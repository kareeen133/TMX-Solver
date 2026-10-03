package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	fhttp "github.com/bogdanfinn/fhttp"
)

type Solver struct {
	t       *Transport
	verbose bool
	profile *Profile
}

type SolveResult struct {
	OrgID           string
	SessionID       string
	Host            string
	BootstrapURL    string
	BootstrapOK     bool
	BootstrapLen    int
	BootstrapMs     int64
	ClearOK         bool
	ClearMs         int64
	StringTable     map[string]string
	ScriptInfo      *ScriptInfo
	Calls           int
	Successful      int
	TotalMs         int64
	GenURLs         int
	ReplayOK        int
	ReplayFail      int
	TmxStarted      bool
	ThxGuid         string
	TmxGuid         string
	TmxNonce        string
	TmxAccepted     bool
	CheckJSURL      string
	CheckJSStatus   int
	CheckJSLen      int
	SubmissionTable *SubmissionTable
	FpPostStatus    int
	FpPostLen       int
	FpPostOK        bool
	FpBodyLen       int
}

func NewSolver(proxy string, verbose bool, profile *Profile) (*Solver, error) {
	if profile == nil {
		profile = DefaultChromeWindowsProfile()
	}
	t, err := NewTransportWithProfile(proxy, TLSProfileForUA(profile.UA))
	if err != nil {
		return nil, err
	}
	t.UA = profile.UA
	t.SecCH = profile.UABrandsHigh
	switch {
	case strings.Contains(strings.ToLower(profile.UA), "android"):
		t.Plat = `"Android"`
	case strings.Contains(strings.ToLower(profile.UA), "iphone"), strings.Contains(strings.ToLower(profile.UA), "ipad"):
		t.Plat = `"iOS"`
	case strings.Contains(profile.UA, "Macintosh"):
		t.Plat = `"macOS"`
	default:
		t.Plat = `"Windows"`
	}
	return &Solver{t: t, verbose: verbose, profile: profile}, nil
}

func (s *Solver) logf(f string, a ...any) {
	if s.verbose {
		fmt.Printf("[tmx] "+f+"\n", a...)
	}
}

func (s *Solver) SolveDeep(orgID, host, sessionID, referer string, mobileConf bool) (*SolveResult, error) {
	if orgID == "" {
		orgID = "usllpic0"
	}
	if host == "" {
		host = "h.online-metrix.net"
	}
	if referer == "" {
		referer = "https://www.example.com/"
	}
	if sessionID == "" {
		sessionID = RandomLowerHex(16)
	}
	res := &SolveResult{OrgID: orgID, SessionID: sessionID, Host: host, StringTable: map[string]string{}}
	overallStart := time.Now()

	var bootstrapURL string
	var body []byte
	var st1 int
	var hdrs1 interface {
		Get(string) string
		Values(string) []string
	}
	var err error
	if mobileConf {
		bootstrapURL = fmt.Sprintf("https://%s/fp/mobile/conf?org_id=%s&session_id=%s",
			host, orgID, sessionID)
		res.BootstrapURL = bootstrapURL
		s.logf("[1] GET /fp/mobile/conf  org=%s session=%s host=%s", orgID, sessionID, host)
		res.Calls++
		t0 := time.Now()
		var rawHdrs fhttp.Header
		st1, body, rawHdrs, err = s.t.GetRaw(bootstrapURL, referer, "*/*")
		res.BootstrapMs = time.Since(t0).Milliseconds()
		hdrs1 = rawHdrs
		if err != nil || st1 != 200 {
			s.logf("    err=%v status=%d", err, st1)
			return res, fmt.Errorf("mobile/conf fetch failed: status=%d", st1)
		}
		res.BootstrapOK = true
		res.Successful++
		res.BootstrapLen = len(body)
	} else {
		bootstrapURL = fmt.Sprintf("https://%s/fp/tags.js?org_id=%s&session_id=%s",
			host, orgID, sessionID)
		res.BootstrapURL = bootstrapURL
		s.logf("[1] GET tags.js  org=%s session=%s host=%s", orgID, sessionID, host)
		res.Calls++
		t0 := time.Now()
		var rawHdrs fhttp.Header
		st1, body, rawHdrs, err = s.t.GetRaw(bootstrapURL, referer, "*/*")
		res.BootstrapMs = time.Since(t0).Milliseconds()
		hdrs1 = rawHdrs
		if err != nil || st1 != 200 || len(body) < 1000 {
			s.logf("    err=%v status=%d len=%d", err, st1, len(body))
			return res, fmt.Errorf("tags.js fetch failed: status=%d", st1)
		}
		res.BootstrapOK = true
		res.Successful++
		res.BootstrapLen = len(body)
	}
	if hdrs1 != nil {
		if v := hdrs1.Get("tmx-nonce"); v != "" {
			res.TmxNonce = v
		}
		for _, sc := range hdrs1.Values("Set-Cookie") {
			if i := strings.Index(sc, "thx_guid="); i >= 0 {
				rest := sc[i+9:]
				if e := strings.IndexAny(rest, "; \r\n"); e >= 0 {
					rest = rest[:e]
				}
				res.ThxGuid = rest
			}
			if i := strings.Index(sc, "tmx_guid="); i >= 0 {
				rest := sc[i+9:]
				if e := strings.IndexAny(rest, "; \r\n"); e >= 0 {
					rest = rest[:e]
				}
				res.TmxGuid = rest
			}
		}
	}
	if res.ThxGuid != "" || res.TmxGuid != "" || res.TmxNonce != "" {
		res.TmxAccepted = true
	}
	s.logf("    status=%d len=%dB time=%dms", st1, len(body), res.BootstrapMs)
	if info, err := ExtractScriptInfo(string(body)); err == nil {
		res.ScriptInfo = info
		res.StringTable = info.StringTable
		s.logf("    cipher=%s.%s table=%d entries", info.CipherObj, info.CipherFn, len(info.StringTable))
	}

	clearURL := fmt.Sprintf("https://%s/fp/clear.png?org_id=%s&session_id=%s", host, orgID, sessionID)
	cis3sid := RandomUpperHex(16)
	checkURL := fmt.Sprintf("https://%s/fp/check.js;CIS3SID=%s?org_id=%s&session_id=%s",
		host, cis3sid, orgID, sessionID)
	s.logf("[2] GET clear.png + check.js (parallel)  CIS3SID=%s", cis3sid)
	res.Calls += 2
	res.TmxStarted = true

	var (
		st2, stC         int
		checkBody        []byte
		hdrsC            fhttp.Header
		errC             error
		clearMs, checkMs int64
		bootWg           sync.WaitGroup
	)
	bootWg.Add(2)
	t1 := time.Now()
	go func() {
		defer bootWg.Done()
		st2, _, _, _ = s.t.GetRaw(clearURL, referer, "image/*")
		clearMs = time.Since(t1).Milliseconds()
	}()
	tC := time.Now()
	go func() {
		defer bootWg.Done()
		stC, checkBody, hdrsC, errC = s.t.GetRaw(checkURL, referer, "*/*")
		checkMs = time.Since(tC).Milliseconds()
	}()
	bootWg.Wait()
	res.ClearMs = clearMs
	if st2 == 200 || st2 == 204 {
		res.ClearOK = true
		res.Successful++
	}
	if errC != nil || stC != 200 || len(checkBody) < 5000 {
		s.logf("    err=%v status=%d len=%d", errC, stC, len(checkBody))
		res.TotalMs = time.Since(overallStart).Milliseconds()
		return res, fmt.Errorf("check.js fetch failed: status=%d", stC)
	}
	res.CheckJSURL = checkURL
	res.CheckJSStatus = stC
	res.CheckJSLen = len(checkBody)
	res.Successful++
	s.logf("    status=%d len=%dB time=%dms", stC, len(checkBody), checkMs)
	if hdrsC != nil {
		if v := hdrsC.Get("tmx-nonce"); v != "" {
			res.TmxNonce = v
		}
		for _, sc := range hdrsC.Values("Set-Cookie") {
			if i := strings.Index(sc, "thx_guid="); i >= 0 && res.ThxGuid == "" {
				rest := sc[i+9:]
				if e := strings.IndexAny(rest, "; \r\n"); e >= 0 {
					rest = rest[:e]
				}
				res.ThxGuid = rest
			}
			if i := strings.Index(sc, "tmx_guid="); i >= 0 && res.TmxGuid == "" {
				rest := sc[i+9:]
				if e := strings.IndexAny(rest, "; \r\n"); e >= 0 {
					rest = rest[:e]
				}
				res.TmxGuid = rest
			}
		}
	}

	tbl := ExtractSubmissionTable(string(checkBody))
	if tbl == nil || tbl.Clear3PNG == "" {
		s.logf("    [tbl] FAIL: could not extract URL submission table")
		res.TotalMs = time.Since(overallStart).Milliseconds()
		return res, fmt.Errorf("submission table extraction failed")
	}
	res.SubmissionTable = tbl
	s.logf("    [tbl] org=%s sid=%s nonce=%s", tbl.OrgID, tbl.SessionID, tbl.Nonce)
	s.logf("    [tbl] clear3.png = %s", trimStr(tbl.Clear3PNG, 130))

	if tbl.SessionID == "" {
		tbl.SessionID = sessionID
	}
	if tbl.OrgID == "" {
		tbl.OrgID = orgID
	}
	sessionProfile := SynthesizeProfile(s.profile, tbl.SessionID)

	builder := &FpBuilder{
		Profile: sessionProfile,
		OrgID:   tbl.OrgID,
		SessID:  tbl.SessionID,
		Nonce:   tbl.Nonce,
		Host:    host,
		CIS3SID: cis3sid,
		Tbl:     tbl,
		PageURL: referer,
		Referer: referer,
	}
	chunks := builder.BuildChunks()
	res.FpBodyLen = 0
	s.logf("[3] sending %d fingerprint chunks (real GET-chunked protocol)", len(chunks))
	dumpPath := os.Getenv("TMX_DUMP_CHUNKS")
	var dumpFile *os.File
	if dumpPath != "" {
		f, err := os.Create(dumpPath)
		if err == nil {
			dumpFile = f
			defer dumpFile.Close()
		}
	}
	totalEnc := 0
	for _, ch := range chunks {
		totalEnc += len(ch.URL)
		if dumpFile != nil {
			fmt.Fprintln(dumpFile, ch.URL)
		}
	}
	res.Calls += len(chunks)

	pacingMs := chunkPacing(chunks, NewSynthRand(res.SessionID, s.profile))
	results := make([]bool, len(chunks))
	accept := "image/avif,image/webp,image/png,image/svg+xml,image/*;q=0.8"
	for i, ch := range chunks {
		if i > 0 {
			time.Sleep(time.Duration(pacingMs[i]) * time.Millisecond)
		}
		stC, _, _, errC := s.t.GetRaw(ch.URL, referer, accept)
		if errC == nil && (stC == 200 || stC == 204) {
			results[i] = true
		} else {
			s.logf("    [chunk FAIL] %s -> %d err=%v url=%s", ch.Kind, stC, errC, trimStr(ch.URL, 120))
		}
	}
	chunkOK, chunkFail := 0, 0
	for _, ok := range results {
		if ok {
			chunkOK++
			res.Successful++
		} else {
			chunkFail++
		}
	}
	res.FpBodyLen = totalEnc
	res.FpPostStatus = 204
	if chunkFail == 0 && chunkOK > 0 {
		res.FpPostOK = true
	}
	s.logf("    chunks ok=%d fail=%d total_url_bytes=%d", chunkOK, chunkFail, totalEnc)

	if res.ThxGuid != "" || res.TmxGuid != "" || res.TmxNonce != "" {
		res.TmxAccepted = true
	}

	res.TotalMs = time.Since(overallStart).Milliseconds()
	return res, nil
}

func trimStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func (s *Solver) Solve(orgID, host, referer, sessionID string) (*SolveResult, error) {
	if orgID == "" {
		orgID = "usllpic0"
	}
	if host == "" {
		host = "h.online-metrix.net"
	}
	if referer == "" {
		referer = "https://www.example.com/"
	}
	if sessionID == "" {
		sessionID = RandomLowerHex(16)
	}

	res := &SolveResult{
		OrgID:       orgID,
		SessionID:   sessionID,
		Host:        host,
		StringTable: map[string]string{},
	}
	overallStart := time.Now()

	bootstrapURL := fmt.Sprintf("https://%s/fp/tags.js?org_id=%s&session_id=%s",
		host, orgID, sessionID)
	res.BootstrapURL = bootstrapURL
	s.logf("[1/2] GET tags.js  org=%s session=%s host=%s", orgID, sessionID, host)
	t0 := time.Now()
	res.Calls++
	st1, body, hdrs1, err := s.t.GetRaw(bootstrapURL, referer, "*/*")
	res.BootstrapMs = time.Since(t0).Milliseconds()
	if err != nil {
		s.logf("    err=%v", err)
	} else {
		res.BootstrapLen = len(body)
		s.logf("    status=%d len=%dB time=%dms", st1, len(body), res.BootstrapMs)
		if st1 == 200 && len(body) > 1000 {
			res.BootstrapOK = true
			res.Successful++
			info, ierr := ExtractScriptInfo(string(body))
			if ierr == nil {
				res.ScriptInfo = info
				res.StringTable = info.StringTable
				s.logf("    cipher=%s.%s  table=%d entries", info.CipherObj, info.CipherFn, len(info.StringTable))
			}
		}
		if hdrs1 != nil {
			if v := hdrs1.Get("tmx-nonce"); v != "" {
				res.TmxNonce = v
			}
			for _, sc := range hdrs1.Values("Set-Cookie") {
				if i := strings.Index(sc, "thx_guid="); i >= 0 {
					rest := sc[i+9:]
					if e := strings.IndexAny(rest, "; \r\n"); e >= 0 {
						rest = rest[:e]
					}
					res.ThxGuid = rest
				}
				if i := strings.Index(sc, "tmx_guid="); i >= 0 {
					rest := sc[i+9:]
					if e := strings.IndexAny(rest, "; \r\n"); e >= 0 {
						rest = rest[:e]
					}
					res.TmxGuid = rest
				}
			}
		}
		if res.ThxGuid != "" || res.TmxGuid != "" || res.TmxNonce != "" {
			res.TmxAccepted = true
		}
	}

	clearURL := fmt.Sprintf("https://%s/fp/clear.png?org_id=%s&session_id=%s",
		host, orgID, sessionID)
	s.logf("[2/2] GET clear.png")
	t1 := time.Now()
	res.Calls++
	st2, _, hdrs2, err := s.t.GetRaw(clearURL, referer, "image/avif,image/webp,image/png,image/svg+xml,image/*;q=0.8")
	res.ClearMs = time.Since(t1).Milliseconds()
	if err != nil {
		s.logf("    err=%v", err)
	} else {
		s.logf("    status=%d time=%dms", st2, res.ClearMs)
		if st2 == 200 || st2 == 204 {
			res.ClearOK = true
			res.Successful++
		}
		if hdrs2 != nil {
			if v := hdrs2.Get("tmx-nonce"); v != "" && res.TmxNonce == "" {
				res.TmxNonce = v
			}
			for _, sc := range hdrs2.Values("Set-Cookie") {
				if res.ThxGuid == "" && strings.Contains(sc, "thx_guid=") {
					i := strings.Index(sc, "thx_guid=")
					rest := sc[i+9:]
					if e := strings.IndexAny(rest, "; \r\n"); e >= 0 {
						rest = rest[:e]
					}
					res.ThxGuid = rest
				}
				if res.TmxGuid == "" && strings.Contains(sc, "tmx_guid=") {
					i := strings.Index(sc, "tmx_guid=")
					rest := sc[i+9:]
					if e := strings.IndexAny(rest, "; \r\n"); e >= 0 {
						rest = rest[:e]
					}
					res.TmxGuid = rest
				}
			}
		}
	}
	if res.ThxGuid != "" || res.TmxGuid != "" || res.TmxNonce != "" {
		res.TmxAccepted = true
	}

	res.TotalMs = time.Since(overallStart).Milliseconds()
	return res, nil
}

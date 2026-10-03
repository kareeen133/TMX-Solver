package main

import (
	"fmt"
	"regexp"
	"strings"
)

type ScriptInfo struct {
	CipherObj     string
	CipherFn      string
	DecryptCtor   string
	DispatchObj   string
	StringTable   map[string]string
	FunctionTable map[string]string
}

var (
	rePrelude  = regexp.MustCompile(`var\s+(td_[A-Za-z0-9]+)=(td_[A-Za-z0-9]+)\|\|\{\};\s*(td_[A-Za-z0-9]+)\.(td_[A-Za-z0-9]+)=function\(([A-Za-z_0-9]+),([A-Za-z_0-9]+)\)\{try\{var\s+([A-Za-z_0-9]+)=\[""\];`)
	reDispatch = regexp.MustCompile(`(td_[A-Za-z0-9]+)\.(tdz_[a-f0-9]+)\s*=\s*new\s+(td_[A-Za-z0-9]+)\.(td_[A-Za-z0-9]+)\(`)
	reTdzDef   = `\.tdz_([a-f0-9]+)\s*=\s*new\s+\S+\.td_[A-Za-z0-9]+\("`
)

func ExtractScriptInfo(src string) (*ScriptInfo, error) {
	info := &ScriptInfo{
		StringTable:   make(map[string]string),
		FunctionTable: make(map[string]string),
	}

	pm := rePrelude.FindStringSubmatch(src)
	if pm == nil {
		return nil, fmt.Errorf("could not locate cipher prelude")
	}
	info.CipherObj = pm[1]
	info.CipherFn = pm[4]

	dm := reDispatch.FindStringSubmatch(src)
	if dm == nil {
		return nil, fmt.Errorf("could not locate tdz dispatch entry")
	}
	info.DispatchObj = dm[1]
	info.DecryptCtor = dm[4]

	pat := regexp.QuoteMeta(info.DispatchObj) + reTdzDef
	re := regexp.MustCompile(pat)
	matches := re.FindAllStringIndex(src, -1)
	for _, m := range matches {
		hashMatch := re.FindStringSubmatch(src[m[0]:m[1]])
		if hashMatch == nil {
			continue
		}
		hashKey := hashMatch[1]
		quoteStart := m[1]
		j := quoteStart
		for j < len(src) {
			if src[j] == '"' && (j == 0 || src[j-1] != '\\') {
				break
			}
			j++
		}
		if j >= len(src) {
			continue
		}
		raw := src[quoteStart:j]
		unesc := unescapeJSHex(raw)
		decoded, err := DecodeTdz(unesc)
		if err == nil {
			info.StringTable[hashKey] = decoded
		}
	}
	return info, nil
}

var reJSHex = regexp.MustCompile(`\\x([0-9a-fA-F]{2})`)

func unescapeJSHex(s string) string {
	return reJSHex.ReplaceAllStringFunc(s, func(m string) string {
		var b byte
		fmt.Sscanf(m[2:], "%x", &b)
		return string([]byte{b})
	})
}

type SubmissionTable struct {
	SidFpHTML    string
	Clear3PNG    string
	DNSCanary    string
	TopFpHTML    string
	H64Clear     string
	IfSid        string
	PageIDPing   string
	BareClear    string
	OrgID        string
	SessionID    string
	Nonce        string
	XKey         string
	AABaseHost   string
	SigningNonce string
}

func ExtractSubmissionTable(src string) *SubmissionTable {
	rePrelude := regexp.MustCompile(`var\s+(td_[A-Za-z0-9]+)=(td_[A-Za-z0-9]+)\|\|\{\};\s*(td_[A-Za-z0-9]+)\.(td_[A-Za-z0-9]+)=function\(([A-Za-z_0-9]+),([A-Za-z_0-9]+)\)\{try\{var\s+([A-Za-z_0-9]+)=\[""\];`)
	pm := rePrelude.FindStringSubmatch(src)
	if pm == nil {
		return nil
	}
	cipherObj := pm[1]
	reBig := regexp.MustCompile(regexp.QuoteMeta(cipherObj) + `\.tdz_([a-f0-9]+)\s*=\s*new\s+\S+\.td_[A-Za-z0-9]+\("([^"]+)"`)
	bigStandalone := regexp.MustCompile(`\b(td_[A-Za-z0-9]+)\s*=\s*new\s+` + regexp.QuoteMeta(cipherObj) + `\.td_[A-Za-z0-9]+\("([^"]+)"`)
	candidates := []string{}
	for _, m := range reBig.FindAllStringSubmatch(src, -1) {
		raw := unescapeJSHex(m[2])
		if dec, err := DecodeTdz(raw); err == nil && strings.Contains(dec, "/fp/clear3.png") {
			candidates = append(candidates, dec)
		}
	}
	for _, m := range bigStandalone.FindAllStringSubmatch(src, -1) {
		raw := unescapeJSHex(m[2])
		if dec, err := DecodeTdz(raw); err == nil && strings.Contains(dec, "/fp/clear3.png") {
			candidates = append(candidates, dec)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	dec := candidates[0]
	t := &SubmissionTable{}

	urlEnd := func(s string) int {
		if e := strings.Index(s, "&pageid=1"); e >= 0 {
			return e + len("&pageid=1")
		}
		if e := strings.Index(s[1:], "https://"); e >= 0 {
			return e + 1
		}
		return len(s)
	}

	if i := strings.Index(dec, "https://"); i >= 0 {
		t.SidFpHTML = dec[i : i+urlEnd(dec[i:])]
	}
	if i := strings.Index(dec, "/fp/clear3.png"); i >= 0 {
		start := strings.LastIndex(dec[:i], "https://")
		if start >= 0 {
			rest := dec[start:]
			t.Clear3PNG = rest[:urlEnd(rest)]
		}
	}
	if i := strings.Index(dec, "/fp/top_fp.html"); i >= 0 {
		start := strings.LastIndex(dec[:i], "https://")
		if start >= 0 {
			rest := dec[start:]
			t.TopFpHTML = rest[:urlEnd(rest)]
		}
	}
	if i := strings.Index(dec, "&di=yes"); i >= 0 {
		start := strings.LastIndex(dec[:i], "https://")
		if start >= 0 {
			t.DNSCanary = dec[start : i+len("&di=yes")]
		}
	}
	if i := strings.Index(dec, "&i=2"); i >= 0 {
		start := strings.LastIndex(dec[:i], "https://h64.online-metrix.net")
		_ = start
		s2 := strings.LastIndex(dec[:i], "https://")
		if s2 >= 0 {
			t.H64Clear = dec[s2 : i+len("&i=2")]
		}
	}
	if i := strings.Index(dec, "&if=sid"); i >= 0 {
		start := strings.LastIndex(dec[:i], "https://")
		if start >= 0 {
			t.IfSid = dec[start : i+len("&if=sid")]
		}
	}

	reOrg := regexp.MustCompile(`org_id=([A-Za-z0-9_]+)`)
	reSid := regexp.MustCompile(`session_id=([a-f0-9]{16,64})`)
	reNonce := regexp.MustCompile(`nonce=([a-f0-9]{8,32})`)
	if m := reOrg.FindStringSubmatch(t.Clear3PNG); len(m) == 2 {
		t.OrgID = m[1]
	}
	if m := reSid.FindStringSubmatch(t.Clear3PNG); len(m) == 2 {
		t.SessionID = m[1]
		t.XKey = m[1]
	}
	if m := reNonce.FindStringSubmatch(t.Clear3PNG); len(m) == 2 {
		t.Nonce = m[1]
	}
	if i := strings.Index(dec, "aa.online-metrix.net"); i >= 0 {
		start := i
		for start > 0 && (dec[start-1] >= 'a' && dec[start-1] <= 'z' || dec[start-1] >= '0' && dec[start-1] <= '9' || dec[start-1] == '.') {
			start--
		}
		t.AABaseHost = dec[start : i+len("aa.online-metrix.net")]
	}

	reSigningNonce := regexp.MustCompile(`/fp/clear1\.png[^"]*?&nonce=([a-f0-9]{8,32})[^"]*?(?:&pageid=1)?([a-f0-9]{32})?([a-f0-9]{16})`)
	for _, m := range bigStandalone.FindAllStringSubmatch(src, -1) {
		raw := unescapeJSHex(m[2])
		dec2, err := DecodeTdz(raw)
		if err != nil || !strings.Contains(dec2, "/fp/clear1.png") {
			continue
		}
		if mm := reSigningNonce.FindStringSubmatch(dec2); len(mm) > 0 {
			if len(mm) >= 4 && mm[3] != "" {
				t.SigningNonce = mm[3]
				break
			}
		}
		idx := strings.Index(dec2, "&pageid=1")
		if idx < 0 {
			continue
		}
		tail := dec2[idx+len("&pageid=1"):]
		reTail := regexp.MustCompile(`^([a-f0-9]{32})([a-f0-9]{16})([a-f0-9]{16})`)
		if mm := reTail.FindStringSubmatch(tail); len(mm) == 4 {
			t.SigningNonce = mm[3]
			break
		}
		reTail2 := regexp.MustCompile(`([a-f0-9]{16})$`)
		if mm := reTail2.FindStringSubmatch(tail); len(mm) == 2 {
			t.SigningNonce = mm[1]
			break
		}
	}
	return t
}

func (s *ScriptInfo) Summary() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CipherObj=%s CipherFn=%s\n", s.CipherObj, s.CipherFn))
	sb.WriteString(fmt.Sprintf("DispatchObj=%s DecryptCtor=%s\n", s.DispatchObj, s.DecryptCtor))
	sb.WriteString(fmt.Sprintf("StringTable entries: %d\n", len(s.StringTable)))
	for h, v := range s.StringTable {
		preview := v
		if len(preview) > 80 {
			preview = preview[:80] + "..."
		}
		sb.WriteString(fmt.Sprintf("  %s: %q\n", h[:16], preview))
	}
	return sb.String()
}

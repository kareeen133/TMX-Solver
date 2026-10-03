package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	org := flag.String("org", "usllpic0", "TMX org_id")
	host := flag.String("host", "h.online-metrix.net", "TMX host (h.online-metrix.net | h64.online-metrix.net | <custom-cname>)")
	referer := flag.String("referer", "https://www.example.com/", "Referer header")
	proxy := flag.String("proxy", "", "proxy URL")
	verbose := flag.Bool("v", true, "verbose")
	jsonOut := flag.Bool("json", false, "JSON output")
	profileName := flag.String("profile", "", "edge-windows | chrome-windows | chrome-android | mobile (defaults to captured Edge)")
	mobile := flag.Bool("mobile", false, "shortcut for chrome-android")
	count := flag.Int("count", 1, "run N times for stress test")
	deep := flag.Bool("deep", false, "deep mode: run tags.js in goja runtime + replay sub-script URLs")
	target := flag.String("target", "", "preset: walmart | ebay | kleinanzeigen | direct (overrides org/host/referer)")
	discoverURL := flag.String("discover", "", "auto-discover TMX endpoint by fetching the given URL and scanning for tags.js + org_id (universal mode for any TMX-protected site)")
	targetsFile := flag.String("targets", "", "JSON file with custom target presets (extends built-in -target list without recompile)")
	verifyFlag := flag.Bool("verify", false, "after solving, hit Walmart login flow and report response shape (Tier-5 verification)")
	verifyEmail := flag.String("verify-email", "test_throwaway_2026@example.com", "test email for Walmart user-check")
	mobileConf := flag.Bool("mobile-conf", false, "use /fp/mobile/conf bootstrap (Kleinanzeigen-style mobile SDK flow)")
	session := flag.String("session", "", "supply the session_id / profilingId to profile (e.g. the cart's fraudPrevention.profilingId). Empty = generate a random one.")
	selfVerify := flag.Bool("self-verify", false, "after solving, run TMX self-verify probes (thx_guid stability, h64 canary, chunk replay) and print risk verdict")
	serve := flag.Bool("serve", false, "run as HTTP API server with admin panel")
	addr := flag.String("addr", ":8080", "API server listen addr (when -serve)")
	apiKey := flag.String("api-key", "", "API key (when -serve). If empty, reads TMX_API_KEY env or generates a random one.")
	logPath := flag.String("log", "requests.jsonl", "request log file (when -serve). Empty = memory-only.")
	maxRecords := flag.Int("max-records", 5000, "max in-memory request records (when -serve)")
	adminUser := flag.String("admin-user", "admin", "admin panel username")
	adminPass := flag.String("admin-pass", "anees3232@", "admin panel password")
	flag.Parse()

	if envPort := os.Getenv("PORT"); envPort != "" {
		*serve = true
		*addr = ":" + envPort
	}

	if *serve {
		runServer(*addr, *apiKey, *proxy, *logPath, *maxRecords, *adminUser, *adminPass)
		return
	}

	if *targetsFile != "" {
		if err := loadCustomTargets(*targetsFile); err != nil {
			fail("targets file: %v", err)
		}
	}
	if *target != "" {
		applyTargetPreset(*target, org, host, referer)
	}

	prof := pickProfile(*profileName, *mobile)

	if *discoverURL != "" {
		t, terr := NewTransport(*proxy)
		if terr != nil {
			fail("discover transport: %v", terr)
		}
		t.UA = prof.UA
		t.SecCH = prof.UABrandsHigh
		dr, derr := DiscoverFromURL(t, *discoverURL, *verbose && !*jsonOut)
		if derr != nil {
			fail("discover: %v", derr)
		}
		*org = dr.OrgID
		*host = dr.Host
		*referer = dr.Referer
		if *verbose && !*jsonOut {
			fmt.Printf("[tmx] discover OK: host=%s org=%s referer=%s\n", *host, *org, *referer)
		}
	}
	if *verbose && !*jsonOut {
		fmt.Printf("[tmx] profile: %s | host: %s | org: %s\n", profileLabel(prof), *host, *org)
	}

	s, err := NewSolver(*proxy, *verbose && !*jsonOut, prof)
	if err != nil {
		fail("init: %v", err)
	}

	totalOK := 0
	totalCalls := 0
	totalDur := int64(0)
	successfulRuns := 0
	var lastRes *SolveResult
	startAll := time.Now()

	for i := 0; i < *count; i++ {
		if *count > 1 && !*jsonOut {
			fmt.Printf("\n=== run %d/%d ===\n", i+1, *count)
		}
		var res *SolveResult
		var err error
		if *deep {
			res, err = s.SolveDeep(*org, *host, *session, *referer, *mobileConf)
		} else {
			res, err = s.Solve(*org, *host, *referer, *session)
		}
		if err != nil {
			fail("solve: %v", err)
		}
		lastRes = res
		totalOK += res.Successful
		totalCalls += res.Calls
		totalDur += res.TotalMs
		if res.Successful == res.Calls {
			successfulRuns++
		}

		if *jsonOut {
			out := map[string]any{
				"success":         res.Successful == res.Calls,
				"org_id":          res.OrgID,
				"session_id":      res.SessionID,
				"host":            res.Host,
				"bootstrap_ok":    res.BootstrapOK,
				"bootstrap_len":   res.BootstrapLen,
				"bootstrap_ms":    res.BootstrapMs,
				"clear_ok":        res.ClearOK,
				"clear_ms":        res.ClearMs,
				"calls":           res.Calls,
				"successful":      res.Successful,
				"string_table":    len(res.StringTable),
				"profile":         profileLabel(prof),
				"total_ms":        res.TotalMs,
				"thx_guid":        res.ThxGuid,
				"tmx_guid":        res.TmxGuid,
				"tmx_nonce":       res.TmxNonce,
				"tmx_accepted":    res.TmxAccepted,
				"check_js_status": res.CheckJSStatus,
				"check_js_len":    res.CheckJSLen,
				"fp_post_status":  res.FpPostStatus,
				"fp_post_ok":      res.FpPostOK,
				"fp_body_len":     res.FpBodyLen,
			}
			b, _ := json.Marshal(out)
			fmt.Println(string(b))
		}
	}

	if !*jsonOut {
		fmt.Println()
		fmt.Println("=== TMX Solver Result ===")
		successRate := 100 * totalOK / max(1, totalCalls)
		fullRunRate := 100 * successfulRuns / max(1, *count)
		if successfulRuns == *count {
			fmt.Printf("[ OK ] %d/%d runs (100%%) | %d/%d calls (%d%%) | avg %dms/run\n",
				successfulRuns, *count, totalOK, totalCalls, successRate, totalDur/int64(max(1, *count)))
		} else {
			fmt.Printf("[FAIL] %d/%d runs (%d%%) | %d/%d calls (%d%%) | avg %dms/run\n",
				successfulRuns, *count, fullRunRate, totalOK, totalCalls, successRate, totalDur/int64(max(1, *count)))
		}
		if lastRes != nil {
			fmt.Printf("  profile         = %s\n", profileLabel(prof))
			fmt.Printf("  host            = %s\n", lastRes.Host)
			fmt.Printf("  org_id          = %s\n", lastRes.OrgID)
			fmt.Printf("  last session    = %s\n", lastRes.SessionID)
			fmt.Printf("  bootstrap       = %v (%dB, %dms)\n", lastRes.BootstrapOK, lastRes.BootstrapLen, lastRes.BootstrapMs)
			fmt.Printf("  clear.png       = %v (%dms)\n", lastRes.ClearOK, lastRes.ClearMs)
			fmt.Printf("  string_table    = %d entries\n", len(lastRes.StringTable))
			fmt.Printf("  total elapsed   = %dms\n", time.Since(startAll).Milliseconds())
			fmt.Println()
			fmt.Println("  --- Tier-2 TMX upstream verification ---")
			fmt.Printf("  thx_guid        = %s\n", shortStr(lastRes.ThxGuid, 60))
			fmt.Printf("  tmx_guid        = %s\n", shortStr(lastRes.TmxGuid, 60))
			fmt.Printf("  tmx-nonce       = %s\n", lastRes.TmxNonce)
			if lastRes.TmxAccepted && (lastRes.ThxGuid != "" || lastRes.TmxGuid != "") {
				fmt.Println("  TMX VERDICT     = [ ACCEPTED ] device cookies issued")
			} else {
				fmt.Println("  TMX VERDICT     = [ NO COOKIES ] no device fingerprint cookies issued")
			}
			if lastRes.CheckJSStatus > 0 {
				fmt.Println()
				fmt.Println("  --- Tier-3 deep fingerprint pipeline ---")
				fmt.Printf("  check.js          = %d (%dB)\n", lastRes.CheckJSStatus, lastRes.CheckJSLen)
				if lastRes.SubmissionTable != nil {
					fmt.Printf("  submit URL        = %s\n", shortStr(lastRes.SubmissionTable.Clear3PNG, 90))
				}
				fmt.Printf("  fp body           = %dB plaintext\n", lastRes.FpBodyLen)
				fmt.Printf("  fp POST           = status=%d ok=%v\n", lastRes.FpPostStatus, lastRes.FpPostOK)
				if lastRes.FpPostOK {
					fmt.Println("  TIER-3 VERDICT  = [ ACCEPTED ] TMX 204'd encoded fingerprint submission")
				} else {
					fmt.Println("  TIER-3 VERDICT  = [ FAIL ] fingerprint POST not accepted")
				}
			}
		}
	}

	if *selfVerify && lastRes != nil && lastRes.SessionID != "" {
		fmt.Println()
		fmt.Println("=== Tier-4 self-verify: TMX risk-score observable probes ===")
		lastChunkURL := ""
		if lastRes.SubmissionTable != nil {
			lastChunkURL = lastRes.SubmissionTable.Clear3PNG
		}
		rsr := s.SelfVerifyRisk(lastRes.OrgID, lastRes.SessionID, lastRes.Host, lastChunkURL)
		fmt.Printf("  thx_guid stable     = %v (first=%s second=%s)\n", rsr.ThxGuidStable, shortStr(rsr.ThxGuidFirst, 32), shortStr(rsr.ThxGuidSecond, 32))
		fmt.Printf("  h64 canary          = status=%d accepted=%v\n", rsr.H64Status, rsr.H64Accepted)
		fmt.Printf("  chunk replay        = status=%d accepted=%v\n", rsr.RepeatChunkStatus, rsr.RepeatChunkAccepted)
		fmt.Printf("  elapsed             = %dms\n", rsr.ElapsedMs)
		for _, n := range rsr.Notes {
			fmt.Printf("  note: %s\n", n)
		}
		fmt.Printf("  RISK VERDICT        = [ %s ]\n", rsr.Verdict)
	}

	if *verifyFlag {
		fmt.Println()
		fmt.Println("=== Tier-5 verify: hitting Walmart login flow ===")
		vr := s.VerifyWalmart(*verifyEmail)
		fmt.Printf("  preLogin status   = %d (len=%d)\n", vr.PreLoginStatus, vr.PreLoginLen)
		fmt.Printf("  preLogin blocked  = %v\n", vr.PreLoginBlock)
		fmt.Printf("  preLogin captcha  = %v\n", vr.PreLoginCaptcha)
		fmt.Printf("  tmx tags.js       = %d (%dms)\n", vr.BootstrapStatus, vr.BootstrapMs)
		fmt.Printf("  gql status        = %d\n", vr.GqlStatus)
		fmt.Printf("  gql body[:220]    = %s\n", vr.GqlBodyHead)
		fmt.Printf("  gql accepted      = %v\n", vr.GqlAccepted)
		fmt.Printf("  gql challenged    = %v\n", vr.GqlChallenged)
		for _, n := range vr.Notes {
			fmt.Printf("  note: %s\n", n)
		}
		if vr.GqlAccepted && !vr.GqlChallenged && !vr.PreLoginBlock {
			fmt.Println("[ OK ] Walmart accepted our session — Tier-5 PASS")
		} else {
			fmt.Println("[FAIL] Walmart did NOT cleanly accept our session — Tier-5 FAIL")
		}
	}

	if successfulRuns != *count {
		os.Exit(1)
	}
}

func runServer(addr, apiKey, proxy, logPath string, maxRecords int, adminUser, adminPass string) {
	if apiKey == "" {
		apiKey = os.Getenv("TMX_API_KEY")
	}
	if apiKey == "" {
		apiKey = RandomLowerHex(16)
		fmt.Printf("[api] no -api-key provided, generated: %s\n", apiKey)
	}
	store, err := NewStore(logPath, maxRecords)
	if err != nil {
		fail("store init: %v", err)
	}
	keys, err := NewKeyStore("apikeys.json")
	if err != nil {
		fail("keystore init: %v", err)
	}
	srv := NewAPIServer(addr, apiKey, proxy, adminUser, adminPass, store, keys)
	if err := srv.Run(); err != nil {
		fail("server: %v", err)
	}
}

func pickProfile(name string, mobile bool) *Profile {
	if mobile && name == "" {
		name = "chrome-android"
	}
	switch strings.ToLower(name) {
	case "chrome-android", "android", "mobile-chrome", "mobile":
		return CapturedChromeAndroidProfile()
	case "safari-ios", "ios", "iphone", "mobile-safari":

		return CapturedChromeAndroidProfile()
	case "chrome-windows", "windows", "chrome":
		return CapturedChromeWindowsProfile()
	case "edge-windows", "edge":
		return CapturedEdge148WindowsProfile()
	case "chrome-mac", "chrome-macos", "mac-chrome", "chrome-osx":
		return CapturedChromeMacOSProfile()
	case "safari-mac", "safari-macos", "mac-safari", "safari", "mac":
		return CapturedSafariMacOSProfile()
	default:
		return CapturedEdge148WindowsProfile()
	}
}

func profileLabel(p *Profile) string {
	if strings.Contains(p.UA, "iPhone") {
		return "Safari/iOS"
	}
	if strings.Contains(p.UA, "Android") {
		return "Chrome/Android"
	}
	if strings.Contains(p.UA, "Edg/") {
		return "Edge/Windows"
	}
	if strings.Contains(p.UA, "Macintosh") {
		if strings.Contains(p.UA, "Chrome/") {
			return "Chrome/macOS"
		}
		return "Safari/macOS"
	}
	return "Chrome/Windows"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func shortStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func fail(f string, a ...any) {
	fmt.Fprintf(os.Stderr, "tmx: "+f+"\n", a...)
	os.Exit(1)
}

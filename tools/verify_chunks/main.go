package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

type harT struct {
	Log struct {
		Entries []struct {
			Request struct {
				URL string `json:"url"`
			} `json:"request"`
			Response struct {
				Status int `json:"status"`
			} `json:"response"`
		} `json:"entries"`
	} `json:"log"`
}

func tdDecode(ciphertextHex, key string) (string, error) {
	if len(ciphertextHex)%2 != 0 {
		return "", fmt.Errorf("odd hex length")
	}
	out := make([]byte, len(ciphertextHex)/2)
	k := 0
	for i := 0; i < len(out); i++ {
		hi, e1 := strconv.ParseUint(ciphertextHex[i*2:i*2+1], 16, 8)
		lo, e2 := strconv.ParseUint(ciphertextHex[i*2+1:i*2+2], 16, 8)
		if e1 != nil || e2 != nil {
			return "", fmt.Errorf("bad hex")
		}
		c := byte(hi<<4) | byte(lo)
		out[i] = c ^ (byte(key[k]) & 0x0A)
		k++
		if k >= len(key) {
			k = 0
		}
	}
	s := string(out)
	if amp := strings.IndexByte(s, '&'); amp >= 0 {
		if _, err := strconv.Atoi(s[:amp]); err == nil {
			return s[amp+1:], nil
		}
	}
	return s, nil
}

func loadHarFields(harPath string) map[string]string {
	data, err := os.ReadFile(harPath)
	must(err)
	var h harT
	must(json.Unmarshal(data, &h))

	var sessID string
	for _, e := range h.Log.Entries {
		u := e.Request.URL
		if !strings.Contains(u, "online-metrix") {
			continue
		}
		p, _ := url.Parse(u)
		if !strings.Contains(p.Path, "/fp/clear") {
			continue
		}
		if sessID == "" {
			sessID = p.Query().Get("session_id")
		}
	}

	fields := map[string]string{}
	collect := func(name, val string) {
		if _, ok := fields[name]; !ok {
			fields[name] = val
		}
	}
	for _, e := range h.Log.Entries {
		u := e.Request.URL
		if !strings.Contains(u, "online-metrix") {
			continue
		}
		p, _ := url.Parse(u)
		if !strings.Contains(p.Path, "/fp/clear") {
			continue
		}
		q := p.Query()
		for _, kind := range []string{"je", "ja", "jb", "jf"} {
			val := q.Get(kind)
			if val == "" {
				continue
			}
			plain, err := tdDecode(val, sessID)
			if err != nil {
				continue
			}
			for seg := range strings.SplitSeq(plain, "&") {
				if seg == "" {
					continue
				}
				eq := strings.IndexByte(seg, '=')
				if eq < 0 {
					continue
				}
				name := seg[:eq]
				v := seg[eq+1:]
				if dec, err := url.QueryUnescape(v); err == nil {
					v = dec
				}
				collect(name, v)
			}
		}
	}
	return fields
}

func runSolverAndCapture(profile, target string) (map[string]string, []string) {
	tmp := os.TempDir()
	dumpPath := tmp + string(os.PathSeparator) + "tmx_solver_dump_" + profile + ".txt"
	defer os.Remove(dumpPath)
	os.Setenv("TMX_DUMP_CHUNKS", dumpPath)

	cmd := exec.Command("go", "run", ".", "-deep", "-profile", profile, "-target", target, "-count", "1")
	cmd.Env = append(os.Environ(), "TMX_DUMP_CHUNKS="+dumpPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		must(err)
	}
	data, err := os.ReadFile(dumpPath)
	must(err)
	urls := strings.Split(strings.TrimSpace(string(data)), "\n")

	var sessID string
	for _, u := range urls {
		p, err := url.Parse(u)
		if err != nil {
			continue
		}
		if !strings.Contains(p.Path, "/fp/clear") {
			continue
		}
		if sessID == "" {
			sessID = p.Query().Get("session_id")
			break
		}
	}
	fields := map[string]string{}
	collect := func(n, v string) {
		if _, ok := fields[n]; !ok {
			fields[n] = v
		}
	}
	for _, u := range urls {
		p, err := url.Parse(u)
		if err != nil {
			continue
		}
		if !strings.Contains(p.Path, "/fp/clear") {
			continue
		}
		q := p.Query()
		for _, kind := range []string{"je", "ja", "jb", "jf"} {
			val := q.Get(kind)
			if val == "" {
				continue
			}
			plain, err := tdDecode(val, sessID)
			if err != nil {
				continue
			}
			for seg := range strings.SplitSeq(plain, "&") {
				if seg == "" {
					continue
				}
				eq := strings.IndexByte(seg, '=')
				if eq < 0 {
					continue
				}
				name := seg[:eq]
				v := seg[eq+1:]
				if dec, err := url.QueryUnescape(v); err == nil {
					v = dec
				}
				collect(name, v)
			}
		}
	}
	return fields, urls
}

func main() {
	harPath := "captures/chrome_win.har"
	if len(os.Args) > 1 {
		harPath = os.Args[1]
	}
	fmt.Println("=== Comparing solver output vs live browser HAR ===")
	fmt.Println("HAR :", harPath)

	profile := "chrome-windows"
	if strings.Contains(harPath, "android") {
		profile = "chrome-android"
	} else if strings.Contains(harPath, "real_session") {
		profile = "edge-windows"
	}
	har := loadHarFields(harPath)
	solver, urls := runSolverAndCapture(profile, "direct")
	fmt.Println("profile:", profile)

	fmt.Printf("solver dumped %d outgoing chunks\n\n", len(urls))

	skip := map[string]bool{
		"sid_rnd": true, "sid_date": true, "sid_sig": true, "sid_key": true,
		"sid_type": true, "lh": true, "dr": true,
		"jftn": true, "jfh": true,
		"hh":  true,
		"wei": true,
	}

	keys := map[string]bool{}
	for k := range har {
		keys[k] = true
	}
	for k := range solver {
		keys[k] = true
	}
	keyList := make([]string, 0, len(keys))
	for k := range keys {
		keyList = append(keyList, k)
	}
	sort.Strings(keyList)

	matches, mismatches, soloHar, soloSolver := 0, 0, 0, 0
	fmt.Printf("%-12s %-9s %s\n", "FIELD", "STATUS", "DETAIL")
	fmt.Println(strings.Repeat("-", 100))
	for _, k := range keyList {
		hv, hOk := har[k]
		sv, sOk := solver[k]
		if skip[k] {
			fmt.Printf("%-12s %-9s (per-session random — excluded)\n", k, "SKIP")
			continue
		}
		switch {
		case hOk && sOk && hv == sv:
			matches++
			disp := hv
			if len(disp) > 70 {
				disp = disp[:70] + "..."
			}
			fmt.Printf("%-12s %-9s %s\n", k, "MATCH", disp)
		case hOk && sOk:
			mismatches++
			fmt.Printf("%-12s %-9s\n", k, "DIFFER")
			fmt.Printf("  har    = %s\n", trim(hv, 90))
			fmt.Printf("  solver = %s\n", trim(sv, 90))
		case hOk && !sOk:
			soloHar++
			fmt.Printf("%-12s %-9s har-only: %s\n", k, "HAR_ONLY", trim(hv, 70))
		case !hOk && sOk:
			soloSolver++
			fmt.Printf("%-12s %-9s solver-only: %s\n", k, "SOLV_ONLY", trim(sv, 70))
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("MATCH      : %d fields byte-identical\n", matches)
	fmt.Printf("DIFFER     : %d fields differ (inspect above)\n", mismatches)
	fmt.Printf("HAR_ONLY   : %d fields in browser but not solver (missing!)\n", soloHar)
	fmt.Printf("SOLV_ONLY  : %d fields in solver but not browser (extra!)\n", soloSolver)
	fmt.Println(strings.Repeat("=", 60))
	if mismatches == 0 && soloHar == 0 && soloSolver == 0 {
		fmt.Println("VERDICT: solver output is BYTE-FAITHFUL to a real Chrome/Windows browser session.")
	} else {
		fmt.Println("VERDICT: solver output diverges from real browser — see fields above.")
	}
}

func trim(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

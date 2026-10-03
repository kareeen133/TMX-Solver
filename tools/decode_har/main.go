package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
)

type HAR struct {
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
		hi, err1 := strconv.ParseUint(ciphertextHex[i*2:i*2+1], 16, 8)
		lo, err2 := strconv.ParseUint(ciphertextHex[i*2+1:i*2+2], 16, 8)
		if err1 != nil || err2 != nil {
			return "", fmt.Errorf("bad hex at %d", i)
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

func main() {
	harPath := flag.String("har", "captures/real_session.har", "HAR file path")
	flag.Parse()

	data, err := os.ReadFile(*harPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read har: %v\n", err)
		os.Exit(1)
	}
	var h HAR
	if err := json.Unmarshal(data, &h); err != nil {
		fmt.Fprintf(os.Stderr, "parse har: %v\n", err)
		os.Exit(1)
	}

	type chunk struct {
		Status int
		Kind   string
		URL    string
		Plain  string
		Path   string
	}
	var chunks []chunk
	var sessID, nonce, orgID, cis3sid string

	for _, e := range h.Log.Entries {
		u := e.Request.URL
		if !strings.Contains(u, "online-metrix") && !strings.Contains(u, "metrix.net") {
			continue
		}
		parsed, err := url.Parse(u)
		if err != nil {
			continue
		}
		if !strings.Contains(parsed.Path, "/fp/clear") {
			continue
		}
		q := parsed.Query()
		if sessID == "" {
			sessID = q.Get("session_id")
			nonce = q.Get("nonce")
			orgID = q.Get("org_id")
		}

		if strings.Contains(parsed.Path, ";CIS3SID=") {
			if i := strings.Index(parsed.Path, "CIS3SID="); i >= 0 {
				rest := parsed.Path[i+8:]
				if e := strings.IndexAny(rest, ";?"); e >= 0 {
					rest = rest[:e]
				}
				cis3sid = rest
			}
		}
		for _, kind := range []string{"je", "ja", "jb", "jf"} {
			val := q.Get(kind)
			if val == "" {
				continue
			}
			plain, err := tdDecode(val, sessID)
			if err != nil {
				plain = "<decode err: " + err.Error() + ">"
			}
			chunks = append(chunks, chunk{
				Status: e.Response.Status,
				Kind:   kind,
				URL:    u,
				Plain:  plain,
				Path:   parsed.Path,
			})
		}
	}

	fmt.Printf("=== HAR decode summary ===\n")
	fmt.Printf("session_id : %s\n", sessID)
	fmt.Printf("nonce      : %s\n", nonce)
	fmt.Printf("org_id     : %s\n", orgID)
	fmt.Printf("CIS3SID    : %s\n", cis3sid)
	fmt.Printf("chunks     : %d\n\n", len(chunks))

	for i, c := range chunks {
		short := c.Path
		if strings.HasSuffix(short, ".png") || strings.Contains(short, ".png;") {
			if last := strings.LastIndex(short, "/"); last >= 0 {
				short = short[last+1:]
			}
		}
		fmt.Printf("--- chunk %d  status=%d  %s  param=%s ---\n", i+1, c.Status, short, c.Kind)
		fmt.Println(c.Plain)
		fmt.Println()
	}

	fmt.Println("=== EXTRACTED FIELDS (paste into Captured*Profile) ===")
	fields := map[string]string{}
	collect := func(name, val string) {
		if _, ok := fields[name]; ok {
			return
		}
		fields[name] = val
	}
	for _, c := range chunks {

		segs := strings.Split(c.Plain, "&")
		for _, s := range segs {
			if s == "" {
				continue
			}
			eq := strings.IndexByte(s, '=')
			if eq < 0 {
				continue
			}
			name := s[:eq]
			val := s[eq+1:]
			switch name {
			case "ex3", "ex4", "ex5", "ex6", "ex6s", "ex7", "ex7s",
				"gl_h", "glh_h", "medh", "ssi", "lsa", "hbd", "audh",
				"mathr", "mt", "ph", "hh", "batst", "uah", "ual", "uistl",
				"pm", "p", "wei", "ccd", "wglv", "wglr", "lq",
				"jso", "jsb", "jsou", "jsbu",
				"c", "z", "f", "af", "sxy", "dpr", "mn", "scd", "lh",
				"pl", "nhc", "ndm", "tzd", "dr":
				dec, _ := url.QueryUnescape(val)
				collect(name, dec)
			}
		}
	}

	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := fields[k]

		disp := v
		if len(disp) > 200 {
			disp = disp[:200] + "...(" + strconv.Itoa(len(v)) + " bytes total)"
		}
		fmt.Printf("  %-8s = %s\n", k, disp)
	}

	fmt.Println()
	fmt.Println("=== GO PASTE SNIPPET ===")
	emit := func(field, key string) {
		if v, ok := fields[key]; ok {
			fmt.Printf("\t%s: %q,\n", field, v)
		}
	}
	emit("Ex3", "ex3")
	emit("Ex4", "ex4")
	emit("Ex5", "ex5")
	emit("Ex6", "ex6")
	emit("Ex6s", "ex6s")
	emit("Ex7", "ex7")
	emit("Ex7s", "ex7s")
	emit("GLH", "gl_h")
	emit("GLHH", "glh_h")
	emit("MedH", "medh")
	emit("SSIH", "ssi")
	emit("LSAH", "lsa")
	emit("HBD", "hbd")
	emit("AudH", "audh")
	emit("MathR", "mathr")
	emit("MtH", "mt")
	emit("PluginH", "ph")
	emit("HistH", "hh")
	emit("BatSt", "batst")
	emit("UAH", "uah")
	emit("UAL", "ual")
	emit("UIStl", "uistl")
	emit("PM", "pm")
	emit("PEnum", "p")
	emit("WeI", "wei")
	emit("CCD", "ccd")
	emit("WebGLVendor", "wglv")
	emit("WebGLRenderer", "wglr")
}

package main

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func TdEncode(plaintext, key string) string {
	if key == "" {
		key = "0"
	}
	prefixed := strconv.Itoa(len(plaintext)) + "&" + plaintext
	const hexChars = "0123456789abcdef"
	var b strings.Builder
	b.Grow(len(prefixed) * 2)
	k := 0
	for i := 0; i < len(prefixed); i++ {
		c := byte(prefixed[i]) ^ (byte(key[k]) & 0x0A)
		k++
		if k >= len(key) {
			k = 0
		}
		b.WriteByte(hexChars[(c>>4)&0xF])
		b.WriteByte(hexChars[c&0xF])
	}
	return b.String()
}

func TdDecode(ciphertextHex, key string) (string, error) {
	if len(ciphertextHex)%2 != 0 {
		return "", fmt.Errorf("ciphertext odd length")
	}
	out := make([]byte, len(ciphertextHex)/2)
	k := 0
	for i := 0; i < len(out); i++ {
		hi, _ := strconv.ParseUint(ciphertextHex[i*2:i*2+1], 16, 8)
		lo, _ := strconv.ParseUint(ciphertextHex[i*2+1:i*2+2], 16, 8)
		c := byte(hi<<4) | byte(lo)
		out[i] = c ^ (byte(key[k]) & 0x0A)
		k++
		if k >= len(key) {
			k = 0
		}
	}
	s := string(out)
	if amp := strings.IndexByte(s, '&'); amp >= 0 {
		return s[amp+1:], nil
	}
	return s, nil
}

type FpChunk struct {
	URL  string
	Kind string
}

type FpBuilder struct {
	Profile *Profile
	OrgID   string
	SessID  string
	Nonce   string
	Host    string
	CIS3SID string
	Tbl     *SubmissionTable
	PageURL string
	Referer string
}

func (b *FpBuilder) qs(extra string) string {
	return fmt.Sprintf("org_id=%s&session_id=%s&nonce=%s&pageid=1%s",
		b.OrgID, b.SessID, b.Nonce, extra)
}

func (b *FpBuilder) clearURL(extra string) string {
	return fmt.Sprintf("https://%s/fp/clear.png?%s", b.Host, b.qs(extra))
}

func (b *FpBuilder) clear3URL(extra string) string {
	return fmt.Sprintf("https://%s/fp/clear3.png;CIS3SID=%s?%s", b.Host, b.CIS3SID, b.qs(extra))
}

func (b *FpBuilder) clear1URL(extra string) string {
	return fmt.Sprintf("https://%s/fp/clear1.png;CIS3SID=%s?%s", b.Host, b.CIS3SID, b.qs(extra))
}

func (b *FpBuilder) checkURL() string {
	return fmt.Sprintf("https://%s/fp/check.js;CIS3SID=%s?org_id=%s&session_id=%s&nonce=%s&pageid=1",
		b.Host, b.CIS3SID, b.OrgID, b.SessID, b.Nonce)
}

func (b *FpBuilder) signingNonce() string {
	if b.Tbl != nil && b.Tbl.SigningNonce != "" {
		return b.Tbl.SigningNonce
	}
	return b.Nonce
}

func (b *FpBuilder) BuildChunks() []FpChunk {
	pf := b.Profile
	if pf == nil {
		pf = CapturedEdge148WindowsProfile()
	}
	out := []FpChunk{}
	enc := func(p string) string { return TdEncode(p, b.SessID) }

	plat := "Windows"
	platLong := "Windows 11"
	browser := "Chrome"
	browserVer := pf.UAFullVersion
	if strings.Contains(pf.UA, "Edg/") {
		browser = "Edge"
		if i := strings.Index(pf.UA, "Edg/"); i >= 0 {
			rest := pf.UA[i+4:]
			if sp := strings.Index(rest, " "); sp >= 0 {
				browserVer = rest[:sp]
			} else {
				browserVer = rest
			}
		}
	}
	if strings.Contains(pf.UA, "Android") {
		plat, platLong = "Android", "Android 13"
	} else if strings.Contains(pf.UA, "iPhone") {
		plat, platLong = "iOS", "iOS 17"
	} else if strings.Contains(strings.ToLower(pf.UA), "mac") {
		plat, platLong = "macOS", "macOS 14"
	}
	jsoLong := platLong
	jsbLong := browser + " " + browserVer
	if pf.JsoLong != "" {
		jsoLong = pf.JsoLong
	}
	if pf.JsbLong != "" {
		jsbLong = pf.JsbLong
	}

	jbBrowser := fmt.Sprintf("&jsou=%s&jso=%s&jsbu=%s&jsb=%s",
		urlEscape(plat), urlEscape(jsoLong), urlEscape(browser), urlEscape(jsbLong))
	if pf.IsMobile {
		jbBrowser += "&jsmu=true"
	}
	out = append(out, FpChunk{URL: b.checkURL() + "&jb=" + enc(jbBrowser), Kind: "check.js+jb"})

	out = append(out, FpChunk{URL: b.clearURL("&ck=0&m=2"), Kind: "ping_m2"})
	out = append(out, FpChunk{URL: b.clearURL("&ck=0&m=1"), Kind: "ping_m1"})

	if pf.LSAH != "" {
		lsaBody := "lsa=" + pf.LSAH + "&la_old="
		out = append(out, FpChunk{URL: b.clearURL("&jb=" + enc(lsaBody)), Kind: "lsa"})

		lsbHash := sha1.Sum([]byte("lsb-" + b.SessID))
		lsbBody := "lsb=" + hex.EncodeToString(lsbHash[:])
		out = append(out, FpChunk{URL: b.clearURL("&jf=" + enc(lsbBody)), Kind: "lsb"})
	}

	pageURL := b.PageURL
	if pageURL == "" {
		pageURL = b.Referer
	}
	dr := b.Referer
	if dr == "" {
		dr = pageURL
	}

	fxStr := pf.FxStr
	afStr := pf.AfStr
	if fxStr == "" {
		fW := int(float64(pf.ScreenW)*pf.PixelRatio + 0.5)
		fH := int(float64(pf.ScreenH)*pf.PixelRatio + 0.5)
		fxStr = fmt.Sprintf("%dx%d", fW, fH)
	}
	if afStr == "" {
		afStr = fxStr
	}
	dprCSV := pf.DprCSV
	if dprCSV == "" {
		dprStr := strconv.FormatFloat(pf.PixelRatio, 'g', -1, 64)
		dprCSV = fmt.Sprintf("%s,%d,%d,%d,%d,%d,%d,%d,%d,10,10",
			dprStr, pf.ScreenW, pf.ScreenH, pf.AvailW, pf.AvailH,
			pf.OuterW, pf.OuterH, pf.InnerW, pf.InnerH)
	}
	sxyVal := fmt.Sprintf("%g", pf.PixelRatio*10)
	if pf.IsMobile {
		sxyVal = "0"
	}
	jaSegs := []string{
		"&c=" + strconv.Itoa(pf.TZOffset),
		"&z=" + strconv.Itoa(pf.DSTOffset),
		"&f=" + fxStr,
		"&af=" + afStr,
		"&sxy=" + sxyVal + "x" + sxyVal,
		"&dpr=" + dprCSV,
		"&mt=" + pf.MtH,
		"&mn=" + strconv.Itoa(len(pf.Mimes)),
		"&scd=" + strconv.Itoa(pf.ColorDepth),
		"&lh=" + urlEscape(pageURL),
		"&pl=" + strconv.Itoa(len(pf.Plugins)),
		"&ph=" + pf.PluginH,
		"&hh=" + pf.HistH,
		"&jso=" + urlEscape(jsoLong),
		"&jsb=" + urlEscape(jsbLong),
		"&jsou=" + urlEscape(plat),
		"&jsbu=" + urlEscape(browser),
	}
	if pf.IsMobile {
		jaSegs = append(jaSegs, "&jsmu=true")
	}
	jaSegs = append(jaSegs,
		"&nhc="+strconv.Itoa(pf.HardwareConc),
		"&ndm="+strconv.Itoa(pf.DeviceMemory),
	)
	if pf.IsMobile {
		jaSegs = append(jaSegs, "&nmtp=1")
	} else {
		jaSegs = append(jaSegs, "&nmtp=0")
	}
	jaSegs = append(jaSegs,
		"&tzd="+urlEscape(pf.TZName),
		"&mathr="+pf.MathR,
		"&dr="+urlEscape(dr),
		"&p="+pf.PEnum,
	)
	jaPayload := strings.Join(jaSegs, "")
	jbLq := "lq=" + urlEscape(pf.UA)
	out = append(out, FpChunk{
		URL:  b.clearURL("&ja=" + enc(jaPayload) + "&jb=" + enc(jbLq)),
		Kind: "ja_screen_plugin+lq",
	})

	if pf.MedH != "" {
		out = append(out, FpChunk{URL: b.clear3URL("&bbv=3&jac=1&je=" + enc("&medh="+pf.MedH)), Kind: "medh"})
	}

	if sig, err := GenerateSidSignature(b.signingNonce()); err == nil {
		jfBody := fmt.Sprintf("sid_rnd=%s&sid_date=%s&sid_type=%s&sid_key=%s&sid_sig=%s&sifr=1",
			sig.Rnd, sig.Date, sig.Type, sig.KeyHex, sig.SigHex)
		out = append(out, FpChunk{URL: b.clear1URL("&jf=" + enc(jfBody)), Kind: "jf_sid_iframe"})
	}

	pmSegs := []string{}
	if pf.PM != "" {
		pmSegs = append(pmSegs, "&pm="+pf.PM)
	}
	if pf.BatSt != "" {
		pmSegs = append(pmSegs, "&batst="+urlEscape(pf.BatSt))
	}
	if pf.AudH != "" {
		pmSegs = append(pmSegs, "&audh="+pf.AudH)
	}
	pmSegs = append(pmSegs, "&jso="+urlEscape(jsoLong))
	if pf.UAH != "" {
		pmSegs = append(pmSegs, "&uah="+urlEscape(pf.UAH))
	}
	if pf.UAL != "" {
		pmSegs = append(pmSegs, "&ual="+urlEscape(pf.UAL))
	}
	if pf.UIStl != "" {
		pmSegs = append(pmSegs, "&uistl="+pf.UIStl)
	}
	if len(pmSegs) > 0 {
		out = append(out, FpChunk{URL: b.clearURL("&jac=1&je=" + enc(strings.Join(pmSegs, ""))), Kind: "pm_bat_aud_uah"})
	}

	hbdSegs := []string{}
	if pf.HBD != "" {
		hbdSegs = append(hbdSegs, "&hbd="+pf.HBD)
	}
	hbdSegs = append(hbdSegs, "&wglv="+urlEscape(pf.WebGLVendor))
	hbdSegs = append(hbdSegs, "&wglr="+urlEscape(pf.WebGLRenderer))
	if pf.Ex3 != "" {
		hbdSegs = append(hbdSegs, "&ex3="+pf.Ex3)
	}
	if pf.Ex6 != "" {
		hbdSegs = append(hbdSegs, "&ex6="+pf.Ex6, "&ex6s="+pf.Ex6s)
	}
	if pf.GLH != "" {
		hbdSegs = append(hbdSegs, "&gl_h="+pf.GLH)
	}
	hbdSegs = append(hbdSegs, "&wglv="+urlEscape(pf.WebGLVendor))
	hbdSegs = append(hbdSegs, "&wglr="+urlEscape(pf.WebGLRenderer))
	if pf.GLHH != "" {
		hbdSegs = append(hbdSegs, "&glh_h="+pf.GLHH)
	}
	if pf.Ex4 != "" {
		hbdSegs = append(hbdSegs, "&ex4="+pf.Ex4)
	}
	if pf.Ex5 != "" {
		hbdSegs = append(hbdSegs, "&ex5="+pf.Ex5)
	}
	if pf.Ex7 != "" {
		hbdSegs = append(hbdSegs, "&ex7="+pf.Ex7, "&ex7s="+pf.Ex7s)
	}
	if pf.CCD != "" {
		hbdSegs = append(hbdSegs, "&ccd="+pf.CCD)
	}
	hbdSegs = append(hbdSegs, "&bbv=3")
	hbdPayload := strings.Join(hbdSegs, "")
	out = append(out, FpChunk{URL: b.clear3URL("&jac=1&je=" + enc(hbdPayload)), Kind: "hbd_canvas_webgl"})

	if sig, err := GenerateSidSignature(b.signingNonce()); err == nil {
		jfBody := fmt.Sprintf("sid_rnd=%s&sid_date=%s&sid_type=%s&sid_key=%s&sid_sig=%s&sifr=0",
			sig.Rnd, sig.Date, sig.Type, sig.KeyHex, sig.SigHex)
		out = append(out, FpChunk{URL: b.clear1URL("&jf=" + enc(jfBody)), Kind: "jf_sid_top"})
	}

	if pf.SSIH != "" {
		ssiBody := ""
		if pf.WnId != "" {
			ssiBody += "&wnid=" + pf.WnId
		}
		ssiBody += "&ssi=" + pf.SSIH
		out = append(out, FpChunk{URL: b.clearURL("&jac=1&je=" + enc(ssiBody)), Kind: "ssi"})
	}

	jfn := pf.Jfn
	if jfn == 0 {
		jfn = 142
	}
	totalLen := len(hbdPayload) + 64
	jfh := md5sum(hbdPayload)
	jftnPayload := fmt.Sprintf("&jfn=%d&jfh=%s&jftn=0:%d:%d&bbv=3", jfn, jfh, totalLen, jfn)
	out = append(out, FpChunk{URL: b.clear3URL("&jac=1&je=" + enc(jftnPayload)), Kind: "jfn_integrity"})

	if !pf.IsMobile {
		rdtPorts := []int{63333, 5900, 5901, 5902, 5903, 3389, 5950, 5931, 5939, 6039, 5944, 6040, 5938, 5279, 7070, 2112}
		rdtParts := make([]string, len(rdtPorts))
		for i, p := range rdtPorts {
			rdtParts[i] = fmt.Sprintf("%d-1500", p)
		}
		rdPayload := "rd=&rdt=" + strings.Join(rdtParts, ",") + "&bbv=3"
		out = append(out, FpChunk{URL: b.clear3URL("&je=" + enc(rdPayload)), Kind: "rd_portscan"})
	}

	if pf.WeI != "" {
		webrtcSegs := []string{
			"&wei=" + pf.WeI,
			"&wii=" + b.synthInternalIP(),
			"&wim=" + b.synthMDNS(),
			"&bbv=3",
		}
		out = append(out, FpChunk{URL: b.clear3URL("&jac=1&je=" + enc(strings.Join(webrtcSegs, ""))), Kind: "webrtc"})
	}

	out = append(out, FpChunk{URL: b.clearURL("&je=" + enc("&shd=closed&bbv=3")), Kind: "shadow_root"})

	bpco := "false"
	if strings.Contains(pf.HBD, "wd_1") || strings.Contains(pf.HBD, "ph_1") ||
		strings.Contains(pf.HBD, "sl_1") || strings.Contains(pf.HBD, "pc_1") {
		bpco = "true"
	}
	out = append(out, FpChunk{URL: b.clearURL("&je=" + enc("&bpco="+bpco+"&bbv=3")), Kind: "bpco_compound"})

	cspNonce := ""
	if b.Tbl != nil && b.Tbl.SessionID != "" {
		h := md5.Sum([]byte("csp-nonce-" + b.SessID))
		cspNonce = hex.EncodeToString(h[:])[:16]
	}
	out = append(out, FpChunk{URL: b.clearURL("&je=" + enc("&csp_nonce="+cspNonce+"&bbv=3")), Kind: "csp_nonce"})

	tdidmBody := "&data=tdidm:tmx-db=&tmx-sid=&tmx-sid1="
	out = append(out, FpChunk{URL: b.clearURL("&je=" + enc(tdidmBody+"&bbv=3")), Kind: "tdidm"})

	if b.Tbl != nil && b.Tbl.OrgID != "" {
		esURL := fmt.Sprintf("https://%s/fp/es.js?org_id=%s&session_id=%s&nonce=%s&pageid=1",
			b.Host, b.OrgID, b.SessID, b.Nonce)
		out = append(out, FpChunk{URL: esURL, Kind: "es_script"})
	}

	if b.Tbl != nil && b.Tbl.SidFpHTML != "" {
		out = append(out, FpChunk{URL: b.Tbl.SidFpHTML, Kind: "sid_fp.html"})
	}
	if b.Tbl != nil && b.Tbl.TopFpHTML != "" {
		out = append(out, FpChunk{URL: b.Tbl.TopFpHTML, Kind: "top_fp.html"})
	}
	if b.Tbl != nil && b.Tbl.H64Clear != "" {
		out = append(out, FpChunk{URL: b.Tbl.H64Clear, Kind: "h64_canary"})
	}
	if b.Tbl != nil && b.Tbl.IfSid != "" {
		out = append(out, FpChunk{URL: b.Tbl.IfSid, Kind: "if=sid"})
	}
	if b.Tbl != nil && b.Tbl.DNSCanary != "" {
		out = append(out, FpChunk{URL: b.Tbl.DNSCanary, Kind: "dns_canary"})
	}

	return out
}

func md5sum(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

func urlEscape(s string) string { return url.QueryEscape(s) }

func (b *FpBuilder) synthInternalIP() string {
	h := md5.Sum([]byte("internal-ip-" + b.SessID))
	switch h[0] & 0x03 {
	case 0:
		return fmt.Sprintf("192.168.%d.%d", h[1]%255, (h[2]%254)+1)
	case 1:
		return fmt.Sprintf("10.%d.%d.%d", h[1]%255, h[2]%255, (h[3]%254)+1)
	default:
		return fmt.Sprintf("172.%d.%d.%d", 16+(h[1]%16), h[2]%255, (h[3]%254)+1)
	}
}

func (b *FpBuilder) synthMDNS() string {
	h := md5.Sum([]byte("mdns-" + b.SessID))
	return hex.EncodeToString(h[:]) + ".local"
}

package main

import (
	"fmt"
	"strings"
)

func chunkPacing(chunks []FpChunk, sr *SynthRand) []int {
	out := make([]int, len(chunks))
	for i, ch := range chunks {
		if i == 0 {
			out[i] = 0
			continue
		}
		out[i] = pacingForKind(ch.Kind, sr)
	}
	return out
}

func pacingForKind(kind string, sr *SynthRand) int {
	switch {
	case kind == "check.js+jb":
		return 0
	case kind == "ping_m2":
		return sr.IntRange(30, 90)
	case kind == "ping_m1":
		return sr.IntRange(20, 60)
	case kind == "lsa":
		return sr.IntRange(60, 140)
	case kind == "ja_screen_plugin+lq":
		return sr.IntRange(80, 200)
	case kind == "medh":
		return sr.IntRange(50, 160)
	case strings.HasPrefix(kind, "jf_sid_iframe"):
		return sr.IntRange(80, 160)
	case strings.HasPrefix(kind, "jf_sid_top"):
		return sr.IntRange(5, 25)
	case kind == "pm_bat_aud_uah":
		return sr.IntRange(100, 220)
	case kind == "hbd_canvas_webgl":
		return sr.IntRange(160, 380)
	case kind == "ssi":
		return sr.IntRange(80, 180)
	case kind == "jfn_integrity":
		return sr.IntRange(30, 90)
	case kind == "rd_portscan":
		return sr.IntRange(200, 480)
	case kind == "webrtc":
		return sr.IntRange(120, 480)
	case kind == "shadow_root":
		return sr.IntRange(20, 60)
	case kind == "es_script":
		return sr.IntRange(80, 200)
	case kind == "lsb":
		return sr.IntRange(10, 40)
	case kind == "bpco_compound", kind == "csp_nonce", kind == "tdidm":
		return sr.IntRange(20, 80)
	case kind == "sid_fp.html":
		return sr.IntRange(80, 180)
	case kind == "top_fp.html":
		return sr.IntRange(50, 140)
	case kind == "h64_canary", kind == "if=sid", kind == "dns_canary":
		return sr.IntRange(40, 120)
	default:
		return sr.IntRange(40, 140)
	}
}

type BehavioralPayload struct {
	MouseTrail   string
	KeyPress     string
	ScrollEvents string
	FocusEvents  string
}

func GenerateBehavioral(sr *SynthRand, isMobile bool) *BehavioralPayload {
	if isMobile {
		return &BehavioralPayload{
			MouseTrail:   "",
			KeyPress:     generateKeyPress(sr, 4, 8),
			ScrollEvents: generateScroll(sr, 8, 14),
			FocusEvents:  generateFocus(sr, 2, 4),
		}
	}
	return &BehavioralPayload{
		MouseTrail:   generateMouseTrail(sr, 18, 36),
		KeyPress:     generateKeyPress(sr, 6, 14),
		ScrollEvents: generateScroll(sr, 3, 8),
		FocusEvents:  generateFocus(sr, 1, 3),
	}
}

func generateMouseTrail(sr *SynthRand, minPoints, maxPoints int) string {
	n := sr.IntRange(minPoints, maxPoints)
	startX := float64(sr.IntRange(50, 800))
	startY := float64(sr.IntRange(50, 600))
	endX := float64(sr.IntRange(50, 1200))
	endY := float64(sr.IntRange(50, 700))
	cx := (startX + endX) / 2.0
	cy := (startY + endY) / 2.0
	cx += sr.FloatRange(-120, 120)
	cy += sr.FloatRange(-120, 120)

	var b strings.Builder
	t := 0
	for i := range n {
		u := float64(i) / float64(n-1)
		x := (1-u)*(1-u)*startX + 2*(1-u)*u*cx + u*u*endX
		y := (1-u)*(1-u)*startY + 2*(1-u)*u*cy + u*u*endY
		x += sr.FloatRange(-1.5, 1.5)
		y += sr.FloatRange(-1.5, 1.5)
		t += sr.IntRange(8, 35)
		if i > 0 {
			b.WriteByte(';')
		}
		fmt.Fprintf(&b, "%d,%d,%d", int(x), int(y), t)
	}
	return b.String()
}

func generateKeyPress(sr *SynthRand, minN, maxN int) string {
	n := sr.IntRange(minN, maxN)
	var b strings.Builder
	t := sr.IntRange(120, 400)
	for i := range n {
		ch := sr.IntRange(48, 122)
		dwell := sr.IntRange(40, 180)
		t += dwell + sr.IntRange(60, 280)
		if i > 0 {
			b.WriteByte(';')
		}
		fmt.Fprintf(&b, "%d,%d", ch, t)
	}
	return b.String()
}

func generateScroll(sr *SynthRand, minN, maxN int) string {
	n := sr.IntRange(minN, maxN)
	var b strings.Builder
	y := 0
	t := sr.IntRange(200, 800)
	for i := range n {
		dy := sr.IntRange(40, 220)
		y += dy
		t += sr.IntRange(80, 400)
		if i > 0 {
			b.WriteByte(';')
		}
		fmt.Fprintf(&b, "%d,%d", y, t)
	}
	return b.String()
}

func generateFocus(sr *SynthRand, minN, maxN int) string {
	n := sr.IntRange(minN, maxN)
	var b strings.Builder
	t := sr.IntRange(500, 1500)
	for i := range n {
		fb := i % 2
		t += sr.IntRange(800, 4000)
		if i > 0 {
			b.WriteByte(';')
		}
		fmt.Fprintf(&b, "%d,%d", fb, t)
	}
	return b.String()
}

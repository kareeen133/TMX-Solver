package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	mrand "math/rand"
	"sort"
	"strconv"
	"strings"
)

type SynthRand struct {
	r *mrand.Rand
}

func NewSynthRand(sessionID string, profile *Profile) *SynthRand {
	h := fnv.New64a()
	h.Write([]byte(sessionID))
	h.Write([]byte(profile.UA))
	h.Write([]byte(profile.WebGLVendor))
	h.Write([]byte(profile.WebGLRenderer))
	seed := int64(h.Sum64())
	return &SynthRand{r: mrand.New(mrand.NewSource(seed))}
}

func (sr *SynthRand) Bytes(n int) []byte {
	b := make([]byte, n)
	sr.r.Read(b)
	return b
}

func (sr *SynthRand) FloatRange(lo, hi float64) float64 {
	return lo + sr.r.Float64()*(hi-lo)
}

func (sr *SynthRand) IntRange(lo, hi int) int {
	if hi <= lo {
		return lo
	}
	return lo + sr.r.Intn(hi-lo)
}

func SynthEx3(sr *SynthRand, profile *Profile) string {
	h := sha1.New()
	h.Write([]byte("canvas2d-ex3-v1"))
	h.Write([]byte(profile.WebGLRenderer))
	h.Write([]byte(profile.UA))
	h.Write([]byte("2d@Browsers~%fingGPRint$&,"))
	h.Write(sr.Bytes(8))
	return hex.EncodeToString(h.Sum(nil))
}

func SynthEx6(sr *SynthRand, profile *Profile) (string, string) {
	h := md5.New()
	h.Write([]byte("canvas2d-ex6-extended-v1"))
	h.Write([]byte(profile.WebGLRenderer))
	h.Write([]byte("Magenta7pxArialPapayaWhip23pxArialDarkOrangeFuchsia0.51.0redD83EDD2918pxArialyellow"))
	h.Write([]byte(profile.UA))
	h.Write(sr.Bytes(8))
	md5hex := hex.EncodeToString(h.Sum(nil))
	ts := int64(1300000000000) + sr.r.Int63n(300000000000)
	return md5hex, strconv.FormatInt(ts, 10)
}

func SynthEx4(sr *SynthRand, profile *Profile) string {
	h := md5.New()
	h.Write([]byte("webgl-ex4-v1"))
	h.Write([]byte(profile.WebGLRenderer))
	h.Write([]byte(profile.WebGLVendor))
	h.Write([]byte(`attribute vec4 position;attribute vec4 color;attribute vec3 corner;varying vec4 v_color;varying vec3 v_corner;void main(){gl_Position=position;v_color=color;v_corner=corner;}`))
	h.Write(sr.Bytes(8))
	return hex.EncodeToString(h.Sum(nil))
}

func SynthEx5(sr *SynthRand, profile *Profile) string {
	h := md5.New()
	h.Write([]byte("webgl2-ex5-v1"))
	h.Write([]byte(profile.WebGLRenderer))
	h.Write([]byte(profile.WebGLVendor))
	h.Write([]byte(`precision highp float;attribute vec4 position;attribute vec4 color;attribute vec3 corner;varying vec4 v_color;varying vec3 v_corner;`))
	h.Write(sr.Bytes(8))
	return hex.EncodeToString(h.Sum(nil))
}

func SynthEx7(sr *SynthRand, profile *Profile) (string, string) {
	h := md5.New()
	h.Write([]byte("webgl-ex7-exotic-v1"))
	h.Write([]byte(profile.WebGLRenderer))
	h.Write(sr.Bytes(8))
	md5hex := hex.EncodeToString(h.Sum(nil))
	ts := int64(1400000000000) + sr.r.Int63n(300000000000)
	return md5hex, strconv.FormatInt(ts, 10)
}

func SynthGLH(sr *SynthRand, profile *Profile) string {
	h := sha1.New()
	h.Write([]byte("webgl-shader-hash-v1"))
	h.Write([]byte(profile.WebGLRenderer))
	h.Write([]byte(profile.WebGLVendor))
	h.Write(sr.Bytes(8))
	return hex.EncodeToString(h.Sum(nil))
}

func SynthGLHH(sr *SynthRand, profile *Profile) string {
	h := sha1.New()
	h.Write([]byte("webgl-shader-highprec-hash-v1"))
	h.Write([]byte(profile.WebGLRenderer))
	h.Write(sr.Bytes(8))
	return hex.EncodeToString(h.Sum(nil))
}

func SynthAudH(sr *SynthRand, profile *Profile) string {
	h := sha256.New()
	h.Write([]byte("audio-context-triangle-v1"))
	h.Write([]byte(profile.UA))
	h.Write([]byte(profile.WebGLRenderer))
	for range 32 {
		v := sr.FloatRange(-100, -30)
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, uint64(int64(v*1e9)))
		h.Write(buf)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func SynthMathR(sr *SynthRand, profile *Profile) string {
	h := sha256.New()
	h.Write([]byte("math-precision-v1"))
	h.Write([]byte(profile.UA))
	if strings.Contains(profile.UA, "Mac") {
		h.Write([]byte("darwin-x86_64"))
	} else if strings.Contains(profile.UA, "Windows") {
		h.Write([]byte("windows-amd64"))
	} else {
		h.Write([]byte("linux-amd64"))
	}
	h.Write(sr.Bytes(4))
	return hex.EncodeToString(h.Sum(nil))
}

func SynthMtH(profile *Profile) string {
	mimes := []string{"application/pdf", "text/pdf"}
	if len(profile.Mimes) > 0 {
		mimes = profile.Mimes
	}
	sort.Strings(mimes)
	h := md5.New()
	h.Write([]byte(strings.Join(mimes, "|")))
	return hex.EncodeToString(h.Sum(nil))
}

func SynthPluginH(profile *Profile) string {
	plugins := profile.Plugins
	if len(plugins) == 0 {
		plugins = []string{"PDF Viewer", "Chrome PDF Viewer"}
	}
	sort.Strings(plugins)
	h := md5.New()
	h.Write([]byte(strings.Join(plugins, "|")))
	return hex.EncodeToString(h.Sum(nil))
}

func SynthHistH(sr *SynthRand) string {
	h := md5.New()
	h.Write([]byte("browser-history-v1"))
	h.Write(sr.Bytes(8))
	return hex.EncodeToString(h.Sum(nil))
}

func SynthLSAH(sr *SynthRand) string {
	h := md5.New()
	h.Write([]byte("localstorage-v1"))
	h.Write(sr.Bytes(8))
	return hex.EncodeToString(h.Sum(nil))
}

func SynthMedH(sr *SynthRand) string {
	audioIn := sr.IntRange(1, 3)
	audioOut := sr.IntRange(1, 4)
	videoIn := sr.IntRange(0, 2)
	h := sha256.New()
	h.Write([]byte("mediadevices-v1"))
	h.Write(sr.Bytes(8))
	hashHex := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("(%d,%d,%d,%s)", audioIn, audioOut, videoIn, hashHex)
}

func SynthSSIH(sr *SynthRand, profile *Profile) string {
	voiceCount := 4
	if strings.Contains(profile.UA, "Mac") {
		voiceCount = sr.IntRange(40, 100)
	} else if strings.Contains(profile.UA, "Android") {
		voiceCount = sr.IntRange(1, 4)
	} else if strings.Contains(profile.UA, "Windows") {
		voiceCount = sr.IntRange(4, 10)
	}
	h := sha1.New()
	h.Write([]byte("speech-synth-v1"))
	h.Write([]byte(profile.UA))
	h.Write(sr.Bytes(8))
	hashHex := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("1,%d,0,%s", voiceCount, hashHex)
}

type HBDOptions struct {
	Profile  *Profile
	IsMobile bool
}

func SynthHBD(opts *HBDOptions) string {
	p := opts.Profile
	if p == nil {
		return ""
	}
	wd := 0
	ch := 1
	pq := 0
	pi := len(p.Plugins)
	la := 1
	ln := 2
	pc := 0
	ph := 0
	mi := 0
	if opts.IsMobile {
		mi = 1
	}
	sl := 0
	cw := 1
	sv := 0

	if strings.Contains(p.UA, "Safari") && !strings.Contains(p.UA, "Chrome") && !strings.Contains(p.UA, "Edg") {
		ch = 0
		la = 0
		ln = 0
		cw = 0
	}

	flags := fmt.Sprintf(":wd_%d:ch_%d:pq_%d:pi_%d:la_%d:ln_%d:pc_%d:ph_%d:mi_%d:sl_%d:cw_%d:sv_%d",
		wd, ch, pq, pi, la, ln, pc, ph, mi, sl, cw, sv)
	geom := fmt.Sprintf(",%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%g",
		p.InnerH, p.InnerW, 0, 10, 0, 0, p.ScreenW, p.ScreenH, p.ScreenW, p.ScreenH, p.PixelRatio)
	tail := ":rt_false,true,true,true:ic_true:ps_default,prompt"
	return flags + geom + tail
}

func SynthUAH(profile *Profile) string {
	major := profile.UAFullVersion
	if major == "" {
		major = "146.0.0.0"
	}
	short := strings.Split(major, ".")[0]
	plat := "Windows"
	platVer := "19.0.0"
	arch := "x86"
	mobile := false
	model := ""
	if strings.Contains(profile.UA, "Android") {
		plat = "Android"
		platVer = "13"
		arch = "arm"
		mobile = true
		model = ""
	} else if strings.Contains(profile.UA, "Mac") {
		plat = "macOS"
		platVer = "14.7.0"
		arch = "arm"
	} else if strings.Contains(profile.UA, "iPhone") || strings.Contains(profile.UA, "iPad") {
		plat = "iOS"
		platVer = "17.5.1"
		arch = "arm"
		mobile = true
	}
	browser := "Google Chrome"
	if strings.Contains(profile.UA, "Edg/") {
		browser = "Microsoft Edge"
	}
	obj := map[string]any{
		"architecture":    arch,
		"bitness":         "64",
		"brands":          []map[string]string{{"brand": "Chromium", "version": short}, {"brand": browser, "version": short}, {"brand": "Not?A_Brand", "version": "99"}},
		"fullVersionList": []map[string]string{{"brand": "Chromium", "version": major}, {"brand": browser, "version": major}, {"brand": "Not?A_Brand", "version": "99.0.0.0"}},
		"mobile":          mobile,
		"model":           model,
		"platform":        plat,
		"platformVersion": platVer,
		"wow64":           false,
	}
	b, _ := json.Marshal(obj)
	return string(b)
}

func SynthUAL(profile *Profile) string {
	major := "146"
	if profile.UAFullVersion != "" {
		major = strings.Split(profile.UAFullVersion, ".")[0]
	}
	plat := "Windows"
	mobile := false
	if strings.Contains(profile.UA, "Android") {
		plat = "Android"
		mobile = true
	} else if strings.Contains(profile.UA, "Mac") {
		plat = "macOS"
	} else if strings.Contains(profile.UA, "iPhone") || strings.Contains(profile.UA, "iPad") {
		plat = "iOS"
		mobile = true
	}
	browser := "Google Chrome"
	if strings.Contains(profile.UA, "Edg/") {
		browser = "Microsoft Edge"
	}
	obj := map[string]any{
		"brands":   []map[string]string{{"brand": "Chromium", "version": major}, {"brand": browser, "version": major}, {"brand": "Not?A_Brand", "version": "99"}},
		"mobile":   mobile,
		"platform": plat,
	}
	b, _ := json.Marshal(obj)
	return string(b)
}

func SynthesizeProfile(base *Profile, sessionID string) *Profile {
	if base == nil {
		return nil
	}
	out := *base
	sr := NewSynthRand(sessionID, &out)

	out.Ex3 = SynthEx3(sr, &out)
	out.Ex4 = SynthEx4(sr, &out)
	out.Ex5 = SynthEx5(sr, &out)
	out.Ex6, out.Ex6s = SynthEx6(sr, &out)
	out.Ex7, out.Ex7s = SynthEx7(sr, &out)
	out.GLH = SynthGLH(sr, &out)
	out.GLHH = SynthGLHH(sr, &out)
	out.AudH = SynthAudH(sr, &out)
	out.MathR = SynthMathR(sr, &out)
	out.MtH = SynthMtH(&out)
	out.PluginH = SynthPluginH(&out)
	out.HistH = SynthHistH(sr)
	out.LSAH = SynthLSAH(sr)
	out.MedH = SynthMedH(sr)
	out.SSIH = SynthSSIH(sr, &out)
	out.HBD = SynthHBD(&HBDOptions{Profile: &out, IsMobile: out.IsMobile})
	out.UAH = SynthUAH(&out)
	out.UAL = SynthUAL(&out)
	out.WeI = SynthWeI(sr)

	return &out
}

func SynthWeI(sr *SynthRand) string {
	for {
		a := sr.IntRange(1, 224)
		if a == 10 || a == 127 || a == 169 || a == 172 || a == 192 || a == 224 || a == 0 {
			continue
		}
		b := sr.IntRange(0, 255)
		c := sr.IntRange(0, 255)
		d := sr.IntRange(1, 254)
		return fmt.Sprintf("%d.%d.%d.%d", a, b, c, d)
	}
}

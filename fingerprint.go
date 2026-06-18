package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

type Profile struct {
	UA              string
	Vendor          string
	Platform        string
	Languages       []string
	HardwareConc    int
	DeviceMemory    int
	ScreenW         int
	ScreenH         int
	AvailW          int
	AvailH          int
	InnerW          int
	InnerH          int
	OuterW          int
	OuterH          int
	ColorDepth      int
	PixelRatio      float64
	TZOffset        int
	TZName          string
	WebGLVendor     string
	WebGLRenderer   string
	GPU             string
	Plugins         []string
	Mimes           []string
	UAFullVersion   string
	UABrandsHigh    string
	WebdriverVal    bool
	CanvasHash      string
	AudioHash       string
	WebRTCFP        string
	FontList        []string
	BatteryLevel    float64
	BatteryCharging bool
	IsMobile        bool

	Ex3      string
	Ex4      string
	Ex5      string
	Ex6      string
	Ex6s     string
	Ex7      string
	Ex7s     string
	GLH      string
	GLHH     string
	MedH     string
	SSIH     string
	LSAH     string
	HBD      string
	AudH     string
	MathR    string
	MtH      string
	PluginH  string
	HistH    string
	BatSt    string
	UAH      string
	UAL      string
	UIStl    string
	PM       string
	PEnum    string
	WeI      string
	JsoLong  string
	JsbLong  string
	CCD      string

	DprCSV    string
	DSTOffset int
	WnId      string
	FxStr     string
	AfStr     string
	Jfn       int
}

func CapturedEdge148WindowsProfile() *Profile {
	return &Profile{
		UA:              "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36 Edg/148.0.0.0",
		Vendor:          "Google Inc.",
		Platform:        "Win32",
		Languages:       []string{"en-US", "en"},
		HardwareConc:    16,
		DeviceMemory:    32,
		ScreenW:         1536,
		ScreenH:         960,
		AvailW:          1536,
		AvailH:          960,
		InnerW:          941,
		InnerH:          1051,
		OuterW:          1051,
		OuterH:          941,
		ColorDepth:      32,
		PixelRatio:      1.25,
		TZOffset:        330,
		TZName:          "Asia/Calcutta",
		WebGLVendor:     "Google Inc. (Intel)",
		WebGLRenderer:   "ANGLE (Intel, Intel(R) Graphics (0x0000A7AB) Direct3D11 vs_5_0 ps_5_0, D3D11)",
		GPU:             "Intel",
		Plugins:         []string{"PDF Viewer", "Chrome PDF Viewer", "Chromium PDF Viewer", "Microsoft Edge PDF Viewer", "WebKit built-in PDF"},
		Mimes:           []string{"application/pdf", "text/pdf"},
		UAFullVersion:   "148.0.0.0",
		UABrandsHigh:    `"Microsoft Edge";v="148.0.3309.43", "Chromium";v="148.0.7340.45", "Not?A_Brand";v="99.0.0.0"`,
		WebdriverVal:    false,
		FontList:        defaultFontList(),
		BatteryLevel:    1.0,
		BatteryCharging: true,
		IsMobile:        false,

		Ex3:     "c5c4ab7ba5f5046a681dba4af707b303c083ee49",
		Ex4:     "f35c32b3eeca65a856565a60664e29a1",
		Ex5:     "061909363ac211314613dc2dbb387668",
		Ex6:     "551f76a6d2d50ea758c7fdd228c97123",
		Ex6s:    "1399942794901",
		Ex7:     "4ba93e9892926cfde2be983d02e9284a",
		Ex7s:    "1514661666859",
		GLH:     "32c3b4a95e053c5e77e279d817fb917580de081b",
		GLHH:    "9fd84c9cb5446da8ec06f32c00ef379ec5c34e33",
		MedH:    "(1,1,1,e12ead95566221e7c4e320e56363dcb477a3c93a4a8510beb25f01bca07d4e8d)",
		SSIH:    "1,1,0,1107a7cfa352275b7770e6e51a09cf838c0593df",
		LSAH:    "ffe63f49c8be4f8ba77f0ada639a24fb",
		HBD:     ":wd_1:ch_1:pq_0:pi_5:la_1:ln_2:pc_0:ph_0:mi_0:sl_0:cw_1:sv_0,941,1051,0,10,0,0,1536,960,1536,960,32,32,1.25:rt_false,true,true,true:ic_true:ps_default,prompt",
		AudH:    "cefbae478677f02fbbd973617692dbd9c6450bf5641669ebef1595ab745a2117",
		MathR:   "71f548e4b3608bfdfb89c996ffd4f70f12ae3d66adf429356b9f3bc677b2fe2f",
		MtH:     "27f51d3149e6bf209b66bd387b0af3c4",
		PluginH: "e802dfa555193f4ebe8993eb4a99290d",
		HistH:   "67a7eb2f315a8ff6473cfd37279b1c49",
		BatSt:   `{"level":1.00,"status":"charging"}`,
		UAH:     `{"architecture":"x86","bitness":"64","brands":[{"brand":"Chromium","version":"148"},{"brand":"Microsoft Edge","version":"148"},{"brand":"Not/A)Brand","version":"99"}],"fullVersionList":[{"brand":"Chromium","version":"148.0.7778.97"},{"brand":"Microsoft Edge","version":"148.0.3967.54"},{"brand":"Not/A)Brand","version":"99.0.0.0"}],"mobile":false,"model":"","platform":"Windows","platformVersion":"19.0.0","wow64":false}`,
		UAL:     `{"brands":[{"brand":"Chromium","version":"148"},{"brand":"Microsoft Edge","version":"148"},{"brand":"Not/A)Brand","version":"99"}],"mobile":false,"platform":"Windows"}`,
		UIStl:   "light",
		PM:      "no",
		PEnum:   "plugin_flash^false!plugin_windows_media_player^false!plugin_adobe_acrobat^false!plugin_quicktime^false!plugin_shockwave^false!plugin_realplayer^false!plugin_vlc_player^false!plugin_devalvr^false!plugin_svg_viewer^false!plugin_java^false",
		JsoLong:   "Windows 11",
		JsbLong:   "Edge 148",
		CCD:       "",
		DprCSV:    "1.25,1536,960,1536,960,1028,850,1051,941,10,10",
		DSTOffset: 0,
		WnId:      "",
		FxStr:     "1920x1200",
		AfStr:     "1920x1200",
		Jfn:       142,
	}
}

func CapturedChromeWindowsProfile() *Profile {
	return &Profile{
		UA:              "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36",
		Vendor:          "Google Inc.",
		Platform:        "Win32",
		Languages:       []string{"en-US", "en"},
		HardwareConc:    16,
		DeviceMemory:    32,
		ScreenW:         1280,
		ScreenH:         800,
		AvailW:          1280,
		AvailH:          800,
		InnerW:          1280,
		InnerH:          800,
		OuterW:          1298,
		OuterH:          890,
		ColorDepth:      32,
		PixelRatio:      1.0,
		TZOffset:        -300,
		TZName:          "America/New_York",
		WebGLVendor:     "Google Inc. (Intel)",
		WebGLRenderer:   "ANGLE (Intel, Intel(R) Graphics (0x0000A7AB) Direct3D11 vs_5_0 ps_5_0, D3D11)",
		GPU:             "Intel",
		Plugins:         []string{"PDF Viewer", "Chrome PDF Viewer", "Chromium PDF Viewer", "Microsoft Edge PDF Viewer", "WebKit built-in PDF"},
		Mimes:           []string{"application/pdf", "text/pdf"},
		UAFullVersion:   "148.0.0.0",
		UABrandsHigh:    `"Chromium";v="148.0.0.0", "Google Chrome";v="148.0.0.0", "Not?A_Brand";v="99.0.0.0"`,
		WebdriverVal:    false,
		FontList:        defaultFontList(),
		BatteryLevel:    1.0,
		BatteryCharging: true,
		IsMobile:        false,

		Ex3:     "c5c4ab7ba5f5046a681dba4af707b303c083ee49",
		Ex4:     "f35c32b3eeca65a856565a60664e29a1",
		Ex5:     "061909363ac211314613dc2dbb387668",
		Ex6:     "551f76a6d2d50ea758c7fdd228c97123",
		Ex6s:    "1399942794901",
		Ex7:     "4ba93e9892926cfde2be983d02e9284a",
		Ex7s:    "1514661666859",
		GLH:     "32c3b4a95e053c5e77e279d817fb917580de081b",
		GLHH:    "9fd84c9cb5446da8ec06f32c00ef379ec5c34e33",
		MedH:    "(1,1,1,e12ead95566221e7c4e320e56363dcb477a3c93a4a8510beb25f01bca07d4e8d)",
		SSIH:    "1,1,0,1107a7cfa352275b7770e6e51a09cf838c0593df",
		LSAH:    "4d8e2011a9d84c03aa0e236b21c47191",
		HBD:     ":wd_1:ch_1:pq_0:pi_5:la_1:ln_1:pc_0:ph_0:mi_0:sl_0:cw_1:sv_0,890,1298,0,10,0,0,1280,800,1280,800,32,32,1.0000000149011612:rt_false,true,true,true:ic_true:ps_default,prompt",
		AudH:    "cefbae478677f02fbbd973617692dbd9c6450bf5641669ebef1595ab745a2117",
		MathR:   "71f548e4b3608bfdfb89c996ffd4f70f12ae3d66adf429356b9f3bc677b2fe2f",
		MtH:     "27f51d3149e6bf209b66bd387b0af3c4",
		PluginH: "e802dfa555193f4ebe8993eb4a99290d",
		HistH:   "781d5d49e6bdfc2af4aba83112ace15d",
		BatSt:   `{"level":1.00,"status":"charging"}`,
		UAH:     `{"architecture":"x86","bitness":"64","brands":[{"brand":"Chromium","version":"148"},{"brand":"Google Chrome","version":"148"},{"brand":"Not?A_Brand","version":"99"}],"fullVersionList":[{"brand":"Chromium","version":"148.0.0.0"},{"brand":"Google Chrome","version":"148.0.0.0"},{"brand":"Not?A_Brand","version":"99.0.0.0"}],"mobile":false,"model":"","platform":"Windows","platformVersion":"10.0","wow64":false}`,
		UAL:     `{"brands":[{"brand":"Chromium","version":"148"},{"brand":"Google Chrome","version":"148"},{"brand":"Not?A_Brand","version":"99"}],"mobile":false,"platform":"Windows"}`,
		UIStl:   "light",
		PM:      "no",
		PEnum:   "plugin_flash^false!plugin_windows_media_player^false!plugin_adobe_acrobat^false!plugin_quicktime^false!plugin_shockwave^false!plugin_realplayer^false!plugin_vlc_player^false!plugin_devalvr^false!plugin_svg_viewer^false!plugin_java^false",
		JsoLong:   "Windows 11",
		JsbLong:   "Chrome 148",
		CCD:       "1",
		DprCSV:    "1.0000000149011612,1280,800,1280,800,1280,800,1298,890,10,10",
		DSTOffset: 60,
		WnId:      "webrtc_no_internal_data",
		FxStr:     "1280x800",
		AfStr:     "1280x800",
		Jfn:       142,
	}
}

func CapturedChromeAndroidProfile() *Profile {
	return &Profile{
		UA:              "Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Mobile Safari/537.36",
		Vendor:          "Google Inc.",
		Platform:        "Linux armv81",
		Languages:       []string{"en-US", "en"},
		HardwareConc:    16,
		DeviceMemory:    32,
		ScreenW:         412,
		ScreenH:         915,
		AvailW:          412,
		AvailH:          915,
		InnerW:          412,
		InnerH:          915,
		OuterW:          412,
		OuterH:          915,
		ColorDepth:      32,
		PixelRatio:      2.625,
		TZOffset:        -300,
		TZName:          "America/New_York",
		WebGLVendor:     "Google Inc. (Intel)",
		WebGLRenderer:   "ANGLE (Intel, Intel(R) Graphics (0x0000A7AB) Direct3D11 vs_5_0 ps_5_0, D3D11)",
		GPU:             "Intel",
		Plugins:         []string{"PDF Viewer", "Chrome PDF Viewer", "Chromium PDF Viewer", "Microsoft Edge PDF Viewer", "WebKit built-in PDF"},
		Mimes:           []string{"application/pdf", "text/pdf"},
		UAFullVersion:   "148.0.0.0",
		UABrandsHigh:    `"Chromium";v="148.0.0.0", "Google Chrome";v="148.0.0.0", "Not?A_Brand";v="99.0.0.0"`,
		WebdriverVal:    false,
		FontList:        mobileFontList(),
		BatteryLevel:    1.0,
		BatteryCharging: true,
		IsMobile:        true,

		Ex3:     "c5c4ab7ba5f5046a681dba4af707b303c083ee49",
		Ex4:     "f35c32b3eeca65a856565a60664e29a1",
		Ex5:     "061909363ac211314613dc2dbb387668",
		Ex6:     "551f76a6d2d50ea758c7fdd228c97123",
		Ex6s:    "1399942794901",
		Ex7:     "4ba93e9892926cfde2be983d02e9284a",
		Ex7s:    "1514661666859",
		GLH:     "32c3b4a95e053c5e77e279d817fb917580de081b",
		GLHH:    "9fd84c9cb5446da8ec06f32c00ef379ec5c34e33",
		MedH:    "(1,1,1,e12ead95566221e7c4e320e56363dcb477a3c93a4a8510beb25f01bca07d4e8d)",
		SSIH:    "1,1,0,1107a7cfa352275b7770e6e51a09cf838c0593df",
		LSAH:    "3324a11d6ca64801a822e2d9d19515cc",
		HBD:     ":wd_1:ch_1:pq_0:pi_5:la_1:ln_1:pc_0:ph_0:mi_0:te_1:sl_0:cw_1:sv_0,915,412,0,0,0,0,412,915,412,915,32,32,2.6249998807907104:rt_false,true,true,true:ic_true:ps_default,prompt",
		AudH:    "cefbae478677f02fbbd973617692dbd9c6450bf5641669ebef1595ab745a2117",
		MathR:   "71f548e4b3608bfdfb89c996ffd4f70f12ae3d66adf429356b9f3bc677b2fe2f",
		MtH:     "27f51d3149e6bf209b66bd387b0af3c4",
		PluginH: "e802dfa555193f4ebe8993eb4a99290d",
		HistH:   "555a035e67f75078f200fef01bef23d7",
		BatSt:   `{"level":1.00,"status":"charging"}`,
		UAH:     `{"architecture":"arm","bitness":"64","brands":[{"brand":"Chromium","version":"148"},{"brand":"Google Chrome","version":"148"},{"brand":"Not?A_Brand","version":"99"}],"fullVersionList":[{"brand":"Chromium","version":"148.0.0.0"},{"brand":"Google Chrome","version":"148.0.0.0"},{"brand":"Not?A_Brand","version":"99.0.0.0"}],"mobile":true,"model":"","platform":"Android","platformVersion":"13","wow64":false}`,
		UAL:     `{"brands":[{"brand":"Chromium","version":"148"},{"brand":"Google Chrome","version":"148"},{"brand":"Not?A_Brand","version":"99"}],"mobile":true,"platform":"Android"}`,
		UIStl:   "light",
		PM:      "no",
		PEnum:   "plugin_flash^false!plugin_windows_media_player^false!plugin_adobe_acrobat^false!plugin_quicktime^false!plugin_shockwave^false!plugin_realplayer^false!plugin_vlc_player^false!plugin_devalvr^false!plugin_svg_viewer^false!plugin_java^false",
		JsoLong:   "Android 13",
		JsbLong:   "Chrome 148",
		CCD:       "4",
		DprCSV:    "2.6249998807907104,412,915,412,915,980,2178,412,915,0,0",
		DSTOffset: 60,
		WnId:      "webrtc_no_internal_data",
		FxStr:     "1080x2400",
		AfStr:     "1081.4999508857727x2401.8748909235",
		Jfn:       14,
	}
}

func DefaultChromeWindowsProfile() *Profile { return CapturedChromeWindowsProfile() }
func DefaultChromeAndroidProfile() *Profile { return CapturedChromeAndroidProfile() }
func DefaultSafariIOSProfile() *Profile     { return CapturedSafariMacOSProfile() }
func DefaultChromeMacOSProfile() *Profile   { return CapturedChromeMacOSProfile() }
func DefaultSafariMacOSProfile() *Profile   { return CapturedSafariMacOSProfile() }

func CapturedChromeMacOSProfile() *Profile {
	return &Profile{
		UA:              "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36",
		Vendor:          "Google Inc.",
		Platform:        "MacIntel",
		Languages:       []string{"en-US", "en"},
		HardwareConc:    10,
		DeviceMemory:    8,
		ScreenW:         1728,
		ScreenH:         1117,
		AvailW:          1728,
		AvailH:          1079,
		InnerW:          1728,
		InnerH:          1015,
		OuterW:          1728,
		OuterH:          1117,
		ColorDepth:      30,
		PixelRatio:      2.0,
		TZOffset:        -300,
		TZName:          "America/New_York",
		WebGLVendor:     "Google Inc. (Apple)",
		WebGLRenderer:   "ANGLE (Apple, ANGLE Metal Renderer: Apple M2, Unspecified Version)",
		GPU:             "Apple M2",
		Plugins:         []string{"PDF Viewer", "Chrome PDF Viewer", "Chromium PDF Viewer", "Microsoft Edge PDF Viewer", "WebKit built-in PDF"},
		Mimes:           []string{"application/pdf", "text/pdf"},
		UAFullVersion:   "146.0.0.0",
		UABrandsHigh:    `"Chromium";v="146", "Google Chrome";v="146", "Not?A_Brand";v="99"`,
		WebdriverVal:    false,
		FontList:        macFontList(),
		BatteryLevel:    1.0,
		BatteryCharging: true,
		IsMobile:        false,

		BatSt:     `{"level":1.00,"status":"charging"}`,
		UIStl:     "light",
		PM:        "no",
		PEnum:     "plugin_flash^false!plugin_windows_media_player^false!plugin_adobe_acrobat^false!plugin_quicktime^false!plugin_shockwave^false!plugin_realplayer^false!plugin_vlc_player^false!plugin_devalvr^false!plugin_svg_viewer^false!plugin_java^false",
		JsoLong:   "Mac OS X 10.15.7",
		JsbLong:   "Chrome 146",
		CCD:       "1",
		DprCSV:    "2,1728,1117,1728,1079,1728,1117,1728,1117,0,0",
		DSTOffset: 60,
		WnId:      "webrtc_no_internal_data",
		FxStr:     "3456x2234",
		AfStr:     "3456x2234",
		Jfn:       142,
	}
}

func CapturedSafariMacOSProfile() *Profile {
	return &Profile{
		UA:              "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Safari/605.1.15",
		Vendor:          "Apple Computer, Inc.",
		Platform:        "MacIntel",
		Languages:       []string{"en-US", "en"},
		HardwareConc:    10,
		DeviceMemory:    8,
		ScreenW:         1728,
		ScreenH:         1117,
		AvailW:          1728,
		AvailH:          1079,
		InnerW:          1728,
		InnerH:          1015,
		OuterW:          1728,
		OuterH:          1117,
		ColorDepth:      30,
		PixelRatio:      2.0,
		TZOffset:        -300,
		TZName:          "America/New_York",
		WebGLVendor:     "Apple Inc.",
		WebGLRenderer:   "Apple GPU",
		GPU:             "Apple M2",
		Plugins:         []string{"PDF Viewer", "Chrome PDF Viewer", "Chromium PDF Viewer", "Microsoft Edge PDF Viewer", "WebKit built-in PDF"},
		Mimes:           []string{"application/pdf", "text/pdf"},
		UAFullVersion:   "17.5",
		UABrandsHigh:    ``,
		WebdriverVal:    false,
		FontList:        macFontList(),
		BatteryLevel:    1.0,
		BatteryCharging: true,
		IsMobile:        false,

		BatSt:     `{"level":1.00,"status":"charging"}`,
		UIStl:     "light",
		PM:        "no",
		PEnum:     "plugin_flash^false!plugin_windows_media_player^false!plugin_adobe_acrobat^false!plugin_quicktime^false!plugin_shockwave^false!plugin_realplayer^false!plugin_vlc_player^false!plugin_devalvr^false!plugin_svg_viewer^false!plugin_java^false",
		JsoLong:   "Mac OS X 10.15.7",
		JsbLong:   "Safari 17",
		CCD:       "",
		DprCSV:    "2,1728,1117,1728,1079,1728,1117,1728,1117,0,0",
		DSTOffset: 60,
		WnId:      "webrtc_no_internal_data",
		FxStr:     "3456x2234",
		AfStr:     "3456x2234",
		Jfn:       142,
	}
}

func macFontList() []string {
	return []string{
		"American Typewriter", "Andale Mono", "Arial", "Arial Black", "Arial Hebrew",
		"Arial Narrow", "Arial Rounded MT Bold", "Arial Unicode MS", "Avenir", "Avenir Next",
		"Avenir Next Condensed", "Baskerville", "Big Caslon", "Bodoni 72", "Bodoni 72 Oldstyle",
		"Bodoni 72 Smallcaps", "Bradley Hand", "Brush Script MT", "Chalkboard", "Chalkboard SE",
		"Chalkduster", "Cochin", "Comic Sans MS", "Copperplate", "Courier", "Courier New",
		"Didot", "DIN Alternate", "DIN Condensed", "Futura", "Geneva", "Georgia", "Gill Sans",
		"Helvetica", "Helvetica Neue", "Herculanum", "Hiragino Maru Gothic ProN",
		"Hiragino Mincho ProN", "Hiragino Sans", "Hoefler Text", "Impact", "Iowan Old Style",
		"Kannada Sangam MN", "Lao Sangam MN", "Lucida Grande", "Luminari", "Marker Felt",
		"Menlo", "Monaco", "Noteworthy", "Optima", "Palatino", "Papyrus", "Phosphate",
		"Rockwell", "Savoye LET", "SignPainter", "Skia", "Snell Roundhand", "Tahoma", "Times",
		"Times New Roman", "Trattatello", "Trebuchet MS", "Verdana", "Zapfino",
	}
}

func mobileFontList() []string {
	return []string{
		"sans-serif", "Roboto", "Noto Sans", "Droid Sans", "Droid Serif",
		"Roboto Mono", "Noto Color Emoji", "Roboto Condensed",
	}
}

func iosFontList() []string {
	return []string{
		"-apple-system", "Helvetica", "Helvetica Neue", "Avenir", "Avenir Next",
		"SF Pro", "SF Pro Display", "SF Pro Text", "Times New Roman",
		"Courier", "Courier New", "Georgia", "Verdana", "Trebuchet MS",
		"Arial", "Arial Hebrew", "Hiragino Sans", "Hiragino Mincho ProN",
		"Apple Color Emoji",
	}
}

func defaultFontList() []string {
	return []string{
		"Arial", "Arial Black", "Arial Narrow", "Calibri", "Cambria",
		"Cambria Math", "Comic Sans MS", "Consolas", "Courier", "Courier New",
		"Georgia", "Helvetica", "Impact", "Lucida Console", "Lucida Sans Unicode",
		"Microsoft Sans Serif", "MS Gothic", "MS PGothic", "MS Sans Serif", "MS Serif",
		"Palatino Linotype", "Segoe Print", "Segoe Script", "Segoe UI", "Segoe UI Light",
		"Segoe UI Semibold", "Segoe UI Symbol", "Tahoma", "Times New Roman", "Trebuchet MS",
		"Verdana", "Wingdings",
	}
}

func randomHex(n int) string {
	b := make([]byte, n/2)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func random16() string { return randomHex(32) }

func RandomNumericString(n int) string {
	const digits = "0123456789"
	b := make([]byte, n)
	rand.Read(b)
	for i := range b {
		b[i] = digits[int(b[i])%10]
	}
	return string(b)
}

func RandomLowerHex(n int) string { return randomHex(n * 2) }

func RandomUpperHex(n int) string { return strings.ToUpper(randomHex(n * 2)) }

func RandomFilename16() string { return randomHex(16) }

func (p *Profile) UserAgentBrandsList() []string {
	out := []string{}
	for _, b := range strings.Split(p.UABrandsHigh, ", ") {
		out = append(out, b)
	}
	return out
}

func mustHex(b []byte) string { return hex.EncodeToString(b) }

func ToHexBytes(s string) string { return hex.EncodeToString([]byte(s)) }

var _ = fmt.Sprint

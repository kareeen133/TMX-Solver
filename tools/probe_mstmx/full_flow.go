//go:build ignore

package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"strings"

	fhttp "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

func ruuid() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0F) | 0x40
	b[8] = (b[8] & 0x3F) | 0x80
	h := hex.EncodeToString(b)
	return fmt.Sprintf("%s-%s-%s-%s-%s", h[0:8], h[8:12], h[12:16], h[16:20], h[20:32])
}
func rhex(n int) string  { b := make([]byte, n/2); rand.Read(b); return hex.EncodeToString(b) }
func ruhex(n int) string { return strings.ToUpper(rhex(n)) }

func main() {
	jar := tls_client.NewCookieJar()
	c, _ := tls_client.NewHttpClient(tls_client.NewNoopLogger(),
		tls_client.WithClientProfile(profiles.Chrome_146_PSK),
		tls_client.WithCookieJar(jar),
		tls_client.WithTimeoutSeconds(15),
	)

	cid := "33e01921-4d64-4f8c-a055-5bdaffd5e33d"
	sid := rhex(32)
	devID := ruuid()
	wTicks := ruhex(15)

	hit := func(label, target string, hdrs map[string]string) (status int, body string, sc []string) {
		req, _ := fhttp.NewRequest("GET", target, nil)
		req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36")
		req.Header.Set("accept-language", "en-US,en;q=0.9")
		req.Header.Set("sec-ch-ua", `"Chromium";v="148", "Google Chrome";v="148", "Not?A_Brand";v="99"`)
		req.Header.Set("sec-ch-ua-mobile", "?0")
		req.Header.Set("sec-ch-ua-platform", `"Windows"`)
		for k, v := range hdrs {
			req.Header.Set(k, v)
		}
		resp, err := c.Do(req)
		if err != nil {
			fmt.Println(label, "ERR", err)
			return 0, "", nil
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(b), resp.Header.Values("set-cookie")
	}

	cookieDump := func() {
		u, _ := url.Parse("https://fpt.live.com/")
		for _, ck := range jar.Cookies(u) {
			v := ck.Value
			if len(v) > 50 {
				v = v[:50] + "..."
			}
			fmt.Printf("  %s.%s = %s\n", ck.Domain, ck.Name, v)
		}
		u2, _ := url.Parse("https://df.cfp.microsoft.com/")
		for _, ck := range jar.Cookies(u2) {
			v := ck.Value
			if len(v) > 50 {
				v = v[:50] + "..."
			}
			fmt.Printf("  %s.%s = %s\n", ck.Domain, ck.Name, v)
		}
	}

	fmt.Println("\n=== STEP 1: FPT bootstrap (gets MUID + fptctx2 cookies) ===")
	url1 := fmt.Sprintf("https://fpt.live.com/?session_id=%s&CustomerId=%s&PageId=SI", sid, cid)
	st1, body1, sc1 := hit("fpt-bootstrap", url1, map[string]string{
		"accept":         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"sec-fetch-dest": "iframe",
		"sec-fetch-mode": "navigate",
		"sec-fetch-site": "cross-site",
		"referer":        "https://signup.live.com/",
	})
	fmt.Println("  status :", st1, " body:", len(body1), "bytes")
	for _, c := range sc1 {
		if len(c) > 120 {
			c = c[:120] + "..."
		}
		fmt.Println("  cookie :", c)
	}
	cookieDump()

	fmt.Println("\n=== STEP 2: TMX (Ls1.0) ===")
	url2 := fmt.Sprintf("https://df.cfp.microsoft.com/Clear.HTML?ctx=Ls1.0&wl=False&session_id=%s&id=%s&w=%s&tkt=&CustomerId=%s",
		sid, devID, wTicks, cid)
	st2, body2, sc2 := hit("tmx-Ls1.0", url2, map[string]string{
		"accept":         "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"sec-fetch-dest": "iframe",
		"sec-fetch-mode": "navigate",
		"sec-fetch-site": "cross-site",
		"referer":        url1,
	})
	fmt.Println("  status :", st2, " body:", len(body2), "bytes")
	for _, c := range sc2 {
		if len(c) > 120 {
			c = c[:120] + "..."
		}
		fmt.Println("  cookie :", c)
	}
	cookieDump()
	containsBaseStamp := strings.Contains(body2, "BaseStamp")
	fmt.Println("  body has BaseStamp() :", containsBaseStamp)

	fmt.Println("\n=== STEP 3: Lscb1.0 returning-user callback (fid=<new MUID>&ofid=<old MUID>) ===")
	muidU, _ := url.Parse("https://fpt.live.com/")
	var muid string
	for _, ck := range jar.Cookies(muidU) {
		if ck.Name == "MUID" {
			muid = ck.Value
		}
	}
	if muid == "" {
		fmt.Println("  no MUID set — cannot fire Lscb1.0")
	} else {
		oldFid := ruhex(32)
		url3 := fmt.Sprintf("https://df.cfp.microsoft.com/Images/Clear.PNG?ctx=Lscb1.0&session_id=%s&CustomerId=%s&fid=%s&ofid=%s&w=%s&auth=",
			sid, cid, muid, oldFid, wTicks)
		st3, body3, sc3 := hit("Lscb1.0", url3, map[string]string{
			"accept":         "image/avif,image/webp,image/png,*/*;q=0.8",
			"sec-fetch-dest": "image",
			"sec-fetch-mode": "no-cors",
			"sec-fetch-site": "cross-site",
			"referer":        url1,
		})
		fmt.Println("  url    :", url3[:130], "...")
		fmt.Println("  status :", st3, " body:", len(body3), "bytes")
		for _, c := range sc3 {
			if len(c) > 120 {
				c = c[:120] + "..."
			}
			fmt.Println("  cookie :", c)
		}
	}

	fmt.Println("\n=== STEP 4: FPT pixel submission with real fingerprint blob ===")
	esiPlain := fmt.Sprintf("bua=%s&os=Win32&lproc=16&ol=true&rtt=50&chrm=true&prosub=20030107&eval=33&appv=%s&ls=true&dm=32&mtp=0&nc=82&pr=1.0000000149011612&sr=1280x720&scd=32&asr=1280x720&tz=330&dst=0&tzo=330&bl=en-US&mth=27f51d3149e6bf209b66bd387b0af3c4&mtn=2&pn=5&ph=f3ac22ac59c6dcb874109d093c5255e8&p=plugin_flash%%3Dfalse&fh=b6d6b51ef9111ad70d5c35d97303db60&fn=127&lh=%s&dr=https%%3A%%2F%%2Flogin.live.com%%2F&w=%s&id=%s&a=s&c=00000000000000000000000000000000",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36",
		"5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36",
		url.QueryEscape(url1),
		wTicks,
		ruuid(),
	)
	esiB64 := base64.StdEncoding.EncodeToString([]byte(esiPlain))
	eciPlain := `{"uvdr":"Google Inc. (Intel)","urdr":"ANGLE (Intel, Intel(R) Graphics (0x0000A7AB) Direct3D11 vs_5_0 ps_5_0, D3D11)","vdr":"WebKit","rdr":"WebKit WebGL","iduh":"3debc60dfd21aba98ca51eafe127f8e9"}`
	eciB64 := base64.StdEncoding.EncodeToString([]byte(eciPlain))

	url4 := fmt.Sprintf("https://fpt.live.com/Images/Clear.PNG?ctx=jscb1.0&session_id=%s&CustomerId=%s&esi=%s&eci=%s&PageId=SI&u1=&u3=19.0.0&u4=x86&u5=64&u2=%%28Chromium%%2C148.0.0.0%%29",
		sid, cid, url.QueryEscape(esiB64), url.QueryEscape(eciB64))
	st4, body4, sc4 := hit("fpt-pixel", url4, map[string]string{
		"accept":         "image/avif,image/webp,image/png,*/*;q=0.8",
		"sec-fetch-dest": "image",
		"sec-fetch-mode": "no-cors",
		"sec-fetch-site": "same-site",
		"referer":        url1,
	})
	fmt.Println("  url len:", len(url4))
	fmt.Println("  status :", st4, " body:", len(body4), "bytes")
	for _, c := range sc4 {
		if len(c) > 120 {
			c = c[:120] + "..."
		}
		fmt.Println("  cookie :", c)
	}
	cookieDump()

	fmt.Println("\n=== VERDICT ===")
	pass := st1 == 200 && st2 == 200 && (st4 == 200 || st4 == 204)
	if pass {
		fmt.Println("  ALL endpoints accepted. Microsoft TMX (Ls1.0) + FPT (jscb1.0) layer IS solvable in pure Go.")
	} else {
		fmt.Println("  Some endpoint failed:", st1, st2, st4)
	}
	fmt.Println("  body has BaseStamp() (Ls1.0):", containsBaseStamp)
}

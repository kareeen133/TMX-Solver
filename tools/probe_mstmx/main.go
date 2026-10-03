package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	fhttp "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

func randUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0F) | 0x40
	b[8] = (b[8] & 0x3F) | 0x80
	h := hex.EncodeToString(b)
	return fmt.Sprintf("%s-%s-%s-%s-%s", h[0:8], h[8:12], h[12:16], h[16:20], h[20:32])
}
func randHex(n int) string {
	b := make([]byte, n/2)
	rand.Read(b)
	return hex.EncodeToString(b)
}
func randUpperHex(n int) string {
	return strings.ToUpper(randHex(n))
}

func main() {
	c, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(),
		tls_client.WithClientProfile(profiles.Chrome_146_PSK),
		tls_client.WithTimeoutSeconds(15),
	)
	if err != nil {
		panic(err)
	}

	cid := "33e01921-4d64-4f8c-a055-5bdaffd5e33d"
	sid := randHex(32)
	devID := randUUID()
	nonce := randUpperHex(15)

	urls := []string{
		fmt.Sprintf("https://df.cfp.microsoft.com/Clear.HTML?ctx=Ls1.0&wl=False&session_id=%s&id=%s&w=%s&tkt=&CustomerId=%s",
			sid, devID, nonce, cid),
		fmt.Sprintf("https://fpt.live.com/?session_id=%s&CustomerId=%s&PageId=SI", sid, cid),
		fmt.Sprintf("https://fpt.live.com/?session_id=%s&CustomerId=%s&PageId=SU", sid, cid),
	}

	for _, u := range urls {
		fmt.Println("\n=== GET", u[:100], "...")
		req, _ := fhttp.NewRequest("GET", u, nil)
		req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36")
		req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
		req.Header.Set("accept-language", "en-US,en;q=0.9")
		req.Header.Set("sec-ch-ua", `"Chromium";v="148", "Google Chrome";v="148", "Not?A_Brand";v="99"`)
		req.Header.Set("sec-ch-ua-mobile", "?0")
		req.Header.Set("sec-ch-ua-platform", `"Windows"`)
		req.Header.Set("sec-fetch-dest", "iframe")
		req.Header.Set("sec-fetch-mode", "navigate")
		req.Header.Set("sec-fetch-site", "cross-site")
		req.Header.Set("referer", "https://signup.live.com/")
		resp, err := c.Do(req)
		if err != nil {
			fmt.Println("  err:", err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Println("  status :", resp.StatusCode)
		fmt.Println("  body   :", len(body), "bytes")
		fmt.Println("  type   :", resp.Header.Get("content-type"))
		for _, sc := range resp.Header.Values("set-cookie") {
			disp := sc
			if len(disp) > 120 {
				disp = disp[:120] + "..."
			}
			fmt.Println("  cookie :", disp)
		}
		preview := string(body)
		if len(preview) > 800 {
			preview = preview[:800] + "...[+" + fmt.Sprintf("%d", len(body)-800) + " bytes]"
		}
		fmt.Println("  body[:800]:")
		fmt.Println(preview)
	}
}

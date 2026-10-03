package main

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	fhttp "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

type Transport struct {
	c     tls_client.HttpClient
	jar   tls_client.CookieJar
	UA    string
	SecCH string
	Lang  string
	Plat  string
	Proxy string
}

func NewTransport(proxy string) (*Transport, error) {
	return NewTransportWithProfile(proxy, profiles.Chrome_146_PSK)
}

func NewTransportWithProfile(proxy string, tlsProfile profiles.ClientProfile) (*Transport, error) {
	jar := tls_client.NewCookieJar()
	opts := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(tlsProfile),
		tls_client.WithCookieJar(jar),
		tls_client.WithNotFollowRedirects(),
	}
	if proxy != "" {
		opts = append(opts, tls_client.WithProxyUrl(proxy))
	}
	c, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), opts...)
	if err != nil {
		return nil, err
	}
	return &Transport{
		c:     c,
		jar:   jar,
		UA:    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36",
		SecCH: `"Chromium";v="146", "Google Chrome";v="146", "Not?A_Brand";v="99"`,
		Lang:  "en-US,en;q=0.9",
		Plat:  `"Windows"`,
		Proxy: proxy,
	}, nil
}

func TLSProfileForUA(ua string) profiles.ClientProfile {
	if strings.Contains(ua, "Safari") && !strings.Contains(ua, "Chrome") && !strings.Contains(ua, "Edg") {
		if strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad") {
			return profiles.Safari_IOS_17_0
		}
		return profiles.Safari_16_0
	}
	if strings.Contains(ua, "Edg/") {
		return profiles.Chrome_133
	}
	if strings.Contains(ua, "Android") {
		return profiles.Chrome_131
	}
	return profiles.Chrome_146_PSK
}

func (t *Transport) baseHeaders() fhttp.Header {
	h := fhttp.Header{}
	h.Set("user-agent", t.UA)
	h.Set("accept-language", t.Lang)
	h.Set("sec-ch-ua", t.SecCH)
	h.Set("sec-ch-ua-mobile", "?0")
	h.Set("sec-ch-ua-platform", t.Plat)
	h["Header-Order:"] = []string{"sec-ch-ua", "sec-ch-ua-mobile", "sec-ch-ua-platform", "user-agent", "accept", "accept-language", "referer", "cookie"}
	return h
}

func (t *Transport) Get(targetURL, referer, accept string, follow bool) (status int, body []byte, headers fhttp.Header, finalURL string, err error) {
	finalURL = targetURL
	for hop := 0; hop < 8; hop++ {
		req, e := fhttp.NewRequest("GET", finalURL, nil)
		if e != nil {
			return 0, nil, nil, finalURL, e
		}
		req.Header = t.baseHeaders()
		if accept != "" {
			req.Header.Set("accept", accept)
		} else {
			req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
		}
		if referer != "" {
			req.Header.Set("referer", referer)
		}
		req.Header.Set("sec-fetch-dest", "document")
		req.Header.Set("sec-fetch-mode", "navigate")
		req.Header.Set("sec-fetch-site", "none")
		req.Header.Set("sec-fetch-user", "?1")
		req.Header.Set("upgrade-insecure-requests", "1")
		resp, e := t.c.Do(req)
		if e != nil {
			return 0, nil, nil, finalURL, e
		}
		body, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
		status = resp.StatusCode
		headers = resp.Header
		if !follow || (status != 301 && status != 302 && status != 303 && status != 307 && status != 308) {
			return
		}
		loc := resp.Header.Get("location")
		if loc == "" {
			return
		}
		ref, _ := url.Parse(finalURL)
		next, _ := url.Parse(loc)
		finalURL = ref.ResolveReference(next).String()
	}
	err = errors.New("too many redirects")
	return
}

func (t *Transport) GetRaw(targetURL, referer, accept string) (status int, body []byte, headers fhttp.Header, err error) {
	req, e := fhttp.NewRequest("GET", targetURL, nil)
	if e != nil {
		return 0, nil, nil, e
	}
	req.Header = t.baseHeaders()
	if accept == "" {
		accept = "*/*"
	}
	req.Header.Set("accept", accept)
	if referer != "" {
		req.Header.Set("referer", referer)
	}
	req.Header.Set("sec-fetch-dest", "script")
	req.Header.Set("sec-fetch-mode", "no-cors")
	req.Header.Set("sec-fetch-site", "cross-site")
	resp, e := t.c.Do(req)
	if e != nil {
		return 0, nil, nil, e
	}
	defer resp.Body.Close()
	body, _ = io.ReadAll(resp.Body)
	return resp.StatusCode, body, resp.Header, nil
}

func (t *Transport) Post(targetURL, referer, contentType string, body []byte) (int, []byte, fhttp.Header, error) {
	req, err := fhttp.NewRequest("POST", targetURL, strings.NewReader(string(body)))
	if err != nil {
		return 0, nil, nil, err
	}
	req.Header = t.baseHeaders()
	req.Header.Set("accept", "*/*")
	if referer != "" {
		req.Header.Set("referer", referer)
	}
	if contentType == "" {
		contentType = "text/plain;charset=UTF-8"
	}
	req.Header.Set("content-type", contentType)
	req.Header.Set("origin", originOf(referer))
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "cross-site")
	resp, err := t.c.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, rb, resp.Header, nil
}

func (t *Transport) GetCookie(domain, name string) string {
	u, _ := url.Parse("https://" + domain)
	for _, ck := range t.jar.Cookies(u) {
		if ck.Name == name {
			return ck.Value
		}
	}
	return ""
}

func (t *Transport) AllCookies(domain string) string {
	u, _ := url.Parse("https://" + domain)
	parts := []string{}
	for _, ck := range t.jar.Cookies(u) {
		parts = append(parts, fmt.Sprintf("%s=%s", ck.Name, ck.Value))
	}
	return strings.Join(parts, "; ")
}

func originOf(ref string) string {
	u, err := url.Parse(ref)
	if err != nil || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

func init() {
	_ = time.Now
}

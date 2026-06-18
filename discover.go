package main

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type DiscoverResult struct {
	OrgID      string
	Host       string
	Referer    string
	SourcePage string
	Hits       []string
}

var (
	tmxScriptRe  = regexp.MustCompile(`(?i)["'\s]((?:https?:)?//[a-z0-9.\-]+/fp/(?:tags\.js|mobile/conf)[^"'\s]*)["'\s]`)
	tmxHostRe    = regexp.MustCompile(`(?i)\b([a-z0-9\-]+(?:\.[a-z0-9\-]+)*\.online-metrix\.net)\b`)
	cnameHostRe  = regexp.MustCompile(`(?i)\b(drfdisvc\.[a-z0-9.\-]+|src\.ebay[a-z0-9.\-\-]*|dfme\.[a-z0-9.\-]+|tmx[a-z0-9\-.]+\.[a-z]{2,})\b`)
	tmxOrgJSONRe = regexp.MustCompile(`(?i)["']?(?:tmx[_]?org[_]?id|orgId|org_id|tmxOrg)["']?\s*[:=]\s*["']([a-z0-9_]{6,32})["']`)
	orgQueryRe   = regexp.MustCompile(`(?i)[?&]org_id=([a-z0-9_]{6,32})\b`)
)

func DiscoverFromURL(t *Transport, pageURL string, verbose bool) (*DiscoverResult, error) {
	u, err := url.Parse(pageURL)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	referer := fmt.Sprintf("%s://%s/", u.Scheme, u.Host)
	st, body, _, finalURL, err := t.Get(pageURL, referer, "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8", true)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", pageURL, err)
	}
	if st < 200 || st >= 400 {
		return nil, fmt.Errorf("fetch %s: status=%d (final=%s)", pageURL, st, finalURL)
	}
	if verbose {
		fmt.Printf("[discover] GET %s -> %d (%d bytes, final=%s)\n", pageURL, st, len(body), finalURL)
	}
	if dumpPath := os.Getenv("TMX_DISCOVER_DUMP"); dumpPath != "" {
		_ = os.WriteFile(dumpPath, body, 0o644)
	}

	res := &DiscoverResult{SourcePage: pageURL, Referer: referer}
	bodyStr := string(body)

	for _, m := range tmxScriptRe.FindAllStringSubmatch(bodyStr, -1) {
		raw := m[1]
		if strings.HasPrefix(raw, "//") {
			raw = "https:" + raw
		}
		if !strings.HasPrefix(raw, "http") {
			continue
		}
		parsed, perr := url.Parse(raw)
		if perr != nil {
			continue
		}
		host := parsed.Host
		org := parsed.Query().Get("org_id")
		if host != "" {
			res.Hits = append(res.Hits, "scriptSrc: "+raw)
			if res.Host == "" {
				res.Host = host
			}
			if res.OrgID == "" && org != "" {
				res.OrgID = org
			}
		}
	}
	if res.Host == "" {
		for _, m := range tmxHostRe.FindAllStringSubmatch(bodyStr, -1) {
			h := m[1]
			if rest, ok := strings.CutPrefix(h, "*."); ok {
				h = "h." + rest
			}
			res.Hits = append(res.Hits, "metrixHost: "+m[1])
			res.Host = h
			break
		}
	}
	if res.Host == "" {
		for _, m := range cnameHostRe.FindAllStringSubmatch(bodyStr, -1) {
			res.Hits = append(res.Hits, "cnameHost: "+m[1])
			res.Host = m[1]
			break
		}
	}
	if res.OrgID == "" {
		for _, m := range tmxOrgJSONRe.FindAllStringSubmatch(bodyStr, -1) {
			c := m[1]
			if c == "user_id" || c == "client_id" || c == "session_id" {
				continue
			}
			res.Hits = append(res.Hits, "jsonOrg: "+c)
			res.OrgID = c
			break
		}
	}
	if res.OrgID == "" {
		for _, m := range orgQueryRe.FindAllStringSubmatch(bodyStr, -1) {
			res.Hits = append(res.Hits, "queryOrg: "+m[1])
			res.OrgID = m[1]
			break
		}
	}

	if res.Host == "" {
		return nil, fmt.Errorf("no TMX endpoint found in %s body (%d bytes scanned)", pageURL, len(body))
	}
	if res.OrgID == "" {
		res.OrgID = "usllpic0"
		if verbose {
			fmt.Printf("[discover] no org_id found, falling back to usllpic0 demo org\n")
		}
	}
	if verbose {
		fmt.Printf("[discover] hits=%d host=%s org_id=%s\n", len(res.Hits), res.Host, res.OrgID)
		for _, h := range res.Hits {
			if len(h) > 150 {
				h = h[:150] + "..."
			}
			fmt.Printf("[discover]   %s\n", h)
		}
	}
	return res, nil
}

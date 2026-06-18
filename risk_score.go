package main

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

type RiskScoreObservable struct {
	SessionID           string
	OrgID               string
	Host                string
	ThxGuidStable       bool
	ThxGuidFirst        string
	ThxGuidSecond       string
	H64Status           int
	H64Accepted         bool
	RepeatChunkStatus   int
	RepeatChunkAccepted bool
	ElapsedMs           int64
	Verdict             string
	Notes               []string
}

func (s *Solver) SelfVerifyRisk(orgID, sessionID, host string, lastSubmittedURL string) *RiskScoreObservable {
	r := &RiskScoreObservable{SessionID: sessionID, OrgID: orgID, Host: host}
	t0 := time.Now()
	referer := "https://www.example.com/"

	hostURL, _ := url.Parse(fmt.Sprintf("https://%s/", host))
	readThxGuid := func() string {
		for _, ck := range s.t.jar.Cookies(hostURL) {
			if ck.Name == "thx_guid" {
				return ck.Value
			}
		}
		return ""
	}

	r.ThxGuidFirst = readThxGuid()

	url1 := fmt.Sprintf("https://%s/fp/tags.js?org_id=%s&session_id=%s", host, orgID, sessionID)
	st1, _, hdrs1, err1 := s.t.GetRaw(url1, referer, "*/*")
	if err1 != nil || st1 != 200 {
		r.Notes = append(r.Notes, fmt.Sprintf("re-tags.js failed: status=%d err=%v", st1, err1))
	} else if hdrs1 != nil {
		for _, sc := range hdrs1.Values("Set-Cookie") {
			if i := strings.Index(sc, "thx_guid="); i >= 0 {
				rest := sc[i+9:]
				if e := strings.IndexAny(rest, "; \r\n"); e >= 0 {
					rest = rest[:e]
				}
				if rest != "" {
					r.ThxGuidFirst = rest
				}
			}
		}
	}

	r.ThxGuidSecond = readThxGuid()
	if r.ThxGuidFirst != "" && r.ThxGuidFirst == r.ThxGuidSecond {
		r.ThxGuidStable = true
	}

	h64URL := fmt.Sprintf("https://h64.online-metrix.net/fp/clear.png?org_id=%s&session_id=%s&i=2", orgID, sessionID)
	st2, _, _, _ := s.t.GetRaw(h64URL, referer, "image/*")
	r.H64Status = st2
	if st2 == 200 || st2 == 204 {
		r.H64Accepted = true
	}

	if lastSubmittedURL != "" {
		st3, _, _, _ := s.t.GetRaw(lastSubmittedURL, referer, "image/*")
		r.RepeatChunkStatus = st3
		if st3 == 200 || st3 == 204 {
			r.RepeatChunkAccepted = true
		}
	}

	r.ElapsedMs = time.Since(t0).Milliseconds()

	score := 0
	if r.ThxGuidStable {
		score++
	}
	if r.H64Accepted {
		score++
	}
	if r.RepeatChunkAccepted {
		score++
	}
	switch {
	case score == 3:
		r.Verdict = "LIKELY_LOW_RISK"
	case score == 2:
		r.Verdict = "LIKELY_MEDIUM_RISK"
	default:
		r.Verdict = "REJECTED_OR_HIGH_RISK"
	}
	return r
}

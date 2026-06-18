package main

import (
	"fmt"
	"strings"
	"time"
)

type WalmartVerifyResult struct {
	PreLoginStatus  int
	PreLoginLen     int
	PreLoginBlock   bool
	PreLoginCaptcha bool
	BootstrapStatus int
	BootstrapMs     int64
	GqlStatus       int
	GqlBodyHead     string
	GqlAccepted     bool
	GqlChallenged   bool
	Notes           []string
}

func (s *Solver) VerifyWalmart(email string) *WalmartVerifyResult {
	r := &WalmartVerifyResult{}
	loginURL := "https://identity.walmart.com/account/login"
	t0 := time.Now()
	st, body, hdrs, _, err := s.t.Get(loginURL, "https://www.walmart.com/", "", true)
	r.PreLoginStatus = st
	r.PreLoginLen = len(body)
	if err != nil {
		r.Notes = append(r.Notes, "preLogin err: "+err.Error())
		return r
	}
	bs := strings.ToLower(string(body))
	if st >= 400 || strings.Contains(bs, "px-captcha") || strings.Contains(bs, "perimeter") || strings.Contains(bs, "akamai") && strings.Contains(bs, "challenge") {
		r.PreLoginBlock = true
		r.Notes = append(r.Notes, fmt.Sprintf("preLogin blocked status=%d", st))
	}
	if strings.Contains(bs, "captcha") || strings.Contains(bs, "press &amp; hold") || strings.Contains(bs, "are you a robot") {
		r.PreLoginCaptcha = true
		r.Notes = append(r.Notes, "preLogin returned captcha challenge")
	}
	_ = hdrs

	bootstrapURL := "https://drfdisvc.walmart.com/fp/tags.js?org_id=hgy2n0ks&session_id=" + RandomLowerHex(16)
	tb := time.Now()
	bst, bbody, _, _ := s.t.GetRaw(bootstrapURL, "https://identity.walmart.com/", "*/*")
	r.BootstrapStatus = bst
	r.BootstrapMs = time.Since(tb).Milliseconds()
	if bst != 200 || len(bbody) < 1000 {
		r.Notes = append(r.Notes, fmt.Sprintf("tmx bootstrap failed status=%d len=%d", bst, len(bbody)))
	}
	clearURL := "https://drfdisvc.walmart.com/fp/clear.png?org_id=hgy2n0ks&session_id=" + RandomLowerHex(16)
	s.t.GetRaw(clearURL, "https://identity.walmart.com/", "image/*")

	gqlURL := "https://identity.walmart.com/orchestra/idp/graphql"
	gqlBody := fmt.Sprintf(`{"query":"query checkAccount($input: AccountInput!) { account(input: $input) { exists status } }","variables":{"input":{"identifier":"%s","type":"email"}}}`, email)
	tg := time.Now()
	gst, gbody, _, gerr := s.t.Post(gqlURL, "https://identity.walmart.com/account/login", "application/json", []byte(gqlBody))
	_ = tg
	r.GqlStatus = gst
	if gerr != nil {
		r.Notes = append(r.Notes, "gql err: "+gerr.Error())
	}
	gs := string(gbody)
	if len(gs) > 220 {
		r.GqlBodyHead = gs[:220]
	} else {
		r.GqlBodyHead = gs
	}
	gsLow := strings.ToLower(gs)
	if gst >= 200 && gst < 300 && (strings.Contains(gsLow, "exists") || strings.Contains(gsLow, "data") || strings.Contains(gsLow, "account")) {
		r.GqlAccepted = true
	}
	if gst == 403 || strings.Contains(gsLow, "captcha") || strings.Contains(gsLow, "challenge") || strings.Contains(gsLow, "blocked") || strings.Contains(gsLow, "perimeter") {
		r.GqlChallenged = true
	}
	r.Notes = append(r.Notes, fmt.Sprintf("preLogin %dB, gql %d %dB total", r.PreLoginLen, r.GqlStatus, len(gbody)))
	_ = t0
	return r
}

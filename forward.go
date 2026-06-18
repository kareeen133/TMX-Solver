package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	fhttp "github.com/bogdanfinn/fhttp"
)

type ForwardRequest struct {
	OrgID         string            `json:"org_id"`
	Host          string            `json:"host"`
	Referer       string            `json:"referer"`
	Target        string            `json:"target"`
	Method        string            `json:"method"`
	Headers       map[string]string `json:"headers"`
	Cookies       map[string]string `json:"cookies"`
	Body          string            `json:"body"`
	Deep          bool              `json:"deep"`
	FollowRedir   bool              `json:"follow_redirects"`
	IncludeCookie bool              `json:"include_tmx_cookies"`
}

type ForwardResponse struct {
	OK            bool              `json:"ok"`
	Error         string            `json:"error,omitempty"`
	SessionID     string            `json:"session_id"`
	OrgID         string            `json:"org_id"`
	Host          string            `json:"host"`
	ThxGuid       string            `json:"thx_guid,omitempty"`
	TmxGuid       string            `json:"tmx_guid,omitempty"`
	TmxNonce      string            `json:"tmx_nonce,omitempty"`
	TmxAccepted   bool              `json:"tmx_accepted"`
	SolveMs       int64             `json:"solve_ms"`
	UpstreamMs    int64             `json:"upstream_ms"`
	Status        int               `json:"status"`
	ResponseHdrs  map[string]string `json:"response_headers"`
	Body          string            `json:"body"`
	BodyLen       int               `json:"body_len"`
	FinalURL      string            `json:"final_url,omitempty"`
	UsedUA        string            `json:"used_ua"`
	UsedProfile   string            `json:"used_profile"`
}

func (s *Solver) Forward(req *ForwardRequest) *ForwardResponse {
	resp := &ForwardResponse{
		OrgID:        req.OrgID,
		Host:         req.Host,
		UsedUA:       s.t.UA,
		UsedProfile:  profileLabel(s.profile),
		ResponseHdrs: map[string]string{},
	}
	if req.Target == "" {
		resp.Error = "target is required"
		return resp
	}
	if req.OrgID == "" || req.Host == "" {
		resp.Error = "org_id and host are required"
		return resp
	}
	if req.Referer == "" {
		req.Referer = req.Target
	}

	t0 := time.Now()
	var sr *SolveResult
	var err error
	if req.Deep {
		sr, err = s.SolveDeep(req.OrgID, req.Host, "", req.Referer, false)
	} else {
		sr, err = s.Solve(req.OrgID, req.Host, req.Referer)
	}
	resp.SolveMs = time.Since(t0).Milliseconds()
	if err != nil {
		resp.Error = "solve failed: " + err.Error()
		return resp
	}
	resp.SessionID = sr.SessionID
	resp.ThxGuid = sr.ThxGuid
	resp.TmxGuid = sr.TmxGuid
	resp.TmxNonce = sr.TmxNonce
	resp.TmxAccepted = sr.TmxAccepted

	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = "GET"
	}

	tu := time.Now()
	var (
		status   int
		body     []byte
		hdrs     fhttp.Header
		finalURL string
		ferr     error
	)
	switch method {
	case "GET":
		status, body, hdrs, finalURL, ferr = s.doForwardGet(req)
	case "POST", "PUT", "PATCH", "DELETE":
		status, body, hdrs, ferr = s.doForwardWithBody(method, req)
		finalURL = req.Target
	default:
		resp.Error = "unsupported method: " + method
		return resp
	}
	resp.UpstreamMs = time.Since(tu).Milliseconds()
	if ferr != nil {
		resp.Error = "upstream error: " + ferr.Error()
		return resp
	}
	resp.Status = status
	resp.FinalURL = finalURL
	resp.BodyLen = len(body)
	resp.Body = string(body)
	for k, v := range hdrs {
		if len(v) > 0 {
			resp.ResponseHdrs[k] = v[0]
		}
	}
	resp.OK = status >= 200 && status < 400
	return resp
}

func (s *Solver) doForwardGet(req *ForwardRequest) (int, []byte, fhttp.Header, string, error) {
	r, err := fhttp.NewRequest("GET", req.Target, nil)
	if err != nil {
		return 0, nil, nil, "", err
	}
	s.applyForwardHeaders(r, req)
	resp, err := s.t.c.Do(r)
	if err != nil {
		return 0, nil, nil, req.Target, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, body, resp.Header, req.Target, nil
}

func (s *Solver) doForwardWithBody(method string, req *ForwardRequest) (int, []byte, fhttp.Header, error) {
	r, err := fhttp.NewRequest(method, req.Target, strings.NewReader(req.Body))
	if err != nil {
		return 0, nil, nil, err
	}
	s.applyForwardHeaders(r, req)
	if r.Header.Get("content-type") == "" && req.Body != "" {
		r.Header.Set("content-type", "application/json")
	}
	r.Header.Set("origin", originOf(req.Referer))
	resp, err := s.t.c.Do(r)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, body, resp.Header, nil
}

func (s *Solver) applyForwardHeaders(r *fhttp.Request, req *ForwardRequest) {
	r.Header = s.t.baseHeaders()
	r.Header.Set("accept", "*/*")
	if req.Referer != "" {
		r.Header.Set("referer", req.Referer)
	}
	r.Header.Set("sec-fetch-dest", "empty")
	r.Header.Set("sec-fetch-mode", "cors")
	r.Header.Set("sec-fetch-site", "same-origin")
	for k, v := range req.Headers {
		r.Header.Set(k, v)
	}
	if len(req.Cookies) > 0 {
		parts := []string{}
		for k, v := range req.Cookies {
			parts = append(parts, fmt.Sprintf("%s=%s", k, v))
		}
		existing := r.Header.Get("cookie")
		if existing != "" {
			r.Header.Set("cookie", existing+"; "+strings.Join(parts, "; "))
		} else {
			r.Header.Set("cookie", strings.Join(parts, "; "))
		}
	}
}

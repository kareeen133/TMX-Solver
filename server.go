package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type APIRecord struct {
	ID          string            `json:"id"`
	Time        string            `json:"time"`
	UnixMs      int64             `json:"unix_ms"`
	ClientIP    string            `json:"client_ip"`
	Endpoint    string            `json:"endpoint"`
	OrgID       string            `json:"org_id"`
	Host        string            `json:"host"`
	Target      string            `json:"target"`
	Method      string            `json:"method"`
	Deep        bool              `json:"deep"`
	SessionID   string            `json:"session_id"`
	TmxAccepted bool              `json:"tmx_accepted"`
	Status      int               `json:"status"`
	OK          bool              `json:"ok"`
	SolveMs     int64             `json:"solve_ms"`
	UpstreamMs  int64             `json:"upstream_ms"`
	BodyLen     int               `json:"body_len"`
	Error       string            `json:"error,omitempty"`
	ReqHeaders  map[string]string `json:"req_headers,omitempty"`
	ReqBody     string            `json:"req_body,omitempty"`
	Response    string            `json:"response_head,omitempty"`
}

type Store struct {
	mu      sync.RWMutex
	records []*APIRecord
	max     int
	logPath string
	logFile *os.File
}

func NewStore(logPath string, max int) (*Store, error) {
	st := &Store{max: max, logPath: logPath}
	if logPath != "" {
		if data, err := os.Open(logPath); err == nil {
			sc := bufio.NewScanner(data)
			sc.Buffer(make([]byte, 1024*1024), 8*1024*1024)
			for sc.Scan() {
				var r APIRecord
				if json.Unmarshal(sc.Bytes(), &r) == nil {
					st.records = append(st.records, &r)
				}
			}
			data.Close()
			if len(st.records) > max {
				st.records = st.records[len(st.records)-max:]
			}
		}
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}
		st.logFile = f
	}
	return st, nil
}

func (s *Store) Add(r *APIRecord) {
	s.mu.Lock()
	s.records = append(s.records, r)
	if len(s.records) > s.max {
		s.records = s.records[len(s.records)-s.max:]
	}
	s.mu.Unlock()
	if s.logFile != nil {
		if b, err := json.Marshal(r); err == nil {
			s.logFile.Write(b)
			s.logFile.Write([]byte("\n"))
		}
	}
}

func (s *Store) List(limit int, offset int, q string) ([]*APIRecord, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	filtered := s.records
	if q != "" {
		ql := strings.ToLower(q)
		filtered = filtered[:0:0]
		for _, r := range s.records {
			hay := strings.ToLower(r.Target + " " + r.OrgID + " " + r.Host + " " + r.SessionID + " " + r.ClientIP + " " + r.Error)
			if strings.Contains(hay, ql) {
				filtered = append(filtered, r)
			}
		}
	}
	total := len(filtered)
	out := make([]*APIRecord, 0, limit)
	for i := total - 1 - offset; i >= 0 && len(out) < limit; i-- {
		out = append(out, filtered[i])
	}
	return out, total
}

func (s *Store) Clear() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := len(s.records)
	if s.logFile != nil && s.logPath != "" {
		s.logFile.Close()
		f, err := os.OpenFile(s.logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			s.logFile, _ = os.OpenFile(s.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			return n, err
		}
		s.logFile = f
	}
	s.records = nil
	return n, nil
}

func (s *Store) Get(id string) *APIRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.records {
		if r.ID == id {
			return r
		}
	}
	return nil
}

func (s *Store) Stats() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	total := len(s.records)
	ok, fail, accepted := 0, 0, 0
	var sumSolve, sumUp int64
	last24 := 0
	cutoff := time.Now().Add(-24 * time.Hour).UnixMilli()
	for _, r := range s.records {
		if r.OK {
			ok++
		} else {
			fail++
		}
		if r.TmxAccepted {
			accepted++
		}
		sumSolve += r.SolveMs
		sumUp += r.UpstreamMs
		if r.UnixMs >= cutoff {
			last24++
		}
	}
	avgSolve, avgUp := int64(0), int64(0)
	if total > 0 {
		avgSolve = sumSolve / int64(total)
		avgUp = sumUp / int64(total)
	}
	return map[string]any{
		"total":         total,
		"ok":            ok,
		"fail":          fail,
		"tmx_accepted":  accepted,
		"avg_solve_ms":  avgSolve,
		"avg_upstream_ms": avgUp,
		"last_24h":      last24,
	}
}

type APIServer struct {
	addr     string
	apiKey   string
	proxy    string
	store    *Store
	adminUser string
	adminPass string
	keys      *KeyStore
	sessions  map[string]int64
	sessMu    sync.Mutex
}

func NewAPIServer(addr, apiKey, proxy, adminUser, adminPass string, store *Store, keys *KeyStore) *APIServer {
	srv := &APIServer{
		addr: addr, apiKey: apiKey, proxy: proxy, store: store,
		adminUser: adminUser, adminPass: adminPass, keys: keys,
		sessions: map[string]int64{},
	}
	return srv
}

func (a *APIServer) newSession() string {
	tok := newID() + newID()
	a.sessMu.Lock()
	a.sessions[tok] = time.Now().Add(24 * time.Hour).UnixMilli()
	a.sessMu.Unlock()
	return tok
}

func (a *APIServer) checkSession(tok string) bool {
	if tok == "" {
		return false
	}
	a.sessMu.Lock()
	defer a.sessMu.Unlock()
	exp, ok := a.sessions[tok]
	if !ok {
		return false
	}
	if time.Now().UnixMilli() > exp {
		delete(a.sessions, tok)
		return false
	}
	return true
}

func (a *APIServer) revokeSession(tok string) {
	a.sessMu.Lock()
	delete(a.sessions, tok)
	a.sessMu.Unlock()
}

func (a *APIServer) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", a.handleHealth)
	mux.HandleFunc("/v1/forward", a.requireKey(a.handleForward))
	mux.HandleFunc("/v1/solve", a.requireKey(a.handleSolve))
	mux.HandleFunc("/admin", a.handleAdminPage)
	mux.HandleFunc("/admin/", a.handleAdminPage)
	mux.HandleFunc("/admin/login", a.handleAdminLogin)
	mux.HandleFunc("/admin/logout", a.handleAdminLogout)
	mux.HandleFunc("/admin/api/requests", a.requireAdminSession(a.handleAdminList))
	mux.HandleFunc("/admin/api/request", a.requireAdminSession(a.handleAdminGet))
	mux.HandleFunc("/admin/api/stats", a.requireAdminSession(a.handleAdminStats))
	mux.HandleFunc("/admin/api/clear", a.requireAdminSession(a.handleAdminClear))
	mux.HandleFunc("/admin/api/keys", a.requireAdminSession(a.handleKeysList))
	mux.HandleFunc("/admin/api/keys/create", a.requireAdminSession(a.handleKeysCreate))
	mux.HandleFunc("/admin/api/keys/adjust", a.requireAdminSession(a.handleKeysAdjust))
	mux.HandleFunc("/admin/api/keys/toggle", a.requireAdminSession(a.handleKeysToggle))
	mux.HandleFunc("/admin/api/keys/delete", a.requireAdminSession(a.handleKeysDelete))
	mux.HandleFunc("/v1/balance", a.handleKeyBalance)
	mux.HandleFunc("/v1/ip", a.requireKey(a.handleIP))
	mux.HandleFunc("/docs", a.handleDocsPage)
	mux.HandleFunc("/", a.handleRoot)

	fmt.Printf("[api] listening on %s\n", a.addr)
	fmt.Printf("[api] docs:         http://%s/docs\n", a.addr)
	fmt.Printf("[api] admin login:  http://%s/admin   (user: %s)\n", a.addr, a.adminUser)
	fmt.Printf("[api] forward URL:  http://%s/v1/forward (X-API-Key: %s)\n", a.addr, a.apiKey)
	return http.ListenAndServe(a.addr, mux)
}

type ctxKey string

const reqKeyCtx ctxKey = "apikey"

func (a *APIServer) requireKey(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")
		if key == "" {
			key = r.URL.Query().Get("key")
		}
		if a.apiKey != "" && key == a.apiKey {
			h(w, r)
			return
		}
		if err := a.keys.Consume(key); err != nil {
			code := 401
			if err == errNoCredits {
				code = 402
			}
			writeJSON(w, code, map[string]any{"error": err.Error()})
			return
		}
		r.Header.Set("X-Resolved-Key", key)
		h(w, r)
	}
}

func (a *APIServer) requireAdminSession(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("tmx_admin")
		if err != nil || !a.checkSession(c.Value) {
			writeJSON(w, 401, map[string]any{"error": "not logged in"})
			return
		}
		h(w, r)
	}
}

func (a *APIServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusFound)
}

func (a *APIServer) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeJSON(w, 405, map[string]any{"error": "POST required"})
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, 400, map[string]any{"error": "invalid json"})
		return
	}
	if body.Username != a.adminUser || body.Password != a.adminPass {
		writeJSON(w, 401, map[string]any{"error": "invalid credentials"})
		return
	}
	tok := a.newSession()
	http.SetCookie(w, &http.Cookie{
		Name:     "tmx_admin",
		Value:    tok,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})
	writeJSON(w, 200, map[string]any{"ok": true})
}

func (a *APIServer) handleAdminLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("tmx_admin"); err == nil {
		a.revokeSession(c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "tmx_admin", Value: "", Path: "/", MaxAge: -1})
	writeJSON(w, 200, map[string]any{"ok": true})
}

func (a *APIServer) handleDocsPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(docsHTML))
}

func (a *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"ok": true, "ts": time.Now().Unix()})
}

func (a *APIServer) handleIP(w http.ResponseWriter, r *http.Request) {
	proxy := r.URL.Query().Get("proxy")
	if proxy == "" {
		proxy = a.proxy
	}
	prof := pickProfile("", false)
	solver, err := NewSolver(proxy, false, prof)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	status, body, _, _, ferr := solver.t.Get("https://api.ipify.org?format=json", "", "*/*", false)
	if ferr != nil {
		writeJSON(w, 500, map[string]any{"error": ferr.Error(), "proxy": proxy})
		return
	}
	writeJSON(w, 200, map[string]any{"status": status, "body": string(body), "proxy": proxy})
}

type forwardEnvelope struct {
	Profile string          `json:"profile"`
	Proxy   string          `json:"proxy"`
	Request *ForwardRequest `json:"request"`
}

func (a *APIServer) handleForward(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeJSON(w, 405, map[string]any{"error": "POST required"})
		return
	}
	var env forwardEnvelope
	if err := json.NewDecoder(r.Body).Decode(&env); err != nil {
		writeJSON(w, 400, map[string]any{"error": "invalid json: " + err.Error()})
		return
	}
	if env.Request == nil {
		writeJSON(w, 400, map[string]any{"error": "missing 'request' field"})
		return
	}
	if env.Request.Method == "" {
		env.Request.Method = "GET"
	}

	proxy := env.Proxy
	if proxy == "" {
		proxy = a.proxy
	}
	prof := pickProfile(env.Profile, false)
	solver, err := NewSolver(proxy, false, prof)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": "solver init: " + err.Error()})
		return
	}

	resp := solver.Forward(env.Request)
	rec := &APIRecord{
		ID:          newID(),
		Time:        time.Now().UTC().Format(time.RFC3339),
		UnixMs:      time.Now().UnixMilli(),
		ClientIP:    clientIP(r),
		Endpoint:    "/v1/forward",
		OrgID:       env.Request.OrgID,
		Host:        env.Request.Host,
		Target:      env.Request.Target,
		Method:      env.Request.Method,
		Deep:        env.Request.Deep,
		SessionID:   resp.SessionID,
		TmxAccepted: resp.TmxAccepted,
		Status:      resp.Status,
		OK:          resp.OK,
		SolveMs:     resp.SolveMs,
		UpstreamMs:  resp.UpstreamMs,
		BodyLen:     resp.BodyLen,
		Error:       resp.Error,
		ReqHeaders:  env.Request.Headers,
		ReqBody:     truncate(env.Request.Body, 4000),
		Response:    truncate(resp.Body, 4000),
	}
	a.store.Add(rec)
	writeJSON(w, 200, resp)
}

type solveRequest struct {
	OrgID   string `json:"org_id"`
	Host    string `json:"host"`
	Referer string `json:"referer"`
	Profile string `json:"profile"`
	Proxy   string `json:"proxy"`
	Deep    bool   `json:"deep"`
}

func (a *APIServer) handleSolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeJSON(w, 405, map[string]any{"error": "POST required"})
		return
	}
	var req solveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "invalid json: " + err.Error()})
		return
	}
	if req.OrgID == "" || req.Host == "" {
		writeJSON(w, 400, map[string]any{"error": "org_id and host required"})
		return
	}
	if req.Referer == "" {
		req.Referer = "https://" + req.Host + "/"
	}

	solveProxy := req.Proxy
	if solveProxy == "" {
		solveProxy = a.proxy
	}
	solvProf := pickProfile(req.Profile, false)
	solver, err := NewSolver(solveProxy, false, solvProf)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": "solver init: " + err.Error()})
		return
	}

	t0 := time.Now()
	var sr *SolveResult
	if req.Deep {
		sr, err = solver.SolveDeep(req.OrgID, req.Host, "", req.Referer, false)
	} else {
		sr, err = solver.Solve(req.OrgID, req.Host, req.Referer)
	}
	dur := time.Since(t0).Milliseconds()
	out := map[string]any{
		"ok":           err == nil && sr != nil && sr.TmxAccepted,
		"session_id":   "",
		"org_id":       req.OrgID,
		"host":         req.Host,
		"thx_guid":     "",
		"tmx_guid":     "",
		"tmx_nonce":    "",
		"tmx_accepted": false,
		"solve_ms":     dur,
		"profile":      profileLabel(solver.profile),
		"used_ua":      solver.t.UA,
	}
	if err != nil {
		out["error"] = err.Error()
	}
	if sr != nil {
		out["session_id"] = sr.SessionID
		out["thx_guid"] = sr.ThxGuid
		out["tmx_guid"] = sr.TmxGuid
		out["tmx_nonce"] = sr.TmxNonce
		out["tmx_accepted"] = sr.TmxAccepted
	}
	rec := &APIRecord{
		ID:          newID(),
		Time:        time.Now().UTC().Format(time.RFC3339),
		UnixMs:      time.Now().UnixMilli(),
		ClientIP:    clientIP(r),
		Endpoint:    "/v1/solve",
		OrgID:       req.OrgID,
		Host:        req.Host,
		Method:      "SOLVE",
		Deep:        req.Deep,
		SolveMs:     dur,
		OK:          err == nil,
	}
	if sr != nil {
		rec.SessionID = sr.SessionID
		rec.TmxAccepted = sr.TmxAccepted
	}
	if err != nil {
		rec.Error = err.Error()
	}
	if (err != nil || sr == nil || !sr.TmxAccepted) {
		if rk := r.Header.Get("X-Resolved-Key"); rk != "" {
			a.keys.Refund(rk)
		}
	}
	a.store.Add(rec)
	writeJSON(w, 200, out)
}

func (a *APIServer) handleKeysList(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"keys": a.keys.List()})
}

func (a *APIServer) handleKeysCreate(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Label   string `json:"label"`
		Credits int64  `json:"credits"`
	}
	json.NewDecoder(r.Body).Decode(&b)
	k := a.keys.Create(b.Label, b.Credits)
	writeJSON(w, 200, k)
}

func (a *APIServer) handleKeysAdjust(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Key   string `json:"key"`
		Delta int64  `json:"delta"`
	}
	if json.NewDecoder(r.Body).Decode(&b) != nil || b.Key == "" {
		writeJSON(w, 400, map[string]any{"error": "key and delta required"})
		return
	}
	k, err := a.keys.Adjust(b.Key, b.Delta)
	if err != nil {
		writeJSON(w, 404, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, k)
}

func (a *APIServer) handleKeysToggle(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Key      string `json:"key"`
		Disabled bool   `json:"disabled"`
	}
	if json.NewDecoder(r.Body).Decode(&b) != nil || b.Key == "" {
		writeJSON(w, 400, map[string]any{"error": "key required"})
		return
	}
	k, err := a.keys.SetDisabled(b.Key, b.Disabled)
	if err != nil {
		writeJSON(w, 404, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, k)
}

func (a *APIServer) handleKeysDelete(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Key string `json:"key"`
	}
	json.NewDecoder(r.Body).Decode(&b)
	if !a.keys.Delete(b.Key) {
		writeJSON(w, 404, map[string]any{"error": "not found"})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

func (a *APIServer) handleKeyBalance(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("X-API-Key")
	if key == "" {
		key = r.URL.Query().Get("key")
	}
	k, ok := a.keys.Get(key)
	if !ok {
		writeJSON(w, 404, map[string]any{"error": "invalid api key"})
		return
	}
	writeJSON(w, 200, map[string]any{"label": k.Label, "credits": k.Credits, "used": k.Used, "disabled": k.Disabled})
}

func (a *APIServer) handleAdminPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(adminHTML))
}

func (a *APIServer) handleAdminList(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	q := r.URL.Query().Get("q")
	items, total := a.store.List(limit, offset, q)
	writeJSON(w, 200, map[string]any{"items": items, "total": total, "limit": limit, "offset": offset})
}

func (a *APIServer) handleAdminGet(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	rec := a.store.Get(id)
	if rec == nil {
		writeJSON(w, 404, map[string]any{"error": "not found"})
		return
	}
	writeJSON(w, 200, rec)
}

func (a *APIServer) handleAdminStats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, a.store.Stats())
}

func (a *APIServer) handleAdminClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeJSON(w, 405, map[string]any{"error": "POST required"})
		return
	}
	n, err := a.store.Clear()
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "cleared": n})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.Encode(v)
}

func newID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func clientIP(r *http.Request) string {
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		if i := strings.Index(v, ","); i >= 0 {
			return strings.TrimSpace(v[:i])
		}
		return strings.TrimSpace(v)
	}
	if v := r.Header.Get("X-Real-IP"); v != "" {
		return v
	}
	host := r.RemoteAddr
	if i := strings.LastIndex(host, ":"); i >= 0 {
		host = host[:i]
	}
	return host
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...[truncated]"
}

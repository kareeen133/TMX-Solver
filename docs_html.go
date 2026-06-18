package main

const docsHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>TMX Solver — API Docs</title>
<style>
:root{
  --bg:#0b0d12; --panel:#11141b; --panel2:#161a23; --line:#222838;
  --text:#e6e8ee; --muted:#8b93a7; --accent:#6ea8ff; --good:#4ade80; --bad:#f87171; --warn:#fbbf24;
  --mono:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;
}
*{box-sizing:border-box}
html,body{margin:0;padding:0;background:var(--bg);color:var(--text);font:15px/1.65 system-ui,-apple-system,Segoe UI,Roboto,sans-serif}
a{color:var(--accent);text-decoration:none}
a:hover{text-decoration:underline}
header{display:flex;align-items:center;gap:16px;padding:14px 22px;border-bottom:1px solid var(--line);background:var(--panel);position:sticky;top:0;z-index:5}
header h1{font-size:16px;margin:0;font-weight:600}
header .spacer{flex:1}
header a.btn{background:var(--panel2);color:var(--text);border:1px solid var(--line);padding:6px 12px;border-radius:6px;font-size:13px}
header a.btn:hover{border-color:var(--accent);color:var(--accent)}

.layout{display:grid;grid-template-columns:240px 1fr;max-width:1200px;margin:0 auto;gap:32px;padding:30px 22px}
nav.toc{position:sticky;top:70px;height:fit-content;border-right:1px solid var(--line);padding-right:18px}
nav.toc h3{font-size:11px;text-transform:uppercase;color:var(--muted);letter-spacing:.6px;margin:14px 0 8px}
nav.toc a{display:block;padding:5px 0;color:var(--muted);font-size:13px}
nav.toc a:hover{color:var(--accent);text-decoration:none}

main{min-width:0}
main h1{font-size:28px;margin:0 0 6px;letter-spacing:-.3px}
main .lead{color:var(--muted);font-size:15px;margin-bottom:30px}
main h2{font-size:20px;margin:36px 0 12px;padding-bottom:6px;border-bottom:1px solid var(--line)}
main h3{font-size:16px;margin:24px 0 8px;color:var(--accent)}
main p{margin:8px 0 14px}
main ul{margin:6px 0 14px;padding-left:22px}
main li{margin:4px 0}
main code{background:var(--panel2);border:1px solid var(--line);padding:1px 6px;border-radius:4px;font:13px var(--mono);color:#e8edf6}
main pre{background:var(--panel);border:1px solid var(--line);border-radius:8px;padding:14px 16px;overflow:auto;font:12.5px/1.55 var(--mono);color:#cdd3e0;margin:8px 0 16px}
main pre code{background:transparent;border:none;padding:0;color:inherit;font-size:inherit}
.endpoint{background:var(--panel);border:1px solid var(--line);border-radius:10px;padding:18px;margin:12px 0}
.endpoint .head{display:flex;align-items:center;gap:10px;margin-bottom:10px}
.method{font:11px var(--mono);font-weight:700;padding:3px 8px;border-radius:5px;letter-spacing:.5px}
.method.post{background:rgba(74,222,128,.15);color:var(--good)}
.method.get{background:rgba(110,168,255,.15);color:var(--accent)}
.url{font:13px var(--mono);color:var(--text)}
.tag{font-size:11px;padding:2px 8px;border-radius:999px;background:var(--panel2);color:var(--muted);border:1px solid var(--line)}
table.fields{width:100%;border-collapse:collapse;margin:8px 0 14px;font-size:13px}
table.fields th,table.fields td{padding:8px 10px;border-bottom:1px solid var(--line);text-align:left;vertical-align:top}
table.fields th{font-size:11px;text-transform:uppercase;color:var(--muted);letter-spacing:.5px;font-weight:500}
table.fields td:first-child{font-family:var(--mono);color:var(--accent);font-size:12.5px;white-space:nowrap}
table.fields td:nth-child(2){font-family:var(--mono);color:var(--warn);font-size:12px;white-space:nowrap}
.note{background:rgba(110,168,255,.08);border-left:3px solid var(--accent);padding:10px 14px;margin:14px 0;border-radius:4px;font-size:13.5px}
.warn{background:rgba(251,191,36,.08);border-left:3px solid var(--warn);padding:10px 14px;margin:14px 0;border-radius:4px;font-size:13.5px}
.danger{background:rgba(248,113,113,.08);border-left:3px solid var(--bad);padding:10px 14px;margin:14px 0;border-radius:4px;font-size:13.5px}
@media(max-width:900px){.layout{grid-template-columns:1fr}nav.toc{position:static;border-right:none;border-bottom:1px solid var(--line);padding:0 0 12px}}
</style>
</head>
<body>
<header>
  <h1>TMX Solver — API Docs</h1>
  <span class="spacer"></span>
  <a class="btn" href="/admin">Admin Panel</a>
</header>

<div class="layout">
<nav class="toc">
  <h3>Getting started</h3>
  <a href="#overview">Overview</a>
  <a href="#auth">Authentication</a>
  <a href="#quickstart">Quick start</a>

  <h3>Endpoints</h3>
  <a href="#forward">POST /v1/forward</a>
  <a href="#solve">POST /v1/solve</a>
  <a href="#health">GET /healthz</a>

  <h3>Reference</h3>
  <a href="#profiles">Profiles</a>
  <a href="#errors">Error responses</a>
  <a href="#admin">Admin panel</a>
  <a href="#why">Why solver-as-proxy?</a>
</nav>

<main>

<h1>TMX Solver API</h1>
<p class="lead">Solve ThreatMetrix fingerprinting on any TMX-protected site, then fire the gated request through the same TLS connection so Akamai and TMX both see one consistent client.</p>

<h2 id="overview">Overview</h2>
<p>This service runs a pure-Go TMX solver behind an HTTP API. It exposes two main endpoints:</p>
<ul>
  <li><code>POST /v1/forward</code> — solver-as-proxy. You give it a target URL, headers, and body; it solves TMX, attaches the cookies, and fires the request through the same TLS+IP. <strong>This is what most clients should use.</strong></li>
  <li><code>POST /v1/solve</code> — returns the raw TMX session credentials (session_id, thx_guid, tmx_guid, etc.) so you can replay them yourself. Only useful if your client can match the solver's TLS fingerprint, IP, and User-Agent.</li>
</ul>

<h2 id="auth">Authentication</h2>
<p>All <code>/v1/*</code> endpoints require an API key, sent as either:</p>
<ul>
  <li><code>X-API-Key: YOUR_KEY</code> header (preferred)</li>
  <li><code>?key=YOUR_KEY</code> query parameter</li>
</ul>
<p>The admin panel at <a href="/admin">/admin</a> uses a separate username/password login (cookie-based session).</p>

<h2 id="quickstart">Quick start</h2>
<p>Solve TMX and fire a Walmart GraphQL request in one call:</p>
<pre><code>curl -X POST https://your-host/v1/forward \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "profile": "edge-windows",
    "request": {
      "org_id": "hgy2n0ks",
      "host": "drfdisvc.walmart.com",
      "referer": "https://identity.walmart.com/account/login",
      "target": "https://identity.walmart.com/orchestra/idp/graphql",
      "method": "POST",
      "headers": {
        "x-acf-sensor-data": "AKAMAI_SENSOR_HERE",
        "content-type": "application/json"
      },
      "body": "{\"query\":\"...\",\"variables\":{...}}"
    }
  }'</code></pre>

<h2 id="forward">POST /v1/forward</h2>
<div class="endpoint">
  <div class="head"><span class="method post">POST</span><span class="url">/v1/forward</span><span class="tag">requires API key</span></div>
  <p>Solves a fresh TMX session and fires your gated request through the same TLS connection and IP. Returns the upstream response.</p>

  <h3>Request body</h3>
<pre><code>{
  "profile": "edge-windows",
  "request": {
    "org_id":   "hgy2n0ks",
    "host":     "drfdisvc.walmart.com",
    "referer":  "https://identity.walmart.com/account/login",
    "target":   "https://identity.walmart.com/orchestra/idp/graphql",
    "method":   "POST",
    "headers":  { "x-acf-sensor-data": "..." },
    "cookies":  { "extra_cookie": "value" },
    "body":     "{...json...}",
    "deep":     false
  }
}</code></pre>

  <h3>Fields</h3>
  <table class="fields">
    <tr><th>Field</th><th>Type</th><th>Description</th></tr>
    <tr><td>profile</td><td>string</td><td>Browser profile: <code>edge-windows</code>, <code>chrome-windows</code>, <code>chrome-android</code>, <code>safari-ios</code>. Default: <code>edge-windows</code>.</td></tr>
    <tr><td>request.org_id</td><td>string</td><td><strong>Required.</strong> TMX org_id (e.g. <code>hgy2n0ks</code> for Walmart).</td></tr>
    <tr><td>request.host</td><td>string</td><td><strong>Required.</strong> TMX host (e.g. <code>drfdisvc.walmart.com</code> or <code>h.online-metrix.net</code>).</td></tr>
    <tr><td>request.referer</td><td>string</td><td>Referer header for both the TMX bootstrap and the gated request. Defaults to <code>target</code>.</td></tr>
    <tr><td>request.target</td><td>string</td><td><strong>Required.</strong> The gated URL to call after solving.</td></tr>
    <tr><td>request.method</td><td>string</td><td>HTTP method: <code>GET</code>, <code>POST</code>, <code>PUT</code>, <code>PATCH</code>, <code>DELETE</code>. Default: <code>GET</code>.</td></tr>
    <tr><td>request.headers</td><td>object</td><td>Extra headers to send on the gated request (akamai sensor, auth tokens, etc.).</td></tr>
    <tr><td>request.cookies</td><td>object</td><td>Extra cookies to attach in addition to TMX's own.</td></tr>
    <tr><td>request.body</td><td>string</td><td>Request body for non-GET methods.</td></tr>
    <tr><td>request.deep</td><td>bool</td><td>Run the full chunked fingerprint pipeline (slower, more thorough). Default: <code>false</code>.</td></tr>
  </table>

  <h3>Response</h3>
<pre><code>{
  "ok": true,
  "session_id": "e78f0f81e4feaf30bd107fd5f6a247ab",
  "org_id": "hgy2n0ks",
  "host": "drfdisvc.walmart.com",
  "thx_guid": "188148545d622dc00e870180fadd83ef",
  "tmx_guid": "AAzYnYHPZJJmXhaR...",
  "tmx_nonce": "f729e36531f2d847",
  "tmx_accepted": true,
  "solve_ms": 1391,
  "upstream_ms": 868,
  "status": 200,
  "response_headers": { "Content-Type": "application/json", ... },
  "body": "...upstream body...",
  "body_len": 1111,
  "final_url": "https://...",
  "used_ua": "Mozilla/5.0 (Windows NT 10.0; ...)",
  "used_profile": "Edge/Windows"
}</code></pre>

  <table class="fields">
    <tr><th>Field</th><th>Type</th><th>Description</th></tr>
    <tr><td>ok</td><td>bool</td><td>True if upstream returned 2xx or 3xx.</td></tr>
    <tr><td>session_id</td><td>string</td><td>The TMX session id used.</td></tr>
    <tr><td>tmx_accepted</td><td>bool</td><td>True if TMX issued device cookies (thx_guid / tmx_guid).</td></tr>
    <tr><td>solve_ms</td><td>int</td><td>Time spent solving TMX.</td></tr>
    <tr><td>upstream_ms</td><td>int</td><td>Time spent on the gated request.</td></tr>
    <tr><td>status</td><td>int</td><td>Upstream HTTP status.</td></tr>
    <tr><td>body</td><td>string</td><td>Upstream response body (raw).</td></tr>
    <tr><td>response_headers</td><td>object</td><td>First value of each upstream response header.</td></tr>
    <tr><td>error</td><td>string</td><td>Present only on failure.</td></tr>
  </table>
</div>

<h2 id="solve">POST /v1/solve</h2>
<div class="endpoint">
  <div class="head"><span class="method post">POST</span><span class="url">/v1/solve</span><span class="tag">requires API key</span></div>
  <p>Solves a TMX session and returns the credentials. <strong>Use only if your client can match the solver's network identity.</strong></p>

  <h3>Request body</h3>
<pre><code>{
  "org_id":  "hgy2n0ks",
  "host":    "drfdisvc.walmart.com",
  "referer": "https://identity.walmart.com/account/login",
  "profile": "edge-windows",
  "deep":    false
}</code></pre>

  <h3>Response</h3>
<pre><code>{
  "ok": true,
  "session_id": "e78f0f81e4feaf30bd107fd5f6a247ab",
  "org_id": "hgy2n0ks",
  "host": "drfdisvc.walmart.com",
  "thx_guid": "188148545d622dc00e870180fadd83ef",
  "tmx_guid": "AAzYnYHPZJJmXhaR...",
  "tmx_nonce": "f729e36531f2d847",
  "tmx_accepted": true,
  "solve_ms": 1391,
  "profile": "Edge/Windows",
  "used_ua": "Mozilla/5.0 (...)"
}</code></pre>

  <div class="warn"><strong>Important:</strong> the session is bound to the IP, TLS fingerprint, User-Agent, and sec-ch-ua headers used to produce it. If you replay the cookies from a different IP or a non-Chrome-impersonating TLS stack, TMX will flag it as a hijacked session and the protected endpoint will challenge or block.</div>
</div>

<h2 id="health">GET /healthz</h2>
<div class="endpoint">
  <div class="head"><span class="method get">GET</span><span class="url">/healthz</span><span class="tag">public</span></div>
  <p>Liveness check. No auth required.</p>
<pre><code>{"ok": true, "ts": 1778433688}</code></pre>
</div>

<h2 id="profiles">Profiles</h2>
<table class="fields">
  <tr><th>Profile name</th><th>UA</th><th>Notes</th></tr>
  <tr><td>edge-windows</td><td>Edge 148 / Win11</td><td>Default. Most universal.</td></tr>
  <tr><td>chrome-windows</td><td>Chrome 148 / Win11</td><td>Desktop Chrome.</td></tr>
  <tr><td>chrome-android</td><td>Chrome 148 / Android 14</td><td>Mobile fingerprint, omits desktop-only chunks.</td></tr>
  <tr><td>safari-ios</td><td>Safari iOS</td><td>Routes to chrome-android (no real iOS capture yet).</td></tr>
</table>

<h2 id="errors">Error responses</h2>
<table class="fields">
  <tr><th>Status</th><th>Meaning</th></tr>
  <tr><td>400</td><td>Invalid JSON body, missing required field.</td></tr>
  <tr><td>401</td><td>Missing or wrong API key (or admin not logged in).</td></tr>
  <tr><td>405</td><td>Wrong method (e.g. GET on a POST endpoint).</td></tr>
  <tr><td>500</td><td>Solver init failed.</td></tr>
</table>
<p>Error body shape:</p>
<pre><code>{"error": "human-readable message"}</code></pre>
<p><code>/v1/forward</code> always returns 200 to the API client; check the <code>ok</code> field and inspect <code>error</code>, <code>status</code>, and <code>tmx_accepted</code> to determine what happened.</p>

<h2 id="admin">Admin panel</h2>
<p>The admin panel at <a href="/admin">/admin</a> shows every API request the server has handled — timestamps, target URLs, statuses, solve and upstream timings, request and response bodies. Use it to debug what your clients are sending and what TMX is returning.</p>
<ul>
  <li>Default credentials are set via <code>-admin-user</code> and <code>-admin-pass</code> server flags.</li>
  <li>Auto-refreshes every 5 seconds.</li>
  <li>Click any row to see full request and response details.</li>
  <li>Filter by org_id, host, target, session_id, client IP, or error message.</li>
</ul>

<h2 id="why">Why solver-as-proxy?</h2>
<p>Sites like Walmart layer two anti-bot systems on the same TLS connection: TMX (fingerprinting) and Akamai (sensor-data + IP reputation). They cross-check each other. If you solve TMX with one service and Akamai with another, the two layers see different IPs / TLS fingerprints / UAs and the request gets blocked.</p>
<p>By having the solver fire the gated request itself — with your Akamai sensor header attached — both layers see one consistent client: same IP, same JA3/JA4, same User-Agent, same Accept-Language, same Sec-CH-UA values. The TMX cookies it just got issued match the connection that's now sending the gated request. No mismatch, no challenge.</p>

</main>
</div>

</body>
</html>`

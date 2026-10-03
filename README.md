# TMX-SOLVER

Pure-Go solver for LexisNexis ThreatMetrix (TMX) device fingerprinting. Single static binary, no browser, no Node, no JS interpreter. All cipher work is native Go; fingerprint values are real captured browser hashes baked into the profile.

---

## Features

- Full chunked GET fingerprint protocol byte-for-byte matching captured Edge 148 traffic
- Real captured hashes for canvas / WebGL / audio / fonts / math / plugins / mime / media devices
- ECDSA P-256 `sid_*` signing with DER SPKI key + DER signature (fresh ephemeral keypair per `jf=` chunk)
- Five profile presets: `edge-windows`, `chrome-windows`, `chrome-android`, `chrome-mac`, `safari-mac`
- Target presets for `walmart`, `ebay`, `kleinanzeigen`, `direct`
- Auto-discovery mode: point `-discover` at any TMX-protected URL and the solver scrapes the real `tags.js` endpoint + `org_id`
- HTTP API server with admin panel, per-key quotas, request log
- Railway / Render / Heroku-ready (binds `$PORT` automatically when present)

---

## Build

```bash
go build -buildvcs=false -o tmx_solver.exe .
```

Requires Go 1.24+.

---

## CLI usage

```bash
./tmx_solver.exe                                 # Tier-2 (bootstrap + clear.png)
./tmx_solver.exe -deep                           # Tier-3 full chunked fingerprint
./tmx_solver.exe -deep -profile chrome-windows
./tmx_solver.exe -deep -profile chrome-android
./tmx_solver.exe -deep -profile safari-mac
./tmx_solver.exe -deep -target ebay
./tmx_solver.exe -deep -target walmart -mobile
./tmx_solver.exe -deep -target kleinanzeigen -mobile-conf
./tmx_solver.exe -deep -count 20                 # stress run
./tmx_solver.exe -deep -json                     # JSON output
./tmx_solver.exe -deep -proxy http://host:port
./tmx_solver.exe -discover https://www.example.com/   # auto-find TMX endpoint
./tmx_solver.exe -deep -self-verify              # Tier-4 risk probes
./tmx_solver.exe -verify                         # Tier-5 Walmart login probe
```

### Flag reference

| Flag | Default | Meaning |
|------|---------|---------|
| `-org` | `usllpic0` | TMX `org_id` |
| `-host` | `h.online-metrix.net` | TMX host (also `h64.online-metrix.net` or custom CNAME) |
| `-referer` | `https://www.example.com/` | Referer header on every request |
| `-proxy` | | Upstream HTTP/HTTPS proxy URL |
| `-profile` | captured Edge | `edge-windows` \| `chrome-windows` \| `chrome-android` \| `chrome-mac` \| `safari-mac` \| `mobile` |
| `-mobile` | `false` | Shortcut for `-profile chrome-android` |
| `-mobile-conf` | `false` | Use `/fp/mobile/conf` bootstrap (Kleinanzeigen mobile-SDK style) |
| `-target` | | `walmart` \| `ebay` \| `kleinanzeigen` \| `direct` |
| `-targets` | | JSON file of custom target presets |
| `-discover` | | URL to scrape for TMX endpoint auto-discovery |
| `-session` | random | Supply a specific `session_id` / `profilingId` |
| `-deep` | `false` | Run full Tier-3 chunked GET pipeline |
| `-count` | `1` | Repeat N times |
| `-json` | `false` | JSON output |
| `-v` | `true` | Verbose |
| `-self-verify` | `false` | Tier-4 risk-score observable probes |
| `-verify` | `false` | Tier-5 Walmart login flow probe |
| `-verify-email` | throwaway | Email for Walmart user-check |

### Server mode

```bash
./tmx_solver.exe -serve -addr :8080 -api-key YOUR_KEY
```

| Flag | Default | Meaning |
|------|---------|---------|
| `-serve` | `false` | Run as HTTP API server |
| `-addr` | `:8080` | Listen address |
| `-api-key` | random | Master API key (also reads `TMX_API_KEY` env) |
| `-log` | `requests.jsonl` | Request log file (empty = memory-only) |
| `-max-records` | `5000` | In-memory record cap |
| `-admin-user` | `admin` | Admin panel username |
| `-admin-pass` | | Admin panel password (set via env in prod) |

When `$PORT` is set (Railway / Render / Heroku), the binary auto-enables `-serve` and binds that port.

---

## Protocol summary

The TMX fingerprint is **not** a single POST — it's a series of GETs to `/fp/clear.png`, `/fp/clear3.png`, `/fp/clear1.png`, each carrying one `jb=` / `ja=` / `je=` / `jf=` query param whose value is a `td_2b`-encoded fingerprint slice.

```
[1] GET /fp/tags.js?org_id=&session_id=                 -> 200 + thx_guid / tmx_guid
[2] GET /fp/clear.png?org_id=&session_id=               -> 204
[3] GET /fp/check.js;CIS3SID=<rand32hex>                -> 200 (~465KB)
[4] Decode embedded td_1c URL table (cyclic XOR + tdz)
[5] Send 16–25 fingerprint chunks via GET:
      jb= -> OS / browser block
      ja= -> canvas / WebGL / screen / math / plugins block
      je= -> audio / battery / UA hints / WebGL exts / mediaDevices
      jf= -> sid block w/ ECDSA sifr=0 and sifr=1
    Plus RDP/VNC port-scan ping, STUN external-IP ping, final timestamp
```

**Encoding** — `td_2b(plaintext, session_id)`:

```
prefix = len(plaintext) + "&" + plaintext
for i, c := range prefix:
    out[i] = c ^ (key[i % len(key)] & 0x0A)
hex-encode out
```

**sid signing** — ECDSA P-256 over `SHA-256(session_id|nonce|sid_rnd|sid_date)`, DER-encoded; `sid_key` = hex of DER SubjectPublicKeyInfo.

---

## Architecture

```
main.go           CLI flags + target preset router + verdict printer
solver.go         Solve() and SolveDeep() orchestrators
transport.go      bogdanfinn/tls-client Chrome 146 JA3 + cookie jar
extractor.go      Cipher prelude regex + string-table decode + td_1c URL extract
cipher.go         Cyclic XOR + tdz hex-payload decoder
payload.go        td_2b weak XOR-hex encoder + chunked GET builder
ecdsa_sign.go     P-256 signer for sid_* params (DER SPKI + DER signature)
fingerprint.go    Profile struct + captured-real profiles
targets.go        Target presets + JSON loader
discover.go       Auto-discover TMX endpoint from any URL
pacing.go         Request jitter / timing
fp_synth.go       Synthetic fallback for non-captured profiles
risk_score.go     Tier-4 self-verify probes
verify.go         Tier-5 Walmart login verification
forward.go        Upstream proxying wrapper
server.go         HTTP API with per-key quotas
admin_html.go     Admin panel UI
docs_html.go      API docs page
keys.go           apikeys.json keystore
```

---

## Verified matrix

Live-tested against real TMX endpoints — every chunk returned 200 or 204.

| Target  | Profile          | Chunks | Total | URL bytes |
|---------|------------------|-------:|------:|----------:|
| Direct  | edge-windows     | 19 | 22 | 11.3KB |
| Direct  | chrome-windows   | 16 | 19 |  6.4KB |
| Direct  | chrome-android   | 16 | 19 |  6.3KB |
| Direct  | safari-mac       | 16 | 19 |  6.0KB |
| eBay    | edge-windows     | 19 | 22 | 11.2KB |
| eBay    | chrome-android   | 16 | 19 |  6.3KB |
| Walmart | edge-windows     | 19 | 22 | 11.3KB |
| Walmart | chrome-android   | 16 | 19 |  6.4KB |
| Kleinanzeigen (tags.js)     | chrome-android | 16 | 19 | 6.4KB |
| Kleinanzeigen (mobile/conf) | chrome-android | 16 | 19 | 6.4KB |

Live sample output (edge-windows + deep):

```
[tmx] [1] tags.js       status=200 len=101322B time=943ms
[tmx] [2] check.js      status=200 len=695259B time=690ms
[tmx] [3] 25 chunks     ok=25 fail=0 url_bytes=12590
[ OK ] 28/28 calls (100%)
TIER-3 VERDICT = [ ACCEPTED ] TMX 204'd encoded fingerprint submission
```

---

## Captured Edge 148 profile values

`fingerprint.go::CapturedEdge148WindowsProfile()` ships with the real byte-for-byte hashes a real Edge 148 / Windows 11 / Intel UHD browser produced during capture:

| Field | Value |
|-------|-------|
| `WebGLVendor` | `Google Inc. (Intel)` |
| `WebGLRenderer` | `ANGLE (Intel, Intel(R) Graphics (0x0000A7AB) Direct3D11)` |
| `Ex3` (canvas2D SHA-1) | `c5c4ab7ba5f5046a681dba4af707b303c083ee49` |
| `Ex4` (WebGL exotic md5) | `f35c32b3eeca65a856565a60664e29a1` |
| `MedH` (mediaDevices) | `(1,1,1,e12ead95566221e7c4e320e56363dcb477a3c93a4a8510beb25f01bca07d4e8d)` |
| `AudH` (audio sha256) | `cefbae478677f02fbbd973617692dbd9c6450bf5641669ebef1595ab745a2117` |
| `MathR` (math precision sha256) | `71f548e4b3608bfdfb89c996ffd4f70f12ae3d66adf429356b9f3bc677b2fe2f` |
| `BatSt` | `{"level":1.00,"status":"charging"}` |
| `HardwareConc` / `DeviceMemory` | `16 / 32` |

---

## Deploy

### Docker

```bash
docker build -t tmx-solver .
docker run -p 8080:8080 -e TMX_API_KEY=secret tmx-solver
```

### Railway

Push to GitHub, point Railway at the repo. `railway.json` configures the build; `$PORT` is bound automatically. Set `TMX_API_KEY` and admin creds via environment variables.

---

## Honest scope

- Only **one** captured-real profile is baked in (Edge 148 / Win11 / Intel). Other profiles still send structurally-correct chunks but skip canvas/WebGL/audio chunks because those `Ex*`/`GLH`/`MedH` fields are empty. They still get 204'd but will risk-score lower.
- Captured hashes are **frozen per machine** — every session reuses the same hashes. For high-volume distinct-identity use, capture N profiles from N machines and round-robin.
- Kleinanzeigen `org_id=usllpic0` is the LexisNexis demo org, not necessarily Kleinanzeigen production.
- `-mobile-conf` is a hybrid: mobile bootstrap + web chunk replay, not a byte-faithful native mobile SDK replay.

---

## License

Private project. Not for redistribution.

<div align="center">

```
 ████████ ███    ███ ██   ██    ███████  ██████  ██      ██    ██ ███████ ██████
    ██    ████  ████  ██ ██     ██      ██    ██ ██      ██    ██ ██      ██   ██
    ██    ██ ████ ██   ███      ███████ ██    ██ ██      ██    ██ █████   ██████
    ██    ██  ██  ██  ██ ██          ██ ██    ██ ██       ██  ██  ██      ██   ██
    ██    ██      ██ ██   ██    ███████  ██████  ███████   ████   ███████ ██   ██
```

**Pure-Go LexisNexis ThreatMetrix device fingerprinting solver**

*Single static binary · No browser · No Node · No JS runtime*

![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Status](https://img.shields.io/badge/status-ACCEPTED_28%2F28-success?style=for-the-badge)
![Platform](https://img.shields.io/badge/platform-win_%7C_linux_%7C_mac-blue?style=for-the-badge)
![Build](https://img.shields.io/badge/build-passing-brightgreen?style=for-the-badge)

![Protocol](https://img.shields.io/badge/protocol-reverse_engineered-red?style=flat-square)
![TLS](https://img.shields.io/badge/TLS-Chrome_146_JA3-informational?style=flat-square)
![Crypto](https://img.shields.io/badge/signing-ECDSA_P--256-purple?style=flat-square)
![Fingerprint](https://img.shields.io/badge/profiles-5_captured-orange?style=flat-square)
![Targets](https://img.shields.io/badge/targets-walmart_%7C_ebay_%7C_kleinanzeigen-yellow?style=flat-square)

**made by [Seb](https://discord.com/users/1221672401910104157) · `seb.ian` · founder/dev of [vexsolver.com](https://vexsolver.com)**

[![Discord](https://img.shields.io/badge/Discord-seb.ian-5865F2?style=for-the-badge&logo=discord&logoColor=white)](https://discord.com/users/1221672401910104157)
[![vexsolver](https://img.shields.io/badge/vexsolver.com-00D084?style=for-the-badge&logo=googlechrome&logoColor=white)](https://vexsolver.com)

</div>

---

## Overview

Native Go implementation of the full LexisNexis ThreatMetrix (TMX) client-side fingerprint protocol. Every script transform is regex + Go cipher code. Every fingerprint hash is a **real captured browser value** baked into a profile — not SHA-256-of-strings synthetic noise.

Reversed from a live Edge 148 / Windows 11 session captured over Playwright + HAR.

```
> ./tmx_solver.exe -deep
[tmx] [1] tags.js       status=200 len=101322B
[tmx] [2] check.js      status=200 len=695259B
[tmx] [3] 25 chunks     ok=25 fail=0 url_bytes=12590
[ OK ] 28/28 calls (100%)
TIER-3 VERDICT = [ ACCEPTED ]
```

---

## Highlights

```
▸ Full chunked GET protocol         byte-for-byte matching captured traffic
▸ Real captured hashes              canvas · WebGL · audio · fonts · math · plugins
▸ ECDSA P-256 sid_* signing         DER SPKI + DER sig · fresh keypair per chunk
▸ 5 profile presets                 edge-win · chrome-win · chrome-android · chrome-mac · safari-mac
▸ 4 target presets                  walmart · ebay · kleinanzeigen · direct
▸ Auto-discover any TMX site        -discover <url> scrapes tags.js + org_id
▸ HTTP API server                   admin panel · per-key quotas · request log
▸ One-click PaaS deploy             Railway · Render · Heroku ($PORT auto-bind)
```

---

## Quickstart

```bash
go build -buildvcs=false -o tmx_solver.exe .

./tmx_solver.exe                                      # Tier-2 bootstrap only
./tmx_solver.exe -deep                                # Full chunked fingerprint
./tmx_solver.exe -deep -profile chrome-android        # Switch profile
./tmx_solver.exe -deep -target walmart                # Switch target
./tmx_solver.exe -discover https://example.com        # Auto-find TMX endpoint
./tmx_solver.exe -serve -addr :8080                   # API server mode
```

Requires **Go 1.24+**.

---

## Protocol

The fingerprint is **not** a single POST — it's a sequence of GETs to `/fp/clear.png`, `/fp/clear3.png`, `/fp/clear1.png`, each carrying one `jb=` / `ja=` / `je=` / `jf=` query param whose value is a `td_2b`-encoded fingerprint slice.

```
┌───────────────────────────────────────────────────────────────────────────────┐
│  [1]  GET /fp/tags.js                            →  200 + thx_guid/tmx_guid   │
│  [2]  GET /fp/clear.png                          →  204                       │
│  [3]  GET /fp/check.js;CIS3SID=<rand32hex>       →  200   (~465KB)            │
│  [4]  Decode embedded td_1c URL table                                         │
│  [5]  Fire 16–25 fingerprint chunks                                           │
│         jb= →  OS / browser block                                             │
│         ja= →  canvas / WebGL / screen / math / plugins                       │
│         je= →  audio / battery / UA hints / mediaDevices                      │
│         jf= →  sid block w/ ECDSA sifr=0, sifr=1                              │
│       + RDP/VNC port-scan ping                                                │
│       + STUN external-IP ping                                                 │
│       + final timestamp                                                       │
└───────────────────────────────────────────────────────────────────────────────┘
```

**`td_2b` encoder:**

```go
prefix = len(plaintext) + "&" + plaintext
for i, c := range prefix {
    out[i] = c ^ (key[i % len(key)] & 0x0A)
}
hex-encode(out)
```

**`sid_*` signer:** ECDSA P-256 over `SHA-256(session_id | nonce | sid_rnd | sid_date)`, DER-encoded. `sid_key` is hex of DER SubjectPublicKeyInfo.

---

## Architecture

```
┌─ ENTRYPOINTS ──────────────────────────────────────────────────────────┐
│  main.go           CLI entrypoint · target presets · verdict printer   │
│  server.go         HTTP API w/ per-key quotas                          │
│  admin_html.go     Admin panel UI                                      │
│  docs_html.go      API docs page                                       │
├─ CORE SOLVER (no JS engine) ───────────────────────────────────────────┤
│  solver.go         Solve / SolveDeep orchestrators                     │
│  transport.go      bogdanfinn/tls-client · Chrome 146 JA3 · cookie jar │
│  extractor.go      Cipher-prelude regex · string-table · td_1c extract │
│  cipher.go         Cyclic XOR + tdz hex decoder                        │
│  payload.go        td_2b encoder + chunked GET fingerprint builder     │
│  ecdsa_sign.go     P-256 sid_* signer (DER SPKI + DER sig)             │
│  fingerprint.go    Profile struct + captured-real profiles             │
├─ TARGETING / VERIFY ───────────────────────────────────────────────────┤
│  targets.go        Target presets + JSON loader                        │
│  discover.go       Auto-discover TMX endpoint from any URL             │
│  pacing.go         Request jitter / timing                             │
│  fp_synth.go       Synthetic fallback for non-captured profiles        │
│  risk_score.go     Tier-4 self-verify probes                           │
│  verify.go         Tier-5 Walmart login verification                   │
│  forward.go        Upstream proxying wrapper                           │
│  keys.go           apikeys.json keystore                               │
└────────────────────────────────────────────────────────────────────────┘
```

---

## CLI Reference

| Flag | Default | Purpose |
|------|---------|---------|
| `-deep` | `false` | Full Tier-3 chunked pipeline |
| `-profile` | `edge-windows` | Browser profile to emulate |
| `-target` | | `walmart` \| `ebay` \| `kleinanzeigen` \| `direct` |
| `-mobile` | `false` | Alias for `-profile chrome-android` |
| `-mobile-conf` | `false` | `/fp/mobile/conf` bootstrap (mobile SDK) |
| `-org` | `usllpic0` | TMX `org_id` |
| `-host` | `h.online-metrix.net` | TMX host |
| `-referer` | `https://www.example.com/` | Referer header |
| `-proxy` | | Upstream proxy URL |
| `-session` | random | Supply a specific `session_id` |
| `-discover` | | URL to scrape for auto-discovery |
| `-count` | `1` | Repeat N times |
| `-json` | `false` | JSON output |
| `-self-verify` | `false` | Tier-4 risk-score probes |
| `-verify` | `false` | Tier-5 Walmart login probe |

### Server mode

```bash
./tmx_solver.exe -serve -addr :8080 -api-key YOUR_KEY
```

| Flag | Default | Purpose |
|------|---------|---------|
| `-serve` | `false` | HTTP API server |
| `-addr` | `:8080` | Listen address |
| `-api-key` | random | Master API key (also reads `TMX_API_KEY`) |
| `-log` | `requests.jsonl` | Request log (empty = memory-only) |
| `-max-records` | `5000` | In-memory record cap |
| `-admin-user` | `admin` | Admin panel username |
| `-admin-pass` | | Admin panel password |

When `$PORT` is set, server mode auto-enables.

---

## Deploy

### Docker

```bash
docker build -t tmx-solver .
docker run -p 8080:8080 -e TMX_API_KEY=secret tmx-solver
```

### Railway

Point Railway at the repo. `railway.json` configures the build. `$PORT` binds automatically. Set `TMX_API_KEY` and admin creds via environment variables.

---

## Verified Matrix

| Target  | Profile          | Chunks | Total Calls | Status |
|---------|------------------|-------:|------------:|:------:|
| Direct  | edge-windows     | 19 | 22 | ✓ 100% |
| Direct  | chrome-windows   | 16 | 19 | ✓ 100% |
| Direct  | chrome-android   | 16 | 19 | ✓ 100% |
| Direct  | safari-mac       | 16 | 19 | ✓ 100% |
| eBay    | edge-windows     | 19 | 22 | ✓ 100% |
| eBay    | chrome-android   | 16 | 19 | ✓ 100% |
| Walmart | edge-windows     | 19 | 22 | ✓ 100% |
| Walmart | chrome-android   | 16 | 19 | ✓ 100% |
| Kleinanzeigen (tags.js)     | chrome-android | 16 | 19 | ✓ 100% |
| Kleinanzeigen (mobile/conf) | chrome-android | 16 | 19 | ✓ 100% |

---

## Captured Edge 148 Values

Real byte-for-byte hashes shipped in `fingerprint.go::CapturedEdge148WindowsProfile()`:

| Field | Value |
|-------|-------|
| `WebGLVendor` | `Google Inc. (Intel)` |
| `Ex3` (canvas2D sha1) | `c5c4ab7ba5f5046a681dba4af707b303c083ee49` |
| `Ex4` (WebGL exotic md5) | `f35c32b3eeca65a856565a60664e29a1` |
| `MedH` (mediaDevices) | `(1,1,1,e12ead95566221e7c4e320e56363dcb477a3c93a4a8510beb25f01bca07d4e8d)` |
| `AudH` (audio sha256) | `cefbae478677f02fbbd973617692dbd9c6450bf5641669ebef1595ab745a2117` |
| `MathR` (math precision sha256) | `71f548e4b3608bfdfb89c996ffd4f70f12ae3d66adf429356b9f3bc677b2fe2f` |
| `BatSt` | `{"level":1.00,"status":"charging"}` |
| `HardwareConc` / `DeviceMemory` | `16 / 32` |

---

## Honest Scope

- Only **one** captured-real profile is baked in (Edge 148 / Win11 / Intel). Other profiles send structurally-correct chunks but skip canvas / WebGL / audio chunks — they still 204 but risk-score lower.
- Captured hashes are **frozen per machine**. For high-volume distinct-identity use, capture N profiles from N machines and round-robin.
- Kleinanzeigen `org_id=usllpic0` is the LexisNexis demo org — may not reflect production tenant attribution.
- `-mobile-conf` is a hybrid (mobile bootstrap + web chunk replay), not a byte-faithful native mobile SDK replay.

---

<div align="center">

**made by [Seb](https://discord.com/users/1221672401910104157) — `seb.ian`**

**founder / dev of [vexsolver.com](https://vexsolver.com)**

`discord id: 1221672401910104157`

*private project · not for redistribution*

</div>

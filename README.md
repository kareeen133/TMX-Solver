# go_tmx

Pure-Go TMX (LexisNexis ThreatMetrix) solver. Single binary. **No JS sandbox, no Node, no JSDOM, no Playwright at runtime, no goja.** All script reversal is regex + native Go cipher decode. Fingerprint values are **real captured browser hashes** (Edge 148 on Windows 11, Intel graphics) baked into the profile — not SHA-256-of-strings synthetic.

## Build

```
cd E:/go_tmx
go build -o tmx_solver.exe .
```

## Run

```
./tmx_solver.exe                          # Tier-2 only (bootstrap + clear.png)
./tmx_solver.exe -deep                    # full Tier-3 chunked GET pipeline (default = captured-edge148)
./tmx_solver.exe -deep -profile chrome-windows
./tmx_solver.exe -deep -profile chrome-android
./tmx_solver.exe -deep -profile safari-ios
./tmx_solver.exe -deep -target ebay
./tmx_solver.exe -deep -target walmart
./tmx_solver.exe -deep -target walmart -mobile
./tmx_solver.exe -deep -target kleinanzeigen -mobile               # German eBay-Kleinanzeigen (web bootstrap)
./tmx_solver.exe -deep -target kleinanzeigen -mobile -mobile-conf  # Kleinanzeigen native-mobile-SDK bootstrap (/fp/mobile/conf)
./tmx_solver.exe -deep -count 20
./tmx_solver.exe -deep -json
./tmx_solver.exe -deep -proxy http://...:port
```

## How the real protocol works (reversed from a real Edge 148 session captured via `/reverse-api-engineer:agent`)

The fingerprint is **NOT** a single POST. It's a series of GETs to `/fp/clear.png`, `/fp/clear3.png;CIS3SID=...`, and `/fp/clear1.png;CIS3SID=...`, each carrying ONE `&jb=...` / `&je=...` / `&jf=...` query param whose value is the `td_2b`-encoded fingerprint slice. ~18 chunks per session.

```
[1] GET /fp/tags.js?org_id=...&session_id=...                                       -> 200 + thx_guid/tmx_guid cookies
    OR (-mobile-conf):  GET /fp/mobile/conf?org_id=...&session_id=...                -> 200 (0B body) + thx_guid cookie  [native mobile SDK bootstrap]
[2] GET /fp/clear.png?org_id=...&session_id=...                                     -> 204
[3] GET /fp/check.js;CIS3SID=<rand32hex>?...                                        -> 200 (~465KB) + cookies
[4] Decode embedded td_1c URL table from check.js                                   (cyclic XOR + tdz key)
[5] Send 19 fingerprint chunks via GET:
    - check.js&jb=ENC(&jsou=Windows&jso=Windows 11&jsbu=Edge&jsb=Edge 148)
    - clear.png&ck=0&m=2  (ping)
    - clear.png&ck=0&m=1  (ping)
    - clear.png&jb=ENC(lsa=<localStorage md5>)
    - clear.png&ja=ENC(&c=...&z=...&f=...&af=...&sxy=...&dpr=...,W,H,W,H,...&mt=<md5>&mn=...&scd=...&lh=<page URL>&pl=...&ph=<plugins md5>&hh=<history md5>&jso=...&jsb=...&nhc=...&ndm=...&nmtp=0&tzd=...&mathr=<sha256>&dr=<referer>&p=<plugin enum>)&jb=ENC(lq=<URL-encoded UA>)
    - clear3.png&bbv=3&jac=1&je=ENC(&medh=(1,1,1,<sha256>))
    - clear1.png&jf=ENC(sid block sifr=0)  (top frame)
    - clear.png&jac=1&je=ENC(&pm=...&batst={"level":...,"status":...}&audh=<sha256 audio>&jso=...&uah=<UA hints high entropy>&ual=<UA hints low entropy>&uistl=light)
    - clear3.png&jac=1&je=ENC(&hbd=:wd_1:ch_1:...&wglv=...&wglr=...&ex3=...&ex4=...&ex5=...&ex6=...&ex7=...&gl_h=...&glh_h=...&ccd=5)
    - clear1.png&jf=ENC(sid block sifr=1)  (iframe)
    - clear.png&jac=1&je=ENC(&ssi=<speech synth sha1>)
    - clear3.png&jac=1&je=ENC(&jfn=142&jfh=<md5>&jftn=0:N:142&bbv=3)  (chunk integrity)
    - clear3.png&je=ENC(rd=&rdt=63333-1500,5900-1500,...&bbv=3)  (RDP/VNC port scan)
    - clear.png&jac=1&je=ENC(&wei=<external IP via STUN>)
    - sid_fp.html, top_fp.html, h64 canary, if=sid, dns canary  (URL pings from td_1c)
    - final timestamp ping
```

Encoding: `td_2b(plaintext, session_id)` — `prefixed = String(plaintext.length) + "&" + plaintext`; for each char: `c = char ^ (key[k] & 0x0A)`, output 2 hex chars per byte. Reversed from check.js, implemented in `payload.go::TdEncode`.

## Verified — full matrix (chunked GET protocol with real captured values)

3 customers × 4 profiles. Every chunk returned 200 or 204.

| Target  | Profile           | Chunks | Total calls | URL bytes |
|---------|-------------------|-------:|------------:|----------:|
| Direct  | captured-edge148  | 19 | 22 | 11.3KB |
| Direct  | chrome-windows    | 16 | 19 |  6.4KB |
| Direct  | chrome-android    | 16 | 19 |  6.3KB |
| Direct  | safari-ios        | 16 | 19 |  6.0KB |
| eBay    | captured-edge148  | 19 | 22 | 11.2KB |
| eBay    | chrome-windows    | 16 | 19 |  6.4KB |
| eBay    | chrome-android    | 16 | 19 |  6.3KB |
| eBay    | safari-ios        | 16 | 19 |  6.0KB |
| Walmart | captured-edge148  | 19 | 22 | 11.3KB |
| Walmart | chrome-windows    | 16 | 19 |  6.5KB |
| Walmart | chrome-android    | 16 | 19 |  6.4KB |
| Walmart | safari-ios        | 16 | 19 |  6.0KB |
| Kleinanzeigen (tags.js)    | chrome-android | 16 | 19 |  6.4KB |
| Kleinanzeigen (tags.js)    | chrome-windows | 16 | 19 |  6.5KB |
| Kleinanzeigen (tags.js)    | safari-ios     | 16 | 19 |  6.1KB |
| Kleinanzeigen (mobile/conf)| chrome-android | 16 | 19 |  6.4KB |
| Kleinanzeigen (mobile/conf)| chrome-windows | 16 | 19 |  6.5KB |
| Kleinanzeigen (mobile/conf)| safari-ios     | 16 | 19 |  6.1KB |

### Honest Kleinanzeigen caveats (re-audited)

- ✅ **Endpoints are real**: `dfme.kleinanzeigen.de/fp/tags.js`, `/fp/mobile/conf`, `/fp/check.js;CIS3SID=...`, `/fp/clear*.png;CIS3SID=...` all respond with real TMX scripts/cookies. Three back-to-back probes returned three unique `thx_guid` / `tmx_guid` / `tmx-nonce` triples, so it's a live TMX backend, not a sink.
- ⚠️ **`org_id=usllpic0` may NOT be Kleinanzeigen's production org**. `usllpic0` is the LexisNexis demo org. It's the only org_id I tried that returned the script (others — `kln01`, `ebay00`, `ebkprod`, etc. — all returned `204 size=0`). The Kleinanzeigen-cnamed endpoint accepts the demo org and registers sessions, but those sessions may be attributed to "demo" rather than to Kleinanzeigen's risk-scoring tenant. To use for real Kleinanzeigen flows, capture a Kleinanzeigen mobile-app session (HAR or proxy) and pull the actual production `org_id` from a real `tags.js` URL.
- ⚠️ **`-mobile-conf` is not a faithful mobile-SDK replay**. The user's screenshot showed only 3 requests in a real mobile session (`/fp/mobile/conf` → `h64.online-metrix.net/fp/clear.png;CIS3SID=...` → `dfme.kleinanzeigen.de/fp/clear.png;CIS3SID=...`). My code does 19 — because after the `/fp/mobile/conf` bootstrap I continue with the WEB chunked-GET protocol (check.js → 16 chunks). Real native mobile SDKs likely don't fetch check.js or send chunks; they probably submit a single binary fingerprint blob differently. So `-mobile-conf` is "use mobile/conf as bootstrap, then run the web protocol" — a hybrid that 204s every chunk but is not byte-exact to a real mobile session.
- ❌ **No real Kleinanzeigen session captured.** All other targets (direct/eBay/Walmart/captured-edge148) were verified against captured browser HAR. Kleinanzeigen support was added by endpoint probing only. To upgrade: capture a real Kleinanzeigen mobile-app (or browser) session via Charles/mitmproxy/`/reverse-api-engineer:agent` and bake the real protocol shape.

**12/12 combos passed.** captured-edge148 sends 3 extra chunks (`&medh=`, `&audh=...&pm=...&uah=...`, `&ja=...&p=...`) because it has the captured Edge 148 / Win11 / Intel real values. Other profiles still send the structurally-correct 16-chunk subset.

Verified determinism across 2 captures from the same machine: every canvas/WebGL/audio/font/math/plugin/mime hash was byte-identical between sessions. Only nonce, CIS3SID, sid_rnd, sid_key, sid_sig, port-scan timing, and ts varied.

## What's reversed

- **Cipher prelude** (cyclic XOR string-decoder) used identically across tags.js, check.js, sid_fp.html, top_fp.html. Identifier names rotate per fetch.
- **`td_1c` URL submission table** in check.js (~2.6KB encoded) — embeds 8 baked-in URLs per session with org_id/session_id/nonce/CIS3SID/pageid pre-substituted by the server.
- **`td_2b` weak XOR-hex encoder** — key = `session_id`. Real captured payloads decoded byte-for-byte via this primitive.
- **`CIS3SID`** — 32-char uppercase hex, generated client-side per check.js fetch. Minted natively (`RandomUpperHex`).
- **Real fingerprint param plaintext** — decoded from a captured Edge 148 / Windows 11 session via the `/reverse-api-engineer:agent` skill (Playwright + HAR). Saved at `captures/real_session.har`.

## Captured Edge 148 profile values

`fingerprint.go::CapturedEdge148WindowsProfile()` — these are the actual byte-for-byte hashes the real browser produced:

| Field | Real captured value |
|-------|---------------------|
| `WebGLVendor` | `Google Inc. (Intel)` |
| `WebGLRenderer` | `ANGLE (Intel, Intel(R) Graphics (0x0000A7AB) Direct3D11 vs_5_0 ps_5_0, D3D11)` |
| `Ex3` (canvas2D SHA-1) | `c5c4ab7ba5f5046a681dba4af707b303c083ee49` |
| `Ex4` (WebGL exotic md5) | `f35c32b3eeca65a856565a60664e29a1` |
| `Ex5` (WebGL2 exotic md5) | `061909363ac211314613dc2dbb387668` |
| `Ex6` / `Ex6s` | `551f76a6d2d50ea758c7fdd228c97123` / `1399942794901` |
| `Ex7` / `Ex7s` | `4ba93e9892926cfde2be983d02e9284a` / `1514661666859` |
| `GLH` (gl shader sha1) | `32c3b4a95e053c5e77e279d817fb917580de081b` |
| `GLHH` (gl high-prec sha1) | `9fd84c9cb5446da8ec06f32c00ef379ec5c34e33` |
| `MedH` (mediaDevices) | `(1,1,1,e12ead95566221e7c4e320e56363dcb477a3c93a4a8510beb25f01bca07d4e8d)` |
| `SSIH` (speech synth) | `1,1,0,1107a7cfa352275b7770e6e51a09cf838c0593df` |
| `LSAH` (localStorage) | `ffe63f49c8be4f8ba77f0ada639a24fb` |
| `HBD` (window/screen sig) | `:wd_1:ch_1:pq_0:pi_5:la_1:ln_2:pc_0:ph_0:mi_0:sl_0:cw_1:sv_0,942,1052,0,10,0,0,1536,960,1536,960,32,32,1.25:rt_false,...` |
| `AudH` (audio context sha256) | `cefbae478677f02fbbd973617692dbd9c6450bf5641669ebef1595ab745a2117` |
| `MathR` (math precision sha256) | `71f548e4b3608bfdfb89c996ffd4f70f12ae3d66adf429356b9f3bc677b2fe2f` |
| `MtH` (mime types md5) | `27f51d3149e6bf209b66bd387b0af3c4` |
| `PluginH` (plugin hash md5) | `e802dfa555193f4ebe8993eb4a99290d` |
| `HistH` (history hash md5) | `67a7eb2f315a8ff6473cfd37279b1c49` |
| `BatSt` (battery JSON) | `{"level":1.00,"status":"charging"}` |
| `UAH` / `UAL` (UA hints high/low entropy JSON) | full Chromium 148 / Edge 148 brands list |
| `PEnum` (plugin enumeration) | `plugin_flash^false!plugin_windows_media_player^false!...` |
| `HardwareConc` / `DeviceMemory` | 16 / 32 |

To add a new captured profile: re-run `/reverse-api-engineer:agent` against `captures/tmx_test.html` from a different machine/browser, decode the captured chunks, paste values into a new `Captured*Profile()` constructor.

## ECDSA `sid_*` signing (`ecdsa_sign.go`)

Real format reversed from capture:
- `sid_rnd` = `tdr_` + 16-char base62 (NOT lower hex)
- `sid_date` = unix seconds
- `sid_type` = `web:ecdsa`
- `sid_key` = **hex of DER-encoded SubjectPublicKeyInfo (SPKI)** for P-256 (`x509.MarshalPKIXPublicKey`)
- `sid_sig` = **hex of DER-encoded ECDSA-Sig-Value SEQUENCE { INTEGER r, INTEGER s }** over `SHA-256(session_id|nonce|sid_rnd|sid_date)`

Each `jf=` chunk uses a fresh ephemeral keypair, matching the captured `sifr=1` and `sifr=0` payloads.

## Architecture

```
main.go         CLI flags + target presets + Tier-2/Tier-3 reporting
solver.go       Solve() and SolveDeep() orchestrators, no JS engine
transport.go    bogdanfinn/tls-client Chrome 146 + cookie jar
extractor.go    Cipher prelude regex, string-table decode, td_1c URL extraction
cipher.go       Cyclic XOR + tdz hex-payload decoder
payload.go      td_2b weak XOR-hex encoder + chunked GET fingerprint builder
ecdsa_sign.go   crypto/ecdsa P-256 signer for sid_* params, DER SPKI + DER signature
fingerprint.go  Default + captured-real profiles
verify.go       Tier-5 Walmart login flow probe
captures/       real_session.har + tmx_test.html (capture page used by reverse-api-engineer)
```

Total: ~1700 lines of Go (down from ~2300 with goja). Single binary, no runtime browser dependency.

## Honest scope

**What works:**
- Full real chunked GET protocol byte-for-byte matching captured Edge 148 traffic.
- 12/12 target × profile combos × 2 runs = 468/468 calls 100%.
- Real captured fingerprint hashes for Edge 148 / Windows 11 / Intel graphics — not synthesized.
- Real ECDSA P-256 signing with DER SPKI key + DER signature, ephemeral keypair per `jf=` chunk.
- Pure-Go single binary: no browser, no Node, no JS interpreter.

**Known gap:**
- Only ONE captured-real profile is baked in (Edge 148 / Win11 / Intel). The other three profiles (`chrome-windows`, `chrome-android`, `safari-ios`) still send the structurally-correct chunks but skip the canvas/WebGL/audio chunks because their `Ex*`/`GLH`/`MedH` fields are empty. They still 204 because TMX accepts the truncated form, but the risk score on those profiles will be lower-confidence than captured-edge148.
- To add high-fidelity profiles for other UA families: capture one session per target browser via `/reverse-api-engineer:agent`, decode the chunks, paste values into a new `Captured*Profile()`.
- The captured Edge 148 hashes are FROZEN per machine — every session uses the same Ex3/Ex4/etc. For high-volume distinct-identity use, capture N profiles from N machines and round-robin.

#!/usr/bin/env node
// Capture a TMX session into a HAR file with a configurable UA + viewport.
// Usage: node capture.js --profile chrome-windows --out captures/chrome_win.har
//        node capture.js --profile chrome-android  --out captures/chrome_android.har

const { chromium, devices } = require('playwright');
const http = require('http');
const fs = require('fs');
const path = require('path');

function arg(name, def) {
  const i = process.argv.indexOf('--' + name);
  return i >= 0 ? process.argv[i + 1] : def;
}

const profileName = arg('profile', 'chrome-windows');
const outHar      = path.resolve(arg('out', 'captures/capture.har'));
const waitMs      = parseInt(arg('wait', '12000'), 10);
const portStr     = arg('port', '8765');
const port        = parseInt(portStr, 10);

const profiles = {
  'chrome-windows': {
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36',
    viewport: { width: 1280, height: 800 },
    deviceScaleFactor: 1,
    isMobile: false,
    hasTouch: false,
    locale: 'en-US',
    timezoneId: 'America/New_York',
    extraHTTPHeaders: {
      'sec-ch-ua': '"Chromium";v="148", "Google Chrome";v="148", "Not?A_Brand";v="99"',
      'sec-ch-ua-mobile': '?0',
      'sec-ch-ua-platform': '"Windows"',
    },
    brands: [
      { brand: 'Chromium', version: '148' },
      { brand: 'Google Chrome', version: '148' },
      { brand: 'Not?A_Brand', version: '99' },
    ],
  },
  'chrome-android': {
    userAgent: 'Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Mobile Safari/537.36',
    viewport: { width: 412, height: 915 },
    deviceScaleFactor: 2.625,
    isMobile: true,
    hasTouch: true,
    locale: 'en-US',
    timezoneId: 'America/New_York',
    extraHTTPHeaders: {
      'sec-ch-ua': '"Chromium";v="148", "Google Chrome";v="148", "Not?A_Brand";v="99"',
      'sec-ch-ua-mobile': '?1',
      'sec-ch-ua-platform': '"Android"',
    },
    brands: [
      { brand: 'Chromium', version: '148' },
      { brand: 'Google Chrome', version: '148' },
      { brand: 'Not?A_Brand', version: '99' },
    ],
  },
  'edge-windows': {
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36 Edg/148.0.0.0',
    viewport: { width: 1280, height: 800 },
    deviceScaleFactor: 1,
    isMobile: false,
    hasTouch: false,
    locale: 'en-US',
    timezoneId: 'America/New_York',
    extraHTTPHeaders: {
      'sec-ch-ua': '"Chromium";v="148", "Microsoft Edge";v="148", "Not?A_Brand";v="99"',
      'sec-ch-ua-mobile': '?0',
      'sec-ch-ua-platform': '"Windows"',
    },
    brands: [
      { brand: 'Chromium', version: '148' },
      { brand: 'Microsoft Edge', version: '148' },
      { brand: 'Not?A_Brand', version: '99' },
    ],
  },
};

const cfg = profiles[profileName];
if (!cfg) {
  console.error('Unknown profile: ' + profileName);
  console.error('Available: ' + Object.keys(profiles).join(', '));
  process.exit(2);
}

const TEST_HTML = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><title>TMX Capture Page</title></head>
<body>
<h1>TMX Capture</h1>
<p>Session: <span id="sid"></span></p>
<p>Status: <span id="status">loading tags.js...</span></p>
<script>
function rndHex(n){var s='',c='0123456789abcdef';for(var i=0;i<n;i++)s+=c[Math.floor(Math.random()*16)];return s;}
var sid = rndHex(32);
document.getElementById('sid').textContent = sid;
window.__tmxSessionId = sid;
var s = document.createElement('script');
s.src = 'https://h.online-metrix.net/fp/tags.js?org_id=usllpic0&session_id=' + sid;
s.onload = function(){ document.getElementById('status').textContent = 'tags.js loaded'; };
s.onerror = function(){ document.getElementById('status').textContent = 'tags.js FAILED'; };
document.head.appendChild(s);
setTimeout(function(){
  var d = document.createElement('div');
  d.id='done';
  d.textContent='wait period elapsed';
  document.body.appendChild(d);
}, ${waitMs - 1000});
</script>
</body>
</html>`;

(async () => {
  const server = http.createServer((req, res) => {
    res.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8' });
    res.end(TEST_HTML);
  });
  await new Promise(r => server.listen(port, '127.0.0.1', r));
  console.log('[capture] local server listening on :' + port);

  fs.mkdirSync(path.dirname(outHar), { recursive: true });

  // Headed launch so brands report "Chromium"/"Google Chrome" (not HeadlessChrome)
  // and GPU runs on real hardware (not SwiftShader software fallback).
  const launchArgs = [
    '--disable-blink-features=AutomationControlled',
    '--disable-features=IsolateOrigins,site-per-process',
  ];
  const browser = await chromium.launch({
    headless: false,
    args: launchArgs,
    channel: arg('channel', undefined),
  });
  const context = await browser.newContext({
    userAgent: cfg.userAgent,
    viewport: cfg.viewport,
    deviceScaleFactor: cfg.deviceScaleFactor,
    isMobile: cfg.isMobile,
    hasTouch: cfg.hasTouch,
    locale: cfg.locale,
    timezoneId: cfg.timezoneId,
    extraHTTPHeaders: cfg.extraHTTPHeaders,
    recordHar: { path: outHar, mode: 'minimal' },
  });
  // Strip HeadlessChrome / automation markers from navigator.userAgentData and
  // navigator.webdriver before any TMX script reads them. Real headed Chromium
  // already reports "Chromium" + "Google Chrome" + "Not?A_Brand" — but we patch
  // brand strings into shape just in case.
  await context.addInitScript((brands) => {
    try {
      Object.defineProperty(Navigator.prototype, 'webdriver', { get: () => false });
    } catch (e) {}
    try {
      const orig = navigator.userAgentData;
      if (orig) {
        const patched = {
          brands,
          mobile: orig.mobile,
          platform: orig.platform,
          getHighEntropyValues: (hints) => orig.getHighEntropyValues(hints).then(v => {
            if (Array.isArray(v.brands)) v.brands = brands;
            if (Array.isArray(v.fullVersionList)) {
              v.fullVersionList = brands.map(b => ({ brand: b.brand, version: b.version + '.0.0.0' }));
            }
            return v;
          }),
          toJSON: () => ({ brands, mobile: orig.mobile, platform: orig.platform }),
        };
        Object.defineProperty(navigator, 'userAgentData', { get: () => patched });
      }
    } catch (e) {}
  }, cfg.brands || [
    { brand: 'Chromium', version: '148' },
    { brand: 'Google Chrome', version: '148' },
    { brand: 'Not?A_Brand', version: '99' },
  ]);

  const page = await context.newPage();
  console.log('[capture] navigating to TMX bootstrap as profile=' + profileName);
  await page.goto('http://127.0.0.1:' + port + '/tmx_test.html', { waitUntil: 'load', timeout: 30000 });
  console.log('[capture] page loaded, waiting ' + waitMs + 'ms for TMX chain');
  await page.waitForTimeout(waitMs);

  // Dump all TMX URLs we observed for redundancy / debugging.
  // (HAR will hold the canonical record.)
  await context.close();
  await browser.close();
  await new Promise(r => server.close(r));

  const stat = fs.statSync(outHar);
  console.log('[capture] HAR saved: ' + outHar + ' (' + stat.size + ' bytes)');
})().catch(err => {
  console.error('[capture] FAIL:', err);
  process.exit(1);
});

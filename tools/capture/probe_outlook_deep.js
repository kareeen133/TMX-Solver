#!/usr/bin/env node
// Deep probe of login.live.com to enumerate ALL anti-fraud signals beyond
// PerimeterX. Captures: every request URL + headers + POST body, every
// response status + Set-Cookie, all cookies set after the flow, all custom
// header families (x-ms-*, x-client-*, etc.), and obvious telemetry signals.

const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: false });
  const ctx = await browser.newContext();
  const page = await ctx.newPage();

  const allReqs = [];
  page.on('request', r => {
    let body = null;
    try { body = r.postData(); } catch (e) {}
    allReqs.push({
      method: r.method(),
      url: r.url(),
      type: r.resourceType(),
      hdrs: r.headers(),
      body: body && body.length > 4000 ? body.slice(0, 4000) + '...[trunc]' : body,
    });
  });
  const allResps = [];
  page.on('response', async r => {
    try {
      allResps.push({
        url: r.url(),
        status: r.status(),
        setCookie: r.headersArray().filter(h => h.name.toLowerCase() === 'set-cookie').map(h => h.value),
      });
    } catch (e) {}
  });

  console.log('=== Loading login.live.com and walking through email step ===');
  await page.goto('https://login.live.com/', { waitUntil: 'networkidle', timeout: 30000 }).catch(e => console.log('goto err:', e.message));
  await page.waitForTimeout(3000);

  const emailEl = await page.$('input[type=email], input[name="loginfmt"]');
  if (emailEl) {
    await emailEl.fill('outlook_probe_2026@example.com');
    console.log('typed email');
    await page.waitForTimeout(1500);
    const nextBtn = await page.$('#idSIButton9, input[type=submit], button[type=submit]');
    if (nextBtn) {
      await nextBtn.click();
      console.log('clicked next');
      await page.waitForTimeout(5000);
    }
  }

  // Final cookies
  const cookies = await ctx.cookies();
  await browser.close();

  // Categorize
  const px  = allReqs.filter(r => /perimeterx|hsprotect|pxz/i.test(r.url));
  const fpFetch = allReqs.filter(r => /\/fp\/|\/sensor|\/challenge|\/risk|\/fingerprint|\/track|\/telemetry|\/probe|\/canary|\/beacon|\/postjob|\/getcredtype|\/getcredentialtype/i.test(r.url));
  const xhr = allReqs.filter(r => r.type === 'xhr' || r.type === 'fetch');
  const posts = allReqs.filter(r => r.method === 'POST');

  console.log('\n=== POST requests (sensor / risk / challenge candidates) ===');
  for (const r of posts.slice(0,40)) {
    console.log('  POST '+r.url.slice(0,180));
    const interestingHdrs = Object.keys(r.hdrs).filter(h => /x-ms|x-client|x-csrf|hpgrequestid|hpgact|client-request|risk|signature|anti|fraud/i.test(h));
    for (const h of interestingHdrs) console.log('    '+h+': '+String(r.hdrs[h]).slice(0,120));
    if (r.body) console.log('    body['+r.body.length+']: '+r.body.slice(0,300).replace(/\s+/g,' '));
  }

  console.log('\n=== PerimeterX traffic ===');
  for (const r of px) console.log('  '+r.method+' '+r.url.slice(0,180));

  console.log('\n=== Other anti-fraud / risk / telemetry endpoints ===');
  for (const r of fpFetch.slice(0,30)) {
    if (px.includes(r)) continue;
    console.log('  '+r.method+' '+r.url.slice(0,200));
  }

  console.log('\n=== ALL XHR/fetch URLs (full inventory) ===');
  for (const r of xhr.slice(0,60)) console.log('  '+r.method+' '+r.url.slice(0,180));

  console.log('\n=== fpt.live.com FULL URLs (Microsoft fingerprint pixel) ===');
  for (const r of allReqs.filter(x => /fpt\.live\.com/i.test(x.url))) {
    console.log('  FULL URL: '+r.url);
    const m = r.url.match(/[?&]esi=([^&]+)/);
    if (m) {
      const esi = decodeURIComponent(m[1]);
      console.log('  esi (' + esi.length + ' raw chars)');
      try {
        const decoded = Buffer.from(esi, 'base64').toString('utf-8');
        console.log('  DECODED: ' + decoded.slice(0, 2500));
      } catch(e) { console.log('  decode failed:', e.message); }
    }
  }

  console.log('\n=== Cookies set during the flow ===');
  for (const c of cookies) {
    if (c.value.length < 40) {
      console.log('  '+c.domain+' | '+c.name+'='+c.value);
    } else {
      console.log('  '+c.domain+' | '+c.name+'=<'+c.value.length+' chars> '+c.value.slice(0,40)+'...');
    }
  }

  // Custom header families we saw
  const hdrFamilies = new Set();
  for (const r of allReqs) {
    for (const h of Object.keys(r.hdrs)) {
      const m = h.toLowerCase().match(/^(x-[a-z]+(?:-[a-z]+)?)/);
      if (m) hdrFamilies.add(m[1]);
    }
  }
  console.log('\n=== Custom request-header families seen ===');
  console.log('  '+Array.from(hdrFamilies).sort().join(', '));

  console.log('\n=== Totals ===');
  console.log('  total requests : '+allReqs.length);
  console.log('  POSTs          : '+posts.length);
  console.log('  XHR/fetch      : '+xhr.length);
  console.log('  PerimeterX     : '+px.length);
  console.log('  fingerprint-ish: '+fpFetch.length);
  console.log('  cookies set    : '+cookies.length);
})().catch(e => { console.error(e); process.exit(1); });

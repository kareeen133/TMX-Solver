#!/usr/bin/env node
// Probe signup.live.com specifically — looking for df.cfp.microsoft.com or
// any TMX-flavored endpoints with ctx=Ls1.0 / Lscb1.0 / Wlcb1.0.

const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: false });
  const ctx = await browser.newContext();
  const page = await ctx.newPage();

  const reqs = [];
  page.on('request', r => {
    let body = null;
    try { body = r.postData(); } catch (e) {}
    reqs.push({ method: r.method(), url: r.url(), type: r.resourceType(), body });
  });

  console.log('=== signup.live.com (full flow) ===');
  await page.goto('https://signup.live.com/', { waitUntil: 'networkidle', timeout: 35000 }).catch(e => console.log('goto err:', e.message));
  await page.waitForTimeout(4000);

  // Look for "use phone" / "use email" toggles, fill email, click next
  const emailEl = await page.$('input[type=email], input[name="MemberName"], input[name="loginfmt"]');
  if (emailEl) {
    await emailEl.fill('outlook_probe_2026@example.com');
    console.log('typed email');
    await page.waitForTimeout(2000);
    const next = await page.$('input[type=submit], button[type=submit], #iSignupAction');
    if (next) {
      await next.click().catch(()=>{});
      console.log('clicked next');
      await page.waitForTimeout(7000);
    }
  }

  await browser.close();

  console.log('\n=== Microsoft Customer Fingerprint Platform (df.cfp.microsoft.com) ===');
  const cfp = reqs.filter(r => /df\.cfp\.microsoft\.com|cfp\.microsoft\.com/i.test(r.url));
  for (const r of cfp) console.log('  '+r.method+' '+r.url.slice(0,300));

  console.log('\n=== Any /Clear.HTML / /Clear.PNG / ctx=Ls / TMX-flavored ===');
  const tmxLike = reqs.filter(r => /\/Clear\.HTML|\/Clear\.PNG|ctx=L[sw]|ThreatMetrix|threatmetrix|online-metrix/i.test(r.url));
  for (const r of tmxLike) console.log('  '+r.method+' '+r.url.slice(0,300));

  console.log('\n=== All non-Microsoft third-party hosts ===');
  const msfp = /(\.microsoft\.com|\.msauth\.net|\.msftauth\.net|\.live\.com|\.outlook\.com|\.office\.com|\.bing\.com|\.windows\.net|\.azureedge\.net|fonts\.gstatic\.com|aria-error\.azure-api\.net)$/i;
  function host(u){ try { return new URL(u).host.toLowerCase(); } catch { return ''; } }
  const tp = {};
  for (const r of reqs) { const h = host(r.url); if (h && !msfp.test('.'+h)) tp[h] = (tp[h]||0)+1; }
  for (const [h,n] of Object.entries(tp).sort((a,b)=>b[1]-a[1])) console.log('  '+n+' '+h);

  console.log('\n=== ALL Microsoft-domain endpoints (filter for fingerprint-ish) ===');
  const interesting = reqs.filter(r => {
    const u = r.url.toLowerCase();
    return /\.microsoft\.com|\.live\.com/.test(u) && /clear|fingerprint|probe|track|telemetry|risk|fraud|fpt|cfp|df\.|tdid|tags\.js|uaid|sensor|verify|prove/.test(u);
  });
  for (const r of interesting.slice(0,40)) console.log('  '+r.method+' '+r.url.slice(0,250));

  console.log('\n=== Totals ===');
  console.log('  total requests:', reqs.length);
  console.log('  cfp.microsoft.com requests:', cfp.length);
})().catch(e => { console.error(e); process.exit(1); });

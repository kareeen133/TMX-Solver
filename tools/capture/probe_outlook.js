#!/usr/bin/env node
// Render outlook/login.live.com in real headed Chromium, simulate typing an
// email, and dump every network request — to expose any lazy-loaded TMX or
// other anti-fraud SDK that static HTML scraping misses.

const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: false });
  const ctx = await browser.newContext();
  const page = await ctx.newPage();
  const reqs = [];
  page.on('request', r => reqs.push({ method: r.method(), url: r.url(), type: r.resourceType() }));
  page.on('response', r => {});

  async function probe(label, url) {
    console.log('\n=== ' + label + ' ===');
    try {
      await page.goto(url, { waitUntil: 'networkidle', timeout: 25000 }).catch(e => console.log('  goto:', e.message));
    } catch (e) { console.log('  err:', e.message); }
    await page.waitForTimeout(2500);

    // Try to find an email input and type something
    const emailSel = ['input[type=email]', 'input[name="loginfmt"]', 'input[name="email"]', 'input#i0116', 'input#emailField'];
    for (const s of emailSel) {
      const el = await page.$(s).catch(()=>null);
      if (el) {
        await el.fill('test@example.com').catch(()=>{});
        console.log('  typed into', s);
        await page.waitForTimeout(2000);
        // Click "Next" if visible
        for (const btn of ['#idSIButton9', 'input[type=submit]', 'button[type=submit]']) {
          const b = await page.$(btn).catch(()=>null);
          if (b) { await b.click().catch(()=>{}); console.log('  clicked', btn); break; }
        }
        await page.waitForTimeout(3500);
        break;
      }
    }
  }

  await probe('outlook.live.com', 'https://outlook.live.com/');
  await probe('login.live.com', 'https://login.live.com/');
  await probe('signup.live.com', 'https://signup.live.com/');

  await browser.close();

  // Filter for anti-fraud SDK candidates
  const interesting = reqs.filter(r => {
    const u = r.url.toLowerCase();
    return /\/fp\/(tags|check|clear|mobile)/i.test(u) ||
           /(online-metrix|threatmetrix|metrixsec|riskifier|kasada|arkose|forter|sift|akamai-bm|imperva|datadome|perimeterx|signifyd|emailage|cdn-cgi\/(challenge|chl)|f\.[a-z]+\.(com|net)\/[a-z0-9]{4,12}\.js)/i.test(u);
  });
  // Find third-party domains (not microsoft/msft/office/live/bing CDNs)
  const msfp = /(\.microsoft\.com|\.msauth\.net|\.msftauth\.net|\.live\.com|\.outlook\.com|\.office\.com|\.bing\.com|\.office365\.com|\.windows\.net|\.azureedge\.net|\.azure\.com|\.windows\.com|fonts\.gstatic\.com)$/i;
  function host(u){ try { return new URL(u).host.toLowerCase(); } catch { return ''; } }
  const thirdParty = {};
  for (const r of reqs) { const h = host(r.url); if (h && !msfp.test('.'+h)) thirdParty[h] = (thirdParty[h]||0)+1; }

  console.log('\n=== TMX-or-antifraud candidates ===');
  if (interesting.length === 0) console.log('  (none)');
  for (const r of interesting.slice(0,30)) console.log('  '+r.method+' '+r.url.slice(0,180));

  console.log('\n=== Third-party hosts (non-microsoft) ===');
  if (Object.keys(thirdParty).length === 0) console.log('  (none — all traffic to microsoft domains)');
  for (const [h,n] of Object.entries(thirdParty).sort((a,b)=>b[1]-a[1]).slice(0,20)) console.log('  '+n+' '+h);

  console.log('\n=== Total requests ===');
  console.log('  '+reqs.length);
})().catch(e => { console.error(e); process.exit(1); });

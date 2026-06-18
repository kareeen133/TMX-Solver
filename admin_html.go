package main

const adminHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>TMX Solver — Admin</title>
<style>
:root{
  --bg:#0b0d12; --panel:#11141b; --panel2:#161a23; --line:#222838;
  --text:#e6e8ee; --muted:#8b93a7; --accent:#6ea8ff; --good:#4ade80; --bad:#f87171; --warn:#fbbf24;
  --mono:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;
}
*{box-sizing:border-box}
html,body{margin:0;padding:0;background:var(--bg);color:var(--text);font:14px/1.5 system-ui,-apple-system,Segoe UI,Roboto,sans-serif}
a{color:var(--accent);text-decoration:none}
a:hover{text-decoration:underline}

/* login screen */
.login-wrap{min-height:100vh;display:flex;align-items:center;justify-content:center;padding:20px}
.login-card{background:var(--panel);border:1px solid var(--line);border-radius:14px;padding:32px;width:100%;max-width:380px;box-shadow:0 24px 60px rgba(0,0,0,.5)}
.login-card h1{margin:0 0 6px;font-size:20px;letter-spacing:.3px}
.login-card .sub{color:var(--muted);font-size:13px;margin-bottom:22px}
.login-card label{display:block;font-size:12px;color:var(--muted);margin-bottom:6px;text-transform:uppercase;letter-spacing:.5px}
.login-card input{width:100%;background:var(--panel2);color:var(--text);border:1px solid var(--line);padding:10px 12px;border-radius:8px;font:13px var(--mono);margin-bottom:14px}
.login-card input:focus{outline:none;border-color:var(--accent)}
.login-card button{width:100%;background:var(--accent);color:#0b0d12;border:none;padding:11px 14px;border-radius:8px;cursor:pointer;font-weight:600;font-size:14px;margin-top:6px}
.login-card button:hover{filter:brightness(1.08)}
.login-card .err{color:var(--bad);font-size:12px;margin-top:10px;min-height:16px}
.login-card .docs-link{display:block;text-align:center;margin-top:18px;color:var(--muted);font-size:12px}

/* dashboard */
header{display:flex;align-items:center;gap:16px;padding:14px 22px;border-bottom:1px solid var(--line);background:var(--panel)}
header h1{font-size:16px;margin:0;font-weight:600;letter-spacing:.3px}
header .who{font:12px var(--mono);color:var(--muted);background:var(--panel2);padding:4px 8px;border-radius:6px;border:1px solid var(--line)}
header .spacer{flex:1}
header button,header a.btn{background:var(--panel2);color:var(--text);border:1px solid var(--line);padding:6px 12px;border-radius:6px;cursor:pointer;font-size:13px;text-decoration:none;display:inline-block}
header button:hover,header a.btn:hover{border-color:var(--accent);color:var(--accent)}
.stats{display:grid;grid-template-columns:repeat(6,1fr);gap:12px;padding:18px 22px}
.stat{background:var(--panel);border:1px solid var(--line);border-radius:10px;padding:14px}
.stat .label{font-size:11px;color:var(--muted);text-transform:uppercase;letter-spacing:.5px}
.stat .value{font-size:22px;font-weight:600;margin-top:4px}
.stat.good .value{color:var(--good)}
.stat.bad .value{color:var(--bad)}
.stat.warn .value{color:var(--warn)}
.toolbar{display:flex;gap:10px;align-items:center;padding:0 22px 14px}
.toolbar input[type=text]{background:var(--panel);color:var(--text);border:1px solid var(--line);padding:8px 12px;border-radius:8px;flex:1;max-width:480px;font:13px var(--mono)}
.toolbar input[type=text]:focus{outline:none;border-color:var(--accent)}
.toolbar select,.toolbar button{background:var(--panel);color:var(--text);border:1px solid var(--line);padding:8px 12px;border-radius:8px;cursor:pointer;font-size:13px}
.toolbar button:hover{border-color:var(--accent);color:var(--accent)}
table{width:calc(100% - 44px);margin:0 22px;border-collapse:collapse;background:var(--panel);border:1px solid var(--line);border-radius:10px;overflow:hidden}
th,td{padding:10px 12px;text-align:left;border-bottom:1px solid var(--line);font-size:13px}
th{background:var(--panel2);color:var(--muted);font-weight:500;text-transform:uppercase;font-size:11px;letter-spacing:.5px}
tr:last-child td{border-bottom:none}
tr.row{cursor:pointer;transition:background .12s}
tr.row:hover{background:var(--panel2)}
.pill{display:inline-block;padding:2px 8px;border-radius:999px;font-size:11px;font-weight:500;font-family:var(--mono)}
.pill.ok{background:rgba(74,222,128,.12);color:var(--good)}
.pill.bad{background:rgba(248,113,113,.12);color:var(--bad)}
.pill.warn{background:rgba(251,191,36,.12);color:var(--warn)}
.pill.muted{background:rgba(139,147,167,.12);color:var(--muted)}
.mono{font-family:var(--mono);font-size:12px}
.short{max-width:280px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;display:inline-block;vertical-align:middle}
.empty{padding:48px 22px;text-align:center;color:var(--muted)}
.modal{position:fixed;inset:0;background:rgba(0,0,0,.7);display:none;align-items:center;justify-content:center;z-index:10}
.modal.open{display:flex}
.modal .card{background:var(--panel);border:1px solid var(--line);border-radius:12px;width:min(900px,92vw);max-height:88vh;overflow:auto;padding:20px}
.modal h2{margin:0 0 12px;font-size:15px}
.modal pre{background:var(--bg);border:1px solid var(--line);border-radius:8px;padding:12px;overflow:auto;font:12px var(--mono);max-height:400px;color:#cdd3e0}
.modal .grid{display:grid;grid-template-columns:140px 1fr;gap:6px 14px;font-size:13px;margin-bottom:14px}
.modal .grid .k{color:var(--muted)}
.modal .grid .v{font-family:var(--mono);font-size:12px;word-break:break-all}
.modal .close{float:right;background:transparent;border:none;color:var(--muted);font-size:20px;cursor:pointer}
.modal .close:hover{color:var(--text)}
.foot{padding:12px 22px;color:var(--muted);font-size:12px;display:flex;gap:14px;align-items:center}
.foot .right{margin-left:auto;display:flex;gap:8px}
.foot button{background:var(--panel);color:var(--text);border:1px solid var(--line);padding:5px 10px;border-radius:6px;cursor:pointer;font-size:12px}
.foot button:disabled{opacity:.4;cursor:not-allowed}
@media(max-width:1100px){.stats{grid-template-columns:repeat(3,1fr)}}
.hidden{display:none!important}
</style>
</head>
<body>

<div id="loginScreen" class="login-wrap">
  <form class="login-card" id="loginForm" onsubmit="return doLogin(event)">
    <h1>TMX Solver Admin</h1>
    <div class="sub">Sign in to access the dashboard</div>
    <label for="u">Username</label>
    <input type="text" id="u" autocomplete="username" required>
    <label for="p">Password</label>
    <input type="password" id="p" autocomplete="current-password" required>
    <button type="submit">Sign in</button>
    <div class="err" id="loginErr"></div>
    <a class="docs-link" href="/docs" target="_blank">Read the API docs →</a>
  </form>
</div>

<div id="app" class="hidden">
<header>
  <h1>TMX Solver — Admin</h1>
  <span class="who" id="whoami">admin</span>
  <span class="spacer"></span>
  <a class="btn" href="/docs" target="_blank">Docs</a>
  <button id="keysBtn">API Keys</button>
  <label style="font-size:12px;color:var(--muted)">
    <input type="checkbox" id="autoRefresh" checked> auto 5s
  </label>
  <button id="refreshBtn">Refresh</button>
  <button id="clearLogsBtn">Clear logs</button>
  <button id="logoutBtn">Logout</button>
</header>

<div class="stats" id="stats"></div>

<div class="toolbar">
  <input type="text" id="search" placeholder="filter by org_id, host, target, session_id, ip, error...">
  <select id="pageSize">
    <option value="25">25 / page</option>
    <option value="50" selected>50 / page</option>
    <option value="100">100 / page</option>
    <option value="250">250 / page</option>
  </select>
  <button id="clearBtn">Clear filter</button>
</div>

<table>
  <thead>
    <tr>
      <th>Time</th><th>Endpoint</th><th>Org / Host</th><th>Target</th><th>Method</th>
      <th>TMX</th><th>Status</th><th>Solve</th><th>Upstream</th><th>Body</th><th>Client</th>
    </tr>
  </thead>
  <tbody id="rows"><tr><td colspan="11" class="empty">Loading…</td></tr></tbody>
</table>

<div class="foot">
  <span id="pageInfo">—</span>
  <span class="right">
    <button id="prev">← Prev</button>
    <button id="next">Next →</button>
  </span>
</div>

<div class="modal" id="keysModal">
  <div class="card" style="max-width:920px">
    <button class="close" onclick="document.getElementById('keysModal').classList.remove('open')">×</button>
    <h2>API Keys &amp; Credits</h2>
    <div class="toolbar" style="padding:0 0 14px;margin:0">
      <input type="text" id="kLabel" placeholder="label / customer name" style="flex:1">
      <input type="number" id="kCredits" placeholder="credits" value="100" style="width:120px">
      <button id="kCreate">Create key</button>
    </div>
    <table style="width:100%;margin:0">
      <thead><tr><th>Label</th><th>Key</th><th>Credits</th><th>Used</th><th>State</th><th>Actions</th></tr></thead>
      <tbody id="keyRows"><tr><td colspan="6" class="empty">Loading…</td></tr></tbody>
    </table>
  </div>
</div>

<div class="modal" id="modal">
  <div class="card">
    <button class="close" onclick="closeModal()">×</button>
    <h2>Request detail</h2>
    <div class="grid" id="detailGrid"></div>
    <h3 style="font-size:13px;color:var(--muted);margin:14px 0 6px">Request body</h3>
    <pre id="reqBody">—</pre>
    <h3 style="font-size:13px;color:var(--muted);margin:14px 0 6px">Response body (head)</h3>
    <pre id="respBody">—</pre>
  </div>
</div>
</div>

<script>
let offset = 0, limit = 50, timer = null;

async function api(path, opts){
  const r = await fetch(path, Object.assign({credentials:'same-origin'}, opts||{}));
  if(r.status === 401){ showLogin(); throw new Error('unauthorized'); }
  if(!r.ok) throw new Error('HTTP '+r.status);
  return r.json();
}

function showLogin(){
  document.getElementById('app').classList.add('hidden');
  document.getElementById('loginScreen').classList.remove('hidden');
}
function showApp(){
  document.getElementById('loginScreen').classList.add('hidden');
  document.getElementById('app').classList.remove('hidden');
  loadStats(); loadRows();
  if(!timer) timer = setInterval(() => { if(document.getElementById('autoRefresh').checked){ loadStats(); loadRows(); } }, 5000);
}

async function doLogin(e){
  e.preventDefault();
  const u = document.getElementById('u').value;
  const p = document.getElementById('p').value;
  const err = document.getElementById('loginErr');
  err.textContent = '';
  try{
    const r = await fetch('/admin/login', {
      method:'POST',
      headers:{'Content-Type':'application/json'},
      credentials:'same-origin',
      body: JSON.stringify({username:u, password:p})
    });
    if(!r.ok){
      const j = await r.json().catch(()=>({error:'login failed'}));
      err.textContent = j.error || 'login failed';
      return false;
    }
    document.getElementById('whoami').textContent = u;
    showApp();
  }catch(ex){ err.textContent = ex.message; }
  return false;
}

async function doLogout(){
  await fetch('/admin/logout', {method:'POST', credentials:'same-origin'});
  if(timer){ clearInterval(timer); timer = null; }
  showLogin();
}

function fmtTime(s){ if(!s) return '—'; const d=new Date(s); return d.toLocaleTimeString()+' · '+d.toLocaleDateString(); }
function pill(t,k){ return '<span class="pill '+k+'">'+t+'</span>'; }

async function loadStats(){
  try{
    const s = await api('/admin/api/stats');
    const el = document.getElementById('stats');
    el.innerHTML = [
      ['Total','muted',s.total],
      ['OK','good',s.ok],
      ['Failed','bad',s.fail],
      ['TMX accepted','good',s.tmx_accepted],
      ['Avg solve','warn',s.avg_solve_ms+' ms'],
      ['Avg upstream','warn',s.avg_upstream_ms+' ms'],
    ].map(([k,c,v])=>'<div class="stat '+c+'"><div class="label">'+k+'</div><div class="value">'+v+'</div></div>').join('');
  }catch(e){}
}

async function loadRows(){
  try{
    const q = document.getElementById('search').value.trim();
    const data = await api('/admin/api/requests?limit='+limit+'&offset='+offset+'&q='+encodeURIComponent(q));
    const tb = document.getElementById('rows');
    if(!data.items || !data.items.length){
      tb.innerHTML = '<tr><td colspan="11" class="empty">No requests yet</td></tr>';
    } else {
      tb.innerHTML = data.items.map(r => {
        const tmx = r.tmx_accepted ? pill('accepted','ok') : pill('—','muted');
        const status = r.status ? pill(r.status, r.ok?'ok':'bad') : pill('—','muted');
        return '<tr class="row" data-id="'+r.id+'">'+
          '<td class="mono">'+fmtTime(r.time)+'</td>'+
          '<td class="mono">'+r.endpoint+'</td>'+
          '<td><div class="mono">'+(r.org_id||'—')+'</div><div class="mono" style="color:var(--muted);font-size:11px">'+(r.host||'')+'</div></td>'+
          '<td><span class="short mono" title="'+(r.target||'')+'">'+(r.target||'—')+'</span></td>'+
          '<td class="mono">'+(r.method||'—')+'</td>'+
          '<td>'+tmx+'</td>'+
          '<td>'+status+'</td>'+
          '<td class="mono">'+(r.solve_ms||0)+' ms</td>'+
          '<td class="mono">'+(r.upstream_ms||0)+' ms</td>'+
          '<td class="mono">'+(r.body_len||0)+' B</td>'+
          '<td class="mono" style="color:var(--muted)">'+(r.client_ip||'—')+'</td>'+
        '</tr>';
      }).join('');
      tb.querySelectorAll('tr.row').forEach(tr => tr.addEventListener('click', () => openDetail(tr.dataset.id)));
    }
    document.getElementById('pageInfo').textContent = 'Showing '+(data.items.length?offset+1:0)+'–'+(offset+data.items.length)+' of '+data.total;
    document.getElementById('prev').disabled = offset === 0;
    document.getElementById('next').disabled = offset + data.items.length >= data.total;
  }catch(e){}
}

async function openDetail(id){
  try{
    const r = await api('/admin/api/request?id='+encodeURIComponent(id));
    const grid = document.getElementById('detailGrid');
    const fields = [
      ['ID', r.id], ['Time', r.time], ['Endpoint', r.endpoint], ['Client IP', r.client_ip],
      ['Org ID', r.org_id], ['Host', r.host], ['Target', r.target], ['Method', r.method],
      ['Deep mode', r.deep ? 'yes' : 'no'], ['Session ID', r.session_id || '—'],
      ['TMX accepted', r.tmx_accepted ? 'yes' : 'no'], ['Status', r.status||'—'],
      ['OK', r.ok ? 'yes' : 'no'], ['Solve', (r.solve_ms||0)+' ms'],
      ['Upstream', (r.upstream_ms||0)+' ms'], ['Body length', (r.body_len||0)+' bytes'],
    ];
    if(r.error) fields.push(['Error', r.error]);
    grid.innerHTML = fields.map(([k,v]) => '<div class="k">'+k+'</div><div class="v">'+(v===''||v==null?'—':String(v))+'</div>').join('');
    document.getElementById('reqBody').textContent = r.req_body || '—';
    document.getElementById('respBody').textContent = r.response_head || '—';
    document.getElementById('modal').classList.add('open');
  }catch(e){ alert('Error: '+e.message); }
}
function closeModal(){ document.getElementById('modal').classList.remove('open'); }
document.getElementById('modal').addEventListener('click', e => { if(e.target.id==='modal') closeModal(); });
document.getElementById('refreshBtn').addEventListener('click', () => { loadStats(); loadRows(); });
document.getElementById('logoutBtn').addEventListener('click', doLogout);
document.getElementById('clearLogsBtn').addEventListener('click', async () => {
  if(!confirm('Clear ALL request logs? This wipes the in-memory buffer and the log file. Cannot be undone.')) return;
  try{
    const d = await api('/admin/api/clear', {method:'POST'});
    offset = 0;
    loadStats(); loadRows();
    alert('Cleared '+(d.cleared||0)+' log records.');
  }catch(e){ alert('Error: '+e.message); }
});
document.getElementById('search').addEventListener('input', () => { offset = 0; loadRows(); });
document.getElementById('clearBtn').addEventListener('click', () => { document.getElementById('search').value=''; offset=0; loadRows(); });
document.getElementById('pageSize').addEventListener('change', e => { limit = parseInt(e.target.value); offset = 0; loadRows(); });
document.getElementById('prev').addEventListener('click', () => { offset = Math.max(0, offset-limit); loadRows(); });
document.getElementById('next').addEventListener('click', () => { offset += limit; loadRows(); });

function copyText(t){ navigator.clipboard.writeText(t); }
async function loadKeys(){
  try{
    const d = await api('/admin/api/keys');
    const tb = document.getElementById('keyRows');
    const ks = (d.keys||[]).sort((a,b)=>b.created_at-a.created_at);
    if(!ks.length){ tb.innerHTML='<tr><td colspan="6" class="empty">No keys yet</td></tr>'; return; }
    tb.innerHTML = ks.map(k=>{
      const st = k.disabled? pill('disabled','bad') : pill('active','ok');
      return '<tr><td>'+(k.label||'—')+'</td>'+
        '<td><code style="cursor:pointer" title="click to copy" onclick="copyText(\''+k.key+'\')">'+k.key+'</code></td>'+
        '<td><b>'+k.credits+'</b></td><td>'+k.used+'</td><td>'+st+'</td>'+
        '<td>'+
          '<button onclick="adjKey(\''+k.key+'\',1)">+1</button> '+
          '<button onclick="adjKey(\''+k.key+'\',10)">+10</button> '+
          '<button onclick="adjKey(\''+k.key+'\',100)">+100</button> '+
          '<button onclick="adjKey(\''+k.key+'\',-1)">-1</button> '+
          '<button onclick="promptAdj(\''+k.key+'\')">±N</button> '+
          '<button onclick="toggleKey(\''+k.key+'\','+(!k.disabled)+')">'+(k.disabled?'enable':'disable')+'</button> '+
          '<button onclick="delKey(\''+k.key+'\')">del</button>'+
        '</td></tr>';
    }).join('');
  }catch(e){}
}
async function adjKey(key,delta){ await api('/admin/api/keys/adjust',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({key,delta})}); loadKeys(); }
function promptAdj(key){ const v=prompt('Credit delta (negative to deduct):','0'); if(v===null)return; const n=parseInt(v); if(!isNaN(n)) adjKey(key,n); }
async function toggleKey(key,disabled){ await api('/admin/api/keys/toggle',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({key,disabled})}); loadKeys(); }
async function delKey(key){ if(!confirm('Delete this key?'))return; await api('/admin/api/keys/delete',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({key})}); loadKeys(); }
document.getElementById('keysBtn').addEventListener('click', ()=>{ document.getElementById('keysModal').classList.add('open'); loadKeys(); });
document.getElementById('kCreate').addEventListener('click', async ()=>{
  const label=document.getElementById('kLabel').value;
  const credits=parseInt(document.getElementById('kCredits').value)||0;
  const k=await api('/admin/api/keys/create',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({label,credits})});
  document.getElementById('kLabel').value='';
  loadKeys();
  alert('Created key (click to copy from table):\n\n'+k.key);
});

(async function(){
  try{
    await api('/admin/api/stats');
    showApp();
  }catch(e){ showLogin(); }
})();
</script>
</body>
</html>`

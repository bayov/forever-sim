// node checklist.js <harness prefix> <mode> -> JSON of feature counts per checklist item
const { chromium } = require('playwright');
const path = require('path');
const [, , hp = 'o', mode = 'plain'] = process.argv;
(async () => {
  const browser = await chromium.launch({ executablePath: process.env.PW_CHROME || '/opt/pw-browsers/chromium-1194/chrome-linux/chrome', args: ['--allow-file-access-from-files'] });
  const ctx = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 1, isMobile: true, hasTouch: true });
  const page = await ctx.newPage();
  const errs = [];
  page.on('pageerror', (e) => errs.push(e.message));
  await page.goto('file://' + path.join(__dirname, `${hp}_${mode}.html`));
  await page.waitForTimeout(600);
  const R = {};
  const q = (fn) => page.evaluate(fn);
  const tab = async (t) => { await page.click(`[data-tab="${t}"]`); await page.waitForTimeout(250); };
  const C = (sel) => `document.querySelectorAll(${JSON.stringify(sel)}).length`;
  const imgs = (sel) => `[...document.querySelectorAll(${JSON.stringify(sel + ' img')})].filter(i=>i.src.startsWith('data:')&&i.naturalWidth>0).length`;
  const txt = (sel) => `(document.querySelector(${JSON.stringify(sel)})||{}).textContent?.trim().length||0`;
  const run = async (items) => { for (const [k, js] of items) R[k] = await q(js); };
  const tipProbe = async (name, sel, i = 0) => {
    const ok = await q(`(()=>{const n=document.querySelectorAll(${JSON.stringify(sel)})[${i}]; if(!n) return 'NO TARGET'; n.scrollIntoView({block:'center'}); const r=n.getBoundingClientRect(); n.dispatchEvent(new MouseEvent('click',{bubbles:true,cancelable:true,clientX:r.left+Math.min(8,r.width/2),clientY:r.top+r.height/2,view:window})); const t=document.querySelector('.fv-tt'); const res = t ? (t.innerText.trim().length>0 ? 'tip ok' + (t.querySelector('img')?' +icon':'') : 'tip EMPTY') : 'NO TIP'; document.body.click(); return res;})()`);
    R['tooltip: ' + name] = ok;
  };
  // header + tabs
  await run([
    ['header: title', txt('.fv h1')], ['header: class crests (img)', imgs('#hero')], ['header: takeaway text', txt('#takeaway')],
    ['tabs: count', C('.fv-tab')],
  ]);
  // ranking
  await run([
    ['ranking: KPI values filled', `[...document.querySelectorAll('#ranking [data-source="overview"]')].filter(n=>/\\d/.test(n.textContent)).length`],
    ['ranking: class chips', C('#rank-classes button')], ['ranking: class chip icons', imgs('#rank-classes')],
    ['ranking: sort control', C('#rank-sort')], ['ranking: length control', C('#rank-len')], ['ranking: defaults toggle', C('#rank-defaults')],
    ['ranking: legend', C('#rank-legend span')],
    ['ranking: rows', C('#rank-chart .fv-row')], ['ranking: row icons', imgs('#rank-chart')], ['ranking: box/whisker marks', C('#rank-chart .fv-box')],
    ['ranking: scatter dots', C('#scatter circle')], ['ranking: length lines', C('#lenchart path')],
    ['ranking: class table rows', C('#class-table tbody tr')], ['ranking: class table icons', imgs('#class-table')],
  ]);
  await page.click('#rank-classes button:nth-child(2)'); await page.waitForTimeout(200);
  R['ranking: class filter works (rows)'] = await q(C('#rank-chart .fv-row'));
  await page.click('#rank-classes button:nth-child(1)'); await page.waitForTimeout(200);
  await q(`(()=>{const s=document.querySelector('#rank-len'); if(s.tagName==='SELECT'){s.value='60';s.dispatchEvent(new Event('change'))} else s.querySelector('[data-v="60"]').click()})()`); await page.waitForTimeout(200);
  R['ranking: 60 s length works (rows)'] = await q(C('#rank-chart .fv-row'));
  await q(`(()=>{const s=document.querySelector('#rank-len'); if(s.tagName==='SELECT'){s.value='180';s.dispatchEvent(new Event('change'))} else s.querySelector('[data-v="180"]').click()})()`); await page.waitForTimeout(200);
  await q(`document.querySelectorAll('#scatter circle')[3].dispatchEvent(new MouseEvent('click',{bubbles:true}))`); await page.waitForTimeout(100);
  R['ranking: scatter tip'] = await q(C('#scatter .fv-tip'));
  await tipProbe('class crest (hover target)', '#hero [data-tt]');
  // races
  await tab('races');
  await run([
    ['races: class chips', C('#heat-classes button')], ['races: heat cells', C('#heat .hc:not(.na)')], ['races: heat race icons', imgs('#heat')],
    ['races: faction rows', C('#faction-table tbody tr')], ['races: faction icons', imgs('#faction-table')], ['races: race spread rows', C('#racespread .tr')],
  ]);
  await q(`document.querySelector('#heat .hc:not(.na)').click()`); await page.waitForTimeout(100);
  R['races: heat cell tip'] = await q(C('#heat-card .fv-tip'));
  await tipProbe('race header', '#heat [data-tt^="r:"]');
  // spec
  await tab('spec');
  await q(`document.querySelectorAll('.fv details').forEach(d=>d.open=true)`); await page.waitForTimeout(300);
  await run([
    ['spec: picker options', C('#sp-pick option')], ['spec: prev/next', C('#sp-prev,#sp-next')], ['spec: spec icon', imgs('#sp-specicon')],
    ['spec: banner art', `!!document.querySelector('#sp-art').style.backgroundImage`], ['spec: meta icons', imgs('#sp-meta')], ['spec: summary', txt('#sp-summary')],
    ['spec: gear round line', txt('#sp-round')], ['spec: tiles', C('#sp-tiles .fv-tile')],
    ['spec: talent tree tabs', C('#sp-tal-seg button')], ['spec: talent buttons', C('#sp-tal .fv-tal')], ['spec: talent icons', imgs('#sp-tal')], ['spec: talent art', `[...document.querySelectorAll('#sp-tal .art')].filter(a=>a.style.backgroundImage).length`],
    ['spec: DPS spread bars', C('#sp-dist rect')], ['spec: percentile cells', C('#sp-pct td')],
    ['spec: ability rows', C('#sp-abil .tr')], ['spec: ability icons', imgs('#sp-abil')],
    ['spec: stat weight rows', C('#sp-weights .tr')], ['spec: weights table rows', C('#sp-wtab tbody tr')],
    ['spec: timeline sample buttons', C('#tl-samples button')], ['spec: timeline zoom buttons', C('#tl-zoom button')],
    ['spec: timeline svg rects', C('#tl-scroll rect')], ['spec: timeline icons', C('#tl-scroll image')], ['spec: timeline lane labels', C('#tl-labels div')], ['spec: timeline cooldown marks', C('#tl-scroll path[data-tt]')],
    ['spec: opener chips', C('#tl-opener [data-tt]')], ['spec: opener icons', imgs('#tl-opener')],
    ['spec: event log rows', C('#tl-log [data-tt]')],
    ['spec: aura rows', C('#sp-auras .tr')], ['spec: aura icons', imgs('#sp-auras')],
    ['spec: races chart rows', C('#sp-races g[data-tt]')], ['spec: races chart icons', C('#sp-races image')],
    ['spec: length chart dots', C('#sp-len circle')],
    ['spec: talent string', txt('#sp-talents')], ['spec: key talents', C('#sp-keytal .fv-tag')], ['spec: rotation steps', C('#sp-rotation li')],
    ['spec: per-ability table rows', C('#sp-abiltab tbody tr')], ['spec: per-ability icons', imgs('#sp-abiltab')],
    ['spec: character sheet rows', C('#sp-stats tbody tr')], ['spec: gear rows', C('#sp-gear tbody tr')], ['spec: gear icons', imgs('#sp-gear')],
    ['spec: setup chips', C('#sp-setup .fv-tag')], ['spec: setup icons', imgs('#sp-setup')], ['spec: caveats', C('#sp-caveats li')],
  ]);
  // timeline interaction
  await q(`(()=>{const sv=document.querySelector('#tl-scroll svg'); const r=sv.getBoundingClientRect(); sv.dispatchEvent(new MouseEvent('click',{bubbles:true,clientX:r.left+120,clientY:r.top+60}))})()`); await page.waitForTimeout(100);
  R['spec: timeline tap readout events'] = await q(C('#tl-readout .ev'));
  await q(`document.querySelector('#tl-zoom [data-z="15"]').click()`); await page.waitForTimeout(200);
  R['spec: timeline zoom 15 s width'] = await q(`Math.round(document.querySelector('#tl-scroll svg').getBoundingClientRect().width)`);
  await q(`document.querySelector('#tl-zoom [data-z="fit"]').click()`); await page.waitForTimeout(200);
  await q(`document.querySelectorAll('#tl-samples button')[1].click()`); await page.waitForTimeout(200);
  R['spec: unlucky sample rects'] = await q(C('#tl-scroll rect'));
  await q(`[...document.querySelectorAll('#sp-tal-seg button')].pop().click()`); await page.waitForTimeout(150);
  R['spec: all three trees'] = await q(C('#sp-tal .fv-tree'));
  await tipProbe('ability row', '#sp-abil .nm[data-tt]');
  await tipProbe('talent', '#sp-tal .fv-tal', 3);
  await tipProbe('timeline cast', '#tl-scroll rect[data-tt]', 5);
  await tipProbe('timeline aura lane', '#tl-scroll g[data-tt]');
  await tipProbe('opener chip', '#tl-opener [data-tt]', 2);
  await tipProbe('aura row', '#sp-auras .nm[data-tt]');
  await tipProbe('gear row', '#sp-gear tbody tr[data-tt]', 13);
  await tipProbe('setup chip', '#sp-setup [data-tt]');
  await tipProbe('stat weight', '#sp-weights [data-tt]');
  await tipProbe('race (meta)', '#sp-meta [data-tt^="r:"]');
  await tipProbe('races chart', '#sp-races g[data-tt]', 1);
  // pet class
  await q(`(()=>{const s=document.querySelector('#sp-pick');s.value='warlock__demonology';s.dispatchEvent(new Event('change'))})()`); await page.waitForTimeout(300);
  R['spec: warlock lanes'] = await q(C('#tl-labels div'));
  R['spec: warlock pet tile'] = await q(`[...document.querySelectorAll('#sp-tiles .k')].map(k=>k.textContent).join('|')`);
  await q(`document.querySelector('#sp-next').click()`); await page.waitForTimeout(250);
  R['spec: next preset'] = await q(`document.querySelector('#sp-title').textContent`);
  // compare
  await tab('compare');
  await run([
    ['compare: pickers', C('#cmp-a option') + '+' + C('#cmp-b option')], ['compare: table rows', C('#cmp-table tbody tr')], ['compare: header icons', imgs('#cmp-table thead')],
    ['compare: spread bars', C('#cmp-dist rect')], ['compare: length dots', C('#cmp-len circle')], ['compare: weights cols', C('#cmp-w .fv-sbar')], ['compare: weights rows', C('#cmp-w .tr')],
    ['compare: damage cols', C('#cmp-abil .fv-sbar')], ['compare: damage rows', C('#cmp-abil .tr')],
  ]);
  await tipProbe('compare ability', '#cmp-abil [data-tt]', 1);
  // method
  await tab('method');
  await run([
    ['method: prose numbers filled', `[...document.querySelectorAll('#method-prose [data-source]')].filter(n=>/\\d/.test(n.textContent)).length`],
    ['method: gear rules', C('#method-gear li')], ['method: before/after rows', C('#round-table tbody tr')], ['method: repro rows', C('#repro-table tbody tr')], ['method: caveat groups', C('#method-caveats details')],
  ]);
  R['no horizontal page scroll (docW)'] = await q('document.documentElement.scrollWidth');
  R['script errors'] = errs.length + (await q('window.__errs.length'));
  console.log(JSON.stringify(R));
  await browser.close();
})().catch((e) => { console.error(e); process.exit(1); });

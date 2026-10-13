// node races_check.js <harness prefix> <mode>: the new Races tab features in every scenario
const { chromium } = require('playwright');
const path = require('path');
const [, , hp = 'h', mode = 'plain'] = process.argv;
(async () => {
  const browser = await chromium.launch({ executablePath: process.env.PW_CHROME || '/opt/pw-browsers/chromium-1194/chrome-linux/chrome', args: ['--allow-file-access-from-files'] });
  const ctx = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 1, isMobile: true, hasTouch: true });
  const page = await ctx.newPage();
  const errs = [];
  page.on('pageerror', (e) => errs.push(e.message));
  await page.goto('file://' + path.join(__dirname, `${hp}_${mode}.html`));
  await page.waitForTimeout(600);
  const q = (js) => page.evaluate(js);
  const C = (s) => `document.querySelectorAll(${JSON.stringify(s)}).length`;
  const R = {};
  await page.click('[data-tab="races"]'); await page.waitForTimeout(300);
  const snap = async (tag) => {
    R[tag] = await q(`({desc: (document.querySelector('#rs-desc')||{}).textContent?.slice(0,40), tag: document.querySelector('.fv-scen-tag').textContent,
      tiles: ${C('#ro-tiles .fv-tile')}, tileImgs: ${C('#ro-tiles img')}, overview: ${C('#race-overview .fv-orow')}, specRows: ${C('#race-spec .fv-rrow')}, segs: ${C('#race-spec .sg')},
      legend: ${C('#rs-legend span')}, take: (document.querySelector('#rs-take')||{}).textContent?.slice(0,60), heat: ${C('#heat .hc:not(.na)')}, stars: [...document.querySelectorAll('#heat .hc')].filter(c=>c.textContent==='★').length,
      find: ${C('#race-find > div')}, faction: ${C('#faction-table tbody tr')}, spread: ${C('#racespread .tr')}, fails: ${C('#races .fv-fail')}})`);
  };
  await snap('baseline');
  for (const [k, v] of [['rs-target', 'beast'], ['rs-target', 'elemental'], ['rs-target', 'undead'], ['rs-weapon', 'weapon'], ['rs-weapon', 'off']]) {
    await q(`document.querySelector('#${k} [data-v="${v}"]').click()`); await page.waitForTimeout(200);
    await snap(v);
  }
  R['weapon resets target'] = await q(`document.querySelector('#rs-target [aria-pressed="true"]').dataset.v`);
  await q(`document.querySelector('#rs-target [data-v="beast"]').click()`); await page.waitForTimeout(150);
  R['target resets weapon'] = await q(`document.querySelector('#rs-weapon [aria-pressed="true"]').dataset.v`);
  await q(`document.querySelector('#rs-target [data-v="humanoid"]').click()`); await page.waitForTimeout(150);
  // tooltips
  const tap = async (name, sel, i = 0) => {
    const h = (await page.$$(sel))[i];
    if (!h) { R['tip ' + name] = 'NO TARGET'; return; }
    await h.evaluate((n) => n.scrollIntoView({ block: 'center' })); await page.waitForTimeout(100);
    const bb = await h.boundingBox();
    await page.touchscreen.tap(bb.x + Math.min(bb.width / 2, 12), bb.y + bb.height / 2); await page.waitForTimeout(200);
    R['tip ' + name] = await q(`(()=>{const t=document.querySelector('.fv-tt')||document.querySelector('.fv-tip'); if(!t) return 'NO TIP'; const b=t.getBoundingClientRect(); return (b.left>=0&&b.right<=innerWidth?'in view ':'OUT ')+t.innerText.replace(/\\n+/g,' | ').slice(0,140)})()`);
    await page.touchscreen.tap(3, 3); await page.waitForTimeout(120);
  };
  await tap('overview row', '#race-overview .fv-orow .nm', 1);
  await tap('overview race icon', '#race-overview .fv-orow > img', 0);
  await tap('best tile', '#ro-tiles .fv-tile', 0);
  await tap('spec segment', '#race-spec .sg', 1);
  await tap('spec base segment', '#race-spec .sg', 0);
  await tap('spec race icon', '#race-spec .fv-rrow > img', 2);
  await tap('class finding crest', '#race-find [data-tt]', 0);
  await tap('heat cell', '#heat .hc:not(.na)', 30);
  // heat cell "Compare races" picks the preset
  const h = (await page.$$('#heat .hc:not(.na)'))[60];
  await h.evaluate((n) => n.scrollIntoView({ block: 'center' })); const bb = await h.boundingBox();
  await page.touchscreen.tap(bb.x + 8, bb.y + bb.height / 2); await page.waitForTimeout(200);
  R['heat act buttons'] = await q(`[...document.querySelectorAll('.fv-tip button')].map(b=>b.textContent).join(' / ')`);
  await q(`[...document.querySelectorAll('.fv-tip button')].find(b=>b.textContent.startsWith('Compare')).click()`); await page.waitForTimeout(600);
  R['picked after compare'] = await q(`document.querySelector('#rs-pick').value + ' top=' + Math.round(document.querySelector('#race-spec-card').getBoundingClientRect().top)`);
  await q(`document.querySelector('#rs-next').click()`); await page.waitForTimeout(150);
  R['next preset'] = await q(`document.querySelector('#rs-pick').value`);
  await q(`document.querySelectorAll('#heat .hl')[5].click()`); await page.waitForTimeout(300);
  R['heat label picks'] = await q(`document.querySelector('#rs-pick').value`);
  // spec page link
  await page.click('[data-tab="spec"]'); await page.waitForTimeout(300);
  R['spec build'] = await q(`document.querySelector('#sp-pick').value`);
  await q(`document.querySelector('#sp-races-link').click()`); await page.waitForTimeout(500);
  R['spec link -> races'] = await q(`document.querySelector('[data-tab="races"]').getAttribute('aria-selected') + ' ' + document.querySelector('#rs-pick').value`);
  R['docW'] = await q('document.documentElement.scrollWidth');
  R['errors'] = errs.length + (await q('window.__errs.length'));
  R['errlist'] = (await q('window.__errs')).slice(0, 3);
  console.log(JSON.stringify(R, null, 1));
  await browser.close();
})().catch((e) => { console.error(e); process.exit(1); });

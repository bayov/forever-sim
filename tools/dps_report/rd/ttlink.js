// node ttlink.js <harness prefix> <mode> [scheme] [shot-prefix]: opens game tooltips of each kind with a real tap and reports the wowhead link
const { chromium } = require('playwright');
const path = require('path');
const [, , hp = 'f1', mode = 'plain', scheme = 'light', shot = ''] = process.argv;
const CASES = [
  ['ability', 'spec', '#sp-abil .nm'],
  ['aura', 'spec', '#sp-auras [data-tt]'],
  ['gear', 'spec', '#sp-gear .fv-slot'],
  ['setup', 'spec', '#sp-setup .fv-pick'],
  ['talent', 'spec', '#sp-tal .fv-tal.max'],
  ['rotation', 'spec', '#sp-rotation .fv-pr [data-tt]'],
  ['racial', 'spec', '#sp-abil [data-tt="k:s20572"]'],
  ['item', 'paladin__level_60', '#sp-gear .fv-slot[data-key$="Ranged"]'],
  ['ability2', 'paladin__level_60', '#sp-abil .nm'],
  ['setup2', 'paladin__level_60', '#sp-setup .fv-pick'],
];
(async () => {
  const browser = await chromium.launch({ executablePath: process.env.PW_CHROME || '/opt/pw-browsers/chromium-1194/chrome-linux/chrome', args: ['--allow-file-access-from-files'] });
  const ctx = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, isMobile: true, hasTouch: true, colorScheme: scheme });
  const page = await ctx.newPage();
  const errs = [];
  page.on('pageerror', (e) => errs.push(e.message));
  await page.goto('file://' + path.join(__dirname, `${hp}_${mode}.html`));
  await page.waitForTimeout(900);
  await page.click('[data-tab="spec"]');
  await page.waitForTimeout(1500);
  await page.evaluate(() => document.querySelectorAll('.fv details').forEach((d) => (d.open = true)));
  await page.waitForTimeout(400);
  const out = [];
  for (const [name, tab, sel] of CASES) {
    if (tab !== 'spec') {
      await page.evaluate((v) => { const s = document.querySelector('#sp-pick'); if (s.value !== v) { s.value = v; s.dispatchEvent(new Event('change')); } }, tab);
      await page.waitForTimeout(900);
      await page.evaluate(() => document.querySelectorAll('.fv details').forEach((d) => (d.open = true)));
      await page.waitForTimeout(300);
    }
    const n = await page.$(sel);
    if (!n) { out.push({ name, missing: sel }); continue; }
    await n.scrollIntoViewIfNeeded();
    const bb = await n.boundingBox();
    await page.touchscreen.tap(bb.x + Math.min(12, bb.width / 2), bb.y + bb.height / 2);
    await page.waitForTimeout(350);
    const r = await page.evaluate(() => {
      const t = document.querySelector('.fv-tt');
      if (!t) return null;
      const a = t.querySelector('.tt-link a');
      const tr = t.getBoundingClientRect();
      return { title: t.querySelector('.tt-title')?.textContent, lines: [...t.querySelectorAll('.tt-body > div')].slice(0, 4).map((d) => d.textContent.slice(0, 80)),
        link: a ? { href: a.href, text: a.textContent, target: a.target, rel: a.rel } : null, inView: tr.left >= 0 && tr.right <= innerWidth };
    });
    out.push({ name, ...r });
    if (shot && r && r.link) {
      const t = await page.$('.fv-tt');
      const tb = await t.boundingBox();
      await page.screenshot({ path: path.join(__dirname, 'shots', `${shot}_${name}.png`), clip: { x: 0, y: Math.max(0, tb.y - 120), width: 390, height: Math.min(844, tb.height + 240) } });
    }
    // tap the link: it must not close the page's state or be prevented; record the popup
    if (r && r.link) {
      const [pop] = await Promise.all([ctx.waitForEvent('page', { timeout: 3000 }).catch(() => null), page.tap('.fv-tt .tt-link a')]);
      out[out.length - 1].opened = pop ? pop.url() : null;
      if (pop) await pop.close();
      await page.waitForTimeout(200);
      out[out.length - 1].closedAfter = await page.evaluate(() => !document.querySelector('.fv-tt'));
    }
    await page.touchscreen.tap(380, 5);
    await page.waitForTimeout(150);
  }
  console.log(JSON.stringify({ out, errs }, null, 1));
  await browser.close();
})();

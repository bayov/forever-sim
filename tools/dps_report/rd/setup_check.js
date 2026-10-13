// node setup_check.js <harness prefix> <mode plain|str> [scheme] [shot-prefix]: the consumes and buffs card's active / inactive state
// For every preset: faded chips = setup rows with active false, faded chips come after active ones in each subsection,
// the "Hide inactive" toggle hides them (and empty subsections), and an inactive chip's tooltip leads with its reason.
// With a shot prefix, saves a phone screenshot of the card with a faded chip's tooltip open.
const { chromium } = require('playwright');
const path = require('path');
const [, , hp = 'sa', mode = 'plain', scheme = 'light', shot = ''] = process.argv;
(async () => {
  const browser = await chromium.launch({ executablePath: process.env.PW_CHROME || '/opt/pw-browsers/chromium-1194/chrome-linux/chrome', args: ['--allow-file-access-from-files'] });
  const ctx = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, colorScheme: scheme, isMobile: true, hasTouch: true });
  const page = await ctx.newPage();
  const errs = [];
  page.on('pageerror', (e) => errs.push(e.message));
  await page.goto('file://' + path.join(__dirname, `${hp}_${mode}.html`));
  await page.waitForTimeout(900);
  await page.click('[data-tab="spec"]');
  await page.waitForTimeout(1500);
  const ids = await page.evaluate(() => [...document.querySelectorAll('#sp-pick option')].map((o) => o.value));
  const report = [];
  for (const id of ids) {
    await page.evaluate((v) => { const s = document.querySelector('#sp-pick'); s.value = v; s.dispatchEvent(new Event('change')); }, id);
    await page.waitForTimeout(mode === 'plain' ? 150 : 250);
    for (let i = 0; i < 20; i++) { if (await page.evaluate(() => document.querySelectorAll('#sp-setup .fv-pick').length)) break; await page.waitForTimeout(150); }
    const r = await page.evaluate((id) => {
      const want = window.__DATA.setup.filter((s) => s.build === id);
      const offWant = want.filter((s) => s.active === false).length;
      const picks = [...document.querySelectorAll('#sp-setup .fv-pick')];
      const off = picks.filter((p) => p.classList.contains('off'));
      const badges = off.filter((p) => p.querySelector('.x')).length;
      const opac = off.length ? getComputedStyle(off[0].querySelector('.nmx')).opacity : null;
      let order = true;
      for (const box of document.querySelectorAll('#sp-setup .fv-picks')) {
        let seenOff = false;
        for (const p of box.children) { if (p.classList.contains('off')) seenOff = true; else if (seenOff) order = false; }
      }
      return { id, chips: picks.length, rows: want.length, off: off.length, offWant, badges, opac, order };
    }, id);
    report.push(r);
  }
  // toggle + tooltip on the current preset (first one with an inactive chip that is visible)
  await page.evaluate((v) => { const s = document.querySelector('#sp-pick'); s.value = v; s.dispatchEvent(new Event('change')); }, 'warrior__fury');
  await page.waitForTimeout(400);
  await page.click('#sp-setup-hide');
  await page.waitForTimeout(150);
  const hidden = await page.evaluate(() => ({
    pressed: document.querySelector('#sp-setup-hide').getAttribute('aria-pressed'),
    visibleOff: [...document.querySelectorAll('#sp-setup .fv-pick.off')].filter((p) => p.offsetParent !== null).length,
    visibleOn: [...document.querySelectorAll('#sp-setup .fv-pick:not(.off)')].filter((p) => p.offsetParent !== null).length,
    emptySubsVisible: [...document.querySelectorAll('#sp-setup .fv-ssub')].filter((s) => s.offsetParent !== null && ![...s.querySelectorAll('.fv-pick')].some((p) => p.offsetParent !== null)).length,
    saved: JSON.parse(localStorage.getItem('fv-state') || '{}').setupHide,
  }));
  await page.click('#sp-setup-hide');
  await page.waitForTimeout(150);
  const shown = await page.evaluate(() => [...document.querySelectorAll('#sp-setup .fv-pick.off')].filter((p) => p.offsetParent !== null).length);
  // tap an inactive chip: the Expose Armor one if there, else the first
  const target = await page.evaluate(() => {
    const offs = [...document.querySelectorAll('#sp-setup .fv-pick.off')];
    const n = offs.find((p) => /Curse of Recklessness/.test(p.textContent)) || offs[0];
    n.scrollIntoView({ block: 'center' });
    window.scrollBy(0, 160);
    return offs.indexOf(n);
  });
  await page.waitForTimeout(200);
  const el = (await page.$$('#sp-setup .fv-pick.off'))[target];
  const bb = await el.boundingBox();
  await page.touchscreen.tap(bb.x + bb.width / 2, bb.y + 12);
  await page.waitForTimeout(350);
  const tip = await page.evaluate(() => {
    const t = document.querySelector('.fv-tt');
    if (!t) return null;
    const first = t.querySelector('.tt-body > div');
    const r = t.getBoundingClientRect();
    return { first: first?.className, firstText: first?.textContent, inView: r.left >= 0 && r.right <= innerWidth, text: t.innerText.slice(0, 200) };
  });
  // an active chip's tooltip has no reason line
  if (shot) await page.screenshot({ path: path.join(__dirname, 'shots', shot + '.png') });
  const activeTip = await page.evaluate(() => {
    document.body.click();
    const n = document.querySelector('#sp-setup .fv-pick:not(.off)');
    n.dispatchEvent(new MouseEvent('click', { bubbles: true }));
    const t = document.querySelector('.fv-tt');
    return t ? t.querySelector('.tt-body > div')?.className : null;
  });
  const docW = await page.evaluate(() => document.documentElement.scrollWidth);
  const bad = report.filter((r) => r.chips !== r.rows || r.off !== r.offWant || r.badges !== r.off || !r.order || (r.off && +r.opac > 0.4));
  console.log(JSON.stringify({ mode, presets: report.length, bad, offPerPreset: Object.fromEntries(report.map((r) => [r.id, `${r.off}/${r.chips}`])), hidden, shownAfterToggleOff: shown, tip, activeTipFirst: activeTip, docW, errs, pageErrs: await page.evaluate(() => window.__errs) }, null, 1));
  await browser.close();
})();

// node ttdesk.js <hp> <mode>: desktop mouse: hover a tooltip target, move into the tooltip, click its wowhead link
const { chromium } = require('playwright');
const path = require('path');
const [, , hp = 'f1', mode = 'plain'] = process.argv;
(async () => {
  const browser = await chromium.launch({ executablePath: process.env.PW_CHROME || '/opt/pw-browsers/chromium-1194/chrome-linux/chrome', args: ['--allow-file-access-from-files'] });
  const ctx = await browser.newContext({ viewport: { width: 1280, height: 900 } });
  const page = await ctx.newPage();
  const errs = []; page.on('pageerror', (e) => errs.push(e.message));
  await page.goto('file://' + path.join(__dirname, `${hp}_${mode}.html`));
  await page.waitForTimeout(800);
  await page.click('[data-tab="spec"]'); await page.waitForTimeout(1200);
  const out = {};
  for (const [name, sel] of [['ability', '#sp-abil .nm'], ['setup', '#sp-setup .fv-pick']]) {
    const n = await page.$(sel); await n.scrollIntoViewIfNeeded();
    const b = await n.boundingBox();
    await page.mouse.move(b.x + 10, b.y + b.height / 2); await page.waitForTimeout(400);
    const a = await page.$('.fv-tt .tt-link a');
    if (!a) { out[name] = 'no link on hover'; continue; }
    const ab = await a.boundingBox();
    await page.mouse.move(ab.x + ab.width / 2, ab.y + ab.height / 2, { steps: 12 }); await page.waitForTimeout(200);
    const still = !!(await page.$('.fv-tt .tt-link a'));
    const [pop] = await Promise.all([ctx.waitForEvent('page', { timeout: 3000 }).catch(() => null), still ? page.mouse.click(ab.x + ab.width / 2, ab.y + ab.height / 2) : Promise.resolve()]);
    out[name] = { reachable: still, opened: pop ? pop.url() : null };
    if (pop) await pop.close();
    await page.mouse.move(5, 5); await page.waitForTimeout(500);
    out[name].closedAfterLeave = !(await page.$('.fv-tt'));
  }
  console.log(JSON.stringify({ out, errs }));
  await browser.close();
})();

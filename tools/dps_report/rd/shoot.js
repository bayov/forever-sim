// node shoot.js <prefix> <mode plain|str> <device phone|desk> <scheme light|dark> <steps-json> [slice]
// steps: [["tab","ranking"],["click","sel"],["eval","js"],["wait",ms],["scroll","sel"]]
// Saves shots/<prefix>.png (full page) and, with slice, shots/<prefix>-N.png viewport-tall slices.
const { chromium } = require('playwright');
const path = require('path');
const [, , prefix, mode = 'plain', device = 'phone', scheme = 'light', stepsJson = '[]', slice = ''] = process.argv;
const H = __dirname;
(async () => {
  const browser = await chromium.launch({ executablePath: process.env.PW_CHROME || '/opt/pw-browsers/chromium-1194/chrome-linux/chrome', args: ['--allow-file-access-from-files'] });
  const vp = device === 'phone' ? { width: 390, height: 844 } : { width: 1280, height: 900 };
  const ctx = await browser.newContext({ viewport: vp, deviceScaleFactor: device === 'phone' ? 2 : 1, colorScheme: scheme, isMobile: device === 'phone', hasTouch: device === 'phone' });
  const page = await ctx.newPage();
  const logs = [];
  page.on('console', (m) => { if (m.type() === 'error' || m.type() === 'warning') logs.push(m.type() + ': ' + m.text()); });
  page.on('pageerror', (e) => logs.push('pageerror: ' + e.message));
  await page.goto('file://' + path.join(H, `${process.env.HP || 'h'}_${mode}.html`));
  await page.waitForTimeout(600);
  for (const [k, a, b] of JSON.parse(stepsJson)) {
    if (k === 'tab') await page.click(`[data-tab="${a}"]`);
    else if (k === 'click') await page.click(a, { force: true });
    else if (k === 'tap') { const el = await page.$(a); const bb = await el.boundingBox(); await page.mouse.click(bb.x + Math.min(10, bb.width / 2), bb.y + bb.height / 2); }
    else if (k === 'hover') await page.hover(a);
    else if (k === 'eval') await page.evaluate(a);
    else if (k === 'scroll') await page.$eval(a, (n) => n.scrollIntoView({ block: 'center' }));
    else if (k === 'wait') await page.waitForTimeout(+a);
    else if (k === 'select') await page.selectOption(a, b);
    if (k !== 'tap' && k !== 'hover') await page.mouse.move(1, 1);
    await page.waitForTimeout(250);
  }
  await page.waitForTimeout(400);
  const errs = await page.evaluate(() => window.__errs);
  const dims = await page.evaluate(() => ({ docW: document.documentElement.scrollWidth, vw: innerWidth, h: document.documentElement.scrollHeight }));
  const out = path.join(H, 'shots', prefix + '.png');
  if (slice === 'view') await page.screenshot({ path: out });
  else await page.screenshot({ path: out, fullPage: true });
  if (slice === 'slice') {
    const n = Math.ceil(dims.h / vp.height);
    for (let i = 0; i < n && i < 14; i++) await page.screenshot({ path: path.join(H, 'shots', `${prefix}-${i}.png`), fullPage: true, clip: { x: 0, y: i * vp.height, width: vp.width, height: Math.min(vp.height, dims.h - i * vp.height) } });
  }
  console.log(JSON.stringify({ prefix, dims, errs, logs: logs.slice(0, 10) }));
  await browser.close();
})().catch((e) => { console.error(e); process.exit(1); });

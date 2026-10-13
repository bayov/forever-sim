// node -r ./slow.js <check>.js ...: scales every page.waitForTimeout by SLOW (default 4) so checks on the ASYNC=1 harness
// (datasets arrive 0.06-1.4 s late) wait for the data the way a reader's page does.
const f = +(process.env.SLOW || 4);
const pw = require('playwright');
const launch = pw.chromium.launch.bind(pw.chromium);
pw.chromium.launch = async (...a) => {
  const b = await launch(...a);
  const nc = b.newContext.bind(b);
  b.newContext = async (...c) => {
    const ctx = await nc(...c);
    const np = ctx.newPage.bind(ctx);
    ctx.newPage = async () => { const p = await np(); const w = p.waitForTimeout.bind(p); p.waitForTimeout = (ms) => w(ms * f); return p; };
    return ctx;
  };
  return b;
};

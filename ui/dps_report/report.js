// The DPS Report's page logic: reads the datasets through `dash` (see data.ts) and draws every tab with d3.
// It began as the script of a Claude dashboard artifact, hence the dash API and data-source marks.
/* global d3 */
import tippy from 'tippy.js';

export function startReport(dash, events) {
  const CLASSES = ['Warrior', 'Rogue', 'Hunter', 'Paladin', 'Shaman', 'Druid', 'Mage', 'Warlock', 'Priest'];
  const HORDE = new Set(['Orc', 'Undead', 'Tauren', 'Troll', 'Windshaper']);
  const ccol = (c) => `var(--fv-${String(c).toLowerCase()})`;
  const fcol = (r) => (HORDE.has(r) ? 'var(--fv-horde)' : 'var(--fv-alliance)');
  const fmt = d3.formatLocale({ decimal: '.', thousands: ',', grouping: [3], currency: ['', ''], minus: '-' }).format;
  const f1 = fmt('.1f'), f0 = fmt('.0f'), f2 = fmt('.2f');
  const sgn = (v, f = f1) => (v > 0 ? '+' : v < 0 ? '-' : '') + f(Math.abs(v));
  const $ = (s) => document.querySelector(s);
  const el = (tag, attrs = {}, ...kids) => {
    const n = document.createElement(tag);
    for (const [k, v] of Object.entries(attrs)) {
      if (v == null) continue;
      if (k === 'class') n.className = v;
      else if (k === 'style') n.style.cssText = v;
      else if (k === 'text') n.textContent = v;
      else if (k.startsWith('on')) n.addEventListener(k.slice(2), v);
      else n.setAttribute(k, v);
    }
    for (const c of kids) if (c != null) n.append(c);
    return n;
  };
  const enc = encodeURIComponent;
  let asmOn = false;
  const ASM_FIELDS = new Set(['dps', 'se', 'p1', 'p10', 'p25', 'p50', 'p75', 'p90', 'min', 'max', 'stdev', 'dps120', 'dps300', 'rank', 'rankAll', 'bestRaceDps', 'hordeDps', 'allianceDps', 'cv']);
  const mark = (n, src, field, where) => {
    if (asmOn && ((src === 'builds' && ASM_FIELDS.has(field)) || src === 'pairs' || (src === 'lengths' && field === 'dps'))) {
      for (const a of ['data-source', 'data-field', 'data-where']) n.removeAttribute(a);
      return n;
    }
    n.setAttribute('data-source', src);
    if (field) n.setAttribute('data-field', field);
    if (where) n.setAttribute('data-where', where);
    return n;
  };
  // A file's cells can arrive as text: numbers, booleans and nested lists are typed back here.
  const normCache = new WeakMap();
  const normVal = (v) => {
    if (typeof v !== 'string') return v;
    if (v === 'true') return true;
    if (v === 'false') return false;
    if (/^-?\d+(\.\d+)?(e-?\d+)?$/.test(v)) return +v;
    if (/^[\[{]/.test(v)) { try { return JSON.parse(v); } catch { return v; } }
    return v;
  };
  const rowsOf = (id) => {
    const r = dash.data(id);
    if (r.status !== 'ok') return null;
    const d = r.data;
    if (!Array.isArray(d)) return d;
    if (!normCache.has(d)) normCache.set(d, d.map((row) => (row && typeof row === 'object' ? Object.fromEntries(Object.entries(row).map(([k, v]) => [k, normVal(v)])) : row)));
    return normCache.get(d);
  };
  const statusOf = (id) => dash.data(id).status;
  // encounter presets (the boss as a creature type) are listed only while "Advantaged" is on; Races always uses the generic boss
  const isEnc = (b) => !!(b && b.encounter && b.encounter !== 'null');
  // duty presets (the preset keeps a raid buff or debuff up itself, standards 23) are listed in the Ranking only, while "Duty"
  // is on; everywhere else their base preset stands for them
  const isDuty = (b) => !!(b && b.duty && b.duty !== 'null');
  const visB = () => { const r = rowsOf('builds'); return r && r.filter((b) => !isDuty(b) && (st.enc || !isEnc(b))).map(adjB); };
  const rankB = () => { const r = rowsOf('builds'); return r && r.filter((b) => (st.duty || !isDuty(b)) && (st.enc || !isEnc(b))).map(adjB); };
  const genB = () => { const r = rowsOf('builds'); return r && r.filter((b) => !isEnc(b) && !isDuty(b)); };
  const dutyBase = (id) => { const b = (rowsOf('builds') || []).find((x) => x.id === id); return isDuty(b) ? b.baseBuild : id; };
  const rankOf = (b, list) => 1 + list.filter((x) => +x.dps > +b.dps).length;
  // ---------- assumptions: each preset's DPS under the picked settings ----------
  const asmSel = () => Object.entries(st.asm || {}).filter(([, v]) => v).sort((p, q) => (p[0] < q[0] ? -1 : 1));
  const asmReady = () => statusOf('assumptions') === 'ok' && statusOf('assumption_dps') === 'ok';
  const asmIdxCache = new WeakMap();
  const asmIdx = () => {
    const d = rowsOf('assumption_dps');
    if (!asmIdxCache.has(d)) { const m = new Map(); for (const r of d) { if (!m.has(r.build)) m.set(r.build, new Map()); m.get(r.build).set(r.combo || '', r); } asmIdxCache.set(d, m); }
    return asmIdxCache.get(d);
  };
  const asmKeysOf = (bid) => new Set((rowsOf('assumptions') || []).filter((a) => arr(a.builds).includes(bid)).map((a) => a.key));
  // the exact combination where it was simmed, else the single settings multiplied
  const asmFactor = (bid, L) => {
    if (!asmOn) return 1;
    bid = dutyBase(bid);
    const ks = asmKeysOf(bid), sel = asmSel().filter(([k]) => ks.has(k));
    const idx = asmIdx().get(bid);
    if (!sel.length || !idx) return 1;
    const f = +L === 120 ? 'dps120' : +L === 300 ? 'dps300' : 'dps180';
    const base = idx.get('');
    if (!base || !+base[f]) return 1;
    const exact = idx.get(sel.map(([k, v]) => `${k}=${v}`).join('|'));
    if (exact && +exact[f]) return +exact[f] / +base[f];
    let x = 1;
    for (const [k, v] of sel) { const one = idx.get(`${k}=${v}`); if (one && +one[f]) x *= +one[f] / +base[f]; }
    return x;
  };
  const adjB = (b) => {
    const x = asmFactor(b.id, 180);
    if (x === 1) return b;
    const o = { ...b, asmF: x };
    for (const k of ['dps', 'se', 'p1', 'p10', 'p25', 'p50', 'p75', 'p90', 'min', 'max', 'stdev', 'bestRaceDps', 'hordeDps', 'allianceDps']) if (o[k] != null) o[k] = +o[k] * x;
    o.dps120 = +b.dps120 * asmFactor(b.id, 120); o.dps300 = +b.dps300 * asmFactor(b.id, 300);
    return o;
  };

  // ---------- calculations ----------
  dash.calc('overview', {
    inputs: ['builds'], title: 'Headline figures',
    description: 'Top, median and lowest preset DPS, how many presets and classes, and the top-to-bottom gap in percent.',
    fn: (builds) => {
      const s = builds.filter((b) => !(b.encounter && b.encounter !== 'null') && !(b.duty && b.duty !== 'null')).sort((a, b) => b.dps - a.dps);
      const n = (v) => +v;
      const top = s[0], bot = s[s.length - 1];
      return [{
        topLabel: top.label, topDps: n(top.dps), bottomLabel: bot.label, bottomDps: n(bot.dps),
        median: Math.round(d3.median(s, (b) => b.dps) * 10) / 10, count: s.length,
        classes: new Set(s.map((b) => b.class)).size, spreadPct: Math.round((top.dps / bot.dps - 1) * 1000) / 10,
      }];
    },
  });
  dash.calc('racespread', {
    inputs: ['race_variants'], title: 'Race spread',
    description: 'Per preset and race scenario, the gap between its best and worst race in percent, and which races those are.',
    fn: (race_variants) => d3.groups(race_variants, (r) => r.build, (r) => r.variant).flatMap(([build, vs]) => vs.map(([variant, rs]) => {
      const best = rs.reduce((x, y) => (+y.dps > +x.dps ? y : x)), worst = rs.reduce((x, y) => (+y.dps < +x.dps ? y : x));
      return { build, variant, spreadPct: Math.round((1 - worst.dps / best.dps) * 1000) / 10, best: best.race, worst: worst.race };
    })),
  });
  dash.calc('pairs', {
    inputs: ['builds'], title: 'Preset against preset',
    description: 'For every pair of presets, the second’s mean, 10th, 50th and 90th percentile, standard deviation and 120s / 300s DPS relative to the first, in percent.',
    fn: (builds) => {
      const out = [];
      for (const a of builds) for (const b of builds) {
        const r = (k) => (a[k] ? Math.round((b[k] / a[k] - 1) * 1000) / 10 : null);
        out.push({ a: a.id, b: b.id, dps: r('dps'), p10: r('p10'), p50: r('p50'), p90: r('p90'), stdev: r('stdev'), dps120: r('dps120'), dps300: r('dps300') });
      }
      return out;
    },
  });

  dash.calc('roles', {
    inputs: ['stats'], title: 'Role per preset',
    description: 'Ranged when ranged attack power beats attack power, caster when spell power does, melee otherwise; read from each preset’s character sheet.',
    fn: (stats) => {
      const by = {};
      for (const s of stats) (by[s.build] = by[s.build] || {})[s.stat] = Number(s.value) || 0;
      return Object.entries(by).map(([build, s]) => ({ build, role: (s.RangedAttackPower || 0) > (s.AttackPower || 0) ? 'Ranged' : (s.SpellPower || 0) > (s.AttackPower || 0) ? 'Caster' : 'Melee' }));
    },
  });

  // ---------- state ----------
  const store = (() => { try { return window.localStorage; } catch { return null; } })();
  const saved = (() => { try { return JSON.parse(store?.getItem('fv-state') || '{}'); } catch { return {}; } })();
  const st = Object.assign({ tab: 'ranking', cls: 'All', role: 'All', fac: 'All', topN: 'all', defaults: false, best: false, sort: 'dps', len: '180', build: null, sample: 0, A: null, B: null, Ar: null, Av: null, Br: null, Bv: null, alwaysOn: true, hlLen: null, hlMana: null, distMode: 'bins', abilMetric: 'share', spreadSort: 'spread', auraOwner: 'All', rsOpt: {}, rsFac: 'All', srFac: 'All', slFac: 'All', rsRole: 'All', rsCls: 'All', rsSpec: 'All', rlFac: 'All', rlRole: 'All', rlCls: 'All', rlSpec: 'All', setupHide: false, enc: false, duty: true, asm: {}, asmOpen: [] }, saved);
  st.len = '180'; // a fight length other than 180s is never restored: the page always opens on the main numbers
  const persist = () => { try { store?.setItem('fv-state', JSON.stringify(st)); } catch {} };

  // ---------- tooltip pieces ----------
  // Every tooltip is built from these, so they all look alike: a title, then the answer (a key figure or one short
  // line), then a grid of labelled values or a small table, signed deltas in the ok and bad colours, and a muted
  // footnote. Longer explanations live in the Method tab.
  const tipTitle = (text, sub, icon) => el('div', { class: 'fv-tip-title' }, icon || null, el('b', { text }), sub ? el('span', { text: sub }) : null);
  // the key figure: a big value, its unit, and what it is or how it compares
  const tipKey = (value, unit, after) => el('div', { class: 'fv-tip-key' }, el('b', {}, value), unit ? el('span', { class: 'u', text: unit }) : null,
    after == null || after === '' ? null : typeof after === 'string' ? el('span', { class: 'fv-tip-m', text: after }) : after);
  const tipDelta = (v, f = f1, unit = '%', eps = 0.05) => el('span', { class: 'fv-tip-d' + (Math.abs(v) < eps ? '' : v > 0 ? ' pos' : ' neg'), text: sgn(v, f) + unit });
  // rows of [label, value, 'txt' for a value that is words rather than a number]; null rows are left out
  const tipGrid = (rows) => {
    const rs = rows.filter(Boolean);
    return rs.length ? el('div', { class: 'fv-tip-grid' }, ...rs.flatMap(([k, v, kind]) => [el('span', { class: 'k' }, k), el('span', { class: 'v' + (kind === 'txt' ? ' txt' : '') }, v ?? '')])) : null;
  };
  // a small table: head is the column titles, rows the cells; the first column is the row's name
  const tipTable = (head, rows) => el('table', { class: 'fv-tip-table' },
    head ? el('thead', {}, el('tr', {}, ...head.map((h) => el('th', {}, h)))) : null,
    el('tbody', {}, ...rows.filter(Boolean).map((r) => el('tr', {}, ...r.map((c) => el('td', {}, c ?? ''))))));
  const tipText = (...kids) => el('div', { class: 'fv-tip-text' }, ...kids);
  const tipNote = (...kids) => el('div', { class: 'fv-tip-note' }, ...kids);
  // Game text inside one of our tooltips (a racial's or a talent's description) takes the game's tooltip gold.
  const GAME_TEXT = 'fv-tip-game';
  const racialText = (part) => !['base', 'swap', 'target'].includes(part);
  const tipSw = (color) => el('i', { class: 'sw', style: `background:${color}` });
  const tipChips = (items, more) => el('div', { class: 'fv-tip-chips' }, ...items.map((t) => el('span', { text: t })), more ? el('em', { text: more }) : null);

  // A chart whose bars start well above 0 makes small gaps look big, so we say so plainly above it. start is where the
  // bars start (null hides the warning), and what is what the gaps are between.
  const axisWarn = (node, start, what) => {
    node.hidden = !start;
    if (start) node.replaceChildren(el('i', { class: 'fas fa-triangle-exclamation', 'aria-hidden': 'true' }),
      el('span', {}, 'Bars start at ', el('b', { text: start }), `, not at 0, so the gaps between ${what} look much bigger than they are.`));
  };

  // ---------- tooltip ----------
  // Chart tooltips: tap (or hover with a mouse) a mark for its exact values; the mark is highlighted; tap elsewhere to close.
  let tipEl = null, tipTippy = null, tipHl = [], tipPinned = false, tipCleanup = null, tipWait = null;
  const hideTip = () => {
    clearTimeout(tipWait);
    if (tipTippy) { tipTippy.destroy(); tipTippy = null; }
    tipEl = null;
    for (const n of tipHl) n.classList.remove('is-hl');
    if (tipCleanup) { tipCleanup(); tipCleanup = null; }
    tipHl = []; tipPinned = false;
  };
  // A tooltip with buttons (like "Open in Specs ›") stays up while the mouse is on it.
  //
  // When the mouse leaves its mark, or moves on to another mark, the hover handlers call tipHover with what to do
  // next (hide the tooltip, or show the other mark's). With buttons in the tooltip we wait a moment first, so the mouse
  // can cross over to the tooltip, and reaching the tooltip cancels it. Any other tooltip changes right away.
  const tipHover = (fn) => {
    clearTimeout(tipWait);
    if (tipEl && tipEl.classList.contains('act') && !tipPinned) tipWait = setTimeout(fn, 300);
    else fn();
  };
  // x and y are where the mark is in host, which the tooltip points at. head is a title (text, or tipTitle(...)), and
  // lines are the tooltip pieces above (plain text lines still work). With opts.end the tooltip ends at x instead of
  // starting there.
  const showTip = (host, x, y, head, lines, opts = {}) => {
    hideTip();
    try { hideTT(); } catch {}
    tipEl = el('div', { class: 'fv-tip' + (opts.open ? ' act' : ''), role: 'status' }, typeof head === 'string' ? tipTitle(head) : head,
      ...lines.filter((l) => l != null && l !== '').map((l) => (typeof l === 'string' ? el('span', { text: l }) : l)));
    const acts = [];
    if (opts.act) {
      const ba = el('button', { class: 'fv-tip-open', text: opts.act[0] });
      ba.addEventListener('click', (e) => { e.stopPropagation(); hideTip(); opts.act[1](); });
      acts.push(ba);
    }
    if (opts.open) {
      const bt = el('button', { class: 'fv-tip-open', text: 'Open in Specs ›' });
      bt.addEventListener('click', (e) => { e.stopPropagation(); hideTip(); openSpec(opts.open); });
      acts.push(bt);
    }
    if (acts.length) { tipEl.classList.add('act'); tipEl.append(el('div', { class: 'fv-tip-row' }, ...acts)); }
    const at = () => { const r = host.getBoundingClientRect(); const px = r.left + x, py = r.top + y; return { width: 0, height: 0, x: px, y: py, left: px, right: px, top: py, bottom: py }; };
    tipTippy = tippy(host, {
      content: tipEl, trigger: 'manual', interactive: acts.length > 0, appendTo: document.body, hideOnClick: false,
      placement: (opts.below ? 'bottom' : 'top') + (opts.end ? '-end' : '-start'), offset: [opts.end ? 0 : 10, 8], getReferenceClientRect: at, duration: 0,
    });
    tipTippy.show();
    if (acts.length) {
      tipTippy.popper.addEventListener('pointerenter', () => clearTimeout(tipWait));
      tipTippy.popper.addEventListener('pointerleave', (e) => {
        if (e.pointerType === 'mouse' && !tipPinned && !tipHl.some((n) => n.contains(e.relatedTarget))) tipHover(opts.onLeave || hideTip);
      });
    }
    tipHl = (Array.isArray(opts.hl) ? opts.hl : opts.hl ? [opts.hl] : []).filter(Boolean);
    for (const n of tipHl) n.classList.add('is-hl');
    tipCleanup = opts.onHide || null;
    tipPinned = opts.pinned !== false;
  };
  // bind a tooltip to d3 marks: fn(event, datum, node) shows it
  const tipOn = (sel, fn) => sel.attr('data-tip', '1').style('cursor', 'pointer')
    .on('click', function (e, d) { e.stopPropagation(); if (tipHl.includes(this) && tipPinned) { hideTip(); return; } fn(e, d, this); tipPinned = true; })
    .on('pointerenter', function (e, d) {
      if (e.pointerType !== 'mouse' || tipPinned) return;
      if (tipHl.includes(this)) { clearTimeout(tipWait); return; }
      tipHover(() => { fn(e, d, this); tipPinned = false; });
    })
    .on('pointerleave', function (e) { if (e.pointerType === 'mouse' && !tipPinned && tipHl.includes(this)) tipHover(hideTip); });
  document.addEventListener('click', (e) => { if (!e.target.closest('[data-tip]') && !e.target.closest('.fv-tip')) hideTip(); });

  // An advantaged preset's "vs Undead" badge says on hover how much it gains over its normal build, and why.
  //
  // The abilities and talents it names (advantageRefs) get their icon in front of their name in the line, and
  // their own game tooltip under it.
  const advTip = (b) => {
    const base = (rowsOf('builds') || []).find((x) => x.id === b.baseBuild);
    const ab = base ? adjB(base) : null;
    const specs = arr(b.advantageRefs).map((v) => { try { return specFor(v, b.id); } catch { return null; } }).filter(Boolean);
    return el('div', { class: 'fv-tip fv-advtip' }, tipTitle(`Advantaged vs ${b.encounter}`),
      ab && +ab.dps ? tipKey(tipDelta(100 * (b.dps / ab.dps - 1)), null, `over ${ab.build} (generic boss)`) : null,
      tipText(...withIcons(b.advantageShort || b.advantage, specs)),
      specs.length ? el('div', { class: 'fv-advtip-cards' }, ...specs.map((sp) => ttCard(sp, true))) : null);
  };
  // text as nodes, with each spec's first mention as its icon and name
  const withIcons = (text, specs) => {
    let parts = [String(text)];
    for (const sp of specs) {
      const name = String(sp.title);
      const k = parts.findIndex((p) => typeof p === 'string' && p.includes(name));
      if (k < 0) continue;
      const [before, ...rest] = parts[k].split(name);
      parts.splice(k, 1, before, el('span', { class: 'fv-advtip-ref' }, sp.icon ? iconImg(sp.icon, 'fv-ico') : null, el('b', { text: name })), rest.join(name));
    }
    return parts.filter((p) => p !== '');
  };
  // A duty preset's badge on its ranking row, and what the duty costs against its base preset (the raid covers it there).
  // What the duty costs, signed: a duty can come out level with its base, or a hair ahead within the noise.
  const dutyCostText = (b, base) => { const c = base.dps - b.dps; return Math.abs(c) < 0.05 ? '±0' : c > 0 ? `−${f1(c)}` : `+${f1(-c)}`; };
  const dutyCostEl = (b, list) => {
    const base = list.find((x) => x.id === b.baseBuild) || adjB((rowsOf('builds') || []).find((x) => x.id === b.baseBuild) || {});
    return +base.dps ? el('span', {}, `${b.duty} · ${dutyCostText(b, base)} DPS `, tipDelta(100 * (b.dps / base.dps - 1))) : b.duty;
  };
  const dutyTip = (b, list) => {
    const base = list.find((x) => x.id === b.baseBuild) || adjB((rowsOf('builds') || []).find((x) => x.id === b.baseBuild) || {});
    return el('div', { class: 'fv-tip fv-advtip' }, tipTitle(`On ${b.duty} duty`),
      +base.dps ? tipKey(dutyCostText(b, base), 'DPS', el('span', { class: 'fv-tip-m' }, tipDelta(100 * (b.dps / base.dps - 1)), ` vs ${base.build}`)) : null,
      b.dutyDefensive === true ? tipText('Defensive: no DPS gain for the raid.') : null,
      tipText(b.summary));
  };
  // The badge's short name for each duty; its tooltip gives the full name, and says when the duty is defensive.
  const DUTY_SHORT = { 'Battle Shout': 'Shout', 'Curse of the Elements': 'Curse', 'Faerie Fire': 'FF', "Hunter's Mark": 'Mark', 'Judgement of the Crusader': 'Crusader', 'Scorpid Sting': 'Scorpid' };
  const dutyBadge = (b, list) => {
    const badge = el('span', { class: 'fv-duty' + (b.dutyDefensive === true ? ' def' : ''), text: `${DUTY_SHORT[b.duty] || b.duty} duty` });
    tippy(badge, { content: dutyTip(b, list), duration: 0, onShow: (t) => t.setContent(dutyTip(b, list)) });
    badge.addEventListener('click', (e) => { e.stopPropagation(); hideTip(); setTimeout(() => badge._tippy?.show()); });
    return badge;
  };
  const encBadge = (b) => {
    const badge = el('span', { class: 'fv-enc', text: 'vs ' + b.encounter });
    if (b.advantage || b.advantageShort) {
      // We build the tooltip each time it shows, because the spell and talent tooltips load after the page.
      tippy(badge, { content: advTip(b), duration: 0, onShow: (t) => t.setContent(advTip(b)) });
      // a tap on the badge shows why, not the ranking row's own tooltip
      badge.addEventListener('click', (e) => { e.stopPropagation(); hideTip(); setTimeout(() => badge._tippy?.show()); });
    }
    return badge;
  };
  // Plain title tooltips become the sim UI's tooltips. We move each title into a tippy as it shows up, in the
  // layout or in anything we draw later.
  const tipTitles = (n) => {
    const t = n.getAttribute('title');
    n.removeAttribute('title');
    if (!t) return;
    const content = el('div', { class: 'fv-tip fv-tip-text', text: t });
    if (n._tippy) n._tippy.setContent(content);
    else tippy(n, { content, duration: 0 });
  };
  {
    const root = $('#fv');
    root.querySelectorAll('[title]').forEach(tipTitles);
    new MutationObserver((muts) => {
      for (const m of muts) {
        if (m.type === 'attributes') { if (m.target.hasAttribute('title')) tipTitles(m.target); continue; }
        for (const n of m.addedNodes) {
          if (n.nodeType !== 1) continue;
          if (n.hasAttribute('title')) tipTitles(n);
          n.querySelectorAll('[title]').forEach(tipTitles);
        }
      }
    }).observe(root, { subtree: true, childList: true, attributes: true, attributeFilter: ['title'] });
  }

  // ---------- game tooltips ----------
  // Anything with data-tt="<type>:<id>" gets a game-style tooltip: tap (data-tt-tap) or long-press on touch, hover with a mouse.
  // Types: k spell or item key, a aura key, g gear slot, su setup chip, tal talent, r race, c class, w stat weight.
  // data-tt-b on the element or an ancestor names the preset; otherwise it is the one on the Spec page.
  const QCOL = { poor: '#9d9d9d', common: '#ffffff', uncommon: '#1eff00', rare: '#0070dd', epic: '#a335ee', legendary: '#ff8000' };
  const CCOL = { Warrior: '#c79c6e', Rogue: '#fff569', Hunter: '#abd473', Paladin: '#f58cba', Shaman: '#2f8cff', Druid: '#ff7d0a', Mage: '#69ccf0', Warlock: '#9482c9', Priest: '#f2f2f2' };
  const GOLD = '#ffd100';
  const tt = (n, v, tap = true, b = null) => {
    if (!n || !v) return n;
    n.setAttribute('data-tt', v);
    if (tap) n.setAttribute('data-tt-tap', '');
    if (b) n.setAttribute('data-tt-b', b);
    return n;
  };
  const byKeyCache = new WeakMap();
  const byKey = (id, key = 'key') => {
    const r = rowsOf(id);
    if (!Array.isArray(r)) return null;
    if (!byKeyCache.has(r)) byKeyCache.set(r, new Map(r.map((x) => [String(x[key]), x])));
    return byKeyCache.get(r);
  };
  const ttRow = (key) => { const m = byKey('tooltips'); return m ? m.get(key) || m.get(key.replace(/t-?\d+$/, '')) || null : null; };
  const arr = (v) => (Array.isArray(v) ? v : []);
  // wowhead page for a real spell or item: Forever's database where it knows the id, else Classic's (labelled), as the sim UI links them
  const whLink = (path, dom) => {
    if (!/^(item|spell)=\d+$/.test(String(path || ''))) return null;
    const d = dom === 'classic' ? 'classic' : 'forever';
    return { href: `https://www.wowhead.com/${d}/${path}`, label: d === 'forever' ? 'Wowhead (Forever)' : 'Wowhead (Classic)', note: d === 'classic' ? 'Not in Forever’s database' : '' };
  };
  const rowLink = (r) => (r ? whLink(r.wh, r.whd) : null);
  const talFill = (t, r) => {
    const vals = arr(t.ranks)[Math.max(0, r - 1)] || [];
    return String(t.description || '').replace(/\{(\d+)\}/g, (m0, i) => (vals[+i] ?? m0)).replace(/\s*\[\[\d+\]\]/g, '');
  };
  const nums = (rows) => rows.filter(Boolean).map(([a, b]) => ['n', a, b]);

  const spellSpec = (key, b, aura) => {
    const bl = (rowsOf('builds') || []).find((x) => x.id === b);
    const r = (bl && byKey('tooltips')?.get(`${key}:${bl.class}`)) || ttRow(key), n = byKey('names')?.get(key);
    const spec = {
      title: String(r?.title ?? n?.name ?? key), icon: r?.icon || n?.icon,
      color: r?.kind === 'item' ? QCOL[r.quality] || QCOL.common : GOLD,
      sub: r?.sub != null ? String(r.sub) : n?.rank ? `Rank ${n.rank}` : '', lines: arr(r?.lines), nums: [], link: rowLink(r),
    };
    const ab = (rowsOf('abilities') || []).find((a) => a.build === b && a.key === key);
    if (ab) {
      spec.nums.push(...nums([
        ab.dps > 0 ? ['DPS', `${f1(ab.dps)} (${f1(ab.share)}% of total)`] : null,
        ab.casts > 0 ? ['Casts per fight', f1(ab.casts)] : null,
        ab.dps > 0 && ab.hits + ab.crits > 0 ? ['Crit', `${f1(ab.critPct)}%`] : null,
        ab.avgHit > 0 ? ['Average hit', f0(ab.avgHit)] : null,
        ab.avgCrit > 0 ? ['Average crit', f0(ab.avgCrit)] : null,
        ab.ticks > 0 && ab.avgTick > 0 ? ['Average tick', `${f0(ab.avgTick)} × ${f1(ab.ticks)}`] : null,
      ]));
    }
    const au = (rowsOf('auras') || []).filter((a) => a.build === b && a.key === key);
    for (const a of au) {
      spec.nums.push(...nums([[a.owner === 'Self' ? 'Uptime' : 'Uptime on the boss', `${f1(a.uptime)}%`], a.procs > 0 ? ['Applied per fight', f1(a.procs)] : null]));
    }
    if (spec.nums.length && bl) spec.numsHead = bl.label;
    if (!spec.lines.length && !spec.nums.length && aura) spec.lines = [['m', 'No tooltip text.']];
    return spec;
  };

  const gearSpec = (slot, b) => {
    const g = (rowsOf('gear') || []).find((x) => x.build === b && x.slot === slot);
    if (!g) return null;
    const r = ttRow('i' + g.itemId);
    const lines = [...arr(r?.lines)];
    const ench = ttRow(`ench:${b}:${slot}`);
    if (ench || g.enchant) lines.push(['g', `Enchanted: ${ench?.title || g.enchant}`], ...arr(ench?.lines).map((l) => ['g', l[1]]));
    if (r?.set) {
      const have = (rowsOf('gear') || []).filter((x) => x.build === b && ttRow('i' + x.itemId)?.set === r.set).length;
      lines.push(['s'], ['y', `${r.set} (${have}/${r.setSize ?? have})`],
        ...arr(r.setBonuses).map(([n, t]) => [have >= n ? 'g' : 'm', `(${n}) Set: ${t}`]));
    }
    return { title: String(r?.title ?? g.item), icon: r?.icon || g.icon, color: QCOL[r?.quality || g.quality] || QCOL.common, sub: '', lines, link: rowLink(r) };
  };

  const setupSpec = (text, b) => {
    const r = ttRow('setup:' + text);
    const s = (rowsOf('setup') || []).find((x) => x.item === text && (!b || x.build === b)) || (rowsOf('setup') || []).find((x) => x.item === text);
    const why = s && s.active === false && s.why ? 'No effect: ' + s.why : s?.note ? String(s.note) : '';
    return { title: String(r?.title ?? text), icon: s?.icon || r?.icon, color: r?.kind === 'item' ? QCOL[r.quality] || QCOL.common : GOLD, sub: String(r?.sub ?? s?.category ?? ''), lines: arr(r?.lines), why, link: rowLink(r) };
  };

  const talentSpec = (cls, tree, idx, b) => {
    const t = (rowsOf('talenttrees') || []).find((x) => x.class === cls && +x.tree === +tree && +x.idx === +idx);
    const bl = (rowsOf('builds') || []).find((x) => x.id === b);
    if (!t) return null;
    const p = bl ? +((String(bl.talents).split('-')[t.tree] || '')[t.idx] || 0) : 0;
    const lines = [['m', `Rank ${p}/${t.maxPoints}`]];
    if (p > 0) lines.push(['y', talFill(t, p)]);
    if (p < t.maxPoints) lines.push(...(p > 0 ? [['s']] : []), ['t', 'Next rank:'], ['y', talFill(t, p + 1)]);
    const wh = String(arr(t.wh)[Math.max(0, p - 1)] || '').split(':');
    return { title: String(t.name), icon: t.icon, color: GOLD, sub: String(t.treeName || ''), lines, link: wh[1] ? whLink('spell=' + wh[1], wh[0]) : null };
  };

  const raceSpec = (race, cls, b) => {
    const r = (cls && ttRow(`race:${race}:${cls}`)) || ttRow(`race:${race}`);
    const rr = (rowsOf('races') || []).find((x) => x.race === race && (!b || x.build === b));
    const spec = { title: race, icon: rr?.raceIcon, color: HORDE.has(race) ? '#ff5a4a' : '#5aa2ff', sub: String(r?.sub ?? (HORDE.has(race) ? 'Horde' : 'Alliance')), lines: arr(r?.lines), nums: [] };
    if (b && rr && rr.build === b) {
      spec.numsHead = (rowsOf('builds') || []).find((x) => x.id === b)?.label;
      spec.nums = nums([['DPS', `${f1(rr.dps)} ± ${f2(rr.se)}`], [rr.vsBest >= -0.005 ? 'Best race' : 'Behind the best race', rr.vsBest >= -0.005 ? '★' : `${f2(-rr.vsBest)}%`], rr.isPreset ? ['Preset race', '✓'] : null]);
    }
    return spec;
  };

  const classSpec = (cls) => {
    const bs = (visB() || []).filter((x) => x.class === cls).sort((x, y) => y.dps - x.dps);
    return {
      title: cls, icon: 'classicon_' + cls.toLowerCase(), color: CCOL[cls] || '#fff', sub: `${bs.length} preset${bs.length === 1 ? '' : 's'} · DPS`,
      lines: bs.map((x) => ['2', x.build, f1(x.dps)]),
    };
  };

  const weightSpec = (stat, b) => {
    const w = (rowsOf('weights') || []).find((x) => x.build === b && x.stat === stat);
    if (!w) return null;
    const s = ttRow('stat:' + stat);
    const pct = w.unit === '1%';
    // The answer goes in the header (what a point is worth), the numbers behind it in the grid.
    const worth = pct ? w.ratingWorth : w.worth;
    const hit = stat === 'MeleeHit' ? 'hit' : 'crit';
    return { title: pct ? String(w.label).replace(' (1%)', ' rating') : String(w.label), color: GOLD, sub: worth == null ? '' : `${f2(worth)} ${w.primaryAbbr} a point`,
      // the stat's description is gold, as on the character sheet; our own notes stay grey
      lines: arr(s?.lines).map((l) => (l[0] === 'm' && !/^Losing hit:/.test(String(l[1])) ? ['y', l[1]] : l)),
      nums: nums(pct ? [[`+1% ${hit}`, `${f2(w.dpsPer)} DPS`], ['Rating for 1%', String(stat === 'MeleeHit' ? 10 : 14)], ['Error', `± ${fmt('.3f')(w.sePer)} DPS`]]
        : [['+1 point', `${fmt('.3f')(w.dpsPer)} DPS`], ['Error', `± ${fmt('.3f')(w.sePer)} DPS`]]) };
  };

  const ttSpec = (node) => specFor(node.getAttribute('data-tt') || '', node.closest('[data-tt-b]')?.getAttribute('data-tt-b') || st.build, node);
  // The game tooltip for a data-tt ref like "k:s10314" or "tal:Rogue:0:4", for preset b.
  const specFor = (v, b, node = null) => {
    const i = v.indexOf(':');
    const type = v.slice(0, i), id = v.slice(i + 1);
    const [p0, p1, p2] = id.split(':');
    switch (type) {
      case 'k': return spellSpec(id, b, false);
      case 'a': return spellSpec(id, b, true);
      case 'g': return gearSpec(id, b);
      case 'su': return setupSpec(id, b);
      case 'rot': return rotSpec(p0, p1, b);
      case 'tal': return talentSpec(p0, p1, p2, b);
      case 'r': return raceSpec(p0, p1, node?.closest('[data-tt-b]') ? b : null);
      case 'c': return classSpec(id);
      case 'w': return weightSpec(id, b);
      default: return null;
    }
  };

  const ttLine = (l) => {
    if (!Array.isArray(l)) return null;
    const [k, a, b] = l;
    if (k === 's') return el('div', { class: 'tt-sp' });
    // a detail kept out of the way: it opens on a tap without closing the tooltip
    if (k === 'd') return el('details', { class: 'tt-more' }, el('summary', { text: String(a ?? '') }), el('div', { class: 'c-m', text: String(b ?? '') }));
    if (k === '2' || k === 'h' || k === 'n') {
      return el('div', { class: 'tt-row' + (k === 'h' ? ' tt-h' : '') }, el('span', { class: k === 'n' ? 'c-m' : null, text: String(a ?? '') }),
        el('span', { class: 'c-t', text: String(b ?? '') }));
    }
    return el('div', { class: 'c-' + (['t', 'm', 'y', 'g'].includes(k) ? k : 't'), text: String(a ?? '') });
  };

  // A head row with an icon (a racial in a race's tooltip) starts an embedded card: a box with the icon, the head row
  // and the lines under it, up to the next head row or spacer.
  const ttLines = (lines) => {
    const out = [];
    let body = null;
    for (const l of arr(lines)) {
      const k = Array.isArray(l) ? l[0] : null;
      if (k === 'h' && l[3]) {
        body = el('div', { class: 'tt-body' }, ttLine(l));
        out.push(el('div', { class: 'tt-embed' }, iconImg(l[3], 'tt-ico'), body));
      } else if (body && k !== 'h' && k !== 's') body.append(ttLine(l));
      else { body = null; out.push(ttLine(l)); }
    }
    return out;
  };

  let ttEl = null, ttTippy = null, ttNode = null, ttPinned = false, ttAt = null;
  let ttHl = [], ttHideTimer = null;
  const hideTT = () => {
    clearTimeout(ttHideTimer);
    if (ttTippy) ttTippy.destroy();
    for (const n of ttHl) n.classList.remove('tt-on');
    ttHl = []; ttEl = null; ttTippy = null; ttNode = null; ttPinned = false;
  };
  // A game tooltip's card: the icon, then the title, its lines and its numbers. A compact card (one embedded in
  // another tooltip) is only the game's own text, without our numbers and the Wowhead link.
  const ttCard = (spec, compact = false) => {
    if (compact) spec = { ...spec, nums: [], link: null };
    const off = /^No effect: /.test(spec.why || '');
    const body = el('div', { class: 'tt-body' }, spec.why ? el('div', { class: 'tt-why' }, off ? el('b', { text: 'No effect' }) : null, off ? ' ' + spec.why.replace(/^No effect: /, '') : spec.why) : null,
      el('div', { class: 'tt-head' }, el('span', { class: 'tt-title', style: `color:${spec.color || GOLD}`, text: spec.title }), spec.sub ? el('span', { class: 'tt-sub', text: spec.sub }) : null),
      ...ttLines(spec.lines),
      ...(spec.nums && spec.nums.length ? [el('div', { class: 'tt-nums' }, spec.numsHead ? el('div', { class: 'c-m', text: spec.numsHead }) : null, ...spec.nums.map(ttLine))] : []),
      spec.link ? el('div', { class: 'tt-link' }, el('a', { href: spec.link.href, target: '_blank', rel: 'noopener noreferrer', text: spec.link.label + ' ↗' }), spec.link.note ? el('span', { text: spec.link.note }) : null) : null);
    return el('div', { class: 'fv-tt' + (compact ? ' tt-embed' : ''), role: 'tooltip' }, spec.icon ? iconImg(spec.icon, 'tt-ico') : null, body);
  };
  const showTT = (node, pt) => {
    let spec = null;
    try { spec = ttSpec(node); } catch (err) { console.error(err); }
    if (!spec) return;
    hideTT(); hideTip();
    const content = ttCard(spec);
    const r = node.getBoundingClientRect();
    // A wide target (a table row, a chart mark) gets the tooltip where the pointer was, not at its left edge.
    const wide = pt && (r.width > 160 || node instanceof SVGElement);
    const dx = wide ? pt.x - 12 - r.left : 0;
    ttTippy = tippy(node, {
      content, trigger: 'manual', interactive: true, appendTo: document.body, hideOnClick: false, placement: 'bottom-start', offset: [0, 6], duration: 0,
      getReferenceClientRect: wide ? () => { const q = node.getBoundingClientRect(); return { width: 0, height: q.height, x: q.left + dx, y: q.top, left: q.left + dx, right: q.left + dx, top: q.top, bottom: q.bottom }; } : null,
    });
    ttTippy.show();
    ttEl = ttTippy.popper;
    ttNode = node; ttAt = { x: r.left, y: r.top };
    const par = node.parentElement, v = node.getAttribute('data-tt');
    ttHl = par && par.classList.contains('fv-sbar') ? [...par.children].filter((c) => c.getAttribute('data-tt') === v) : [node];
    for (const n of ttHl) n.classList.add('tt-on');
  };

  {
    const root = $('#fv');
    let lpTimer = null, lpFired = false, lpStart = null;
    // The tooltip is in body, outside the report, so we watch the whole page. That way moving onto the tooltip
    // keeps it up.
    document.addEventListener('pointerover', (e) => {
      if (e.pointerType !== 'mouse') return;
      if (ttEl && ttEl.contains(e.target)) { clearTimeout(ttHideTimer); return; }
      const n = e.target.closest('[data-tt]');
      if (n && n === ttNode) clearTimeout(ttHideTimer);
      if (!n || n === ttNode) return;
      if (!ttPinned) showTT(n, { x: e.clientX, y: e.clientY });
    });
    document.addEventListener('pointerout', (e) => {
      if (e.pointerType !== 'mouse') return;
      const into = (x) => x && e.relatedTarget && x.contains(e.relatedTarget);
      // a short grace period lets the mouse cross over to the tooltip and its wowhead link
      if (ttEl && ttEl.contains(e.target)) { if (!ttPinned && !into(ttEl) && !into(ttNode)) { clearTimeout(ttHideTimer); ttHideTimer = setTimeout(hideTT, 250); } return; }
      const n = e.target.closest('[data-tt]');
      if (!n || into(n)) return;
      if (n === ttNode && !ttPinned && !into(ttEl)) { clearTimeout(ttHideTimer); ttHideTimer = setTimeout(hideTT, 250); }
    });
    root.addEventListener('pointerdown', (e) => {
      clearTimeout(lpTimer); lpFired = false;
      if (e.pointerType === 'mouse') return;
      const n = e.target.closest('[data-tt]');
      if (!n || n.hasAttribute('data-tt-tap')) return;
      lpStart = { x: e.clientX, y: e.clientY };
      lpTimer = setTimeout(() => { lpFired = true; showTT(n, lpStart); ttPinned = true; }, 450);
    });
    root.addEventListener('pointermove', (e) => { if (lpStart && Math.hypot(e.clientX - lpStart.x, e.clientY - lpStart.y) > 10) clearTimeout(lpTimer); });
    for (const ev of ['pointerup', 'pointercancel']) root.addEventListener(ev, () => { clearTimeout(lpTimer); lpStart = null; });
    root.addEventListener('contextmenu', (e) => { if (lpFired || lpStart) e.preventDefault(); });
    document.addEventListener('click', (e) => {
      if (lpFired) { lpFired = false; e.preventDefault(); e.stopPropagation(); return; }
      if (ttEl && ttEl.contains(e.target)) {
        // the wowhead link opens in a new tab; any other tap on the tooltip closes it
        if (e.target.closest('.tt-link a')) { e.stopPropagation(); setTimeout(hideTT, 0); return; }
        // a folded detail opens and closes in place, and the tooltip stays
        if (e.target.closest('.tt-more')) { e.stopPropagation(); ttPinned = true; return; }
        hideTT(); e.preventDefault(); e.stopPropagation(); return;
      }
      const n = e.target.closest ? e.target.closest('[data-tt]') : null;
      // A toggle, like a class icon in the class filter, does its job on click, so its tooltip closes instead of
      // staying pinned.
      if (n && n.hasAttribute('aria-pressed')) { hideTT(); return; }
      if (n && n.hasAttribute('data-tt-tap')) {
        if (n === ttNode && ttPinned) hideTT();
        else { showTT(n, { x: e.clientX, y: e.clientY }); ttPinned = true; }
        return;
      }
      if (n !== ttNode) hideTT();
      else if (n) ttPinned = true;
    }, true);
    document.addEventListener('keydown', (e) => { if (e.key === 'Escape') hideTT(); });
    // a scroll box that moves the tooltip's target away (the timeline, a wide table) closes it
    document.addEventListener('scroll', (e) => {
      if (!ttNode || e.target === document || !e.target.contains || !e.target.contains(ttNode)) return;
      const r = ttNode.getBoundingClientRect();
      if (Math.abs(r.left - ttAt.x) > 24 || Math.abs(r.top - ttAt.y) > 24) hideTT();
    }, true);
  }

  // ---------- tabs ----------
  const pages = ['ranking', 'spec', 'races', 'compare', 'method'];
  const showTab = (tab, link = true) => {
    const changed = st.tab !== tab;
    st.tab = tab; persist();
    if (changed) { const pg = document.getElementById(tab); pg.classList.remove('in'); void pg.offsetWidth; pg.classList.add('in'); }
    document.querySelectorAll('.fv-tab').forEach((b) => { b.setAttribute('aria-selected', String(b.dataset.tab === tab)); b.classList.toggle('active', b.dataset.tab === tab); });
    document.querySelectorAll('.fv-page').forEach((p) => { p.hidden = p.dataset.page !== tab; p.classList.toggle('active', p.dataset.page === tab); });
    if (link) dash.setLink(tab === 'spec' && st.build ? st.build : tab);
    hideTip(); hideTT();
    render();
  };
  document.querySelectorAll('.fv-tab').forEach((b) => b.addEventListener('click', (e) => { e.preventDefault(); showTab(b.dataset.tab); }));
  const openSpec = (id) => {
    id = dutyBase(id);
    const swap = st.tab === 'spec' && st.build !== id;
    st.build = id; st.sample = 0; persist(); showTab('spec'); $('.sim-ui').scrollTo({ top: 0 });
    if (swap) for (const n of [$('#sp-head'), $('#sp-tiles')]) { n.classList.remove('fv-swap'); void n.offsetWidth; n.classList.add('fv-swap'); }
  };
  // swipe the spec banner for the next or previous preset
  {
    let sx = null, sy = null;
    const head = $('#sp-head');
    head.addEventListener('touchstart', (e) => { const t = e.touches[0]; sx = t.clientX; sy = t.clientY; }, { passive: true });
    head.addEventListener('touchend', (e) => {
      if (sx == null) return;
      const t = e.changedTouches[0], dx = t.clientX - sx, dy = t.clientY - sy;
      sx = null;
      if (Math.abs(dx) > 60 && Math.abs(dy) < 45) (dx < 0 ? $('#sp-next') : $('#sp-prev')).click();
    }, { passive: true });
  }
  // jump to a section of the spec page
  // It sticks under the header, and the section we are reading gets the gold line.
  {
    const cards = [...document.querySelectorAll('#spec [data-jump]')];
    const bar = $('#sp-jump-bar'), scroller = document.querySelector('.dps-report-sim-ui') || document.scrollingElement;
    let picked = -1;
    const links = cards.map((card, i) => {
      const bt = el('button', { class: 'nav-link', type: 'button', text: card.dataset.jump });
      bt.addEventListener('click', () => {
        picked = i;
        const d = card.querySelector(':scope > details'); if (d) d.open = true;
        // Aim for where the bar sits once it sticks under the header, not where it is now.
        const stuckBottom = document.querySelector('.sim-header').getBoundingClientRect().bottom + bar.offsetHeight;
        const top = card.getBoundingClientRect().top - stuckBottom - 8;
        scroller.scrollBy({ top, behavior: matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth' });
      });
      return bt;
    });
    $('#sp-jump').replaceChildren(...links.map((bt) => el('li', { class: 'nav-item' }, bt)));
    const spy = () => {
      if (bar.offsetParent === null) return;
      const r = bar.getBoundingClientRect(), head = document.querySelector('.sim-header').getBoundingClientRect().bottom;
      bar.classList.toggle('stuck', r.top <= head + 1);
      // Two cards side by side start at the same height. Of those, the one we clicked wins, or else the left one.
      const tops = cards.map((c) => c.getBoundingClientRect().top);
      const passed = tops.filter((t) => t <= r.bottom + 24);
      const rowTop = passed.length ? Math.max(...passed) : tops[0];
      const row = cards.map((_, i) => i).filter((i) => Math.abs(tops[i] - rowTop) < 4);
      const on = row.includes(picked) ? picked : row[0];
      links.forEach((bt, i) => bt.classList.toggle('active', i === on));
    };
    scroller.addEventListener('scroll', spy, { passive: true });
    new ResizeObserver(spy).observe(bar);
  }

  // ---------- shared pieces ----------
  // The class filter looks like the sim UI's race picker: a label that says what we picked, then a row of class icons.
  // The picked class is lit with a gold ring and the others are dimmed. Clicking the picked class again shows all.
  //
  // avail, when given, is the classes to offer. The others are hidden.
  const classChips = (host, key, avail) => {
    const m = iconMap();
    const name = st[key] === 'All' ? 'All classes' : st[key];
    if (host.childElementCount && (host.dataset.icons === '1' || !m)) {
      host.querySelector('.fv-pick-value').textContent = name;
      host.querySelectorAll('button').forEach((b) => { b.setAttribute('aria-pressed', String(b.dataset.c === st[key])); b.classList.toggle('active', b.dataset.c === st[key]); b.hidden = !!avail && !avail.has(b.dataset.c); });
      return;
    }
    host.dataset.icons = m ? '1' : '0';
    host.classList.add('fv-icon-picker');
    host.replaceChildren(el('span', { class: 'fv-label fv-rowlabel', text: 'Class' }),
      el('div', { class: 'fv-pick-options' }, ...CLASSES.map((c) => {
        const u = m && m.get('classicon_' + c.toLowerCase());
        const b = el('button', { type: 'button', class: 'fv-pick-option' + (c === st[key] ? ' active' : ''), 'data-c': c, 'aria-label': c, 'aria-pressed': String(c === st[key]), 'data-tt': 'c:' + c, style: u ? `background-image:url("${u}")` : `background:${ccol(c)}` });
        b.addEventListener('click', () => { st[key] = st[key] === c ? 'All' : c; persist(); render(); });
        b.hidden = !!avail && !avail.has(c);
        return b;
      })), el('span', { class: 'fv-pick-value', text: name }));
  };
  const fail = (host) => { host.replaceChildren(el('div', { class: 'fv-fail', text: 'Failed to load' })); };
  const waitOr = (host, ids) => {
    const ss = ids.map(statusOf);
    if (ss.some((s) => s === 'error' || s === 'connect' || s === 'declined' || s === 'missing')) { fail(host); return false; }
    if (ss.some((s) => s !== 'ok')) { host.classList.add('dash-skeleton'); return false; }
    host.classList.remove('dash-skeleton');
    return true;
  };
  // A Spec card never shows another preset: while this preset's data loads its old content is cleared, and every
  // drawn host is stamped with the preset it shows (data-build).
  const specWait = (host, ids, b, extras = []) => {
    if (waitOr(host, ids)) { host.dataset.build = b.id; return true; }
    if (host.querySelector(':scope > .fv-fail')) { host.dataset.build = b.id; return false; }
    if (host.dataset.build !== b.id) { host.replaceChildren(); for (const x of extras) if (x) { x.replaceChildren(); (x.closest('table') || x).dataset.build = ''; } host.dataset.build = ''; }
    return false;
  };
  const sortedBuildOptions = (sel, builds, value) => {
    if (sel.dataset.filled !== String(builds.length)) {
      sel.replaceChildren();
      for (const c of CLASSES) {
        const g = el('optgroup', { label: c });
        builds.filter((b) => b.class === c).sort((a, b) => b.dps - a.dps).forEach((b) => g.append(el('option', { value: b.id, text: `${b.label} · ${f1(b.dps)}` })));
        if (g.childElementCount) sel.append(g);
      }
      sel.dataset.filled = String(builds.length);
    }
    sel.value = value;
  };
  // A dropdown with an icon on each option, in place of a <select> (which can't show images). The button looks
  // like the page's selects, and the menu like the sim UI's dropdowns. A re-render (when a dataset finishes
  // loading) keeps an open menu open.
  //
  // items are { header } rows or { value, text, icon, end, disabled } options. icon is a function, because we
  // need a fresh icon element for the button and for the menu.
  const iselShow = (host, open) => {
    host.classList.toggle('open', open);
    host.querySelector('.fv-isel-menu')?.classList.toggle('show', open);
    host.querySelector('.fv-isel-btn')?.setAttribute('aria-expanded', String(open));
  };
  const iconSelect = (host, items, value, onPick, btnText) => {
    host.classList.add('fv-isel');
    const open = host.classList.contains('open');
    const cur = items.find((it) => !it.header && it.value === value);
    const btn = el('button', { type: 'button', class: 'form-select fv-isel-btn', 'aria-haspopup': 'listbox', 'aria-expanded': String(open), 'aria-label': host.dataset.label || '' },
      cur?.icon?.() || '', el('span', { class: 'fv-isel-text', text: cur ? btnText?.(cur) ?? cur.text : '' }));
    const menu = el('div', { class: 'dropdown-menu fv-isel-menu' + (open ? ' show' : ''), role: 'listbox' },
      ...items.map((it) => (it.header ? el('h6', { class: 'dropdown-header', style: it.color ? `color:${it.color}` : '', text: it.header })
        : el('button', { type: 'button', role: 'option', class: 'dropdown-item' + (it.value === value ? ' active' : ''), 'aria-selected': String(it.value === value), 'data-v': it.value, ...(it.disabled ? { disabled: '' } : {}) },
          it.icon?.() || '', el('span', { class: 'fv-isel-text', text: it.text }), it.end ? el('small', { class: 'fv-isel-end', text: it.end }) : ''))));
    const opts = () => [...menu.querySelectorAll('.dropdown-item:not([disabled])')];
    btn.addEventListener('click', (e) => {
      e.stopPropagation();
      const o = !host.classList.contains('open');
      document.querySelectorAll('.fv-isel.open').forEach((h) => h !== host && iselShow(h, false));
      iselShow(host, o);
      // We scroll the picked option into the middle of the menu, without scrolling the page.
      const a = menu.querySelector('.active');
      if (o && a) menu.scrollTop = a.offsetTop - (menu.clientHeight - a.offsetHeight) / 2;
    });
    btn.addEventListener('keydown', (e) => {
      if (e.key !== 'ArrowDown' && e.key !== 'ArrowUp') return;
      e.preventDefault();
      if (!host.classList.contains('open')) btn.click();
      (menu.querySelector('.active:not([disabled])') || opts()[0])?.focus();
    });
    menu.addEventListener('keydown', (e) => {
      const all = opts(), i = all.indexOf(document.activeElement);
      if (e.key === 'ArrowDown' || e.key === 'ArrowUp') { e.preventDefault(); all[Math.max(0, Math.min(all.length - 1, i + (e.key === 'ArrowDown' ? 1 : -1)))]?.focus(); }
      if (e.key === 'Escape' || e.key === 'Tab') { iselShow(host, false); if (e.key === 'Escape') btn.focus(); }
    });
    menu.addEventListener('click', (e) => {
      const it = e.target.closest('.dropdown-item');
      if (!it || it.disabled) return;
      e.stopPropagation();
      iselShow(host, false);
      btn.focus();
      if (it.dataset.v !== value) onPick(it.dataset.v);
    });
    host.replaceChildren(btn, menu);
  };
  document.addEventListener('click', (e) => { for (const h of document.querySelectorAll('.fv-isel.open')) if (!h.contains(e.target)) iselShow(h, false); });
  document.addEventListener('keydown', (e) => { if (e.key === 'Escape') document.querySelectorAll('.fv-isel.open').forEach((h) => iselShow(h, false)); });
  // The Specs tab's preset picker: each preset with its spec icon and DPS, grouped by class (best first) under a
  // header in the class colour. The hidden #sp-pick select keeps the same choice for scripts that drive the page.
  const specPicker = (builds, value) => {
    const items = CLASSES.flatMap((c) => {
      const bs = builds.filter((b) => b.class === c).sort((a, b) => b.dps - a.dps);
      return bs.length ? [{ header: c, color: ccol(c) }, ...bs.map((b) => ({ value: b.id, text: b.build, label: `${b.build} · ${b.class}`, icon: () => specIco(b, ''), end: f1(b.dps) }))] : [];
    });
    iconSelect($('#sp-isel'), items, value, (v) => openSpec(v), (it) => it.label);
  };
  const buildOrder = (builds) => CLASSES.flatMap((c) => builds.filter((b) => b.class === c).sort((a, b) => b.dps - a.dps));

  // ---------- ranking ----------
  // a segmented control bound to one state key
  const seg = (host, key, opts, onPick) => {
    const sig = opts.map((o) => o[0] + (o[2] ? '+' : '')).join('|');
    if (host.dataset.built !== sig) {
      host.replaceChildren(...opts.map(([v, lbl, ico]) => {
        const b = el('button', { 'data-v': v }, ico || null, lbl);
        b.addEventListener('click', (e) => { e.stopPropagation(); st[key] = v; persist(); hideTip(); if (onPick) onPick(v); else render(); });
        return b;
      }));
      host.dataset.built = sig;
    }
    host.querySelectorAll('button').forEach((b) => b.setAttribute('aria-pressed', String(b.dataset.v === String(st[key]))));
  };
  const SORTS = [['dps', 'Mean'], ['p10', 'Unlucky'], ['p90', 'Lucky'], ['cv', 'Steadiest']];
  // a duty preset plays its base preset's role
  const roleOf = () => {
    const m = new Map((rowsOf('roles') || []).map((r) => [r.build, r.role]));
    for (const b of rowsOf('builds') || []) if (isDuty(b) && !m.has(b.id)) m.set(b.id, m.get(b.baseBuild));
    return m;
  };
  // Each part of a ranking bar names itself in a small tooltip while the mouse is on it: the mean bar, the box
  // of the middle half of fights, the 10th to 90th percentile whisker and the median tick. They're too thin to
  // aim at one by one, so we go by where the mouse is along the bar. All bars share one tooltip.
  //
  // The tooltip is one block of inline text, so the sim UI's sticky tooltips don't offer to pin it.
  let partTippy = null;
  const partTip = (track, b, max) => {
    let last = null;
    const hide = () => { partTippy?.hide(); last = null; };
    track.addEventListener('pointermove', (e) => {
      if (e.pointerType !== 'mouse') return;
      const r = track.getBoundingClientRect(), v = (e.clientX - r.left) / r.width * max, pxPerDps = r.width / max;
      const part = Math.abs(v - b.p50) * pxPerDps <= 4 ? ['Median', f1(b.p50)]
        : v >= b.p25 && v <= b.p75 ? ['Middle half of fights', `${f1(b.p25)}–${f1(b.p75)}`]
          : v >= b.p10 && v <= b.p90 ? ['10th–90th percentile', `${f1(b.p10)}–${f1(b.p90)}`]
            : v <= b.dps ? ['Mean', f1(b.dps)] : null;
      if (!part) { hide(); return; }
      if (!partTippy) partTippy = tippy(el('span'), { trigger: 'manual', appendTo: document.body, placement: 'top', offset: [0, 12], duration: 0, content: '' });
      if (last !== part[0]) { partTippy.setContent(el('span', { class: 'fv-parttip' }, el('span', { class: 't', text: part[0] }), el('br'), el('b', { text: part[1] }), ' DPS')); last = part[0]; }
      const px = e.clientX, py = r.top + r.height / 2;
      partTippy.setProps({ getReferenceClientRect: () => ({ width: 0, height: 0, x: px, y: py, left: px, right: px, top: py, bottom: py }) });
      partTippy.show();
    });
    track.addEventListener('pointerleave', hide);
  };
  const drawRanking = () => {
    const host = $('#rank-chart');
    classChips($('#rank-classes'), 'cls');
    const lens = statusOf('lengths') === 'ok' ? [...new Set([180, ...rowsOf('lengths').map((r) => +r.secs)])].sort((a, b) => a - b).map(String) : ['180'];
    seg($('#rank-len'), 'len', lens.map((v) => [v, v + 's']));
    $('#rank-lenbar').classList.toggle('off', st.len !== '180');
    seg($('#rank-sort'), 'sort', SORTS);
    seg($('#rank-role'), 'role', [['All', 'All roles'], ['Melee', 'Melee'], ['Ranged', 'Ranged'], ['Caster', 'Caster']]);
    // With a class picked, the roles it has no preset for are off. A class with one role shows that role, with the
    // whole role filter off. The picked role comes back when the class filter is cleared.
    const roleMap = roleOf();
    const clsRoles = st.cls !== 'All' ? new Set((rankB() || []).filter((b) => b.class === st.cls).map((b) => roleMap.get(b.id)).filter(Boolean)) : null;
    const role = !clsRoles?.size ? st.role : clsRoles.size === 1 ? [...clsRoles][0] : clsRoles.has(st.role) ? st.role : 'All';
    $('#rank-role').querySelectorAll('button').forEach((bt) => {
      const v = bt.dataset.v;
      bt.setAttribute('aria-pressed', String(v === role));
      bt.disabled = !!clsRoles?.size && (clsRoles.size === 1 || (v !== 'All' && !clsRoles.has(v)));
    });
    const fi = (n) => (iconMap() ? iconImg(n, 'fv-ico') : null);
    seg($('#rank-fac'), 'fac', [['All', 'Both factions'], ['Horde', 'Horde', fi('ui_hordeicon')], ['Alliance', 'Alliance', fi('ui_allianceicon')]]);
    // With one class picked there's nothing to pick the best of, so the chip is off until the class filter is cleared.
    const oneCls = st.cls !== 'All', best = st.best && !oneCls;
    $('#rank-best').disabled = oneCls;
    $('#rank-best').title = oneCls ? 'Only one class is shown' : '';
    $('#rank-best').setAttribute('aria-pressed', String(best));
    $('#rank-enc').setAttribute('aria-pressed', String(!!st.enc));
    $('#rank-duty').setAttribute('aria-pressed', String(!!st.duty));
    const filtered = st.cls !== 'All' || (role !== 'All' && !(clsRoles?.size === 1)) || st.fac !== 'All' || best;
    $('#rank-reset').hidden = !filtered;
    if (!waitOr(host, ['builds', 'lengths'])) return;
    const builds = rankB(), lengths = rowsOf('lengths'), roles = roleOf();
    const L = +st.len;
    $('#rank-sort').parentElement.hidden = L !== 180;
    const lenMap = new Map(lengths.filter((r) => r.secs === L).map((r) => [r.build, r]));
    let rows = builds.filter((b) => (st.cls === 'All' || b.class === st.cls) && (st.fac === 'All' || b.faction === st.fac) && (role === 'All' || roles.get(b.id) === role));
    const val = (b) => (L === 180 ? b.dps : (lenMap.get(b.id)?.dps ?? 0) * asmFactor(b.id, L));
    const sortKey = L === 180 ? st.sort : 'dps';
    const key = { dps: val, p10: (b) => b.p10, p90: (b) => b.p90, cv: (b) => -b.cv }[sortKey] || val;
    rows = rows.sort((a, b) => key(b) - key(a));
    if (best) { const seen = new Set(); rows = rows.filter((b) => !seen.has(b.class) && seen.add(b.class)); }
    const max = d3.max(builds, (b) => Math.max(L === 180 ? b.p90 : 0, val(b))) * 1.02;
    const x = (v) => (100 * v / max) + '%';
    const top = d3.max(builds, val);
    if (!rows.length) { host.replaceChildren(el('p', { class: 'fv-note', text: 'No preset matches.' })); return; }
    host.replaceChildren(...rows.map((b, i) => {
      const v = val(b);
      const valEl = el('span', { class: 'fv-row-val', text: f1(v) });
      if (L === 180) mark(valEl, 'builds', 'dps', 'id=' + enc(b.id));
      else mark(valEl, 'lengths', 'dps', `build=${enc(b.id)}&secs=${L}`);
      const gap = 100 * (1 - v / top);
      const extra = L === 180 && sortKey === 'p10' ? `10th ${f1(b.p10)}` : L === 180 && sortKey === 'p90' ? `90th ${f1(b.p90)}` : L === 180 && sortKey === 'cv' ? `spread ${f2(b.cv)}%` : gap < 0.05 ? 'top' : `−${f1(gap)}%`;
      const track = el('div', { class: 'fv-track' }, el('span', { class: 'fv-bar', style: `width:${x(v)};background:${ccol(b.class)}` }));
      if (L === 180) {
        track.append(el('span', { class: 'fv-whisk', style: `left:${x(b.p10)};width:calc(${x(b.p90)} - ${x(b.p10)})` }),
          el('span', { class: 'fv-box', style: `left:${x(b.p25)};width:calc(${x(b.p75)} - ${x(b.p25)})` }),
          el('span', { class: 'fv-med', style: `left:${x(b.p50)}` }));
        partTip(track, b, max);
      }
      const row = el('button', { class: 'fv-row' + (i === 0 && rows.length > 1 ? ' top' : ''), 'data-tip': '1', 'data-id': b.id, 'aria-label': `${b.label}, ${f1(v)} DPS` },
        el('span', { class: 'fv-row-pos', text: String(i + 1) }),
        specIco(b) || el('span', { class: 'fv-dot', style: `background:${ccol(b.class)}` }),
        el('span', { class: 'fv-row-main' },
          el('span', { class: 'fv-row-name' }, el('b', { text: isEnc(b) ? String(b.build).replace(/ · vs .*$/, '') : isDuty(b) ? String(b.build).replace(/ · [^·]* duty$/, '') : b.build }), el('span', { text: '· ' + b.class })),
          el('span', { class: 'fv-row-sub' }, tt(raceIco(b.raceIcon), `r:${b.race}:${b.class}`, false, b.id), el('em', { text: b.race }),
            iconMap() ? iconImg(b.faction === 'Horde' ? 'ui_hordeicon' : 'ui_allianceicon', 'fv-ico') : null,
            isEnc(b) ? encBadge(b) : null, isDuty(b) ? dutyBadge(b, builds) : null)),
        el('span', { class: 'fv-row-end' }, valEl, el('span', { class: 'fv-row-gap', text: extra })),
        track);
      row.addEventListener('click', (e) => {
        e.stopPropagation();
        if (row.classList.contains('is-hl')) { openSpec(b.id); return; }
        const card = $('#rank-card'), cr = card.getBoundingClientRect(), rr = row.getBoundingClientRect();
        const behind = gap < 0.05 ? el('span', { class: 'fv-tip-m', text: 'top' }) : tipDelta(-gap);
        const race = el('span', {}, raceIco(b.raceIcon), ` ${b.race}`);
        const se = lenMap.get(b.id)?.se;
        const lines = L === 180
          ? [tipKey(f1(b.dps), 'DPS', `± ${f2(b.se)}`), tipGrid([['Rank', el('span', {}, `${rankOf(b, builds)} of ${builds.length} · `, behind)], ['10th–90th', `${f1(b.p10)}–${f1(b.p90)}`],
            ['Middle half', `${f1(b.p25)}–${f1(b.p75)}`], ['Spread', `${f2(b.cv)}%`], ['Race', race], isEnc(b) ? ['Boss', `${b.encounter} · advantaged`] : null, isDuty(b) ? ['Duty', dutyCostEl(b, builds)] : null])]
          : [tipKey(f1(v), 'DPS', se != null ? `± ${f2(se)} at ${L}s` : `at ${L}s`), tipGrid([['vs 180s', el('span', {}, `${f1(b.dps)} · `, tipDelta(100 * (v / b.dps - 1)))], ['vs the top', behind], ['Race', race]])];
        showTip(card, rr.left - cr.left + 30, rr.bottom - cr.top + 6, tipTitle(b.label, null, specIco(b, '')), lines, { hl: row, open: b.id, below: true });
      });
      return row;
    }));
  };
  $('#rank-best').addEventListener('click', (e) => { e.stopPropagation(); st.best = !st.best; persist(); render(); });
  $('#rank-duty').addEventListener('click', (e) => { e.stopPropagation(); st.duty = !st.duty; persist(); hideTip(); render(); });
  $('#rank-enc').addEventListener('click', (e) => { e.stopPropagation(); st.enc = !st.enc; if (!st.enc && st.build && isEnc((rowsOf('builds') || []).find((b) => b.id === st.build))) st.build = (rowsOf('builds') || []).find((b) => b.id === st.build).baseBuild; persist(); hideTip(); render(); });
  $('#rank-reset').addEventListener('click', (e) => { e.stopPropagation(); Object.assign(st, { cls: 'All', role: 'All', fac: 'All', best: false }); persist(); render(); });

  // line chart of DPS by fight length; tap or hover anywhere for a crosshair readout of the nearest line
  const drawLenChart = (host, series, opts = {}) => {
    const W = host.clientWidth, H = opts.h || Math.max(220, Math.min(300, W * 0.7));
    const m = { l: 42, r: opts.endLabels ? 70 : 14, t: 12, b: 28 };
    const xs = [...new Set(series.flatMap((s) => s.pts.map((p) => p.s)))].sort((a, b) => a - b);
    const x = d3.scaleLinear(d3.extent(xs), [m.l + 8, W - m.r]);
    const all = series.flatMap((s) => s.pts.map((p) => p.v)).concat(opts.ref != null ? [opts.ref] : []);
    const y = d3.scaleLinear(opts.yDomain || d3.extent(all), [H - m.b, m.t]).nice();
    const xTicks = opts.xTicks || xs, xFmt = opts.xFmt || ((d) => d + 's');
    const svg = d3.create('svg').attr('width', '100%').attr('viewBox', `0 0 ${W} ${H}`).attr('height', H);
    svg.append('g').attr('transform', `translate(0,${H - m.b})`).call(d3.axisBottom(x).tickValues(xTicks).tickFormat((d, i) => (W >= 480 || i === xTicks.length - 1 ? xFmt(d) : String(d))).tickSizeOuter(0)).attr('font-size', null).attr('font-family', null)
      .call((g) => g.selectAll('line,path').style('stroke', 'var(--cds-chart-axis)'));
    svg.append('g').attr('transform', `translate(${m.l},0)`).call(d3.axisLeft(y).ticks(5).tickFormat(opts.fmt || f0).tickSize(-(W - m.l - m.r))).attr('font-size', null).attr('font-family', null)
      .call((g) => { g.select('.domain').remove(); g.selectAll('line').style('stroke', 'var(--cds-chart-grid)'); });
    if (opts.ref != null) svg.append('line').attr('x1', m.l).attr('x2', W - m.r).attr('y1', y(opts.ref)).attr('y2', y(opts.ref)).style('stroke', 'var(--cds-chart-reference)').attr('stroke-dasharray', '4 3');
    const line = d3.line().x((p) => x(p.s)).y((p) => y(p.v));
    const paths = new Map();
    for (const s of series) {
      paths.set(s, svg.append('path').attr('d', line(s.pts)).style('fill', 'none').style('stroke', s.color).attr('stroke-width', s.w || 2)
        .attr('stroke-linecap', 'round').attr('stroke-linejoin', 'round').attr('opacity', s.op ?? 1));
      if (s.dots) svg.append('g').selectAll('circle').data(s.pts).join('circle').attr('cx', (p) => x(p.s)).attr('cy', (p) => y(p.v)).attr('r', 3.5).style('fill', s.color).style('pointer-events', 'none');
      if (opts.endLabels) {
        const last = s.pts[s.pts.length - 1];
        svg.append('text').attr('x', x(last.s) + 6).attr('y', y(last.v) + 4).text(`${s.short || s.name} ${opts.fmt ? opts.fmt(last.v) : f0(last.v)}`).style('fill', s.color);
      }
    }
    const rule = svg.append('line').attr('y1', m.t).attr('y2', H - m.b).style('stroke', 'var(--color-fg)').attr('stroke-dasharray', '3 3').attr('opacity', 0).style('pointer-events', 'none');
    const dot = svg.append('circle').attr('r', 5).style('fill', 'var(--color-panel)').attr('stroke-width', 2.5).attr('opacity', 0).style('pointer-events', 'none');
    const pickable = series.filter((s) => s.pick !== false);
    const focusOn = (s) => { for (const [ss, pa] of paths) pa.attr('opacity', !s ? ss.op ?? 1 : ss === s ? 1 : Math.min(ss.op ?? 1, 0.12)).attr('stroke-width', ss === s ? 3 : ss.w || 2); };
    // With no line asked for, the picked one (opts.hlId) stays lit, or none.
    const highlight = (id) => { const want = id || opts.hlId; focusOn(want ? series.find((q) => q.id === want) || null : null); };
    // The tooltip shows where the pointer is. While the mouse stays on the same line and length, we leave the tooltip
    // where it is, so the mouse can reach its button.
    let shown = null;
    const leave = () => { hideTip(); rule.attr('opacity', 0); dot.attr('opacity', 0); highlight(null); if (opts.onHover) opts.onHover(null); };
    const pick = (e, pinned, now) => {
      const [px, py] = d3.pointer(e, svg.node());
      const sx = xs.reduce((a, b) => (Math.abs(x(b) - px) < Math.abs(x(a) - px) ? b : a));
      let best = null, bd = Infinity;
      for (const s of pickable) { const p = s.pts.find((q) => q.s === sx); if (p && Math.abs(y(p.v) - py) < bd) { bd = Math.abs(y(p.v) - py); best = [s, p]; } }
      if (!best) return;
      const [s, p] = best;
      const key = s.id + '|' + sx;
      if (!pinned && tipEl && shown === key) { clearTimeout(tipWait); return; }
      if (!pinned && !now && tipEl) { tipHover(() => pick(e, false, true)); return; }
      shown = key;
      rule.attr('x1', x(sx)).attr('x2', x(sx)).attr('opacity', 0.5);
      dot.attr('cx', x(sx)).attr('cy', y(p.v)).style('stroke', s.color).attr('opacity', 1);
      if (opts.focus) focusOn(s);
      if (opts.onHover) opts.onHover(s.id);
      // A few lines: the others' values at this length in a grid. Many lines: a small table of the picked line at every
      // length takes their place.
      const others = series.length <= 3 ? tipGrid(series.filter((q) => q !== s).map((q) => { const pp = q.pts.find((r) => r.s === sx); return pp ? [el('span', {}, tipSw(q.color), q.short || q.name), opts.tipMain ? opts.tipMain(pp)[0] : `${f1(pp.v)} DPS`] : null; })) : null;
      const [hx, hy] = d3.pointer(e, host);
      const strip = series.length > 3 && opts.strip !== false ? el('div', { class: 'fv-tip-strip' }, ...s.pts.flatMap((q) => [
        el('span', { class: 'k' + (q.s === sx ? ' on' : ''), text: `${q.s}s` }),
        el('span', { class: 'v' + (q.s === sx ? ' on' : ''), text: opts.fmt ? opts.fmt(q.v) : f0(q.v) }),
      ])) : null;
      // tipMain gives the key figure and what it compares to; tipRows any further [label, value] rows
      const main = opts.tipMain ? opts.tipMain(p) : [`${f1(p.v)} DPS`];
      showTip(host, hx, hy, tipTitle(s.name, opts.tipSub ? opts.tipSub(sx) : `${sx}s`, tipSw(s.color)), [tipKey(main[0], null, main[1]), opts.tipRows ? tipGrid(opts.tipRows(p)) : null, others, strip],
        { open: s.id && opts.open ? s.id : null, pinned, onLeave: leave, onHide: () => { shown = null; rule.attr('opacity', 0); dot.attr('opacity', 0); } });
      if (pinned && opts.onPick) opts.onPick(s.id);
    };
    if (opts.refLabel && opts.ref != null) svg.append('text').attr('x', W - m.r).attr('y', y(opts.ref) - 4).attr('text-anchor', 'end').style('fill', 'var(--color-fg-muted)').text(opts.refLabel);
    svg.append('rect').attr('x', m.l).attr('y', 0).attr('width', Math.max(0, W - m.l - m.r + 8)).attr('height', H - m.b).style('fill', 'transparent').attr('data-tip', '1').style('cursor', 'crosshair')
      .on('click', (e) => { e.stopPropagation(); pick(e, true); })
      .on('pointermove', (e) => { if (e.pointerType === 'mouse' && !tipPinned) pick(e, false); })
      .on('pointerleave', (e) => { if (e.pointerType === 'mouse' && !tipPinned) tipHover(leave); });
    host.replaceChildren(svg.node());
    highlight(null);
    // Something outside the chart, like a list of the same presets, can light up one line.
    return { highlight };
  };

  const drawLenAll = () => {
    const host = $('#lenchart');
    if (!waitOr(host, ['lengths', 'builds'])) return;
    const builds = visB(), lengths = rowsOf('lengths');
    const by = d3.group(lengths, (r) => r.build);
    const series = builds.map((b) => {
      const pts = (by.get(b.id) || []).sort((p, q) => p.secs - q.secs);
      const base = pts.find((p) => p.secs === 180)?.dps || 1;
      const on = st.cls === 'All' || b.class === st.cls;
      return { id: b.id, name: b.label, color: ccol(b.class), op: on ? 0.9 : 0.08, pick: on, w: on && st.cls !== 'All' ? 2 : 1.5, dots: on && st.cls !== 'All', pts: pts.map((p) => ({ s: p.secs, v: (100 * p.dps * asmFactor(b.id, p.secs)) / (base * asmFactor(b.id, 180)), dps: p.dps * asmFactor(b.id, p.secs) })) };
    }).sort((a, b) => a.op - b.op);
    drawLenChart(host, series, { fmt: (v) => d3.format('.0f')(v) + '%', ref: 100, focus: true, open: true, hlId: st.hlLen, onPick: (id) => { st.hlLen = id; persist(); }, tipMain: (p) => [`${f1(p.dps)} DPS`, p.s === 180 ? null : el('span', {}, tipDelta(p.v - 100), ' vs 180s')] });
  };

  // ---------- mana ----------
  // How a preset's mana runs down over a long fight, from 100 logged 480s fights each. The chart is the average mana
  // and the list says when a typical fight first drops below 10%. Warriors and rogues don't use mana.
  const mmss = (v) => `${Math.floor(v / 60)}:${String(Math.round(v % 60)).padStart(2, '0')}`;
  const drawMana = () => {
    const host = $('#manachart'), list = $('#mana-list');
    if (!waitOr(host, ['mana', 'mana_summary', 'builds'])) return;
    const builds = visB(), curves = d3.group(rowsOf('mana'), (r) => r.build);
    const sum = new Map(rowsOf('mana_summary').map((r) => [r.build, r]));
    const manaBuilds = builds.filter((b) => b.class !== 'Warrior' && b.class !== 'Rogue');
    const shown = manaBuilds.filter((b) => sum.has(b.id) && (st.cls === 'All' || b.class === st.cls));
    const missing = manaBuilds.filter((b) => !sum.has(b.id) && (st.cls === 'All' || b.class === st.cls));
    $('#mana-sum').textContent = st.cls === 'Warrior' || st.cls === 'Rogue' ? `${st.cls}s don't use mana.`
      : missing.length ? `No mana data for ${[...new Set(missing.map((b) => b.class))].join(', ').replace(/, ([^,]*)$/, ' and $1')} (${missing.length} preset${missing.length > 1 ? 's' : ''}).` : '';
    if (!shown.length) { host.replaceChildren(); list.replaceChildren(); return; }
    const lowPct = rowsOf('mana_summary')[0]?.lowPct ?? 10;
    const series = manaBuilds.filter((b) => sum.has(b.id)).map((b) => {
      const on = shown.includes(b);
      const pts = (curves.get(b.id) || []).map((r) => ({ s: r.t, v: r.pct, p25: r.p25, p75: r.p75 }));
      return { id: b.id, name: b.label, color: ccol(b.class), op: on ? 0.9 : 0.06, pick: on, w: on && st.cls !== 'All' ? 2 : 1.5, pts };
    }).sort((a, b) => a.op - b.op);
    const hoverRow = (id) => list.querySelectorAll('.fv-row').forEach((r) => r.classList.toggle('is-hover', r.dataset.id === id));
    const chart = drawLenChart(host, series, {
      onHover: hoverRow,
      fmt: (v) => d3.format('.0f')(v) + '%', yDomain: [0, 100], ref: lowPct, refLabel: `${lowPct}%`, xTicks: [0, 60, 120, 180, 240, 300, 360, 420, 480], xFmt: (d) => mmss(d),
      strip: false, focus: true, open: true, hlId: st.hlMana, onPick: (id) => { st.hlMana = id; persist(); },
      tipSub: (t) => `at ${mmss(t)}`, tipMain: (p) => [`${d3.format('.0f')(p.v)}%`, 'mana, average'], tipRows: (p) => [['Middle half', `${d3.format('.0f')(p.p25)}–${d3.format('.0f')(p.p75)}%`]],
    });
    const rows = shown.map((b) => ({ b, m: sum.get(b.id) })).sort((p, q) => (p.m.lowAt ?? 1e9) - (q.m.lowAt ?? 1e9) || q.m.pctAt480 - p.m.pctAt480);
    const dur = d3.max(rows, (r) => r.m.dur) || 480;
    const x = (v) => `${(100 * Math.min(v, dur)) / dur}%`;
    list.replaceChildren(...rows.map(({ b, m }) => {
      const low = m.lowAt != null;
      const track = el('div', { class: 'fv-track' });
      if (low) track.append(el('span', { class: 'fv-box', style: `left:${x(m.lowAtP25 ?? m.lowAt)};width:calc(${x(m.lowAtP75 ?? dur)} - ${x(m.lowAtP25 ?? m.lowAt)})` }), el('span', { class: 'fv-med', style: `left:${x(m.lowAt)}` }));
      else track.append(el('span', { class: 'fv-bar', style: `width:100%;background:${ccol(b.class)};opacity:0.35` }));
      const row = el('button', { class: 'fv-row', 'data-tip': '1', 'data-id': b.id, 'aria-label': `${b.label}, ${low ? `below ${lowPct}% mana at ${mmss(m.lowAt)}` : `never below ${lowPct}% mana in a typical fight`}` },
        specIco(b) || el('span', { class: 'fv-dot', style: `background:${ccol(b.class)}` }),
        el('span', { class: 'fv-row-main' }, el('span', { class: 'fv-row-name' }, el('b', { text: b.build }), el('span', { text: '· ' + b.class }))),
        el('span', { class: 'fv-row-end' }, el('span', { class: 'fv-row-val', text: low ? mmss(m.lowAt) : '8:00+' }), el('span', { class: 'fv-row-gap', text: `${d3.format('.0f')(m.pctAt480)}% left at 8:00` })),
        track);
      // Pointing at a row shows its numbers, and a click keeps them up. A second click opens the preset in Specs.
      const tip = (pinned) => {
        const card = $('#mana-card'), cr = card.getBoundingClientRect(), rr = row.getBoundingClientRect();
        // Only what the row doesn't show already: the middle half of fights, how many fights run low at all, and the
        // mana pool and its net drain.
        const share = m.shareLow ?? m.share10;
        showTip(card, rr.left - cr.left + 30, rr.bottom - cr.top + 6, tipTitle(b.label, null, specIco(b, '')), [
          tipKey(low ? mmss(m.lowAt) : '8:00+', null, low ? `below ${lowPct}% mana` : `never below ${lowPct}%`),
          tipGrid([low && m.lowAtP25 != null ? ['Middle half', `${mmss(m.lowAtP25)}–${m.lowAtP75 != null ? mmss(m.lowAtP75) : 'never'}`] : null,
            ['Fights running low', share >= 99.5 ? 'all' : `${d3.format('.0f')(share)}%`], ['Mana pool', d3.format(',')(m.maxMana)], ['Net drain', `${f1(m.spentPerSec - m.gainedPerSec)}/s`]]),
        ], { hl: row, open: b.id, below: true, pinned });
      };
      row.addEventListener('pointerenter', (e) => {
        chart.highlight(b.id); row.classList.add('is-hover');
        if (e.pointerType !== 'mouse' || tipPinned) return;
        if (tipHl.includes(row)) { clearTimeout(tipWait); return; }
        tipHover(() => tip(false));
      });
      row.addEventListener('pointerleave', (e) => { chart.highlight(null); row.classList.remove('is-hover'); if (e.pointerType === 'mouse' && !tipPinned && tipHl.includes(row)) tipHover(hideTip); });
      row.addEventListener('click', (e) => {
        e.stopPropagation();
        if (row.classList.contains('is-hl') && tipPinned) { openSpec(b.id); return; }
        tip(true);
      });
      return row;
    }));
  };

  // ---------- races ----------
  const drawRaceSpread = () => {
    const host = $('#racespread');
    if (!waitOr(host, ['builds', 'race_variants', 'race_breakdown'])) return;
    const V = rsVar('spread');
    const builds = new Map(rowsOf('builds').map((b) => [b.id, b]));
    seg($('#spread-sort'), 'spreadSort', [['spread', 'Biggest gap'], ['class', 'By class']]);
    const order = new Map(buildOrder(rowsOf('builds')).map((b, i) => [b.id, i]));
    const races = raceRows(V);
    // The gap between each preset's best and worst entry. The rows are sorted best first.
    const rows = d3.groups(races.filter((r) => builds.has(r.build)), (r) => r.build).map(([build, rs]) => ({ build, spreadPct: Math.round((1 - rs[rs.length - 1].dps / rs[0].dps) * 1000) / 10, best: rs[0].name, worst: rs[rs.length - 1].name }))
      .sort((a, b) => (st.spreadSort === 'class' ? order.get(a.build) - order.get(b.build) : b.spreadPct - a.spreadPct));
    const max = d3.max(rows, (r) => r.spreadPct) || 1;
    const wrap = el('div', { class: 'fv-sbar fv-srows' });
    for (const r of rows) {
      const b = builds.get(r.build);
      const v = el('span', { class: 'vl' }, el('span', { text: f2(r.spreadPct) }), '%');
      const tr = el('span', { class: 'tr' }, el('span', { style: `width:${100 * r.spreadPct / max}%;background:${ccol(b.class)}` }));
      const row = el('div', { class: 'fv-srow', 'data-tip': '1' }, el('span', { class: 'nm' }, specIco(b, '') || el('span', { class: 'fv-dot', style: `background:${ccol(b.class)}` }), `${b.label} `,
        el('small', { class: 'fv-note', text: `${r.best} over ${r.worst}` })), v, tr);
      // The tooltip shows on hover and lists every entry on the preset, best first. A click pins it, so we can
      // reach its buttons.
      const tip = (pinned) => {
        const sib = races.filter((q) => q.build === b.id).sort((p, q) => q.dps - p.dps);
        const card = $('#racespread-card'), cr = card.getBoundingClientRect(), nr = tr.getBoundingClientRect();
        showTip(card, 20, nr.bottom - cr.top + 4, tipTitle(b.label, null, specIco(b, '')), [tipKey(`${f2(r.spreadPct)}%`, null, `best to worst, ${sib.length} entries`),
          tipGrid(sib.map((q, k) => [q.name + (q.isPreset ? ' (preset race)' : ''), k === 0 ? `${f1(q.dps)} DPS` : `−${f2(100 * (1 - q.dps / sib[0].dps))}%`]))],
          { hl: row, open: b.id, below: true, pinned, act: ['Compare races ›', () => pickRaceSpec(b.id)] });
      };
      hoverTip(row, tip);
      row.addEventListener('click', (e) => {
        e.stopPropagation();
        if (e.pointerType && e.pointerType !== 'mouse') return;
        if (row.classList.contains('is-hl') && tipPinned) { hideTip(); return; }
        tip(true);
      });
      wrap.append(row);
    }
    host.replaceChildren(wrap);
    host.style.minHeight = 'auto';
  };

  // ---------- races: scenarios ----------
  // Each Races card shows every race as the preset plays it, on the generic boss. A switch on each card adds the
  // boss types a racial works on as entries of their own (t, off by default). Since client 70425 no racial depends
  // on the weapon, so there are no racial weapon entries any more.
  const RS_OPTS = [['t', 'Racial enemy types', 'Troll and Dwarf vs Beast, Windshaper and High Order vs Elemental, Dwarf vs Humanoid (Priests)']];
  const rsVar = (card) => {
    const m = (st.rsOpt?.[card] ?? '').replace('w', '');
    const host = $(`#${card}-opts`);
    if (host && !host.dataset.bound) {
      host.dataset.bound = '1';
      host.replaceChildren(...RS_OPTS.map(([k, lbl, tip]) => {
        const bt = el('button', { type: 'button', class: 'fv-chip', 'data-k': k, title: tip, text: lbl });
        bt.addEventListener('click', (e) => {
          e.stopPropagation();
          const cur = (st.rsOpt?.[card] ?? '').replace('w', '');
          st.rsOpt = { ...st.rsOpt, [card]: RS_OPTS.map((o) => o[0]).filter((x) => (x === k) !== cur.includes(x)).join('') };
          persist(); hideTip(); render();
        });
        return bt;
      }));
    }
    host?.querySelectorAll('button').forEach((b) => b.setAttribute('aria-pressed', String(m.includes(b.dataset.k))));
    return m;
  };

  // A race with an advantage shows up a second time, as its own entry: Troll vs Beast, Windshaper vs Elemental. A
  // variant counts where its DPS differs from the race's own on the generic boss by more than the sim's noise (three
  // standard errors). High Order vs Beast comes out 0.1 DPS apart, for example.
  //
  // Each entry is a race_variants row with a key (the race, or race|variant), a name, a tag like "vs Beast", and its
  // rank and gap among all of the preset's entries.
  const ADV = ['beast', 'elemental', 'humanoid'];
  const weaponOf = (build, race) => (rowsOf('race_breakdown') || []).find((q) => q.build === build && q.race === race && q.scenario === 'weapon' && q.part !== 'swap')?.part;
  const advTag = (v, weapon) => (v === 'weapon' ? (weapon ? `with ${/^[aeiou]/.test(weapon) ? 'an' : 'a'} ${weapon}` : 'with its racial weapon') : `vs ${v[0].toUpperCase()}${v.slice(1)}`);
  // One tag for an entry across presets: Human swaps to a sword for some presets and to a mace for others.
  const groupTag = (rows) => {
    const v = rows[0].variant;
    if (v !== 'weapon') return advTag(v);
    const ws = [...new Set(rows.map((r) => r.weapon).filter(Boolean))];
    return ws.length ? advTag('weapon', ws.join(' or ')) : advTag('weapon');
  };
  let rrCache = {};
  const raceRows = (mode) => {
    const rv = rowsOf('race_variants');
    if (!rv) return null;
    const rb = rowsOf('race_breakdown') || [];
    if (rrCache.rv === rv && rrCache.rb === rb && rrCache.mode === mode) return rrCache.rows;
    const tgt = mode.includes('t');
    const rows = rv.filter((r) => r.variant === 'baseline').map((r) => ({ ...r, key: r.race, name: r.race, tag: null }));
    const base = new Map(rows.map((r) => [r.build + '|' + r.race, r]));
    for (const r of rv) {
      if (!ADV.includes(r.variant) || !r.simmed || !tgt) continue;
      const own = base.get(r.build + '|' + r.race);
      if (!own || Math.abs(r.dps - (own.dropped ? own.dps + own.dropped.dps : own.dps)) < 3 * Math.hypot(r.se, own.se)) continue;
      const weapon = r.variant === 'weapon' ? weaponOf(r.build, r.race) : null;
      const tag = advTag(r.variant, weapon);
      rows.push({ ...r, key: `${r.race}|${r.variant}`, name: `${r.race} ${tag}`, tag, weapon, adv: true, isPreset: false });
    }
    // Each preset's entries come out together, best first.
    const out = d3.groups(rows, (r) => r.build).flatMap(([, rs]) => {
      rs.sort((p, q) => q.dps - p.dps);
      rs.forEach((r, i) => { r.rank = i + 1; r.of = rs.length; r.vsBest = Math.round(10000 * (r.dps / rs[0].dps - 1)) / 100; });
      return rs;
    });
    rrCache = { rv, rb, mode, rows: out };
    return out;
  };
  // A race's name with its advantage, if any, as a smaller tag after it.
  // An advantage tag is a colored pill with an icon, so the extra entries stand out from the races' own. Weapon
  // racials are gray, the boss types orange, and the entry without its weapon racial is a dashed outline.
  const TAG_ICON = { sword: 'inv_sword_04', mace: 'inv_mace_25', axe: 'inv_axe_09', beast: 'ability_hunter_pet_bear', elemental: 'spell_frost_summonwaterelemental', humanoid: 'spell_holy_chastise' };
  const tagKind = (v) => (v === 'weapon' ? 'wpn' : v === 'noweapon' ? 'nowpn' : ADV.includes(v) ? 'tgt' : '');
  const tagIcon = (r) => TAG_ICON[r.variant === 'weapon' ? String(r.weapon || '').split(' ')[0] : r.variant];
  const raceTag = (r, tag = r.tag) => {
    if (!tag) return null;
    const ic = r.variant !== 'noweapon' && tagIcon(r) && iconMap() ? iconImg(tagIcon(r), '') : null;
    return el('span', { class: `fv-rtag ${tagKind(r.variant)}` }, ic, tag);
  };
  const raceName = (r, tag = r.tag) => el('span', { class: 'fv-rname' }, el('b', { text: r.race, style: `color:${fcol(r.race)}` }), raceTag(r, tag));
  const safeTop = (fn, host) => { try { fn(); } catch (err) { console.error(err); if (host) fail(host); } };
  // a racial's colour follows what kind it is: a passive, a cooldown, or a boss type; base stats and the weapon swap are neutral
  const PART_KIND = { elune: 1, eureka: 1, bloodfury: 1, berserking: 1, chastise: 1, starshards: 1, darksacrifice: 1, target: 2, beast_troll: 2, beast_dwarf: 2, insight: 2 };
  const partCol = (p) => (p === 'base' ? 'var(--cds-chart-muted)' : p === 'swap' ? 'var(--color-fg-muted)' : dash.colors[PART_KIND[p] ?? 0]);
  const COND = { beast: /^beast/, elemental: /^insight$/ };
  const KIND_NAME = ['Passive racial', 'Racial cooldown', 'Boss type racial'];
  const noteRow = (kind, key) => (rowsOf('race_notes') || []).find((n) => n.kind === kind && n.key === key);
  const shortLabel = (s) => String(s).replace('Specialization', 'Spec.');
  // A racial's game icon (race_notes) with a ring in its kind's colour, so it still ties to its part of the bar. The
  // weapon swap and a boss type with no racial behind it keep a plain swatch.
  const partIco = (p, cls = 'fv-pico') => {
    const ic = noteRow('part', p)?.icon;
    const img = ic && iconMap() ? iconImg(ic, cls) : null;
    if (!img) return el('i', { style: `background:${partCol(p)}` });
    img.style.setProperty('--ring', partCol(p));
    return img;
  };
  const pct = (v) => `${v > 99.995 ? '100' : f1(Math.min(v, 99.9))}%`;
  const ordinal = (n) => n + (n % 100 >= 11 && n % 100 <= 13 ? 'th' : ['th', 'st', 'nd', 'rd'][n % 10] || 'th');
  // Shows one preset in the Best races card: the class picked, the spec picked, and the card in view.
  const pickRaceSpec = (id) => {
    const b = (rowsOf('builds') || []).find((x) => x.id === id);
    Object.assign(st, { rsRole: 'All', rsCls: b ? b.class : 'All', rsSpec: id });
    persist(); hideTip();
    if (st.tab !== 'races') showTab('races'); else render();
    requestAnimationFrame(() => $('#race-rank-card').scrollIntoView({ behavior: matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth', block: 'start' }));
  };
  // Pointing at a row shows its tooltip, and on a phone (no pointer) a tap does. Without tap, a tap only does what the
  // row does (like a button that picks a setting), and the tooltip is for the mouse only.
  const hoverTip = (row, tip, { tap = true } = {}) => {
    row.addEventListener('pointerenter', (e) => {
      if (e.pointerType !== 'mouse' || tipPinned) return;
      if (tipHl.includes(row)) { clearTimeout(tipWait); return; }
      tipHover(() => tip(false));
    });
    row.addEventListener('pointerleave', (e) => { if (e.pointerType === 'mouse' && !tipPinned && tipHl.includes(row)) tipHover(hideTip); });
    row.addEventListener('pointerup', (e) => {
      if (!tap || e.pointerType === 'mouse' || e.target.closest('[data-tt]')) return;
      if (row.classList.contains('is-hl') && tipPinned) { hideTip(); return; }
      tip(true);
    });
  };

  // A row whose parts each have their own tooltip. A part is any element with data-zone, and the parts marked
  // data-zone="" all show the row's own tooltip (zone ''). The rest of the row shows none. With a mouse we show the
  // tooltip of the part under the pointer, and switch it as the pointer moves to another part. A tap shows and pins
  // the tooltip of the part it lands on, and a second tap on the same part closes it. tip(zone, node, pinned) gets
  // the part's element.
  const hoverZones = (row, tip) => {
    let cur = null, curN = null;
    const at = (e) => e.target.closest('[data-zone]');
    const zone = (n) => (n && row.contains(n) ? n.dataset.zone : null);
    const off = () => { cur = null; if (tipHl.includes(row)) hideTip(); };
    row.addEventListener('pointerover', (e) => {
      if (e.pointerType !== 'mouse' || tipPinned) return;
      const n = at(e), z = zone(n);
      if (z == null) return off();
      if (z === cur && n === curN && tipHl.includes(row)) return;
      cur = z; curN = n;
      tip(z, n, false);
    });
    row.addEventListener('pointerleave', (e) => { cur = null; if (e.pointerType === 'mouse' && !tipPinned && tipHl.includes(row)) hideTip(); });
    row.addEventListener('pointerup', (e) => {
      if (e.pointerType === 'mouse' || e.target.closest('[data-tt]')) return;
      const n = at(e), z = zone(n);
      if (z == null || (tipPinned && tipHl.includes(row) && z === cur)) return off();
      cur = z; curN = n;
      tip(z, n, true);
    });
  };

  // ---------- races: the scope filter ----------
  // Both Races cards narrow down the presets they average over with one filter: a faction, a role, a class, then
  // one spec. Each step lists only what fits the steps before it, so a role leaves its classes and a class leaves
  // its specs. Each card keeps its own choice, under its own prefix (rs for races, rl for racials).
  const ROLES = [['All', 'All roles'], ['Melee', 'Melee'], ['Ranged', 'Ranged'], ['Caster', 'Caster']];
  const scopeOf = (p) => {
    const roleMap = roleOf(), has = new Set((rowsOf('race_variants') || []).map((r) => r.build));
    const all = buildOrder(genB() || []).filter((b) => has.has(b.id));
    const role = st[p + 'Role'] || 'All';
    const inRole = all.filter((b) => role === 'All' || roleMap.get(b.id) === role);
    const classes = CLASSES.filter((c) => inRole.some((b) => b.class === c));
    if (!classes.includes(st[p + 'Cls'])) st[p + 'Cls'] = 'All';
    const cls = st[p + 'Cls'];
    const inCls = inRole.filter((b) => cls === 'All' || b.class === cls);
    // We pick a spec only within a class, so with all classes the spec goes back to all of them.
    if (cls === 'All' || !inCls.some((b) => b.id === st[p + 'Spec'])) st[p + 'Spec'] = 'All';
    const spec = st[p + 'Spec'];
    const builds = spec === 'All' ? inCls : inCls.filter((b) => b.id === spec);
    return { fac: st[p + 'Fac'] || 'All', role, cls, spec, classes, inCls, builds, ids: new Set(builds.map((b) => b.id)), single: builds.length === 1 ? builds[0] : null };
  };
  const drawScope = (p, sc) => {
    const fi = (n) => (iconMap() ? iconImg(n, 'fv-ico') : null);
    seg($(`#${p}-fac`), p + 'Fac', [['All', 'Both factions'], ['Horde', 'Horde', fi('ui_hordeicon')], ['Alliance', 'Alliance', fi('ui_allianceicon')]]);
    seg($(`#${p}-role`), p + 'Role', ROLES);
    classChips($(`#${p}-cls`), p + 'Cls', new Set(sc.classes));
    const all = `All ${sc.cls} specs`;
    const items = [{ value: 'All', text: `${all} (${sc.inCls.length})` }];
    for (const b of sc.inCls) items.push({ value: b.id, text: b.build, icon: () => specIco(b, '') });
    // The spec selector shows only once a class is picked, with that class's specs.
    const host = $(`#${p}-spec`);
    host.parentElement.hidden = sc.cls === 'All';
    host.dataset.label = 'Spec';
    iconSelect(host, items, sc.spec, (v) => { st[p + 'Spec'] = v; persist(); hideTip(); render(); }, (it) => (it.value === 'All' ? all : it.text));
  };

  // The Specs page's race cards show one preset, so their filter is only the faction.
  const fixedScope = (p, b) => {
    const fi = (n) => (iconMap() ? iconImg(n, 'fv-ico') : null);
    seg($(`#${p}-fac`), p + 'Fac', [['All', 'Both factions'], ['Horde', 'Horde', fi('ui_hordeicon')], ['Alliance', 'Alliance', fi('ui_allianceicon')]]);
    return { fac: st[p + 'Fac'] || 'All', builds: [b], ids: new Set([b.id]), single: b };
  };

  // An entry's bar on one preset: base stats, then each racial that works in its scenario (its variant). The entry
  // without its weapon racial leaves that racial out. A boss type's part is named after the racial it switches on.
  const segsOf = (row, mine) => {
    const V = row.variant;
    const out = [{ part: 'base', r: mine.find((q) => q.part === 'base') }];
    for (const r of mine.filter((q) => q.scenario === 'all' && q.part !== 'base' && q.active && q !== row.dropped)) out.push({ part: r.part, r });
    const cond = mine.find((q) => q.scenario === 'all' && !q.active && COND[V] && COND[V].test(q.part));
    for (const r of mine.filter((q) => q.scenario === V)) out.push(r.part === 'target' && cond ? { part: cond.part, r: { ...r, label: cond.label } } : { part: r.part, r });
    const idle = (row.dropped ? [{ ...row.dropped, note: 'left out of this entry' }] : []).concat(mine.filter((q) => q.scenario === 'all' && !q.active && !(COND[V] && COND[V].test(q.part))
      && !(V === 'weapon' && mine.some((w) => w.scenario === 'weapon' && w.part === q.part))));
    return { segs: out.filter((sg) => sg.r), idle };
  };
  const legend = (host, kinds, base) => host.replaceChildren(base ? el('span', {}, el('i', { class: 'sw2', style: 'background:var(--cds-chart-muted)' }), "Baseline (each race's stats, no racials)") : '', ...[0, 1, 2, 'swap'].filter((k) => kinds.has(k)).map((k) => el('span', {}, el('i', { class: 'sw2', style: `background:${k === 'swap' ? 'var(--color-fg-muted)' : dash.colors[k]}` }), k === 'swap' ? 'Weapon swap' : KIND_NAME[k])));

  // Each racial's caption sits under its piece of the bar, starting where the piece starts. When a caption would run
  // into the one before it, we move it down to the first line where it fits, so short pieces side by side still get
  // readable captions. A caption that would run past the right edge of the card ends at the right edge instead.
  //
  // We place captions in percent, or against the right edge, not in pixels. So they stay under their piece when the
  // card changes width, or when the web font loads after we measured and the text gets wider.
  const layoutCaps = (root) => {
    for (const cap of root.querySelectorAll('.fv-rcap')) {
      const W = cap.clientWidth, spans = [...cap.children];
      if (!W || !spans.length) continue;
      const lineH = spans[0].offsetHeight + 2, lanes = [];
      for (const sp of spans) {
        const w = sp.offsetWidth, want = (W * +sp.dataset.at) / 100, clamp = want + w > W;
        const from = clamp ? Math.max(0, W - w) : want;
        let lane = lanes.findIndex((end) => end + 10 <= from);
        if (lane < 0) lane = lanes.push(0) - 1;
        lanes[lane] = from + w;
        Object.assign(sp.style, { left: clamp ? 'auto' : `${sp.dataset.at}%`, right: clamp ? '0' : 'auto', top: `${lane * lineH}px` });
      }
      cap.style.height = `${lanes.length * lineH}px`;
    }
  };

  // ---------- races: best races ----------
  // Every race, and each advantage as its own entry, ranked over the presets in the filter. On one preset a bar is
  // the entry's DPS. Over several presets it's the entry's DPS as a share of the best entry's on the same preset,
  // averaged, so a full bar is the best race on every preset. Either way the bar is split into the race's base
  // stats (the gray body) and what each of its racials adds, which shows how strong each racial is.
  //
  // The Specs page shows the same card for its preset (fixed), with only the faction and the enemy types switch.
  const drawRaceRank = (fixed) => {
    const p = fixed ? 'sr' : 'rs';
    const host = fixed ? $('#sp-races') : $('#race-rank'), axis = $(`#${p}-axis`), card = host.closest('.fv-card');
    const ids = ['race_variants', 'race_breakdown', 'builds', 'stats'];
    if (!(fixed ? specWait(host, ids, fixed) : waitOr(host, ids))) return;
    const sc = fixed ? fixedScope(p, fixed) : scopeOf(p);
    if (!fixed) drawScope(p, sc);
    const V = rsVar(p), one = sc.single;
    const rr = raceRows(V).filter((r) => sc.ids.has(r.build) && (sc.fac === 'All' || r.faction === sc.fac));
    if (!rr.length) {
      host.replaceChildren(el('p', { class: 'fv-note', text: fixed ? 'No race numbers.' : 'No race of this faction plays these presets.' }));
      axis.textContent = ''; $(`#${p}-legend`).replaceChildren();
      return;
    }
    const bd = d3.group(rowsOf('race_breakdown').filter((q) => sc.ids.has(q.build)), (q) => q.build + '|' + q.race);
    const bmap = new Map(sc.builds.map((b) => [b.id, b]));
    // Per entry: each part's value summed over the presets (DPS on one preset, else a share of the best entry's
    // DPS), its share of the best on each preset, and its place.
    const acc = new Map();
    for (const [build, rs] of d3.groups(rr, (r) => r.build)) {
      const best = rs[0].dps, u = one ? 1 : 100 / best;
      rs.forEach((r, i) => {
        const { segs, idle } = segsOf(r, bd.get(build + '|' + r.race) || []);
        if (!acc.has(r.key)) acc.set(r.key, { ...r, parts: new Map(), shares: [], ranks: [], idle });
        const a = acc.get(r.key);
        a.shares.push({ build, share: (100 * r.dps) / best });
        a.ranks.push(i + 1);
        for (const sg of segs) {
          if (!a.parts.has(sg.part)) a.parts.set(sg.part, { part: sg.part, label: sg.r.label, sum: 0 });
          a.parts.get(sg.part).sum += +sg.r.dps * u;
        }
      });
    }
    const rows = [...acc.values()].map((a) => {
      const n = a.shares.length;
      const avg = d3.mean(a.shares, (g) => g.share);
      return { ...a, n, value: one ? a.dps : avg, avg, timesBest: a.shares.filter((g) => g.share > 99.995).length, medianRank: d3.median(a.ranks),
        segs: [...a.parts.values()].map((p) => ({ ...p, v: p.sum / n })) };
    }).sort((p, q) => q.value - p.value);
    for (const r of rows) { let at = 0; for (const sg of r.segs) { sg.from = at; at += sg.v; sg.to = at; } }
    const ends = rows.flatMap((r) => r.segs.flatMap((sg) => [sg.from || sg.to, sg.to]).filter((v) => v > 0).concat(one ? [r.dps - r.se, r.dps + r.se] : [r.value]));
    const hi = d3.max(ends), lo0 = Math.min(d3.min(ends), d3.min(rows, (r) => r.segs[0].to));
    const lo = Math.max(0, lo0 - 0.35 * Math.max(hi - lo0, one ? 4 : 0.5));
    const x = (v) => (100 * (Math.max(lo, v) - lo)) / (hi - lo);
    const fv = (v) => (one ? f1(v) : `${f2(v)}%`), fd = (v) => (one ? sgn(v, f1) : `${sgn(v, f2)}%`);
    const who = (r) => (r.tag ? `${r.race} ${r.tag}` : r.race);
    const kinds = new Set();
    // The weakest base stats in view, which the tooltip measures each race's base stats against.
    const baseOf = (r) => r.segs.find((sg) => sg.part === 'base')?.v;
    const low = d3.least(rows.filter((r) => baseOf(r) != null), baseOf);
    // The weapon swap's tooltip, told for this entry. It starts with what the preset holds and why the entry holds
    // another weapon type, then what that changes before the racial: each talent whose bonus depends on the weapon
    // type, and the weapon's speed. Over several presets it lists what the swap changes on each preset instead.
    const art = (t) => (/^[aeiou]/.test(t) ? 'an ' : 'a ') + t;
    const swapTip = (r) => {
      const sw = r.shares.map((g) => (bd.get(g.build + '|' + r.race) || []).find((q) => q.part === 'swap')).filter(Boolean);
      const racial = r.segs.find((sg) => sg.part === r.weapon)?.label || 'Its weapon racial';
      const W = r.weapon || 'weapon';
      const talent = (t, lines) => el('div', { class: 'fv-tip-racial' }, iconImg(t[1], 'fv-ico') || el('i'),
        el('span', {}, el('b', { text: lines[0] }), ...lines.slice(1).map((l, k) => el('span', { class: k === 0 ? GAME_TEXT : null, text: l }))));
      const bonus = (type, v) => `With ${art(type)}: ${v || 'no bonus'}.`;
      if (one) {
        const q = sw[0];
        if (!q) return { text: null, nodes: [] };
        const ws = arr(q.weapons), spec = bmap.get(one.id)?.build || one.label;
        const holds = ws.length === 2 ? (ws[0][1] === ws[1][1] ? `${art(ws[0][1])} in each hand` : `${art(ws[0][1])} and ${art(ws[1][1])}`)
          : ws[0][0] === 'off hand' ? `${art(ws[0][1])} in the off hand` : ws[0][0] === 'two-hand' ? `a two-hand ${ws[0][1]}` : art(ws[0][1]);
        const gets = ws.length === 2 ? `${art(W)} in each hand` : ws[0][0] === 'off hand' ? `${art(W)} in the off hand` : q.caster ? `a one-hand ${W}` : art(W);
        const text = `${spec} holds ${holds}. ${racial} needs ${art(W)}, so this entry gets ${gets} with the same stats. `
          + `This piece is the swap alone, before the racial bonus.`;
        const moved = ws.filter((w) => w[3] !== w[4]);
        const nodes = [];
        if (q.caster) nodes.push(el('span', { class: 'fv-tip-m', text: "Casters don't attack with their weapon. No change." }));
        else {
          for (const t of arr(q.talents)) nodes.push(talent(t, [`${t[0]} ${t[2]}/${t[3]}`, t[4], bonus(t[5], t[6]), bonus(t[7], t[8])]));
          if (moved.length) {
            nodes.push(el('span', {}, el('b', { text: 'Speed: ' }), moved.map(([hand, , , s0, s1]) => `${hand} ${f2(s0)}s to ${f2(s1)}s`).join(', ') + '. ',
              Math.abs(q.dps) < 0.05 ? "Not a melee preset. No change." : 'Same weapon DPS, but a slower weapon hits harder and less often.'));
          }
          if (!arr(q.talents).length && !moved.length) nodes.push(el('span', { class: 'fv-tip-m', text: 'No talent uses the weapon type and the speed is the same. No change.' }));
        }
        return { text, nodes };
      }
      const text = `${racial} needs ${art(W)}. Presets with another weapon get ${art(W)} with the same stats. `
        + `This piece is the swap alone, before the racial bonus.`;
      const why = (q) => (arr(q.talents).length ? arr(q.talents).map((t) => t[0]).join(' and ') + ' bonus changes' : 'the weapon speed changes');
      const from = (q) => { const t = [...new Set(arr(q.weapons).map((w) => w[1]))].join(' and '); return t[0].toUpperCase() + t.slice(1); };
      const live = sw.filter((q) => Math.abs(+q.dps) >= 0.05), idle = sw.filter((q) => Math.abs(+q.dps) < 0.05);
      const talents = new Map();
      for (const q of live) for (const t of arr(q.talents)) if (!talents.has(t[0])) talents.set(t[0], t);
      return { text, nodes: [
        el('div', { class: 'fv-tip-specs' }, ...live.map((q) => {
          const b = bmap.get(q.build);
          return el('span', {}, b ? specIco(b, '') : el('i'), el('span', {}, el('b', { text: `${b?.label || q.build}: ${sgn(+q.dps, f1)} DPS. ` }), `${from(q)} to ${W}, ${why(q)}.`));
        })),
        idle.length ? el('span', { class: 'fv-tip-m', text: `No effect on ${idle.map((q) => bmap.get(q.build)?.label || q.build).join(', ')}. `
          + 'Casters and ranged Hunters don\'t attack with it. The others keep the same speed and talent bonus.' }) : null,
        ...[...talents.values()].map((t) => talent(t, [t[0], t[4]])),
      ] };
    };
    // Every race of a class starts from the same class stats plus its own small stat offsets, so the offsets are the
    // whole base stats gap between two races. For one preset we also price each stat with that preset's stat weights,
    // so the sum is close to the simmed gap.
    const wOne = one ? new Map((rowsOf('weights') || []).filter((w) => w.build === one.id).map((w) => [w.stat, +w.dpsPer])) : null;
    const statGap = (race, ref) => {
      const a = noteRow('stats', race)?.stats, b = noteRow('stats', ref)?.stats;
      if (!a || !b) return [];
      const gaps = Object.keys(a).map((k) => [k, a[k] - b[k]]).filter(([, d]) => d);
      if (!gaps.length) return [el('span', { class: 'fv-tip-m', text: `Same stats as ${ref}.` })];
      const w = wOne?.size ? wOne : null;
      const worth = (k, d) => (w.has(k) ? d * w.get(k) : 0);
      const sum = w ? d3.sum(gaps, ([k, d]) => worth(k, d)) : 0;
      return [tipTable([`Stats vs ${ref}`, '', ...(w ? ['DPS'] : [])], [
        ...gaps.map(([k, d]) => [k, sgn(d, f0), ...(w ? [w.has(k) && w.get(k) ? tipDelta(worth(k, d), f1, '') : '0'] : [])]),
        w ? ['Sum', '', tipDelta(sum, f1, '')] : null])];
    };
    // The race's numbers show the row's own tooltip.
    const head = (n) => { n.dataset.zone = ''; return n; };
    // The race's icon and name show the race's game tooltip instead, with every racial it has, also the ones that
    // add no DPS here. On one preset it also gives that preset's DPS for the race.
    const cls1 = new Set(sc.builds.map((b) => b.class)).size === 1 ? sc.builds[0].class : null;
    const raceTT = (r, n) => tt(n, `r:${r.race}${cls1 ? ':' + cls1 : ''}`, true, one ? one.id : null);
    // An entry's "vs Elemental" tag shows the tooltip of the racial that works on that boss type.
    const tagZone = (r, name) => {
      const tag = name.querySelector('.fv-rtag');
      const part = r.segs.find((sg) => COND[r.variant]?.test(sg.part))?.part ?? r.segs.find((sg) => sg.part === 'target')?.part;
      if (tag && part) tag.dataset.zone = part;
      return name;
    };
    const raceNameTT = (r) => { const n = raceName(r); raceTT(r, n.querySelector('b')); return n; };
    host.replaceChildren(...rows.map((r, i) => {
      const track = el('div', { class: 'fv-rtrack' });
      r.segs.forEach((sg, k) => {
        const a = Math.min(sg.from, sg.to), z = Math.max(sg.from, sg.to);
        if (sg.part !== 'base') kinds.add(sg.part === 'swap' ? 'swap' : PART_KIND[sg.part] ?? 0);
        const neg = sg.v < 0 && sg.part !== 'base';
        track.append(el('span', { class: 'sg' + (neg ? ' neg' : '') + (sg.part === 'swap' ? ' swap' : '') + (k === r.segs.length - 1 ? ' last' : ''), 'data-zone': sg.part,
          style: `left:${k === 0 ? 0 : x(a)}%;width:${Math.max(k === 0 ? 0 : 0.6, x(z) - (k === 0 ? 0 : x(a)))}%;background:${partCol(sg.part)};${neg ? 'z-index:1;' : ''}` }));
      });
      if (one) track.append(el('span', { class: 'wk', style: `left:${x(r.dps - r.se)}%;width:${Math.max(0.4, x(r.dps + r.se) - x(r.dps - r.se))}%` }));
      const cap = el('div', { class: 'fv-rsub fv-rcap' }, ...r.segs.map((sg, k) => ({ sg, at: k === 0 ? 0 : x(Math.min(sg.from, sg.to)) }))
        .filter(({ sg }) => sg.part !== 'base' && Math.abs(sg.v) >= (one ? 0.05 : 0.005))
        .map(({ sg, at }) => el('span', { 'data-zone': sg.part, 'data-at': at, style: `left:${at}%` }, partIco(sg.part), `${shortLabel(sg.label)} ${fd(sg.v)}`)));
      // A racial with no effect here gets its reason after the race name, so it costs no extra line.
      const idle = one ? r.idle.slice(0, 3).map((q) => el('small', { 'data-zone': 'idle:' + q.part, text: `${shortLabel(q.label)}: ${q.note}` })) : [];
      const row = el('div', { class: 'fv-rrow' + (i === 0 ? ' top' : ''), 'data-tip': '1' },
        raceTT(r, raceIco(r.raceIcon) || el('span', { class: 'fv-dot', style: `background:${fcol(r.race)}` })),
        el('span', { class: 'nm' }, tagZone(r, raceNameTT(r)), one && r.isPreset ? head(el('i', { text: 'preset race' })) : null, ...idle),
        el('span', { class: 'end', 'data-zone': '' }, el('small', { text: one ? (i === 0 ? '★ best' : `−${f2(100 * (1 - r.dps / rows[0].dps))}%`) : r.timesBest ? `best on ${r.timesBest} of ${r.n}` : 'never best' }),
          el('b', { text: one ? f1(r.dps) : pct(r.avg) })),
        track, cap);
      // Each part of the entry has its own tooltip. A piece of the bar, or its caption under the bar, says what
      // that part is worth and what the racial does. The rest of the row gives the entry's DPS, and over several
      // presets the specs it's the best race for, one line per class, or the three where it comes closest.
      hoverZones(row, (zone, node, pinned) => {
        const cr = card.getBoundingClientRect(), rw = row.getBoundingClientRect();
        const title = tipTitle(who(r), one ? one.label : null, raceIco(r.raceIcon));
        const opts = { below: true, pinned };
        if (zone) {
          const [what, part] = zone.startsWith('idle:') ? ['idle', zone.slice(5)] : ['seg', zone];
          const sg = r.segs.find((g) => g.part === part), q = r.idle.find((g) => g.part === part);
          // A boss type's piece with no racial of its own name (Dwarf vs Humanoid) is told as the racial behind it.
          const ip = (part === 'target' && TARGET_PART[`${r.race}|${r.variant}`]) || part;
          const info = noteRow('part', ip);
          let name, sub = [];
          if (what === 'idle' && q) {
            name = shortLabel(q.label);
            sub.push(el('span', { class: 'fv-tip-m', text: q.note[0].toUpperCase() + q.note.slice(1) + '.' }));
          } else if (part === 'base') {
            const base = baseOf(r);
            name = `Base stats ${fv(base)}${one ? ' DPS' : ''}`;
            if (low && rows.length > 1) {
              const gap = one ? (100 * (base - baseOf(low))) / baseOf(low) : base - baseOf(low);
              sub.push(el('span', { class: 'fv-tip-m', text: low.race === r.race ? 'Weakest base stats here.'
                : gap < 0.005 ? `Same as ${low.race} (weakest base here).`
                : `${f2(gap)}%${one ? ` (${f1(base - baseOf(low))} DPS)` : ''} above ${low.race} (weakest base here).` }));
              const ref = low.race === r.race ? d3.greatest(rows.filter((q) => baseOf(q) != null), baseOf) : low;
              if (ref.race !== r.race) sub.push(...statGap(r.race, ref.race));
            }
          } else if (sg) name = `${ip !== part && info ? info.title : sg.label} ${fd(sg.v)}`;
          if (!name) return;
          let text = info?.text;
          if (part === 'swap') { const t = swapTip(r); text = t.text || text; sub.push(...t.nodes); }
          if (!one && what === 'seg') sub.push(el('span', { class: 'fv-tip-m', text: `Share of the best race's DPS, averaged over ${r.n} presets.` }));
          const segEl = what === 'seg' ? row.querySelector(`.sg[data-zone="${part}"]`) : null;
          const capEl = row.querySelector(`.fv-rsub [data-zone="${zone}"]`);
          // We point the tooltip at the caption or the tag when the pointer is on it, else at the piece of the bar.
          const own = node.closest('.fv-rsub, .fv-rtag');
          const at = (own ? node : segEl || node).getBoundingClientRect();
          showTip(card, Math.max(8, at.left - cr.left), (own ? at.bottom : rw.bottom) - cr.top + 4, title,
            [el('div', { class: 'fv-tip-racial' }, part === 'base' ? el('i', { style: 'background:var(--cds-chart-muted)' }) : partIco(ip, 'fv-ico'),
              el('span', {}, el('b', { text: name }), text ? el('span', { class: text === info?.text && racialText(ip) ? GAME_TEXT : null, text }) : null)), ...sub],
            { ...opts, hl: [row, segEl, capEl] });
          return;
        }
        let head = [tipKey(f1(r.dps), 'DPS', `± ${f2(r.se)}`)];
        if (!one) {
          const order = new Map(sc.builds.map((b, k) => [b.id, k]));
          const best = r.shares.filter((g) => g.share > 99.995).sort((p, q) => order.get(p.build) - order.get(q.build));
          const pick = best.length ? best : [...r.shares].sort((p, q) => q.share - p.share).slice(0, 3);
          const list = d3.groups(pick.map((g) => ({ g, b: bmap.get(g.build) })).filter((o) => o.b), (o) => o.b.class).map(([cls, os]) => el('span', {},
            ringed('classicon_' + cls.toLowerCase(), 'fv-ico ring', ccol(cls)), el('span', {}, el('b', { text: `${cls}: `, style: `color:${ccol(cls)}` }),
              os.map((o) => o.b.build + (best.length ? '' : ` (${pct(o.g.share)})`)).join(', '))));
          head = [tipKey(best.length ? `Best race on ${best.length}` : 'Never the best race', `of ${r.n} presets`),
            best.length ? null : el('span', { class: 'fv-tip-m', text: 'Closest on:' }), el('div', { class: 'fv-tip-specs' }, ...list)];
        }
        // The tooltip goes under the part the pointer is on. Under the numbers on the right it ends where they end.
        const at = node.getBoundingClientRect(), end = node.classList.contains('end');
        showTip(card, end ? at.right - cr.left : Math.max(8, at.left - cr.left), at.bottom - cr.top + 4, title, [...head, null], { ...opts, end, hl: row });
      });
      return row;
    }));
    axisWarn(axis, lo > 0 ? (one ? `${f0(lo)} DPS` : `${f1(lo)}%`) : null, 'races');
    legend($(`#${p}-legend`), kinds, true);
    layoutCaps(host);
    if (!host.dataset.capWatch) {
      host.dataset.capWatch = '1';
      let w = 0;
      new ResizeObserver(() => { if (host.clientWidth !== w) { w = host.clientWidth; layoutCaps(host); } }).observe(host);
      document.fonts?.ready.then(() => layoutCaps(host));
      document.fonts?.addEventListener('loadingdone', () => layoutCaps(host));
    }
  };

  // ---------- races: strongest racials ----------
  // Each racial ranked by how much it adds to its race's DPS, averaged over the presets in the filter that its race
  // can play. Where a racial doesn't work (Beast Slaying on the generic boss) it
  // counts as nothing. The switches decide whether a weapon racial counts with its weapon swapped in, and a boss
  // type racial against its boss type. On one preset the numbers are DPS.
  const TARGET_PART = { 'Troll|beast': 'beast_troll', 'Dwarf|beast': 'beast_dwarf', 'High Order|elemental': 'insight', 'Windshaper|elemental': 'insight', 'Dwarf|humanoid': 'chastise' };
  //
  // The Specs page shows the same card for its preset (fixed), with only the faction and the enemy types switch.
  const drawRacialRank = (fixed) => {
    const p = fixed ? 'sl' : 'rl';
    const host = fixed ? $('#sp-racials') : $('#racial-rank'), dead = $(`#${p}-dead`), card = host.closest('.fv-card');
    const ids = ['race_variants', 'race_breakdown', 'builds', 'stats'];
    if (!(fixed ? specWait(host, ids, fixed) : waitOr(host, ids))) return;
    const sc = fixed ? fixedScope(p, fixed) : scopeOf(p);
    if (!fixed) drawScope(p, sc);
    const V = rsVar(p), one = sc.single, tgt = V.includes('t');
    const base = rowsOf('race_variants').filter((r) => r.variant === 'baseline' && sc.ids.has(r.build) && (sc.fac === 'All' || r.faction === sc.fac));
    const bd = d3.group(rowsOf('race_breakdown').filter((q) => sc.ids.has(q.build)), (q) => q.build + '|' + q.race);
    const bmap = new Map(sc.builds.map((b) => [b.id, b]));
    const acc = new Map();
    for (const r of base) {
      const mine = bd.get(r.build + '|' + r.race) || [];
      const tv = {};
      if (tgt) for (const q of mine) if (q.part === 'target' && TARGET_PART[r.race + '|' + q.scenario]) tv[TARGET_PART[r.race + '|' + q.scenario]] = q;
      for (const q of mine.filter((x) => x.scenario === 'all' && x.part !== 'base')) {
        let v = q.active ? +q.dps : 0, works = q.active;
        if (tv[q.part]) { v += +tv[q.part].dps; works = true; }
        const k = r.race + '|' + q.part;
        if (!acc.has(k)) acc.set(k, { key: k, race: r.race, part: q.part, label: q.label, raceIcon: r.raceIcon, n: 0, works: 0, share: 0, dps: 0, per: [], notes: [] });
        const a = acc.get(k);
        a.n++; a.share += (100 * v) / r.dps; a.dps += v;
        a.per.push({ build: r.build, share: (100 * v) / r.dps, v });
        if (works && Math.abs(v) >= 0.05) a.works++;
        else if (q.note && !a.notes.includes(q.note)) a.notes.push(q.note);
      }
    }
    const all = [...acc.values()].map((a) => ({ ...a, value: one ? a.dps : a.share / a.n })).sort((p, q) => q.value - p.value);
    const rows = all.filter((a) => a.works), idle = all.filter((a) => !a.works);
    if (!rows.length) {
      host.replaceChildren(el('p', { class: 'fv-note', text: 'No racial helps these presets.' }));
      dead.replaceChildren(); $(`#${p}-legend`).replaceChildren();
      return;
    }
    const max = Math.max(d3.max(rows, (a) => a.value), one ? 1 : 0.1);
    const fd = (v) => (one ? sgn(v, f1) : `${sgn(v, f2)}%`);
    const kinds = new Set();
    host.replaceChildren(...rows.map((a, i) => {
      kinds.add(PART_KIND[a.part] ?? 0);
      const row = el('div', { class: 'fv-orow' + (i === 0 ? ' top' : ''), 'data-tip': '1' },
        noteRow('part', a.part)?.icon && iconMap() ? iconImg(noteRow('part', a.part).icon, 'fv-ico') : raceIco(a.raceIcon) || el('span', { class: 'fv-dot', style: `background:${fcol(a.race)}` }),
        el('span', { class: 'nm' }, el('b', { text: a.label }), tt(el('small', { style: `color:${fcol(a.race)}` }, raceIco(a.raceIcon), a.race), `r:${a.race}`)),
        el('span', { class: 'end' }, el('b', { text: fd(a.value), class: a.value < 0 ? 'neg' : '' }),
          el('small', { text: one ? 'DPS' : a.works === a.n ? `works on all ${a.n}` : `works on ${a.works} of ${a.n}` })),
        el('div', { class: 'fv-otrack' }, el('span', { class: 'bar', style: `width:${(100 * Math.max(0, a.value)) / max}%;background:${partCol(a.part)}` })));
      hoverTip(row, (pinned) => {
        const cr = card.getBoundingClientRect(), rw = row.getBoundingClientRect();
        const info = noteRow('part', a.part);
        const top = [...a.per].sort((p, q) => q.share - p.share).filter((g) => Math.abs(g.v) >= 0.05).slice(0, 3);
        showTip(card, 24, rw.bottom - cr.top + 4, tipTitle(a.label, a.race, noteRow('part', a.part)?.icon ? iconImg(noteRow('part', a.part).icon) : raceIco(a.raceIcon)), [
          tipKey(fd(a.value), one ? 'DPS' : `of ${a.race} DPS`, one ? null : 'on average'),
          info ? tipText(racialText(a.part) ? el('span', { class: GAME_TEXT, text: info.short || info.text }) : info.short || info.text) : null,
          !one && top.length ? tipTable(['Strongest on', ''], top.map((g) => {
            const b = bmap.get(g.build);
            return [b ? el('span', { class: 'fv-tip-spec' }, specIco(b, ''), b.label) : g.build, `${sgn(g.share, f2)}%`];
          })) : null,
          a.works < a.n && a.notes.length ? tipNote(`Does nothing on ${a.n - a.works} of ${a.n}: ${a.notes.slice(0, 2).join(' · ')}`) : null,
        ], { hl: row, below: true, pinned });
      });
      return row;
    }));
    dead.textContent = idle.length ? `No effect on ${one ? 'this preset' : 'these presets'}: ${idle.map((a) => `${a.label} (${a.race})`).join(', ')}.` : '';
    legend($(`#${p}-legend`), kinds);
  };

  // ---------- icons ----------
  // The class, race, faction and preset icons come first. The spell, item and slot icons (icons_more) start loading once
  // the first view is up, or as soon as something asks for an icon that isn't in the first set.
  let iconCache = { src: null, more: null, m: null };
  let wantMore = false;
  const iconMap = () => {
    const r = rowsOf('icons');
    if (!r) return null;
    const more = wantMore ? rowsOf('icons_more') : null;
    if (iconCache.src !== r || iconCache.more !== more) iconCache = { src: r, more, m: new Map([...r, ...(more || [])].map((x) => [x.icon, x.url])) };
    return iconCache.m;
  };
  const iconImg = (name, cls = 'fv-ico') => {
    const m = iconMap();
    if (m && name && !m.has(name) && !wantMore) { wantMore = true; dash.data('icons_more'); }
    const u = m && (m.get(name) || m.get('inv_misc_questionmark'));
    return u ? el('img', { src: u, alt: '', class: cls }) : null;
  };
  const ringed = (name, cls, ring) => { const i = iconImg(name, cls); if (i && ring) i.style.setProperty('--ring', ring); return i; };
  const specIco = (b, size = 'md') => ringed(b.specIcon, `fv-ico ${size} ring`, ccol(b.class));
  const raceIco = (icon) => (icon ? iconImg(icon, 'fv-ico') : null);
  const svgIcon = (sel, name, x, y, sz) => { const m = iconMap(); const u = m && (m.get(name) || null); if (u) sel.append('image').attr('href', u).attr('x', x).attr('y', y).attr('width', sz).attr('height', sz); };
  const keyIcon = (key) => { const n = rowsOf('names'); if (!n) return null; const r = n.find((x) => x.key === key); return r?.icon ? iconImg(r.icon) : null; };

  // ---------- talents ----------
  // The preset the talents show, for redrawing them when the card's width changes.
  let talB = null;
  const drawTalents = (b) => {
    const host = $('#sp-tal');
    host.setAttribute('data-source', 'talenttrees builds icons treeart');
    if (!specWait(host, ['talenttrees'], b, [$('#sp-tal-seg')])) return;
    iconMap();
    const all = rowsOf('talenttrees').filter((t) => t.class === b.class);
    const parts = String(b.talents).split('-');
    const pts = (t) => +((parts[t.tree] || '')[t.idx] || 0);
    const trees = d3.groups(all, (t) => t.tree).sort((x, y) => x[0] - y[0]).map(([ti, ts]) => ({ ti, name: ts[0].treeName, ts, spent: d3.sum(ts, pts) }));
    // On a narrow screen the trees don't fit side by side, and the tree picker shows one at a time, the one with the
    // most points first. When all three fit, they all show and the picker hides.
    if (st.talBuild !== b.id || st.talTree == null) { st.talTree = String(trees.reduce((x, y) => (y.spent > x.spent ? y : x)).ti); st.talBuild = b.id; }
    $('#sp-tal-seg').replaceChildren(...[...trees.map((t) => [String(t.ti), `${t.name} ${t.spent}`]), ['all', 'All three']].map(([v, lbl]) => {
      const bt = el('button', { 'aria-pressed': String(st.talTree === v), text: lbl });
      bt.addEventListener('click', () => { st.talTree = v; drawTalents(b); });
      return bt;
    }));
    // We redraw when the card's width changes, in case the trees fit now or no longer do. The Specs tab is hidden
    // until we open it, and then its width changes from 0 too.
    talB = b;
    const card = $('#sp-tal-card');
    if (!card.dataset.fitWatch) {
      card.dataset.fitWatch = '1';
      let w = card.clientWidth;
      new ResizeObserver(() => { if (card.clientWidth !== w) { w = card.clientWidth; if (talB && st.tab === 'spec') { try { drawTalents(talB); } catch (e) { console.error(e); } } } }).observe(card);
    }
    const arrow = (pr, pc, r, c, active) => {
      // Placed like TalentReqArrow in the sim UI: it starts on the parent's cell and its length is in cells.
      const a = el('div', { class: 'talent-picker-req-arrow' });
      let rowEnd = Math.max(pr, r) + 1, colEnd = Math.max(pc, c) + 1, dir;
      if (pr === r) {
        dir = pc < c ? 'right' : 'left';
        a.dataset.reqArrowColSize = String(Math.abs(pc - c));
        colEnd = dir === 'left' ? colEnd + 1 : colEnd - 1;
      } else if (pc === c) {
        dir = 'down';
        a.dataset.reqArrowRowSize = String(Math.abs(pr - r));
        rowEnd += 1;
      } else {
        dir = pc < c ? 'rightdown' : 'leftdown';
        a.dataset.reqArrowColSize = String(Math.abs(pc - c));
        a.dataset.reqArrowRowSize = String(Math.abs(pr - r));
        rowEnd += 1;
        colEnd = dir === 'rightdown' ? colEnd + 1 : colEnd - 1;
        a.append(el('div'));
      }
      a.classList.add(`talent-picker-req-arrow-${dir}`);
      Object.assign(a.style, { gridRow: String(pr + 1), gridColumn: String(pc + 1), gridRowEnd: String(rowEnd), gridColumnEnd: String(colEnd), zIndex: '1' });
      if (active) a.dataset.reqActive = 'true';
      return a;
    };
    const icons = iconMap();
    // The trees use the sim UI's talent picker markup and styles (talents_picker.tsx), read only.
    const paint = (shown) => host.replaceChildren(el('div', { class: 'talents-picker-root' }, el('div', { class: 'talents-picker-inner' }, el('div', { id: 'talents-carousel', class: 'carousel' },
      el('div', { class: 'carousel-inner' }, ...shown.map((tr) => {
        const rows = d3.max(tr.ts, (t) => t.row) + 1;
        const cell = 'calc(var(--talent-icon-size) * 1.4)';
        const main = el('div', { class: 'talent-tree-main', style: `grid-template-rows:repeat(${rows},1fr);grid-template-columns:repeat(4,1fr);height:calc(${cell} * ${rows});max-width:calc(${cell} * 4)` });
        for (const t of tr.ts) if (t.preRow != null) main.append(arrow(t.preRow, t.preCol, t.row, t.col, pts(t) > 0));
        for (const t of tr.ts) {
          const p = pts(t);
          const u = icons && (icons.get(t.icon) || icons.get('inv_misc_questionmark'));
          main.append(tt(el('a', {
            class: `talent-picker-root${p >= t.maxPoints ? ' talent-full' : ''}`, 'data-points': String(p), 'data-max-points': String(t.maxPoints),
            style: `grid-row:${t.row + 1};grid-column:${t.col + 1};z-index:2;${u ? `background-image:url("${u}")` : ''}`, 'aria-label': `${t.name} ${p} of ${t.maxPoints}`,
          }, el('span', { class: 'talent-picker-points', text: `${p}/${t.maxPoints}` })), `tal:${b.class}:${t.tree}:${t.idx}`, true, b.id));
        }
        const art = (rowsOf('treeart') || []).find((a) => a.class === b.class && +a.tree === tr.ti);
        return el('div', { class: 'carousel-item active' }, el('div', { class: 'talent-tree-picker-root' },
          el('div', { class: 'talent-tree-header' },
            el('span', { class: 'talent-tree-icon' }, art ? iconImg(art.icon, '') : null, el('label', { class: 'talent-tree-points', text: String(tr.spent) })),
            el('span', { class: 'talent-tree-title', text: tr.name })),
          el('div', { class: 'talent-tree-background', style: art && art.bg ? `background-image:url("${art.bg}")` : '' }),
          main));
      }))))));
    paint(trees);
    const car = host.querySelector('#talents-carousel');
    const fits = !car.clientWidth || car.scrollWidth <= car.clientWidth + 1;
    $('#sp-tal-seg').closest('.fv-toolbar').hidden = fits;
    if (!fits && st.talTree !== 'all') paint(trees.filter((t) => String(t.ti) === st.talTree));
  };

  // ---------- stat weights ----------
  const drawWeights = (b) => {
    const host = $('#sp-weights');
    host.setAttribute('data-source', 'weights tooltips');
    if (!specWait(host, ['weights'], b, [$('#sp-wtab tbody')])) return;
    const rows = rowsOf('weights').filter((w) => w.build === b.id);
    if (!rows.length) { host.replaceChildren(el('span', { class: 'fv-note', text: 'No stat weights for this preset.' })); $('#sp-wtab tbody').replaceChildren(); $('#sp-wtab').dataset.build = b.id; return; }
    const ab = rows[0].primaryAbbr;
    $('#sp-weights-note').textContent = `In points of ${ab}. Hit and crit per rating point (10 hit or 14 crit rating = 1%).`;
    const val = (w) => (w.unit === '1%' ? w.ratingWorth : w.worth);
    const shown = rows.filter((w) => w.stat !== 'MeleeHitDown').sort((x, y) => val(y) - val(x));
    const down = rows.find((w) => w.stat === 'MeleeHitDown');
    const max = d3.max(shown, (w) => Math.abs(val(w))) || 1;
    host.replaceChildren(...shown.flatMap((w) => {
      const where = `build=${enc(b.id)}&stat=${enc(w.stat)}`;
      const nm = w.unit === '1%' ? w.label.replace(' (1%)', ' rating') : w.label;
      return [
        tt(el('span', { class: 'nm', text: nm }), 'w:' + w.stat),
        tt(el('span', { class: 'vl' }, mark(el('span', { text: f2(val(w)) }), 'weights', w.unit === '1%' ? 'ratingWorth' : 'worth', where), ` ${ab}`), 'w:' + w.stat),
        tt(el('span', { class: 'tr' }, el('span', { style: `width:${Math.max(0, 100 * val(w) / max)}%;background:${ccol(b.class)}` })), 'w:' + w.stat),
      ];
    }), ...(down ? [el('span', { class: 'fv-note', style: 'grid-column:1/-1', text: `Losing 1% hit costs ${f1(down.dpsPer)} DPS (${f2(down.ratingWorth)} ${ab} per rating point)${shown.find((w) => w.stat === 'MeleeHit')?.dpsPer < 0.5 ? '. At the hit cap, so more hit adds nothing.' : '.'}` })] : []));
    $('#sp-wtab tbody').replaceChildren(...rows.map((w) => el('tr', { 'data-key': `${enc(b.id)},${enc(w.stat)}` },
      el('td', { text: w.label }), el('td', { class: 'n', text: fmt('.3f')(w.dpsPer) }), el('td', { class: 'n', text: fmt('.3f')(w.sePer) }),
      el('td', { class: 'n', text: w.worth == null ? '' : f2(w.worth) }), el('td', { class: 'n', text: w.ratingWorth == null ? '' : f2(w.ratingWorth) }))));
    $('#sp-wtab').dataset.build = b.id;
  };

  // ---------- spec ----------
  const ABILITY_COLORS = (abil) => {
    const m = new Map();
    abil.filter((a) => a.dps > 0).slice(0, 8).forEach((a, i) => m.set(a.key, nameColor(a.key) || dash.colors[i]));
    return m;
  };
  const tile = (k, v, s, src, field, where, unit, icon) => el('div', { class: 'fv-tile' }, el('span', { class: 'k', text: k }), el('span', { class: 'v' }, icon ? iconImg(icon, 'fv-ico md') : null, icon ? ' ' : null, mark(el('span', { text: v }), src, field, where), unit ? el('small', { text: unit }) : null), s ? el('span', { class: 's', text: s }) : null);

  const drawSpec = () => {
    if (!waitOr($('#sp-head'), ['builds'])) return;
    const builds = visB();
    const order = buildOrder(builds);
    if (!st.build || !builds.find((b) => b.id === st.build)) st.build = order.find((b) => b.isDefault)?.id || order[0].id;
    const b = builds.find((x) => x.id === st.build);
    const W = 'id=' + enc(b.id);
    sortedBuildOptions($('#sp-pick'), builds, b.id);
    specPicker(builds, b.id);
    $('#sp-dot').style.background = ccol(b.class);
    $('#sp-dot').hidden = !!iconMap();
    $('#sp-specicon').replaceChildren(...(iconMap() ? [ringed(b.specIcon + ':l', 'fv-ico lg', ccol(b.class)) || specIco(b, 'lg')] : []));
    const arts = rowsOf('treeart');
    if (arts) {
      const parts = String(b.talents).split('-').map((p) => [...p].reduce((a, c) => a + (+c || 0), 0));
      const main = parts.indexOf(Math.max(...parts));
      const art = arts.find((a) => a.class === b.class && +a.tree === main);
      $('#sp-art').style.backgroundImage = art && art.bg ? `url("${art.bg}")` : '';
    }
    mark($('#sp-title'), 'builds', 'label', W).textContent = b.label;
    // replaceChildren writes a null as the text "null", so we leave out the encounter badge when there's none.
    $('#sp-meta').replaceChildren(...[
      tt(el('span', {}, ringed('classicon_' + b.class.toLowerCase(), 'fv-ico ring', ccol(b.class)), ` ${b.spec}`), 'c:' + b.class),
      tt(el('span', {}, raceIco(b.raceIcon), ' ', mark(el('span', { text: b.race }), 'builds', 'race', W), ' ', iconImg(b.faction === 'Horde' ? 'ui_hordeicon' : 'ui_allianceicon')), `r:${b.race}:${b.class}`, true, b.id),
      mark(el('span', { text: b.treePoints }), 'builds', 'treePoints', W),
      isEnc(b) ? mark(encBadge(b), 'builds', 'encounter', W) : null].filter(Boolean));
    mark($('#sp-summary'), 'builds', 'summary', W).textContent = b.summary;
    const extra = b.petDps > 0 ? tile('Pet damage', f1(b.petShare), `${f1(b.petDps)} DPS from pets`, 'builds', 'petShare', W, '%')
      : b.oomSecs > 0 ? tile('Time out of mana', f1(b.oomSecs), 'per fight', 'builds', 'oomSecs', W, 's')
        : tile('Threat per second', f0(b.tps), 'for tank aggro', 'builds', 'tps', W);
    $('#sp-tiles').replaceChildren(
      tile('Mean DPS', f1(b.dps), `± ${f2(b.se)} (1 s.e.)`, 'builds', 'dps', W),
      tile('Rank', String(rankOf(b, builds)), `of ${builds.length} presets`, 'builds', st.enc ? 'rankAll' : 'rank', W),
      tile('Unlucky fight', f1(b.p10), '10th percentile', 'builds', 'p10', W),
      tile('Lucky fight', f1(b.p90), '90th percentile', 'builds', 'p90', W),
      tt(tile('Best race', b.bestRace, `${f1(b.bestRaceDps)} DPS`, 'builds', 'bestRace', W, null, b.bestRaceIcon), `r:${b.bestRace}:${b.class}`, true, b.id),
      extra);
    $('#sp-asm').hidden = !b.asmF;
    for (const h of ['#sp-head', '#sp-tiles']) $(h).dataset.build = b.id;
    const i = order.findIndex((x) => x.id === b.id);
    $('#sp-prev').onclick = () => openSpec(order[(i - 1 + order.length) % order.length].id);
    $('#sp-next').onclick = () => openSpec(order[(i + 1) % order.length].id);

    // caveats: the same groups as Method's, for this preset
    const cavHost = $('#sp-caveats');
    if (specWait(cavHost, ['notes', 'builds'], b)) {
      const builds = new Map(visB().map((x) => [x.id, x]));
      if (!builds.has(b.id)) builds.set(b.id, b);
      const ns = rowsOf('notes').filter((n) => n.kind === 'caveats' && builds.get(n.build)?.class === b.class);
      const groups = cavGroups(b.class, ns, builds, b.id);
      cavHost.replaceChildren(groups.childElementCount ? groups : el('p', { class: 'fv-note', text: 'None recorded.' }));
    }

    // a card that throws shows a failure note (never the previous preset's content)
    const safe = (fn, host, extras = []) => { try { fn(); } catch (err) { console.error(err); if (host) { fail(host); host.dataset.build = b.id; } for (const x of extras) if (x) x.replaceChildren(); } };
    safe(() => drawTalents(b), $('#sp-tal'));
    safe(() => drawDist(b), $('#sp-dist'));
    safe(() => drawWeights(b), $('#sp-weights'));
    const abil = (rowsOf('abilities') || []).filter((a) => a.build === b.id);
    const colors = ABILITY_COLORS(abil);
    safe(() => drawAbilities(b, abil, colors), $('#sp-abil'));
    safe(() => drawEvents(b), $('#sp-events'));
    safe(() => drawAuras(b), $('#sp-auras'));
    safe(() => drawRaceRank(b), $('#sp-races'));
    safe(() => drawRacialRank(b), $('#sp-racials'));
    $('#sp-races-link').onclick = (e) => { e.stopPropagation(); pickRaceSpec(b.id); };
    safe(() => drawSpecLen(b), $('#sp-len'));
    safe(() => drawTables(b));
  };
  $('#sp-pick').addEventListener('change', (e) => openSpec(e.target.value));

  // histogram of fights per 5 DPS bin; tap or hover a bin for its share, count and the share of fights at or below it
  const drawHist = (host, sets, opts = {}) => {
    const W = host.clientWidth, H = opts.h || 200;
    const m = { l: 8, r: 8, t: 22, b: 26 };
    const cum = opts.mode === 'cum';
    for (const s of sets) { let c = 0; for (const d of s.bins) { c += d.share; d.cum = c; } }
    const val = (d) => (cum ? d.cum : d.share);
    const lo = d3.min(sets, (s) => d3.min(s.bins, (d) => d.lo)), hi = d3.max(sets, (s) => d3.max(s.bins, (d) => d.lo)) + 5;
    const x = d3.scaleLinear([lo, hi], [m.l, W - m.r]);
    const y = d3.scaleLinear([0, d3.max(sets, (s) => d3.max(s.bins, val))], [H - m.b, m.t]);
    const svg = d3.create('svg').attr('width', '100%').attr('viewBox', `0 0 ${W} ${H}`).attr('height', H);
    svg.append('g').attr('transform', `translate(0,${H - m.b})`).call(d3.axisBottom(x).ticks(Math.max(3, W / 80)).tickSizeOuter(0)).attr('font-size', null).attr('font-family', null)
      .call((g) => g.selectAll('line,path').style('stroke', 'var(--cds-chart-axis)'));
    if (cum) for (const q of [0.5, 1]) svg.append('text').attr('x', m.l).attr('y', y(q) - 3).text(d3.format('.0%')(q)).style('font-size', '9px');
    const bars = new Map();
    for (const s of sets) {
      const g = svg.append('g');
      g.selectAll('rect').data(s.bins).join('rect').attr('x', (d) => x(d.lo) + 0.5).attr('width', (d) => Math.max(1, x(d.lo + 5) - x(d.lo) - 1))
        .attr('y', (d) => y(val(d))).attr('height', (d) => y(0) - y(val(d))).attr('rx', 1.5).style('fill', s.color).attr('opacity', s.op ?? 0.85)
        .each(function (d) { const k = d.lo; if (!bars.has(k)) bars.set(k, []); bars.get(k).push(this); });
    }
    for (const mk of opts.marks || []) {
      svg.append('line').attr('x1', x(mk.v)).attr('x2', x(mk.v)).attr('y1', m.t - 4).attr('y2', H - m.b).style('stroke', 'var(--color-fg)').attr('stroke-width', mk.strong ? 1.5 : 1).attr('stroke-dasharray', mk.strong ? null : '3 3').style('pointer-events', 'none');
      svg.append('text').attr('x', x(mk.v)).attr('y', m.t - 8).attr('text-anchor', 'middle').text(mk.label);
    }
    const los = [...bars.keys()].sort((a, b) => a - b);
    const hit = svg.append('g').selectAll('rect').data(los).join('rect').attr('x', (k) => x(k)).attr('width', (k) => Math.max(2, x(k + 5) - x(k))).attr('y', m.t).attr('height', H - m.b - m.t).style('fill', 'transparent');
    tipOn(hit, (e, k) => {
      const [px, py] = d3.pointer(e, host);
      const bin = (s) => s.bins.find((q) => q.lo === k);
      // One preset: its share of fights as the key figure. Two: a row each.
      const lines = sets.length === 1
        ? (() => { const d = bin(sets[0]); return d ? [tipKey(d3.format('.2%')(d.share), 'of fights', `(${f0(d.count)})`), tipGrid([['At or below', d3.format('.1%')(d.cum)]])] : [tipText('No fights')]; })()
        : [tipTable(['', 'Fights', 'At or below'], sets.map((s) => { const d = bin(s); return [el('span', {}, tipSw(s.color), s.name.split(' · ')[0]), d ? d3.format('.2%')(d.share) : '–', d ? d3.format('.1%')(d.cum) : '–']; }))];
      showTip(host, px, py, `${k}–${k + 5} DPS`, lines, { hl: bars.get(k) });
    });
    host.replaceChildren(svg.node());
  };
  const binsFor = (hist, id) => {
    const rows = hist.filter((h) => h.build === id).sort((a, b) => a.lo - b.lo);
    const n = d3.sum(rows, (r) => r.count);
    return rows.map((r) => ({ lo: r.lo, count: r.count, share: r.count / n }));
  };

  const drawDist = (b) => {
    const host = $('#sp-dist');
    host.setAttribute('data-source', 'hist builds');
    if (!specWait(host, ['hist'], b, [$('#sp-pct thead tr'), $('#sp-pct tbody tr')])) return;
    seg($('#sp-dist-mode'), 'distMode', [['bins', 'Per bin'], ['cum', 'Cumulative']], () => drawDist(b));
    drawHist(host, [{ name: b.build, color: ccol(b.class), bins: binsFor(rowsOf('hist'), b.id) }],
      { mode: st.distMode, marks: [{ v: b.p10, label: '10th' }, { v: b.dps, label: 'mean', strong: true }, { v: b.p90, label: '90th' }] });
    const t = $('#sp-pct'), W = 'id=' + enc(b.id);
    const cols = [['Worst', 'min'], ['1st', 'p1'], ['10th', 'p10'], ['25th', 'p25'], ['Median', 'p50'], ['75th', 'p75'], ['90th', 'p90'], ['99th', 'p99'], ['Best', 'max'], ['Std dev', 'stdev']];
    t.querySelector('thead tr').replaceChildren(...cols.map(([h]) => el('th', { class: 'n', text: h })));
    t.querySelector('tbody tr').replaceChildren(...cols.map(([, k]) => mark(el('td', { class: 'n', text: f1(b[k]) }), 'builds', k, W)));
    t.dataset.build = b.id;
    $('#sp-dist-note').textContent = `${f0(b.iters)} fights.`;
  };

  const drawAbilities = (b, abil, colors) => {
    const host = $('#sp-abil');
    host.setAttribute('data-source', 'abilities tooltips');
    if (!specWait(host, ['abilities'], b, [$('#sp-abiltab tbody')])) return;
    seg($('#sp-abil-metric'), 'abilMetric', [['share', 'Damage share'], ['casts', 'Casts'], ['critPct', 'Crit %']], () => drawAbilities(b, abil, colors));
    const mt = ['share', 'casts', 'critPct'].includes(st.abilMetric) ? st.abilMetric : 'share';
    const rows = abil.filter((a) => a.dps > 0 && (mt === 'share' || a[mt] > 0)).sort((p, q) => q[mt] - p[mt]);
    const max = d3.max(rows, (a) => a[mt]) || 1;
    host.replaceChildren(...rows.flatMap((a) => {
      const where = `build=${enc(b.id)}&key=${enc(a.key)}`;
      const vl = mt === 'share' ? [mark(el('span', { text: f1(a.share) }), 'abilities', 'share', where), '% · ', mark(el('span', { text: f1(a.dps) }), 'abilities', 'dps', where)]
        : mt === 'casts' ? [mark(el('span', { text: f1(a.casts) }), 'abilities', 'casts', where), ' per fight'] : [mark(el('span', { text: f1(a.critPct) }), 'abilities', 'critPct', where), '%'];
      return [
        tt(el('span', { class: 'nm' }, el('span', { class: 'fv-dot', style: `background:${colors.get(a.key) || 'var(--cds-chart-muted)'}` }), keyIcon(a.key), mark(el('span', { text: a.ability }), 'abilities', 'ability', where)), 'k:' + a.key),
        tt(el('span', { class: 'vl' }, ...vl), 'k:' + a.key),
        tt(el('span', { class: 'tr' }, el('span', { style: `width:${100 * a[mt] / max}%;background:${colors.get(a.key) || 'var(--cds-chart-muted)'}` })), 'k:' + a.key),
      ];
    }));
    // A column that doesn't apply to an ability shows a dash, not 0: Deep Wounds or a proc that only ticks has no
    // hits or casts, and a direct hit has no ticks.
    const HIT_COLS = ['hits', 'crits', 'critPct', 'avgHit', 'avgCrit', 'misses', 'dodges', 'glances'];
    const abilCell = (a, k) => {
      const direct = a.hits + a.crits + a.misses + a.dodges + a.glances > 0;
      if ((HIT_COLS.includes(k) && !direct) || ((k === 'ticks' || k === 'avgTick') && !(a.ticks > 0)) || (k === 'casts' && !(a.casts > 0))) return '–';
      return ['avgHit', 'avgCrit', 'avgTick'].includes(k) ? f0(a[k]) : f1(a[k]);
    };
    const tb = $('#sp-abiltab tbody');
    tb.replaceChildren(...abil.map((a) => tt(el('tr', { 'data-key': `${enc(b.id)},${enc(a.key)}` },
      el('td', { class: 'first' }, keyIcon(a.key), ' ' + a.ability), ...['dps', 'share', 'casts', 'hits', 'crits', 'critPct', 'ticks', 'avgHit', 'avgCrit', 'avgTick', 'misses', 'dodges', 'glances']
        .map((k) => el('td', { class: 'n', text: abilCell(a, k) }))), 'k:' + a.key)));
    $('#sp-abiltab').dataset.build = b.id;
  };

  // each aura's average uptime (every fight), and a mini timeline of the sample fight picked in the events view
  const drawAuras = (b) => {
    const host = $('#sp-auras');
    const src = 'timeline_' + b.class.toLowerCase();
    host.setAttribute('data-source', `auras ${src} names tooltips`);
    $('#sp-alwayson').setAttribute('aria-pressed', String(st.alwaysOn));
    if (!specWait(host, ['auras'], b)) return;
    seg($('#sp-aura-owner'), 'auraOwner', [['All', 'All'], ['Self', 'On the player'], ['Target', 'On the boss']]);
    const T = statusOf(src) === 'ok' ? tlSamples(b) : null;
    const s = T && T.list.length ? T.list[Math.min(st.sample || 0, T.list.length - 1)] : null;
    const t0 = s ? s.t0 : 0, t1 = s ? s.dur : 1;
    const pos = (t) => (100 * (Math.max(t0, Math.min(t1, t)) - t0)) / (t1 - t0);
    const rows = rowsOf('auras').filter((a) => a.build === b.id && (!st.alwaysOn || a.uptime < 99.5) && (st.auraOwner === 'All' || (st.auraOwner === 'Self') === (a.owner === 'Self'))).sort((x, y) => (x.owner === y.owner ? y.uptime - x.uptime : x.owner === 'Self' ? -1 : 1));
    const card = $('#sp-auras-card');
    const out = [];
    let last = null;
    for (const a of rows) {
      if (a.owner !== last) { out.push(el('span', { class: 'nm', style: 'grid-column:1/-1;font-weight:500;color:var(--color-fg)', text: a.owner === 'Self' ? 'On the player' : 'On the boss' })); last = a.owner; }
      const where = `build=${enc(b.id)}&key=${enc(a.key)}&owner=${a.owner}`;
      const col = nameColor(a.key) || (a.owner === 'Self' ? dash.colors[0] : dash.colors[1]);
      const iv = s ? s.auras.filter((q) => q[0] === a.key && (a.owner === 'Target' ? q[1] === 't' : q[1] !== 't') && q[3] > q[2]) : [];
      const upHere = s ? d3.sum(iv, (q) => Math.min(q[3], t1) - Math.max(q[2], 0)) / t1 * 100 : null;
      out.push(tt(el('span', { class: 'nm' }, a.icon ? iconImg(a.icon) : null, mark(el('span', { text: a.aura }), 'auras', 'aura', where), a.procs > 1.5 ? el('small', { class: 'fv-note', text: `${f1(a.procs)}× per fight` }) : null), 'a:' + a.key),
        tt(el('span', { class: 'vl' }, mark(el('span', { text: f1(a.uptime) }), 'auras', 'uptime', where), '% avg'), 'a:' + a.key));
      let bar;
      if (s) {
        bar = el('span', { class: 'tr fv-mini', 'data-tip': '1', role: 'button', 'aria-label': `${a.aura}: uptime in the sample fight` },
          t0 < -0.55 ? el('i', { class: 'pre', style: `width:${pos(0)}%` }) : null,
          ...iv.map((q) => el('span', { style: `left:${pos(q[2])}%;width:${Math.max(0.4, pos(q[3]) - pos(q[2]))}%;background:${col}` }, q[4] > 1 && pos(q[3]) - pos(q[2]) > 6 ? el('b', { text: '×' + q[4] }) : null)),
          el('i', { class: 'pull', style: `left:${pos(0)}%` }));
        bar.addEventListener('click', (e) => {
          e.stopPropagation();
          if (bar.classList.contains('is-hl')) { hideTip(); return; }
          const r = card.getBoundingClientRect(), nb = bar.getBoundingClientRect();
          const tm = (t) => `${Math.max(0, t).toFixed(0)}`;
          const lines = [tipKey(`${f1(a.uptime)}%`, 'uptime', 'every fight'), tipGrid([[`This ${s.label.toLowerCase().replace(' fight', '')} fight`, `${f1(upHere)}% · ${iv.length}×`],
            d3.max(iv, (q) => q[4]) > 1 ? ['Most stacks', String(d3.max(iv, (q) => q[4]))] : null]),
            iv.length ? tipChips(iv.slice(0, 8).map((q) => `${tm(q[2])}–${q[3] >= s.dur - 0.05 ? 'end' : tm(q[3])}s`), iv.length > 8 ? `+${iv.length - 8} more` : null) : null];
          showTip(card, e.clientX - r.left - 20, nb.bottom - r.top, tipTitle(a.aura, null, a.icon ? iconImg(a.icon) : null), lines, { hl: bar, below: true });
        });
      } else {
        bar = el('span', { class: 'tr' }, el('span', { style: `width:${a.uptime}%;background:${col}` }));
      }
      out.push(bar);
    }
    host.replaceChildren(...(out.length ? out : [el('span', { class: 'fv-note', text: 'Every aura is always up.' })]));
    $('#sp-auras-note').textContent = s ? `Number: average uptime. Bar: the ${s.label.toLowerCase()}${t0 < -0.55 ? ' (pre-pull shaded)' : ''}.` : 'Average uptime.';
  };
  $('#sp-alwayson').addEventListener('click', (e) => { e.stopPropagation(); st.alwaysOn = !st.alwaysOn; persist(); render(); });
  $('#sp-setup-hide').addEventListener('click', (e) => { e.stopPropagation(); st.setupHide = !st.setupHide; persist(); $('#sp-setup').classList.toggle('hide-off', st.setupHide); e.currentTarget.setAttribute('aria-pressed', String(st.setupHide)); });

  const drawSpecLen = (b) => {
    const host = $('#sp-len');
    host.setAttribute('data-source', 'lengths');
    if (!specWait(host, ['lengths'], b)) return;
    const pts = rowsOf('lengths').filter((r) => r.build === b.id).sort((p, q) => p.secs - q.secs).map((p) => ({ s: p.secs, v: p.dps, oom: p.oomSecs }));
    drawLenChart(host, [{ name: b.build, color: ccol(b.class), dots: true, pts }], { h: 200, tipMain: (p) => [`${f1(p.v)} DPS`, p.s === 180 ? null : el('span', {}, tipDelta(100 * (p.v / b.dps - 1)), ' vs 180s')], tipRows: (p) => [p.oom > 0 ? ['Out of mana', `${f1(p.oom)}s`] : null] });
  };

  // ---------- rotation ----------
  // the exact APL condition, compactly: what a reader of the sim UI's rotation tab would recognise
  const aplText = (v) => {
    if (!v || typeof v !== 'object') return String(v ?? '');
    const [t, a] = Object.entries(v)[0] || [];
    const nm = (r) => { if (!r) return '?'; const k = r.spellId ? 's' + r.spellId + (r.tag ? 't' + r.tag : '') : r.itemId ? 'i' + r.itemId : r.otherId ? 'o' + r.otherId : null; return (k && byKey('names')?.get(k)?.name) || k || '?'; };
    const OP = { OpLt: '<', OpLe: '≤', OpGt: '>', OpGe: '≥', OpEq: '=', OpNe: '≠', OpAdd: '+', OpSub: '−', OpMul: '×', OpDiv: '÷' };
    if (t === 'const') return String(a.val);
    if (t === 'and' || t === 'or') return (a.vals || []).map((x) => { const s = aplText(x); return /\b(and|or)\b/.test(Object.keys(x)[0]) ? `(${s})` : s; }).join(t === 'and' ? ' AND ' : ' OR ');
    if (t === 'not') return `NOT ${aplText(a.val)}`;
    if (t === 'cmp' || t === 'math') return `${aplText(a.lhs)} ${OP[a.op] || a.op} ${aplText(a.rhs)}`;
    const ref = a && (a.spellId || a.auraId);
    const args = [ref ? nm(ref) : null, a && a.delay ? 'delay ' + aplText(a.delay) : null, a && a.autoType ? a.autoType : null, a && a.threshold ? a.threshold : null].filter(Boolean);
    return `${t}(${args.join(', ')})`;
  };
  const rotRows = (bid) => (rowsOf('rotation') || []).filter((r) => r.build === bid);
  const rotSpec = (section, order, b) => {
    const r = rotRows(b).find((x) => x.section === section && String(x.order) === String(order));
    if (!r) return null;
    const base = r.key ? spellSpec(r.key, b, false) : { title: r.name, color: GOLD, lines: [], nums: [] };
    const cond = r.cond ? tlParse(r.cond, null) : null;
    // A priority line's card already shows its condition and note, so the tooltip doesn't repeat them; pre-pull casts
    // and folded lines (other races' racials, lines switched off) show them here. The exact APL condition is folded.
    const shown = section === 'priority' && !r.otherRace && !/^Never/.test(r.when || '');
    const lines = [
      ['h', section === 'prepull' ? `Pre-pull · ${fmtT(+r.at)}` : `Priority ${r.order}`, r.race ? (r.otherRace ? `${r.race} only · skipped` : `${r.race} racial`) : String(r.kind)],
      !shown && r.when ? ['t', r.when] : null,
      !shown && r.extra ? ['t', r.extra] : null,
      !shown && r.notes ? ['m', r.notes] : null,
      cond ? ['d', 'Exact condition', aplText(cond)] : null,
    ].filter(Boolean);
    return { ...base, icon: base.icon || r.icon, title: String(r.key && !/^oOther/.test(r.key) ? base.title : r.name), lines: [...lines, ...(base.lines && base.lines.length ? [['s'], ...base.lines] : [])] };
  };
  const KIND_TAG = { cast: 'Cast', channel: 'Channel', move: 'Move', auto: 'Auto' };
  const drawRotation = (b) => {
    const host = $('#sp-rotation'), pre = $('#sp-prepull');
    if (!specWait(host, ['rotation'], b, [pre])) return;
    const rows = rotRows(b.id);
    const W = (r) => `build=${enc(b.id)}&section=${r.section}&order=${r.order}`;
    const ico = (r, cls) => (r.icon ? iconImg(r.icon, cls) : el('span', { class: cls + ' fv-noico', text: r.kind === 'move' ? '➜' : r.kind === 'auto' ? '⟳' : String(r.name).slice(0, 2) }));
    const pp = rows.filter((r) => r.section === 'prepull').sort((p, q) => p.at - q.at);
    pre.replaceChildren(...(pp.length ? [el('span', { class: 'fv-label', text: 'Before the pull' }),
      el('div', { class: 'fv-pp' }, ...pp.map((r) => tt(el('button', { class: 'fv-ppi' }, ico(r, 'fv-ico'), el('span', {}, mark(el('b', { text: r.name }), 'rotation', 'name', W(r)), mark(el('em', { text: fmtT(+r.at) }), 'rotation', 'at', W(r)))), `rot:prepull:${r.order}`, true, b.id)),
        el('span', { class: 'fv-ppi pull' }, el('b', { text: 'Pull' }), el('em', { text: '0s' })))] : []));
    pre.dataset.build = b.id;
    const pl = rows.filter((r) => r.section === 'priority').sort((p, q) => p.order - q.order);
    const items = [];
    for (let i = 0; i < pl.length; i++) {
      const r = pl[i];
      const fold = (q) => (q.otherRace ? 'race' : /^Never/.test(q.when) ? 'off' : null);
      if (fold(r)) { // other races' racials, and lines switched off, fold into one row each
        const why = fold(r), run = [r];
        while (pl[i + 1] && fold(pl[i + 1]) === why) run.push(pl[++i]);
        items.push(el('li', { class: 'fv-pr other', value: String(run[0].order) },
          el('span', { class: 'no', text: run.length > 1 ? `${run[0].order}–${run[run.length - 1].order}` : String(run[0].order) }),
          el('div', { class: 'tx' }, el('div', { class: 'ics' }, ...run.map((q) => tt(ico(q, 'fv-ico'), `rot:priority:${q.order}`, true, b.id)), el('b', { text: run.map((q) => q.name).join(', ') })), el('span', { class: 'wh', text: why === 'race' ? `${[...new Set(run.map((q) => q.race))].join(', ')} racials.` : `Off in this preset${run[0].notes ? ': ' + run[0].notes.replace(/^Never:\s*/i, '') : '.'}` }))));
        continue;
      }
      items.push(tt(el('li', { class: 'fv-pr', value: String(r.order) },
        el('span', { class: 'no', text: String(r.order) }),
        ico(r, 'fv-ico big'),
        el('div', { class: 'tx' },
          el('div', { class: 'hd' }, mark(el('b', { text: r.name }), 'rotation', 'name', W(r)), r.kind !== 'cast' ? el('span', { class: 'kd', text: KIND_TAG[r.kind] || r.kind }) : null, r.race ? el('span', { class: 'kd race', text: r.race }) : null),
          r.when ? mark(el('span', { class: 'wh', text: r.when }), 'rotation', 'when', W(r)) : null,
          r.extra ? el('span', { class: 'wh', text: r.extra }) : null,
          r.notes ? el('span', { class: 'nt', text: r.notes }) : null),
        r.casts > 0 ? el('span', { class: 'ct' }, mark(el('b', { text: f1(r.casts) }), 'rotation', 'casts', W(r)), el('small', { text: 'casts' })) : null), `rot:priority:${r.order}`, true, b.id));
    }
    host.replaceChildren(...(items.length ? items : [el('li', { class: 'fv-note', text: 'No rotation recorded.' })]));
    $('#sp-rot-note').textContent = 'The sim casts the first line whose condition holds.';
  };

  // ---------- gear (paper doll, as the sim UI's gear picker) ----------
  const SLOT_ICON = { Head: 'head', Neck: 'neck', Shoulders: 'shoulders', Back: 'shirt', Chest: 'chest', Wrist: 'wrists', Hands: 'hands', Waist: 'waist', Legs: 'legs', Feet: 'feet', 'Ring 1': 'finger', 'Ring 2': 'finger', 'Trinket 1': 'trinket', 'Trinket 2': 'trinket', 'Main hand': 'mainhand', 'Off hand': 'offhand', Ranged: 'ranged' };
  // The sim UI's gear picker (gear_picker.tsx) and character sheet (character_stats.tsx) markup, read only.
  const GEAR_LEFT = ['Head', 'Neck', 'Shoulders', 'Back', 'Chest', 'Wrist', 'Main hand', 'Off hand', 'Ranged'];
  const GEAR_RIGHT = ['Hands', 'Waist', 'Legs', 'Feet', 'Ring 1', 'Ring 2', 'Trinket 1', 'Trinket 2'];
  const drawGear = (b) => {
    const host = $('#sp-gear');
    if (!specWait(host, ['gear'], b)) return;
    const gear = rowsOf('gear').filter((g) => g.build === b.id);
    const by = new Map(gear.map((g) => [g.slot, g]));
    const twoHand = (g) => /Two-Hand|Staff|Polearm/i.test(arr(ttRow('i' + g.itemId)?.lines).map((l) => l[1]).join(' '));
    const sets = d3.rollup(gear, (v) => v.length, (g) => ttRow('i' + g.itemId)?.set || null);
    const icons = iconMap();
    const iconUrl = (name) => (icons && name ? icons.get(name) || null : null);
    const slot = (name) => {
      const g = by.get(name);
      const W = `build=${enc(b.id)}&slot=${enc(name)}`;
      const iconEl = (u) => el('a', { class: 'item-picker-icon', role: 'button', style: u ? `background-image:url("${u}")` : '' });
      if (!g) {
        const mh = by.get('Main hand');
        return el('div', { class: 'item-picker-root fv-slot-empty' },
          el('div', { class: 'item-picker-icon-wrapper' }, iconEl(iconUrl('slot_' + SLOT_ICON[name]))),
          el('div', { class: 'item-picker-labels-container' },
            el('span', { class: 'item-picker-name text-muted', text: name === 'Off hand' && mh && twoHand(mh) ? 'Two-hand weapon in the main hand' : `${name}: empty` })));
      }
      const r = ttRow('i' + g.itemId);
      const q = r?.quality || g.quality || 'common';
      const set = r?.set ? r.set : null;
      const use = arr(r?.lines).some((l) => /^Use:/.test(String(l[1])));
      const synth = /Synthetic/.test(g.item) && /^Trinket/.test(name);
      return tt(el('div', { class: 'item-picker-root', 'data-key': `${enc(b.id)},${enc(name)}` },
        el('div', { class: 'item-picker-icon-wrapper' },
          g.ilvl ? mark(el('span', { class: 'item-picker-ilvl', text: String(g.ilvl) }), 'gear', 'ilvl', W) : null,
          iconEl(iconUrl(g.icon) || iconUrl('slot_' + SLOT_ICON[name]))),
        el('div', { class: 'item-picker-labels-container' },
          mark(el('a', { class: `item-picker-name item-quality-${q}`, role: 'button', text: g.item.replace(/ \(Phase 1\)$/, '') }), 'gear', 'item', W),
          g.enchant ? mark(el('a', { class: 'item-picker-enchant', role: 'button', text: g.enchant }), 'gear', 'enchant', W) : null,
          set || use || synth ? el('span', { class: 'fv-gear-tags' }, set ? el('em', { class: 'set', text: `Set ${sets.get(set)}/${r.setSize ?? sets.get(set)}` }) : null, synth ? el('em', { text: use ? 'On use' : 'Stat trinket' }) : use ? el('em', { text: 'On use' }) : null) : null)),
      'g:' + name);
    };
    host.replaceChildren(el('div', { class: 'gear-picker-left tab-panel-col' }, ...GEAR_LEFT.map(slot)), el('div', { class: 'gear-picker-right tab-panel-col' }, ...GEAR_RIGHT.map(slot)));
  };

  // ---------- character stats (as the sim UI's character sheet) ----------
  const STAT_GROUPS = [['Primary', ['Health', 'Mana']], ['Attributes', ['Strength', 'Agility', 'Stamina', 'Intellect', 'Spirit']],
    ['Physical', ['AttackPower', 'RangedAttackPower', 'MeleeHit', 'MeleeCrit', 'MeleeHaste', 'ArmorPenetration', 'Expertise']],
    ['Spell', ['SpellPower', 'SpellDamage', 'ArcanePower', 'FirePower', 'FrostPower', 'HolyPower', 'NaturePower', 'ShadowPower', 'SpellHit', 'SpellCrit', 'SpellHaste', 'MP5']],
    ['Defense', ['Armor']]];
  const PCT_STATS = new Set(['MeleeHit', 'MeleeCrit', 'SpellHit', 'SpellCrit', 'MeleeHaste', 'SpellHaste']);
  const drawStats = (b) => {
    const host = $('#sp-stats');
    if (!specWait(host, ['stats'], b)) return;
    const st2 = new Map(rowsOf('stats').filter((s) => s.build === b.id).map((s) => [s.stat, s]));
    host.replaceChildren(el('table', { class: 'character-stats-table' }, ...STAT_GROUPS.map(([, ks]) => {
      const have = ks.filter((k) => st2.has(k));
      if (!have.length) return null;
      return el('tbody', {}, ...have.map((k) => {
        const s = st2.get(k);
        return el('tr', { class: 'character-stats-table-row' }, el('td', { class: 'character-stats-table-label', text: s.label }),
          el('td', { class: 'character-stats-table-value' }, mark(el('span', { text: PCT_STATS.has(k) ? f2(s.value) + '%' : d3.format(',')(Math.round(s.value)) }), 'stats', 'value', `build=${enc(b.id)}&stat=${enc(k)}`)));
      }));
    }).filter(Boolean)));
  };

  // ---------- consumes and buffs (as the sim UI's icon pickers) ----------
  // Sections and subsections follow the sim UI: the Consumables picker's rows, the Raid and Party
  // Buffs subsections and the Offensive and Defensive Debuffs subsections (ui/core/components).
  // The report is Forever's, so it groups the buffs as the sim UI does under Forever: Battle
  // Shout, Trueshot Aura, the totems but Mana Spring and Stoneskin, the crit auras and the
  // resistances reach the whole raid (client 1.60.1.70425; see RAID_BUFF_SUBSECTIONS).
  const SETUP_SECTIONS = ['Consumables', 'Raid Buffs', 'Party Buffs', 'Offensive Debuffs', 'Defensive Debuffs', 'Professions'];
  const SETUP_SUBS = {
    'Consumables': ['Potion', 'Special', 'Flask', 'Weapon', 'Food', 'Attributes', 'Physical', 'Defensive', 'Spell', 'Misc', 'Engineering', 'Pet'],
    'Raid Buffs': ['Stats', 'Blessings', 'Attack Power', 'Totems', 'Crit', 'Resistances', 'Mana', 'Other'],
    'Party Buffs': ['Totems', 'Defense', 'Other'],
    'Offensive Debuffs': ['Sunder / Expose', 'CoR / FF', 'CoE', 'Other'],
    'Defensive Debuffs': ['Attack Power', 'Attack Speed', 'Chance to Hit', 'Other'],
    'Professions': [''],
  };
  const CONSUME_SUB = [
    [/^Default Potion|Potion$/, 'Potion'], [/Conjured|Rune|Thistle/, 'Special'], [/^Flask/, 'Flask'], [/Imbue|Weapon/, 'Weapon'],
    [/^Food/, 'Food'], [/^(Agility|Strength|Intellect|Spirit|Stamina)|Scroll/, 'Attributes'],
    [/^(Attack Power|Hit|Force)/, 'Physical'], [/^(Health|Armor|Troll)/, 'Defensive'],
    [/^(Spell Power|Fire Power|Frost Power|Shadow Power|Nature Power|Holy Power|Arcane Power|Healing|Mana Regen|Mp5)/, 'Spell'],
    [/Explosive|Sapper|Engineering/, 'Engineering'], [/^Pet/, 'Pet'],
  ];
  const BUFF_SUB = [
    [/^(Gift Of The Wild|Mark Of The Wild|Power Word Fortitude|Prayer Of Fortitude|Arcane Brilliance|Arcane Intellect|Divine Spirit|Prayer Of Spirit)/, 'Raid Buffs', 'Stats'],
    [/^Blessing Of/, 'Raid Buffs', 'Blessings'], [/^Innervate/, 'Raid Buffs', 'Mana'], [/^(Power Infusion|Thorns)/, 'Raid Buffs', 'Other'],
    [/^(Battle Shout|Trueshot Aura)/, 'Raid Buffs', 'Attack Power'], [/^(Mana Spring Totem|Stoneskin Totem)/, 'Party Buffs', 'Totems'],
    [/Protection|Resistance/, 'Raid Buffs', 'Resistances'], [/Totem/, 'Raid Buffs', 'Totems'],
    [/^(Leader Of The Pack|Moonkin Aura)/, 'Raid Buffs', 'Crit'], [/^(Devotion Aura|Retribution Aura)/, 'Party Buffs', 'Defense'],
  ];
  const DEBUFF_SUB = [
    [/^(Sunder Armor|Expose Armor)/, 'Offensive Debuffs', 'Sunder / Expose'], [/^(Curse Of Recklessness|Faerie Fire)/, 'Offensive Debuffs', 'CoR / FF'],
    [/^Curse Of Elements/, 'Offensive Debuffs', 'CoE'],
    [/^(Demoralizing Roar|Demoralizing Shout)/, 'Defensive Debuffs', 'Attack Power'], [/^Thunder Clap/, 'Defensive Debuffs', 'Attack Speed'],
    [/^(Insect Swarm|Scorpid Sting)/, 'Defensive Debuffs', 'Chance to Hit'], [/^(Curse Of Weakness|Judgement Of Light)/, 'Defensive Debuffs', 'Other'],
  ];
  const setupPlace = (s) => {
    const g = s.group || s.category;
    if (g === 'Consumables') {
      const key = s.item.split(':')[0];
      const hit = CONSUME_SUB.find(([re]) => re.test(key));
      return ['Consumables', hit ? hit[1] : 'Misc'];
    }
    if (g === 'Professions') return ['Professions', ''];
    if (g === 'Boss debuffs') {
      const hit = DEBUFF_SUB.find(([re]) => re.test(s.item));
      return hit ? [hit[1], hit[2]] : ['Offensive Debuffs', 'Other'];
    }
    const hit = BUFF_SUB.find(([re]) => re.test(s.item));
    return hit ? [hit[1], hit[2]] : [g === 'Self buffs' ? 'Raid Buffs' : 'Party Buffs', 'Other'];
  };
  const drawSetup = (b) => {
    const host = $('#sp-setup');
    const bossEl = $('#sp-boss');
    bossEl.replaceChildren(el('b', { text: 'Boss: ' }), mark(el('span', { text: b.bossText || 'Level 63 boss' }), 'builds', 'bossText', 'id=' + enc(b.id)), isEnc(b) ? ' (advantaged)' : '');
    if (!specWait(host, ['setup'], b)) return;
    const rows = rowsOf('setup').filter((s) => s.build === b.id).map((s) => ({ s, p: setupPlace(s) }));
    const sections = d3.groups(rows, (x) => x.p[0]).sort((p, q) => SETUP_SECTIONS.indexOf(p[0]) - SETUP_SECTIONS.indexOf(q[0]));
    const pick = (s) => {
      const off = s.active === false;
      const r = ttRow('setup:' + s.item);
      // Forever took out every talent that improved a raid buff or debuff (Improved Blessing of Might,
      // Improved Mark of the Wild, ...), and the sim treats an improved one as regular, so no 'Imp' tag.
      const name = String(r?.title || s.item.split(': ').pop()).replace(/ \(improved\)/i, '').replace(/ (Of|The) /g, (m) => m.toLowerCase());
      // The icon only, like the sim's settings page. The name is in its tooltip. A hand (MH or OH) or a level goes in a
      // label at its bottom, like the stat labels of the sim's elixirs.
      const label = s.hand || (s.level && s.level !== 'Improved' ? s.level : null);
      return tt(el('button', { class: 'fv-pick' + (off ? ' off' : ''), 'aria-label': off ? name + ' (no effect)' : name },
        s.icon ? iconImg(s.icon, 'fv-ico') : el('span', { class: 'fv-ico fv-noico', text: name.slice(0, 2) }),
        label ? el('span', { class: 'lb', text: label }) : null), 'su:' + s.item);
    };
    // Each section looks like its card on the sim's settings page: a gold small caps title, then its groups side by
    // side, each with a muted label over its icons. Every icon is a bag slot like the sim's picks. One that's on but
    // does nothing is gray, like an unpicked one.
    host.replaceChildren(...sections.map(([sec, xs]) => {
      const order = SETUP_SUBS[sec] || [];
      const subs = d3.groups(xs, (x) => x.p[1]).sort((p, q) => order.indexOf(p[0]) - order.indexOf(q[0]));
      const on = xs.filter((x) => x.s.active !== false).length;
      return el('section', { class: 'content-block fv-sgrp' + (sec === 'Consumables' ? ' wide' : '') },
        el('div', { class: 'content-block-header' }, el('h6', { class: 'content-block-title', text: sec }),
          on < xs.length ? el('small', { class: 'fv-sgrp-n', text: `${on} of ${xs.length} active` }) : null),
        el('div', { class: 'fv-ssubs' }, ...subs.map(([sub, ys]) => el('div', { class: 'fv-ssub' },
          sub ? el('div', { class: 'sh', text: sub }) : null,
          el('div', { class: 'fv-picks' }, ...ys.sort((p, q) => (p.s.active === false) - (q.s.active === false) || (p.s.slotOrder || 0) - (q.s.slotOrder || 0)).map((y) => pick(y.s)))))));
    }));
    host.classList.toggle('hide-off', !!st.setupHide);
    $('#sp-setup-hide').setAttribute('aria-pressed', String(!!st.setupHide));
  };

  const drawTables = (b) => {
    for (const [fn, h] of [[drawGear, '#sp-gear'], [drawStats, '#sp-stats'], [drawSetup, '#sp-setup'], [drawRotation, '#sp-rotation']]) {
      try { fn(b); } catch (err) { console.error(err); fail($(h)); $(h).dataset.build = b.id; }
    }
  };

  // ---------- sample fights ----------
  // The fight itself shows in the sim UI's events view (events.ts). These read the compact timeline datasets, which
  // the aura bars and the rotation table still use.
  const tlParse = (v, dflt) => { if (typeof v !== 'string') return v ?? dflt; try { return JSON.parse(v); } catch { return dflt; } };
  const tlCache = { rows: null, id: null, list: null };
  // the sample fights of one preset (typical, unlucky, lucky), parsed once per dataset load
  const tlSamples = (b) => {
    const src = 'timeline_' + b.class.toLowerCase();
    const rows = rowsOf(src);
    if (!rows) return null;
    if (tlCache.rows !== rows || tlCache.id !== b.id) {
      const samples = rows.filter((s) => s.build === b.id).map((s) => ({
        ...s, pct: +s.pct, seed: +s.seed, dps: +s.dps, dur: +s.dur,
        casts: tlParse(s.casts, []), auras: tlParse(s.auras, []),
      }));
      const ord = [samples.find((s) => s.pct === 50), samples.find((s) => s.pct === 10), samples.find((s) => s.pct === 90)].filter(Boolean);
      for (const s of ord) s.t0 = Math.min(0, d3.min(s.casts, (c) => c[0]) ?? 0, d3.min(s.auras, (a) => a[2]) ?? 0) - (d3.min(s.casts, (c) => c[0]) < 0 ? 0.6 : 0.5);
      Object.assign(tlCache, { rows, id: b.id, list: ord });
    }
    return { src, list: tlCache.list };
  };
  const nameColor = (k) => { const n = byKey('names')?.get(k); return n && n.color ? n.color : null; };
  const fmtT = (t) => `${t < 0 ? '−' : ''}${Math.abs(t).toFixed(1)}s`;

  // The events view of the preset's sample fights. Its Typical, Unlucky and Lucky iterations are st.sample 0, 1 and
  // 2, and the aura bars follow the one picked there.
  const drawEvents = (b) => {
    const T = statusOf('timeline_' + b.class.toLowerCase()) === 'ok' ? tlSamples(b) : null;
    const s = T && T.list.length ? T.list[Math.min(st.sample || 0, T.list.length - 1)] : null;
    $('#sp-time-note').textContent = s
      ? `${s.label}: ${f1(s.dps)} DPS over ${f1(s.dur)}s.`
      : '';
    events.show(b.id, st.sample || 0);
  };
  events.onSampleChange((sample) => {
    if (st.sample === sample) return;
    st.sample = sample; persist();
    const b = visB().find((x) => x.id === st.build);
    if (b) { drawEvents(b); drawAuras(b); }
  });

  // ---------- compare ----------
  // A side is a preset, or one of its race and setup variants from the Races tab. A variant only has its
  // mean DPS, because we sim the spread, fight lengths, stat weights and damage sources for the preset's own
  // race only.
  const CMP_SETUPS = [['baseline', 'Normal setup'], ['beast', 'Beast target'], ['elemental', 'Elemental target'], ['humanoid', 'Humanoid target']];
  const cmpSide = (k, b, rv) => {
    const mine = rv ? rv.filter((r) => r.build === b.id) : [];
    let race = st[k + 'r'] || b.race, v = st[k + 'v'] || 'baseline';
    if (!mine.some((r) => r.race === race)) race = b.race;
    const row = (vv) => mine.find((r) => r.race === race && r.variant === vv);
    // The racial weapon setup only exists for races with a weapon racial the preset can use. A setup can also
    // go missing from the data, like the Undead target that the generic boss replaced.
    if (!rv || !row(v) || !CMP_SETUPS.some((o) => o[0] === v) || (v === 'weapon' && !row('weapon')?.simmed)) v = 'baseline';
    const tag = [race !== b.race ? race : null, v !== 'baseline' ? CMP_SETUPS.find((o) => o[0] === v)[1].toLowerCase() : null].filter(Boolean).join(' · ');
    return { k, b, race, v, row: row(v), mine, full: race === b.race && v === 'baseline', name: tag ? `${b.build} · ${tag}` : b.build };
  };
  // Two classes take their class colors. Two presets of one class take a brighter and a darker shade of it.
  //
  // The darker shade stays at 52% lightness or more, so the shaman's blue (already fairly dark) stays readable
  // on the dark page. The brighter shade sits a fixed step above it, so the priest's white still gets a gray
  // to pair with.
  const cmpColors = (A, B) => {
    if (A.class !== B.class) return [ccol(A.class), ccol(B.class)];
    const c = ccol(A.class), dark = 'max(0.52, l - 0.2)';
    return [`oklch(from ${c} min(0.97, ${dark} + 0.3) c h)`, `oklch(from ${c} ${dark} c h)`];
  };
  const fillCmpVariants = (s) => {
    const lk = s.k.toLowerCase();
    const base = s.mine.filter((r) => r.variant === 'baseline').sort((p, q) => q.dps - p.dps);
    const races = base.length ? base : [{ race: s.b.race, raceIcon: s.b.raceIcon }];
    iconSelect($(`#cmp-${lk}r`), races.map((r) => ({ value: r.race, text: r.race + (r.race === s.b.race ? ' (preset)' : ''), icon: () => raceIco(r.raceIcon), end: r.dps ? f1(r.dps) : '' })), s.race,
      (v) => { st[s.k + 'r'] = v; persist(); render(); });
    iconSelect($(`#cmp-${lk}v`), CMP_SETUPS.map(([v, lbl]) => {
      const r = s.mine.find((x) => x.race === s.race && x.variant === v);
      const off = !s.mine.length || (v === 'weapon' && !r?.simmed) || (v !== 'baseline' && !r);
      return { value: v, text: off && r?.note ? `${lbl} (${r.note})` : lbl, end: r && !off ? f1(r.dps) : '', disabled: off && v !== 'baseline' };
    }), s.v, (v) => { st[s.k + 'v'] = v; persist(); render(); });
  };
  // The preset picker groups the presets by class, best first, like the page's other preset pickers.
  const cmpPresetPicker = (k, builds) => {
    const items = CLASSES.flatMap((c) => {
      const bs = builds.filter((b) => b.class === c).sort((a, b) => b.dps - a.dps);
      return bs.length ? [{ header: c, color: ccol(c) }, ...bs.map((b) => ({ value: b.id, text: b.build, label: b.label, icon: () => specIco(b, ''), end: f1(b.dps) }))] : [];
    });
    // A new preset starts on its own race and setup.
    iconSelect($(`#cmp-${k.toLowerCase()}`), items, st[k], (v) => { st[k] = v; st[k + 'r'] = null; st[k + 'v'] = null; persist(); render(); }, (it) => `${it.label} · ${it.end}`);
  };
  const drawCompare = () => {
    if (statusOf('builds') !== 'ok') return;
    const builds = visB();
    const ord = buildOrder(builds);
    if (!st.A || !builds.find((b) => b.id === st.A)) st.A = ord[0].id;
    if (!st.B || !builds.find((b) => b.id === st.B)) st.B = [...builds].sort((a, b) => b.dps - a.dps).find((b) => b.id !== st.A).id;
    cmpPresetPicker('A', builds); cmpPresetPicker('B', builds);
    const A = builds.find((b) => b.id === st.A), B = builds.find((b) => b.id === st.B);
    const rv = rowsOf('race_variants');
    const sA = cmpSide('A', A, rv), sB = cmpSide('B', B, rv);
    fillCmpVariants(sA); fillCmpVariants(sB);
    const [ca, cb] = cmpColors(A, B);
    const colOf = (s) => (s.k === 'A' ? ca : cb);
    // Every widget on the tab marks the two sides with the same lettered badge, A and B, in their chart colors.
    const abBadge = (k) => el('span', { class: 'fv-ab', style: `--ab:${k === 'A' ? ca : cb}`, text: k });
    $('#cmp-dot-a').style.setProperty('--ab', ca); $('#cmp-dot-b').style.setProperty('--ab', cb);
    $('#cmp-ha').replaceChildren(el('span', { class: 'fv-ab-head' }, abBadge('A'), specIco(A, '') || '', sA.name));
    $('#cmp-hb').replaceChildren(el('span', { class: 'fv-ab-head' }, abBadge('B'), specIco(B, '') || '', sB.name));
    $('#cmp-hd').replaceChildren(el('span', { class: 'fv-ab-head' }, abBadge('B'), 'vs', abBadge('A')));

    // With a variant on either side, both sides read their mean DPS from the race sims, so they share one setup.
    const anyVar = !sA.full || !sB.full;
    const W = (s) => `build=${enc(s.b.id)}&race=${enc(s.race)}&variant=${s.v}`;
    const pr = asmOn ? Object.fromEntries(['dps', 'p10', 'p50', 'p90', 'stdev', 'dps120', 'dps300'].map((k) => [k, +A[k] ? Math.round((B[k] / A[k] - 1) * 1000) / 10 : null]))
      : (rowsOf('pairs') || []).find((p) => p.a === A.id && p.b === B.id);
    const pct = (a, b) => (a ? Math.round((b / a - 1) * 10000) / 100 : null);
    const diffTd = (v, f, src) => (v == null ? el('td', { class: 'n', text: '' }) : el('td', { class: 'n' + (Math.abs(v) < 0.05 ? '' : v > 0 ? ' pos' : ' neg') }, src ? mark(el('span', { text: sgn(v, f) }), ...src) : el('span', { text: sgn(v, f) }), '%'));
    const NA = '–';
    const valTd = (s, k, f) => {
      if (k === 'dps' && anyVar) return s.row ? mark(el('td', { class: 'n', text: f1(s.row.dps) }), 'race_variants', 'dps', W(s)) : el('td', { class: 'n', text: NA });
      if (k === 'se') return anyVar && s.row ? mark(el('td', { class: 'n', text: `± ${f2(s.row.se)}` }), 'race_variants', 'se', W(s)) : mark(el('td', { class: 'n', text: `± ${f2(s.b.se)}` }), 'builds', 'se', 'id=' + enc(s.b.id));
      if (k === 'race') return el('td', { class: 'n', text: s.race });
      if (k === 'rank' && s.full) return mark(el('td', { class: 'n', text: String(rankOf(s.b, builds)) }), 'builds', st.enc ? 'rankAll' : 'rank', 'id=' + enc(s.b.id));
      if (k === 'bossType') return mark(el('td', { class: 'n', text: s.b.bossType || 'Generic boss' }), 'builds', 'bossType', 'id=' + enc(s.b.id));
      if (k === 'setup') return el('td', { class: 'n', text: CMP_SETUPS.find((o) => o[0] === s.v)[1] });
      if (k === 'raceRank') return s.row ? mark(el('td', { class: 'n', text: `${s.row.rank} of ${s.row.of}` }), 'race_variants', 'rank', W(s)) : el('td', { class: 'n', text: NA });
      if (k === 'bestRace') { const top = s.mine.find((r) => r.variant === s.v && r.rank === 1); return el('td', { class: 'n', text: top ? top.race : s.b.bestRace }); }
      if (!s.full) return el('td', { class: 'n', text: NA });
      return mark(el('td', { class: 'n', text: f(s.b[k]) }), 'builds', k, 'id=' + enc(s.b.id));
    };
    const diffOf = (k) => {
      if (k === 'dps' && anyVar) return diffTd(sA.row && sB.row ? pct(sA.row.dps, sB.row.dps) : null, f2);
      if (anyVar || !pr || pr[k] == null) return diffTd(null);
      const td = diffTd(pr[k], f1, ['pairs', k, `a=${enc(A.id)}&b=${enc(B.id)}`]);
      // A wider spread is neither good nor bad, so it gets no color.
      if (k === 'stdev') td.className = 'n';
      return td;
    };
    const rows = [['Mean DPS', 'dps', f1], ['Error of the mean', 'se', f2], ['Median', 'p50', f1], ['Unlucky (10th)', 'p10', f1], ['Lucky (90th)', 'p90', f1], ['Std dev', 'stdev', f1], ['120s fight', 'dps120', f1], ['300s fight', 'dps300', f1], ['Race', 'race', String],
      ...(anyVar ? [['Setup', 'setup', String]] : []), ...(rv ? [['Race rank', 'raceRank', String]] : []), ['Best race', 'bestRace', String], ['Boss', 'bossType', String], ['Rank', 'rank', String]];
    $('#cmp-table tbody').replaceChildren(...rows.map(([lbl, k, f]) => el('tr', {}, el('td', { text: lbl }), valTd(sA, k, f), valTd(sB, k, f), diffOf(k))));

    // The charts and the side by side cards show the sides that were simmed in full.
    const full = [sA, sB].filter((s) => s.full), part = [sA, sB].filter((s) => !s.full);
    const gap = $('#cmp-gap');
    gap.replaceChildren(...(part.length ? [part.map((s) => `${s.k} (${s.name})`).join(' and '), `: mean DPS only. ${full.length ? `The cards below show ${full[0].k} only.` : 'The charts below need a preset on its own race and setup.'}`] : []));
    for (const id of ['#cmp-dist-card', '#cmp-len-card', '#cmp-w-card', '#cmp-abil-card']) $(id).hidden = !full.length;
    const hd = $('#cmp-dist');
    if (full.length && waitOr(hd, ['hist'])) {
      const h = rowsOf('hist');
      drawHist(hd, full.map((s) => ({ name: `${s.k} · ${s.name}`, color: colOf(s), op: 0.6, bins: binsFor(h, s.b.id) })), { marks: full.map((s) => ({ v: s.b.dps, label: s.k })) });
    }
    const hl = $('#cmp-len');
    if (full.length && waitOr(hl, ['lengths'])) {
      const l = rowsOf('lengths');
      const ser = (s) => ({ name: `${s.k} · ${s.name}`, short: s.k, color: colOf(s), dots: true, pts: l.filter((r) => r.build === s.b.id).sort((p, q) => p.secs - q.secs).map((p) => ({ s: p.secs, v: p.dps })) });
      drawLenChart(hl, full.map(ser), { endLabels: true, h: 220 });
    }
    const hw = $('#cmp-w');
    if (full.length && waitOr(hw, ['weights'])) {
      const wr = rowsOf('weights');
      hw.replaceChildren(...full.map((s) => {
        const rs = wr.filter((w) => w.build === s.b.id && w.stat !== 'MeleeHitDown');
        const v = (w) => (w.unit === '1%' ? w.ratingWorth : w.worth);
        rs.sort((x, y) => v(y) - v(x));
        const max = d3.max(rs, (w) => Math.abs(v(w))) || 1;
        return el('div', { style: 'display:grid;gap:0.4rem;min-width:0;align-content:start' }, el('span', { class: 'fv-ab-head' }, abBadge(s.k), s.b.label),
          el('div', { class: 'fv-sbar', 'data-tt-b': s.b.id }, ...rs.flatMap((w) => [tt(el('span', { class: 'nm', text: w.unit === '1%' ? w.label.replace(' (1%)', ' rating') : w.label }), 'w:' + w.stat),
            tt(el('span', { class: 'vl' }, mark(el('span', { text: f2(v(w)) }), 'weights', w.unit === '1%' ? 'ratingWorth' : 'worth', `build=${enc(s.b.id)}&stat=${enc(w.stat)}`), ' ' + w.primaryAbbr), 'w:' + w.stat),
            tt(el('span', { class: 'tr' }, el('span', { style: `width:${Math.max(0, 100 * v(w) / max)}%;background:${colOf(s)}` })), 'w:' + w.stat)])));
      }));
    }
    const ha = $('#cmp-abil');
    if (full.length && waitOr(ha, ['abilities'])) {
      const ab = rowsOf('abilities');
      ha.replaceChildren(...full.map((s) => {
        const rows2 = ab.filter((a) => a.build === s.b.id && a.dps > 0).slice(0, 10);
        const max = d3.max(rows2, (a) => a.share) || 1;
        return el('div', { style: 'display:grid;gap:0.4rem;min-width:0;align-content:start' }, el('span', { class: 'fv-ab-head' }, abBadge(s.k), s.b.label),
          el('div', { class: 'fv-sbar', 'data-tt-b': s.b.id }, ...rows2.flatMap((a) => [tt(el('span', { class: 'nm', text: a.ability }), 'k:' + a.key), tt(el('span', { class: 'vl' }, mark(el('span', { text: f1(a.share) }), 'abilities', 'share', `build=${enc(s.b.id)}&key=${enc(a.key)}`), '%'), 'k:' + a.key), tt(el('span', { class: 'tr' }, el('span', { style: `width:${100 * a.share / max}%;background:${colOf(s)}` })), 'k:' + a.key)])));
      }));
    }
  };
  // ---------- method ----------
  const drawMethod = () => {
    const gh = $('#method-gear');
    if (statusOf('notes') === 'ok') gh.replaceChildren(...rowsOf('notes').filter((n) => n.build === '_round' && n.kind === 'gear').sort((a, b) => a.order - b.order).map((n) => mark(el('li', { text: n.text }), 'notes', 'text', `build=_round&kind=gear&order=${n.order}`)));
    const host = $('#method-caveats');
    if (!waitOr(host, ['notes', 'builds'])) return;
    const builds = new Map(visB().map((b) => [b.id, b]));
    const cav = rowsOf('notes').filter((n) => n.kind === 'caveats' && builds.has(n.build));
    const byClass = d3.groups(cav, (n) => builds.get(n.build).class).sort((a, b) => CLASSES.indexOf(a[0]) - CLASSES.indexOf(b[0]));
    host.replaceChildren(...byClass.map(([cls, ns]) => {
      const groups = cavGroups(cls, ns, builds);
      return el('details', { class: 'fv-cav-cls' }, el('summary', {}, el('span', { class: 'fv-cav-cls-name' }, el('span', { class: 'fv-dot', style: `background:${ccol(cls)}` }),
        `${cls} (${new Set(ns.map((n) => n.text)).size})`)), groups);
    }));
  };
  // A caveat that starts with "Assumed:", "Open:", "Left out:" or "Unverified:" shows that word as a tag.
  const CAV_TAG = /^(Assumed|Open|Left out|Unverified):\s*/;
  const cavItem = (t) => {
    const m = t.match(CAV_TAG), rest = m ? t.slice(m[0].length) : t, txt = rest.charAt(0).toUpperCase() + rest.slice(1);
    return el('li', { class: 'fv-cav' }, m ? el('span', { class: 'fv-cav-tag', text: m[1] }) : null, /[.!?)]$/.test(txt) ? txt : txt + '.');
  };
  // One class's caveats (ns), under the presets they apply to. The ones every preset of the class shares come first,
  // then each preset's own, in the ranking's order. A caveat a few presets share goes under all their names. Method
  // shows every class, and a preset's page only the groups with its preset in them (only).
  const cavGroups = (cls, ns, builds, only) => {
    const order = new Map([...builds.keys()].map((id, k) => [id, k]));
    const seen = new Map();
    for (const n of ns) { if (!seen.has(n.text)) seen.set(n.text, []); seen.get(n.text).push(n.build); }
    const all = [...builds.values()].filter((b) => b.class === cls).length;
    const groups = d3.groups([...seen], ([, ids]) => ids.join('|'))
      .map(([, items]) => ({ ids: items[0][1], texts: items.map(([t]) => t) }))
      .filter((g) => !only || g.ids.includes(only))
      .sort((p, q) => (q.ids.length === all) - (p.ids.length === all) || order.get(p.ids[0]) - order.get(q.ids[0]));
    const head = (ids) => el('h4', { class: 'fv-cav-preset' }, ...(ids.length === all && all > 1 ? [`All ${cls} presets`]
      : ids.flatMap((id, k) => [k ? ', ' : null, specIco(builds.get(id), ''), ' ', builds.get(id).build])));
    return el('div', { class: 'fv-cav-groups' }, ...groups.map((g) => el('section', { class: 'fv-cav-group' }, head(g.ids), el('ul', { class: 'fv-cav-list' }, ...g.texts.map(cavItem)))));
  };

  // ---------- static placeholders (headline figures, method text) ----------
  const fillStatic = () => {
    document.querySelectorAll('.dps-report-sim-ui [data-source][data-field][data-row]').forEach((n) => {
      const src = n.getAttribute('data-source');
      const r = dash.data(src);
      if (r.status !== 'ok') {
        if (r.status !== 'loading') n.classList.remove('dash-skeleton');
        return;
      }
      const row = r.data[+n.getAttribute('data-row')];
      const v = row ? row[n.getAttribute('data-field')] : null;
      n.textContent = v == null ? '—' : typeof v === 'number' ? (Number.isInteger(v) ? String(v) : f1(v)) : String(v);
      n.classList.remove('dash-skeleton');
    });
  };

  // ---------- assumptions panel ----------
  // the picks live in st, this browser's saved view
  const SEG_NAME = { low: 'Low', mid: 'Mid', expected: 'Expected', high: 'High', higher: 'Higher' };
  const asmPick = (a, id) => {
    st.asm = { ...(st.asm || {}) };
    if (id === a.defaultSetting) delete st.asm[a.key]; else st.asm[a.key] = id;
    persist(); hideTip(); render();
  };
  $('#asm-reset').addEventListener('click', (e) => { e.stopPropagation(); st.asm = {}; persist(); hideTip(); render(); });
  const bName = (id) => (rowsOf('builds') || []).find((b) => b.id === id)?.build || id;
  // Effects are each setting's change against the presets' own setting at 180s, from `sens`. A signed number takes
  // the ok or bad colour; a range over several presets runs from the smaller change to the larger.
  const SPEC_SHORT = { Retribution: 'Ret', Enhancement: 'Enh', Elemental: 'Ele', 'Beast Mastery': 'BM', Marksmanship: 'MM', Assassination: 'Assassin', Destruction: 'Destro', Demonology: 'Demo' };
  const asmSigned = (v, f, unit, eps) => el('span', { class: 'fv-asm-d' + (Math.abs(v) < eps ? '' : v > 0 ? ' pos' : ' neg'), text: (Math.abs(v) < eps ? '' : v > 0 ? '+' : '−') + f(Math.abs(v)) + unit });
  const asmRange = (xs, f = f1, unit = '%', eps = 0.05) => {
    if (!xs.length) return null;
    const lo = d3.min(xs), hi = d3.max(xs);
    if (hi - lo < eps || (Math.sign(lo) === Math.sign(hi) && f(Math.abs(lo)) === f(Math.abs(hi)))) return el('span', { class: 'fv-asm-eff' }, asmSigned((lo + hi) / 2, f, unit, eps));
    const [p, q] = lo >= 0 ? [lo, hi] : hi <= 0 ? [hi, lo] : [lo, hi];
    return el('span', { class: 'fv-asm-eff' }, asmSigned(p, f, '', eps), el('span', { class: 'fv-asm-to', text: ' to ' }), asmSigned(q, f, unit, eps));
  };
  const asmPcts = (a, sid, ids) => arr(a.sens).filter((x) => x.setting === sid && (!ids || ids.includes(x.build))).map((x) => +x.pct);
  // how much an assumption can move a preset at 180s, for the order within a class and the size badge
  const asmImpact = (a) => (a.status === 'live' ? d3.max(arr(a.sens), (x) => Math.abs(+x.pct)) || 0 : -1);
  const asmSize = (v) => (v >= 4 ? ['big', 'Big'] : v >= 1.5 ? ['med', 'Medium'] : v >= 0.05 ? ['small', 'Small'] : ['none', 'No effect']);
  const asmSort = (as) => [...as].sort((p, q) => asmImpact(q) - asmImpact(p) || p.order - q.order);
  // The presets an assumption touches, by spec. A spec covers all its presets, the vs Undead / Humanoid / Dragonkin
  // copies included, so the page names the spec only.
  const asmSpecs = (a) => {
    const byId = new Map((rowsOf('builds') || []).map((b) => [b.id, b])), groups = new Map();
    for (const id of arr(a.builds)) {
      const b = byId.get(id);
      if (!b) continue;
      const k = `${b.class}|${b.spec}`;
      if (!groups.has(k)) groups.set(k, { cls: b.class, spec: b.spec, icon: b.specIcon, ids: [], names: [] });
      const g = groups.get(k);
      g.ids.push(id);
      if (!b.encounter) g.names.push(b.build);
    }
    return [...groups.values()].map((g) => ({ ...g, n: g.names.length || g.ids.length, short: SPEC_SHORT[g.spec] || g.spec }));
  };
  // A reference row: the only value considered, which the presets use, so there is nothing to compare or pick.
  const asmRef = (a) => arr(a.settings).length === 1;
  const REF_CAP = 'Reference value.';
  const asmRefLine = (a) => arr(a.settings).map((x) => {
    const p = el('p', { class: 'fv-asm-ref' }, el('b', { text: x.label }), ...asmMarks(a, x),
      x.value && x.value !== x.label ? el('span', { text: ` · ${x.value}` }) : null);
    if (x.rationale) asmHover(p, x.label, () => asmSetTip(a, x));
    return p;
  });
  // The presets' setting is the default: what the ranking uses, and the expected value since the 2026-10-15 rulings.
  const asmMarks = (a, x) => [x.id === a.defaultSetting ? el('span', { class: 'def', text: ' · default' }) : null];
  // One table per assumption. Each cell answers "if the truth is this column, this row's spec changes by this much at
  // 180s, against what the ranking shows now": a column per setting, a row per spec, the % in the ok or bad colour
  // and the DPS under it. The presets' column is the zero point ("–") whichever column is picked, so the numbers never
  // shift. The column heads are the buttons that apply a setting.
  const ASM_CAP = 'DPS change at 180s';
  const asmHead = (a, x, cur) => [
    el('b', { text: x.label }),
    x.id === a.defaultSetting ? el('small', { class: 'fv-asm-tag def', text: 'Default' }) : el('small', { class: 'fv-asm-tag', text: SEG_NAME[x.id] || x.id }),
    cur && x.id === cur && x.id !== a.defaultSetting ? el('small', { class: 'fv-asm-use now', text: 'now applied' }) : null,
  ];
  const asmMatrix = (a, cur, onPick) => {
    const live = a.status === 'live' && arr(a.sens).length > 0, sets = arr(a.settings);
    const dpsOf = (sid, ids) => arr(a.sens).filter((x) => x.setting === sid && ids.includes(x.build)).map((x) => +x.dps);
    const zero = () => el('span', { class: 'fv-asm-zero', title: 'Default: no change', text: '–' });
    const col = (x) => [x.id === a.defaultSetting ? 'is-def' : '', x.id === cur ? 'is-cur' : ''].filter(Boolean).join(' ') || null;
    const head = el('tr', {}, el('th', { class: 'fv-asm-corner', scope: 'col' }, live ? 'Spec' : ''), ...sets.map((x) => {
      const bt = el('button', { class: 'fv-asm-col', 'data-setting': x.id, 'aria-pressed': String(x.id === cur),
        'aria-label': `${x.label} (${SEG_NAME[x.id] || x.id})${x.id === a.defaultSetting ? ', default' : ''}` }, ...asmHead(a, x, cur));
      if (!live) bt.disabled = true;
      bt.addEventListener('click', (e) => { e.stopPropagation(); onPick(x.id); });
      // A tap on the head picks the setting, so its tooltip is for the mouse only.
      if (x.rationale) asmHover(bt, x.label, () => asmSetTip(a, x), { tap: false });
      return el('th', { scope: 'col', class: col(x) }, bt);
    }));
    const body = live ? el('tbody', {}, ...asmSpecs(a).map((g) => el('tr', {},
      el('th', { scope: 'row', title: `${g.cls} · ${g.spec}` }, el('span', { class: 'fv-asm-tspec' }, ringed(g.icon, 'fv-ico ring', ccol(g.cls)), el('span', { text: g.short }))),
      ...sets.map((x) => (x.id === a.defaultSetting ? el('td', { class: col(x) }, zero())
        : el('td', { class: col(x) }, asmRange(asmPcts(a, x.id, g.ids)) || zero(), el('small', {}, asmRange(dpsOf(x.id, g.ids), f0, ' DPS', 0.5)))))))) : null;
    return el('div', { class: 'fv-asm-mx' }, el('table', { class: 'fv-asm-tbl' }, el('thead', {}, head), body));
  };
  // Pointing at an assumption shows where it comes from, and its full description when the panel shows a short one.
  // Each setting explains itself on its column head (or its line in a reference row), see asmSetTip.
  const asmTip = (a) => [
    a.short && a.description && a.description !== a.short ? tipText(a.description) : null,
    a.source ? tipNote(`Source: ${a.source}`) : null,
    a.status === 'pending' ? tipNote('Not simulated yet') : a.status === 'inert' ? tipNote('No longer applies') : null,
  ];
  // What a setting means, and why it's a candidate.
  const asmSetTip = (a, x) => [
    el('div', { class: 'fv-tip-set' }, el('span', { class: 'fv-tip-m', text: SEG_NAME[x.id] || '' }), ...asmMarks(a, x),
      el('small', { text: `${x.value && x.value !== x.label ? x.value + '. ' : ''}${x.rationale}` })),
  ];
  // Shows the tooltip under node, from node's left edge.
  const asmHover = (node, head, lines, opts = {}) => {
    node.setAttribute('data-tip', '1');
    hoverTip(node, (pinned) => {
      const card = $('#asm-card'), cr = card.getBoundingClientRect(), nr = node.getBoundingClientRect();
      showTip(card, nr.left - cr.left, nr.bottom - cr.top + 4, head, lines(), { hl: node, below: true, pinned, act: opts.act });
    }, { tap: opts.tap !== false });
  };
  // Everything about an assumption the row leaves out, behind "More details": the full description, each setting's
  // value and reason, and the source.
  const asmMore = (a) => el('details', { class: 'fv-asm-more' }, el('summary', { text: 'More details' }),
    a.description && a.description !== (a.short || a.description) ? el('p', { text: a.description }) : null,
    el('div', { class: 'fv-asm-msets' }, ...arr(a.settings).map((x) => el('small', {}, el('b', { text: x.label }),
      ` (${x.id === a.defaultSetting ? 'default' : SEG_NAME[x.id] || x.id}): ${x.value && x.value !== x.label ? x.value + '. ' : ''}${x.rationale || ''}`))),
    a.source ? el('small', { class: 'fv-asm-src', text: `Source: ${a.source}` }) : null);
  const drawAsm = () => {
    const host = $('#asm-list'), custom = asmSel().length > 0;
    $('#asm-badge').hidden = !custom; $('#asm-reset').hidden = !custom;
    if (!waitOr(host, ['assumptions', 'assumption_dps', 'builds'])) return;
    const As = rowsOf('assumptions'), open = new Set(arr(st.asmOpen));
    host.replaceChildren(...d3.groups(As, (a) => a.class).map(([cls, as]) => {
      const changed = as.filter((a) => st.asm?.[a.key]).length, top = d3.max(as, asmImpact);
      const d = el('details', { class: 'fv-asm-cls', 'data-cls': cls });
      d.open = open.has(cls);
      d.addEventListener('toggle', () => { const o = new Set(arr(st.asmOpen)); if (d.open) o.add(cls); else o.delete(cls); st.asmOpen = [...o]; persist(); });
      d.append(el('summary', {}, cls === 'All' ? null : ringed('classicon_' + cls.toLowerCase(), 'fv-ico ring', ccol(cls)), cls === 'All' ? 'All classes' : cls,
        el('small', { text: `${as.length} assumption${as.length === 1 ? '' : 's'}${top > 0 ? ` · up to ${f1(top)}%` : ''}` }), changed ? el('span', { class: 'fv-asm-badge', text: `${changed} changed` }) : null));
      for (const a of asmSort(as)) {
        const cur = st.asm?.[a.key] || a.defaultSetting, ref = asmRef(a), live = a.status === 'live' && !ref, imp = asmImpact(a);
        const [sz, szText] = ref ? ['ref', 'Reference'] : asmSize(imp);
        const picked = cur !== a.defaultSetting;
        const hd = el('div', { class: 'fv-asm-hd' },
          el('div', { class: 'fv-asm-lbl' }, el('b', { text: a.label }), live || ref ? el('span', { class: `fv-asm-size ${sz}`, title: ref ? 'One value only, kept for reference' : `Up to ${f1(imp)}% on a preset at 180s`, text: szText }) : null),
          el('p', { class: 'fv-asm-short', text: a.short || a.description }));
        asmHover(hd, a.label, () => asmTip(a));
        d.append(el('div', { class: 'fv-asm-row', 'data-key': a.key, 'data-size': sz },
          el('div', { class: 'fv-asm-txt' }, hd, asmMore(a)),
          el('div', { class: 'fv-asm-body' }, ...(ref ? [...asmRefLine(a), el('p', { class: 'fv-asm-cap', text: REF_CAP })] : [
            live ? el('p', { class: 'fv-asm-cap', text: picked ? 'DPS change at 180s, vs the ranking without your pick' : ASM_CAP }) : null,
            asmMatrix(a, cur, (id) => asmPick(a, id)),
            !live ? el('p', { class: 'fv-asm-cap', text: a.status === 'pending' ? 'Not simulated yet.' : 'No longer applies.' }) : null]))));
      }
      return d;
    }));
  };
  // headline figures: the overview calc, or the same figures from the adjusted presets
  const asmOverview = () => {
    const ns = document.querySelectorAll('#fv [data-asm-f]');
    if (!ns.length) document.querySelectorAll('#fv [data-source="overview"]').forEach((n) => { n.dataset.asmF = n.getAttribute('data-field'); });
    const all = document.querySelectorAll('#fv [data-asm-f]');
    if (!asmOn) { all.forEach((n) => { n.setAttribute('data-source', 'overview'); n.setAttribute('data-field', n.dataset.asmF); n.setAttribute('data-row', '0'); }); return; }
    const s = (genB() || []).map(adjB).sort((p, q) => q.dps - p.dps);
    if (!s.length) return;
    const top = s[0], bot = s[s.length - 1];
    const v = { topLabel: top.label, topDps: f1(top.dps), bottomLabel: bot.label, bottomDps: f1(bot.dps), median: f1(d3.median(s, (b) => b.dps)), count: String(s.length),
      classes: String(new Set(s.map((b) => b.class)).size), spreadPct: f1((top.dps / bot.dps - 1) * 100) };
    all.forEach((n) => { for (const a of ['data-source', 'data-field', 'data-row']) n.removeAttribute(a); n.textContent = v[n.dataset.asmF] ?? '—'; n.classList.remove('dash-skeleton'); });
  };

  // ---------- render ----------
  // What the first view doesn't need waits until it is drawn: tooltips, more icons and the tree art behind the
  // header. Then it loads while the reader looks at the page, so it's there by the first hover or tab change.
  let later = false;
  const loadLater = () => {
    if (later) return;
    later = wantMore = true;
    for (const id of ['tooltips', 'icons_more', 'treeart']) dash.data(id);
  };
  function render() {
    asmOn = asmSel().length > 0 && asmReady();
    asmOverview();
    fillStatic();
    if (!later && (st.tab !== 'ranking' || (statusOf('builds') === 'ok' && statusOf('icons') === 'ok'))) {
      if (st.tab !== 'ranking') loadLater();
      else (window.requestIdleCallback || setTimeout)(loadLater);
    }
    if (ttNode && !ttNode.isConnected) hideTT();
    try {
      if (st.tab === 'ranking') { safeTop(drawAsm, $('#asm-list')); drawRanking(); drawLenAll(); drawMana(); }
      else if (st.tab === 'races') { for (const [fn, h] of [[drawRaceRank, '#race-rank'], [drawRacialRank, '#racial-rank'], [drawRaceSpread, '#racespread']]) safeTop(fn, h && $(h)); }
      else if (st.tab === 'spec') drawSpec();
      else if (st.tab === 'compare') drawCompare();
      else if (st.tab === 'method') { drawMethod(); }
    } catch (err) { console.error(err); throw err; }
  }

  // deep links: a tab name or a preset id
  const h = location.hash.slice(1);
  if (pages.includes(h)) st.tab = h;
  else if (/^[a-z]+__[\w-]+$/.test(h)) { st.build = h; st.tab = 'spec'; }
  document.querySelectorAll('.fv-tab').forEach((b) => { b.setAttribute('aria-selected', String(b.dataset.tab === st.tab)); b.classList.toggle('active', b.dataset.tab === st.tab); });
  document.querySelectorAll('.fv-page').forEach((p) => { p.hidden = p.dataset.page !== st.tab; p.classList.toggle('active', p.dataset.page === st.tab); });
  dash.onData(render);
}

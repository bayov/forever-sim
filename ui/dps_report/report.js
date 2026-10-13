// The DPS Report's page logic: reads the datasets through `dash` (see data.ts) and draws every tab with d3.
// It began as the script of a Claude dashboard artifact, hence the dash API and data-source marks.
/* global d3 */
import tippy from 'tippy.js';

export function startReport(dash) {
  const CLASSES = ['Warrior', 'Rogue', 'Hunter', 'Paladin', 'Shaman', 'Druid', 'Mage', 'Warlock', 'Priest'];
  const RACES_ALL = ['Human', 'Dwarf', 'Night Elf', 'Gnome', 'High Order', 'Orc', 'Undead', 'Tauren', 'Troll', 'Windshaper'];
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
  const mark = (n, src, field, where) => {
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
  const visB = () => { const r = rowsOf('builds'); return r && r.filter((b) => st.enc || !isEnc(b)); };
  const genB = () => { const r = rowsOf('builds'); return r && r.filter((b) => !isEnc(b)); };
  const rankOf = (b, list) => 1 + list.filter((x) => +x.dps > +b.dps).length;

  // ---------- calculations ----------
  dash.calc('overview', {
    inputs: ['builds'], title: 'Headline figures',
    description: 'Top, median and lowest preset DPS, how many presets and classes, and the top-to-bottom gap in percent.',
    fn: (builds) => {
      const s = builds.filter((b) => !(b.encounter && b.encounter !== 'null')).sort((a, b) => b.dps - a.dps);
      const n = (v) => +v;
      const top = s[0], bot = s[s.length - 1];
      return [{
        topLabel: top.label, topDps: n(top.dps), bottomLabel: bot.label, bottomDps: n(bot.dps),
        median: Math.round(d3.median(s, (b) => b.dps) * 10) / 10, count: s.length,
        classes: new Set(s.map((b) => b.class)).size, spreadPct: Math.round((top.dps / bot.dps - 1) * 1000) / 10,
      }];
    },
  });
  dash.calc('factionbest', {
    inputs: ['race_variants', 'builds'], title: 'Best race by faction',
    description: 'Per preset and race scenario, the best Horde and best Alliance race with their DPS, and Horde over Alliance in percent.',
    fn: (race_variants, builds) => {
      const lab = new Map(builds.map((b) => [b.id, b.label]));
      return d3.groups(race_variants, (r) => r.build, (r) => r.variant).flatMap(([build, vs]) => vs.map(([variant, rs]) => {
        const top = (f) => rs.filter((r) => r.faction === f).reduce((x, r) => (!x || +r.dps > +x.dps ? r : x), null);
        const h = top('Horde'), al = top('Alliance');
        return { build, variant, label: lab.get(build), hordeRace: h ? h.race : null, hordeDps: h ? +h.dps : null, allianceRace: al ? al.race : null, allianceDps: al ? +al.dps : null,
          gapPct: h && al ? Math.round((h.dps / al.dps - 1) * 1000) / 10 : null };
      }));
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
    description: 'For every pair of presets, the second’s mean, 10th, 50th and 90th percentile, standard deviation and 120 s / 300 s DPS relative to the first, in percent.',
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
  const st = Object.assign({ tab: 'ranking', cls: 'All', heatCls: 'All', heatFac: 'All', role: 'All', fac: 'All', topN: 'all', defaults: false, best: false, sort: 'dps', len: '180', build: null, sample: 0, zoom: 'fit', A: null, B: null, Ar: null, Av: null, Br: null, Bv: null, alwaysOn: true, filter: 'all', hlLen: null, hlMana: null, distMode: 'bins', abilMetric: 'share', spreadSort: 'spread', auraOwner: 'All', lanes: 'all', focus: '', rsTarget: 'generic', rsWeapon: 'preset', rsFac: 'All', rsRole: 'All', rsBuild: null, setupHide: false, enc: false }, saved);
  st.rsTarget = 'generic'; st.rsWeapon = 'preset';
  st.len = '180'; // a fight length other than 180 s is never restored: the page always opens on the main numbers
  if (typeof st.focus !== 'string') st.focus = '';
  const persist = () => { try { store?.setItem('fv-state', JSON.stringify(st)); } catch {} };

  // ---------- tooltip ----------
  // Chart tooltips: tap (or hover with a mouse) a mark for its exact values; the mark is highlighted; tap elsewhere to close.
  let tipEl = null, tipTippy = null, tipHl = [], tipPinned = false, tipCleanup = null;
  const hideTip = () => {
    if (tipTippy) { tipTippy.destroy(); tipTippy = null; }
    tipEl = null;
    for (const n of tipHl) n.classList.remove('is-hl');
    if (tipCleanup) { tipCleanup(); tipCleanup = null; }
    tipHl = []; tipPinned = false;
  };
  // x and y are where the mark is in host, which the tooltip points at.
  const showTip = (host, x, y, head, lines, opts = {}) => {
    hideTip();
    try { hideTT(); } catch {}
    tipEl = el('div', { class: 'fv-tip' + (opts.open ? ' act' : ''), role: 'status' }, el('span', { class: 'h', text: head }),
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
      placement: opts.below ? 'bottom-start' : 'top-start', offset: [10, 8], getReferenceClientRect: at, duration: [120, 0],
    });
    tipTippy.show();
    tipHl = (Array.isArray(opts.hl) ? opts.hl : opts.hl ? [opts.hl] : []).filter(Boolean);
    for (const n of tipHl) n.classList.add('is-hl');
    tipCleanup = opts.onHide || null;
    tipPinned = opts.pinned !== false;
  };
  // bind a tooltip to d3 marks: fn(event, datum, node) shows it
  const tipOn = (sel, fn) => sel.attr('data-tip', '1').style('cursor', 'pointer')
    .on('click', function (e, d) { e.stopPropagation(); if (tipHl.includes(this) && tipPinned) { hideTip(); return; } fn(e, d, this); tipPinned = true; })
    .on('pointerenter', function (e, d) { if (e.pointerType === 'mouse' && !tipPinned) { fn(e, d, this); tipPinned = false; } })
    .on('pointerleave', function (e) { if (e.pointerType === 'mouse' && !tipPinned && tipHl.includes(this)) hideTip(); });
  document.addEventListener('click', (e) => { if (!e.target.closest('[data-tip]') && !e.target.closest('.fv-tip')) hideTip(); });

  const aBoss = (type) => `${/^[AEIOU]/.test(type) ? 'an' : 'a'} ${type} boss`;
  // An advantaged preset's "vs Undead" badge says on hover why the preset does better against that kind of boss.
  const encBadge = (b) => {
    const badge = el('span', { class: 'fv-enc', text: 'vs ' + b.encounter });
    if (b.advantage) tippy(badge, { content: el('div', { class: 'fv-advtip' }, el('b', { text: `Advantaged against ${aBoss(b.encounter)}` }), el('p', { text: b.advantage })) });
    return badge;
  };
  // Plain title tooltips become the sim UI's tooltips. We move each title into a tippy as it shows up, in the
  // layout or in anything we draw later.
  const tipTitles = (n) => {
    const t = n.getAttribute('title');
    n.removeAttribute('title');
    if (!t) return;
    const content = el('span', { text: t });
    if (n._tippy) n._tippy.setContent(content);
    else tippy(n, { content });
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
    if (!spec.lines.length && !spec.nums.length && aura) spec.lines = [['m', 'No tooltip text for this effect.']];
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
      title: cls, icon: 'classicon_' + cls.toLowerCase(), color: CCOL[cls] || '#fff', sub: `${bs.length} preset${bs.length === 1 ? '' : 's'}`,
      lines: bs.length ? [['m', 'Level 60 presets, mean DPS at 180 s'], ...bs.map((x) => ['2', x.build, f1(x.dps)])] : [],
    };
  };

  const weightSpec = (stat, b) => {
    const w = (rowsOf('weights') || []).find((x) => x.build === b && x.stat === stat);
    if (!w) return null;
    const s = ttRow('stat:' + stat);
    const pct = w.unit === '1%';
    const lines = [pct
      ? ['y', `1% more ${stat === 'MeleeHit' ? 'hit' : 'crit'} adds ${f2(w.dpsPer)} DPS. One rating point (${stat === 'MeleeHit' ? 10 : 14} make 1%) is worth ${w.ratingWorth == null ? '—' : f2(w.ratingWorth)} ${w.primaryAbbr}.`]
      : ['y', `One more point adds ${fmt('.3f')(w.dpsPer)} DPS, as much as ${w.worth == null ? '—' : f2(w.worth)} ${w.primaryAbbr}.`], ...arr(s?.lines)];
    return { title: pct ? String(w.label).replace(' (1%)', ' rating') : String(w.label), color: GOLD, sub: w.primaryAbbr ? `in ${w.primaryAbbr}` : '', lines,
      nums: nums([['DPS per ' + (pct ? '1%' : 'point'), `${fmt('.3f')(w.dpsPer)} ± ${fmt('.3f')(w.sePer)}`]]) };
  };

  const ttSpec = (node) => {
    const v = node.getAttribute('data-tt') || '';
    const i = v.indexOf(':');
    const type = v.slice(0, i), id = v.slice(i + 1);
    const b = node.closest('[data-tt-b]')?.getAttribute('data-tt-b') || st.build;
    const [p0, p1, p2] = id.split(':');
    switch (type) {
      case 'k': return spellSpec(id, b, false);
      case 'a': return spellSpec(id, b, true);
      case 'g': return gearSpec(id, b);
      case 'su': return setupSpec(id, b);
      case 'rot': return rotSpec(p0, p1, b);
      case 'tal': return talentSpec(p0, p1, p2, b);
      case 'r': return raceSpec(p0, p1, node.closest('[data-tt-b]') ? b : null);
      case 'c': return classSpec(id);
      case 'w': return weightSpec(id, b);
      default: return null;
    }
  };

  const ttLine = (l) => {
    if (!Array.isArray(l)) return null;
    const [k, a, b] = l;
    if (k === 's') return el('div', { class: 'tt-sp' });
    if (k === '2' || k === 'h' || k === 'n') {
      return el('div', { class: 'tt-row' + (k === 'h' ? ' tt-h' : '') }, el('span', { class: k === 'n' ? 'c-m' : null, text: String(a ?? '') }),
        el('span', { class: k === 'h' ? 'c-m' : 'c-t', text: String(b ?? '') }));
    }
    return el('div', { class: 'c-' + (['t', 'm', 'y', 'g'].includes(k) ? k : 't'), text: String(a ?? '') });
  };

  let ttEl = null, ttTippy = null, ttNode = null, ttPinned = false, ttAt = null;
  let ttHl = [], ttHideTimer = null;
  const hideTT = () => {
    clearTimeout(ttHideTimer);
    if (ttTippy) ttTippy.destroy();
    for (const n of ttHl) n.classList.remove('tt-on');
    ttHl = []; ttEl = null; ttTippy = null; ttNode = null; ttPinned = false;
  };
  const showTT = (node, pt) => {
    let spec = null;
    try { spec = ttSpec(node); } catch (err) { console.error(err); }
    if (!spec) return;
    hideTT(); hideTip();
    const body = el('div', { class: 'tt-body' }, spec.why ? el('div', { class: 'tt-why', text: spec.why }) : null,
      el('div', { class: 'tt-head' }, el('span', { class: 'tt-title', style: `color:${spec.color || GOLD}`, text: spec.title }), spec.sub ? el('span', { class: 'tt-sub', text: spec.sub }) : null),
      ...arr(spec.lines).map(ttLine),
      ...(spec.nums && spec.nums.length ? [el('div', { class: 'tt-nums' }, spec.numsHead ? el('div', { class: 'c-m', text: spec.numsHead }) : null, ...spec.nums.map(ttLine))] : []),
      spec.link ? el('div', { class: 'tt-link' }, el('a', { href: spec.link.href, target: '_blank', rel: 'noopener noreferrer', text: spec.link.label + ' ↗' }), spec.link.note ? el('span', { text: spec.link.note }) : null) : null);
    const content = el('div', { class: 'fv-tt', role: 'tooltip' }, spec.icon ? iconImg(spec.icon, 'tt-ico') : null, body);
    const r = node.getBoundingClientRect();
    // A wide target (a table row, a chart mark) gets the tooltip where the pointer was, not at its left edge.
    const wide = pt && (r.width > 160 || node instanceof SVGElement);
    const dx = wide ? pt.x - 12 - r.left : 0;
    ttTippy = tippy(node, {
      content, trigger: 'manual', interactive: true, appendTo: document.body, hideOnClick: false, placement: 'bottom-start', offset: [0, 6], duration: [120, 0],
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
    let hoverTimer = null, lpTimer = null, lpFired = false, lpStart = null;
    root.addEventListener('pointerover', (e) => {
      if (e.pointerType !== 'mouse') return;
      if (ttEl && ttEl.contains(e.target)) { clearTimeout(ttHideTimer); return; }
      const n = e.target.closest('[data-tt]');
      clearTimeout(hoverTimer);
      if (n && n === ttNode) clearTimeout(ttHideTimer);
      if (!n || n === ttNode) return;
      hoverTimer = setTimeout(() => { if (!ttPinned) showTT(n, { x: e.clientX, y: e.clientY }); }, 120);
    });
    root.addEventListener('pointerout', (e) => {
      if (e.pointerType !== 'mouse') return;
      const into = (x) => x && e.relatedTarget && x.contains(e.relatedTarget);
      // a short grace period lets the mouse cross over to the tooltip and its wowhead link
      if (ttEl && ttEl.contains(e.target)) { if (!ttPinned && !into(ttEl) && !into(ttNode)) { clearTimeout(ttHideTimer); ttHideTimer = setTimeout(hideTT, 250); } return; }
      const n = e.target.closest('[data-tt]');
      if (!n || into(n)) return;
      clearTimeout(hoverTimer);
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
  const pages = ['ranking', 'races', 'spec', 'compare', 'method'];
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
    const swap = st.tab === 'spec' && st.build !== id;
    st.build = id; st.sample = 0; st.focus = ''; persist(); showTab('spec'); $('.sim-ui').scrollTo({ top: 0 });
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
  const classChips = (host, key) => {
    const m = iconMap();
    const name = st[key] === 'All' ? 'All classes' : st[key];
    if (host.childElementCount && (host.dataset.icons === '1' || !m)) {
      host.querySelector('.fv-pick-value').textContent = name;
      host.querySelectorAll('button').forEach((b) => { b.setAttribute('aria-pressed', String(b.dataset.c === st[key])); b.classList.toggle('active', b.dataset.c === st[key]); });
      return;
    }
    host.dataset.icons = m ? '1' : '0';
    host.classList.add('fv-icon-picker');
    host.replaceChildren(el('span', { class: 'fv-label', text: 'Class' }), el('span', { class: 'fv-pick-value', text: name }),
      el('div', { class: 'fv-pick-options' }, ...CLASSES.map((c) => {
        const u = m && m.get('classicon_' + c.toLowerCase());
        const b = el('button', { type: 'button', class: 'fv-pick-option' + (c === st[key] ? ' active' : ''), 'data-c': c, 'aria-label': c, 'aria-pressed': String(c === st[key]), 'data-tt': 'c:' + c, style: u ? `background-image:url("${u}")` : `background:${ccol(c)}` });
        b.addEventListener('click', () => { st[key] = st[key] === c ? 'All' : c; persist(); render(); });
        return b;
      })));
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
  const roleOf = () => new Map((rowsOf('roles') || []).map((r) => [r.build, r.role]));
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
        : v >= b.p25 && v <= b.p75 ? ['Middle half of fights', `${f1(b.p25)} to ${f1(b.p75)}`]
          : v >= b.p10 && v <= b.p90 ? ['10th to 90th percentile', `${f1(b.p10)} to ${f1(b.p90)}`]
            : v <= b.dps ? ['Mean', f1(b.dps)] : null;
      if (!part) { hide(); return; }
      if (!partTippy) partTippy = tippy(el('span'), { trigger: 'manual', appendTo: document.body, placement: 'top', offset: [0, 12], duration: [80, 0], content: '' });
      if (last !== part[0]) { partTippy.setContent(el('div', { class: 'fv-parttip' }, el('b', { text: part[0] }), el('br'), `${part[1]} DPS`)); last = part[0]; }
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
    seg($('#rank-len'), 'len', lens.map((v) => [v, v + ' s']));
    $('#rank-lenbar').classList.toggle('off', st.len !== '180');
    seg($('#rank-sort'), 'sort', SORTS);
    seg($('#rank-role'), 'role', [['All', 'All roles'], ['Melee', 'Melee'], ['Ranged', 'Ranged'], ['Caster', 'Caster']]);
    const fi = (n) => (iconMap() ? iconImg(n, 'fv-ico') : null);
    seg($('#rank-fac'), 'fac', [['All', 'Both factions'], ['Horde', 'Horde', fi('ui_hordeicon')], ['Alliance', 'Alliance', fi('ui_allianceicon')]]);
    $('#rank-best').setAttribute('aria-pressed', String(st.best));
    $('#rank-enc').setAttribute('aria-pressed', String(!!st.enc));
    const filtered = st.cls !== 'All' || st.role !== 'All' || st.fac !== 'All' || st.best;
    $('#rank-reset').hidden = !filtered;
    if (!waitOr(host, ['builds', 'lengths'])) return;
    const builds = visB(), lengths = rowsOf('lengths'), roles = roleOf();
    const L = +st.len;
    $('#rank-sort').parentElement.hidden = L !== 180;
    const lenMap = new Map(lengths.filter((r) => r.secs === L).map((r) => [r.build, r]));
    let rows = builds.filter((b) => (st.cls === 'All' || b.class === st.cls) && (st.fac === 'All' || b.faction === st.fac) && (st.role === 'All' || roles.get(b.id) === st.role));
    const val = (b) => (L === 180 ? b.dps : lenMap.get(b.id)?.dps ?? 0);
    const sortKey = L === 180 ? st.sort : 'dps';
    const key = { dps: val, p10: (b) => b.p10, p90: (b) => b.p90, cv: (b) => -b.cv }[sortKey] || val;
    rows = rows.sort((a, b) => key(b) - key(a));
    if (st.best) { const seen = new Set(); rows = rows.filter((b) => !seen.has(b.class) && seen.add(b.class)); }
    const max = d3.max(builds, (b) => Math.max(L === 180 ? b.p90 : 0, val(b))) * 1.02;
    const x = (v) => (100 * v / max) + '%';
    const top = d3.max(builds, val);
    if (!rows.length) { host.replaceChildren(el('p', { class: 'fv-note', text: 'No preset matches these filters.' })); return; }
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
          el('span', { class: 'fv-row-name' }, el('b', { text: isEnc(b) ? String(b.build).replace(/ · vs .*$/, '') : b.build }), el('span', { text: '· ' + b.class })),
          el('span', { class: 'fv-row-sub' }, tt(raceIco(b.raceIcon), `r:${b.race}:${b.class}`, false, b.id), el('em', { text: b.race }),
            iconMap() ? iconImg(b.faction === 'Horde' ? 'ui_hordeicon' : 'ui_allianceicon', 'fv-ico') : null, roles.get(b.id) ? el('em', { text: '· ' + roles.get(b.id) }) : null,
            isEnc(b) ? encBadge(b) : null)),
        el('span', { class: 'fv-row-end' }, valEl, el('span', { class: 'fv-row-gap', text: extra })),
        track);
      row.addEventListener('click', (e) => {
        e.stopPropagation();
        if (row.classList.contains('is-hl')) { openSpec(b.id); return; }
        const card = $('#rank-card'), cr = card.getBoundingClientRect(), rr = row.getBoundingClientRect();
        const lines = L === 180
          ? [`${f1(b.dps)} DPS ± ${f2(b.se)} (1 s.e.)`, `Rank ${rankOf(b, builds)} of ${builds.length} · ${gap < 0.05 ? 'top preset' : `${f1(gap)}% behind the top`}`, `10th–90th percentile: ${f1(b.p10)}–${f1(b.p90)}`, `Middle half: ${f1(b.p25)}–${f1(b.p75)}`, `Fight-to-fight spread ${f2(b.cv)}%`, `${b.race} (${b.faction})${roles.get(b.id) ? ' · ' + roles.get(b.id) : ''}`]
          : [`${f1(v)} DPS in ${L} s fights${lenMap.get(b.id)?.se != null ? ` ± ${f2(lenMap.get(b.id).se)}` : ''}`, `${f1(b.dps)} DPS at 180 s (${sgn(100 * (v / b.dps - 1))}%)`, gap < 0.05 ? 'Top preset at this length' : `${f1(gap)}% behind the top at ${L} s`, `${b.race} (${b.faction})`];
        if (isEnc(b)) lines.push(`Advantaged: against ${aBoss(b.encounter)}`);
        showTip(card, rr.left - cr.left + 30, rr.bottom - cr.top + 6, b.label, lines, { hl: row, open: b.id, below: true });
      });
      return row;
    }));
  };
  $('#rank-best').addEventListener('click', (e) => { e.stopPropagation(); st.best = !st.best; persist(); render(); });
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
    const xTicks = opts.xTicks || xs, xFmt = opts.xFmt || ((d) => d + ' s');
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
    const pick = (e, pinned) => {
      const [px, py] = d3.pointer(e, svg.node());
      const sx = xs.reduce((a, b) => (Math.abs(x(b) - px) < Math.abs(x(a) - px) ? b : a));
      let best = null, bd = Infinity;
      for (const s of pickable) { const p = s.pts.find((q) => q.s === sx); if (p && Math.abs(y(p.v) - py) < bd) { bd = Math.abs(y(p.v) - py); best = [s, p]; } }
      if (!best) return;
      const [s, p] = best;
      rule.attr('x1', x(sx)).attr('x2', x(sx)).attr('opacity', 0.5);
      dot.attr('cx', x(sx)).attr('cy', y(p.v)).style('stroke', s.color).attr('opacity', 1);
      if (opts.focus) focusOn(s);
      if (opts.onHover) opts.onHover(s.id);
      const others = series.length <= 3 ? series.filter((q) => q !== s).map((q) => { const pp = q.pts.find((r) => r.s === sx); return pp ? el('span', {}, el('i', { class: 'sw', style: `background:${q.color}` }), `${q.name}: ${opts.tipFmt ? opts.tipFmt(pp) : f1(pp.v) + ' DPS'}`) : null; }) : [];
      const [hx, hy] = d3.pointer(e, host);
      // With many lines, a small table of the picked line at every length takes the place of the other lines' values.
      const strip = series.length > 3 && opts.strip !== false ? el('div', { class: 'fv-tip-strip' }, ...s.pts.flatMap((q) => [
        el('span', { class: 'k' + (q.s === sx ? ' on' : ''), text: `${q.s} s` }),
        el('span', { class: 'v' + (q.s === sx ? ' on' : ''), text: opts.fmt ? opts.fmt(q.v) : f0(q.v) }),
      ])) : null;
      const main = opts.tipMain ? opts.tipMain(p) : [opts.tipFmt ? opts.tipFmt(p) : `${f1(p.v)} DPS`];
      showTip(host, hx, hy, s.name, [el('span', { class: 'fv-tip-sub', text: opts.tipSub ? opts.tipSub(sx) : `${sx} s fights` }),
        el('span', { class: 'fv-tip-main' }, el('i', { class: 'sw', style: `background:${s.color}` }), el('b', { text: main[0] }), main[1] ? el('span', { class: 'fv-tip-sub', text: main[1] }) : null), ...others, strip],
        { open: s.id && opts.open ? s.id : null, pinned, onHide: () => { rule.attr('opacity', 0); dot.attr('opacity', 0); } });
      if (pinned && opts.onPick) opts.onPick(s.id);
    };
    if (opts.refLabel && opts.ref != null) svg.append('text').attr('x', W - m.r).attr('y', y(opts.ref) - 4).attr('text-anchor', 'end').style('fill', 'var(--color-fg-muted)').text(opts.refLabel);
    svg.append('rect').attr('x', m.l).attr('y', 0).attr('width', Math.max(0, W - m.l - m.r + 8)).attr('height', H - m.b).style('fill', 'transparent').attr('data-tip', '1').style('cursor', 'crosshair')
      .on('click', (e) => { e.stopPropagation(); pick(e, true); })
      .on('pointermove', (e) => { if (e.pointerType === 'mouse' && !tipPinned) pick(e, false); })
      .on('pointerleave', (e) => { if (e.pointerType === 'mouse' && !tipPinned) { hideTip(); rule.attr('opacity', 0); dot.attr('opacity', 0); highlight(null); if (opts.onHover) opts.onHover(null); } });
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
      return { id: b.id, name: b.label, color: ccol(b.class), op: on ? 0.9 : 0.08, pick: on, w: on && st.cls !== 'All' ? 2 : 1.5, dots: on && st.cls !== 'All', pts: pts.map((p) => ({ s: p.secs, v: 100 * p.dps / base, dps: p.dps })) };
    }).sort((a, b) => a.op - b.op);
    drawLenChart(host, series, { fmt: (v) => d3.format('.0f')(v) + '%', ref: 100, focus: true, open: true, hlId: st.hlLen, onPick: (id) => { st.hlLen = id; persist(); }, tipMain: (p) => [`${f1(p.dps)} DPS`, p.s === 180 ? null : `${sgn(p.v - 100)}% vs 180 s`] });
  };

  // ---------- mana ----------
  // How a preset's mana runs down over a long fight, from 100 logged 480 s fights each. The chart is the average mana
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
      : missing.length ? `No mana data yet for ${[...new Set(missing.map((b) => b.class))].join(', ').replace(/, ([^,]*)$/, ' and $1')} (${missing.length} preset${missing.length > 1 ? 's' : ''}).` : '';
    if (!shown.length) { host.replaceChildren(); list.replaceChildren(); return; }
    const lowPct = rowsOf('mana_summary')[0]?.lowPct ?? 15;
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
      tipSub: (t) => `at ${mmss(t)}`, tipMain: (p) => [`${d3.format('.0f')(p.v)}% mana`, `middle half of fights ${d3.format('.0f')(p.p25)} to ${d3.format('.0f')(p.p75)}%`],
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
        showTip(card, rr.left - cr.left + 30, rr.bottom - cr.top + 6, b.label, [
          low && m.lowAtP25 != null ? `Middle half of fights: ${mmss(m.lowAtP25)} to ${m.lowAtP75 != null ? mmss(m.lowAtP75) : 'never'}` : null,
          share >= 99.5 ? 'Every fight runs low' : `${d3.format('.0f')(share)}% of fights run low`,
          `${d3.format(',')(m.maxMana)} mana, drains ${f1(m.spentPerSec - m.gainedPerSec)}/s`,
        ], { hl: row, open: b.id, below: true, pinned });
      };
      row.addEventListener('pointerenter', (e) => { chart.highlight(b.id); row.classList.add('is-hover'); if (e.pointerType === 'mouse' && !tipPinned) tip(false); });
      row.addEventListener('pointerleave', (e) => { chart.highlight(null); row.classList.remove('is-hover'); if (e.pointerType === 'mouse' && !tipPinned) hideTip(); });
      row.addEventListener('click', (e) => {
        e.stopPropagation();
        if (row.classList.contains('is-hl') && tipPinned) { openSpec(b.id); return; }
        tip(true);
      });
      return row;
    }));
  };

  // ---------- races ----------
  const drawHeat = () => {
    const host = $('#heat');
    host.setAttribute('data-source', 'race_variants builds tooltips');
    classChips($('#heat-classes'), 'heatCls');
    const fi = (n) => (iconMap() ? iconImg(n, 'fv-ico') : null);
    seg($('#heat-fac'), 'heatFac', [['All', 'All'], ['Horde', 'Horde', fi('ui_hordeicon')], ['Alliance', 'Alliance', fi('ui_allianceicon')]]);
    if (!waitOr(host, ['race_variants', 'builds'])) return;
    const V = rsVar();
    const RACES = RACES_ALL.filter((r) => st.heatFac === 'All' || (st.heatFac === 'Horde') === HORDE.has(r));
    const races = rowsOf('race_variants').filter((r) => r.variant === V);
    const has = new Set(races.map((r) => r.build));
    const builds = buildOrder(genB()).filter((b) => (st.heatCls === 'All' || b.class === st.heatCls) && has.has(b.id));
    const idx = new Map(races.map((r) => [r.build + '|' + r.race, r]));
    const grid = el('div', { class: 'fv-heat', style: `grid-template-columns: minmax(6.5rem, 8.5rem) repeat(${RACES.length}, minmax(2.9rem, 1fr))` });
    grid.append(el('div', { class: 'hh' }));
    const RI = new Map(races.map((r) => [r.race, r.raceIcon]));
    for (const r of RACES) grid.append(tt(el('div', { class: 'hh', style: `color:${fcol(r)}` }, RI.get(r) ? iconImg(RI.get(r), '') : null, r.replace('High Order', 'High Ord.').replace('Windshaper', 'Windshp.')), 'r:' + r));
    const maxGap = d3.max(races, (r) => -r.vsBest) || 1;
    for (const b of builds) {
      const hl = el('div', { class: 'hl', title: b.label, role: 'button', tabindex: '0', style: 'cursor:pointer' }, specIco(b, '') || el('span', { class: 'fv-dot', style: `background:${ccol(b.class)}` }), b.build);
      hl.addEventListener('click', (e) => { e.stopPropagation(); pickRaceSpec(b.id); });
      grid.append(hl);
      for (const rc of RACES) {
        const r = idx.get(b.id + '|' + rc);
        if (!r) { grid.append(el('div', { class: 'hc na', text: '' })); continue; }
        const best = r.vsBest >= -0.005;
        const t = Math.min(1, -r.vsBest / Math.min(maxGap, 4));
        const c = el('button', {
          class: 'hc', 'data-tip': '1',
          style: `background:color-mix(in srgb, ${best ? 'var(--color-ok)' : 'var(--cds-chart-muted)'} ${best ? 35 : Math.round(8 + 50 * t)}%, transparent);`,
          text: best ? '★' : '−' + f1(-r.vsBest),
        });
        c.addEventListener('click', (e) => {
          e.stopPropagation();
          const card = $('#heat-card'), cr = card.getBoundingClientRect(), cb = c.getBoundingClientRect();
          if (c.classList.contains('is-hl')) { hideTip(); return; }
          const sib = races.filter((q) => q.build === b.id).sort((p, q) => q.dps - p.dps);
          showTip(card, cb.left - cr.left, cb.top - cr.top, `${b.label} · ${rc}`, [`${f1(r.dps)} DPS ± ${f2(r.se)}`, best ? 'Best race for this preset' : `${d3.format('.2f')(-r.vsBest)}% behind the best (${sib[0].race})`, `Race ${sib.indexOf(r) + 1} of ${sib.length} for this preset`, r.note ? `Same as with its normal weapon: ${r.note}` : '', r.isPreset ? 'The preset’s race' : ''], { hl: c, open: b.id, act: ['Compare races ›', () => pickRaceSpec(b.id)] });
        });
        grid.append(c);
      }
    }
    host.replaceChildren(grid);
  };

  const drawFactionTable = () => {
    const tb = $('#faction-table tbody');
    if (statusOf('builds') !== 'ok' || statusOf('factionbest') !== 'ok') return;
    const V = rsVar();
    const fb = new Map(rowsOf('factionbest').filter((g) => g.variant === V).map((g) => [g.build, g]));
    const ic = new Map(RACES_ALL.map((r) => [r, (rowsOf('race_variants') || []).find((x) => x.race === r)?.raceIcon]));
    if (iconMap()) { $('#fi-h').replaceChildren(iconImg('ui_hordeicon')); $('#fi-a').replaceChildren(iconImg('ui_allianceicon')); }
    tb.replaceChildren(...buildOrder(rowsOf('builds')).filter((b) => fb.has(b.id)).map((b) => {
      const g = fb.get(b.id);
      return el('tr', { class: 'clk', 'data-key': `${enc(b.id)},${enc(V)}`, onclick: () => pickRaceSpec(b.id) },
        el('td', { class: 'first' }, specIco(b, '') || el('span', { class: 'fv-dot', style: `background:${ccol(b.class)}` }), ' ' + g.label),
        tt(el('td', {}, raceIco(ic.get(g.hordeRace)), ' ' + (g.hordeRace || '—')), g.hordeRace ? `r:${g.hordeRace}:${b.class}` : null, false, b.id), el('td', { class: 'n', text: g.hordeDps ? f1(g.hordeDps) : '—' }),
        tt(el('td', {}, raceIco(ic.get(g.allianceRace)), ' ' + (g.allianceRace || '—')), g.allianceRace ? `r:${g.allianceRace}:${b.class}` : null, false, b.id), el('td', { class: 'n', text: g.allianceDps ? f1(g.allianceDps) : '—' }),
        el('td', { class: 'n', text: g.gapPct == null ? '' : f1(g.gapPct) }));
    }));
  };

  const drawRaceSpread = () => {
    const host = $('#racespread');
    if (!waitOr(host, ['racespread', 'builds', 'race_variants'])) return;
    const V = rsVar();
    const builds = new Map(rowsOf('builds').map((b) => [b.id, b]));
    seg($('#spread-sort'), 'spreadSort', [['spread', 'Biggest gap'], ['class', 'By class']]);
    const order = new Map(buildOrder(rowsOf('builds')).map((b, i) => [b.id, i]));
    const rows = rowsOf('racespread').filter((r) => r.variant === V && builds.has(r.build)).sort((a, b) => (st.spreadSort === 'class' ? order.get(a.build) - order.get(b.build) : b.spreadPct - a.spreadPct));
    const races = rowsOf('race_variants').filter((r) => r.variant === V);
    const max = d3.max(rows, (r) => r.spreadPct) || 1;
    const wrap = el('div', { class: 'fv-sbar' });
    for (const r of rows) {
      const b = builds.get(r.build);
      const v = el('span', { class: 'vl' }, mark(el('span', { text: f2(r.spreadPct) }), 'racespread', 'spreadPct', `build=${enc(r.build)}&variant=${V}`), '%');
      const parts = [el('span', { class: 'nm', 'data-tip': '1' }, specIco(b, '') || el('span', { class: 'fv-dot', style: `background:${ccol(b.class)}` }), `${b.label} `, el('small', { class: 'fv-note', text: `${r.best} over ${r.worst}` })), v,
        el('span', { class: 'tr', 'data-tip': '1' }, el('span', { style: `width:${100 * r.spreadPct / max}%;background:${ccol(b.class)}` }))];
      v.setAttribute('data-tip', '1');
      for (const n of parts) n.addEventListener('click', (e) => {
        e.stopPropagation();
        if (parts[0].classList.contains('is-hl')) { hideTip(); return; }
        const sib = races.filter((q) => q.build === b.id).sort((p, q) => q.dps - p.dps);
        const bb = sib[0], ww = sib[sib.length - 1], card = $('#racespread-card'), cr = card.getBoundingClientRect(), nr = parts[2].getBoundingClientRect();
        showTip(card, 20, nr.bottom - cr.top + 4, b.label, [`Best: ${r.best}${bb ? ` · ${f1(bb.dps)} DPS` : ''}`, `Worst: ${r.worst}${ww ? ` · ${f1(ww.dps)} DPS` : ''}`, `Gap ${f2(r.spreadPct)}% across ${sib.length} races`, `Preset race: ${b.race}`], { hl: parts, open: b.id, below: true, act: ['Compare races ›', () => pickRaceSpec(b.id)] });
      });
      wrap.append(...parts);
    }
    host.replaceChildren(wrap);
    host.style.minHeight = 'auto';
  };

  // ---------- races: scenarios ----------
  // Target and weapon are two views of one scenario: the weapon views are simmed against the generic boss only.
  const RS_T = [['generic', 'Generic boss'], ['beast', 'Beast'], ['elemental', 'Elemental'], ['humanoid', 'Humanoid']];
  const RS_W = [['preset', 'Normal weapon'], ['weapon', 'Racial weapon'], ['off', 'Racials off']];
  const rsVar = () => (st.rsTarget !== 'generic' ? st.rsTarget : st.rsWeapon === 'preset' ? 'baseline' : st.rsWeapon);
  const safeTop = (fn, host) => { try { fn(); } catch (err) { console.error(err); if (host) fail(host); } };
  // a racial's colour follows what kind it is: a passive, a cooldown, or the target type; base stats and the weapon swap are neutral
  const PART_KIND = { elune: 1, eureka: 1, bloodfury: 1, berserking: 1, chastise: 1, starshards: 1, darksacrifice: 1, target: 2 };
  const partCol = (p) => (p === 'base' ? 'var(--cds-chart-muted)' : p === 'swap' ? 'var(--color-fg-muted)' : dash.colors[PART_KIND[p] ?? 0]);
  const COND = { beast: /^beast/, elemental: /^insight$/ };
  const KIND_NAME = ['Passive racial', 'Racial cooldown', 'Target type'];
  const noteRow = (kind, key) => (rowsOf('race_notes') || []).find((n) => n.kind === kind && n.key === key);
  const pickRaceSpec = (id) => {
    st.rsBuild = id; persist(); hideTip();
    if (st.tab !== 'races') showTab('races'); else render();
    requestAnimationFrame(() => $('#race-spec-card').scrollIntoView({ behavior: matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth', block: 'start' }));
  };
  document.querySelectorAll('#races .fv-scen-tag').forEach((t) => t.addEventListener('click', (e) => { e.stopPropagation(); $('#race-scen-card').scrollIntoView({ behavior: 'smooth', block: 'start' }); }));

  const drawScen = () => {
    seg($('#rs-target'), 'rsTarget', RS_T, (v) => { if (v !== 'generic') st.rsWeapon = 'preset'; persist(); render(); });
    seg($('#rs-weapon'), 'rsWeapon', RS_W, (v) => { if (v !== 'preset') st.rsTarget = 'generic'; persist(); render(); });
    const V = rsVar(), d = $('#rs-desc');
    if (statusOf('race_notes') !== 'ok') { d.classList.add('dash-skeleton'); return; }
    d.classList.remove('dash-skeleton');
    const n = noteRow('scenario', V);
    mark(d, 'race_notes', 'text', `kind=scenario&key=${V}`).textContent = n ? n.text : '';
    document.querySelectorAll('#races .fv-scen-tag').forEach((t) => { mark(t, 'race_notes', 'title', `kind=scenario&key=${V}`).textContent = n ? n.title : ''; });
  };

  // ---------- races: overview ----------
  const drawRaceOverview = () => {
    const host = $('#race-overview'), tiles = $('#ro-tiles');
    const fi = (n) => (iconMap() ? iconImg(n, 'fv-ico') : null);
    seg($('#ro-fac'), 'rsFac', [['All', 'All races'], ['Horde', 'Horde', fi('ui_hordeicon')], ['Alliance', 'Alliance', fi('ui_allianceicon')]]);
    seg($('#ro-role'), 'rsRole', [['All', 'All roles'], ['Melee', 'Melee'], ['Ranged', 'Ranged'], ['Caster', 'Caster']]);
    if (!waitOr(host, ['race_summary'])) { tiles.replaceChildren(); return; }
    const V = rsVar(), scope = st.rsFac === 'All' ? 'All' : 'Faction';
    const all = rowsOf('race_summary').filter((r) => r.variant === V && r.scope === scope && (st.rsFac === 'All' || r.faction === st.rsFac));
    const W = (r) => `variant=${V}&scope=${scope}&role=${r.role}&race=${enc(r.race)}`;
    const better = (p, q) => q.timesBest - p.timesBest || p.avgGap - q.avgGap;
    // the best race for each role: most often best, then the smallest average gap
    tiles.replaceChildren(...['Melee', 'Ranged', 'Caster'].map((rl) => {
      const top = all.filter((r) => r.role === rl).sort(better)[0];
      if (!top) return el('div', { class: 'fv-tile' }, el('span', { class: 'k', text: rl }), el('span', { class: 'v', text: '—' }));
      return tt(el('div', { class: 'fv-tile' }, el('span', { class: 'k', text: `Best ${rl.toLowerCase()}` }),
        el('span', { class: 'v' }, raceIco(top.raceIcon), mark(el('span', { text: top.race }), 'race_summary', 'race', W(top))),
        el('span', { class: 's' }, mark(el('span', { text: String(top.timesBest) }), 'race_summary', 'timesBest', W(top)), ' of ', mark(el('span', { text: String(top.presets) }), 'race_summary', 'presets', W(top)), ' presets')), `r:${top.race}`);
    }));
    const rows = all.filter((r) => r.role === st.rsRole).sort((p, q) => p.avgGap - q.avgGap || q.timesBest - p.timesBest);
    if (!rows.length) { host.replaceChildren(el('p', { class: 'fv-note', text: 'No race plays this role in this faction.' })); return; }
    const max = Math.max(0.5, d3.max(rows, (r) => r.worstGap));
    const x = (v) => `${(100 * v) / max}%`;
    const card = $('#race-overview-card');
    host.replaceChildren(...rows.map((r, i) => {
      const row = el('div', { class: 'fv-orow' + (i === 0 ? ' top' : ''), 'data-tip': '1', role: 'button', tabindex: '0', 'aria-label': `${r.race}: best on ${r.timesBest} of ${r.presets}` },
        tt(raceIco(r.raceIcon) || el('span', { class: 'fv-dot', style: `background:${fcol(r.race)}` }), `r:${r.race}`),
        el('span', { class: 'nm' }, el('b', { text: r.race, style: `color:${fcol(r.race)}` }), el('small', { text: `${r.presets} preset${r.presets === 1 ? '' : 's'}` })),
        el('span', { class: 'end' }, el('b', {}, '★ ', mark(el('span', { text: String(r.timesBest) }), 'race_summary', 'timesBest', W(r))),
          el('small', {}, r.avgGap < 0.005 ? 'always best' : el('span', {}, '−', mark(el('span', { text: f2(r.avgGap) }), 'race_summary', 'avgGap', W(r)), '% avg'))),
        el('div', { class: 'fv-otrack' }, el('span', { class: 'bar', style: `width:${x(r.avgGap)};background:${fcol(r.race)}` }), el('span', { class: 'wk', style: `left:0;width:${x(r.worstGap)}` })));
      row.addEventListener('click', (e) => {
        if (e.target.closest('[data-tt]')) return;
        e.stopPropagation();
        if (row.classList.contains('is-hl')) { hideTip(); return; }
        const cr = card.getBoundingClientRect(), rr = row.getBoundingClientRect();
        const bf = arr(r.bestFor).map((lbl, k) => `${lbl}${+arr(r.bestForGap)[k] > 0.005 ? ` (−${f2(arr(r.bestForGap)[k])}%)` : ' (best)'}`);
        showTip(card, 24, rr.bottom - cr.top + 4, `${r.race} · ${st.rsRole === 'All' ? 'all roles' : st.rsRole.toLowerCase()}`,
          [`Best race on ${r.timesBest} of ${r.presets} presets${scope === 'Faction' ? ` (within the ${r.faction})` : ''}`, `${f2(r.avgGap)}% behind the best on average`, `Worst: ${f2(r.worstGap)}% behind`, `Median rank ${r.medianRank}`, 'Closest to the best on:', ...bf], { hl: row, below: true });
      });
      return row;
    }));
  };

  // ---------- races: one preset, every race ----------
  const drawRaceSpec = () => {
    const host = $('#race-spec');
    if (!waitOr(host, ['race_variants', 'race_breakdown', 'builds'])) return;
    const builds = genB();
    if (!builds.some((b) => b.id === st.rsBuild)) st.rsBuild = builds.some((b) => b.id === st.build) ? st.build : [...builds].sort((p, q) => q.dps - p.dps)[0].id;
    const pick = $('#rs-pick');
    sortedBuildOptions(pick, builds, st.rsBuild);
    if (!pick.dataset.bound) {
      pick.dataset.bound = '1';
      pick.addEventListener('change', () => { st.rsBuild = pick.value; persist(); hideTip(); render(); });
      const order = () => buildOrder(genB()).map((b) => b.id);
      const step = (d) => { const o = order(); st.rsBuild = o[(o.indexOf(st.rsBuild) + d + o.length) % o.length]; persist(); hideTip(); render(); };
      $('#rs-prev').addEventListener('click', (e) => { e.stopPropagation(); step(-1); });
      $('#rs-next').addEventListener('click', (e) => { e.stopPropagation(); step(1); });
    }
    const b = builds.find((x) => x.id === st.rsBuild), V = rsVar();
    const rows = rowsOf('race_variants').filter((r) => r.build === b.id && r.variant === V).sort((p, q) => q.dps - p.dps);
    if (!rows.length) { host.replaceChildren(el('p', { class: 'fv-note', text: V === 'humanoid' ? 'No racial of this class cares about Humanoids: only Priests (Dwarf Chastise) have this scenario.' : 'No race numbers for this preset.' })); $('#rs-take').textContent = ''; return; }
    const bd = rowsOf('race_breakdown').filter((r) => r.build === b.id);
    const segsOf = (race) => {
      const mine = bd.filter((r) => r.race === race);
      const base = mine.find((r) => r.part === 'base');
      const out = [{ part: 'base', r: base }];
      if (V !== 'off') {
        for (const r of mine.filter((q) => q.scenario === 'all' && q.part !== 'base' && q.active)) out.push({ part: r.part, r });
        const cond = mine.find((q) => q.scenario === 'all' && !q.active && COND[V] && COND[V].test(q.part));
        for (const r of mine.filter((q) => q.scenario === (V === 'baseline' ? '-' : V))) out.push({ part: r.part, r: cond && r.part === 'target' ? { ...r, label: `${r.label}, ${cond.label}` } : r });
      }
      let at = 0;
      for (const sg of out) { sg.from = at; at += +sg.r.dps; sg.to = at; }
      return { segs: out, idle: mine.filter((q) => q.scenario === 'all' && !q.active && !(COND[V] && COND[V].test(q.part)) && !(V === 'weapon' && mine.some((w) => w.scenario === 'weapon' && w.part === q.part))) };
    };
    const S = new Map(rows.map((r) => [r.race, segsOf(r.race)]));
    const ends = rows.flatMap((r) => S.get(r.race).segs.flatMap((sg) => [sg.from || sg.to, sg.to]).filter((v) => v > 0).concat([r.dps - r.se, r.dps + r.se]));
    const bases = rows.map((r) => S.get(r.race).segs[0].to);
    const hi = d3.max(ends), lo0 = Math.min(d3.min(ends), d3.min(bases));
    const lo = Math.max(0, lo0 - 0.35 * Math.max(hi - lo0, 4));
    const x = (v) => (100 * (Math.max(lo, v) - lo)) / (hi - lo);
    const card = $('#race-spec-card');
    const W = (r) => `build=${enc(b.id)}&race=${enc(r.race)}&variant=${V}`;
    const kinds = new Set();
    host.replaceChildren(...rows.map((r, i) => {
      const { segs, idle } = S.get(r.race);
      const track = el('div', { class: 'fv-rtrack' });
      segs.forEach((sg, k) => {
        const a = Math.min(sg.from, sg.to), z = Math.max(sg.from, sg.to);
        if (sg.part !== 'base') kinds.add(sg.part === 'swap' ? 'swap' : PART_KIND[sg.part] ?? 0);
        const neg = +sg.r.dps < 0 && sg.part !== 'base';
        const n = el('span', { class: 'sg' + (neg ? ' neg' : '') + (sg.part === 'swap' ? ' swap' : '') + (k === segs.length - 1 ? ' last' : ''), 'data-tip': '1',
          style: `left:${k === 0 ? 0 : x(a)}%;width:${Math.max(k === 0 ? 0 : 0.6, x(z) - (k === 0 ? 0 : x(a)))}%;background:${partCol(sg.part)};${neg ? 'z-index:1;' : ''}` });
        n.addEventListener('click', (e) => {
          e.stopPropagation();
          if (n.classList.contains('is-hl')) { hideTip(); return; }
          const info = noteRow('part', sg.part === 'base' ? 'base' : sg.part === 'target' ? 'target' : sg.part);
          const d = +sg.r.dps, se = +sg.r.se;
          const cr = card.getBoundingClientRect(), nr = n.getBoundingClientRect();
          showTip(card, nr.left - cr.left, nr.bottom - cr.top + 2, `${r.race} · ${sg.r.label}`,
            [sg.part === 'base' ? `${f1(d)} DPS on base stats alone` : `${sgn(d, f1)} DPS ± ${f1(se)}${Math.abs(d) < 2 * se ? ' (within the noise)' : ''}`,
              sg.part === 'base' ? null : `${sgn((100 * d) / r.dps, f2)}% of this race's DPS`, sg.r.note || null, info ? info.text : null], { hl: n, below: true });
        });
        track.append(n);
      });
      track.append(el('span', { class: 'wk', style: `left:${x(r.dps - r.se)}%;width:${Math.max(0.4, x(r.dps + r.se) - x(r.dps - r.se))}%` }));
      const sub = el('div', { class: 'fv-rsub' }, ...segs.filter((sg) => sg.part !== 'target' || Math.abs(+sg.r.dps) >= 0.05).map((sg) => el('span', {}, el('i', { style: `background:${partCol(sg.part)}` }),
        sg.part === 'base' ? `Base ${f0(+sg.r.dps)}` : `${String(sg.r.label).replace('Specialization', 'Spec.')} ${sgn(+sg.r.dps, f1)}`)),
        ...idle.slice(0, 3).map((q) => el('span', {}, el('em', { text: `${String(q.label).replace('Specialization', 'Spec.')}: ${q.note}` }))));
      return el('div', { class: 'fv-rrow' + (i === 0 ? ' top' : '') },
        tt(raceIco(r.raceIcon) || el('span', { class: 'fv-dot', style: `background:${fcol(r.race)}` }), `r:${r.race}:${b.class}`, true, b.id),
        el('span', { class: 'nm' }, el('b', { text: r.race, style: `color:${fcol(r.race)}` }), r.isPreset ? el('i', { text: 'preset race' }) : null),
        el('span', { class: 'end' }, mark(el('b', { text: f1(r.dps) }), 'race_variants', 'dps', W(r)), el('small', { text: i === 0 ? '★ best' : `−${f2(-r.vsBest)}%` })),
        track, sub);
    }));
    const top = rows[0], last = rows[rows.length - 1];
    $('#rs-take').replaceChildren(el('b', { text: b.label }), ': ', mark(el('b', { text: top.race }), 'race_variants', 'race', W(top)), ' leads with ', mark(el('span', { text: f1(top.dps) }), 'race_variants', 'dps', W(top)), ' DPS',
      ...(rows.length > 1 ? ['; ', rows[1].race, ' is ', f2(-rows[1].vsBest), '% behind and ', last.race, ' last at −', f2(-last.vsBest), '%.'] : ['.']));
    $('#rs-axis').textContent = `Bars start at ${f0(lo)} DPS.`;
    $('#rs-legend').replaceChildren(el('span', {}, el('i', { class: 'sw2', style: 'background:var(--cds-chart-muted)' }), 'Base stats'),
      ...[0, 1, 2, 'swap'].filter((k) => kinds.has(k)).map((k) => el('span', {}, el('i', { class: 'sw2', style: `background:${k === 'swap' ? 'var(--color-fg-muted)' : dash.colors[k]}` }), k === 'swap' ? 'Weapon swap' : KIND_NAME[k])));
  };

  // ---------- races: a finding per class ----------
  const drawRaceFind = () => {
    const host = $('#race-find');
    if (!waitOr(host, ['race_notes'])) return;
    const ns = rowsOf('race_notes').filter((n) => n.kind === 'class' && (st.heatCls === 'All' || n.key === st.heatCls));
    host.replaceChildren(...CLASSES.map((c) => ns.find((n) => n.key === c)).filter(Boolean).map((n) => el('div', {},
      tt(ringed('classicon_' + n.key.toLowerCase(), 'fv-ico ring', ccol(n.key)) || el('span', { class: 'fv-dot', style: `background:${ccol(n.key)}` }), 'c:' + n.key),
      el('span', {}, el('b', { text: n.key + '. ' }), mark(el('span', { text: n.text }), 'race_notes', 'text', `kind=class&key=${n.key}`)))));
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
  const drawTalents = (b) => {
    const host = $('#sp-tal');
    host.setAttribute('data-source', 'talenttrees builds icons treeart');
    if (!specWait(host, ['talenttrees'], b, [$('#sp-tal-seg')])) return;
    iconMap();
    const all = rowsOf('talenttrees').filter((t) => t.class === b.class);
    const parts = String(b.talents).split('-');
    const pts = (t) => +((parts[t.tree] || '')[t.idx] || 0);
    const trees = d3.groups(all, (t) => t.tree).sort((x, y) => x[0] - y[0]).map(([ti, ts]) => ({ ti, name: ts[0].treeName, ts, spent: d3.sum(ts, pts) }));
    const wide = $('#sp-tal-card').clientWidth > 700;
    if (st.talBuild !== b.id || st.talTree == null) { st.talTree = wide ? 'all' : String(trees.reduce((x, y) => (y.spent > x.spent ? y : x)).ti); st.talBuild = b.id; }
    $('#sp-tal-seg').replaceChildren(...[...trees.map((t) => [String(t.ti), `${t.name} ${t.spent}`]), ['all', 'All three']].map(([v, lbl]) => {
      const bt = el('button', { 'aria-pressed': String(st.talTree === v), text: lbl });
      bt.addEventListener('click', () => { st.talTree = v; drawTalents(b); });
      return bt;
    }));
    // The trees use the sim UI's talent picker markup and styles (talents_picker.tsx), read only.
    const shown = st.talTree === 'all' ? trees : trees.filter((t) => String(t.ti) === st.talTree);
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
    host.replaceChildren(el('div', { class: 'talents-picker-root' }, el('div', { class: 'talents-picker-inner' }, el('div', { id: 'talents-carousel', class: 'carousel' },
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
  };

  // ---------- stat weights ----------
  const drawWeights = (b) => {
    const host = $('#sp-weights');
    host.setAttribute('data-source', 'weights tooltips');
    if (!specWait(host, ['weights'], b, [$('#sp-wtab tbody')])) return;
    const rows = rowsOf('weights').filter((w) => w.build === b.id);
    if (!rows.length) { host.replaceChildren(el('span', { class: 'fv-note', text: 'No stat weights for this preset.' })); $('#sp-wtab tbody').replaceChildren(); $('#sp-wtab').dataset.build = b.id; return; }
    const ab = rows[0].primaryAbbr;
    $('#sp-weights-note').textContent = `What one more point of each stat is worth, in ${rows[0].primary.toLowerCase()} (${ab}): how many points of ${ab} give the same DPS. Hit and crit are per rating point (10 hit rating or 14 crit rating make 1%). From ${f0(rows[0].iters)} fights per stat at 180 s.`;
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
    }), ...(down ? [el('span', { class: 'fv-note', style: 'grid-column:1/-1', text: `Losing 1% hit would cost ${f1(down.dpsPer)} DPS, worth ${f2(down.ratingWorth)} ${ab} per rating point${shown.find((w) => w.stat === 'MeleeHit')?.dpsPer < 0.5 ? ': this preset sits at the hit cap, so more hit adds nothing.' : '.'}` })] : []));
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
    $('#sp-meta').replaceChildren(
      tt(el('span', {}, ringed('classicon_' + b.class.toLowerCase(), 'fv-ico ring', ccol(b.class)), ` ${b.spec}`), 'c:' + b.class),
      tt(el('span', {}, raceIco(b.raceIcon), ' ', mark(el('span', { text: b.race }), 'builds', 'race', W), ' ', iconImg(b.faction === 'Horde' ? 'ui_hordeicon' : 'ui_allianceicon')), `r:${b.race}:${b.class}`, true, b.id),
      mark(el('span', { text: b.treePoints }), 'builds', 'treePoints', W),
      isEnc(b) ? mark(encBadge(b), 'builds', 'encounter', W) : null);
    mark($('#sp-summary'), 'builds', 'summary', W).textContent = b.summary;
    const extra = b.petDps > 0 ? tile('Pet damage', f1(b.petShare), `${f1(b.petDps)} DPS from pets`, 'builds', 'petShare', W, '%')
      : b.oomSecs > 0 ? tile('Time out of mana', f1(b.oomSecs), 'per fight on average', 'builds', 'oomSecs', W, ' s')
        : tile('Threat per second', f0(b.tps), 'for tanks keeping aggro', 'builds', 'tps', W);
    $('#sp-tiles').replaceChildren(
      tile('Mean DPS', f1(b.dps), `± ${f2(b.se)} (1 s.e.)`, 'builds', 'dps', W),
      tile('Rank', String(rankOf(b, builds)), `of ${builds.length} presets`, 'builds', st.enc ? 'rankAll' : 'rank', W),
      tile('Unlucky fight', f1(b.p10), '10th percentile', 'builds', 'p10', W),
      tile('Lucky fight', f1(b.p90), '90th percentile', 'builds', 'p90', W),
      tt(tile('Best race', b.bestRace, `${f1(b.bestRaceDps)} DPS`, 'builds', 'bestRace', W, null, b.bestRaceIcon), `r:${b.bestRace}:${b.class}`, true, b.id),
      extra);
    for (const h of ['#sp-head', '#sp-tiles']) $(h).dataset.build = b.id;
    const i = order.findIndex((x) => x.id === b.id);
    $('#sp-prev').onclick = () => openSpec(order[(i - 1 + order.length) % order.length].id);
    $('#sp-next').onclick = () => openSpec(order[(i + 1) % order.length].id);

    // caveats
    const notes = rowsOf('notes');
    const nb = (notes || []).filter((n) => n.build === b.id);
    $('#sp-caveats').dataset.build = notes ? b.id : '';
    const cav = nb.filter((n) => n.kind === 'caveats');
    $('#sp-caveats').replaceChildren(...(cav.length ? cav.map((n) => mark(el('li', { text: n.text }), 'notes', 'text', `build=${enc(b.id)}&kind=caveats&order=${n.order}`)) : notes ? [el('li', { class: 'fv-note', text: 'None recorded.' })] : []));

    // a card that throws shows a failure note (never the previous preset's content)
    const safe = (fn, host, extras = []) => { try { fn(); } catch (err) { console.error(err); if (host) { fail(host); host.dataset.build = b.id; } for (const x of extras) if (x) x.replaceChildren(); } };
    safe(() => drawTalents(b), $('#sp-tal'));
    safe(() => drawDist(b), $('#sp-dist'));
    safe(() => drawWeights(b), $('#sp-weights'));
    const abil = (rowsOf('abilities') || []).filter((a) => a.build === b.id);
    const colors = ABILITY_COLORS(abil);
    safe(() => drawAbilities(b, abil, colors), $('#sp-abil'));
    safe(() => drawTimeline(b, colors), $('#tl-scroll'));
    safe(() => drawAuras(b), $('#sp-auras'));
    safe(() => drawSpecRaces(b), $('#sp-races'));
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
      const lines = sets.map((s) => { const d = s.bins.find((q) => q.lo === k); return el('span', {}, sets.length > 1 ? el('i', { class: 'sw', style: `background:${s.color}` }) : null, `${sets.length > 1 ? s.name + ': ' : ''}${d ? `${d3.format('.2%')(d.share)} of fights (${f0(d.count)}), ${d3.format('.1%')(d.cum)} at or below` : 'no fights'}`); });
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
    $('#sp-dist-note').textContent = `DPS of each of the ${f0(b.iters)} simulated fights, in 5 DPS bins.`;
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
    const tb = $('#sp-abiltab tbody');
    tb.replaceChildren(...abil.map((a) => tt(el('tr', { 'data-key': `${enc(b.id)},${enc(a.key)}` },
      el('td', { class: 'first' }, keyIcon(a.key), ' ' + a.ability), ...['dps', 'share', 'casts', 'hits', 'crits', 'critPct', 'ticks', 'avgHit', 'avgCrit', 'avgTick', 'misses', 'dodges', 'glances']
        .map((k) => el('td', { class: 'n', text: ['avgHit', 'avgCrit', 'avgTick'].includes(k) ? f0(a[k]) : f1(a[k]) }))), 'k:' + a.key)));
    $('#sp-abiltab').dataset.build = b.id;
  };

  // each aura's average uptime (every fight), and a mini timeline of the sample fight picked on the timeline
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
        bar = el('span', { class: 'tr fv-mini', 'data-tip': '1', role: 'button', 'aria-label': `${a.aura}: when it was up in the sample fight` },
          t0 < -0.55 ? el('i', { class: 'pre', style: `width:${pos(0)}%` }) : null,
          ...iv.map((q) => el('span', { style: `left:${pos(q[2])}%;width:${Math.max(0.4, pos(q[3]) - pos(q[2]))}%;background:${col}` }, q[4] > 1 && pos(q[3]) - pos(q[2]) > 6 ? el('b', { text: '×' + q[4] }) : null)),
          el('i', { class: 'pull', style: `left:${pos(0)}%` }));
        bar.addEventListener('click', (e) => {
          e.stopPropagation();
          if (bar.classList.contains('is-hl')) { hideTip(); return; }
          const r = card.getBoundingClientRect(), nb = bar.getBoundingClientRect();
          const lines = [`${f1(a.uptime)}% average over every fight`, `${f1(upHere)}% in this ${s.label.toLowerCase()}, ${iv.length} time${iv.length === 1 ? '' : 's'}`,
            ...iv.slice(0, 10).map((q) => `${fmtT(q[2])} to ${q[3] >= s.dur - 0.05 ? 'the end' : fmtT(q[3])}${q[4] > 1 ? ` · up to ${q[4]} stacks` : ''}`),
            iv.length > 10 ? `and ${iv.length - 10} more` : null];
          showTip(card, e.clientX - r.left - 20, nb.bottom - r.top, a.aura, lines, { hl: bar, below: true });
        });
      } else {
        bar = el('span', { class: 'tr' }, el('span', { style: `width:${a.uptime}%;background:${col}` }));
      }
      out.push(bar);
    }
    host.replaceChildren(...(out.length ? out : [el('span', { class: 'fv-note', text: 'Every aura is up the whole fight.' })]));
    $('#sp-auras-note').textContent = s ? `The number is the average uptime over every simulated fight. Each bar is the ${s.label.toLowerCase()} picked on the timeline, filled only while the aura was up${t0 < -0.55 ? ', with the pre-pull seconds shaded' : ''}. Tap a name for its tooltip, a bar for its times.` : 'Average uptime over every simulated fight. Tap a name for its tooltip.';
  };
  $('#sp-alwayson').addEventListener('click', (e) => { e.stopPropagation(); st.alwaysOn = !st.alwaysOn; persist(); render(); });
  $('#sp-setup-hide').addEventListener('click', (e) => { e.stopPropagation(); st.setupHide = !st.setupHide; persist(); $('#sp-setup').classList.toggle('hide-off', st.setupHide); e.currentTarget.setAttribute('aria-pressed', String(st.setupHide)); });

  const drawSpecRaces = (b) => {
    const host = $('#sp-races');
    host.setAttribute('data-source', 'races');
    if (!specWait(host, ['races'], b)) return;
    const rows = rowsOf('races').filter((r) => r.build === b.id).sort((x, y) => y.dps - x.dps);
    const W = host.clientWidth, rowH = 26, m = { l: 100, r: 64, t: 4, b: 4 };
    const H = m.t + m.b + rows.length * rowH;
    const lo = d3.min(rows, (r) => r.dps) * 0.985, hi = d3.max(rows, (r) => r.dps);
    const x = d3.scaleLinear([lo, hi], [m.l, W - m.r]);
    const svg = d3.create('svg').attr('width', '100%').attr('viewBox', `0 0 ${W} ${H}`).attr('height', H);
    rows.forEach((r, i) => {
      const yy = m.t + i * rowH;
      const rg = svg.append('g').attr('data-tt', `r:${r.race}:${b.class}`).attr('data-tt-tap', '').attr('data-tt-b', b.id);
      rg.append('rect').attr('x', 0).attr('y', yy).attr('width', W).attr('height', rowH).style('fill', 'transparent');
      rg.append('text').attr('x', m.l - 26).attr('y', yy + rowH / 2 + 4).attr('text-anchor', 'end').text(r.race).style('fill', fcol(r.race));
      if (r.raceIcon) svgIcon(rg, r.raceIcon, m.l - 22, yy + 4, rowH - 8);
      rg.append('rect').attr('x', m.l).attr('y', yy + 6).attr('height', rowH - 12).attr('width', Math.max(2, x(r.dps) - m.l)).attr('rx', 4).style('fill', fcol(r.race)).attr('opacity', 0.8);
      svg.append('text').attr('x', x(r.dps) + 6).attr('y', yy + rowH / 2 + 4).text(f1(r.dps) + (i === 0 ? ' ★' : ''));
    });
    svg.append('text').attr('x', m.l).attr('y', H).attr('font-size', 9).text(`axis starts at ${f0(lo)}`);
    host.replaceChildren(svg.node());
    $('#sp-races-link').onclick = (e) => { e.stopPropagation(); st.rsBuild = b.id; persist(); showTab('races'); requestAnimationFrame(() => $('#race-spec-card').scrollIntoView({ block: 'start' })); };
  };

  const drawSpecLen = (b) => {
    const host = $('#sp-len');
    host.setAttribute('data-source', 'lengths');
    if (!specWait(host, ['lengths'], b)) return;
    const pts = rowsOf('lengths').filter((r) => r.build === b.id).sort((p, q) => p.secs - q.secs).map((p) => ({ s: p.secs, v: p.dps, oom: p.oomSecs }));
    drawLenChart(host, [{ name: b.build, color: ccol(b.class), dots: true, pts }], { h: 200, tipFmt: (p) => `${f1(p.v)} DPS (${sgn(100 * (p.v / b.dps - 1))}% vs 180 s)${p.oom > 0 ? ` · ${f1(p.oom)} s out of mana` : ''}` });
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
    const lines = [
      ['h', section === 'prepull' ? `Before the pull, at ${fmtT(+r.at)}` : `Priority ${r.order}`, String(r.kind)],
      r.when ? ['t', 'When: ' + r.when] : null,
      r.extra ? ['t', r.extra] : null,
      r.notes ? ['m', r.notes] : null,
      r.race ? ['m', r.otherRace ? `${r.race} only: skipped on this preset's race.` : `${r.race} racial.`] : null,
      cond ? ['m', 'Condition: ' + aplText(cond)] : null,
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
        el('span', { class: 'fv-ppi pull' }, el('b', { text: 'Pull' }), el('em', { text: '0 s' })))] : []));
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
          el('div', { class: 'tx' }, el('div', { class: 'ics' }, ...run.map((q) => tt(ico(q, 'fv-ico'), `rot:priority:${q.order}`, true, b.id)), el('b', { text: run.map((q) => q.name).join(', ') })), el('span', { class: 'wh', text: why === 'race' ? `Racials for ${[...new Set(run.map((q) => q.race))].join(', ')}: skipped on ${b.race}.` : `Switched off in this preset${run[0].notes ? ': ' + run[0].notes.replace(/^Never:\s*/i, '') : '.'}` }))));
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
    $('#sp-rot-note').textContent = `${b.label}'s own priority list, as the sim runs it: whenever it can act, it takes the first line whose condition holds.${pp.length ? ' Pre-pull casts go out before the boss is pulled.' : ''} Casts are per fight on average, for the spell as a whole. Tap a line for the spell and its exact condition.`;
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
  const SETUP_SECTIONS = ['Consumables', 'Raid Buffs', 'Party Buffs', 'Offensive Debuffs', 'Defensive Debuffs', 'Professions'];
  const SETUP_SUBS = {
    'Consumables': ['Potion', 'Special', 'Flask', 'Weapon', 'Food', 'Attributes', 'Physical', 'Defensive', 'Spell', 'Misc', 'Engineering', 'Pet'],
    'Raid Buffs': ['Stats', 'Blessings', 'Mana', 'Other'],
    'Party Buffs': ['Attack Power', 'Totems', 'Crit', 'Defense', 'Resistances', 'Other'],
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
    [/^(Battle Shout|Trueshot Aura)/, 'Party Buffs', 'Attack Power'], [/Totem/, 'Party Buffs', 'Totems'],
    [/^(Leader Of The Pack|Moonkin Aura)/, 'Party Buffs', 'Crit'], [/^(Devotion Aura|Retribution Aura)/, 'Party Buffs', 'Defense'],
    [/Protection|Resistance/, 'Party Buffs', 'Resistances'],
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
      const q = r?.kind === 'item' ? r.quality : null;
      return tt(el('button', { class: 'fv-pick' + (q ? ' q-' + q : '') + (off ? ' off' : ''), 'aria-label': off ? name + ' (no effect)' : null },
        el('span', { class: 'ib' }, s.icon ? iconImg(s.icon, 'fv-ico') : el('span', { class: 'fv-ico fv-noico', text: name.slice(0, 2) }),
          s.level && s.level !== 'Improved' ? el('span', { class: 'lv', text: s.level }) : null, s.hand ? el('span', { class: 'hd', text: s.hand }) : null, off ? el('span', { class: 'x', text: '×', 'aria-hidden': 'true' }) : null),
        mark(el('span', { class: 'nmx', text: name }), 'setup', 'item', `build=${enc(b.id)}&item=${enc(s.item)}`)), 'su:' + s.item);
    };
    host.replaceChildren(...sections.map(([sec, xs]) => {
      const order = SETUP_SUBS[sec] || [];
      const subs = d3.groups(xs, (x) => x.p[1]).sort((p, q) => order.indexOf(p[0]) - order.indexOf(q[0]));
      return el('section', { class: 'fv-sgrp' + (sec === 'Consumables' ? ' wide' : '') },
        el('div', { class: 'gh' }, el('span', { text: sec }), el('small', { text: (() => { const on = xs.filter((x) => x.s.active !== false).length; return on === xs.length ? String(on) : `${on} of ${xs.length} active`; })() })),
        ...subs.map(([sub, ys]) => el('div', { class: 'fv-ssub' },
          sub ? el('div', { class: 'sh', text: sub }) : null,
          el('div', { class: 'fv-picks' }, ...ys.sort((p, q) => (p.s.active === false) - (q.s.active === false) || (p.s.slotOrder || 0) - (q.s.slotOrder || 0)).map((y) => pick(y.s))))));
    }));
    host.classList.toggle('hide-off', !!st.setupHide);
    $('#sp-setup-hide').setAttribute('aria-pressed', String(!!st.setupHide));
  };

  const drawTables = (b) => {
    for (const [fn, h] of [[drawGear, '#sp-gear'], [drawStats, '#sp-stats'], [drawSetup, '#sp-setup'], [drawRotation, '#sp-rotation']]) {
      try { fn(b); } catch (err) { console.error(err); fail($(h)); $(h).dataset.build = b.id; }
    }
  };

  // ---------- timeline ----------
  // Resources take the game's colours; spells and items take a colour from their icon (names.color).
  const RES_COL = { mana: '#3f7fe0', rage: '#d2413a', energy: '#a8bd2a', focus: '#e0832c', 'combo points': '#d9a62c', health: '#3faa4f' };
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
        casts: tlParse(s.casts, []), dmg: tlParse(s.dmg, []), auras: tlParse(s.auras, []), cds: tlParse(s.cds, []), res: tlParse(s.res, {}), units: tlParse(s.units, {}),
      }));
      const ord = [samples.find((s) => s.pct === 50), samples.find((s) => s.pct === 10), samples.find((s) => s.pct === 90)].filter(Boolean);
      for (const s of ord) s.t0 = Math.min(0, d3.min(s.casts, (c) => c[0]) ?? 0, d3.min(s.auras, (a) => a[2]) ?? 0) - (d3.min(s.casts, (c) => c[0]) < 0 ? 0.6 : 0.5);
      Object.assign(tlCache, { rows, id: b.id, list: ord });
    }
    return { src, list: tlCache.list };
  };
  const nameColor = (k) => { const n = byKey('names')?.get(k); return n && n.color ? n.color : null; };
  const fmtT = (t) => `${t < 0 ? '−' : ''}${Math.abs(t).toFixed(1)} s`;
  const kfmt = (v) => (v >= 10000 ? d3.format('.1f')(v / 1000) + 'k' : d3.format(',')(Math.round(v)));
  const OUTC = { H: 'hit', C: 'crit', G: 'glance', M: 'miss', D: 'dodge', P: 'parry', B: 'block', R: 'resist' };

  let tl = null; // current timeline context
  const tlReset = () => { $('#tl-readout').replaceChildren(el('span', { class: 'fv-note', text: 'Tap the timeline to see the casts, hits and buffs at that moment.' })); };
  const drawTimeline = (b, colors) => {
    const card = $('#sp-time-card');
    const sc = $('#tl-scroll');
    const src = 'timeline_' + b.class.toLowerCase();
    sc.setAttribute('data-source', src + ' names icons tooltips');
    iconMap();
    if (!['lanes', 'single'].includes(st.tlMode)) st.tlMode = 'lanes';
    seg($('#tl-mode'), 'tlMode', [['lanes', 'Lanes'], ['single', 'Single lane']], (v) => { if (v === 'single' && (st.zoom === 'fit' || st.zoom === '60')) st.zoom = 'auto'; persist(); drawSpec(); });
    const single = st.tlMode === 'single';
    card.querySelector('.fv-lanes').hidden = single;
    // while this preset's fights load, show nothing of the previous preset's
    if (!specWait(sc, [src, 'names'], b, [$('#tl-samples'), $('#tl-labels'), $('#tl-lanes'), $('#tl-log'), $('#tl-legend')])) {
      if (sc.dataset.build !== b.id) { tlReset(); tl = null; sc.style.minHeight = '16rem'; $('#tl-readout').dataset.build = ''; }
      return;
    }
    const names = byKey('names');
    const T = tlSamples(b);
    const ord = T ? T.list : [];
    if (!ord.length) { sc.replaceChildren(el('div', { class: 'fv-fail', text: 'No sample fight' })); return; }
    if (st.sample >= ord.length) st.sample = 0;
    const segS = $('#tl-samples');
    segS.replaceChildren(...ord.map((s, i) => {
      const bt = el('button', { 'aria-pressed': String(i === st.sample) }, `${s.label.replace(' fight', '')} `, mark(el('span', { text: f0(s.dps) }), src, 'dps', `build=${enc(b.id)}&seed=${s.seed}`));
      bt.addEventListener('click', () => { st.sample = i; persist(); drawSpec(); });
      return bt;
    }));
    segS.dataset.build = b.id;
    const s = ord[st.sample];
    // single lane opens zoomed in far enough to tell casts apart: 5 s windows for busy melee, 15 s otherwise
    const isAuto = (k) => /^oOtherAction(Attack|Shoot)/.test(k);
    if (st.zoom === 'auto') { st.zoom = s.casts.some((c) => /^oOtherActionAttack/.test(c[1])) && s.casts.filter((c) => c[4] === 'p' && !isAuto(c[1])).length / s.dur > 0.55 ? '5' : '15'; persist(); }
    if (!['fit', '60', '30', '15', '5'].includes(st.zoom)) st.zoom = 'fit';
    document.querySelectorAll('#tl-zoom button').forEach((bt) => bt.setAttribute('aria-pressed', String(bt.dataset.z === st.zoom)));
    const key = b.id + '|' + st.sample;
    if (sc.dataset.tlKey !== key) { tlReset(); sc.dataset.tlKey = key; }
    // which names have several ranks in this sample
    const nameOf = (k) => names.get(k)?.name || k;
    const rankCount = d3.rollup(new Set([...s.casts.map((c) => c[1]), ...s.dmg.map((d) => d[1])]), (v) => v.length, (k) => nameOf(k));
    const label = (k) => { const n = names.get(k); const nm = nameOf(k); return n?.rank && rankCount.get(nm) > 1 ? `${nm} r${n.rank}` : nm; };
    const unitName = (u) => (u === 'p' ? 'Player' : u === 't' ? 'Boss' : s.units?.[u] || 'Pet');
    const colorOf = (k) => nameColor(k) || colors.get(k) || 'var(--cds-chart-muted)';
    const t0 = s.t0, t1 = s.dur;
    const viewW = Math.max(200, sc.clientWidth);
    const pps = st.zoom === 'fit' ? viewW / (t1 - t0) : viewW / +st.zoom;
    const Wd = Math.max(viewW, (t1 - t0) * pps);
    const x = d3.scaleLinear([t0, t1], [0, Wd]);
    const permanent = (a) => a[2] <= 0.05 && a[3] >= s.dur - 0.05;
    const auraGroups = (who) => {
      const g = d3.groups(s.auras.filter((a) => (who === 't' ? a[1] === 't' : a[1] !== 't') && !permanent(a) && a[3] > a[2] && (who !== 't' || s.auras.some((q) => q[0] === a[0] && q[2] > 0.05))), (a) => a[0] + '|' + a[1]);
      return g.map(([k, iv]) => ({ key: iv[0][0], unit: iv[0][1], iv, up: d3.sum(iv, (a) => a[3] - a[2]) })).sort((p, q) => q.up - p.up).slice(0, who === 't' ? 7 : 9);
    };
    const buffs = auraGroups('p'), debuffs = auraGroups('t');
    const resKey = ['p:rage', 'p:energy', 'p:focus', 'p:mana'].find((k) => s.res[k]?.length > 2);
    const resName = resKey ? resKey.slice(2) : null;
    const resCol = resName ? RES_COL[resName] : null;
    const cpKey = s.res['p:combo points'] ? 'p:combo points' : null;
    const petCasts = s.casts.filter((c) => c[4] !== 'p' && c[4] !== 't');
    const resAt = (k, t) => { const pts = s.res[k]; if (!pts) return null; let v = null; for (const p of pts) { if (p[0] > t) break; v = p[1]; } return v; };
    const resMax = resKey ? d3.max(s.res[resKey], (p) => p[1]) || 1 : 1;
    const aurasAt = (t, who) => s.auras.filter((a) => (who === 't' ? a[1] === 't' : a[1] !== 't') && a[2] <= t && a[3] > t && !permanent(a))
      .map((a) => label(a[0]) + (a[4] > 1 ? ` ×${a[4]}` : '') + (a[1] !== 'p' && a[1] !== 't' ? ` (${unitName(a[1])})` : ''));
    const hid = new Set(String(st.lanes || '').split(',').filter(Boolean));
    // highlight one ability: everything else fades
    const fsel = $('#tl-focus');
    const keys = [...new Set(s.casts.map((c) => c[1]).concat(s.dmg.map((d) => d[1])))].filter((k) => names.has(k)).sort((p, q) => label(p).localeCompare(label(q)));
    fsel.replaceChildren(el('option', { value: '', text: 'Everything' }), ...keys.map((k) => el('option', { value: k, text: label(k) })));
    if (!keys.includes(st.focus)) st.focus = '';
    fsel.value = st.focus;
    const fo = (k, on) => (!st.focus || k === st.focus ? on : 0.1);
    // lanes
    const lanes = [];
    let yy = 22;
    const lane = (id, lbl, h, k) => { const l = { id, lbl, y: yy, h, key: k }; lanes.push(l); yy += h + 4; return l; };
    let Ld = null, Lc = null, Lr = null, Lx = null;
    if (!single) {
      Lc = lane('casts', 'Casts', 24);
      const groups = [['pet', 'Pet casts', petCasts.length > 0], ['cds', 'Cooldowns', s.cds.length > 0], ['b', 'Buffs', buffs.length > 0], ['t', 'Debuffs', debuffs.length > 0], ['res', resName ? resName[0].toUpperCase() + resName.slice(1) : '', !!resKey], ['dps', 'Running DPS', true]].filter((g) => g[2]);
      $('#tl-lanes').replaceChildren(...groups.map(([k, lbl]) => {
        const bt = el('button', { class: 'fv-chip', 'aria-pressed': String(!hid.has(k)), text: lbl });
        bt.addEventListener('click', (e) => { e.stopPropagation(); if (hid.has(k)) hid.delete(k); else hid.add(k); st.lanes = [...hid].join(','); persist(); drawTimeline(b, colors); });
        return bt;
      }));
    }
    const Lp = !single && petCasts.length && !hid.has('pet') ? lane('pet', 'Pet casts', 16) : null;
    if (!single) Ld = lane('dmg', 'Damage', 44);
    const Lcd = !single && s.cds.length && !hid.has('cds') ? lane('cds', 'Cooldowns', 16) : null;
    const Lb = single || hid.has('b') ? [] : buffs.map((a) => lane('b', label(a.key) + (a.unit !== 'p' ? ` (${unitName(a.unit)})` : ''), 10, a.key));
    if (!single && debuffs.length && !hid.has('t')) yy += 4;
    const Lt = single || hid.has('t') ? [] : debuffs.map((a) => lane('t', '◆ ' + label(a.key), 10, a.key));
    if (!single) {
      Lr = resKey && !hid.has('res') ? lane('res', resName[0].toUpperCase() + resName.slice(1), 40) : null;
      Lx = hid.has('dps') ? null : lane('dps', 'Running DPS', 40);
    }
    // single lane: damage numbers above, cast icons on the track, resource changes below
    let Sd = null, Si = null, Sr = null;
    if (single) {
      Sd = lane('sd', 'Damage', 30);
      Si = lane('si', 'Casts', 30);
      Sr = resKey ? lane('sr', resName[0].toUpperCase() + resName.slice(1), 14) : null;
    }
    const H = yy + 4;
    const svg = d3.create('svg').attr('width', Wd).attr('height', H).attr('viewBox', `0 0 ${Wd} ${H}`);
    // pre-pull shading, grid, axis and the pull marker
    if (t0 < -0.55) {
      svg.append('rect').attr('x', 0).attr('y', 14).attr('width', Math.max(0, x(0))).attr('height', H - 14).style('fill', 'var(--fv-prepull)');
      if (x(0) > 44) svg.append('text').attr('x', x(0) - 4).attr('y', H - 4).attr('text-anchor', 'end').text('pre-pull').style('font-size', '9px').style('fill', 'var(--color-fg-muted)');
    }
    const step = pps > 30 ? 1 : pps > 10 ? 5 : pps > 3 ? 10 : 30;
    const ticks = d3.range(Math.ceil(t0 / step) * step, t1 + 0.01, step).filter((d) => d !== 0);
    svg.append('g').selectAll('line').data(ticks).join('line').attr('x1', x).attr('x2', x).attr('y1', 16).attr('y2', H).style('stroke', 'var(--cds-chart-grid)').attr('stroke-width', 1);
    svg.append('g').selectAll('text').data(ticks.filter((d, i) => step >= 5 || Math.round(d) % 5 === 0)).join('text').attr('x', x)
      .attr('y', 11).attr('text-anchor', (d) => (x(d) < 14 ? 'start' : x(d) > Wd - 14 ? 'end' : 'middle')).text((d) => (d < 0 ? '−' : '') + Math.abs(d) + 's')
      .attr('opacity', (d) => (x(d) - x(0) < 58 && x(d) - x(0) > -22 ? 0 : 1));
    svg.append('line').attr('x1', x(0)).attr('x2', x(0)).attr('y1', 13).attr('y2', H).style('stroke', 'var(--fv-gold)').attr('stroke-width', 1.5).attr('stroke-dasharray', '3 2');
    svg.append('text').attr('x', x(0) + 3).attr('y', 11).text('Pull 0 s').style('fill', 'var(--fv-gold)').style('font-weight', '600').style('font-size', '9.5px');
    for (const l of lanes) if (l.id === 'b' || l.id === 't') svg.append('rect').attr('x', 0).attr('width', Wd).attr('y', l.y).attr('height', l.h).style('fill', 'var(--color-panel)').attr('opacity', 0.6);
    const icoOf = (k) => names.get(k)?.icon;
    if (!single) {
      // casts
      const castRect = (g, list, L) => {
        g.selectAll('rect').data(list).join('rect').attr('x', (c) => x(c[0])).attr('y', L.y).attr('height', L.h)
          .attr('width', (c) => Math.max(2, x(c[0] + Math.max(c[2], c[3], 0.15)) - x(c[0]) - 1)).attr('rx', 2)
          .style('fill', (c) => colorOf(c[1])).attr('opacity', (c) => fo(c[1], 0.92)).attr('data-tt', (c) => 'k:' + c[1]).attr('data-tt-tap', '');
        if (pps > 8) for (const c of list) { const w = Math.max(c[2], c[3], 0.15) * pps; const ic = icoOf(c[1]); if (ic && w >= L.h) svgIcon(g, ic, x(c[0]) + 1, L.y + 1, L.h - 2); }
        if (pps > 14) g.selectAll('text').data(list.filter((c) => (Math.max(c[2], c[3], 0.15)) * pps > 34 + L.h)).join('text')
          .attr('x', (c) => x(c[0]) + 3 + (icoOf(c[1]) ? L.h : 0)).attr('y', L.y + L.h / 2 + 3.5).text((c) => { const n = label(c[1]); const w = Math.max(c[2], c[3]) * pps; return n.length * 5.5 > w - 6 ? n.slice(0, Math.max(2, Math.floor((w - 6) / 5.5))) : n; })
          .style('fill', '#fff').style('font-size', '9.5px').style('paint-order', 'stroke').style('stroke', 'rgba(0,0,0,0.35)').style('stroke-width', '2px');
      };
      castRect(svg.append('g'), s.casts.filter((c) => c[4] === 'p'), Lc);
      if (Lp) castRect(svg.append('g'), petCasts, Lp);
      // damage
      const maxD = d3.max(s.dmg, (d) => d[2]) || 1;
      const hs = d3.scaleSqrt([0, maxD], [0, Ld.h - 2]);
      const dg = svg.append('g');
      dg.selectAll('rect').data(s.dmg.filter((d) => d[2] > 0)).join('rect').attr('x', (d) => x(d[0]) - 0.75).attr('width', pps > 20 ? 2.5 : 1.5)
        .attr('y', (d) => Ld.y + Ld.h - hs(d[2])).attr('height', (d) => Math.max(1, hs(d[2]))).style('fill', (d) => colorOf(d[1]))
        .attr('opacity', (d) => fo(d[1], d[3] === 'C' ? 1 : d[3] === 'G' ? 0.35 : 0.6)).attr('data-tt', (d) => 'k:' + d[1]).attr('data-tt-tap', '');
      dg.selectAll('text').data(s.dmg.filter((d) => d[2] === 0 && ['M', 'D', 'P'].includes(d[3]))).join('text').attr('x', (d) => x(d[0])).attr('y', Ld.y + Ld.h).attr('text-anchor', 'middle').text('×').style('font-size', '9px');
      // cooldowns
      if (Lcd) {
        const cg = svg.append('g');
        cg.selectAll('path').data(s.cds).join('path').attr('d', (c) => `M${x(c[0])},${Lcd.y} l5,${Lcd.h / 2} l-5,${Lcd.h / 2} l-5,-${Lcd.h / 2}z`).style('fill', (c) => colorOf(c[1])).style('stroke', 'var(--color-fg)').attr('stroke-width', 0.75).attr('data-tt', (c) => 'k:' + c[1]).attr('data-tt-tap', '');
        if (pps > 6) cg.selectAll('text').data(s.cds).join('text').attr('x', (c) => x(c[0]) + 7 + Lcd.h).attr('y', Lcd.y + Lcd.h - 3).text((c) => label(c[1])).style('font-size', '9px');
        if (pps > 6) for (const c of s.cds) { const ic = icoOf(c[1]); if (ic) svgIcon(cg, ic, x(c[0]) + 6, Lcd.y, Lcd.h); }
      }
      // auras
      const auraLane = (a, L, col) => {
        const g = svg.append('g');
        g.attr('data-tt', 'a:' + a.key).attr('data-tt-tap', '');
        g.selectAll('rect').data(a.iv).join('rect').attr('x', (v) => x(Math.max(t0, v[2]))).attr('width', (v) => Math.max(1.5, x(v[3]) - x(Math.max(t0, v[2])))).attr('y', L.y + 1).attr('height', L.h - 2).attr('rx', 2).style('fill', col).attr('opacity', 0.85);
        if (pps > 14) g.selectAll('text').data(a.iv.filter((v) => v[4] > 1 && (v[3] - v[2]) * pps > 14)).join('text').attr('x', (v) => x(v[2]) + 2).attr('y', L.y + L.h - 2).text((v) => '×' + v[4]).style('font-size', '8px').style('fill', '#fff');
      };
      buffs.forEach((a, i) => { if (Lb[i]) auraLane(a, Lb[i], nameColor(a.key) || colors.get(a.key) || dash.colors[i % 8]); });
      debuffs.forEach((a, i) => { if (Lt[i]) auraLane(a, Lt[i], nameColor(a.key) || colors.get(a.key) || 'var(--cds-chart-muted)'); });
      // resource, in the game's colour
      if (Lr) {
        const pts = s.res[resKey];
        const ry = d3.scaleLinear([0, resMax], [Lr.y + Lr.h, Lr.y]);
        const area = d3.area().curve(d3.curveStepAfter).x((p) => x(p[0])).y0(Lr.y + Lr.h).y1((p) => ry(p[1]));
        svg.append('path').attr('d', area([...pts, [t1, pts[pts.length - 1][1]]])).style('fill', resCol).attr('opacity', 0.28);
        svg.append('path').attr('d', d3.line().curve(d3.curveStepAfter).x((p) => x(p[0])).y((p) => ry(p[1]))([...pts, [t1, pts[pts.length - 1][1]]])).style('fill', 'none').style('stroke', resCol).attr('stroke-width', 1.5);
        svg.append('text').attr('x', Math.max(3, x(0) + 3)).attr('y', Lr.y + 9).text(f0(resMax)).style('font-size', '9px');
        if (cpKey) {
          const cp = s.res[cpKey];
          svg.append('path').attr('d', d3.line().curve(d3.curveStepAfter).x((p) => x(p[0])).y((p) => Lr.y + Lr.h - (p[1] / 5) * Lr.h)([...cp, [t1, cp[cp.length - 1][1]]])).style('fill', 'none').style('stroke', RES_COL['combo points']).attr('stroke-width', 1.2).attr('stroke-dasharray', '2 2');
        }
      }
    }
    // running dps
    const cum = [];
    let tot = 0;
    for (const d of [...s.dmg].sort((p, q) => p[0] - q[0])) { tot += d[2]; if (d[0] > 2) cum.push([d[0], tot / d[0]]); }
    const dy = Lx ? d3.scaleLinear([0, (d3.max(cum, (p) => p[1]) || 1) * 1.05], [Lx.y + Lx.h, Lx.y]) : null;
    if (Lx) svg.append('path').attr('d', d3.line().x((p) => x(p[0])).y((p) => dy(p[1]))(cum)).style('fill', 'none').style('stroke', ccol(b.class)).attr('stroke-width', 1.8).attr('stroke-linejoin', 'round');
    if (Lx && cum.length) svg.append('text').attr('x', Wd - 3).attr('y', dy(cum[cum.length - 1][1]) - 4).attr('text-anchor', 'end').text(f0(cum[cum.length - 1][1]) + ' DPS').style('font-size', '9.5px');

    // ---- single lane ----
    // one moment's state, for the tooltip: resource left, buffs and debuffs up, recent cooldowns
    const stateLines = (t) => {
      const out = [];
      const rv = resKey ? resAt(resKey, t + 0.001) : null;
      if (rv != null) out.push(el('span', {}, el('i', { class: 'sw', style: `background:${resCol}` }), `${resName[0].toUpperCase() + resName.slice(1)} left: ${f0(rv)} of ${f0(resMax)}`));
      const cp = cpKey ? resAt(cpKey, t + 0.001) : null;
      if (cp != null) out.push(`Combo points: ${f0(cp)}`);
      const bu = aurasAt(t, 'p'), de = aurasAt(t, 't');
      if (bu.length) out.push(`Buffs: ${bu.slice(0, 8).join(', ')}${bu.length > 8 ? ` +${bu.length - 8}` : ''}`);
      if (de.length) out.push(`Debuffs: ${de.slice(0, 6).join(', ')}${de.length > 6 ? ` +${de.length - 6}` : ''}`);
      const cdr = s.cds.filter((c) => c[0] <= t + 0.01 && c[0] >= t - 30);
      if (cdr.length) out.push(`Cooldowns in the last 30 s: ${cdr.map((c) => `${label(c[1])} (${fmtT(c[0])})`).join(', ')}`);
      return out;
    };
    if (single) {
      const evC = s.casts.map((c) => ({ t: c[0], k: c[1], u: c[4], cast: c[2], cost: c[5], kind: 'c' }));
      const evD = s.dmg.map((d) => ({ t: d[0], k: d[1], amt: d[2], o: d[3], tick: d[5], u: d[4], kind: 'd' }));
      const evR = [];
      if (resKey) {
        const pts = s.res[resKey];
        for (let i = 1; i < pts.length; i++) { const dv = pts[i][1] - pts[i - 1][1]; if (Math.abs(dv) >= 0.5) evR.push({ t: pts[i][0], dv }); }
      }
      // greedy clusters: neighbours closer than `gap` pixels share one mark
      const cluster = (list, gapOf) => {
        const out = [];
        for (const e of [...list].sort((p, q) => p.t - q.t)) {
          const last = out[out.length - 1];
          if (last && x(e.t) - x(last.t0) < gapOf(last)) { last.ev.push(e); last.t1 = e.t; } else out.push({ t0: e.t, t1: e.t, ev: [e] });
        }
        return out;
      };
      const ic = Si.h - 7;
      // a cast's damage and what it cost, for its tooltip
      const tipFor = (g, what) => (e, d, node) => {
        const r = card.getBoundingClientRect(), nb = node.getBoundingClientRect();
        const evs = what === 'c' ? d.ev : what === 'd' ? d.ev : d.ev;
        const head = evs.length === 1 && what !== 'r' ? `${label(evs[0].k)} · ${fmtT(evs[0].t)}` : `${fmtT(d.t0)}${d.t1 > d.t0 ? ` to ${fmtT(d.t1)}` : ''} · ${evs.length} ${what === 'c' ? 'casts' : what === 'd' ? 'hits' : 'changes'}`;
        const lines = [];
        if (what === 'c') {
          for (const c of evs.slice(0, 8)) {
            const hits = s.dmg.filter((q) => q[1] === c.k && q[0] >= c.t - 0.01 && q[0] <= c.t + Math.max(c.cast, 0) + 1.2 && (q[2] > 0 || q[3] !== 'H'));
            const dmgTxt = hits.length ? hits.slice(0, 3).map((q) => q[2] ? `${kfmt(q[2])}${q[3] === 'C' ? ' crit' : ''}` : OUTC[q[3]] || q[3]).join(', ') : '';
            lines.push(el('span', {}, evs.length > 1 ? el('b', { text: `${fmtT(c.t)} ${label(c.k)}` }) : el('b', { text: c.cast > 0 ? `${f1(c.cast)} s cast` : 'Instant' }),
              c.cost ? ` · −${f0(c.cost)}${resName ? ' ' + resName : ''}` : '', dmgTxt ? ` · ${dmgTxt}` : '', c.u !== 'p' ? ` · ${unitName(c.u)}` : ''));
          }
          if (evs.length > 8) lines.push(`and ${evs.length - 8} more`);
        } else if (what === 'd') {
          const sum = d3.sum(evs, (q) => q.amt), cr = evs.filter((q) => q.o === 'C').length;
          if (evs.length > 1) lines.push(el('b', { text: `${kfmt(sum)} damage${cr ? `, ${cr} crit${cr > 1 ? 's' : ''}` : ''}` }));
          for (const q of evs.slice(0, 8)) lines.push(`${evs.length > 1 ? fmtT(q.t) + ' ' + label(q.k) + ' ' : ''}${q.amt ? kfmt(q.amt) : ''} ${q.tick ? 'tick' : ''} ${OUTC[q.o] || q.o}`.replace(/\s+/g, ' ').trim());
          if (evs.length > 8) lines.push(`and ${evs.length - 8} more`);
        } else {
          const net = d3.sum(evs, (q) => q.dv);
          lines.push(el('b', { text: `${resName[0].toUpperCase() + resName.slice(1)} ${sgn(net, f0)}` }));
          for (const q of evs.slice(0, 6)) lines.push(`${fmtT(q.t)} ${sgn(q.dv, f0)}`);
        }
        lines.push(el('span', { class: 'fv-tip-sep' }), ...stateLines(d.t1));
        showTip(card, nb.left - r.left + nb.width / 2 - 10, nb.bottom - r.top, head, lines, { hl: node, below: true });
      };
      // casts: icons framed in their colour; pets smaller, on the lower half
      const cg = svg.append('g');
      const CW = ic + 4;
      // auto attacks: small ticks under the track, so the spells stand out
      const ag = cg.append('g');
      for (const g of cluster(evC.filter((q) => q.u === 'p' && isAuto(q.k)), () => 6)) {
        const gg = ag.append('g').datum(g);
        gg.append('rect').attr('x', x(g.t0) - 2).attr('y', Si.y + Si.h - 3).attr('width', Math.max(4, x(g.t1) - x(g.t0) + 4)).attr('height', 5).attr('rx', 2.5).style('fill', 'var(--cds-chart-muted)').attr('class', 'sl-cast');
        tipOn(gg, tipFor(g, 'c'));
      }
      for (const g of cluster(evC.filter((q) => q.u === 'p' && !isAuto(q.k)), () => CW)) {
        const e0 = g.ev[0], cx = x(g.t0);
        const gg = cg.append('g').datum(g).attr('opacity', g.ev.some((q) => !st.focus || q.k === st.focus) ? 1 : 0.15);
        gg.append('rect').attr('x', cx - 1).attr('y', Si.y).attr('width', ic + 2).attr('height', ic + 2).attr('rx', 3).style('fill', colorOf(e0.k)).attr('class', 'sl-cast');
        const ico = icoOf(e0.k);
        if (ico) svgIcon(gg, ico, cx + 1, Si.y + 2, ic - 2);
        else gg.append('text').attr('x', cx + ic / 2).attr('y', Si.y + ic / 2 + 4).attr('text-anchor', 'middle').text(label(e0.k).slice(0, 2)).style('fill', '#fff').style('font-size', '9px').style('font-weight', '600');
        if (g.ev.length > 1) {
          gg.append('rect').attr('x', cx + ic - 8).attr('y', Si.y - 4).attr('width', 13).attr('height', 10).attr('rx', 5).style('fill', 'var(--color-fg)');
          gg.append('text').attr('x', cx + ic - 1.5).attr('y', Si.y + 3.5).attr('text-anchor', 'middle').text('+' + (g.ev.length - 1)).style('fill', 'var(--color-bg)').style('font-size', '7.5px').style('font-weight', '700');
        }
        tipOn(gg, tipFor(g, 'c'));
      }
      for (const g of cluster(evC.filter((q) => q.u !== 'p' && q.u !== 't'), () => 14)) {
        const e0 = g.ev[0], cx = x(g.t0);
        const gg = cg.append('g').datum(g);
        gg.append('circle').attr('cx', cx + 6).attr('cy', Si.y + Si.h - 5).attr('r', 6).style('fill', colorOf(e0.k)).style('stroke', 'var(--color-bg)').attr('stroke-width', 1.5);
        const ico = icoOf(e0.k);
        if (ico && pps > 20) svgIcon(gg, ico, cx + 2, Si.y + Si.h - 9, 8);
        tipOn(gg, tipFor(g, 'c'));
      }
      // damage numbers: crits bold with a star, misses in words
      const dgp = svg.append('g');
      const wOf = (g) => 8 + 5.6 * (g.ev.length > 1 ? kfmt(d3.sum(g.ev, (q) => q.amt)).length + 3 : (g.ev[0].amt ? kfmt(g.ev[0].amt).length : 5));
      const dcl = cluster(evD.filter((q) => q.u !== 't'), wOf);
      dcl.forEach((g, i) => {
        const sum = d3.sum(g.ev, (q) => q.amt), crit = g.ev.some((q) => q.o === 'C'), miss = g.ev.length === 1 && !g.ev[0].amt;
        const big = g.ev.reduce((p, q) => (q.amt > p.amt ? q : p));
        const gg = dgp.append('g').datum(g).attr('opacity', g.ev.some((q) => !st.focus || q.k === st.focus) ? 1 : 0.15);
        const txt = miss ? (OUTC[g.ev[0].o] || g.ev[0].o) : `${kfmt(sum)}${g.ev.length > 1 ? ` ×${g.ev.length}` : ''}`;
        const yb = Sd.y + (i % 2 ? 26 : 13);
        gg.append('rect').attr('x', x(g.t0) - 2).attr('y', yb - 10).attr('width', wOf(g)).attr('height', 13).style('fill', 'transparent');
        gg.append('text').attr('x', x(g.t0)).attr('y', yb).text(txt).style('fill', miss ? 'var(--color-fg-muted)' : colorOf(big.k))
          .style('font-weight', crit ? '700' : '500').style('font-style', miss ? 'italic' : 'normal').style('font-size', crit ? '10.5px' : '9.5px');
        if (crit) gg.append('text').attr('x', x(g.t0) + wOf(g) - 7).attr('y', yb - 4).text('✦').style('fill', 'var(--fv-gold)').style('font-size', '8px');
        tipOn(gg, tipFor(g, 'd'));
      });
      // resource changes
      if (Sr) {
        const rg = svg.append('g');
        for (const g of cluster(evR, () => 26)) {
          const net = d3.sum(g.ev, (q) => q.dv);
          if (Math.abs(net) < 0.5) continue;
          const gg = rg.append('g').datum(g);
          gg.append('rect').attr('x', x(g.t0) - 2).attr('y', Sr.y).attr('width', 26).attr('height', Sr.h).style('fill', 'transparent');
          gg.append('text').attr('x', x(g.t0)).attr('y', Sr.y + 10).text(sgn(net, f0)).style('fill', resCol).style('font-weight', net < 0 ? '600' : '500').style('font-size', '9px');
          tipOn(gg, tipFor(g, 'r'));
        }
      }
    }
    // cursor
    const cursor = svg.append('line').attr('y1', 16).attr('y2', H).style('stroke', 'var(--color-fg)').attr('stroke-width', 1).attr('opacity', 0);
    svg.on('click', (e) => {
      const [px] = d3.pointer(e);
      const t = x.invert(px);
      cursor.attr('x1', px).attr('x2', px).attr('opacity', 0.8);
      readout(t);
    });
    // double-tap (or double-click) zooms in one step around that moment
    svg.on('dblclick', (e) => {
      e.preventDefault();
      const t = x.invert(d3.pointer(e)[0]);
      const steps = ['fit', '60', '30', '15', '5'];
      const nx = steps[Math.min(steps.length - 1, steps.indexOf(st.zoom) + 1)];
      if (nx === st.zoom) return;
      st.zoom = nx; persist(); drawSpec();
      const sc2 = $('#tl-scroll');
      if (tl) sc2.scrollLeft = Math.max(0, tl.x(t) - sc2.clientWidth / 2);
    });
    sc.replaceChildren(svg.node());
    sc.style.minHeight = '';
    $('#tl-labels').style.height = H + 'px';
    $('#tl-labels').replaceChildren(...lanes.map((l) => tt(el('div', { style: `top:${l.y + l.h / 2 - 5}px;font-size:${l.h < 14 ? '0.62rem' : '0.68rem'}${l.id === 'sr' || l.id === 'res' ? `;color:${resCol}` : ''}`, title: l.key ? null : l.lbl, text: l.lbl }), l.key ? 'a:' + l.key : null)));
    // legend: the resource colour, the crit star and the pull marker
    $('#tl-legend').replaceChildren(...[
      resKey ? el('span', {}, el('i', { class: 'sw', style: `background:${resCol}` }), resName[0].toUpperCase() + resName.slice(1)) : null,
      cpKey ? el('span', {}, el('i', { class: 'sw', style: `background:${RES_COL['combo points']}` }), 'Combo points') : null,
      single ? el('span', {}, el('b', { text: '✦', style: 'color:var(--fv-gold)' }), ' Crit') : null,
      el('span', {}, el('i', { class: 'sw pull' }), 'Pull at 0 s', t0 < -0.55 ? ', shaded before it' : ''),
      el('span', { class: 'fv-note', text: single ? 'Tap a mark for its numbers and the buffs up at that moment.' : 'Bars take their icon’s colour.' }),
    ].filter(Boolean));
    const readout = (t) => {
      const near = (tt2) => Math.abs(tt2 - t) <= (pps > 20 ? 0.6 : 1.5);
      const evs = [];
      for (const c of s.casts) if (near(c[0])) evs.push([c[0], `${unitName(c[4])} casts ${label(c[1])}`, c[5] ? `${f0(c[5])} cost` : '', 'k:' + c[1]]);
      for (const d of s.dmg) if (near(d[0])) evs.push([d[0], `${label(d[1])}${d[5] ? ' tick' : ''} ${OUTC[d[3]] || d[3]}`, d[2] ? f0(d[2]) : '', 'k:' + d[1]]);
      for (const a of s.auras) { if (near(a[2]) && a[2] > t0 + 0.6) evs.push([a[2], `${label(a[0])} on ${unitName(a[1]).toLowerCase()}`, 'gained', 'a:' + a[0]]); if (near(a[3]) && a[3] < s.dur - 0.05) evs.push([a[3], `${label(a[0])} on ${unitName(a[1]).toLowerCase()}`, 'faded', 'a:' + a[0]]); }
      evs.sort((p, q) => p[0] - q[0]);
      const up = buffs.filter((a) => a.iv.some((v) => v[2] <= t && v[3] >= t)).map((a) => label(a.key));
      const rv = resKey ? resAt(resKey, t) : null;
      const cd = cum.filter((p) => p[0] <= t).pop();
      const head = el('div', { class: 'ev' }, el('b', { text: fmtT(t) + (t < 0 ? ' (before the pull)' : '') }),
        el('span', { class: 'fv-note', text: [rv != null ? `${resName} ${f0(rv)}` : null, cd ? `running ${f0(cd[1])} DPS` : null].filter(Boolean).join(' · ') }));
      const list = evs.slice(0, 14).map((e2) => tt(el('div', { class: 'ev' }, el('span', { class: 't', text: e2[0].toFixed(2) }), el('span', { style: 'flex:1', text: e2[1] }), el('span', { text: e2[2] })), e2[3]));
      $('#tl-readout').replaceChildren(head, ...list, ...(up.length ? [el('div', { class: 'fv-note', text: 'Active: ' + up.join(', ') })] : []));
    };
    // log
    tl = { s, x, label, unitName, t0, t1, b: b.id };
    $('#tl-filter').value = st.filter;
    updateLog();
    for (const n of [$('#tl-labels'), $('#tl-log'), $('#tl-legend'), $('#tl-lanes'), $('#tl-readout')]) n.dataset.build = b.id;
    $('#sp-time-note').textContent = `${s.label}: ${f1(s.dps)} DPS over ${f1(s.dur)} s${t0 < -0.55 ? `, from ${fmtT(t0 + 0.6).replace('−', '')} before the pull` : ''}. ${single ? 'Every cast, hit and resource change on one track. Zoom in to pull clusters apart.' : 'Zoom in, scroll sideways, and tap anywhere to see what happened at that moment.'}`;
  };
  const updateLog = () => {
    if (!tl) return;
    const sc = $('#tl-scroll');
    const a = tl.x.invert(sc.scrollLeft), z = tl.x.invert(sc.scrollLeft + sc.clientWidth);
    const s = tl.s, f = st.filter, evs = [];
    if (f === 'all' || f === 'cast') for (const c of s.casts) if (c[0] >= a && c[0] <= z) evs.push([c[0], `${tl.unitName(c[4])}: ${tl.label(c[1])}`, c[2] ? `${c[2].toFixed(2)} s cast` : 'instant', 'k:' + c[1]]);
    if (f === 'all' || f === 'dmg') for (const d of s.dmg) if (d[0] >= a && d[0] <= z) evs.push([d[0], `${tl.label(d[1])}${d[5] ? ' (tick)' : ''}`, d[2] ? `${f0(d[2])}${d[3] === 'C' ? ' crit' : d[3] === 'G' ? ' glance' : ''}` : ({ M: 'miss', D: 'dodge', P: 'parry' })[d[3]] || d[3], 'k:' + d[1]]);
    if (f === 'all' || f === 'aura') for (const au of s.auras) { if (au[2] >= a && au[2] <= z && au[2] > tl.t0 + 0.6) evs.push([au[2], `${tl.label(au[0])} on ${tl.unitName(au[1]).toLowerCase()}`, '+', 'a:' + au[0]]); if (au[3] >= a && au[3] <= z && au[3] < s.dur - 0.05) evs.push([au[3], `${tl.label(au[0])} on ${tl.unitName(au[1]).toLowerCase()}`, '−', 'a:' + au[0]]); }
    evs.sort((p, q) => p[0] - q[0]);
    $('#tl-log').replaceChildren(...evs.slice(0, 400).map((e) => tt(el('div', {}, el('span', { class: 't', text: e[0].toFixed(2) }), el('span', { text: e[1] }), el('span', { class: 'x', text: e[2] })), e[3])),
      ...(evs.length > 400 ? [el('span', { class: 'fv-note', text: `First 400 of ${evs.length} events. Zoom in to see the rest.` })] : evs.length ? [] : [el('span', { class: 'fv-note', text: 'No events on screen. Scroll the timeline or pick another filter.' })]));
  };
  let logTimer = null;
  $('#tl-scroll').addEventListener('scroll', () => { clearTimeout(logTimer); logTimer = setTimeout(updateLog, 150); }, { passive: true });
  document.addEventListener('toggle', (e) => { if (e.target && e.target.id === 'tl-logbox') updateLog(); }, true);
  $('#tl-logbox summary').addEventListener('click', () => setTimeout(updateLog, 50));
  $('#tl-focus').addEventListener('change', (e) => { st.focus = e.target.value; persist(); drawSpec(); });
  $('#tl-filter').addEventListener('change', (e) => { st.filter = e.target.value; persist(); updateLog(); });
  document.querySelectorAll('#tl-zoom button').forEach((bt) => bt.addEventListener('click', () => {
    const sc = $('#tl-scroll');
    const mid = tl ? tl.x.invert(sc.scrollLeft + sc.clientWidth / 2) : 0;
    st.zoom = bt.dataset.z; persist();
    drawSpec();
    if (tl) sc.scrollLeft = Math.max(0, tl.x(mid) - sc.clientWidth / 2);
  }));

  // ---------- compare ----------
  // A side is a preset, or one of its race and setup variants from the Races tab. A variant only has its
  // mean DPS, because we sim the spread, fight lengths, stat weights and damage sources for the preset's own
  // race only.
  const CMP_SETUPS = [['baseline', 'Normal setup'], ['weapon', 'Racial weapon'], ['off', 'Racials off'], ['beast', 'Beast target'], ['elemental', 'Elemental target'], ['humanoid', 'Humanoid target']];
  const cmpSide = (k, b, rv) => {
    const mine = rv ? rv.filter((r) => r.build === b.id) : [];
    let race = st[k + 'r'] || b.race, v = st[k + 'v'] || 'baseline';
    if (!mine.some((r) => r.race === race)) race = b.race;
    const row = (vv) => mine.find((r) => r.race === race && r.variant === vv);
    // The racial weapon setup only exists for races with a weapon racial the preset can use. A setup can also
    // go missing from the data, like the Undead target that the generic boss replaced.
    if (!rv || !row(v) || (v === 'weapon' && !row('weapon')?.simmed)) v = 'baseline';
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
    const pr = (rowsOf('pairs') || []).find((p) => p.a === A.id && p.b === B.id);
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
    const rows = [['Mean DPS', 'dps', f1], ['Error of the mean', 'se', f2], ['Median', 'p50', f1], ['Unlucky (10th)', 'p10', f1], ['Lucky (90th)', 'p90', f1], ['Std dev', 'stdev', f1], ['120 s fight', 'dps120', f1], ['300 s fight', 'dps300', f1], ['Race', 'race', String],
      ...(anyVar ? [['Setup', 'setup', String]] : []), ...(rv ? [['Race rank', 'raceRank', String]] : []), ['Best race', 'bestRace', String], ['Boss', 'bossType', String], ['Rank', 'rank', String]];
    $('#cmp-table tbody').replaceChildren(...rows.map(([lbl, k, f]) => el('tr', {}, el('td', { text: lbl }), valTd(sA, k, f), valTd(sB, k, f), diffOf(k))));

    // The charts and the side by side cards show the sides that were simmed in full.
    const full = [sA, sB].filter((s) => s.full), part = [sA, sB].filter((s) => !s.full);
    const gap = $('#cmp-gap');
    gap.replaceChildren(...(part.length ? [part.map((s) => `${s.k} (${s.name})`).join(' and '), ` only ${part.length > 1 ? 'have their' : 'has its'} mean DPS. ${full.length ? `The cards below show ${full[0].k} only` : 'The spread, fight length, stat weights and damage sources show up when a side is a preset on its own race and setup'}, because we sim them for each preset's own race only.`] : []));
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
      const seen = new Map();
      for (const n of ns) { if (!seen.has(n.text)) seen.set(n.text, []); seen.get(n.text).push(builds.get(n.build).build); }
      return el('details', { style: 'border-top:1px solid var(--color-border-line);padding-block:0.5rem' }, el('summary', {}, el('span', { style: 'display:flex;gap:0.4rem;align-items:center;font-size:var(--cds-font-size-body)' }, el('span', { class: 'fv-dot', style: `background:${ccol(cls)}` }), `${cls} (${seen.size})`)),
        el('ul', { class: 'fv-list' }, ...[...seen].map(([t, bs]) => el('li', {}, t, el('span', { class: 'fv-note', text: ` · ${bs.join(', ')}` })))));
    }));
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

  // ---------- hero ----------
  const drawHero = () => {
    const host = $('#hero');
    const builds = genB(), m = iconMap();
    if (!builds || !m) return;
    const arts = later ? rowsOf('treeart') : null;
    if (arts && !$('#top').dataset.art) {
      const top = builds.reduce((p, q) => (+q.dps > +p.dps ? q : p));
      const parts = String(top.talents).split('-').map((t) => [...t].reduce((x, c) => x + (+c || 0), 0));
      const art = arts.find((a) => a.class === top.class && +a.tree === parts.indexOf(Math.max(...parts)));
      if (art && art.bg) { $('#top').style.setProperty('--fv-hero-art', `url("${art.bg}")`); $('#top').dataset.art = '1'; }
    }
    host.replaceChildren(...CLASSES.map((c) => {
      const b = builds.find((x) => x.class === c);
      const img = iconImg((b && b.classIcon) || 'classicon_' + c.toLowerCase() + ':l', '');
      if (img) { img.src = m.get('classicon_' + c.toLowerCase() + ':l') || img.src; img.style.setProperty('--ring', ccol(c)); img.alt = c; }
      const bt = el('button', { 'aria-pressed': String(st.cls === c), 'aria-label': `Show ${c}`, 'data-tt': 'c:' + c }, img);
      bt.addEventListener('click', () => { st.cls = st.cls === c ? 'All' : c; st.heatCls = st.cls; persist(); if (st.tab !== 'ranking' && st.tab !== 'races') showTab('ranking'); else render(); });
      return bt;
    }));
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
    fillStatic();
    if (!later && (st.tab !== 'ranking' || (statusOf('builds') === 'ok' && statusOf('icons') === 'ok'))) {
      if (st.tab !== 'ranking') loadLater();
      else (window.requestIdleCallback || setTimeout)(loadLater);
    }
    if (ttNode && !ttNode.isConnected) hideTT();
    try { drawHero(); } catch (e) { console.error(e); }
    try {
      if (st.tab === 'ranking') { drawRanking(); drawLenAll(); drawMana(); }
      else if (st.tab === 'races') { for (const [fn, h] of [[drawScen, null], [drawRaceOverview, '#race-overview'], [drawRaceSpec, '#race-spec'], [drawHeat, '#heat'], [drawRaceFind, '#race-find'], [drawFactionTable, null], [drawRaceSpread, '#racespread']]) safeTop(fn, h && $(h)); }
      else if (st.tab === 'spec') drawSpec();
      else if (st.tab === 'compare') drawCompare();
      else if (st.tab === 'method') drawMethod();
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

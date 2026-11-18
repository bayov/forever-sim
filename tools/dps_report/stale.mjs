// Says which DPS Report builds are out of date, so a regeneration only re-sims what changed.
//
// Each build (a level 60 preset, id "<class>__<preset>") depends on:
//   - shared inputs every build uses: the sim's own packages (sim/core, sim/common, sim/encounters, sim/lib), the
//     top-level sim/*.go that registers the classes, the protos the sim compiles (not ui.proto or test.proto, which
//     nothing in the sim imports), the CLI and go.mod, the report's local patches and dashstats, and the report's step
//     scripts (their sims, iterations and post-processing, adaptive.py's fight counts among them);
//   - its class's sim package sim/<class>/;
//   - its own inputs: its request tools/dps_report/req/<id>.json (what is simmed), the fields of its meta file the
//     step scripts read (class, spec, race, baseBuild, duty, pet, encounter), its base build's request (a variant's race
//     overrides depend on it), its races' own rotations and gear tools/dps_report/req/races/<id>.<Race>.json (and its
//     base build's, race_override.py), and its assumption scenario rotations tools/dps_report/asm_apls/<id>.*
//     (run_assumptions.py);
//   - the item database entries it wears: the items, enchants and random suffixes in its request and in its race
//     files' own gear sets (`equipment`), from assets/database/db.json (the sim reads nothing else of the database at
//     run time: sim/core/database_load.go).
// Test files, golden results, testdata and docs inside those folders are left out of every hash: the sim binary never
// reads them. tools/rotopt and the class's ui/<spec> folders are not inputs either: they generate requests (mksettings),
// and the requests are hashed themselves.
//
// ui/dps_report/provenance.json records, per build, the commit its data is current at and the hash of each input there.
// A build is stale when any hash differs at the commit being checked, or when it has no record. A record from before
// this input set (no `format`, or an older one) is compared by computing this input set at its recorded commit (git
// fetch --shallow-since=<date> when that commit is missing from a shallow clone).
//
// Stale says "may have moved", not "moved": pipeline.sh probe (probe.py) re-runs each stale build's sims at a tenth of
// their iterations and compares them with the probe stored in its raw pack, and only the builds that differ are
// re-simmed. Builds it clears are recorded with --by probe.
//
// The Tanking and Healing modes (tankheal.py) are recorded per preset under provenance.json's `benches` (tank, heal:
// {presets: {<stem>: record}}), the same way as the DPS builds. A preset (a request in tools/tankbench/req or
// tools/healbench/req) depends on the sim's packages (not the DPS-only CLI, patches and step scripts), the top-level
// sim/*.go, its bench's sources (tools/tankbench or tools/healbench, its req/ folder left out), tools/tankscore, the
// seven boss files in tools/encounters, tankheal.py and the race sims' scripts and racials switch (bench_races.py,
// adaptive.py, dashpaths.py, run_build.py's CLASS_RACES, patches 0002 and 0006), its own request, the class packages
// of every class in its raid, and the database entries those requests wear. A tank preset's raid is the tank alone. A
// healing preset's raid is the healer and healbench's frost mage fillers (-filler), plus its warrior tank (-tank) for a
// tank healer: a raid healer's raid has no tank (healbench -role raid). They are listed apart (`benches` in --json, a
// second block in the text), so probe.py, which reads `stale`, sees DPS builds only.
//
// Usage:
//   node tools/dps_report/stale.mjs [--commit <rev>] [--json]       list stale builds (default rev: HEAD)
//   node tools/dps_report/stale.mjs --record <id...|all> [--commit <rev>] [--iterations N] [--by sim|probe|hand]
//                                                                   record builds as current at <rev>
//   node tools/dps_report/stale.mjs --record-bench <tank|heal|tank:<stem,...>|heal:<stem,...> ...> [--commit <rev>]
//                                   [--iterations N]                record a mode's presets (all, or the ones named)
//                                                                   as current at <rev>
//   node tools/dps_report/stale.mjs --bench-fingerprint <tank|heal> [--commit <rev>]
//                                                                   print a mode's per-preset input fingerprints, the
//                                                                   ones --record-bench records (tankheal.py --resume)
import { execFileSync } from 'child_process';
import crypto from 'crypto';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');
const dataDir = path.join(root, 'ui/dps_report/data');
const provPath = path.join(root, 'ui/dps_report/provenance.json');

export const FORMAT = 3;
const STEP_SCRIPTS = ['run_build', 'run_races', 'adaptive', 'weights', 'mana_curves', 'ablate', 'run_assumptions', 'assumptions_def', 'race_override', 'dashpaths'].map(
	s => `tools/dps_report/${s}.py`,
);
export const SHARED_INPUTS = [
	'sim/core',
	'sim/common',
	'sim/encounters',
	'sim/lib',
	'proto',
	'cmd/wowsimcli',
	'go.mod',
	'go.sum',
	'tools/dps_report/patches',
	'tools/dps_report/dashstats',
	...STEP_SCRIPTS,
];
// files inside the inputs that the sim binary and the step scripts never read
const IGNORED = /(_test\.go|\.results|\.results\.tmp|\.md)$|\/testdata\/|^proto\/(ui|test)\.proto$/;
const META_FIELDS = ['class', 'spec', 'race', 'baseBuild', 'duty', 'pet', 'encounter'];

const args = process.argv.slice(2);
const opt = name => {
	const i = args.indexOf(name);
	return i >= 0 ? args[i + 1] : undefined;
};
const rev = opt('--commit') || 'HEAD';

const git = (...a) => execFileSync('git', a, { cwd: root, encoding: 'utf8', maxBuffer: 1 << 30 });
const sha = s => crypto.createHash('sha256').update(s).digest('hex');
const memo = f => {
	const m = new Map();
	return (...a) => {
		const k = a.join('\0');
		if (!m.has(k)) m.set(k, f(...a));
		return m.get(k);
	};
};
const hasCommit = memo(r => {
	try {
		git('cat-file', '-e', `${r}^{commit}`);
		return true;
	} catch {
		return false;
	}
});
const show = memo((r, p) => {
	try {
		return git('show', `${r}:${p}`);
	} catch {
		return null; // not present at this revision
	}
});
const showJson = (r, p) => {
	const s = show(r, p);
	return s === null ? null : JSON.parse(s);
};
const lsTree = memo((r, dir) =>
	git('ls-tree', '-r', r, '--', dir)
		.split('\n')
		.filter(Boolean)
		.map(l => ({ blob: l.split(/\s+/)[2], path: l.split('\t')[1] })),
);
// a folder's or file's hash with the ignored files left out; null when it doesn't exist at the revision
const inputHash = memo((r, p) => {
	const files = lsTree(r, p).filter(f => !IGNORED.test(f.path));
	return files.length ? sha(files.map(f => `${f.blob} ${f.path}`).join('\n')) : null;
});

// the database the sim loads, by kind and id
const db = memo(r => {
	const d = showJson(r, 'assets/database/db.json') || {};
	return {
		item: new Map((d.items || []).map(x => [x.id, x])),
		enchant: new Map((d.enchants || []).map(x => [x.effectId, x])),
		suffix: new Map((d.randomSuffixes || []).map(x => [x.id, x])),
	};
});
// the requests' players' gear: every player's equipment
const equipmentOf = req => (req?.raid?.parties || []).flatMap(party => (party.players || []).map(pl => pl.equipment));
// the database entries a set of gear wears: every item, its enchant and its random suffix
export const dbSlice = (r, equipments) => {
	const d = db(r);
	const rows = [];
	for (const eq of equipments)
		for (const it of eq?.items || []) {
			if (it.id) rows.push(['item', it.id, d.item.get(it.id) ?? null]);
			if (it.enchant) rows.push(['enchant', it.enchant, d.enchant.get(it.enchant) ?? null]);
			if (it.randomSuffix) rows.push(['suffix', it.randomSuffix, d.suffix.get(it.randomSuffix) ?? null]);
		}
	rows.sort((a, b) => (a[0] + a[1] < b[0] + b[1] ? -1 : 1));
	return sha(JSON.stringify(rows));
};
// the top-level sim/*.go (register_all.go: which classes the binaries compile)
const topSim = r => sha(lsTree(r, 'sim/').filter(f => /^sim\/[^/]+\.go$/.test(f.path) && !IGNORED.test(f.path)).map(f => `${f.blob} ${f.path}`).join('\n'));

const metaAt = (r, id) => showJson(r, `tools/dps_report/req/${id}.meta.json`);

// every input of a build at a revision, as {name: hash}
export const fingerprint = memo((r, id) => {
	const cls = id.split('__')[0];
	const fp = {};
	for (const p of [...SHARED_INPUTS, `sim/${cls}`]) fp[p] = inputHash(r, p);
	fp['sim/*.go'] = topSim(r);
	const reqPath = `tools/dps_report/req/${id}.json`;
	fp[reqPath] = inputHash(r, reqPath);
	const meta = metaAt(r, id);
	fp[`${reqPath.slice(0, -5)}.meta.json (sim fields)`] = meta ? sha(JSON.stringify(META_FIELDS.map(k => meta[k] ?? null))) : null;
	const base = meta?.baseBuild;
	if (base) fp[`tools/dps_report/req/${base}.json`] = inputHash(r, `tools/dps_report/req/${base}.json`);
	const ids = [id, base].filter(Boolean);
	// a race's own gear set (race_override.py `equipment`) is worn in its race sims: its database entries count too
	const gear = equipmentOf(showJson(r, reqPath));
	for (const f of lsTree(r, 'tools/dps_report/req/races/'))
		if (ids.some(i => path.basename(f.path).startsWith(i + '.'))) {
			fp[f.path] = f.blob;
			const eq = showJson(r, f.path)?.equipment;
			if (eq) gear.push(eq);
		}
	for (const f of lsTree(r, 'tools/dps_report/asm_apls/')) if (path.basename(f.path).startsWith(id + '.')) fp[f.path] = f.blob;
	fp['assets/database (the entries it wears)'] = dbSlice(r, gear);
	return fp;
});

// the Tanking and Healing modes' inputs (tankheal.py), per preset
export const BENCH_FORMAT = 3;
const BENCH_SHARED = ['sim/core', 'sim/common', 'sim/encounters', 'sim/lib', 'proto', 'go.mod', 'go.sum', 'tools/tankscore', 'tools/dps_report/tankheal.py',
	'tools/dps_report/bench_races.py', 'tools/dps_report/adaptive.py', 'tools/dps_report/dashpaths.py', 'tools/dps_report/run_build.py',
	'tools/dps_report/patches/0002-dash-DASH_RACIALS_OFF-switch-to-leave-racial-parts-o.patch',
	'tools/dps_report/patches/0006-dash-DASH_RACIALS_OFF_CLASS-scopes-the-racials-switch.patch'];
const BOSSES = ['magmadar', 'golemagg', 'onyxia', 'broodlord', 'vaelastrasz', 'chromaggus', 'ragnaros'];
const NOT_PRESETS = ['warrior_protection']; // the healing bench's tank (tankheal.py NOT_PRESETS)
// healbench's default tank and filler requests (tools/healbench/main.go -tank, -filler)
const HEAL_TANK = { cls: 'warrior', req: 'tools/tankbench/req/warrior_protection.json' };
const HEAL_FILLER = { cls: 'mage', req: 'tools/dps_report/req/mage__frost.json' };
export const BENCH = {
	// tankbench sims the tank alone, against a stand-in healing stream
	tank: { dir: 'tools/tankbench', raid: () => [] },
	// healbench -role tank: the tank, the healer and the fillers; -role raid (a _raid preset): no tank
	heal: { dir: 'tools/healbench', raid: stem => (stem.endsWith('_raid') ? [HEAL_FILLER] : [HEAL_TANK, HEAL_FILLER]) },
};
// a mode's presets at a revision: its bench's requests, less the healing bench's tank
export const benchPresets = memo((r, mode) =>
	lsTree(r, `${BENCH[mode].dir}/req/`)
		.map(f => f.path)
		.filter(p => /^[^/]+\.json$/.test(p.slice(BENCH[mode].dir.length + 5)))
		.map(p => path.basename(p, '.json'))
		.filter(s => !NOT_PRESETS.includes(s))
		.sort(),
);
export const benchFingerprint = memo((r, mode, stem) => {
	const b = BENCH[mode];
	const raid = b.raid(stem);
	const reqs = [`${b.dir}/req/${stem}.json`, ...raid.map(x => x.req)];
	const classes = [...new Set([stem.split('_')[0], ...raid.map(x => x.cls)])];
	const fp = {};
	for (const p of [...BENCH_SHARED, ...reqs, ...classes.map(c => `sim/${c}`), ...BOSSES.map(x => `tools/encounters/${x}.json`)]) fp[p] = inputHash(r, p);
	const src = lsTree(r, b.dir).filter(f => !f.path.startsWith(`${b.dir}/req/`) && !IGNORED.test(f.path));
	fp[`${b.dir} (the bench, its requests left out)`] = src.length ? sha(src.map(f => `${f.blob} ${f.path}`).join('\n')) : null;
	fp['sim/*.go'] = topSim(r);
	fp['assets/database (the entries it wears)'] = dbSlice(r, reqs.flatMap(p => equipmentOf(showJson(r, p))));
	return fp;
});
const benchModes = Object.keys(BENCH).filter(m => fs.existsSync(path.join(dataDir, `${m}_builds.json`)));

// results simmed with DASH_ALLOW_BIN_MISMATCH (build_bins.py), which --record and --record-bench refuse to mark current
const dashDir = process.env.DASH_DIR || path.join(root, 'tools/dps_report');
const binOverride = rel => (fs.existsSync(path.join(dashDir, rel)) ? JSON.parse(fs.readFileSync(path.join(dashDir, rel), 'utf8')) : null);
const refuseOverridden = (names, rel) => {
	const bad = names.filter(n => binOverride(rel(n)));
	for (const n of bad)
		console.error(`REFUSED to record ${n} current: simmed with binaries that don't match the code (DASH_ALLOW_BIN_MISMATCH=${binOverride(rel(n)).reason}, ${rel(n)}). Rebuild (pipeline.sh bins) and re-sim it.`);
	return names.filter(n => !bad.includes(n));
};

const diff = (a, b) => [...new Set([...Object.keys(a), ...Object.keys(b)])].filter(k => (a[k] ?? null) !== (b[k] ?? null)).sort();

// every build in the report, and every request not in it yet (a new preset)
const published = JSON.parse(fs.readFileSync(path.join(dataDir, 'builds.json'), 'utf8')).map(b => b.id);
const requested = fs
	.readdirSync(path.join(root, 'tools/dps_report/req'))
	.filter(f => f.endsWith('.meta.json'))
	.map(f => f.slice(0, -'.meta.json'.length));
const builds = [...new Set([...published, ...requested])].filter(id => requested.includes(id)).sort();
const removed = published.filter(id => !requested.includes(id));
const prov = fs.existsSync(provPath) ? JSON.parse(fs.readFileSync(provPath, 'utf8')) : { builds: {} };

// a preset's record: its own, else the whole mode's from before per-preset records (format 2: {commit, inputs})
const benchRecord = (m, stem) => {
	const rec = (prov.benches || {})[m];
	return rec?.presets ? rec.presets[stem] : rec?.commit ? { commit: rec.commit, recordedAt: rec.recordedAt, iterations: rec.iterations, format: rec.format } : undefined;
};

if (args.includes('--bench-fingerprint')) {
	// tankheal.py --resume: the mode's per-preset fingerprints at --commit, as JSON ({<stem>: the inputs --record-bench
	// records for it}), so a stopped run resumes across commits exactly when no preset's recorded inputs moved
	const m = opt('--bench-fingerprint');
	if (!BENCH[m]) throw new Error(`Unknown mode ${m}: tank or heal`);
	const commit = git('rev-parse', rev).trim();
	console.log(JSON.stringify(Object.fromEntries(benchPresets(commit, m).map(s => [s, benchFingerprint(commit, m, s)])), null, '\t'));
	process.exit(0);
}

if (args.includes('--record-bench')) {
	// <mode> records all of the mode's presets, <mode>:<stem>,<stem> only those (tankheal.py fresh prints these)
	const i = args.indexOf('--record-bench');
	const commit = git('rev-parse', rev).trim();
	const done = [];
	prov.benches = prov.benches || {};
	for (const a of args.slice(i + 1)) {
		if (a.startsWith('--')) break;
		const [m, list] = a.split(':');
		if (!BENCH[m]) throw new Error(`Unknown mode ${m}: tank or heal`);
		// a mode simmed on binaries that don't match the code (build_bins.py): its main bench ran whole, so no preset of it
		if (!refuseOverridden([m], x => `${x}/bin_override.json`).length) continue;
		const all = benchPresets(commit, m);
		const stems = list ? list.split(',').filter(Boolean) : all;
		for (const s of stems) if (!all.includes(s)) throw new Error(`Unknown ${m} preset ${s}`);
		// the presets left as they were keep their records (an older whole-mode one becomes theirs), and removed ones go
		const presets = {};
		for (const s of all) if (benchRecord(m, s)) presets[s] = benchRecord(m, s);
		for (const s of stems)
			presets[s] = {
				commit,
				recordedAt: new Date().toISOString(),
				iterations: Number(opt('--iterations')) || undefined,
				format: BENCH_FORMAT,
				inputs: benchFingerprint(commit, m, s),
			};
		prov.benches[m] = { presets };
		done.push(`${m} (${stems.length} of ${all.length} presets)`);
	}
	fs.writeFileSync(provPath, JSON.stringify(prov, null, '\t') + '\n');
	console.log(`Recorded ${done.join(', ') || 'no mode'} at ${commit.slice(0, 10)}`);
	process.exit(0);
}

if (args.includes('--record')) {
	const i = args.indexOf('--record');
	let ids = [];
	for (const a of args.slice(i + 1)) {
		if (a.startsWith('--')) break;
		ids.push(a);
	}
	if (ids.includes('all')) ids = builds;
	ids = refuseOverridden(ids, id => `bin_override/${id}.json`);
	const commit = git('rev-parse', rev).trim();
	const iterations = Number(opt('--iterations')) || undefined;
	const by = opt('--by') || 'sim';
	if (!['sim', 'probe', 'hand'].includes(by)) throw new Error(`--by ${by}: sim, probe or hand`);
	for (const id of ids) {
		if (!builds.includes(id)) throw new Error(`Unknown build ${id}`);
		const old = prov.builds[id];
		prov.builds[id] = {
			commit,
			recordedAt: new Date().toISOString(),
			// a probe or hand record keeps the sims it vouches for: their commit and iterations
			iterations: by === 'sim' ? iterations : old?.iterations,
			by,
			simCommit: by === 'sim' ? commit : old?.simCommit || old?.commit,
			format: FORMAT,
			inputs: fingerprint(commit, id),
		};
	}
	fs.writeFileSync(provPath, JSON.stringify(prov, null, '\t') + '\n');
	console.log(`Recorded ${ids.length} build(s) at ${commit.slice(0, 10)} (by ${by})`);
	process.exit(0);
}

const stale = [];
const missing = new Set();
for (const id of builds) {
	const rec = prov.builds[id];
	if (!rec) {
		stale.push({ id, reasons: ['no provenance record'] });
		continue;
	}
	let then = rec.inputs;
	if (rec.format !== FORMAT) {
		// an older record: this input set at the commit it was recorded at
		if (!hasCommit(rec.commit)) {
			missing.add(rec.commit);
			stale.push({ id, reasons: [`recorded at ${rec.commit.slice(0, 10)}, which this clone doesn't have`] });
			continue;
		}
		then = fingerprint(rec.commit, id);
	}
	const changed = diff(then, fingerprint(rev, id));
	if (changed.length) stale.push({ id, reasons: changed.map(p => `${p} changed`) });
}

// the Tanking and Healing modes: each preset like a DPS build, listed by mode
const benches = [];
for (const m of benchModes) {
	const stems = benchPresets(rev, m);
	const st = [];
	for (const s of stems) {
		const rec = benchRecord(m, s);
		if (!rec) {
			st.push({ id: s, reasons: ['no provenance record'] });
			continue;
		}
		let then = rec.inputs;
		if (rec.format !== BENCH_FORMAT) {
			if (!hasCommit(rec.commit)) {
				missing.add(rec.commit);
				st.push({ id: s, reasons: [`recorded at ${rec.commit.slice(0, 10)}, which this clone doesn't have`] });
				continue;
			}
			then = benchFingerprint(rec.commit, m, s);
		}
		const changed = diff(then, benchFingerprint(rev, m, s));
		if (changed.length) st.push({ id: s, reasons: changed.map(p => `${p} changed`) });
	}
	if (st.length) benches.push({ id: m, total: stems.length, presets: st.map(x => x.id), stale: st });
}

if (args.includes('--json')) {
	console.log(JSON.stringify({ rev, total: builds.length, stale, removed, benches }, null, 2));
} else {
	console.log(`${stale.length} of ${builds.length} builds are stale at ${rev}.`);
	for (const s of stale) console.log(`  ${s.id}: ${s.reasons.join(', ')}`);
	if (removed.length) console.log(`In the report with no request (removed presets; pipeline.sh publish drops them): ${removed.join(' ')}`);
	console.log(`${benches.length} of ${benchModes.length} Tanking and Healing modes are stale (tankheal.py run <mode> --stale, then publish).`);
	for (const b of benches) {
		console.log(`  ${b.id}: ${b.stale.length} of ${b.total} presets`);
		// presets with the same reasons on one line
		const by = new Map();
		for (const s of b.stale) by.set(s.reasons.join(', '), [...(by.get(s.reasons.join(', ')) || []), s.id]);
		for (const [why, ids] of by) console.log(`    ${ids.join(' ')}: ${why}`);
	}
	if (missing.size)
		console.log(`Fetch the recorded commits to compare them (git fetch --shallow-since=<a date before them> origin master): ${[...missing].join(' ')}`);
}

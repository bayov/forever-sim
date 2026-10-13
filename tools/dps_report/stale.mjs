// Says which DPS Report builds are out of date, so a regeneration only re-sims what changed.
//
// Each build (a level 60 preset, id "<class>__<preset>") depends on:
//   - shared inputs: the sim core, encounters, item database, proto and the settings tooling, which every build uses;
//   - its class's inputs: sim/<class>/ and the class's ui/<spec>/ folders (presets, gear sets, rotations).
// ui/dps_report/provenance.json records, per build, the git tree hash of each input when its data was made.
// A build is stale when any of those hashes differs from the commit being checked, or when it has no record.
//
// Usage:
//   node tools/dps_report/stale.mjs [--commit <rev>] [--json]       list stale builds (default rev: HEAD)
//   node tools/dps_report/stale.mjs --record <id...|all> [--commit <rev>] [--iterations N]
//                                                                   record builds as freshly generated at <rev>
import { execFileSync } from 'child_process';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');
const dataDir = path.join(root, 'ui/dps_report/data');
const provPath = path.join(root, 'ui/dps_report/provenance.json');

export const SHARED_INPUTS = ['sim/core', 'sim/common', 'sim/encounters', 'sim/lib', 'proto', 'assets/database', 'tools/rotopt'];

const args = process.argv.slice(2);
const opt = name => {
	const i = args.indexOf(name);
	return i >= 0 ? args[i + 1] : undefined;
};
const rev = opt('--commit') || 'HEAD';

const git = (...a) => execFileSync('git', a, { cwd: root, encoding: 'utf8' }).trim();
const treeHash = p => {
	try {
		return git('rev-parse', `${rev}:${p}`);
	} catch {
		return null; // not present at this revision
	}
};

// ui/<dir> folders belonging to a class: "warrior", "tank_warrior", "balance_druid", ...
const classInputs = cls => {
	const dirs = fs
		.readdirSync(path.join(root, 'ui'), { withFileTypes: true })
		.filter(d => d.isDirectory() && d.name.split('_').includes(cls))
		.map(d => `ui/${d.name}`);
	return [`sim/${cls}`, ...dirs];
};

const fingerprint = id => {
	const cls = id.split('__')[0];
	return Object.fromEntries([...SHARED_INPUTS, ...classInputs(cls)].map(p => [p, treeHash(p)]));
};

const builds = JSON.parse(fs.readFileSync(path.join(dataDir, 'builds.json'), 'utf8')).map(b => b.id);
const prov = fs.existsSync(provPath) ? JSON.parse(fs.readFileSync(provPath, 'utf8')) : { builds: {} };

if (args.includes('--record')) {
	const i = args.indexOf('--record');
	let ids = [];
	for (const a of args.slice(i + 1)) {
		if (a.startsWith('--')) break;
		ids.push(a);
	}
	if (ids.includes('all')) ids = builds;
	const commit = git('rev-parse', rev);
	const iterations = Number(opt('--iterations')) || undefined;
	for (const id of ids) {
		if (!builds.includes(id)) throw new Error(`Unknown build ${id}`);
		prov.builds[id] = { commit, recordedAt: new Date().toISOString(), iterations, inputs: fingerprint(id) };
	}
	fs.writeFileSync(provPath, JSON.stringify(prov, null, '\t') + '\n');
	console.log(`Recorded ${ids.length} build(s) at ${commit.slice(0, 10)}`);
	process.exit(0);
}

const stale = [];
for (const id of builds) {
	const rec = prov.builds[id];
	if (!rec) {
		stale.push({ id, reasons: ['no provenance record'] });
		continue;
	}
	const now = fingerprint(id);
	const changed = Object.keys(now).filter(p => now[p] !== rec.inputs[p]);
	if (changed.length) stale.push({ id, reasons: changed.map(p => `${p} changed`) });
}

if (args.includes('--json')) {
	console.log(JSON.stringify({ rev, total: builds.length, stale }, null, 2));
} else {
	console.log(`${stale.length} of ${builds.length} builds are stale at ${rev}.`);
	for (const s of stale) console.log(`  ${s.id}: ${s.reasons.join(', ')}`);
}

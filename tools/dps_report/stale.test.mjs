// Tests for stale.mjs: node --test tools/dps_report/stale.test.mjs
// Each test builds a small git repository with stale.mjs copied in, records it, commits a change and reads
// `stale.mjs --json`.
import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const script = path.join(path.dirname(fileURLToPath(import.meta.url)), 'stale.mjs');

const req = (cls, items) => ({ raid: { parties: [{ players: [{ class: `Class${cls}`, equipment: { items: items.map(id => ({ id })) } }] }] } });
const HEAL = ['priest_holy', 'priest_holy_raid', 'shaman_restoration', 'shaman_restoration_raid'];
const TANK = ['shaman_warden_balanced', 'warrior_protection_balanced'];

function repo() {
	const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'stale-test-'));
	const put = (p, v) => {
		fs.mkdirSync(path.dirname(path.join(dir, p)), { recursive: true });
		fs.writeFileSync(path.join(dir, p), typeof v === 'string' ? v : JSON.stringify(v));
	};
	const git = (...a) => execFileSync('git', a, { cwd: dir, encoding: 'utf8' });
	const stale = (...a) => execFileSync('node', [path.join(dir, 'tools/dps_report/stale.mjs'), ...a], { cwd: dir, encoding: 'utf8' });
	const commit = msg => {
		git('add', '-A');
		git('-c', 'user.name=t', '-c', 'user.email=t@t', 'commit', '-qm', msg);
		return git('rev-parse', 'HEAD').trim();
	};
	git('init', '-q');
	put('tools/dps_report/stale.mjs', fs.readFileSync(script, 'utf8'));
	put('ui/dps_report/data/builds.json', [{ id: 'hunter__bm' }]);
	put('ui/dps_report/data/tank_builds.json', []);
	put('ui/dps_report/data/heal_builds.json', []);
	put('assets/database/db.json', { items: [1, 2, 3, 10, 20, 30].map(id => ({ id, name: `item ${id}` })), enchants: [], randomSuffixes: [] });
	put('tools/dps_report/req/hunter__bm.json', req('Hunter', [1]));
	put('tools/dps_report/req/hunter__bm.meta.json', { class: 'hunter' });
	// the race's own gear set: item 2 is worn only here
	put('tools/dps_report/req/races/hunter__bm.Dwarf.json', { gearSet: 'x', equipment: { items: [{ id: 2 }] } });
	put('tools/dps_report/req/mage__frost.json', req('Mage', [30]));
	for (const p of ['sim/register_all.go', 'sim/core/core.go', 'tools/healbench/main.go', 'tools/tankbench/main.go'])
		put(p, 'package x\n');
	for (const c of ['hunter', 'mage', 'priest', 'shaman', 'warrior']) put(`sim/${c}/${c}.go`, 'package x\n');
	for (const s of HEAL) put(`tools/healbench/req/${s}.json`, req(s.split('_')[0], [10]));
	for (const s of TANK) put(`tools/tankbench/req/${s}.json`, req(s.split('_')[0], [10]));
	put('tools/tankbench/req/warrior_protection.json', req('Warrior', [20]));
	commit('base');
	stale('--record', 'all');
	stale('--record-bench', 'tank', 'heal');
	const base = commit('record');
	return { dir, put, git, stale, commit, base, json: (...a) => JSON.parse(stale('--json', ...a)) };
}
const staleOf = (j, mode) => (j.benches.find(b => b.id === mode)?.presets || []).sort();
const edit = (r, p) => r.put(p, fs.readFileSync(path.join(r.dir, p), 'utf8') + '// changed\n');

test('a fresh record is current', () => {
	const r = repo();
	const j = r.json();
	assert.deepEqual(j.stale, []);
	assert.deepEqual(j.benches, []);
});

test('an item worn only in a race gear set stales the build', () => {
	const r = repo();
	const db = JSON.parse(fs.readFileSync(path.join(r.dir, 'assets/database/db.json'), 'utf8'));
	db.items.find(x => x.id === 3).name = 'unworn, changed';
	r.put('assets/database/db.json', db);
	r.commit('an item nobody wears');
	assert.deepEqual(r.json().stale, []);
	db.items.find(x => x.id === 2).name = 'race gear, changed';
	r.put('assets/database/db.json', db);
	r.commit('an item the Dwarf gear set wears');
	assert.deepEqual(r.json().stale, [{ id: 'hunter__bm', reasons: ['assets/database (the entries it wears) changed'] }]);
});

test('the DPS builds hash the top-level sim/*.go and adaptive.py', () => {
	const r = repo();
	edit(r, 'sim/register_all.go');
	r.commit('register');
	assert.deepEqual(r.json().stale.map(s => s.reasons), [['sim/*.go changed']]);
	r.put('tools/dps_report/adaptive.py', '# fight counts\n');
	r.commit('adaptive');
	assert.deepEqual(r.json('--commit', 'HEAD').stale.map(s => s.reasons), [['sim/*.go changed', 'tools/dps_report/adaptive.py changed']]);
});

test('tools/dashstats and the report switches stale the DPS builds, a patch file nothing', () => {
	const r = repo();
	r.put('tools/dps_report/patches/0002-old.patch', 'no input any more\n');
	r.commit('a patch file');
	assert.deepEqual(r.json().stale, []);
	assert.deepEqual(r.json().benches, []);
	r.put('tools/dashstats/main.go', 'package main\n');
	r.commit('dashstats');
	assert.deepEqual(r.json().stale.map(s => s.reasons), [['tools/dashstats changed']]);
	assert.deepEqual(r.json().benches, []);
	r.put('sim/core/report_switches.go', 'package core\n');
	r.commit('switches');
	assert.deepEqual(r.json('--commit', 'HEAD').stale.map(s => s.reasons), [['sim/core changed', 'tools/dashstats changed']]);
	assert.deepEqual(staleOf(r.json(), 'heal'), HEAL.slice().sort());
});

test('a class package stales only the presets whose raid has the class', () => {
	const r = repo();
	edit(r, 'sim/shaman/shaman.go');
	r.commit('shaman');
	let j = r.json();
	assert.deepEqual(staleOf(j, 'heal'), ['shaman_restoration', 'shaman_restoration_raid']);
	assert.deepEqual(staleOf(j, 'tank'), ['shaman_warden_balanced']);
	assert.deepEqual(j.stale, []);
	// the healing raid's tank is a warrior, in tank healers' raids only (healbench -role raid has no tank)
	r.git('reset', '-q', '--hard', r.base);
	edit(r, 'sim/warrior/warrior.go');
	r.commit('warrior');
	j = r.json();
	assert.deepEqual(staleOf(j, 'heal'), ['priest_holy', 'shaman_restoration']);
	assert.deepEqual(staleOf(j, 'tank'), ['warrior_protection_balanced']);
});

test('the mage fillers stale every healing preset and no tank preset', () => {
	const r = repo();
	edit(r, 'sim/mage/mage.go');
	r.commit('mage');
	const j = r.json();
	assert.deepEqual(staleOf(j, 'heal'), HEAL);
	assert.deepEqual(staleOf(j, 'tank'), []);
});

test('sim/core stales every preset, run_build.py every bench preset', () => {
	const r = repo();
	edit(r, 'sim/core/core.go');
	r.commit('core');
	let j = r.json();
	assert.deepEqual(staleOf(j, 'heal'), HEAL);
	assert.deepEqual(staleOf(j, 'tank'), TANK);
	assert.equal(j.stale.length, 1);
	r.git('reset', '-q', '--hard', r.base);
	r.put('tools/dps_report/run_build.py', 'CLASS_RACES = {}\n');
	r.commit('races');
	j = r.json();
	assert.deepEqual(staleOf(j, 'heal'), HEAL);
	assert.deepEqual(staleOf(j, 'tank'), TANK);
});

test('a preset request stales that preset alone, the healing tank request the tank healers', () => {
	const r = repo();
	r.put('tools/healbench/req/priest_holy_raid.json', req('Priest', [10, 1]));
	r.put('tools/tankbench/req/warrior_protection.json', req('Warrior', [20, 1]));
	r.commit('requests');
	const j = r.json();
	assert.deepEqual(staleOf(j, 'heal'), ['priest_holy', 'priest_holy_raid', 'shaman_restoration']);
	assert.deepEqual(staleOf(j, 'tank'), []);
});

test('record-bench <mode>:<stems> records only those presets', () => {
	const r = repo();
	edit(r, 'sim/shaman/shaman.go');
	r.commit('shaman');
	r.stale('--record-bench', 'heal:shaman_restoration');
	const j = r.json();
	assert.deepEqual(staleOf(j, 'heal'), ['shaman_restoration_raid']);
	const prov = JSON.parse(fs.readFileSync(path.join(r.dir, 'ui/dps_report/provenance.json'), 'utf8'));
	assert.deepEqual(Object.keys(prov.benches.heal.presets).sort(), HEAL);
	assert.equal(prov.benches.heal.presets.priest_holy.commit, prov.benches.heal.presets.shaman_restoration_raid.commit);
	assert.notEqual(prov.benches.heal.presets.priest_holy.commit, prov.benches.heal.presets.shaman_restoration.commit);
	assert.throws(() => r.stale('--record-bench', 'heal:no_such_preset'));
});

test('an older whole-mode record is compared at its own commit', () => {
	const r = repo();
	const provPath = path.join(r.dir, 'ui/dps_report/provenance.json');
	const prov = JSON.parse(fs.readFileSync(provPath, 'utf8'));
	prov.benches.heal = { commit: r.base, recordedAt: 'then', iterations: 1000, format: 2, inputs: { old: 'hash' } };
	fs.writeFileSync(provPath, JSON.stringify(prov));
	assert.deepEqual(r.json().benches, []);
	edit(r, 'sim/priest/priest.go');
	r.commit('priest');
	assert.deepEqual(staleOf(r.json(), 'heal'), ['priest_holy', 'priest_holy_raid']);
	// recording some presets turns the old record into each other preset's own
	r.stale('--record-bench', 'heal:priest_holy');
	const now = JSON.parse(fs.readFileSync(provPath, 'utf8')).benches.heal;
	assert.equal(now.commit, undefined);
	assert.equal(now.presets.shaman_restoration.commit, r.base);
	assert.equal(now.presets.shaman_restoration.format, 2);
	assert.deepEqual(staleOf(r.json(), 'heal'), ['priest_holy_raid']);
});

test('bench-fingerprint prints the per-preset inputs record-bench stores, and moves only with a sim input', () => {
	const r = repo();
	const fp = JSON.parse(r.stale('--bench-fingerprint', 'heal'));
	const prov = JSON.parse(fs.readFileSync(path.join(r.dir, 'ui/dps_report/provenance.json'), 'utf8'));
	assert.deepEqual(Object.keys(fp).sort(), HEAL);
	for (const s of HEAL) assert.deepEqual(fp[s], prov.benches.heal.presets[s].inputs);
	// a commit that touches no input keeps the fingerprint (tankheal.py --resume carries on across it)
	r.put('README.md', 'docs\n');
	r.commit('docs');
	assert.equal(r.stale('--bench-fingerprint', 'heal'), r.stale('--bench-fingerprint', 'heal', '--commit', r.base));
	edit(r, 'sim/priest/priest.go');
	r.commit('priest');
	assert.notEqual(r.stale('--bench-fingerprint', 'heal'), r.stale('--bench-fingerprint', 'heal', '--commit', r.base));
	assert.throws(() => r.stale('--bench-fingerprint', 'dps'));
});

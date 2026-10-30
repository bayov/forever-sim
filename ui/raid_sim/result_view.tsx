// The comp search's result (docs/raid_sim.md, The result): every item a CompSearchResult holds, for each
// scenario (faction and camp setting). It leads with what could change the answer: the tie set and the
// sensitivity scenarios that change the winner.
import {
	CompMemberValue,
	CompResult,
	CompResultMember,
	CompScenarioResult,
	CompSearchRequest,
	CompSearchResult,
	Duty,
	PaladinJudgement,
	RaidPackage,
	RaidPreset,
	RaidRole,
} from '../core/proto/api';
import { Faction, MobType } from '../core/proto/common';
import {
	blessingName,
	dutyName,
	factionName,
	judgementName,
	mobTypeName,
	packageName,
	presetInfo,
	presetName,
	presetsFor,
	raceLabel,
	RaidSimData,
	roleName,
	totemName,
} from './data';
import {
	badge,
	ci95,
	ci95Labelled,
	collapsible,
	empty,
	errorStatus,
	fmt,
	fmtInt,
	fmtSigned,
	list,
	MEMBER_VALUE_COLUMNS,
	memberLabel,
	membersLabel,
	NA,
	partsToggle,
	plural,
	section,
	table,
} from './view_utils';

// A sensitivity scenario's winner: its members and tanks and healers, as the comps table names them. A
// result from before the winners carried their members has only the buff set's text.
const winnerText = (data: RaidSimData | null, w: CompResult): Node | string => {
	if (!w.members.length) return w.buffSet || 'another comp';
	return (
		<span>
			{dpsSummary(data, w)}
			{w.combo.length ? `; ${comboText(w.combo)}` : ''}
		</span>
	);
};

// A scenario's name: readable since the search names them from the report's assumptions; an older
// result's raw key ("raid_hits=high|natural_flask_zone=low") keeps a code font.
const scenarioLabel = (name: string) => <span className={`rs-scenario-name ${/[=|]/.test(name) ? 'rs-scenario-key' : ''}`}>{name}</span>;

// What the two DPS figures of a comp are, wherever both show.
const TWO_FIGURES =
	'Score: the members’ own single-player sims under the comp’s buffs, added up, which is how the search ranks comps. Raid sim DPS: the whole raid simmed together, which the confirmation runs for the top comps and which settles their order.';

export interface ResultViewOptions {
	// Opens a comp in the comp editor.
	onOpenComp?: (comp: CompResult, scenario: CompScenarioResult) => void;
	// Where the result came from, shown above it.
	source?: string;
}

export const scenarioName = (sc: CompScenarioResult) => `${factionName(sc.faction)}, camps ${sc.campBuffs ? 'on' : 'off'}`;

const comboText = (presets: RaidPreset[]) => presets.map(presetName).join(', ');

// The DPS members of a comp, counted, like "4× Mage: Frostfire". Each entry keeps its count and its name
// on one line, and the list wraps between entries.
const dpsSummary = (data: RaidSimData | null, comp: CompResult) => {
	const counts = new Map<string, { member: CompResultMember; n: number }>();
	comp.members
		.filter(m => m.role == RaidRole.RaidRoleDps)
		.forEach(m => {
			const key = m.buildId;
			const entry = counts.get(key);
			if (entry) entry.n++;
			else counts.set(key, { member: m, n: 1 });
		});
	const out: Node[] = [];
	[...counts.values()].forEach(({ member, n }, i) => {
		if (i > 0) out.push(document.createTextNode(', '));
		out.push(
			<span className="rs-comp-entry">
				{`${n}× `}
				{memberLabel(data, member.buildId)}
				{member.duties.length && !data?.byId.has(member.buildId) ? ` (${member.duties.map(dutyName).join(', ')})` : ''}
			</span>,
		);
	});
	return <span className="rs-comp-summary">{out}</span>;
};

let viewSeq = 0;

export class ResultView {
	readonly rootElem: HTMLElement;
	private readonly result: CompSearchResult;
	private readonly data: RaidSimData | null;
	private readonly opts: ResultViewOptions;
	private scenarioIndex = 0;
	private readonly scenarioHost: HTMLElement;
	private tabs: HTMLButtonElement[] = [];
	private readonly panelId = `rs-scenario-${++viewSeq}`;
	// The bar that jumps to each section of the long report.
	private readonly jumpHost = (<nav className="rs-jump" attributes={{ 'aria-label': 'Sections' }} />) as HTMLElement;

	constructor(result: CompSearchResult, data: RaidSimData | null, opts: ResultViewOptions = {}) {
		this.result = result;
		this.data = data;
		this.opts = opts;
		this.scenarioHost = (<div className="rs-scenario" attributes={{ role: 'tabpanel' }} />) as HTMLElement;
		this.scenarioHost.id = this.panelId;
		this.rootElem = (
			<div className="rs-result">
				{this.header()}
				{result.error ? <div className="alert alert-danger rs-error">{errorStatus('The search failed', result.error.message)}</div> : null}
				{result.scenarios.length ? this.overview() : empty('This result has no scenarios.')}
				{this.jumpHost}
				{this.scenarioHost}
				{this.global()}
			</div>
		) as HTMLElement;
		this.showScenario(0);
	}

	private header() {
		const r = this.result;
		const req = r.request ?? CompSearchRequest.create();
		const save = (
			<button className="btn btn-outline-primary btn-sm rs-save-json" type="button" title="Save this result as the JSON that Open a result reads">
				<i className="fas fa-download me-1" />
				Save JSON
			</button>
		) as HTMLButtonElement;
		save.addEventListener('click', () => saveJson(r));
		return (
			<header className="rs-result-header">
				<div className="rs-result-title-row">
					<h2 className="rs-result-title">
						Comp search: {plural(req.size, 'player')}, {plural(req.tanks, 'tank')}, {plural(req.healers, 'healer')}
					</h2>
					{save}
				</div>
				<p className="rs-result-subtitle">
					{encounterText(req)}. {fmtInt(r.simsRun)} single-player sims at {fmtInt(req.iterations)} iterations (contenders at{' '}
					{fmtInt(req.contenderIterations)}),{' '}
					<span title="Sims the search didn’t run again: from the cache, or from earlier in this search">{fmtInt(r.cacheHits)} reused</span>,{' '}
					{fmt(r.seconds, 0)} s. Sim version {r.version || 'unknown'}.
				</p>
				{this.opts.source ? <p className="rs-result-source">From {this.opts.source}</p> : null}
			</header>
		);
	}

	// Each scenario's best comp, and the picker for the scenario shown below.
	// The scenario tabs: a tablist whose arrow keys move between them, each tab naming the panel below.
	private overview() {
		const buttons = this.result.scenarios.map((sc, i) => {
			const best = sc.comps[0];
			const button = (
				<button className="rs-scenario-tab" type="button" attributes={{ role: 'tab' }}>
					<span className="rs-scenario-tab-name">{scenarioName(sc)}</span>
					<span className="rs-scenario-tab-dps">
						{!best ? 'no comp' : best.confirmed ? `${fmt(best.raidSimDps)} raid sim DPS` : `${fmt(best.dps)} DPS score`}
					</span>
					{sc.tieSet > 1 ? <span className="rs-scenario-tab-note">a tie of {sc.tieSet}</span> : null}
				</button>
			) as HTMLButtonElement;
			button.id = `${this.panelId}-tab-${i}`;
			button.setAttribute('aria-controls', this.panelId);
			button.addEventListener('click', () => this.showScenario(i));
			button.addEventListener('keydown', e => {
				const n = this.result.scenarios.length;
				const to = { ArrowRight: i + 1, ArrowDown: i + 1, ArrowLeft: i - 1, ArrowUp: i - 1, Home: 0, End: n - 1 }[e.key];
				if (to === undefined) return;
				e.preventDefault();
				const next = (to + n) % n;
				this.showScenario(next);
				this.tabs[next]?.focus();
			});
			return button;
		});
		this.tabs = buttons;
		return (
			<nav className="rs-scenario-tabs" attributes={{ role: 'tablist', 'aria-label': 'Faction and camp setting' }}>
				{buttons}
			</nav>
		);
	}

	private showScenario(i: number) {
		this.scenarioIndex = i;
		this.tabs.forEach((b, j) => {
			b.classList.toggle('active', j == i);
			b.setAttribute('aria-selected', j == i ? 'true' : 'false');
			b.tabIndex = j == i ? 0 : -1;
		});
		if (this.tabs[i]) this.scenarioHost.setAttribute('aria-labelledby', this.tabs[i].id);
		const sc = this.result.scenarios[i];
		if (!sc) {
			this.scenarioHost.replaceChildren();
			return;
		}
		this.scenarioHost.replaceChildren(...new ScenarioView(sc, this.data, this.opts).elements);
		this.renderJumps();
	}

	// A link to each section the report shows, in its order.
	private renderJumps() {
		const sections = [...this.rootElem.querySelectorAll<HTMLElement>('.rs-section')];
		this.jumpHost.replaceChildren(
			...sections.map((sec, i) => {
				sec.id ||= `${this.panelId}-section-${i}`;
				const link = (
					<a className="rs-jump-link" href={`#${sec.id}`}>
						{sec.querySelector('.rs-section-title')?.textContent ?? ''}
					</a>
				) as HTMLAnchorElement;
				// The address keeps its hash for raid links, so a jump only scrolls.
				link.addEventListener('click', e => {
					e.preventDefault();
					sec.scrollIntoView({ behavior: 'smooth', block: 'start' });
				});
				return link;
			}),
		);
	}

	// What the whole run shares: the assumptions, what it left out and what isn't modelled, and the request.
	private global() {
		const r = this.result;
		return (
			<div className="rs-global">
				{section('Assumptions', r.assumptions.length ? list(r.assumptions) : empty('None listed.'), {
					note: 'The rules the search ran under that the client or the launch checks have not settled.',
				})}
				{section('Left out', r.leftOut.length ? list(r.leftOut) : empty('Nothing.'), { note: 'Role options no build can express yet.' })}
				{section('Not modelled', r.notModelled.length ? list(r.notModelled) : empty('Nothing listed.'))}
				{section('The request', requestTable(r.request ?? CompSearchRequest.create()))}
			</div>
		);
	}
}

// A result as a JSON file, like tools/compsearch's <out>.json.
const saveJson = (result: CompSearchResult) => {
	const name = `comp-search-${result.request?.size ?? 'raid'}-${new Date().toISOString().slice(0, 16).replace(/[:T]/g, '-')}.json`;
	const url = URL.createObjectURL(new Blob([CompSearchResult.toJsonString(result, { prettySpaces: 1 })], { type: 'application/json' }));
	const link = (<a href={url} download={name} className="d-none" />) as HTMLAnchorElement;
	document.body.appendChild(link); // Firefox wants it in the page
	link.click();
	link.remove();
	setTimeout(() => URL.revokeObjectURL(url), 1000);
};

const encounterText = (req: CompSearchRequest) => {
	const e = req.encounter;
	if (!e) return 'Default fight';
	const parts = [`${e.name || 'Generic'} boss`];
	parts.push(
		e.useHealth ? `${fmtInt(e.bossHealth)} health (first guess ${fmt(e.duration, 0)} s)` : `${fmt(e.duration, 0)} ± ${fmt(e.durationVariation, 0)} s`,
	);
	if (e.mobType != MobType.MobTypeUnknown) parts.push(mobTypeName(e.mobType));
	parts.push(e.naturalFlaskZone ? 'in the natural flask zone' : 'outside the natural flask zone');
	return parts.join(', ');
};

const requestTable = (req: CompSearchRequest) => {
	const e = req.encounter;
	const a = req.assumptions;
	const rows: [string, string][] = [
		['Raid size', `${req.size}`],
		['Tank slots', `${req.tanks}`],
		['Healer slots', `${req.healers}`],
		['Fixed tanks', req.fixedTanks.map(presetName).join(', ') || 'none (searched)'],
		['Fixed healers', req.fixedHealers.map(presetName).join(', ') || 'none (searched)'],
		['Fixed DPS members', req.fixedDps.join(', ') || 'none'],
		['Encounter', e ? encounterText(req) : 'default'],
		['Factions', req.factions.map(f => factionName(f as Faction)).join(', ') || 'both'],
		['Camp buffs', req.campBuffs.map(c => (c ? 'on' : 'off')).join(', ') || 'on and off'],
		['Iterations', fmtInt(req.iterations)],
		['Contender iterations', fmtInt(req.contenderIterations)],
		['Seed', `${req.seed}`],
		['Confirm with raid sims', req.confirm ? `yes, at ${fmtInt(req.confirmIterations)} and ${fmtInt(req.confirmTopIterations)} iterations` : 'no'],
		['Sensitivity', req.skipSensitivity ? 'skipped' : 'run'],
		['Max combos', req.maxCombos ? `${req.maxCombos}` : 'all'],
		[
			'Assumption switches',
			a
				? [
						a.campChairSpellCritOnly ? 'the Camp Chair loses only its spell crit to Moonkin Aura' : '',
						a.campBuffsStackWithGroupBuffs ? 'camp buffs stack with group buffs' : '',
						a.journeymanCamp ? 'one Journeyman camp (asked for, not built yet)' : '',
				  ]
						.filter(s => s)
						.join('; ') || 'none'
				: 'none',
		],
	];
	return table(['Field', 'Value'], rows, 'rs-table-request');
};

// One scenario: a faction and camp setting.
class ScenarioView {
	readonly elements: HTMLElement[];
	private readonly sc: CompScenarioResult;
	private readonly data: RaidSimData | null;
	private readonly opts: ResultViewOptions;
	private readonly membersHost = (<div className="rs-members-host" />) as HTMLElement;

	constructor(sc: CompScenarioResult, data: RaidSimData | null, opts: ResultViewOptions) {
		this.sc = sc;
		this.data = data;
		this.opts = opts;
		this.elements = [
			this.leading(),
			this.comps(),
			section('Members by party', this.membersHost, {
				note: 'Every slot with its party, its duties and their cost, what it chose (judgement, totems) and what it gets (blessings, package, Power Infusion, Innervate).',
			}),
			this.combos(),
			this.roles(),
			this.buffs(),
			this.fillers(),
			this.memberValues(),
			this.adding(),
			this.sensitivity(),
			this.modelCheck(),
			this.notes(),
		];
		this.showMembers(0);
	}

	private get d() {
		return this.data;
	}

	// The tie set and the sensitivity scenarios that change the winner.
	private leading() {
		const sc = this.sc;
		const changes = sc.sensitivity.filter(s => !s.winnerHolds);
		const items: Node[] = [];
		const best = sc.comps[0];
		if (best) {
			items.push(
				best.confirmed ? (
					<p className="rs-lead-line">
						<strong>{scenarioName(sc)}:</strong> the best comp gets <strong>{fmt(best.raidSimDps)} raid DPS in the raid sim</strong>
						{` ${ci95Labelled(best.raidSimSe)}, with a score of ${fmt(best.dps)} ${ci95Labelled(best.se)}, over a ${fmt(
							sc.fightSeconds,
							0,
						)} s fight, with `}
						{comboText(best.combo)}.
					</p>
				) : (
					<p className="rs-lead-line">
						<strong>{scenarioName(sc)}:</strong> the best comp scores <strong>{fmt(best.dps)} raid DPS</strong>
						{` ${ci95Labelled(best.se)} over a ${fmt(sc.fightSeconds, 0)} s fight, with `}
						{comboText(best.combo)}. No raid sim confirmed it.
					</p>
				),
			);
			if (best.confirmed) items.push(<p className="rs-section-note">{TWO_FIGURES}</p>);
		}
		items.push(
			<p className="rs-lead-line">
				{sc.tieSet > 1
					? `The first ${sc.tieSet} comps are a tie: their ${
							best?.confirmed ? 'raid sim' : 'score'
					  } gaps are under the larger of 2 standard errors and 0.1%, so any of them is the answer.`
					: 'No other comp ties with the best.'}
			</p>,
		);
		if (sc.sensitivity.length) {
			items.push(
				<p className="rs-lead-line">
					The winner holds under {sc.sensitivity.length - changes.length} of {plural(sc.sensitivity.length, 'sensitivity scenario')}
					{changes.length == 1 ? ' and changes under this one:' : changes.length ? ' and changes under these:' : '.'}
				</p>,
			);
			if (changes.length) {
				items.push(
					list(
						changes.map(s => (
							<span>
								{scenarioLabel(s.scenario)}:{' '}
								{s.winners.length ? (
									<>
										{winnerText(this.d, s.winners[0])}, ahead by {fmt(s.winners[0].gap)}
										{s.winners.length > 1 ? ` (a tie of ${s.winners.length})` : ''}
									</>
								) : (
									s.note || 'another comp wins'
								)}
							</span>
						)),
					),
				);
			}
		}
		return section('What could change the answer', items, { className: 'rs-lead' });
	}

	private comps() {
		const sc = this.sc;
		if (!sc.comps.length) return section('The best comps', empty('No comp.'));
		// With a confirmation, the comps are ranked by their raid sims, and the raid sim columns join.
		const confirmed = sc.comps.some(c => c.confirmed);
		const rows = sc.comps.map((c, i) => {
			const show = (
				<button className="btn btn-sm btn-outline-primary rs-show-members" type="button" title="Show this comp's members below">
					Members
				</button>
			) as HTMLButtonElement;
			show.addEventListener('click', () => this.showMembers(i));
			const actions: Node[] = [show];
			if (this.opts.onOpenComp) {
				const open = (
					<button className="btn btn-sm btn-outline-secondary rs-open-comp" type="button" title="Load this comp into the comp editor">
						Edit
					</button>
				) as HTMLButtonElement;
				open.addEventListener('click', () => this.opts.onOpenComp!(c, sc));
				actions.push(open);
			}
			return [
				<span>
					{i + 1}
					{i < sc.tieSet && sc.tieSet > 1 ? badge('tie', 'info', 'In the tie set') : null}
				</span>,
				<span className="rs-dps">
					{fmt(c.dps)} <span className="rs-ci">{ci95(c.se)}</span>
				</span>,
				// The first comp is the one the gaps are measured from: by score, or by raid sim with a confirmation.
				i == 0 ? (
					confirmed ? (
						NA
					) : (
						'best'
					)
				) : (
					<span>
						{fmt(c.gap)} <span className="rs-ci">{ci95(c.gapSe)}</span>
					</span>
				),
				...(confirmed
					? [
							c.confirmed ? (
								<span className="rs-dps">
									{fmt(c.raidSimDps)} <span className="rs-ci">{ci95(c.raidSimSe)}</span>
								</span>
							) : (
								<span className="rs-muted">not raid-simmed</span>
							),
							!c.confirmed ? (
								NA
							) : i == 0 ? (
								'best'
							) : (
								<span>
									{fmt(c.raidSimGap)} <span className="rs-ci">{ci95(c.raidSimGapSe)}</span>
								</span>
							),
					  ]
					: []),
				fmt(c.modelDps),
				comboText(c.combo),
				dpsSummary(this.d, c),
				<span className="rs-actions">{actions}</span>,
			];
		});
		return section(
			'The best comps',
			[
				table(
					[
						'#',
						{ label: 'Score ±95%', numeric: true, tooltip: 'The members’ single-player DPS added up, with its 95% interval' },
						{
							label: 'Score gap ±95%',
							numeric: true,
							tooltip: `Behind the ${
								confirmed ? 'first comp' : 'best'
							}’s score, with the difference’s 95% interval, the two errors taken as independent, so on the wide side (below zero: ahead of it on score)`,
						},
						...(confirmed
							? [
									{ label: 'Raid sim DPS ±95%', numeric: true, tooltip: 'The whole raid simmed together, with its 95% interval' },
									{
										label: 'Raid sim gap ±95%',
										numeric: true,
										tooltip:
											'Behind the best confirmed comp in the raid sims, with the difference’s 95% interval, the two errors taken as independent, so on the wide side',
									},
							  ]
							: []),
						{ label: 'Model', numeric: true, tooltip: 'The model’s estimate of the score, before the exact single-player sims' },
						{ label: 'Tanks and healers', className: 'rs-col-words' },
						{ label: 'DPS members', className: 'rs-col-words' },
						'',
					],
					rows,
					'rs-table-comps',
				),
				<p className="rs-section-note">{TWO_FIGURES}</p>,
				<div className="rs-buffsets">
					{list(
						sc.comps.map((c, i) => (
							<span>
								<strong>#{i + 1}</strong> buff set: {c.buffSet || 'not described'}
							</span>
						)),
					)}
				</div>,
			],
			{
				note: confirmed
					? 'The best comp and its runners-up, ranked by raid sim DPS: the comps the raid sims confirmed come first, the rest follow by score.'
					: 'The best comp and its runners-up, ranked by score. This run didn’t confirm them with raid sims (the search’s Confirm option).',
			},
		);
	}

	private showMembers(i: number) {
		const comp = this.sc.comps[i];
		if (!comp) {
			this.membersHost.replaceChildren(empty('No comp.'));
			return;
		}
		this.membersHost.replaceChildren(
			<p className="rs-members-of">
				{`Comp #${i + 1}: a score of ${fmt(comp.dps)} ${ci95Labelled(comp.se)}`}
				{comp.confirmed ? `, ${fmt(comp.raidSimDps)} raid sim DPS ${ci95Labelled(comp.raidSimSe)}` : ''}
			</p>,
			partiesView(this.d, comp),
		);
	}

	private combos() {
		const sc = this.sc;
		if (!sc.combos.length && !sc.presetValues.length) return section('Tank and healer combos', empty('This run ranked no combos.'));
		// Which preset values the ranked combos give, as sim/compsearch's presetValues finds them: at the
		// best needs the best combo with that preset's slot changed to be ranked, and with minus without
		// needs a ranked combo with it and one without it.
		const key = (presets: RaidPreset[]) => [...presets].sort((a, b) => a - b).join(',');
		const ranked = new Set(sc.combos.map(c => key(c.presets)));
		const best = sc.combos[0]?.presets ?? [];
		// A result from before the search marked them says nothing, so it falls back to the ranked combos.
		const flagged = sc.presetValues.some(p => p.hasAtBest || p.hasBestWithMinusWithout);
		const presetValueApplies = {
			atBest: (p: RaidPreset) => {
				const idx = best.indexOf(p);
				const role = presetInfo(p)?.role;
				if (idx < 0 || role === undefined) return false;
				return presetsFor(role).some(q => q.id != p && ranked.has(key(best.map((x, j) => (j == idx ? q.id : x)))));
			},
			withMinusWithout: (p: RaidPreset) => sc.combos.some(c => c.presets.includes(p)) && sc.combos.some(c => !c.presets.includes(p)),
		};
		const combosTable = (n: number) =>
			table(
				[
					'#',
					'Tanks and healers',
					{ label: 'Best comp’s score', numeric: true, tooltip: 'The score of the best comp the search filled around this combo' },
					{ label: 'Behind the best combo', numeric: true },
					{
						label: 'Signature',
						tooltip: 'The tanks and healers that change what the raid gets: the combos with one signature share their buff sets',
					},
				],
				sc.combos
					.slice(0, n)
					.map((c, i) => [`${i + 1}`, comboText(c.presets), fmt(c.dps), fmt(c.gap), <span className="rs-muted">{c.signature}</span>]),
			);
		return section(
			'Tank and healer combos',
			[
				collapsible(combosTable, sc.combos.length, 10),
				<h4 className="rs-subtitle">Each preset’s DPS value</h4>,
				table(
					[
						'Preset',
						{
							label: 'At the winner’s combo',
							numeric: true,
							tooltip: 'The winner’s combo minus the best combo with that slot’s preset changed (screening scores)',
						},
						{
							label: 'With minus without',
							numeric: true,
							tooltip: 'The best combo with at least one of the preset, minus the best combo with none',
						},
					],
					sc.presetValues.map(p => [
						presetName(p.preset),
						(flagged ? p.hasAtBest : presetValueApplies.atBest(p.preset)) ? fmt(p.atBest) : NA,
						(flagged ? p.hasBestWithMinusWithout : presetValueApplies.withMinusWithout(p.preset)) ? fmt(p.bestWithMinusWithout) : NA,
					]),
				),
				<p className="rs-section-note">
					{NA} where the ranked combos can’t say: at the best, for a preset the best combo doesn’t have or no ranked combo swaps out; with minus
					without, for a preset in every ranked combo or in none.
				</p>,
			],
			{
				note: 'Every combination of tank and healer presets that fits the slots, ranked by its best comp’s score: the members’ single-player sims added up, from the screening fill, before any raid sim. Their own damage is left out.',
			},
		);
	}

	private roles() {
		const sc = this.sc;
		// A result from before the measured field has none set: its rows show their numbers as they are.
		const old = !sc.roles.some(r => r.measured);
		const notMeasured = () => <span className="rs-muted">not measured</span>;
		return section(
			'The role table',
			sc.roles.length
				? table(
						['Role', '', { label: 'Gain', numeric: true }, 'Provider', { label: 'Provider cost', numeric: true }, { label: 'Net', numeric: true }],
						sc.roles.map(r => {
							const measured = r.measured || old;
							return [
								r.role,
								r.inSet ? badge('in the set', 'success') : badge('missing', 'secondary', 'The value of adding it'),
								!measured ? notMeasured() : r.inSet ? fmt(r.gain) : '',
								r.inSet ? membersLabel(this.d, r.provider) : 'value of adding',
								measured && r.inSet ? fmt(r.providerCost) : '',
								measured ? fmt(r.net) : notMeasured(),
							];
						}),
				  )
				: empty('No roles.'),
			{
				note: 'Each role’s gain at the winning set, with that role alone removed and the raid re-filled; a role the set lacks shows its value of adding instead. A role only the placeholders give costs nothing and can’t be removed, so it isn’t measured.',
			},
		);
	}

	private buffs() {
		const sc = this.sc;
		return section(
			'Buffs and debuffs',
			[
				sc.buffs.length
					? table(
							['Effect', 'Scope', 'How', { label: 'Receivers', numeric: true }, 'Provider'],
							sc.buffs.map(b => [b.effect, b.scope, b.given ? 'given' : 'cast', `${b.receivers}`, membersLabel(this.d, b.provider)]),
					  )
					: empty('No buffs recorded.'),
				<h4 className="rs-subtitle">What the raid goes without</h4>,
				sc.misses.length ? list(sc.misses, 'rs-misses') : empty('Nothing: every buff the raid could give is up.'),
			],
			{ note: 'From the resolver’s record of the best comp. A given buff is up all fight; a cast one comes from a member’s rotation.' },
		);
	}

	private fillers() {
		const sc = this.sc;
		if (!sc.fillers.length) return section('The filler ranking', empty('No fillers.'));
		const lo = Math.min(...sc.fillers.map(f => f.assumptionLow || f.dps));
		const hi = Math.max(...sc.fillers.map(f => f.assumptionHigh || f.dps));
		const pos = (v: number) => `${((v - lo) / Math.max(1, hi - lo)) * 100}%`;
		return section(
			'The filler ranking',
			table(
				[
					'Build',
					{ label: 'DPS', numeric: true },
					{ label: 'Standard error', numeric: true, tooltip: 'One standard error; the 95% interval is 1.96 times it either side' },
					'Assumption range',
					'',
				],
				sc.fillers.map(f => [
					memberLabel(this.d, f.buildId, { icon: true }),
					fmt(f.dps),
					fmt(f.se),
					<span className="rs-range">
						<span className="rs-range-text">
							{fmt(f.assumptionLow, 0)} to {fmt(f.assumptionHigh, 0)}
						</span>
						<span className="rs-range-bar">
							<span
								className="rs-range-span"
								style={{ left: pos(f.assumptionLow), width: `calc(${pos(f.assumptionHigh)} - ${pos(f.assumptionLow)})` }}
							/>
							<span className="rs-range-dot" style={{ left: pos(f.dps) }} />
						</span>
					</span>,
					f.interchangeable ? badge('interchangeable', 'info', 'Within 2 standard errors of the best filler') : '',
				]),
				'rs-table-fillers',
			),
			{ note: 'Each build under the winning buff set without a package, with the spread of the DPS Report’s open assumptions (assumption_dps.json).' },
		);
	}

	private memberValues() {
		const sc = this.sc;
		const seats = sc.comps[0] ? valueSeats(sc.comps[0], sc.memberValues) : new Map<CompMemberValue, CompResultMember[]>();
		return section(
			'Member values',
			sc.memberValues.length
				? table(
						MEMBER_VALUE_COLUMNS,
						sc.memberValues.map(mv => {
							const seat = seats.get(mv) ?? [];
							const duties = seat[0]?.duties.filter(d => d != Duty.DutyNone) ?? [];
							return [
								memberLabel(this.d, mv.buildId, { icon: true, count: mv.count, seat: seat.length ? seatText(seat) : undefined }),
								fmt(mv.ownDps),
								fmtSigned(mv.contribution - mv.ownDps),
								fmt(mv.contribution),
								mv.bestReplacement ? fmt(mv.replacementMargin) : NA,
								mv.bestReplacement ? memberLabel(this.d, mv.bestReplacement) : NA,
								duties.length ? duties.map(dutyName).join(', ') : NA,
								duties.length ? fmt(mv.dutyCost) : NA,
								mv.reportDifference || mv.reportDifferenceParts.length ? fmtSigned(mv.reportDifference) : NA,
								partsToggle(mv.reportDifferenceParts),
							];
						}),
						'rs-table-values',
				  )
				: empty('No member values.'),
			{
				note: 'One row per distinct member in the best comp, with its group and package; identical members share one value. A replacement margin within noise means two builds are interchangeable.',
			},
		);
	}

	private adding() {
		const sc = this.sc;
		return section(
			'Value of adding',
			sc.adding.length
				? collapsible(
						n =>
							table(
								['Build', { label: 'Value', numeric: true }],
								sc.adding.slice(0, n).map(a => [memberLabel(this.d, a.buildId, { icon: true }), fmtSigned(a.value)]),
							),
						sc.adding.length,
						8,
				  )
				: empty('No builds left to add.'),
			{ note: 'A build that isn’t in the comp: the best raid DPS from swapping it in for any member, minus raid DPS.' },
		);
	}

	private sensitivity() {
		const sc = this.sc;
		return section(
			'Sensitivity',
			sc.sensitivity.length
				? collapsible(
						n =>
							table(
								['Scenario', '', 'Winner there', { label: 'Ahead by', numeric: true }, 'Note'],
								sc.sensitivity
									.slice()
									.sort((a, b) => +a.winnerHolds - +b.winnerHolds)
									.slice(0, n)
									.map(s => [
										scenarioLabel(s.scenario),
										s.winnerHolds ? badge('holds', 'success') : badge('changes', 'warning'),
										s.winnerHolds ? '' : s.winners.length ? winnerText(this.d, s.winners[0]) : '',
										s.winnerHolds || !s.winners.length ? '' : fmt(s.winners[0].gap),
										s.note,
									]),
							),
						sc.sensitivity.length,
						12,
				  )
				: empty('This run skipped the sensitivity scenarios.'),
			{ note: 'Whether the winner holds under each of the report’s open assumptions, and what wins where it doesn’t.' },
		);
	}

	private modelCheck() {
		const mc = this.sc.modelCheck;
		return section(
			'The model check',
			mc
				? table(
						['Buff sets scored', 'Simmed exactly', { label: 'Largest miss', numeric: true }, { label: 'Mean miss', numeric: true }],
						[[fmtInt(mc.setsScored), fmtInt(mc.setsSimmed), fmt(mc.maxMiss), fmt(mc.meanMiss)]],
				  )
				: empty('No model check.'),
			{ note: 'How far the model’s scores were from the exact sims of the sets it ranked best.' },
		);
	}

	private notes() {
		const sc = this.sc;
		return section('Notes and timings', [
			sc.notes.length ? list(sc.notes) : empty('No notes.'),
			<h4 className="rs-subtitle">Sims and seconds by phase</h4>,
			sc.timings.length ? list(sc.timings, 'rs-timings') : empty('No timings.'),
		]);
	}
}

// A comp's members, party by party, with what each one holds and gets.
export const partiesView = (data: RaidSimData | null, comp: CompResult) => {
	const parties = new Map<number, CompResultMember[]>();
	comp.members.forEach(m => {
		if (!parties.has(m.party)) parties.set(m.party, []);
		parties.get(m.party)!.push(m);
	});
	const sorted = [...parties.entries()].sort((a, b) => a[0] - b[0]);
	return (
		<div className="rs-parties">
			{sorted.map(([party, members]) => {
				const dps = members.reduce((s, m) => s + (m.role == RaidRole.RaidRoleDps ? m.dps : 0), 0);
				return (
					<div className="rs-party">
						<div className="rs-party-header">
							<span className="rs-party-name">Group {party}</span>
							<span className="rs-party-dps">{fmt(dps)} DPS</span>
						</div>
						{members.map(m => memberCard(data, m))}
					</div>
				);
			})}
		</div>
	);
};

// What makes one seat of a build differ from another: the seat's package and cooldowns, and its class's
// blessings (the search's askKey).
const seatKey = (m: CompResultMember) => [m.buildId, m.package, +m.powerInfusion, +m.innervate, m.blessings.join('.')].join('|');

// The best comp's members each member value stands for. A value's member index counts the search's own
// member list, which isn't the result's (that one has the placeholders first and is sorted by party), so
// a value is matched to the seats of its build that share one value: as many as its count, the closest
// in DPS when two seats fit.
export const valueSeats = (comp: CompResult, values: CompMemberValue[]): Map<CompMemberValue, CompResultMember[]> => {
	const groups = new Map<string, CompResultMember[]>();
	comp.members
		.filter(m => m.role == RaidRole.RaidRoleDps)
		.forEach(m => {
			const k = seatKey(m);
			if (!groups.has(k)) groups.set(k, []);
			groups.get(k)!.push(m);
		});
	const used = new Set<string>();
	const out = new Map<CompMemberValue, CompResultMember[]>();
	values.forEach(mv => {
		const fits = [...groups.entries()].filter(([k, ms]) => !used.has(k) && ms[0].buildId == mv.buildId);
		const exact = fits.filter(([, ms]) => ms.length == (mv.count || 1));
		const pool = exact.length ? exact : fits;
		if (!pool.length) return;
		pool.sort((a, b) => Math.abs(a[1][0].dps - mv.ownDps) - Math.abs(b[1][0].dps - mv.ownDps));
		used.add(pool[0][0]);
		out.set(mv, pool[0][1]);
	});
	return out;
};

// A seat in words, like "group 2 · Mana Spring · Power Infusion".
export const seatText = (members: CompResultMember[]) => {
	const m = members[0];
	const groups = [...new Set(members.map(x => x.party))].sort((a, b) => a - b);
	const parts = [`${groups.length > 1 ? 'groups' : 'group'} ${groups.join(', ')}`];
	if (m.package != RaidPackage.RaidPackageNone) parts.push(packageName(m.package));
	if (m.powerInfusion) parts.push('Power Infusion');
	if (m.innervate) parts.push('Innervate');
	return parts.join(' · ');
};

const memberCard = (data: RaidSimData | null, m: CompResultMember) => {
	const holds: string[] = [];
	// The result has one duty cost per member, its base build less it on duty. A member holds one duty, so
	// that's the duty's own cost; with more than one, it's their cost together.
	const duties = m.duties.filter(d => d != Duty.DutyNone);
	if (duties.length == 1) holds.push(`${dutyName(duties[0])}${m.dutyCost ? ` (costs ${fmt(m.dutyCost)} DPS)` : ''}`);
	else if (duties.length > 1) holds.push(`${duties.map(dutyName).join(', ')}${m.dutyCost ? ` (cost ${fmt(m.dutyCost)} DPS together)` : ''}`);
	if (m.judgement != PaladinJudgement.PaladinJudgementNone) holds.push(`judges ${judgementName(m.judgement)}`);
	if (m.totems.length) holds.push(`totems: ${m.totems.map(totemName).join(', ')}`);
	const gets: string[] = [];
	if (m.blessings.length) gets.push(`${m.blessings.length == 1 ? 'blessing' : 'blessings'}: ${m.blessings.map(blessingName).join(', ')}`);
	if (m.package != RaidPackage.RaidPackageNone) gets.push(packageName(m.package));
	if (m.powerInfusion) gets.push('Power Infusion');
	if (m.innervate) gets.push('Innervate');
	const isDps = m.role == RaidRole.RaidRoleDps;
	return (
		<div className={`rs-member ${isDps ? '' : 'rs-member-placeholder'}`}>
			<div className="rs-member-top">
				{memberLabel(data, m.buildId, { icon: true })}
				{isDps ? (
					<span className="rs-member-dps">
						{fmt(m.dps)}{' '}
						<span className="rs-ci" title="Its 95% interval">
							{ci95(m.se)}
						</span>
					</span>
				) : (
					badge(roleName(m.role), m.role == RaidRole.RaidRoleTank ? 'warning' : 'success')
				)}
			</div>
			<div className="rs-member-details">
				{m.race ? <span>{raceLabel(m.race)}</span> : null}
				{isDps && m.reportDps ? (
					<span title="Its DPS here minus its report build’s published DPS">vs report {fmtSigned(m.dps - m.reportDps)}</span>
				) : null}
				{m.tps ? <span>{fmt(m.tps)} TPS</span> : null}
			</div>
			{holds.length ? <div className="rs-member-holds">{holds.join('; ')}</div> : null}
			{gets.length ? <div className="rs-member-gets">gets {gets.join('; ')}</div> : null}
		</div>
	);
};

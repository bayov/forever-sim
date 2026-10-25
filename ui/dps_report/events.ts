// The Specs page's sample fights in the sim UI's own events view: the timeline, the chart, the combat log and the
// replay of a preset's typical, unlucky and lucky fight. tools/dps_report/events.py sims those fights again with
// their debug logs and writes them to data/events/<build>.json.gz.
import { EventsView, IterationKind } from '../core/components/detailed_results/events_view';
import { SimResultData } from '../core/components/detailed_results/result_component';
import { RaidSimRequest, RaidSimResult } from '../core/proto/api';
import { IconData, UIDatabase } from '../core/proto/ui';
import { ActionId, resourceTypeToIcon } from '../core/proto_utils/action_id';
import { Database } from '../core/proto_utils/database';
import { SimResult } from '../core/proto_utils/sim_result';
import { TypedEvent } from '../core/typed_event';
import type { Dash } from './data';
import { loadEvents, standalone } from './data_source';

// The report numbers its sample fights 0, 1 and 2 (st.sample in report.js).
const SAMPLE_KINDS: Array<IterationKind> = ['typical', 'unlucky', 'lucky'];

const ITERATION_TOOLTIPS: Partial<Record<IterationKind, string>> = {
	typical: 'A fight with about the median DPS, the 50th percentile.',
	unlucky: 'A fight at about the 10th percentile of DPS.',
	lucky: 'A fight at about the 90th percentile of DPS.',
};

const ICON_URL = /\/icons\/(?:large|medium|small)\/([\w-]+)\.jpg$/i;

interface NameRow {
	key: string;
	name: string;
	icon: string | null;
	rank: number | null;
}

interface IconRow {
	icon: string;
	url: string;
}

export class ReportEvents {
	private readonly host: HTMLElement;
	private readonly dash: Dash;
	private readonly emitter = new TypedEvent<SimResultData | null>('ReportEvents');
	private view: EventsView | null = null;
	private lookups: Promise<void> | null = null;
	private build: string | null = null;
	private token = 0;
	private sampleListener: ((sample: number) => void) | null = null;

	constructor(host: HTMLElement, dash: Dash) {
		this.host = host;
		this.dash = dash;
	}

	// Called when we pick another fight in the view.
	onSampleChange(listener: (sample: number) => void) {
		this.sampleListener = listener;
	}

	// Shows this preset's fights, on the given one. The report calls this each time it draws the Specs page.
	async show(build: string, sample: number) {
		const kind = SAMPLE_KINDS[sample] ?? 'typical';
		if (build == this.build) {
			this.view?.setIteration(kind);
			// The timeline draws only while we can see it, so we tell it when the page may have opened.
			document.dispatchEvent(new Event('shown.bs.tab'));
			return;
		}
		this.build = build;
		const token = ++this.token;
		this.setStatus('Loading the fight');

		try {
			await (this.lookups ??= this.loadLookups());
			const bytes = await loadEvents(build);
			if (token != this.token) return;
			if (!bytes) {
				this.setStatus('No sample fight for this preset.');
				return;
			}
			const { request, result } = JSON.parse(await gunzip(bytes));
			// The file keeps the typical fight's log only once, in typicalIteration.
			result.logs = result.typicalIteration.logs;
			result.firstIterationDuration = result.typicalIteration.duration;
			const simResult = await SimResult.makeNew(
				RaidSimRequest.fromJson(request, { ignoreUnknownFields: true }),
				RaidSimResult.fromJson(result, { ignoreUnknownFields: true }),
			);
			if (token != this.token) return;

			this.setStatus(null);
			this.view ??= new EventsView({
				parent: this.host,
				resultsEmitter: this.emitter,
				simUI: null,
				iterationTooltips: ITERATION_TOOLTIPS,
				onIterationChange: kind => this.sampleListener?.(Math.max(0, SAMPLE_KINDS.indexOf(kind))),
				// The replay shows a fight best to someone who hasn't seen the sim UI. A format picked
				// here is remembered under the report's own key, so a pick in the sim UI doesn't carry over.
				defaultFormat: 'replay',
				formatStorageKey: '__dps_report_events_format__',
			});
			const eventID = TypedEvent.nextEventID();
			this.emitter.emit(eventID, { eventID, result: simResult, filter: { player: 0, target: null } });
			this.view.setIteration(kind);
		} catch (e) {
			console.error(`DPS report: failed to load the fights of ${build}`, e);
			if (token == this.token) {
				this.build = null;
				this.setStatus('The fight failed to load.');
			}
		}
	}

	private setStatus(text: string | null) {
		let status = this.host.querySelector<HTMLElement>('.fv-events-status');
		if (!status) {
			status = document.createElement('div');
			status.className = 'fv-events-status fv-note';
			this.host.prepend(status);
		}
		status.textContent = text ?? '';
		status.hidden = text == null;
		this.host.classList.toggle('loading', text != null);
	}

	// The events view names and draws each spell and item through the sim UI's Database and wowhead's icons. We
	// give it the names and icons the report carries instead: the single file can't load anything from another site,
	// and the full item database is far bigger than what a few fights need.
	private async loadLookups() {
		const [names, ...iconSets] = await datasets<[NameRow[], IconRow[], IconRow[], IconRow[]]>(this.dash, [
			'event_names',
			'icons',
			'icons_more',
			'event_icons',
		]);
		const urls = new Map(iconSets.flat().map(row => [row.icon.toLowerCase(), row.url]));
		const questionMark = urls.get('inv_misc_questionmark') ?? '';
		const cache = new Map<string, string>();
		const mapIcon = (url: string) => {
			let mapped = cache.get(url);
			if (mapped === undefined) {
				const label = url.match(ICON_URL)?.[1].toLowerCase();
				mapped = (label && urls.get(label)) || (standalone ? questionMark : url);
				cache.set(url, mapped);
			}
			return mapped;
		};
		ActionId.setIconUrlMapper(mapIcon);
		for (const [type, url] of Object.entries(resourceTypeToIcon)) {
			if (url) resourceTypeToIcon[Number(type) as keyof typeof resourceTypeToIcon] = mapIcon(url);
		}

		const iconData = (prefix: string) =>
			names
				.filter(row => row.key[0] == prefix)
				.map(row => IconData.create({ id: Number(row.key.slice(1)), name: row.name, icon: row.icon ?? '', rank: row.rank ?? 0 }));
		Database.use(UIDatabase.create({ spellIcons: iconData('s'), itemIcons: iconData('i') }));
	}
}

// Resolves with the datasets once all of them have loaded.
const datasets = <T extends unknown[]>(dash: Dash, ids: Array<string>): Promise<T> =>
	new Promise((resolve, reject) => {
		let done = false;
		dash.onData(() => {
			if (done) return;
			const results = ids.map(id => dash.data(id));
			const bad = results.find(r => r.status == 'error' || r.status == 'missing');
			if (bad) {
				done = true;
				reject(new Error(bad.message));
			} else if (results.every(r => r.status == 'ok')) {
				done = true;
				resolve(results.map(r => r.data) as T);
			}
		});
	});

const gunzip = async (bytes: Uint8Array): Promise<string> => {
	// A server may have unpacked the file on the way already (Content-Encoding: gzip).
	if (bytes[0] != 0x1f || bytes[1] != 0x8b) return new TextDecoder().decode(bytes);
	const stream = new Blob([bytes]).stream().pipeThrough(new DecompressionStream('gzip'));
	return new Response(stream).text();
};

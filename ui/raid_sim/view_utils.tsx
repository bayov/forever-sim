// Small building blocks the raid sim page's views share: numbers, tables and member labels.
import { Class } from '../core/proto/common';
import { classNames, cssClassForClass } from '../core/proto_utils/utils';
import { buildLabel, CLASS_ICONS, RAID_PRESETS, RaidSimData } from './data';

// A value that rounds to zero prints as 0, never as -0.0.
const rounded = (n: number, digits: number) => {
	const text = n.toFixed(digits);
	return /^-0\.?0*$/.test(text) ? text.slice(1) : text;
};
export const fmt = (n: number | undefined, digits = 1) => (n === undefined || isNaN(n) ? '' : rounded(n, digits));
export const fmtSigned = (n: number | undefined, digits = 1) => {
	if (n === undefined || isNaN(n)) return '';
	const text = rounded(n, digits);
	return text.startsWith('-') || /^0\.?0*$/.test(text) ? text : `+${text}`;
};
// The dash a table shows for a value that doesn't apply.
export const NA = '—';
// A 95% interval from a standard error.
export const ci95 = (se: number | undefined, digits = 1) => (se ? `±${(1.96 * se).toFixed(digits)}` : '');
export const fmtInt = (n: number | bigint | undefined) => (n === undefined ? '' : Number(n).toLocaleString('en-US'));

export type Cell = Node | string | number | null | undefined | Array<Node | string>;

export interface Column {
	label: string;
	numeric?: boolean;
	className?: string;
	tooltip?: string;
}

// A table that scrolls sideways on a narrow screen instead of making the page wider.
export const table = (columns: Array<Column | string>, rows: Cell[][], className = ''): HTMLElement => {
	const cols = columns.map(c => (typeof c == 'string' ? { label: c } : c));
	return (
		<div className={`rs-table-wrap ${className}`}>
			<table className="rs-table">
				<thead>
					<tr>
						{cols.map(c => (
							<th className={`${c.numeric ? 'numeric' : ''} ${c.className ?? ''}`} title={c.tooltip ?? ''}>
								{c.label}
							</th>
						))}
					</tr>
				</thead>
				<tbody>
					{rows.map(row => (
						<tr>
							{row.map((cell, i) => (
								<td className={`${cols[i]?.numeric ? 'numeric' : ''} ${cols[i]?.className ?? ''}`}>{cell as any}</td>
							))}
						</tr>
					))}
				</tbody>
			</table>
		</div>
	) as HTMLElement;
};

export const section = (title: string, body: Node | Array<Node | string | null>, opts: { id?: string; note?: string; className?: string } = {}) => {
	const elem = (
		<section className={`rs-section ${opts.className ?? ''}`}>
			<h3 className="rs-section-title">{title}</h3>
			{opts.note ? <p className="rs-section-note">{opts.note}</p> : null}
			<div className="rs-section-body">{body as any}</div>
		</section>
	) as HTMLElement;
	if (opts.id) elem.id = opts.id;
	return elem;
};

export const list = (items: Array<Node | string>, className = '') => (
	<ul className={`rs-list ${className}`}>
		{items.map(i => (
			<li>{i}</li>
		))}
	</ul>
);

export const badge = (text: string, variant = 'secondary', title = '') => (
	<span className={`badge rs-badge bg-${variant}`} title={title}>
		{text}
	</span>
);

export const empty = (text: string) => <p className="rs-empty">{text}</p>;

// An error for a status line: its first line, and the rest (a Go stack trace, say) in a collapsible, so
// a long error never makes the line wider than the screen.
export const errorStatus = (prefix: string, message: string | undefined): HTMLElement => {
	const text = (message ?? '').trim();
	const lines = text.split('\n');
	let summary = lines[0].trim();
	let details = lines.slice(1).join('\n').trim();
	// A one-line error that carries a trace ("...: goroutine 1 [running]: ...") keeps only its message up front.
	const trace = summary.search(/\s(goroutine \d+ \[|\tat |github\.com\/)/);
	if (trace > 0) {
		details = `${summary.slice(trace).trim()}\n${details}`.trim();
		summary = summary.slice(0, trace).replace(/[:\s]+$/, '');
	}
	if (summary.length > 240) {
		details = `${summary}\n${details}`.trim();
		summary = `${summary.slice(0, 200).trimEnd()}…`;
	}
	return (
		<div className="rs-error-status">
			<span className="rs-error-summary">{summary ? `${prefix}: ${summary}` : `${prefix}.`}</span>
			{details ? (
				<details className="rs-error-details">
					<summary>Details</summary>
					<pre>{details}</pre>
				</details>
			) : null}
		</div>
	) as HTMLElement;
};

// What we know about a build id or a preset's name: its class, its name and its icon. A result can
// name a build this page's data doesn't have (a newer report), so it then shows the id.
export const memberInfo = (data: RaidSimData | null, id: string): { label: string; detail: string; cls: Class; icon: string } => {
	const found = data?.byId.get(id);
	if (found) {
		const { build, variant } = found;
		return {
			label: buildLabel(build, variant),
			detail: id,
			cls: build.class,
			icon: CLASS_ICONS[build.class] ?? '',
		};
	}
	const preset = presetNamed(id);
	if (preset) return { label: preset.name, detail: 'placeholder', cls: preset.class, icon: preset.icon };
	return { label: id, detail: '', cls: Class.ClassUnknown, icon: '' };
};

// An icon's alt text: its class, since the name beside it says the rest.
export const iconAlt = (cls: Class) => (cls == Class.ClassUnknown ? '' : classNames[cls] ?? '');

const presetNamed = (name: string) => RAID_PRESETS.find(p => p.name == name);

// A member's name in its class color, with the report id in its tooltip.
export const memberLabel = (data: RaidSimData | null, id: string, opts: { icon?: boolean; count?: number; seat?: string } = {}) => {
	const info = memberInfo(data, id);
	const cls = info.cls == Class.ClassUnknown ? '' : `text-${cssClassForClass(info.cls)}`;
	return (
		<span className={`rs-member-label ${cls}`} title={info.detail}>
			{opts.icon && info.icon ? <img className="rs-member-icon" src={info.icon} alt={iconAlt(info.cls)} /> : null}
			{info.label}
			{opts.count && opts.count > 1 ? <span className="rs-member-count">×{opts.count}</span> : null}
			{opts.seat ? <span className="rs-member-seat">{opts.seat}</span> : null}
		</span>
	);
};

// A list of build ids, as the result writes several providers: "a, b".
export const membersLabel = (data: RaidSimData | null, ids: string) => {
	const parts = ids
		.split(', ')
		.filter(s => s)
		.map(id => (data?.byId.has(id) || presetNamed(id) ? memberLabel(data, id) : document.createTextNode(id)));
	const out: Node[] = [];
	parts.forEach((p, i) => {
		if (i > 0) out.push(document.createTextNode(', '));
		out.push(p);
	});
	return <span>{out}</span>;
};

// A collapsible list: the first n rows, and a button for the rest.
export const collapsible = (build: (rows: number) => HTMLElement, total: number, first: number) => {
	if (total <= first) return build(total);
	const host = (<div className="rs-collapsible" />) as HTMLElement;
	let open = false;
	const render = () => {
		const button = (
			<button className="btn btn-link rs-show-more" type="button">
				{open ? 'Show fewer' : `Show all ${total}`}
			</button>
		) as HTMLButtonElement;
		button.addEventListener('click', () => {
			open = !open;
			render();
		});
		host.replaceChildren(build(open ? total : first), button);
	};
	render();
	return host;
};

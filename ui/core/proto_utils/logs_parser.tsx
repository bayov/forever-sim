import { RaidSimResult, ResourceType } from '../proto/api.js';
import { bucket, getEnumValues, stringComparator, sum } from '../utils.js';
import { ActionId } from './action_id.js';
import { resourceNames, stringToResourceType } from './names.js';

export class Entity {
	readonly name: string;
	readonly ownerName: string; // Blank if not a pet.

	// Either target index, player index, or owner index depending on what kind
	// of entity this is.
	readonly index: number;

	readonly isTarget: boolean;
	readonly isPet: boolean;

	constructor(name: string, ownerName: string, index: number, isTarget: boolean, isPet: boolean) {
		this.name = name;
		this.ownerName = ownerName;
		this.index = index;
		this.isTarget = isTarget;
		this.isPet = isPet;
	}

	// A player, not a target or a pet.
	get isPlayer(): boolean {
		return !this.isTarget && !this.isPet;
	}

	equals(other: Entity) {
		return this.isTarget == other.isTarget && this.isPet == other.isPet && this.index == other.index && this.name == other.name;
	}

	toString(): string {
		if (this.isTarget) {
			return `Target ${this.index + 1}`;
		} else if (this.isPet) {
			return `${this.ownerName} (#${this.index + 1}) - ${this.name}`;
		} else {
			return `${this.name} (#${this.index + 1})`;
		}
	}

	toHTML() {
		if (this.isTarget) {
			return <span className="text-danger">[Target {this.index + 1}]</span>;
		} else if (this.isPet) {
			return (
				<>
					<span className="text-primary">{`[${this.ownerName} ${this.index + 1}]`}</span>
					{` - `}
					{this.name}
				</>
			);
		} else {
			return <span className="text-primary">{`[${this.name} ${this.index + 1}]`}</span>;
		}
	}

	// Parses one or more Entities from a string.
	// Each entity label should be one of:
	//   'Target 1' if a target,
	//   'PlayerName (#1)' if a player, or
	//   'PlayerName (#1) - PetName' if a pet.
	static parseRegex = /\[(Target (\d+))|(([a-zA-Z0-9]+) \(#(\d+)\) - ([a-zA-Z0-9\s]+))|(([a-zA-Z0-9\s]+) \(#(\d+)\))\]/g;
	static parseAll(str: string): Array<Entity> {
		return Array.from(str.matchAll(Entity.parseRegex)).map(match => {
			if (match[1]) {
				return new Entity(match[1], '', parseInt(match[2]) - 1, true, false);
			} else if (match[3]) {
				return new Entity(match[6], match[4], parseInt(match[5]) - 1, false, true);
			} else if (match[7]) {
				return new Entity(match[8], '', parseInt(match[9]) - 1, false, false);
			} else {
				throw new Error('Invalid Entity match');
			}
		});
	}
}

interface SimLogParams {
	raw: string;
	logIndex: number;
	timestamp: number;
	source: Entity | null;
	target: Entity | null;
	actionId: ActionId | null;
	threat: number;
}

export class SimLog {
	readonly raw: string;

	// Index of this log within the full log output.
	// When comparing timestamps this should be used instead of timestamp, because
	// timestamp is scraped from log text and doesn't have enough precision.
	readonly logIndex: number;

	// Time in seconds from the encounter start.
	readonly timestamp: number;

	readonly source: Entity | null;
	readonly target: Entity | null;
	readonly actionId: ActionId | null;

	// Amount of threat generated from this event. Note that not all events generate threat, so this will be 0.
	readonly threat: number;

	// Logs for auras that were active at this timestamp.
	// This is only filled if populateActiveAuras() is called.
	activeAuras: Array<AuraUptimeLog>;

	cachedHTML: Record<string | number, Element | null> = {};

	constructor(params: SimLogParams) {
		this.raw = params.raw;
		this.logIndex = params.logIndex;
		this.timestamp = params.timestamp;
		this.source = params.source;
		this.target = params.target;
		this.actionId = params.actionId;
		this.threat = params.threat;
		this.activeAuras = [];
	}

	toHTML(includeTimestamp = true) {
		return this.cacheOutput(includeTimestamp, () => {
			let html = <>{this.raw}</>;
			// Base logs already have the timestamp appended by default
			if (!includeTimestamp) {
				const regexp = /(\[[0-9.-]+\]) (\[[0-9a-zA-Z\s\-()#]+\])?(.*)/;
				if (this.raw.match(regexp)) {
					// TypeScript doesn't handle regex capture typing well
					const captureArr = regexp.exec(this.raw);

					if (captureArr && captureArr.length == 4) {
						html = <>{captureArr[3]}</>;
					}
				}
			}

			if (this.source) {
				html = (
					<>
						{this.source.toHTML()} {html}
					</>
				);
			}
			return html;
		});
	}

	toPrefix(includeTimestamp = true) {
		let prefix = '';
		if (includeTimestamp) {
			prefix = `[${this.timestamp.toFixed(2)}]`;
		}

		return (
			<>
				{prefix}
				{this.source?.toHTML()}
			</>
		);
	}

	formattedTimestamp(): string {
		const positiveTimestamp = Math.abs(this.timestamp);
		const minutes = Math.floor(positiveTimestamp / 60);
		const seconds = Math.floor(positiveTimestamp - minutes * 60);
		const milliseconds = ((positiveTimestamp - Math.floor(positiveTimestamp)) * 1000).toFixed();

		let formatted = `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}:${String(milliseconds).padStart(3, '0')}`;
		if (this.timestamp < 0) {
			formatted = `-${formatted}`;
		}
		return formatted;
	}

	protected newActionIdLink(actionId: ActionId, isAura?: boolean) {
		const iconElem = <span className="icon icon-sm me-1"></span>;
		const actionAnchor = (
			<a className="log-action" target="_blank">
				<span className="d-inline-flex align-items-center align-bottom">
					{iconElem} {actionId.name}
				</span>
			</a>
		);
		actionId.setBackground(iconElem as HTMLAnchorElement);
		actionId.setWowheadHref(actionAnchor as HTMLAnchorElement);
		actionId.setWowheadDataset(actionAnchor as HTMLAnchorElement, { useBuffAura: isAura });
		return actionAnchor;
	}

	static async parseAll(result: RaidSimResult): Promise<SimLog[]> {
		const lines = result.logs.split('\n');
		const logs = await SimLog.parseLines(lines);
		CastBeganLog.findCostResources(logs);
		return logs;
	}

	private static parseLines(lines: string[]): Promise<SimLog[]> {
		return Promise.all(
			lines.map((line, lineIndex) => {
				const params: SimLogParams = {
					raw: line,
					logIndex: lineIndex,
					timestamp: 0,
					source: null,
					target: null,
					actionId: null,
					threat: 0,
				};

				const threatMatch = line.match(/ \(Threat: (-?[0-9]+\.[0-9]+)\)/);
				if (threatMatch) {
					params.threat = parseFloat(threatMatch[1]);
					line = line.substring(0, threatMatch.index);
				}

				const match = line.match(/\[(-?[0-9]+\.[0-9]+)\]\w*(.*)/);
				if (!match || !match[1]) {
					return new SimLog(params);
				}

				params.timestamp = parseFloat(match[1]);
				const remainder = match[2];

				const entities = Entity.parseAll(remainder);
				params.source = entities[0] || null;
				params.target = entities[1] || null;

				// Order from most to least common to reduce number of checks.
				return (
					DamageDealtLog.parse(params) ||
					ResourceChangedLog.parse(params) ||
					AuraEventLog.parse(params) ||
					AuraStacksChangeLog.parse(params) ||
					MajorCooldownUsedLog.parse(params) ||
					CastBeganLog.parse(params) ||
					CastCompletedLog.parse(params) ||
					CastInterruptedLog.parse(params) ||
					StatChangeLog.parse(params) ||
					ExtraAttackLog.parse(params) ||
					CooldownLog.parse(params) ||
					MaxResourcesLog.parse(params) ||
					TotemLog.parse(params) ||
					Promise.resolve(new SimLog(params))
				);
			}),
		);
	}

	isDamageDealt(): this is DamageDealtLog {
		return this instanceof DamageDealtLog;
	}

	isResourceChanged(): this is ResourceChangedLog {
		return this instanceof ResourceChangedLog;
	}

	isAuraEvent(): this is AuraEventLog {
		return this instanceof AuraEventLog;
	}

	isAuraStacksChange(): this is AuraStacksChangeLog {
		return this instanceof AuraStacksChangeLog;
	}

	isMajorCooldownUsed(): this is MajorCooldownUsedLog {
		return this instanceof MajorCooldownUsedLog;
	}

	isCastBegan(): this is CastBeganLog {
		return this instanceof CastBeganLog;
	}

	isCastCompleted(): this is CastCompletedLog {
		return this instanceof CastCompletedLog;
	}

	isCastInterrupted(): this is CastInterruptedLog {
		return this instanceof CastInterruptedLog;
	}

	isStatChange(): this is StatChangeLog {
		return this instanceof StatChangeLog;
	}

	isExtraAttack(): this is ExtraAttackLog {
		return this instanceof ExtraAttackLog;
	}

	isCooldown(): this is CooldownLog {
		return this instanceof CooldownLog;
	}

	isMaxResources(): this is MaxResourcesLog {
		return this instanceof MaxResourcesLog;
	}

	isTotem(): this is TotemLog {
		return this instanceof TotemLog;
	}

	// Group events that happen at the same time.
	static groupDuplicateTimestamps<LogType extends SimLog>(logs: Array<LogType>): Array<Array<LogType>> {
		const grouped: Array<Array<LogType>> = [];
		let curGroup: Array<LogType> = [];

		logs.forEach(log => {
			if (curGroup.length == 0 || log.timestamp == curGroup[0].timestamp) {
				curGroup.push(log);
			} else {
				grouped.push(curGroup);
				curGroup = [log];
			}
		});
		if (curGroup.length > 0) {
			grouped.push(curGroup);
		}

		return grouped;
	}

	// We cache each line's markup and hand out a copy. The markup is often a fragment, and a
	// fragment empties when it goes into the page, so handing out the cached one again (when the
	// Log view draws a run a second time) would show an empty line.
	cacheOutput(cacheKey: string | number | boolean, fn: () => Element) {
		const key = String(cacheKey);
		if (!this.cachedHTML[key]) this.cachedHTML[key] = fn();
		return (this.cachedHTML[key] as Element).cloneNode(true) as Element;
	}
}

export class DamageDealtLog extends SimLog {
	readonly amount: number;
	readonly type: string;
	readonly miss: boolean;
	readonly hit: boolean;
	readonly crit: boolean;
	readonly crush: boolean;
	readonly glance: boolean;
	readonly dodge: boolean;
	readonly parry: boolean;
	readonly block: boolean;
	readonly blockedCrit: boolean;
	readonly tick: boolean;
	// The share of the damage a partial resist took, in percent: 25, 50 or 75 under Classic, a 10%
	// step under Forever, 0 for none.
	readonly resistPercent: number;

	constructor(
		params: SimLogParams,
		amount: number,
		type: string,
		miss: boolean,
		crit: boolean,
		crush: boolean,
		glance: boolean,
		dodge: boolean,
		parry: boolean,
		block: boolean,
		blockedCrit: boolean,
		tick: boolean,
		resistPercent: number,
	) {
		super(params);
		this.amount = amount;
		this.type = type;
		this.miss = miss;
		this.glance = glance;
		this.dodge = dodge;
		this.parry = parry;
		this.block = block;
		this.blockedCrit = blockedCrit;
		this.hit = !miss && !crit;
		this.crit = crit;
		this.crush = crush;
		this.tick = tick;
		this.resistPercent = resistPercent;
	}

	isDamage(): boolean {
		return this.type == 'damage';
	}

	isHealing(): boolean {
		return this.type == 'healing';
	}

	isShielding(): boolean {
		return this.type == 'shielding';
	}

	result() {
		return (
			<>
				{this.isHealing() ? `Heal ` : ''}
				{this.isShielding() ? `Shield ` : ''}
				{this.miss
					? 'Miss'
					: this.dodge
					? 'Dodge'
					: this.parry
					? 'Parry'
					: this.glance
					? 'Glance'
					: this.blockedCrit
					? 'Blocked Crit'
					: this.block
					? 'Block'
					: this.crit
					? 'Crit'
					: this.crush
					? 'Crush'
					: this.tick
					? 'Tick'
					: 'Hit'}
				{` `}
				{this.target?.toHTML() || ''}
				{!this.miss && !this.dodge && !this.parry ? (
					<>
						{` `}
						for <strong className="text-danger">{this.amount.toFixed(2)} damage</strong>
						{this.resistPercent > 0 ? <> ({this.resistPercent}% Resist)</> : ''}
						.
					</>
				) : (
					''
				)}
			</>
		);
	}

	toHTML(includeTimestamp = true) {
		return this.cacheOutput(includeTimestamp, () => {
			const threatPostfix = this.source?.isTarget ? '' : ` (${this.threat.toFixed(2)} Threat)`;
			return (
				<>
					{this.toPrefix(includeTimestamp)} {this.newActionIdLink(this.actionId!)} {this.result()}
					{threatPostfix}
				</>
			);
		});
	}

	static parse(params: SimLogParams): Promise<DamageDealtLog> | null {
		const match = params.raw.match(
			/] (.*?) (tick )?((Miss)|(Hit)|(Crit)|(Crush)|(Glance)|(Dodge)|(Parry)|(BlockedCrit)|(Block))( \((\d+)% Resist\))?( for (\d+\.\d+) ((damage)|(healing)|(shielding)))?/,
		);
		if (match) {
			return ActionId.fromLogString(match[1])
				.fill(params.source?.index, params.source?.isPlayer)
				.then(cause => {
					params.actionId = cause;

					// Note: You must change these indeces when adding or removing capture groups to the regex above
					const amount = match[16] ? parseFloat(match[16]) : 0;
					const type = match[17] || '';

					return new DamageDealtLog(
						params,
						amount,
						type,
						match[3] == 'Miss',
						match[3] == 'Crit',
						match[3] == 'Crush',
						match[3] == 'Glance',
						match[3] == 'Dodge',
						match[3] == 'Parry',
						match[3] == 'Block',
						match[3] == 'BlockedCrit',
						Boolean(match[2]) && match[2].includes('tick'),
						match[13] ? parseInt(match[13]) : 0,
					);
				});
		} else {
			return null;
		}
	}
}

export class DpsLog extends SimLog {
	readonly dps: number;

	// Damage events that occurred at the same time as this log.
	readonly damageLogs: Array<DamageDealtLog>;

	constructor(params: SimLogParams, dps: number, damageLogs: Array<DamageDealtLog>) {
		super(params);
		this.dps = dps;
		this.damageLogs = damageLogs;
	}

	static DPS_WINDOW = 15; // Window over which to calculate DPS.
	static DPS_MIN_WINDOW = 5;

	// The seconds we divide the window's damage by at a time in the fight.
	//
	// Before 15 sec into the fight, the window holds only the damage since the pull. Dividing by
	// the full 15 sec would draw the opening as a slow climb from 0, so we divide by the time so
	// far instead. We never divide by less than 5 sec, because the first hit or two would
	// otherwise read as thousands of DPS.
	static windowAt(timestamp: number): number {
		return Math.min(DpsLog.DPS_WINDOW, Math.max(DpsLog.DPS_MIN_WINDOW, timestamp));
	}

	static fromLogs(damageDealtLogs: Array<DamageDealtLog>): Array<DpsLog> {
		const groupedDamageLogs = SimLog.groupDuplicateTimestamps(damageDealtLogs);

		let curDamageLogs: Array<DamageDealtLog> = [];
		let curDamageTotal = 0;

		return groupedDamageLogs.map(ddLogGroup => {
			ddLogGroup.forEach(ddLog => {
				curDamageLogs.push(ddLog);
				curDamageTotal += ddLog.amount;
			});

			const newStartIdx = curDamageLogs.findIndex(curLog => {
				const inWindow = curLog.timestamp > ddLogGroup[0].timestamp - DpsLog.DPS_WINDOW;
				if (!inWindow) {
					curDamageTotal -= curLog.amount;
				}
				return inWindow;
			});
			if (newStartIdx == -1) {
				curDamageLogs = [];
			} else {
				curDamageLogs = curDamageLogs.slice(newStartIdx);
			}

			const dps = curDamageTotal / DpsLog.windowAt(ddLogGroup[0].timestamp);
			if (isNaN(dps)) {
				console.warn('NaN dps!');
			}

			return new DpsLog(
				{
					raw: '',
					logIndex: ddLogGroup[0].logIndex,
					timestamp: ddLogGroup[0].timestamp,
					source: ddLogGroup[0].source,
					target: null,
					actionId: null,
					threat: 0,
				},
				dps,
				ddLogGroup,
			);
		});
	}
}

export class ThreatLogGroup extends SimLog {
	readonly threatBefore: number;
	readonly threatAfter: number;
	readonly logs: Array<SimLog>;

	constructor(params: SimLogParams, threatBefore: number, threatAfter: number, logs: Array<SimLog>) {
		super(params);
		this.threatBefore = threatBefore;
		this.threatAfter = threatAfter;
		this.logs = logs;
	}

	static fromLogs(logs: Array<SimLog>): Array<ThreatLogGroup> {
		const groupedLogs = SimLog.groupDuplicateTimestamps(logs.filter(log => log.threat != 0));
		let curThreat = 0;
		return groupedLogs.map(logGroup => {
			const newThreat = sum(logGroup.map(log => log.threat));
			const threatLog = new ThreatLogGroup(
				{
					raw: '',
					logIndex: logGroup[0].logIndex,
					timestamp: logGroup[0].timestamp,
					source: logGroup[0].source,
					target: logGroup[0].target,
					actionId: null,
					threat: newThreat,
				},
				curThreat,
				curThreat + newThreat,
				logGroup,
			);

			curThreat += newThreat;
			return threatLog;
		});
	}
}

export class AuraEventLog extends SimLog {
	readonly isGained: boolean;
	readonly isFaded: boolean;
	readonly isRefreshed: boolean;
	// When a gained or refreshed aura runs out, if it does. Results simmed before the sim
	// wrote it don't have it.
	readonly expiresAt: number | null;

	constructor(params: SimLogParams, isGained: boolean, isFaded: boolean, isRefreshed: boolean, expiresAt: number | null = null) {
		super(params);
		this.isGained = isGained;
		this.isFaded = isFaded;
		this.isRefreshed = isRefreshed;
		this.expiresAt = expiresAt;
	}

	toHTML(includeTimestamp = true) {
		return this.cacheOutput(includeTimestamp, () => (
			<>
				{this.toPrefix(includeTimestamp)}
				{`  Aura  `}
				{this.isGained ? 'gained' : this.isFaded ? 'faded' : 'refreshed'}: {this.newActionIdLink(this.actionId!, true)}.
			</>
		));
	}

	static parse(params: SimLogParams): Promise<AuraEventLog> | null {
		const match = params.raw.match(/Aura ((gained)|(faded)|(refreshed)): (.*)/);
		if (match && match[5]) {
			return ActionId.fromLogString(match[5])
				.fill(params.source?.index, params.source?.isPlayer)
				.then(aura => {
					params.actionId = aura;
					const event = match[1];
					const expires = match[5].match(/\(expires at (-?\d+\.?\d*)\)/);
					return new AuraEventLog(params, event == 'gained', event == 'faded', event == 'refreshed', expires ? parseFloat(expires[1]) : null);
				});
		} else {
			return null;
		}
	}
}

export class AuraStacksChangeLog extends SimLog {
	readonly oldStacks: number;
	readonly newStacks: number;

	constructor(params: SimLogParams, oldStacks: number, newStacks: number) {
		super(params);
		this.oldStacks = oldStacks;
		this.newStacks = newStacks;
	}

	toHTML(includeTimestamp = true) {
		return this.cacheOutput(includeTimestamp, () => (
			<>
				{this.toPrefix(includeTimestamp)} {this.newActionIdLink(this.actionId!, true)} stacks: {this.oldStacks} &rarr; {this.newStacks}.
			</>
		));
	}

	static parse(params: SimLogParams): Promise<AuraStacksChangeLog> | null {
		const match = params.raw.match(/(.*) stacks: ([0-9]+) --> ([0-9]+)/);
		if (match && match[1]) {
			return ActionId.fromLogString(match[1])
				.fill(params.source?.index, params.source?.isPlayer)
				.then(aura => {
					params.actionId = aura;
					return new AuraStacksChangeLog(params, parseInt(match[2]), parseInt(match[3]));
				});
		} else {
			return null;
		}
	}
}

export class AuraUptimeLog extends SimLog {
	readonly gainedAt: number;
	readonly fadedAt: number;
	readonly stacksChange: Array<AuraStacksChangeLog>;
	// When the aura was set to run out as it was gained or refreshed, if the log says.
	readonly expiresAt: number | null;

	constructor(params: SimLogParams, fadedAt: number, stacksChange: Array<AuraStacksChangeLog>, expiresAt: number | null = null) {
		super(params);
		this.gainedAt = params.timestamp;
		this.fadedAt = fadedAt;
		this.stacksChange = stacksChange;
		this.expiresAt = expiresAt;
	}

	static fromLogs(logs: Array<SimLog>, entity: Entity, encounterDuration: number): Array<AuraUptimeLog> {
		const unmatchedGainedLogs: Array<{ gained: AuraEventLog; stacks: Array<AuraStacksChangeLog> }> = [];
		const uptimeLogs: Array<AuraUptimeLog> = [];

		logs.forEach((log: SimLog) => {
			if (!log.source || !log.source.equals(entity)) {
				return;
			}

			if (log.isAuraStacksChange()) {
				const matchingGainedIdx = unmatchedGainedLogs.findIndex(gainedLog => gainedLog.gained.actionId!.equals(log.actionId!));
				if (matchingGainedIdx == -1) {
					console.warn('Unmatched aura stacks change log: ' + log.actionId!.name);
					return;
				}
				unmatchedGainedLogs[matchingGainedIdx].stacks.push(log);
				return;
			}

			if (!log.isAuraEvent()) {
				return;
			}

			if (log.isGained) {
				unmatchedGainedLogs.push({ gained: log, stacks: [] });
				return;
			}

			const matchingGainedIdx = unmatchedGainedLogs.findIndex(gainedLog => gainedLog.gained.actionId!.equals(log.actionId!));
			if (matchingGainedIdx == -1) {
				console.warn('Unmatched aura faded log: ' + log.actionId!.name);
				return;
			}
			const { gained: gainedLog, stacks: stacksChangeLogs } = unmatchedGainedLogs.splice(matchingGainedIdx, 1)[0];

			uptimeLogs.push(
				new AuraUptimeLog(
					{
						raw: log.raw,
						logIndex: gainedLog.logIndex,
						timestamp: gainedLog.timestamp,
						source: log.source,
						target: log.target,
						actionId: gainedLog.actionId,
						threat: gainedLog.threat,
					},
					log.timestamp,
					stacksChangeLogs,
					gainedLog.expiresAt,
				),
			);

			if (log.isRefreshed) {
				unmatchedGainedLogs.push({ gained: log, stacks: [] });
			}
		});

		// Auras active at the end won't have a faded log, so need to add them separately.
		unmatchedGainedLogs.forEach(unmatchedLog => {
			const { gained: gainedLog, stacks: stacksChangeLogs } = unmatchedLog;
			uptimeLogs.push(
				new AuraUptimeLog(
					{
						raw: gainedLog.raw,
						logIndex: gainedLog.logIndex,
						timestamp: gainedLog.timestamp,
						source: gainedLog.source,
						target: gainedLog.target,
						actionId: gainedLog.actionId,
						threat: gainedLog.threat,
					},
					encounterDuration,
					stacksChangeLogs,
					gainedLog.expiresAt,
				),
			);
		});

		uptimeLogs.sort((a, b) => a.gainedAt - b.gainedAt);
		return uptimeLogs;
	}

	// Populates the activeAuras field for all logs using the provided auras.
	static populateActiveAuras(logs: Array<SimLog>, auraLogs: Array<AuraUptimeLog>) {
		let curAuras: Array<AuraUptimeLog> = [];
		let auraLogsIndex = 0;

		logs.forEach(log => {
			while (auraLogsIndex < auraLogs.length && auraLogs[auraLogsIndex].gainedAt <= log.timestamp) {
				curAuras.push(auraLogs[auraLogsIndex]);
				auraLogsIndex++;
			}
			curAuras = curAuras.filter(curAura => curAura.fadedAt > log.timestamp);

			const activeAuras = curAuras.slice();
			activeAuras.sort((a, b) => stringComparator(a.actionId!.name, b.actionId!.name));
			log.activeAuras = activeAuras;
		});
	}
}

export class ResourceChangedLog extends SimLog {
	readonly resourceType: ResourceType;
	readonly valueBefore: number;
	readonly valueAfter: number;
	readonly isSpend: boolean;

	constructor(params: SimLogParams, resourceType: ResourceType, valueBefore: number, valueAfter: number, isSpend: boolean) {
		super(params);
		this.resourceType = resourceType;
		this.valueBefore = valueBefore;
		this.valueAfter = valueAfter;
		this.isSpend = isSpend;
	}

	toHTML(includeTimestamp = true) {
		return this.cacheOutput(includeTimestamp, () => {
			const signedDiff = (this.valueAfter - this.valueBefore) * (this.isSpend ? -1 : 1);
			const isHealth = this.resourceType == ResourceType.ResourceTypeHealth;
			const verb = isHealth ? (this.isSpend ? 'Lost' : 'Recovered') : this.isSpend ? 'Spent' : 'Gained';
			const resourceName = resourceNames.get(this.resourceType)!;
			const resourceClass = `resource-${resourceName.replace(/\s/g, '-').toLowerCase()}`;
			
			return (
				<>
					{this.toPrefix(includeTimestamp)} {verb}{' '}
					<strong className={resourceClass}>
						{signedDiff.toFixed(1)} {resourceName}
					</strong>
					{this.target ? <>{` on `} {this.target?.toHTML()}</> : null}
					{` from `}
					{this.newActionIdLink(this.actionId!)}. 
					({this.valueBefore.toFixed(1)} &rarr; {this.valueAfter.toFixed(1)})
				</>
			);
		});
	}

	resultString(): string {
		const delta = this.valueAfter - this.valueBefore;
		if (delta < 0) {
			return delta.toFixed(1);
		} else {
			return '+' + delta.toFixed(1);
		}
	}

	static parse(params: SimLogParams): Promise<ResourceChangedLog> | null {
		// ((Gained)|(Spent)) \d+\.?\d* ((health)|(mana)|(energy)|(focus)|(rage)|(combo points)) from (\{.*?\}) (?:on (\[[\w ]+\]) )?\((\d+\.?\d*) --> (\d+\.?\d*)\)

		const match = params.raw.match(
			/((Gained)|(Spent)) \d+\.?\d* ((health)|(mana)|(energy)|(focus)|(rage)|(combo points)) (?:on (\[[\w ]+\]) )?from (\{.*?\}) \((\d+\.?\d*) --> (\d+\.?\d*)\)/,
		);
		if (match) {
			const resourceType = stringToResourceType(match[4]);
			// combo points have an additional Target included in its log lines, adjust for that here
			const valueBefore = match[13]
			const valueAfter = match[14];
			return ActionId.fromLogString(match[12])
				.fill(params.source?.index, params.source?.isPlayer)
				.then(cause => {
					params.actionId = cause;
					return new ResourceChangedLog(params, resourceType, parseFloat(valueBefore), parseFloat(valueAfter), match[1] == 'Spent');
				});
		} else {
			return null;
		}
	}
}

export class ResourceChangedLogGroup extends SimLog {
	readonly resourceType: ResourceType;
	readonly valueBefore: number;
	readonly valueAfter: number;
	readonly logs: Array<ResourceChangedLog>;

	constructor(params: SimLogParams, resourceType: ResourceType, valueBefore: number, valueAfter: number, logs: Array<ResourceChangedLog>) {
		super(params);
		this.resourceType = resourceType;
		this.valueBefore = valueBefore;
		this.valueAfter = valueAfter;
		this.logs = logs;
	}

	toHTML(includeTimestamp = true) {
		return this.cacheOutput(includeTimestamp, () => (
			<>
				{this.toPrefix(includeTimestamp)} {resourceNames.get(this.resourceType)}: {this.valueBefore.toFixed(1)} &rarr; {this.valueAfter.toFixed(1)}
			</>
		));
	}

	static fromLogs(logs: Array<SimLog>): Record<ResourceType, Array<ResourceChangedLogGroup>> {
		const allResourceChangedLogs = logs.filter((log): log is ResourceChangedLog => log.isResourceChanged());

		const results: Partial<Record<ResourceType, Array<ResourceChangedLogGroup>>> = {};
		const resourceTypes = (getEnumValues(ResourceType) as Array<ResourceType>).filter(val => val != ResourceType.ResourceTypeNone);
		resourceTypes.forEach(resourceType => {
			const resourceChangedLogs = allResourceChangedLogs.filter(log => log.resourceType == resourceType);

			const groupedLogs = SimLog.groupDuplicateTimestamps(resourceChangedLogs);
			results[resourceType] = groupedLogs.map(
				logGroup =>
					new ResourceChangedLogGroup(
						{
							raw: '',
							logIndex: logGroup[0].logIndex,
							timestamp: logGroup[0].timestamp,
							source: logGroup[0].source,
							target: logGroup[0].target,
							actionId: null,
							threat: 0,
						},
						resourceType,
						logGroup[0].valueBefore,
						logGroup[logGroup.length - 1].valueAfter,
						logGroup,
					),
			);
		});

		return results as Record<ResourceType, Array<ResourceChangedLogGroup>>;
	}
}

export class MajorCooldownUsedLog extends SimLog {
	constructor(params: SimLogParams) {
		super(params);
	}

	toHTML(includeTimestamp = true) {
		return this.cacheOutput(includeTimestamp, () => (
			<>
				{this.toPrefix(includeTimestamp)} Major cooldown used: {this.newActionIdLink(this.actionId!)}.
			</>
		));
	}

	static parse(params: SimLogParams): Promise<MajorCooldownUsedLog> | null {
		const match = params.raw.match(/Major cooldown used: (.*)/);
		if (match) {
			return ActionId.fromLogString(match[1])
				.fill(params.source?.index, params.source?.isPlayer)
				.then(cooldownId => {
					params.actionId = cooldownId;
					return new MajorCooldownUsedLog(params);
				});
		} else {
			return null;
		}
	}
}

export class CastBeganLog extends SimLog {
	readonly manaCost: number;
	readonly castTime: number;
	readonly effectiveTime: number;
	// What the cost is paid in, from the cast's spend line, see findCostResources. Null when the
	// cast spent nothing.
	costResource: ResourceType | null = null;

	constructor(params: SimLogParams, manaCost: number, castTime: number, effectiveTime: number) {
		super(params);
		this.manaCost = manaCost;
		this.castTime = castTime;
		this.effectiveTime = effectiveTime;
	}

	toHTML(includeTimestamp = true) {
		return this.cacheOutput(includeTimestamp, () => (
			<>
				{this.toPrefix(includeTimestamp)} Casting {this.newActionIdLink(this.actionId!)} (Cast time: {this.castTime.toFixed(2)}s, {this.costText()}).
			</>
		));
	}

	private costText(): string {
		if (!this.manaCost) return 'no cost';
		const resource = this.costResource === null ? '' : ` ${resourceNames.get(this.costResource)}`;
		return `Cost: ${this.manaCost.toFixed(1)}${resource}`;
	}

	// The sim's cast line has the cost but not what it's paid in, so a warrior's Whirlwind would
	// read "22.0 Mana". The spend comes in its own line ("Spent 22.0 rage from Whirlwind"), so we
	// take the resource from there. An instant spends at the same moment, next to the cast line.
	// A cast with a cast time spends when it finishes. A finisher also spends its combo points,
	// and the cost is the energy.
	static findCostResources(logs: SimLog[]) {
		logs.forEach((log, i) => {
			if (!log.isCastBegan() || !log.manaCost) return;
			// The log rounds times to 0.01s.
			const end = log.timestamp + Math.max(log.castTime, log.effectiveTime) + 0.011;
			for (let j = Math.max(0, i - 8); j < logs.length && (j <= i || logs[j].timestamp <= end); j++) {
				const other = logs[j];
				if (
					other.isResourceChanged() &&
					other.isSpend &&
					other.timestamp >= log.timestamp &&
					other.resourceType != ResourceType.ResourceTypeComboPoints &&
					!!log.source &&
					!!other.source?.equals(log.source) &&
					other.actionId?.equalsIgnoringTag(log.actionId!)
				) {
					log.costResource = other.resourceType;
					return;
				}
			}
		});
	}

	static parse(params: SimLogParams): Promise<CastBeganLog> | null {
		const match = params.raw.match(/Casting (.*) \(Cost = (\d+\.?\d*), Cast Time = (\d+\.?\d*)(m?s), Effective Time = (\d+\.?\d*)(m?s)\)/);
		if (match) {
			let castTime = parseFloat(match[3]);
			if (match[4] == 'ms') {
				castTime /= 1000;
			}
			let effectiveTime = parseFloat(match[5]);
			if (match[6] == 'ms') {
				effectiveTime /= 1000;
			}
			return ActionId.fromLogString(match[1])
				.fill(params.source?.index, params.source?.isPlayer)
				.then(castId => {
					params.actionId = castId;
					return new CastBeganLog(params, parseFloat(match[2]), castTime, effectiveTime);
				});
		} else {
			return null;
		}
	}
}

export class CastCompletedLog extends SimLog {
	constructor(params: SimLogParams) {
		super(params);
	}

	toHTML(includeTimestamp = true) {
		return this.cacheOutput(includeTimestamp, () => (
			<>
				{this.toPrefix(includeTimestamp)} Completed cast {this.actionId!.name}.
			</>
		));
	}

	static parse(params: SimLogParams): Promise<CastCompletedLog> | null {
		const match = params.raw.match(/Completed cast (.*)/);
		if (match) {
			return ActionId.fromLogString(match[1])
				.fill(params.source?.index, params.source?.isPlayer)
				.then(castId => {
					params.actionId = castId;
					return new CastCompletedLog(params);
				});
		} else {
			return null;
		}
	}
}

// A hard cast cut off before it finished, like "Cast of {SpellID: 8008} interrupted" when we
// start moving or die.
export class CastInterruptedLog extends SimLog {
	constructor(params: SimLogParams) {
		super(params);
	}

	toHTML(includeTimestamp = true) {
		return this.cacheOutput(includeTimestamp, () => (
			<>
				{this.toPrefix(includeTimestamp)} Cast of {this.newActionIdLink(this.actionId!)} interrupted.
			</>
		));
	}

	static parse(params: SimLogParams): Promise<CastInterruptedLog> | null {
		const match = params.raw.match(/Cast of ({.*?}) interrupted/);
		if (!match) return null;
		return ActionId.fromLogString(match[1])
			.fill(params.source?.index, params.source?.isPlayer)
			.then(castId => {
				params.actionId = castId;
				return new CastInterruptedLog(params);
			});
	}
}

export class CastLog extends SimLog {
	readonly castTime: number;
	readonly effectiveTime: number;
	readonly travelTime: number;

	readonly castBeganLog: CastBeganLog;
	readonly castCompletedLog: CastCompletedLog | null;
	// Set when the cast was cut off before it finished, like by moving or dying.
	readonly castInterruptedLog: CastInterruptedLog | null;

	// All instances of damage dealt from the completion of this cast until the completion of the next cast.
	readonly damageDealtLogs: Array<DamageDealtLog>;

	constructor(
		castBeganLog: CastBeganLog,
		castCompletedLog: CastCompletedLog | null,
		damageDealtLogs: Array<DamageDealtLog>,
		castInterruptedLog: CastInterruptedLog | null = null,
	) {
		super({
			raw: castBeganLog.raw,
			logIndex: castBeganLog.logIndex,
			timestamp: castBeganLog.timestamp,
			source: castBeganLog.source,
			target: castBeganLog.target,
			actionId: castCompletedLog?.actionId || castBeganLog.actionId, // Use completed log because of arcane blast
			threat: castCompletedLog?.threat || castBeganLog.threat,
		});
		this.castTime = castBeganLog.castTime;
		this.effectiveTime = castBeganLog.effectiveTime;
		this.castBeganLog = castBeganLog;
		this.castCompletedLog = castCompletedLog;
		this.castInterruptedLog = castInterruptedLog;
		this.damageDealtLogs = damageDealtLogs;

		if (this.castCompletedLog && this.castBeganLog) {
			this.castTime = this.castCompletedLog.timestamp - this.castBeganLog.timestamp;
		} else if (this.castInterruptedLog) {
			this.castTime = this.castInterruptedLog.timestamp - this.castBeganLog.timestamp;
		}
		if (
			this.castCompletedLog &&
			this.damageDealtLogs.length == 1 &&
			this.castCompletedLog.timestamp < this.damageDealtLogs[0].timestamp &&
			!this.damageDealtLogs[0].tick
		) {
			this.travelTime = this.damageDealtLogs[0].timestamp - this.castCompletedLog.timestamp;
		} else {
			this.travelTime = 0;
		}
	}

	toHTML(includeTimestamp = true) {
		return this.cacheOutput(includeTimestamp, () => (
			<>
				{this.toPrefix(includeTimestamp)} Casting {this.actionId!.name} (Cast time = {this.castTime.toFixed(2)}s).
			</>
		));
	}

	totalDamage(): number {
		return sum(this.damageDealtLogs.map(ddl => ddl.amount));
	}

	static fromLogs(logs: Array<SimLog>): Array<CastLog> {
		const castBeganLogs = logs.filter((log): log is CastBeganLog => log.isCastBegan());
		const castCompletedLogs = logs.filter((log): log is CastCompletedLog => log.isCastCompleted());
		const castInterruptedLogs = logs.filter((log): log is CastInterruptedLog => log.isCastInterrupted());
		const damageDealtLogs = logs.filter((log): log is DamageDealtLog => log.isDamageDealt());

		const toBucketKey = (actionId: ActionId) => {
			if (actionId.spellId == 30451) {
				// Arcane Blast is unique because it can finish its cast as a different
				// spell than it started (if stacks drop).
				return actionId.toStringIgnoringTag();
			} else {
				return actionId.toString();
			}
		};
		const castBeganLogsByAbility = bucket(castBeganLogs, log => toBucketKey(log.actionId!));
		const castCompletedLogsByAbility = bucket(castCompletedLogs, log => toBucketKey(log.actionId!));
		const castInterruptedLogsByAbility = bucket(castInterruptedLogs, log => toBucketKey(log.actionId!));
		const damageDealtLogsByAbility = bucket(damageDealtLogs, log => toBucketKey(log.actionId!));

		const castLogs: Array<CastLog> = [];
		Object.keys(castBeganLogsByAbility).forEach(bucketKey => {
			const abilityCastsBegan = castBeganLogsByAbility[bucketKey]!;
			const abilityCastsCompleted = castCompletedLogsByAbility[bucketKey];
			const abilityDamageDealt = damageDealtLogsByAbility[bucketKey];

			// A cast that was interrupted, like a Lesser Healing Wave cut off when we died, never
			// completes. It ends at its "Cast of ... interrupted" line, which belongs to the last cast
			// of the ability that began before it. Without this, each later cast took the completion
			// of the one after it, and the interrupted cast lasted until the next one completed.
			//
			// Once in a while a cast also ends with neither line, and the next cast of the ability
			// begins later on (we've seen it with Lesser Healing Wave in a leveling session). It gets
			// no completion either, so it shows for its cast time.
			const endedEarly = new Map<CastBeganLog, CastInterruptedLog | null>();
			(castInterruptedLogsByAbility[bucketKey] ?? []).forEach(interruptedLog => {
				let began: CastBeganLog | undefined;
				for (const log of abilityCastsBegan) {
					if (log.logIndex > interruptedLog.logIndex) break;
					began = log;
				}
				if (began && !endedEarly.has(began)) endedEarly.set(began, interruptedLog);
			});
			const completedIndexes = (abilityCastsCompleted ?? []).map(log => log.logIndex);
			let nextCompleted = 0;
			abilityCastsBegan.forEach((log, i) => {
				while (nextCompleted < completedIndexes.length && completedIndexes[nextCompleted] < log.logIndex) nextCompleted++;
				const next = abilityCastsBegan[i + 1];
				// Two casts that begin at the same moment, like a double extra attack, may log both
				// begins before the completions, so those keep pairing in order.
				if (!next || endedEarly.has(log) || next.timestamp <= log.timestamp) return;
				const completesBeforeNext = nextCompleted < completedIndexes.length && completedIndexes[nextCompleted] < next.logIndex;
				if (!completesBeforeNext) endedEarly.set(log, null);
			});

			let ddIdx = 0;
			// The other casts pair with the completed logs in order.
			let completedIdx = 0;
			for (let cbIdx = 0; cbIdx < abilityCastsBegan.length; cbIdx++) {
				const cbLog = abilityCastsBegan[cbIdx];
				if (endedEarly.has(cbLog)) {
					castLogs.push(new CastLog(cbLog, null, [], endedEarly.get(cbLog)));
					continue;
				}

				// Assume cast completed log is the same index because they always come in pairs.
				// Only exception is final pair, where there might be a cast began without a cast completed.
				let ccLog: CastCompletedLog | null = null;
				let nextCcLog: CastCompletedLog | null = null;
				if (abilityCastsCompleted && completedIdx < abilityCastsCompleted.length) {
					ccLog = abilityCastsCompleted[completedIdx];
					if (completedIdx + 1 < abilityCastsCompleted.length) {
						nextCcLog = abilityCastsCompleted[completedIdx + 1];
					}
				}
				completedIdx++;

				// Find all damage dealt logs between the cur and next cast completed logs.
				const ddLogs = [];
				while (abilityDamageDealt && ddIdx < abilityDamageDealt.length && (!nextCcLog || abilityDamageDealt[ddIdx].timestamp < nextCcLog.timestamp)) {
					ddLogs.push(abilityDamageDealt[ddIdx]);
					ddIdx++;
				}
				castLogs.push(new CastLog(cbLog, ccLog, ddLogs));
			}
		});

		castLogs.sort((a, b) => a.timestamp - b.timestamp);
		return castLogs;
	}
}

export class StatChangeLog extends SimLog {
	readonly isGain: boolean;
	readonly stats: string;

	constructor(params: SimLogParams, isGain: boolean, stats: string) {
		super(params);
		this.isGain = isGain;
		this.stats = stats;
	}

	toHTML(includeTimestamp = true) {
		return this.cacheOutput(includeTimestamp, () => {
			if (this.isGain) {
				return (
					<>
						{this.toPrefix(includeTimestamp)} Gained {this.stats} from {this.newActionIdLink(this.actionId!)}.
					</>
				);
			} else {
				return (
					<>
						{this.toPrefix(includeTimestamp)} Lost {this.stats} from fading {this.newActionIdLink(this.actionId!)}.
					</>
				);
			}
		});
	}

	static parse(params: SimLogParams): Promise<StatChangeLog> | null {
		const match = params.raw.match(/((Gained)|(Lost)) ({.*}) from (fading )?(.*)/);
		if (match) {
			return ActionId.fromLogString(match[6])
				.fill(params.source?.index, params.source?.isPlayer)
				.then(effectId => {
					params.actionId = effectId;
					const sign = match[1] == 'Lost' ? -1 : 1;
					return new StatChangeLog(params, sign == 1, match[4]);
				});
		} else {
			return null;
		}
	}
}

export class ExtraAttackLog extends SimLog {
	readonly attacks: number;
	readonly attacksText: string;
	readonly attackID: ActionId;
	readonly triggerID: ActionId;

	constructor(params: SimLogParams, attacks: number, attacksText: string, attackID: ActionId, triggerID: ActionId) {
		super(params);
		this.attacks = attacks;
		this.attacksText = attacksText;
		this.attackID = attackID;
		this.triggerID = triggerID;
	}

	toHTML(includeTimestamp = true) {
		return this.cacheOutput(includeTimestamp, () => {
			return (
				<>
					{this.toPrefix(includeTimestamp)} Gained {this.attacks} extra {this.attacksText} from {this.newActionIdLink(this.attackID)} triggered by{' '}
					{this.newActionIdLink(this.triggerID)}.
				</>
			);
		});
	}

	static parse(params: SimLogParams): Promise<ExtraAttackLog> | null {
		const match = params.raw.match(/gained ([0-9]+) extra (?:main-hand|off-hand|ranged) (attacks?) from ({.*?}) triggered by ({.*?})/);
		if (match) {
			return ActionId.fromLogString(match[3])
				.fill(params.source?.index, params.source?.isPlayer)
				.then(attackID => {
					return ActionId.fromLogString(match[4])
						.fill(params.source?.index, params.source?.isPlayer)
						.then(triggerID => {
							params.actionId = attackID;
							return new ExtraAttackLog(params, parseInt(match[1]), match[2], attackID, triggerID);
						});
				});
		} else {
			return null;
		}
	}
}

// A spell or item went on cooldown, like "{SpellID: 17364} on cooldown until 12.500 (timer 4)".
//
// Spells that share a cooldown log the same timer number. One cast can log more than one timer,
// like a trinket with its own cooldown and the cooldown it shares with the other trinkets.
export class CooldownLog extends SimLog {
	readonly readyAt: number;
	readonly timer: number;

	constructor(params: SimLogParams, readyAt: number, timer: number) {
		super(params);
		this.readyAt = readyAt;
		this.timer = timer;
	}

	toHTML(includeTimestamp = true) {
		return this.cacheOutput(includeTimestamp, () => (
			<>
				{this.toPrefix(includeTimestamp)} {this.newActionIdLink(this.actionId!)} on cooldown until {this.readyAt.toFixed(2)}s.
			</>
		));
	}

	static parse(params: SimLogParams): Promise<CooldownLog> | null {
		const match = params.raw.match(/] ({.*?}) on cooldown until (-?\d+\.?\d*) \(timer (-?\d+)\)/);
		if (!match) return null;
		return ActionId.fromLogString(match[1])
			.fill(params.source?.index, params.source?.isPlayer)
			.then(actionId => {
				params.actionId = actionId;
				return new CooldownLog(params, parseFloat(match[2]), parseInt(match[3]));
			});
	}
}

// The size of each of a unit's bars at the pull, like `Max resources: {"Health": 4210.000}`.
export class MaxResourcesLog extends SimLog {
	readonly maxValues: Partial<Record<ResourceType, number>>;

	constructor(params: SimLogParams, maxValues: Partial<Record<ResourceType, number>>) {
		super(params);
		this.maxValues = maxValues;
	}

	toHTML(includeTimestamp = true) {
		return this.cacheOutput(includeTimestamp, () => (
			<>
				{this.toPrefix(includeTimestamp)} Max{' '}
				{Object.entries(this.maxValues)
					.map(([type, value]) => `${resourceNames.get(Number(type) as ResourceType)?.toLowerCase()} ${value!.toFixed(0)}`)
					.join(', ')}
				.
			</>
		));
	}

	static parse(params: SimLogParams): Promise<MaxResourcesLog> | null {
		const match = params.raw.match(/Max resources: ({.*})/);
		if (!match) return null;
		const maxValues: Partial<Record<ResourceType, number>> = {};
		try {
			Object.entries(JSON.parse(match[1]) as Record<string, number>).forEach(([name, value]) => {
				const type = stringToResourceType(name);
				if (type != ResourceType.ResourceTypeNone) maxValues[type] = value;
			});
		} catch {
			return null;
		}
		return Promise.resolve(new MaxResourcesLog(params, maxValues));
	}
}

// A shaman's totem went down in its element's slot, like
// "{SpellID: 10438} down in the Fire slot until 55.000". A new totem in the slot takes the old
// one's place.
export class TotemLog extends SimLog {
	readonly element: string;
	readonly until: number;

	constructor(params: SimLogParams, element: string, until: number) {
		super(params);
		this.element = element;
		this.until = until;
	}

	toHTML(includeTimestamp = true) {
		return this.cacheOutput(includeTimestamp, () => (
			<>
				{this.toPrefix(includeTimestamp)} {this.newActionIdLink(this.actionId!)} down in the {this.element} slot until {this.until.toFixed(2)}s.
			</>
		));
	}

	static parse(params: SimLogParams): Promise<TotemLog> | null {
		const match = params.raw.match(/] ({.*?}) down in the (\w+) slot until (-?\d+\.?\d*)/);
		if (!match) return null;
		return ActionId.fromLogString(match[1])
			.fill(params.source?.index, params.source?.isPlayer)
			.then(actionId => {
				params.actionId = actionId;
				return new TotemLog(params, match[2], parseFloat(match[3]));
			});
	}
}

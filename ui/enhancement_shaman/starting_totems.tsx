import tippy from 'tippy.js';

import { EnumValueConfig } from '../core/components/enum_picker';
import { IconEnumRowPicker } from '../core/components/individual_sim_ui/player_pickers';
import { otherShamanWindfuryTotem } from '../core/components/inputs/buffs_debuffs';
import { NumberPicker } from '../core/components/number_picker';
import { markUnstacked } from '../core/components/unstacked_mark';
import { IndividualSimUI } from '../core/individual_sim_ui';
import { Player } from '../core/player';
import { APLPrepullAction, APLRotation_Type as APLRotationType } from '../core/proto/apl';
import { Spec } from '../core/proto/common';
import { AirTotem, EarthTotem, FireTotem, StartingTotems, WaterTotem } from '../core/proto/shaman';
import { ActionId } from '../core/proto_utils/action_id';
import { EventID, TypedEvent } from '../core/typed_event';

type EnhPlayer = Player<Spec.SpecEnhancementShaman>;

// A totem we can start the fight with, and the level each of its ranks is learned at. Keep the
// ranks in sync with sim/shaman/*_totems.go. The sim puts down the highest rank we know.
interface TotemOption {
	name: string;
	value: number;
	icon: string;
	ranks: Array<{ level: number; spellId: number }>;
}

interface TotemElement {
	name: string;
	field: 'earth' | 'air' | 'fire' | 'water';
	secondsField: 'earthSecondsBeforePull' | 'airSecondsBeforePull' | 'fireSecondsBeforePull' | 'waterSecondsBeforePull';
	// When we pick a totem for an element that had none, it went down this many seconds before
	// the pull. We put them down one per second, water first.
	defaultSecondsBeforePull: number;
	totems: Array<TotemOption>;
}

const ranks = (levels: Array<number>, spellIds: Array<number>) => levels.map((level, i) => ({ level, spellId: spellIds[i] }));

const ELEMENTS: Array<TotemElement> = [
	{
		name: 'Earth',
		field: 'earth',
		secondsField: 'earthSecondsBeforePull',
		defaultSecondsBeforePull: 28,
		totems: [
			{
				name: 'Strength of Earth Totem',
				value: EarthTotem.StrengthOfEarthTotem,
				icon: 'spell_nature_earthbindtotem',
				ranks: ranks([10, 24, 38, 52, 60], [8075, 8160, 8161, 10442, 25361]),
			},
			{
				name: 'Stoneskin Totem',
				value: EarthTotem.StoneskinTotem,
				icon: 'spell_nature_stoneskintotem',
				ranks: ranks([4, 14, 24, 34, 44, 54], [8071, 8154, 8155, 10406, 10407, 10408]),
			},
			{ name: 'Tremor Totem', value: EarthTotem.TremorTotem, icon: 'spell_nature_tremortotem', ranks: ranks([18], [8143]) },
		],
	},
	{
		name: 'Air',
		field: 'air',
		secondsField: 'airSecondsBeforePull',
		defaultSecondsBeforePull: 27,
		totems: [
			{ name: 'Windfury Totem', value: AirTotem.WindfuryTotem, icon: 'spell_nature_windfury', ranks: ranks([32, 42, 52], [8512, 10613, 10614]) },
			{
				name: 'Grace of Air Totem',
				value: AirTotem.GraceOfAirTotem,
				icon: 'spell_nature_invisibilitytotem',
				ranks: ranks([42, 56, 60], [8835, 10627, 25359]),
			},
		],
	},
	{
		name: 'Fire',
		field: 'fire',
		secondsField: 'fireSecondsBeforePull',
		defaultSecondsBeforePull: 29,
		totems: [
			{
				name: 'Searing Totem',
				value: FireTotem.SearingTotem,
				icon: 'spell_fire_searingtotem',
				ranks: ranks([10, 20, 30, 40, 50, 60], [3599, 6363, 6364, 6365, 10437, 10438]),
			},
			{ name: 'Magma Totem', value: FireTotem.MagmaTotem, icon: 'spell_fire_selfdestruct', ranks: ranks([26, 36, 46, 56], [8190, 10585, 10586, 10587]) },
			{
				name: 'Flametongue Totem',
				value: FireTotem.FlametongueTotem,
				icon: 'spell_nature_guardianward',
				ranks: ranks([28, 38, 48, 58], [8227, 8249, 10526, 16387]),
			},
		],
	},
	{
		name: 'Water',
		field: 'water',
		secondsField: 'waterSecondsBeforePull',
		defaultSecondsBeforePull: 30,
		totems: [
			{
				name: 'Mana Spring Totem',
				value: WaterTotem.ManaSpringTotem,
				icon: 'spell_nature_manaregentotem',
				ranks: ranks([26, 36, 46, 56], [5675, 10495, 10496, 10497]),
			},
			{
				name: 'Healing Stream Totem',
				value: WaterTotem.HealingStreamTotem,
				icon: 'inv_spear_04',
				ranks: ranks([20, 30, 40, 50, 60], [5394, 6375, 6377, 10462, 10463]),
			},
		],
	},
];

const STARTING_TOTEMS_TOOLTIP =
	'The totems we have standing when the fight starts, one per element. ' +
	'We take them as put down before the sim starts, so they cost no mana and no global cooldown. ' +
	'The sim puts down the highest rank our level knows.';

const SECONDS_BEFORE_PULL_TOOLTIP =
	'How many seconds before the pull we put the totem down. It has that much less time left at the pull ' +
	'(totems last 5 min, Searing Totem 30 to 55 sec by rank, Magma Totem 20 sec). 0 means at the pull.';

const startingTotems = (player: EnhPlayer): StartingTotems => player.getSpecOptions().startingTotems ?? StartingTotems.create();

function setStartingTotems(eventID: EventID, player: EnhPlayer, change: (totems: StartingTotems) => void) {
	const options = player.getSpecOptions();
	options.startingTotems = StartingTotems.clone(startingTotems(player));
	change(options.startingTotems);
	player.setSpecOptions(eventID, options);
}

const changedEvent = (player: EnhPlayer) => TypedEvent.onAny([player.specOptionsChangeEmitter, player.miscOptionsChangeEmitter]);

const learned = (totem: TotemOption, player: Player<any>) => totem.ranks[0].level <= player.getEffectiveLevel();

// The spell of the highest rank our level knows, for the totem's icon and tooltip.
function highestRankSpellId(totem: TotemOption, player: Player<any>): number {
	const level = player.getEffectiveLevel();
	return totem.ranks.filter(rank => rank.level <= level).pop()?.spellId ?? totem.ranks[0].spellId;
}

// Totems the rotation can put down that we can't start the fight with, by element.
const OTHER_TOTEMS: Record<TotemElement['field'], Array<{ name: string; spellIds: Array<number> }>> = {
	earth: [],
	air: [{ name: 'Windwall Totem', spellIds: [15107, 15111, 15112] }],
	fire: [],
	water: [{ name: 'Mana Tide Totem', spellIds: [16190, 17354, 17359] }],
};

// The element and name of the totem a spell puts down, or nothing when it isn't a totem.
function totemOfSpell(spellId: number): { element: TotemElement; name: string } | undefined {
	for (const element of ELEMENTS) {
		const totem =
			element.totems.find(totem => totem.ranks.some(rank => rank.spellId === spellId)) ??
			OTHER_TOTEMS[element.field].find(totem => totem.spellIds.includes(spellId));
		if (totem) {
			return { element, name: totem.name };
		}
	}
	return undefined;
}

// The totem a prepull action in the rotation puts down, when a starting totem replaces it.
//
// The sim puts the starting totems down at the pull, after the rotation's prepull actions. So
// with Grace of Air as the starting air totem, a prepull Windfury Totem at -3s is replaced at
// the pull and never stands in the fight. Our level has to know the starting totem, or the sim
// doesn't put it down.
function replacedPrepullTotem(player: EnhPlayer, action: APLPrepullAction) {
	if (player.getRotationType() !== APLRotationType.TypeAPL || action.hide) {
		return undefined;
	}
	const cast = action.action?.action;
	const id = cast?.oneofKind === 'castSpell' ? cast.castSpell.spellId?.rawId : undefined;
	const totem = id?.oneofKind === 'spellId' ? totemOfSpell(id.spellId) : undefined;
	if (!totem) {
		return undefined;
	}
	const value = startingTotems(player)[totem.element.field];
	const starting = totem.element.totems.find(option => option.value === value && learned(option, player));
	return starting ? { ...totem, starting } : undefined;
}

const doAt = (action: APLPrepullAction) => (action.doAtValue?.value.oneofKind === 'const' ? action.doAtValue.value.const.val : '');

// The note on a prepull totem in the Rotation tab that a starting totem replaces.
export function startingTotemPrepullNote(player: EnhPlayer, action: APLPrepullAction): string | undefined {
	const replaced = replacedPrepullTotem(player, action);
	if (!replaced) {
		return undefined;
	}
	return (
		`At the pull, the starting ${replaced.starting.name} replaces this ${replaced.name}, so this action does nothing. ` +
		`Remove it, or set the ${replaced.element.name} starting totem to None in the Settings tab, under Class Settings.`
	);
}

// The note on an element's seconds before the pull, when its starting totem replaces prepull
// totems in the rotation.
function replacedPrepullNote(player: EnhPlayer, totemElement: TotemElement): string | undefined {
	const replaced = player.aplRotation.prepullActions.flatMap(action => {
		const totem = replacedPrepullTotem(player, action);
		return totem?.element === totemElement ? [`${totem.name} at ${doAt(action)}`] : [];
	});
	if (!replaced.length) {
		return undefined;
	}
	const actions = replaced.length > 1 ? 'actions do' : 'action does';
	return `At the pull, this starting totem replaces the rotation's prepull ${replaced.join(', ')}, so that ${actions} nothing.`;
}

// The totems an Enhancement shaman starts the fight with, in the Class Settings section.
//
// It's a table with a row per element: the element's name, its totem icons (like the weapon
// imbue, clicking the picked one again leaves the element without a totem), and the seconds
// before the pull we put the totem down. The field is off while the element has no totem.
export function buildStartingTotemsSettings(parent: HTMLElement, simUI: IndividualSimUI<Spec.SpecEnhancementShaman>) {
	const player = simUI.player;
	const title = (<span className="starting-totems-title form-label">Starting totems</span>) as HTMLElement;
	const secondsTitle = (<span className="starting-totems-seconds-title form-label">Sec before pull</span>) as HTMLElement;
	const root = (
		<div className="starting-totems">
			{title}
			{secondsTitle}
		</div>
	) as HTMLElement;
	tippy(title, { content: STARTING_TOTEMS_TOOLTIP });
	tippy(secondsTitle, { content: SECONDS_BEFORE_PULL_TOOLTIP });
	parent.appendChild(root);

	ELEMENTS.forEach(totemElement => {
		const values: Array<EnumValueConfig> = [
			{ name: 'None', value: 0 },
			...totemElement.totems.map(totem => ({
				name: totem.name,
				value: totem.value,
				icon: totem.icon,
				showWhen: (player: EnhPlayer) => learned(totem, player),
			})),
		];
		const picker = new IconEnumRowPicker(root, player, {
			id: `starting-totem-${totemElement.field}`,
			label: totemElement.name,
			values,
			changedEvent,
			getValue: player => startingTotems(player)[totemElement.field],
			setValue: (eventID, player, newValue) =>
				setStartingTotems(eventID, player, totems => {
					if (totems[totemElement.field] === 0) totems[totemElement.secondsField] = totemElement.defaultSecondsBeforePull;
					(totems[totemElement.field] as number) = newValue;
				}),
		});

		// Under Forever, Grace of Air doesn't stack with Windfury Totem for now, like in the party
		// buffs.
		const graceOfAir = picker.rootElem.querySelector<HTMLElement>('[style*="spell_nature_invisibilitytotem"]');
		if (graceOfAir) {
			markUnstacked(
				graceOfAir,
				() =>
					startingTotems(player).air === AirTotem.GraceOfAirTotem && otherShamanWindfuryTotem(player)
						? "Gives no Agility while another shaman's Windfury Totem is up. Under Forever they don't stack for now."
						: undefined,
				simUI.changeEmitter,
			);
		}

		const secondsPicker = new NumberPicker(root, player, {
			id: `starting-totem-${totemElement.field}-seconds`,
			label: `${totemElement.name} totem sec before pull`,
			positive: true,
			changedEvent,
			getValue: player => startingTotems(player)[totemElement.secondsField],
			setValue: (eventID, player, newValue) => setStartingTotems(eventID, player, totems => (totems[totemElement.secondsField] = newValue)),
			enableWhen: player => startingTotems(player)[totemElement.field] !== 0,
		});
		markUnstacked(secondsPicker.rootElem, () => replacedPrepullNote(player, totemElement), player.changeEmitter);
	});
}

// The starting totems at the top of the Rotation tab, since the rotation plays from them. They
// are set in the Settings tab, so a click there takes us to it.
export function buildStartingTotemsSummary(parent: HTMLElement, simUI: IndividualSimUI<Spec.SpecEnhancementShaman>) {
	const player = simUI.player;
	const icons = (<div className="starting-totems-icons"></div>) as HTMLElement;
	const root = (
		<button type="button" className="starting-totems-summary">
			<span className="starting-totems-summary-label">Starting totems</span>
			{icons}
		</button>
	) as HTMLElement;
	parent.prepend(root);

	tippy(root, { content: `${STARTING_TOTEMS_TOOLTIP} Edit them in the Settings tab, under Class Settings.` });
	root.addEventListener('click', () => simUI.simHeader.activateTab('settings-tab'));

	const update = () => {
		const totems = startingTotems(player);
		const picked = ELEMENTS.map(totemElement => ({
			totem: totemElement.totems.find(totem => totem.value === totems[totemElement.field] && learned(totem, player)),
			secondsBeforePull: totems[totemElement.secondsField],
		})).filter(picked => picked.totem);

		icons.replaceChildren();
		if (!picked.length) {
			icons.appendChild(<span className="starting-totems-none">None</span>);
			return;
		}
		picked.forEach(({ totem, secondsBeforePull }) => {
			const icon = (<span className="starting-totem-icon" aria-label={totem!.name}></span>) as HTMLElement;
			ActionId.fromSpellId(highestRankSpellId(totem!, player))
				.fill()
				.then(actionId => actionId.setBackground(icon));
			icons.appendChild(
				<span className="starting-totem">
					{icon}
					<span className="starting-totem-time">{secondsBeforePull > 0 ? `-${secondsBeforePull}s` : 'at pull'}</span>
				</span>,
			);
		});
	};
	changedEvent(player).on(update);
	update();
}

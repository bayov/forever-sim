import tippy from 'tippy.js';

import { EnumValueConfig } from '../core/components/enum_picker';
import { IconEnumRowPicker } from '../core/components/individual_sim_ui/player_pickers';
import { NumberPicker } from '../core/components/number_picker';
import { IndividualSimUI } from '../core/individual_sim_ui';
import { Player } from '../core/player';
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
	secondsField: 'earthSecondsLeft' | 'airSecondsLeft' | 'fireSecondsLeft' | 'waterSecondsLeft';
	totems: Array<TotemOption>;
}

const ranks = (levels: Array<number>, spellIds: Array<number>) => levels.map((level, i) => ({ level, spellId: spellIds[i] }));

const ELEMENTS: Array<TotemElement> = [
	{
		name: 'Earth',
		field: 'earth',
		secondsField: 'earthSecondsLeft',
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
		secondsField: 'airSecondsLeft',
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
		secondsField: 'fireSecondsLeft',
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
		secondsField: 'waterSecondsLeft',
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
	'We take them as put down before the pull, so they cost no mana and no global cooldown. ' +
	'The sim puts down the highest rank our level knows.';

const SECONDS_LEFT_TOOLTIP =
	'Seconds the totem has left at the pull. 0 means we just put it down, so it has its full duration ' +
	'(5 min, Searing Totem 30 to 55 sec by rank, Magma Totem 20 sec).';

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

// The totems an Enhancement shaman starts the fight with, in the Class Settings section. Each
// element has a row of totem icons, like the weapon imbue, and the seconds its totem has left.
// Clicking the picked totem again leaves the element without one.
export function buildStartingTotemsSettings(parent: HTMLElement, simUI: IndividualSimUI<Spec.SpecEnhancementShaman>) {
	const player = simUI.player;
	const root = (<div className="starting-totems"></div>) as HTMLElement;
	const header = (<div className="starting-totems-header form-label">Starting totems</div>) as HTMLElement;
	tippy(header, { content: STARTING_TOTEMS_TOOLTIP });
	root.appendChild(header);
	parent.appendChild(root);

	ELEMENTS.forEach(totemElement => {
		const row = (<div className="starting-totem-row"></div>) as HTMLElement;
		root.appendChild(row);

		const values: Array<EnumValueConfig> = [
			{ name: 'None', value: 0 },
			...totemElement.totems.map(totem => ({
				name: totem.name,
				value: totem.value,
				icon: totem.icon,
				showWhen: (player: EnhPlayer) => learned(totem, player),
			})),
		];
		new IconEnumRowPicker(row, player, {
			id: `starting-totem-${totemElement.field}`,
			label: totemElement.name,
			values,
			changedEvent,
			getValue: player => startingTotems(player)[totemElement.field],
			setValue: (eventID, player, newValue) => setStartingTotems(eventID, player, totems => ((totems[totemElement.field] as number) = newValue)),
		});

		new NumberPicker(row, player, {
			id: `starting-totem-${totemElement.field}-seconds`,
			label: 'sec left',
			labelTooltip: SECONDS_LEFT_TOOLTIP,
			positive: true,
			changedEvent,
			getValue: player => startingTotems(player)[totemElement.secondsField],
			setValue: (eventID, player, newValue) => setStartingTotems(eventID, player, totems => (totems[totemElement.secondsField] = newValue)),
			showWhen: player => startingTotems(player)[totemElement.field] !== 0,
		});
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
			secondsLeft: totems[totemElement.secondsField],
		})).filter(picked => picked.totem);

		icons.replaceChildren();
		if (!picked.length) {
			icons.appendChild(<span className="starting-totems-none">None</span>);
			return;
		}
		picked.forEach(({ totem, secondsLeft }) => {
			const icon = (<span className="starting-totem-icon" aria-label={totem!.name}></span>) as HTMLElement;
			ActionId.fromSpellId(highestRankSpellId(totem!, player))
				.fill()
				.then(actionId => actionId.setBackground(icon));
			icons.appendChild(
				<span className="starting-totem">
					{icon}
					<span className="starting-totem-time">{secondsLeft > 0 ? `${secondsLeft}s` : 'full'}</span>
				</span>,
			);
		});
	};
	changedEvent(player).on(update);
	update();
}

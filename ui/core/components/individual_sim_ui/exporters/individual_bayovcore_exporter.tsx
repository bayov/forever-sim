import { IndividualSimUI } from '../../../individual_sim_ui';
import { PseudoStat, Spec, Stat } from '../../../proto/common';
import { classNames } from '../../../proto_utils/names';
import { UnitStat } from '../../../proto_utils/stats';
import { specNames } from '../../../proto_utils/utils';
import { IndividualExporter } from './individual_exporter';

// The Current EP as stat weights for BayovCore's Upgrades, which we paste into its Stat
// Weights page with Import.
//
// It's a Pawn string, which Upgrades reads, but with Upgrades' own names for the stats. Pawn's
// names would lose some of ours: it has one name for spell power and spell damage, and
// one for attack power and ranged attack power. Upgrades has hit and crit in percent, one
// each for melee and spells, because under Forever an item's hit and crit count for both. So
// we give it the larger of the two.
//
// Three more fields tell Upgrades what the weights are for. Level is our level and Tree our
// leading talent tree (1 to 3), which Upgrades' Auto picks stat weights by. Survival=own says
// the weights price Stamina or Armor themselves, like the leveling and PvP ones. Without it
// Upgrades drops those and works them out from its survival slider.
export class IndividualBayovCoreExporter<SpecType extends Spec> extends IndividualExporter<SpecType> {
	constructor(parent: HTMLElement, simUI: IndividualSimUI<SpecType>) {
		super(parent, simUI, { title: 'BayovCore Export', allowDownload: true });
	}

	getData(): string {
		const player = this.simUI.player;
		const epValues = player.getEpWeights();

		const weights: Record<string, number> = {};
		UnitStat.getAll().forEach(stat => {
			const name = IndividualBayovCoreExporter.getName(stat);
			const weight = epValues.getUnitStat(stat);
			if (weight == 0 || !name) return;
			weights[name] = Math.max(weights[name] ?? weight, weight);
		});

		const level = player.getEffectiveLevel();
		const leveling = this.simUI.sim.encounter.isLeveling();
		const name = `${specNames[player.spec]}, Level ${level}${leveling ? ' Leveling' : ''}`;
		const fields = [`Class=${classNames.get(player.getClass())}`, `Level=${level}`];
		const treePoints = player.getTalentTreePoints();
		if (treePoints.some(points => points > 0)) fields.push(`Tree=${player.getTalentTree() + 1}`);
		if (IndividualBayovCoreExporter.survivalStats.some(stat => weights[stat])) fields.push('Survival=own');
		Object.entries(weights).forEach(([stat, weight]) => fields.push(`${stat}=${Number(weight.toFixed(3))}`));
		return `( Pawn: v1: "${name}": ${fields.join(', ')} )`;
	}

	static getName(stat: UnitStat): string | undefined {
		return stat.isStat() ? IndividualBayovCoreExporter.statNames[stat.getStat()] : IndividualBayovCoreExporter.pseudoStatNames[stat.getPseudoStat()];
	}

	// The stats Upgrades counts as survival. Its STATS table in Modules/Upgrades.lua marks them.
	static survivalStats = ['Stamina', 'Armor', 'BonusArmor', 'Defense', 'Dodge', 'Parry', 'Block', 'BlockValue', 'Health'];

	// Upgrades' names for the stats (STATS in BayovCore's Modules/Upgrades.lua). The ones it
	// doesn't score, like armor penetration and resilience, are left out.
	static statNames: Partial<Record<Stat, string>> = {
		[Stat.StatStrength]: 'Strength',
		[Stat.StatAgility]: 'Agility',
		[Stat.StatStamina]: 'Stamina',
		[Stat.StatIntellect]: 'Intellect',
		[Stat.StatSpirit]: 'Spirit',
		[Stat.StatSpellPower]: 'SpellPower',
		[Stat.StatSpellDamage]: 'SpellDamage',
		[Stat.StatArcanePower]: 'ArcanePower',
		[Stat.StatFirePower]: 'FirePower',
		[Stat.StatFrostPower]: 'FrostPower',
		[Stat.StatHolyPower]: 'HolyPower',
		[Stat.StatNaturePower]: 'NaturePower',
		[Stat.StatShadowPower]: 'ShadowPower',
		[Stat.StatHealingPower]: 'Healing',
		[Stat.StatMP5]: 'MP5',
		[Stat.StatSpellHit]: 'Hit',
		[Stat.StatMeleeHit]: 'Hit',
		[Stat.StatSpellCrit]: 'Crit',
		[Stat.StatMeleeCrit]: 'Crit',
		[Stat.StatSpellHaste]: 'CastSpeed',
		[Stat.StatMeleeHaste]: 'AttackSpeed',
		[Stat.StatSpellPenetration]: 'SpellPen',
		[Stat.StatAttackPower]: 'AttackPower',
		[Stat.StatRangedAttackPower]: 'RangedAttackPower',
		[Stat.StatFeralAttackPower]: 'FeralAttackPower',
		[Stat.StatExpertise]: 'Expertise',
		[Stat.StatMana]: 'Mana',
		[Stat.StatHealth]: 'Health',
		[Stat.StatArmor]: 'Armor',
		[Stat.StatBonusArmor]: 'BonusArmor',
		[Stat.StatDefense]: 'Defense',
		[Stat.StatDodge]: 'Dodge',
		[Stat.StatParry]: 'Parry',
		[Stat.StatBlock]: 'Block',
		[Stat.StatBlockValue]: 'BlockValue',
		[Stat.StatArcaneResistance]: 'ArcaneResistance',
		[Stat.StatFireResistance]: 'FireResistance',
		[Stat.StatFrostResistance]: 'FrostResistance',
		[Stat.StatNatureResistance]: 'NatureResistance',
		[Stat.StatShadowResistance]: 'ShadowResistance',
	};
	static pseudoStatNames: Partial<Record<PseudoStat, string>> = {
		[PseudoStat.PseudoStatMainHandDps]: 'MainHandDps',
		[PseudoStat.PseudoStatOffHandDps]: 'OffHandDps',
		[PseudoStat.PseudoStatRangedDps]: 'RangedDps',
	};
}

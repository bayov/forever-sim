# Usage: python3 hacks_optional.py <scratch checkout>  (patches 3 env-gated switches into a LOCAL checkout for one-off measurements:
#   FAIR_NO_WEAPON_RACIAL, FAIR_SKIP_SET_PIECES="4,5", FAIR_RAID_HITS="perMin:damage". Never commit the patched tree.)
import sys
for root in sys.argv[1:]:
    p=root+"/sim/core/specializations.go"; s=open(p).read()
    if "FAIR_NO_WEAPON_RACIAL" not in s:
        s=s.replace("func (character *Character) AddWeaponSpecializationCrit(critPercent float64, types ...proto.WeaponType) {\n","func (character *Character) AddWeaponSpecializationCrit(critPercent float64, types ...proto.WeaponType) {\n\tif os.Getenv(\"FAIR_NO_WEAPON_RACIAL\") != \"\" {\n\t\treturn\n\t}\n",1)
        s=s.replace("import (","import (\n\t\"os\"",1); open(p,"w").write(s)
    p=root+"/sim/core/item_sets.go"; s=open(p).read()
    if "FAIR_SKIP" not in s:
        old="""	for _, activeSetBonus := range activeSetBonuses {
		activeSetBonus.BonusEffect(agent)
	}
}"""
        new="""	skip := os.Getenv("FAIR_SKIP_SET_PIECES")
	for _, activeSetBonus := range activeSetBonuses {
		if skip != "" && strings.Contains(activeSetBonus.Name, "Synthetic") && strings.Contains(","+skip+",", fmt.Sprintf(",%d,", activeSetBonus.NumPieces)) {
			continue
		}
		activeSetBonus.BonusEffect(agent)
	}
}"""
        assert old in s; s=s.replace(old,new,1)
        imp=s[s.index("import ("):s.index(")",s.index("import ("))]
        add=""
        if '"os"' not in imp: add+='\t"os"\n'
        if '"strings"' not in imp: add+='\t"strings"\n'
        s=s.replace("import (\n","import (\n"+add,1); open(p,"w").write(s)
    print("hacked", root)
for root in sys.argv[1:]:
    p=root+"/sim/core/rage.go"; s=open(p).read()
    if "FAIR_RAID_HITS" not in s:
        old="""		OnReset: func(aura *Aura, sim *Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt"""
        new="""		OnReset: func(aura *Aura, sim *Simulation) {
			aura.Activate(sim)
			var perMin, dmg float64
			if v := os.Getenv("FAIR_RAID_HITS"); v != "" {
				fmt.Sscanf(v, "%f:%f", &perMin, &dmg)
			}
			if perMin > 0 {
				wait := float64(time.Minute) / perMin
				hit := &PendingAction{}
				hit.OnAction = func(sim *Simulation) {
					if unit.GetCurrentPowerBar() == RageBar {
						r := dmg*2.5/GetRageConversion(63)*unit.rageBar.damageTakenMultiplier + unit.rageBar.flatDamageTakenBonusRage
						unit.AddRage(sim, r, rageFromDamageTakenMetrics)
					}
					hit.NextActionAt = sim.CurrentTime + time.Duration(sim.RandomExpFloat("Fair Raid Hit")*wait)
					sim.AddPendingAction(hit)
				}
				hit.NextActionAt = sim.CurrentTime + time.Duration(sim.RandomExpFloat("Fair Raid Hit")*wait)
				sim.AddPendingAction(hit)
			}
		},
		OnSpellHitDealt"""
        assert old in s; s=s.replace(old,new,1)
        imp=s[s.index("import ("):s.index(")",s.index("import ("))]
        add=""
        for m in ("os","fmt","time"):
            if '"%s"'%m not in imp: add+='\t"%s"\n'%m
        s=s.replace("import (\n","import (\n"+add,1); open(p,"w").write(s)

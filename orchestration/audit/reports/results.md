# Results pass 110 (master f9fa79917)

Range 2af970d74..f9fa79917 (233 commits listed, 30 watched files). Sim changes: Rend and Tear style bear hit scaling (sim/druid/bear_attacks.go, tank only),
Totem of Sustaining +1 Water Shield charge (sim/shaman/water_shield.go, items.go; healer/Warden only). Presets and gear: tank warrior, protection paladin, warden shaman, healers, bear.
No change to proto, mksettings, database, assets/database, DPS requests or builds.json.

- Shared run (68, 20k, seed 1) vs metas: 14 of 68 differ by more than 0.5, largest +1.4 (warrior fury_2h-duty-sunder), noise-level; none explained by a commit.
- builds.json unchanged against the previous pass; compare.py builds checked 68.
- Drift: mksettings and requests unchanged, so equal to the reference (regen.py not re-run).
- Open finding results-20261104-2 stays as is (metas within 1.4).

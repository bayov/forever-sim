# healer_gear pass 18 (master f9fa79917)

- Trigger: tools/rotopt/phase0_specs.json changed; the diff is the enh row's talents only, no heal row touched.
- phase0_check.py: healer rows priest_holy, priest_discipline, paladin_holy, shaman_restoration list, no flags; druid_restoration has no list yet (expected).
- Healer lists changed since 3e9748641 only by price notes (266e5357f), picks unchanged.
- Requests (tank and raid) of priest_holy, paladin_holy, priest_discipline, shaman_restoration wear their gear sets (ids and enchants match).
- sim/shaman/items.go gained Totem of Sustaining (23200, Water Shield 4th globe); not in the shaman_restoration list.
- Open finding healer_gear-20261126-1 not re-measured (no sims this pass).

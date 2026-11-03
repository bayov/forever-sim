# Priest, round 70753

Worker ledger for the Priest DPS specs (Shadow, Shadow + Holy, Smite) against the 2026-11-02 development notes
(post 9) and client 1.60.1.70753. Final summary goes at the top once Phase B lands.

## Phase A: sources

- Client: `python3 tools/sources/wago.py fetch 1.60.1.70753 --set client_update` and the same for 70707, then a diff of
  every row whose spell has SpellClassSet 6 (Priest), across SpellEffect, SpellMisc, SpellCooldowns, SpellPower,
  SpellCategories, SpellLevels, SpellAuraOptions, SpellClassOptions, SpellShapeshift, SpellInterrupts,
  SpellAuraRestrictions, SpellTargetRestrictions, SpellLabel, SkillLineAbility. Every Priest change in 70753:
  - Shadow Word: Death, all four ranks (1309595, 1309633, 1309635, 1309636): SpellEffect EffectBasePointsF 10 -> 30
    (the self-damage effect).
  - Holy Binding added (1328569 talent aura, proc type 16384, 50%; 1328570 the self heal). Healing.
  - Cure Disease 528, Abolish Disease 552: SpellShapeshift ShapeshiftExclude 0x8000000 (Shadowform) -> ShapeshiftMask
    0x8000000, Attributes_2 +0x80000. The Shadowform fix.
  - Mind Control: Attributes_4 +0x800. Not DPS.
  - Attributes_14 +0x40 on 230 Priest spell rows. Not a Priest change: 1442 rows across every class get it,
    Blizzard included (10, 6141, 8427, 10187), so it is not the low-rank proc penalty exemption (Blizzard keeps the
    penalty). Most likely the notes' "learning new spells no longer adds them to your action bars". No sim effect.
  - No damage, coefficient, cost, cooldown, duration or proc change on any Priest DPS spell or Priest racial spell
    (Devouring Plague, Shadowguard, Starshards, Chastise, Dark Sacrifice, Penance, Smite, Holy Fire, Mind Blast, Mind
    Flay, Shadow Word: Pain).
- Server hotfixes: `python3 tools/sources/hotfixes.py fetch 1.60.1.70753 --set client_update` found 0 rows for 70753
  (2026-11-03).
- Talent trees (Trait tables) are A1's port; Phase B reads `orchestration/talent_remap_70753.md`.

## Ledger

| note line | tier and evidence | sim status | fix |
|---|---|---|---|
| Power Word: Shield can be replaced by an equal or higher rank | notes | n/a | healer session (no DPS preset casts it) |
| Shadow Word: Death: 30% of max health to the priest if the target lives (was 10%), none on a miss | client 70753 SpellEffect bp 10 -> 30, all ranks; notes | missing (the sim left the self-damage out) | shadow_word_death.go: on a landed hit that doesn't kill, `RemoveHealth(0.30 x max health)`; nothing on a miss. Test `TestShadowWordDeathSelfDamage`. No DPS effect: nothing in the DPS presets reads the priest's health (goldens unchanged). The Shadow rotations cast it on cooldown; Vampiric Embrace's 20% of Shadow damage back covers it in a raid |
| Undead Dark Sacrifice no longer breaks crowd control | notes | n/a | the sim has no crowd control |
| Holy: Binding Heal removed, Holy Binding in its slot | client 70753 (1328569, 1328570); notes | n/a for DPS | healer session. No DPS preset takes the slot (Smite's Holy string `30505110002` stops before it, the Shadow + Holy build has 3 in Twilight Focus only). The proto field is already `holy_binding`; A1 remaps the tree |
| Smite builds: every Holy talent they take | Smite takes Twilight Focus 3, Holy Specialization 5, Divine Fury 5, Holy Nova 1, Blessed Recovery 1, Searing Light 2 | no note line touches them | re-check after A1's tree port (Phase B) |
| Devouring Contagion no longer spreads Devouring Plague from a critter | notes | n/a | the sim models only its cost cut, no spread |
| Vampiric Embrace healing can't crit | notes | landed | vampiric_embrace.go heals a flat 20% of the damage with `GainHealth`, no crit roll |
| Cure Disease and Abolish Disease no longer break Shadowform | client 70753 SpellShapeshift; notes | n/a | the sim doesn't cast either |
| Talents no longer have a low-rank proc penalty (except Blizzard) | notes | landed (never modelled) | no Priest talent proc in the sim has a rank penalty (no roll under 100% outside the 5% Vestments set proc); `downrankCoefficient` is the separate spell power penalty and stays. A2 owns any shared part |
| Races: Undead Touch of the Grave 15 s ICD, Troll Berserking 15 s, Human Spirit 10% and Ruthless 1%, Dwarf Earth Fury out and Artisan's Precision 1% hit in, Orc Savage out and Blood Fury 2 min, Tauren Endurance health only, Skyborne Wind-Blessed 2%, all races' base stats | notes | A2 | shared racials (A2). Phase B's race sweep picks them up |
| Throw windup, proc roots DR | notes | n/a | not a Priest mechanic |
| Threat and healing rows | | n/a | tank/healer sessions |

## Phase A commits

- Shadow Word: Death self-damage (this branch, `round-70753-priest`).

## Notes-only values

None. The Priest section of the notes has no damage line, and the one value it gives (Shadow Word: Death's 30%
self-damage) is in client 70753. The 2026-11-03 hotfix
(`orchestration/audit/ref/client_update/sources/foreverchanges_2026-11-03_hotfix_and_70753.txt`) has no Priest line;
its note that 70753 "writes the October 30 hotfix into the client files" for priests matches the 70707 -> 70753 diff
above (Shadow Word: Death, Holy Binding, Cure / Abolish Disease).

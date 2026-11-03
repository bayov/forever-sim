# Priest, round 70753

Worker ledger for the Priest DPS specs (Shadow, Shadow + Holy, Smite) against the 2026-11-02 development notes
(post 9) and client 1.60.1.70753.

## Summary

Landed on master at 69b9c38 (2026-11-03).

**What changed.** One Priest mechanic: Shadow Word: Death now costs the priest 30% of max health when the target lives
(10% before) and nothing on a miss (client 70753 SpellEffect bp 10 -> 30 on all ranks; notes). The sim used to leave the
self-damage out; it now removes the health. DPS effect 0 (nothing in the DPS presets reads the priest's health); the
three report hashes move because the damage-taken metrics changed. Every other Priest note line is healing, crowd
control, or already true in the sim (ledger below). No notes-only values. The client talent tree is unchanged
(`orchestration/talent_remap_70753.md`), so no talent string moved. A2's Classic race base stats move Smite +0.6.

**Builds** (20,000 iterations, seed 1; seed 1000003 within 0.4; Undead; talents unchanged):

| build | talents | 120 s old / new | 180 s old / new | 300 s new |
|---|---|---|---|---|
| priest__shadow | 00530400102-3-50052101231302251 | 944.7 / 944.7 | 948.7 / 948.7 | 946.0 |
| priest__shadow-holy | -3025011303-50052101231302251 | 902.3 / 902.3 | 906.6 / 906.6 | 873.9 |
| priest__smite | 525035231030001001-30505110002-3 | 777.8 / 777.8 | 761.4 / 762.0 | 687.9 |

(The 120 s "old" numbers are the pre-A2 rotopt runs of this round; the 180 s old ones are the published values.)

**Re-optimization.** Talent search at 120/180/300 s (3000 iterations, 20,000 to confirm) from each preset: Shadow and
Smite keep their talents. Shadow + Holy's unconstrained search leaves Holy for Discipline (947.3 combined), which is the
Level 60 Shadow build's idea, not this variant's, so its Holy talents stay; its Shadow tree (53042101211312251) on the
Level 60 build loses 0.1 to 0.2 at every length on both seeds, so Level 60 keeps its tree. Knob search: Shadow keeps
every knob at all three lengths. Smite dpMana 60 (+4.4 at 120, -3.0 at 180), Holy Fire on (+9.4 at 300, -11 to -22
elsewhere), Dark Sacrifice 20 and Smite mana rate 45, and Shadow + Holy Dark Sacrifice 20 (+15 at 300, -1.5 at 180), are
fight-length trade-offs, so none ships (handoff rule: positive at 120, 180 and 300 on both seed sets). No rotation file
changed.

**Races** (180 s, seed 1, each race on its own rotation and gear from req/races; both seed sets agree within 0.4):

| build | Undead | Troll | Human | Night Elf | Dwarf | Gnome |
|---|---|---|---|---|---|---|
| Shadow | **948.7** | 945.5 | 944.8 | 944.5 | 941.9 | 941.5 |
| Shadow + Holy | **906.6** | 903.2 | 902.4 | 902.1 | 899.7 | 899.4 |
| Smite | **762.0** | 732.9 | 737.7 | 734.7 | 730.1 | 741.0 |

Best overall and Horde: Undead for all three. Best Alliance: Human for Shadow and Shadow + Holy (0.3 over Night Elf on
both seeds), Gnome for Smite. No race change.

**Audit** (Phase B step 4):
- results: all three requests re-simmed on master at 20k seed 1; Shadow and Shadow + Holy equal published, Smite +0.6
  (A2's base stats), meta updated. `unread_fields.py`: clean.
- rotations: breakdown and one timeline of Level 60 Shadow: Shadow Word: Pain kept up (recast every 24.25 s), Mind Blast
  21.8 and Shadow Word: Death 9.7 casts in 180 s (on cooldown), Mind Flay fills, no idle GCDs; the Troll Shadowguard lines
  are inert for the Undead (settled). Knob candidates as above. Standards 31: the APL notes state each line's intent and
  are unchanged.
- gear: no Priest gear change since e8313ea (2026-11-02); gear pass 52 (2026-11-03) found nothing for Priest.
  `tools/audit/gear/lin2.py` crashes parsing the current `orchestration/gear_targets.md` table (ValueError on 'Sta');
  shared audit tool, not fixed here.
- consumables: `scan.sh` clean for shadow_priest.
- mechanics, client_update: Phase A below (client diff 70707 -> 70753 for every Priest spell, hotfixes 0 rows).

**Open questions and shared changes.** No shared code changed. Question: the brief's Claude-Session trailer names
another session id; this worker's commits carry its own session's id.

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

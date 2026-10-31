# Consumables pass 58 (master e05fd732e)

No findings. Last audited master a2a3e66e1; client 1.60.1.70707 is still the newest build. Class branches are not ahead in any audited way (stale copies only).

Moved watched files, all checked:
- ui/*/presets.ts (all eleven specs): mostly renames (build, encounter and APL labels such as "Generic boss", "Humanoid boss", "Mutilate", "Faerie Fire duty"), removal of the Level 30 and low-level encounters and gear, comment trims, and new hunter Wolf and Tallstrider pet builds. No Consumes, buff, debuff or enchant value changed. scan.sh is clean in every ban section.
- Default Level 60 encounters are still MobTypeUnknown; Humanoid appears only in labelled "vs Humanoid" encounters.
- tools/rotopt/mksettings (withConsumes helper, ele_mana_tide, shadow_holy, smite, rogue_combat_swords/daggers, rogue_assassination): each matches its presets.ts Consumes (ManaTideConsumes Major Mana Potion; ConsumesHoly and ConsumesSmite; P1SapperOnly/CombatSwords/CombatDaggers with no Land Mine).
- ui/warlock gear_sets: the synthetic demonology set was deleted (no live use left); no enchants touched.
- tools/dps_report/assumptions_def.py: Windfury Totem ICD 0.1 s, extra attacks keep the swing, new rogue evisc_ap. No buff values.

scan.sh on master: minor-armor exclusive-effect count 3 (Faerie Fire, Curse of Recklessness, Tallstrider Dust Cloud; intended shared slot). Totem counts as the ruling.

No open items.

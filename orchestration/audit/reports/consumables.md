# Consumables pass 44

Master f62cb7971 (last audited f796d9953). Result: quiet, no findings.

- Moved: tools/dps_report/assumptions_def.py (+7 single-value assumptions: hunter_multishot_cast, dot_live_ticks, hunter_base_rap, druid_bite_miss, wf_totem_icd, extra_attack_swing, channel_fsr) and orchestration/standards.md (standard 17 reworded, new 27 and 28).
- No buff, debuff, consumable, enchant or oil_proc_damage change. wf_totem_icd keeps the 1.5s ruling (standard 17); client has 100 ms, recorded as a ruled assumption.
- scan.sh on master: all banned/Revelation/trinket/crit-elixir sections empty. Air/weapon totem counts unchanged.
- Note: scan's minor-armor count reads 3 (expect 2). The third is ForeverDustCloudAura (tallstrider pet ability, client 70532), not used by any preset. Standard 20 holds (FF and CoR share the slot). The scan's expected count is stale.
- wago latest: 1.60.1.70532, already audited. Class branches carry no new commits in audited files beyond what master has.
- Open items: none.

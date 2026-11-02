# Consumables pass 64 (master 67084d0ec)

Trigger: tools/dps_report/assumptions_def.py changed since fa865a3b8b. No other watched file moved.

## Verified
- Parsed the old and new assumptions_def.py and compared every entry ignoring the new `group` field and ordering: no buff, oil, flask, imbue or totem entry changed (oil_proc_damage, hunter_tsa_hm_rank, natural_flask_zone, lock_stone_with_oil, imbue scrolls all identical). Added keys: wf_totem_nodamage, flurry_batch, warr_uw_queued (swing and rage rules, not this audit's).
- scan.sh on master: banned, Revelation, trinket and crit-elixir sections empty. Counts as before (minor-armor count 3, known: Dust Cloud; ref expects 2).
- wago latest 1.60 build: 1.60.1.70707, unchanged since last pass.

## Open
- consumables-20261101-1 (Sapper on ranged hunters, low) remains open.

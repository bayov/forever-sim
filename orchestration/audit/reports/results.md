# Results audit, pass 61 (master d50909a37)

Checked with the loop's shared run (20k, seed 1, 180 s, all 77 requests) and a fresh mksettings/wowsimcli build at d50909a37.

- Why due: 118 watched files changed since 334c7decd: 19 new duty and pet-variant requests (tools/dps_report/req), the raid sim's compsearch (new, not read by the report's sims), sim/core debuffs and buffs, Seal of Wisdom (sim/paladin/sow.go), Faerie Fire in Cat/Bear as the caster spell (sim/druid/faerie_fire.go), the Feral preset.
- Published vs mine: 75 of 77 within 0.15, including all 19 new requests. The two others are the known stale metas: mage__frostfire meta 1087.5 vs run 1088.4 (+0.9), shaman__elemental-level60 772.2 vs 772.7 (+0.5). Same numbers on binaries at 8bf7ed6b1 and 334c7decd, builds.json already carries the run's values. Stale meta only, as in pass 60; not filed.
- builds.json vs the run: all 77 match.
- Moves against the reference: only druid__feral-cat-duty-ff, 862.2 -> 858.5. Explained: 0bb3ec9b6 re-simmed it after 0cf1a8c8b (Faerie Fire is the caster spell now). The 334c7decd binary gives 823.2 for that request, so the sim change is worth +35 and the published 858.5 matches master. ADDED: the 19 new requests.
- Drift: every previous line is unchanged. New lines are all standing kinds: hunter pet variants (petType Tallstrider or Wolf, mksettings writes Cat; 6 requests), Ret and Ret Shock duty presets (Crusader and Wisdom debuffs absent in the request; primarySeal Command; Ret Shock potion = open finding results-20261025-1), warlock Recklessness duty (curseOfRecklessness and faerieFire off), warrior Sunder duty (sunderArmor off).
- regen.py: mapping extended for the -tallstrider and -wolf hunter ids (tools/audit/results/regen.py).
- Unread fields: none printed.
- New duty rows cost what a duty should: Fury Sunder -34.1, Arms Sunder -42.9, Fury 2H Sunder -35.8, Ret Wisdom -13.3, Warlock Recklessness -14.7 (Aff), Hunter Tallstrider -71 (BM), Wolf -57 (BM). Not modelling errors by arithmetic, no finding.
- Baseline: ref/results req_summary.txt and drift_report.txt replaced with this run.

# Consumables pass 108, master 5536db4bc

Changed watched file: tools/dps_report/assumptions_def.py (diff from 226ada86b).
- Removed the low settings of natural_flask_zone, orc_bf_weapon, lock_stone_with_oil, warr_exec_wf2 (settled assumptions); descs gained cost figures. No sim, preset or mksettings values changed.
- lock_stone_with_oil now cites Blizzard's warlock preview (stones stack with Wizard Oil): consistent with the pass doc.
- Finding consumables-20261125-1 (low): natural_flask_zone desc still says 60 Stamina and 4-5%; sim has 30 and 2% since client 70862.
- Wago latest build 1.60.1.71091 = last audited; no new client build.
Open items: the three low findings in findings/consumables.md.

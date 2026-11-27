# healer_gear pass 20, master 5ab589d85
Checked (no sims): phase0_check.py healer rows; per-slot id/enchant/gems compare of each gear set with its requests; other-faction sets.

| spec | list score | faction | requests wear list | notes |
|---|---|---|---|---|
| druid_restoration | 126.55 | Horde | yes (tank, raid) | new list; neck 22659 out (-20271127-2); no _alliance set |
| paladin_holy | 122.89 | Alliance | yes | Armaments rewards removed, requests fixed in d74a53261 |
| shaman_restoration | 127.82 | Horde | yes | _alliance set differs in slots 3, 6 only (not examined further) |
| priest_holy | 132.29 | Alliance | yes | re-scored 11-27 |
| priest_discipline | 123.67 | Alliance | yes | re-scored 11-27 |

No request holds synthetic ids (99xxxx). -20271127-1 resolved. -20261126-1 (Holy lists score ~1.4% above bench) not re-measured (no sims); priest lists were re-scored since, so recheck next sim pass.

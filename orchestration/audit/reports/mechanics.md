# Mechanics pass 99 (2026-11-23, master 5518f4873)

Range 7a6d1beb9..5518f4873: 5 sim commits, 4 non-test sim files. No new client build (1.60.1.71037 still newest).
No new Blizzard post. No daily-update commits to TASKS.md or the ledger. One new finding (low).

## Get current

| Source | Seen | Newer? |
|---|---|---|
| Client | 1.60.1.71037 (`wago.py latest --probe 300`) | no |
| Hotfix thread t/2403040 | last post 2026-11-21 00:18:29Z, post 1 edited 00:19:18Z | no |
| Dev notes t/2360696 | last 2026-11-04 19:35Z, post 1 edited 2026-11-03 03:20Z | no |
| Daily update | no TASKS.md / update_ledger.md commits, no forever-update merges | no |

## Commits checked (client 1.60.1.71037, `sp.py`, SpellCategories, SpellCooldowns, ItemEffect)

| Commit | Change | Verdict |
|---|---|---|
| 79c688d0c | Empty weapon slot never swings (swingAt = NeverExpires); APL wait capped at NeverExpires | engine safety fix; no effect on any equipped weapon |
| 078699747 | Axe of the Deep Woods: 18104 Wrath, bp 126 var 0.333 = 105-147 Nature, 1.7 PPM | damage matches tier 1; PPM from wowsims/forever's client store (client has no PPM row) |
| 65f071726 | Headmaster's Charge: 18264 +20 Int, 15 min, 10 min CD | matches (ItemEffect 99609, eff 35 party aura 29 bp 20); wearer-only is fine for a solo sim |
| 592444f24 | Flame Wrath shield ignores self damage | correct; no reflect of own damage |
| efe009fe1 | Helm of Fire 10578: 331 var 0.272, 33 per 2 s for 8 s, 5 min CD, 10 s cat 1141 | matches ItemEffect 98606 and SpellEffect |
| efe009fe1 | Circle of Flame 17447: 79 var 0.5 per 1 s, 10 s, 5 min CD | matches ItemEffect 99200 |
| efe009fe1 | Robe of the Archmage 18385: 500 var 0.5, 5 min CD | values match; but spell Category 30 with 60 s category recovery is missed: mechanics-20261123-1 |

## Conclusions

- mechanics-20261123-1 (low): the robe's Replenish Mana shares spell category 30 (60 s) with mana gems,
  healthstones and Demonic Rune; the sim gives it no shared cooldown. No preset wears it.
- Open findings unchanged: 20261120-2, 20261118-2, 20261118-1 (shaman half), 20261113-2, 20261104-1.

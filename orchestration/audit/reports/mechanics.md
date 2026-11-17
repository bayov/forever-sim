# Mechanics pass 86 (2026-11-17, master 5f20dfdf4)

Range 1a85df1d7..5f20dfdf4. Client 1.60.1.70938 -> 1.60.1.70959. No new findings; both open findings unchanged.

## Get current

| Source | Newest | New since pass 85 |
|---|---|---|
| Client | 1.60.1.70959 (gate) | diffed against 70938 (`wago.py diff --by-spell`, 161 spells) |
| Server hotfixes on 70959 | push 114157 (2026-11-17 18:03) | 114097 touches SpellAuraOptions 124699 / 126501 / 249144 (Unbridled Wrath, Flurry, Reckless Swings); row data is null in the feed, so the "ProcTypeMask 4 -> 20" value in ad1d7f469 is unverified here (class change) |
| Dev notes t/2360696 | post 10, 2026-11-04 | none |
| Hotfix thread t/2403040 | post 11, 2026-11-17T02:17Z (Nov 16 hotfixes) | Windfury Totem armor / ICD / Rage / roll fixes; Fury UW, Flurry, RS on HS/Cleave; Maelstrom double use; VE 10% |
| Blue posts | Aggrend 11-05, Linxy (retail) | none for Forever mechanics; Kaivax only via the hotfix thread |
| TASKS / daily update | 5f20dfdf4: build 70959, Windfury Totem reshape and Unsealed Wrath 100 ms ICD rulings, Nov 16 leftovers | waiting on the user, not findings |

## 70959 shared deltas

| Spells | Change | Verdict |
|---|---|---|
| ~90 consumables, flasks, jujus, every Well Fed | Attributes_2 +0x100; ProcChance 101 SpellAuraOptions rows dropped | logging flag / passive auras: no effect |
| Windfury Totem 8515/8516, 10608-10612 | buff spells lose AP aura + extra attack; AP on raid aura effect 1 | user ruling pending (TASKS, PR #133); Nov 16 hotfix points already landed (ledger SHR-HF-02..04) |
| Gift of Arthas / Echoes of Lordaeron | SpellAuraOptions rows removed, target change | not in Level 60 raid set |
| Class rows (Stormstrike 8 s / 100% already in sim, Maelstrom 408498, Judgements DefenseType 3, Unsealed Wrath, Berserker Rage, Wild Growth) | | client update auditor's ledger (Build 70959 section) |

## Commits triaged

| Commit | Change | Check | Verdict |
|---|---|---|---|
| 6ea6e47b4 | Manual Crowd Pummeler 3 min cooldown | 70959 ItemEffect 98990 Charges 3, CoolDownMSec 180000; 13494 +50% 30 s | ok (tier 1) |
| 6875e3b1e | Five Thunders 1679 set bonuses | 18679 / 18378 / 450626->27775 (65 SP 10 s) / 14047 (23 SP) match 70959; 5% proc rate is an assumption | ok |
| b075ae64d | core.Pet.InheritedStats accessor; hunter pet white swing without owner crit | accessor only in core | class change |
| cf95c3e32 | Death Mark always hits | 1316709 / 1325756-8 Attributes_3 0x40000 on 70959 | ok (tier 1) |
| ad1d7f469 | Warrior UW / Flurry / RS on HS and Cleave | push 114097 touches those three rows; Nov 16 text agrees | class change, value not readable |
| 2ebdc14d6, 29481dd0e, eada6763e, 4a8541e46 | Execute refund, trap hit, Raptor DW penalty, Predator's Edge | log/wowsims-based class rules | left to class audits |
| 39b94e364 | rotopt set search | tooling | n/a |

## Open

mechanics-20261113-2 (Weakening Disease: `item_effects.go:1347` still Strength -50), mechanics-20261104-1 (druid 70790 scales: `tuning_70753.go` unchanged): still open.

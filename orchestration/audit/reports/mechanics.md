# Mechanics pass 74 (2026-11-11, master 707ae92cc5)

No new finding. mechanics-20261107-1 (natural flasks) is resolved: 032692409, merged by ca07df649, gives 30 Stamina
and 2% zone bonuses, which match 70862 and 70888. mechanics-20261104-1 (druid scales) is still open, since no commit
touched `tuning_70753.go`.

## Range
27d0d59d67..707ae92cc5: 15 sim commits. Shared engine and raid commits:

| commit | what | verdict |
|---|---|---|
| 032692409 | Natural flasks 30 Stamina and a 2% zone bonus | Matches 70862/70888 SpellEffect (tier 1). Resolves mechanics-20261107-1. |
| cb651085c, eb6c042a4 | No imbue turns off a weapon totem buff (`totem_weapon_buffs.go`, `buffs.go`, `character.go`) | Matches the Nov 10 hotfixes (t/2403040 post 6, tier 1): "Weapon imbues and Weapon totems can now be used at the same time." Separate WF Weapon and WF Totem cooldowns are report assumption wf_weapon_totem_icd. On 70888 the procs have separate per-spell icds and no shared category (439431 20%/1500 ms; 8515/10609/10612 20%/100 ms), so the assumption fits the client. No finding. |
| 54a15ba24 | mksettings spec_enh comment only | No mechanic change. |
| f20d5938a | core dot.go/aura.go: no ghost tick after a lazily processed expiry (Aura.endedAt) | An engine timing fix, not a Forever value. The guard looks right: it ticks only when the aura is active or the tick was due by the time the aura ended. The report hashes moved on purpose (Affliction -6.3, rogues +-0.3). No finding. |

Class commits (Wrack/Soul Siphon 113766, Imp Firebolt 113766, Arcane Shot/Summon Hawk 11% base mana 113892, Hawk +80%,
NS on Wild Growth, Idol of Ferocity, Arcane Concentration over Missiles, Sweeping Strikes charges, Shaman Nov 10 tuning)
follow the Nov 10 hotfix post. They are worker territory, and no shared formula moved.

## Client
Still 1.60.1.70888 (the gate's probe found nothing newer). Hotfix rows for 70888 were not read. `hotfixes.py fetch 1.60.1.70888 --set mechanics` was still splitting the "70888 spell" search after 40 min (about 340 parts), so I stopped it and `pushes` printed nothing. Open item for the next pass: read the newest pushes on wago.tools/hotfixes?product=wowf (searches "70888 spell" and "113892", newest first), not the full fetch. The Nov 10 post's shared items are covered under Blizzard below.

## Blizzard
- Dev notes t/2360696: still 10 posts, newest post 10 (2026-11-04), and post 1 has not been edited since 2026-11-03.
- Hotfixes t/2403040 post 6 (2026-11-11 03:27, Nov 10 hotfixes, also edited into post 1). The shared items in it:
  weapon imbues with weapon totems (on master), the Revelation enchant no longer procs from Vampiric Embrace heals (settled
  on master per TASKS), and Lesser Spirit enchants (no DPS). Hunter pet health 150% -> 115% is a class item, tracked in TASKS.md.
- Blue search (Kaivax, Aggrend, Gnedi, Linxy, Brenoch, after Nov 9): apart from that, only a fix for missing shaman totem
  items (t/2409772), with no mechanic in it.
- Daily update: TASKS.md 707ae92cc adds the Nov 10 items above. Merge ca07df649 brought the flask fix.

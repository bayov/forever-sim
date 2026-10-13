# Pet buffs under Forever: hunter pets and warlock demons

Research note, 2026-10-12. It tests standards.md item 14 against the beta (client 1.60.1.70338 and
beta reports).

Source tiers, in order: **Forever beta**, then **SoD fallback** (Season of Discovery client,
hotfixes, blue posts and findings) where the beta is silent, then **Era fallback** only where SoD
is silent too. Every verdict below is Forever beta unless it carries another tag. Classic Era client
1.15.8.66129 is used once, as the baseline for the flag comparison, and that is a diff, not a
verdict.

## Verdict in short

- **Hunter pets: the ruling holds.** No player buffs and no pet consumables. The pet gets its
  share of the hunter's stats through the Hunter Pet Scaling aura instead.
- **Warlock demons: the ruling should change.** The beta treats demons the same as hunter pets:
  - Every group-cast buff, party aura and stat totem in the client now carries a new "players
    only" target flag. Classic Era does not have this flag. The flagged spells include Gift of the
    Wild, Prayer of Fortitude, Arcane Brilliance, Prayer of Spirit, every Greater Blessing, Battle
    Shout, Leader of the Pack, Moonkin Aura, Devotion Aura, Strength of Earth and Grace of Air.
  - A beta tester reported on day 2 that any class buff on a warlock or hunter pet returns
    "Invalid Target". The report was closed as "Not a Bug".
  - Blizzard's warlock deep dive says demons "now scale from their summoner's stats".

  A raid demon therefore gets none of the raid buffs it gets in the sim today. It gets the
  warlock's stats through Warlock Pet Scaling instead, and the sim has no model of that.
- Target debuffs are not touched. They are on the target, so pet damage gets them as before.

## Sources

| # | Source | Date | Type |
|---|---|---|---|
| S1 | [Blizzard, Forever class deep dives: Hunter and Druid](https://worldofwarcraft.blizzard.com/news/24301515), Pets section: "Pets now get increased stats based on their Hunter's gear." "Pets can no longer receive player buffs that increase their stats." | 2026-09-30 | official |
| S2 | [Blizzard, Forever class deep dives: Rogue and Warlock](https://worldofwarcraft.blizzard.com/en-us/news/24310968): "Pets now scale from their summoner's stats." It says nothing about buffs on demons. | 2026-10-07 | official |
| S3 | Beta client 1.60.1.70338, wago.tools `SpellMisc` compared with Era 1.15.8.66129: `Attributes_3` bit 0x100 (SPELL_ATTR3_ONLY_TARGET_PLAYERS) is set on 72 spells that lack it in Era. The full list is under "Client flag" below. No spell lost it except one boss spell. | build 70338 | client |
| S4 | Beta client 70338 `SpellEffect`: Hunter Pet Scaling 415429 and Warlock Pet Scaling 416189 (the Season of Discovery auras, carried into Forever). Both are server-scripted, so every amount is 0 and the client gives no coefficients. | build 70338 | client |
| S5 | [forever-bugs #27, "Player Pets (Warlock/Hunter) Cannot be buffed"](https://github.com/ClassicWoWCommunity/forever-bugs/issues/27): "When trying to buff a player's pet (warlock or Hunter), it responds with Invalid Target." Closed by a tracker maintainer as Not A Bug: "This was introduced to remove the need for pets to be buffed and having to keep Scrolls on you." This is a community tracker. No Blizzard post is shown. | 2026-09-18 | beta tester |
| S6 | [EU beta forum, "Pets can't be buffed"](https://eu.forums.blizzard.com/en/wow/t/pets-cant-be-buffed/629896): Blessings, Scrolls and Power Word: Fortitude fail on a pet; Blessing of Might, Strength of Earth, Kings, Mark and JuJu Power named. Not a blue post. | 2026-09-19/21 | beta tester |
| S7 | [ForeverChanges, hunter pets](https://foreverchanges.pro/hunter-pets), from the forever-hunter wiki. The pet gets 2 health per point of the hunter's Stamina, 30% of the hunter's armor, 10% of the hunter's highest attack power, and all of the hunter's crit chance, racials included. "Your pet no longer takes buffs from other players, even those it took in Classic." Beta testing, not client data. | page updated 2026-09-24 | beta tester |
| S8 | [Beta development notes](https://us.forums.blizzard.com/en/wow/t/2360696). 09-24: "Voidwalker Sacrifice now correctly scales with 10% of the Warlock's spell healing" (a demon ability that scales from the warlock). 10-08: "Unleashed Fury now correctly increases your pet's damage". Known issue on 10-08: the Ferocity talent does not raise the pet's crit. | 2026-09-24 to 10-08 | official |
| S9 | [Beta known issues, 09-17](https://us.forums.blizzard.com/en/wow/t/wow-forever-beta-known-issues-september-17/2352687), via Warcraft Tavern: "Warlock pets do not display Mana as a resource on the character sheet, and do not mention the scaling from their master." Also: "Pet Hit chance for both Hunters and Warlocks do not display the master's scaling coefficient". | 2026-09-17 | official |
| S10 | [ForeverChanges, warlock](https://foreverchanges.pro/class/warlock), build 70338, data from 10-08: Improved Imp no longer raises Blood Pact. Demonic Knowledge raises "your spell damage and your Demon pet's spell damage". | 2026-10-08 | client text |

### Client flag (S3)

`Attributes_3 & 0x100` is the spell attribute that limits a spell to player targets. In 70338 it
is new on these spells (all ranks unless a rank is named):
- Gift of the Wild; Prayer of Fortitude; Arcane Brilliance; Prayer of Spirit; Prayer of Shadow
  Protection.
- Greater Blessing of Kings, Might, Wisdom and Salvation.
- Battle Shout; Leader of the Pack (24932); Moonkin Aura (24907).
- Devotion Aura; Fire, Frost and Shadow Resistance Aura.
- The buff spells of the Strength of Earth and Grace of Air totems; Aspect of the Wild.
- Every Scroll of Agility, Strength, Protection (Armor), Intellect, Spirit and Stamina, ranks I to
  IV.

The flag is **not** on these:
- The single-target Mark of the Wild, Power Word: Fortitude, Arcane Intellect, Divine Spirit and
  the small Blessings.
- Trueshot Aura (Forever's is ranged attack power only), Blood Pact, Paranoia, Retribution Aura,
  Sanctity Aura, Mana Spring, Windfury Totem, Tranquil Air.
- Inspiration, Power Infusion, Bloodlust, Juju Power and Juju Might.
- Every world buff.

S5 and S6 report that the single-target buffs fail on pets too ("Invalid Target"). The block is
therefore wider than this flag and partly server-side. The flag alone settles every group-cast
buff, and it applies to any non-player unit, demons included.

### Scaling auras (S4)

Effects in client order:
- **Hunter Pet Scaling 415429:**
  - max health; melee attack power; spell damage (all magic schools);
  - six resistances; melee and spell hit; melee and spell haste; melee and spell crit;
  - +50 max Focus; expertise (two effects); dodge.
- **Warlock Pet Scaling 416189:** the same list without the Focus and dodge effects, plus
  Intellect, mana regeneration and healing done.

The amounts come from the server. Only the hunter's numbers have been measured (S7: attack power
and crit). Hit, haste and expertise are slots that can follow the owner. Nobody has measured them.

## Matrix

Legend:
- **No** = the buff does not land on that pet in Forever.
- **Via owner** = it does not land on the pet, but it raises the owner's stats, and the pet takes
  a share of those through its scaling aura.
- **Conf.** = confidence. Every row is tier **Forever beta**: no row needed a fallback for its
  verdict. SoD only fills in the scaling numbers (section 5).
- In SoD, as in Era, pets could still be buffed (SoD fallback and Era fallback). The block is
  new in Forever, so neither fallback applies to the "does it land" question.
- **Sim now:**
  - The hunter branch sets `IgnoresPlayerBuffs`, which skips all raid, party, individual and
    world buffs on the pet. It inherits 10% of the hunter's ranged attack power and all of the
    hunter's melee crit.
  - The warlock branch gives the demon the preset's raid buffs. Battle Shout, Trueshot Aura and
    Leader of the Pack are left out of the preset. The demon inherits nothing.

### 1. Raid and party buffs

| Buff | Hunter pet | Demon | Forever source | Conf. | Sim now (hunter / warlock) |
|---|---|---|---|---|---|
| Gift of the Wild | No; via owner (Agility to AP and crit) | **No** | S3 flag, S1, S5 | high | no / **yes, Improved** |
| Mark of the Wild (single) | No | No (reported) | S1, S5, S6; no client flag | medium for demons | no / covered by Gift |
| Prayer of Fortitude | No | **No** | S3, S5 | high | no / **yes** |
| Power Word: Fortitude (single) | No | No (reported) | S5, S6 | medium | no / **yes** |
| Arcane Brilliance | No | **No**; via owner Intellect (scaling has an Intellect slot) | S3 | high | no / **yes** |
| Arcane Intellect (single) | No | No (reported) | S5 | medium | no / **yes** |
| Divine Spirit / Prayer of Spirit | No | **No** | S3 (Prayer), S5 | high (Prayer) | no / **yes** |
| Greater Blessings (Might, Kings, Wisdom, Salvation) | No; via owner | **No** | S3, S1 | high | no / **Kings and Wisdom yes** |
| Small Blessings | No | No (reported) | S5, S6 | medium | no / **yes** |
| Battle Shout | No; via owner (10% of AP, if melee AP is the highest) | No | S3 | high | no / no |
| Trueshot Aura | No effect: Forever's is ranged AP only (client 1299346 and 1299348: aura 124, 30 or 40). Via owner: 10% of it, 3 to 4 AP | No effect | client 70338 | high | no / no |
| Leader of the Pack (Forever: 3% crit to spells and melee) | No; **via owner, in full** (the pet takes all of the hunter's crit) | No; via owner (probably the warlock's spell crit) | S3, S7 | high (No), medium (via owner) | no buff, inherits crit / no |
| Moonkin Aura (Forever: 3% crit) | No; via owner, in full | **No**; via owner | S3 | high | no / **yes** |
| Strength of Earth, Grace of Air | No; via owner | **No** | S3 | high | no / caster group has none |
| Windfury Totem | Not flagged, and a pet has no weapon to imbue | n/a | client 70338 | medium | no / no |
| Mana Spring, Tranquil Air | Not flagged | Not flagged, so they may still reach the demon (mana only) | client 70338 | low | no / Mana Spring yes |
| Devotion and resistance auras, Retribution and Sanctity Aura | No (Devotion and resistance flagged); Retribution and Sanctity not flagged | Same | S3 | high for the flagged ones | no / **yes** (Sanctity: none in Forever; Retribution: damage taken only) |
| Blood Pact | Not flagged (Stamina only, no DPS) | Not flagged | client 70338, S10 | medium | no / yes |
| Paranoia | Not flagged (detection only) | n/a | client 70338 | - | - |
| Shadow Weaving, Inspiration, Power Infusion | Not flagged; Shadow Weaving is a debuff (see 6) | Same | client 70338 | - | n/a |
| Furious Howl (wolf, AP reduced 40% on 09-24) | The pet's own buff, not flagged | n/a | S8 | medium | hunter sim: the pet's own ability |

### 2. Consumables on pets

| Item | Hunter pet | Demon | Source | Conf. | Sim now |
|---|---|---|---|---|---|
| Scrolls (Agility, Strength, Protection, Spirit, Intellect, Stamina, I to IV) | No | No | S3 flag on every rank; S5's closing note names scrolls | high | `ApplyPetConsumeEffects` still runs for both pets. It is off only because the presets leave the inputs empty. |
| Juju Power, Juju Might | No (reported) | No (reported) | S6 (JuJu Power); not flagged in the client, so the block is server-side | medium | Same as above |
| Pet food (happiness) | Not a buff; the sim models happiness as a flat x1.25 | n/a | S1 | - | x1.25 |
| Forever pet consumables | None found in client or news | - | - | low | - |

### 3. World buffs

| Buff | Hunter pet | Demon | Source | Conf. | Sim now |
|---|---|---|---|---|---|
| Rallying Cry, Warchief's Blessing, Spirit of Zandalar, DM Tribute, Songflower | Not flagged in the client. By S1 and S7 they don't land on a hunter pet. On the owner they reach the pet through crit (Rallying Cry and Songflower, in full) and through 10% of the hunter's AP | Unknown directly; probably the same (S5) | S1, S7 | medium | hunter: none / warlock: off in the presets |

The world buffs are off in every preset, so this changes no number.

### 4. Owner talents and abilities that buff the pet (Forever)

| Talent | Applies? | Source | Sim now |
|---|---|---|---|
| Unleashed Fury (pet damage +3% a rank, also the Hawk) | Yes. Fixed on 10-08: "now correctly increases your pet's damage" | S1, S8 | hunter: x(1 + 0.03 a rank) |
| Ferocity (pet crit +2% a rank, also the Hawk) | Intended yes. **Known issue on 10-08: it does not raise the pet's crit on the beta** | S1, S8 | hunter: +2% a rank |
| Focused Fire (hunter and pet damage while the pet is active) | Yes | S1 | hunter: +1% a rank |
| Aspect of the Monkey (pet gets half) | Yes, no DPS | S1 | - |
| Ferocious Inspiration | **Not in Forever's tree** | forever-wiki/classes/hunter.md | - |
| Bestial Wrath, Intimidation (+100% crit on the next attack) | Yes | S1 | hunter branch |
| Unholy Power (all demon damage +2% a rank, Classic 4% melee only) | Yes | S2, S10 | warlock: x(1 + 0.02 a rank) |
| Master Demonologist (both warlock and demon) | Yes | S2, S10 | warlock branch |
| Soul Link (both +3% damage) | Yes | S10 | warlock branch |
| Demonic Knowledge (warlock and demon spell damage, up to 33/67/100% of level) | Yes, on both | S2, S10 | warlock: on both |
| Demonic Pact (31) | Keeps the sacrifice buff when another demon is summoned; no pet stat | S2 | - |
| Fel Vitality, Demonic Energies | Health, mana and healing only | S2 | - |
| Fel Intellect, Fel Stamina | Removed or absent in Forever's tree (not in forever-wiki/classes/warlock.md) | - | - |

### 5. Stat inheritance

- **Hunter pets:**
  - They take a share of the hunter's stats, server-scripted (S1, S4).
  - Measured by beta testers (S7): 10% of the hunter's **highest** attack power, all of the
    hunter's crit, 2 health per Stamina point and 30% of the hunter's armor.
  - The client aura also has slots for hit, haste and expertise that can follow the owner (S4).
    The 09-17 known issue says the sheet doesn't show "the master's scaling coefficient" for pet
    hit (S9), so pet hit follows the hunter's hit by some coefficient.
  - So a buff on the hunter does reach the pet:
    - 10% of the hunter's attack power. That covers Blessing of Might, Battle Shout if melee AP is
      the highest, Agility from Mark and Kings, and Trueshot Aura's ranged AP.
    - All of every crit buff: Leader of the Pack, Moonkin Aura, Agility crit, Rallying Cry,
      Songflower.
  - The sim already inherits AP and crit dynamically, so these reach the pet in the sim. It uses
    ranged AP rather than the higher of the two, which is right for every current build (S7 and
    the comment in sim/hunter/pet.go).
- **Warlock demons:**
  - They take a share of the warlock's stats (S2 official, S4, S9; S8 shows a demon ability
    scaling from the warlock's spell healing).
  - Coefficients are not known: the client is server-scripted and no beta measurement exists. The
    aura's slots are health, melee AP, spell damage, resistances, hit, spell hit, haste, crit,
    Intellect, mana regeneration and expertise.
  - **SoD fallback**, the same aura 416189 in Season of Discovery:
    - Blizzard hotfixes, 2024-12-04
      ([source](https://worldofwarcraft.blizzard.com/en-us/news/24148555/hotfixes-december-4-2024)):
      - "Warlock pets will now correctly inherit their master's Fire or Shadow spell hit chance
        based on whichever is higher."
      - "...gain full value of spell hit chance for their spells."
      - "...gain 80% of their master's spell hit as physical hit."
    - Player test, SoD forums, 2024-01-06
      ([thread](https://us.forums.blizzard.com/en/wow/t/why-does-hunter-get-pet-scaling-but-not-warlock/1759290)):
      - demons take the warlock's Stamina, Intellect and armor;
      - "one attack power per 2 of your highest spell power";
      - spell power "around 1/4". The poster says this one is unsure.
      - A Succubus lost about 25% of its health and 14% of its armor when the warlock undressed.
    - No SoD source gives crit inheritance.
  - SoD fallback for hunters, February 2024 hotfix (via Wowhead and Warcraft Tavern): pets take
    30% of the hunter's armor and 20% of the hunter's Stamina. Forever's measured 30% of armor
    agrees. The beta's 2 health per Stamina point is the Forever number and wins.
  - **The sim inherits nothing.** That makes this the bigger gap: the demon's damage misses the
    warlock's spell power, hit and crit.

### 6. Target debuffs

Sunder Armor, Faerie Fire, Curse of Recklessness, Expose Armor, Curse of the Elements and Curse
of Shadow are on the target. Nothing in S1 to S10 touches them, and the player flag in S3 is about
who can be **buffed**. Pet damage gets them as usual (conf. high for armor debuffs, medium for
school debuffs).

The open part is Forever's own rule that Improved Shadow Bolt, Shadow Weaving, Improved Scorch
and Winter's Chill help only the player who applies them (the warlock presets' comment). It is
not known whether the warlock's own Improved Shadow Bolt counts for its own Succubus's Lash of
Pain. A log would answer it. The sim doesn't give it to the demon.

## Sim changes recommended (for the orchestrator to rule on)

1. **Warlock demons take no player buffs under Forever.**
   - Set `IgnoresPlayerBuffs` on the warlock's demons, the field the hunter branch adds in
     sim/core/pet.go.
   - Rewrite standards.md item 14 to say "no pet takes player buffs".
   - DPS: **down** for every warlock spec with an active demon. The demon loses Kings, Gift of the
     Wild, Moonkin crit, Arcane Brilliance and Divine Spirit. Its loss is likely larger than the
     ~7 DPS the melee party buffs were worth. The warlock's own DPS doesn't change, except through
     Soul Link and Master Demonologist, which don't depend on buffs.
2. **Warlock demons inherit the warlock's stats.**
   - Forever's coefficients are unknown. The SoD fallback placeholder, labelled as such:
     - spell hit 100% (the higher of the warlock's Fire and Shadow hit);
     - physical hit 80% of the warlock's spell hit;
     - AP 0.5 per point of the warlock's highest spell power;
     - spell power about 25% (player-measured, weak).
   - Crit has no source in either tier, so leave it out until a log.
   - DPS: **up**. Together with change 1 the net direction is unknown until there are
     coefficients. Shipping change 1 without change 2 would understate demons.
3. **Gate `ApplyPetConsumeEffects` on Forever for both classes**, so a UI user can't put Juju or
   scrolls on a pet. DPS: none for the presets (already off). It only stops a wrong option.
4. **Hunter, optional:** inherit the hunter's hit, once a log gives the coefficient (S4, S9). DPS:
   **up**, pet only.
5. **Hunter, keep as is:** Ferocity at +2% a rank, as intended, with the 10-08 known issue noted
   in the preset text. Unleashed Fury at 3% a rank (fixed on 10-08).

## Open questions that need a beta log or a test

1. Can a single-target buff (Mark, Fortitude, Intellect, small Blessing) still be cast on a
   **demon**? The client allows it, S5 says "Invalid Target" for warlock pets. A screenshot from a
   warlock's demon is enough.
2. Warlock Pet Scaling coefficients: what share of the warlock's spell power the Imp's Firebolt and
   the Succubus's Lash of Pain get, plus the demon's AP, hit and crit. A demon with its owner
   naked, then geared, is enough.
3. Does the hunter pet inherit the hunter's hit, and by what coefficient? Does it inherit haste
   (Rapid Fire, Quick Shots) or expertise?
4. Is the pet's AP 10% of the **higher** of the hunter's melee and ranged AP, as S7 says? This
   matters only if Battle Shout or Blessing of Might ever made melee AP the higher one.
5. Do Mana Spring and Tranquil Air (unflagged) still reach a demon in the warlock's party? This
   affects demon mana only.
6. Does the warlock's own Improved Shadow Bolt (or Shadow Weaving) count for its own demon's
   Shadow damage under Forever's "helps only the player who applied it" rule?
7. Is the Ferocity crit bug fixed in a build after 70338?

Not checked: SpellTargetRestrictions was downloaded but not compared with Era. No client field for
the block on the unflagged single-target buffs was found, so it is most likely a server-side rule.

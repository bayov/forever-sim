# Practice room

A realtime 3D room where you play your character against an encounter, driven by the same sim that makes the DPS
Report. You press your own keybinds, and the sim decides every outcome exactly as it does in a sim run.

It's a sim tool first. The point is to play a fight by hand under the sim's rules, to try a rotation or a talent
choice before using it in game, and later to practice a raid fight with friends. It is not a game we ship.

**How to read this doc:**

- Decisions marked "(the user, 2026-11-02)" are settled, and [Decisions](#decisions) lists them.
- Everything else is a proposal. [Open questions](#open-questions) lists what the user still has to decide.
- An implementer starts at [Milestones](#milestones). [Spikes](#spikes) are short experiments we run before building
  on an answer.
- Milestones are "Milestone 1" to "Milestone 4". "M2" alone always means WoW's model file format.

**Terms:**

- **Pull:** the moment the fight starts. It is time 0 in the sim. What triggers it depends on the fight mode.
- **Prepull window:** the time before the pull, when you walk around, buff and drop totems.
- **Bit-identical:** two runs give exactly the same numbers, down to the last digit.
- **Golden file test:** a test that compares output with a saved file that is known to be right.
- **Hot path:** code the sim runs millions of times per run, where even one extra check costs time.
- **p50 and p99:** the median, and the value only 1 in 100 cases is worse than.
- **Combat reach:** how far a unit's body extends from its center, as the game measures it.

## What it answers

- **What you really do in a fight under the sim's rules,** at your ping: your DPS from one fight, and where you lost
  time (idle GCDs, buffs that fell off, time out of range).
- **How a talent, gear or rotation choice plays by hand** before you respec in game.
- **Later: a raid fight**, with bots or friends in the other spots.

## Decisions

1. **Browser app, not a native client** (the user, 2026-11-02). It runs at a local URL from the practice server.
2. **Stack** (the user, 2026-11-02): TypeScript + Vite for the client, Three.js for 3D, our Go sim as the game logic.
   The sim runs as WASM in a Web Worker now, and natively in a Go server once we add multiplayer.
3. **Local only, with real WoW assets** (the user, 2026-11-02). Character models, items, spell icons, sounds and
   visuals come from the WoW Forever client on this PC. They are never deployed, committed or put in `dist/`.
4. **Keybinds and action bars come from the Forever client** (the user, 2026-11-02). Later, other players import
   their own. Keys the client doesn't set use Blizzard's default bindings. Which spell sits in which slot, and which
   key presses each button, can be edited in the room.
5. **First milestone is a basic 3D room** (the user, 2026-11-02): enh shaman, target dummy, flat arena, your keybinds,
   a live DPS meter and a fight summary. No 2D milestone first.
6. **Don't disrupt the regular sim** (the user, 2026-11-02). The practice room is kept as separate as possible. It
   uses the game rules the sim codes, as a consumer of the sim. See [Separation](#separation-from-the-regular-sim).
7. **One iteration, no averages** (the user, 2026-11-02). The room sims only the fight you play. It doesn't run other
   seeds or compare with the sim's average, for now.
8. **Ping is modeled** (the user, 2026-11-02), as a setting with your in-game latency.
9. **You do the prepull, and you start the fight** (the user, 2026-11-02). Totems and buffs before the pull are your
   own casts. Nothing in the room acts for you.
10. **Two fight modes** (the user, 2026-11-02). See [Fight modes and the pull](#fight-modes-and-the-pull).
    - **Free (the default):** like a training dummy in game. The fight starts with your first attack and goes on until
      you click Stop.
    - **Encounter:** a "Start Encounter" button starts a countdown (5s by default, you set it), and the fight lasts a
      set length you choose.
11. **It plays like WoW** (the user, 2026-11-02). Movement, casting, the camera and the small quirks match the game as
    closely as we can, so you play the way you do in game and the sim result is accurate. See
    [Playing like WoW](#playing-like-wow).
12. **Full manual control** (the user, 2026-11-02). Your character has no rotation. As in WoW, it casts only what you
    press. Auto attack starts off. It turns on with a melee ability, a `/startattack` macro (for example on Earth
    Shock) or a right click on an enemy, and then your character swings whenever it's in melee range.
13. **Imbues are a setting you can override** (the user, 2026-11-02). You enter the room with the imbue from the
    settings already up. Casting a different imbue replaces it, at any time, before or during the fight.
14. **Your whole class toolkit is castable** (the user, 2026-11-02). You can cast anything your class knows at your
    level, as in game, not only what the sim's rotation uses. See [Your toolkit](#your-toolkit).

## Separation from the regular sim

The practice room is a consumer of the sim core, the way the DPS Report is a consumer of sim runs. It builds a normal
sim from a normal config and drives it from outside through what the core already exports. The game rules (spells,
procs, auras, resources, the APLs of every other unit) stay in the core, and the practice room doesn't change them.

### Where things live

| Part | Place | Notes |
| --- | --- | --- |
| Session, manual control, world, fight summary | `sim/practice/` | imports `sim` and `sim/core`, nothing imports it |
| WASM | `practice.wasm`, built from `sim/practice/wasm/` | `//go:build wasm` tag, written to `dist/practice/` |
| Server | `cmd/practice/` | a new small Go HTTP server, nothing shared with `sim/web` |
| Page | `ui/practice/practice.html`, own Vite config | built to `dist/practice/`, never to `dist/forever/` |
| Assets, game tables, attempts, your edits | `~/forever-practice/assets/`, `db2/`, `attempts/`, `settings/` | outside the repo |

The page's entry is `practice.html`, not `index.html`. The regular build takes every `ui/**/index.html` as an input
(`vite.config.mts:119`) and `.gitignore` ignores `ui/*/index.html`.

Generated files live outside the repo, like `~/forever-logs`. jj tracks every new file under 8 MB on its own, in all
five workspaces, so extracted game files inside the repo would end up in commits.

### Rules for touching shared code

- The class sims, the DPS Report, rotopt and the leveling sim give bit-identical results after any practice change.
  Every core touch is checked with `make test` (24 golden files, kept to 5 significant digits) and then with the exact
  check: `identical()` in `tools/dps_report/run_assumptions.py`, which runs every DPS Report request on two binaries
  (one built from `master`, one from the change) and compares the full results at fixed seeds. `tmp/lvwork/abcmp.sh`
  (local, untracked) does the same for 92 DPS and 60 leveling requests.
- A core change is only a small, opt-in hook that does nothing in a normal run, with no extra work in the hot paths.
  Each one is listed in the table below with why we need it.
- When the practice room needs a game rule the core doesn't have (for example a facing check for casts), it checks
  the rule on the practice side, before an input reaches the sim. A rule that would also make the regular sim more
  correct goes into the core as a normal sim change with its own review, never as part of practice work. The first
  one: a totem buff replaced or expired before time 0 stays up until 0, because the aura hold search starts from 0
  (`aura_holds.go:36`). It should start from the earliest sim time (`startingCDTime`). It changes results only for
  holds that end before 0, and it still gets the `identical()` run.
- In `ui/core` we import what's already exported. When reuse would need a refactor of a shared file (like
  `replay.tsx`), we copy the piece into `ui/practice/` instead.
- The regular sim UI doesn't change. Settings come into the practice page through the class sim's existing export
  (link or JSON). A "Practice" button in the class sim needs the user's OK first.

### Core touch points (the whole list)

| Hook | Why | Cost in a normal run |
| --- | --- | --- |
| `Simulation.NextEventTime()`, read-only | `Run(until)` stops before the next event without moving the clock | none, nothing calls it |
| `Unit.ManualControl bool`, default false | `DoNextAction` returns at once for your unit, after the due hard cast block (`apl.go:209`) | one bool check per `DoNextAction` |
| `AutoAttacks.SkipPullStart bool`, default false | the sim's pull at time 0 leaves your auto attacks off | one bool check per unit at the pull |
| Spell limit 200 to 400, timer limit 100 to 200 (`spell.go:202`, `cooldown.go:21`) | enh at 60 already has 154 spells and 49 timers, and the whole toolkit adds about 60 spells | none, they're sanity checks |
| Shaman: `SwapMainHandImbue(sim, imbue)` | see [Your toolkit](#your-toolkit) | none, nothing calls it |
| Shaman: exported `PlaceTotem(sim, slot, spell, duration)` | totems the sim doesn't have take their element's slot, Fire Nova and Totemic Recall see them | none |
| Shaman: `FireTotemInReach func(totem, target) bool`, nil by default | fire totems and Fire Nova hit only targets in reach (`totems.go:80` is always true outside leveling) | one nil check per fire totem tick |
| Shaman: Nature's Swiftness skips its "no cast speed buff up" condition under `ManualControl` (`talents.go:380`) | that condition is a rotation helper, and it would refuse a legal press | one bool check per cast |

Other classes get their own class hooks when we add them (blessings, auras, aspects, poisons).

Extra spells and auras come in through `core.EnvHooks` (`environment.go:180`), which already exists for tools that
watch a sim from outside, like `tools/logcheck`. It runs before the sim is finalized, while spells and auras can still
be registered.

`ManualControl` is needed because an empty APL isn't idle. Each time it finds nothing to do, it waits 50ms through
`WaitUntil`, which sets the GCD (`apl.go:227`, `gcd.go:90`). It runs after every swing, cast, and energy or rage tick
(`attack.go:432`, `cast.go:368`, `energy.go:166`, `rage.go:440`). So the GCD would look busy most of the time, a key
press at the GCD's end would lose to it, and the log would get "Pausing GCD for 50ms" 20 times a second.

`SkipPullStart` keeps the pull from starting your swings at time 0. A session action at 0 with a priority above the
pull's would also work (`sim.go:689`), but the hook says what it means. The core also turns auto attack on and off on
its own in places (Ravager, Shadowmeld) and has no exported "is it on" flag, so the session keeps its own flag.

### Other shared files (one time, each its own commit)

- `tsconfig.json`: exclude `ui/practice`, which gets its own `ui/practice/tsconfig.json`. Otherwise one practice type
  error breaks `make` for every session (`makefile:53` runs `tsc --noEmit` on the whole tree).
- `makefile`: filter `ui/practice` out of `UI_SRC`, and `sim/practice` out of `lib.wasm`'s prerequisites (it takes
  every `.go` file under `sim/`, so each practice edit would relink it). Add the targets `practice-wasm`, `practice`
  (wasm, page, server) and `practice-test`.
- `package.json`: add `three` and `@types/three` once, both pinned at 0.183.x. Newer `@types/three` needs TypeScript
  5.6, and the repo pins 5.2.2. Changing the lock file makes every session's next `make` run `npm ci` again, so it
  happens once, on its own.
- `go.mod` and `go.sum`, only if a practice Go dependency comes up.
- `tools/dps_report/build_artifact.mjs`: the deploy guard. It fails when the build's module ids contain `/ui/practice/`
  or the page contains `forever-practice`. About five lines.

Every Go file in `sim/practice` and `cmd/practice` carries the build tag `practice`, not only the tests. `make test`
runs `go test ./sim/...`, which compiles every package under `sim/` even when its tests are skipped, so one practice
compile error would fail it for every session. A cgo CascLib binding, if we ever need it, sits behind the same tag.
Practice messages are plain Go and TS types, never a `.proto` in `proto/`, because `protoc ./proto/*.proto` writes into
`sim/core/proto`.

`make practice-test` runs `go test -tags with_db,practice ./sim/practice/... ./cmd/practice/...` and the practice
type-check. "`lib.wasm` unchanged" means `go list -deps ./sim/wasm | grep practice` prints nothing.

## Architecture

```
 Browser tab (main thread)                           Practice worker
 ┌──────────────────────────────────────┐            ┌─────────────────────────────────┐
 │ Input: keydown ──► input message ────┼── input ──►│ practice.wasm (Go)              │
 │                                      │            │  practice session               │
 │ HUD state store ◄── event batches ◄──┼── events ◄─│   sim: your unit takes inputs,  │
 │   (predicts GCD and cast bar)        │            │        every other unit its APL │
 │ Three.js scene, HTML overlay         │◄─ units ───│   world: positions, facing,     │
 └──────────────────────────────────────┘            │          range, movement        │
   ▲ assets, keybinds, bars from the practice server │   own clock, pushes events      │
                                                     └─────────────────────────────────┘
```

The sim and the world model both live in Go. The browser sends only your inputs and your own movement, and gets every
unit's state back. That keeps one owner for positions when the boss moves (Milestone 2), bots stand around you
(Milestone 3) and a server judges range for several players (Milestone 4).

### Starting a session

The engine already has `SimOptions.Interactive` and a session API in `sim/lib/library.go`. We don't use the flag: it
stops the rotation of every character, pets included (`character.go:525`). We follow the session API's steps, which
are all exported.

**The request.** The class sim's export leaves out the item database (`ui/core/player.ts:1609`), and `practice.wasm`
has none. So the page sends the link or JSON to the practice server (`POST /api/request`, built with `-tags with_db`).
It turns it into a `RaidSimRequest` the way `tools/rotopt/run.go:44` does, and fills `player.database` the way
`getDatabase` does (`library.go:96`). Then the server makes these changes to the request:

- Your unit's rotation becomes an empty APL: `&proto.APLRotation{Type: TypeAPL}`, with no priority list and no prepull
  actions. A nil rotation panics (`cast.go:368`), and the imported prepull actions would cast for you, because the APL
  registers them when the sim is built (`apl.go:111`).
- `EndFightAtHealth` 0 and no duration variation, so the fight length is exactly the mode's.
- PvP melee downtime 0, because it would move you in and out of range on its own (`pvp.go:68`). The world decides
  range.
- A leveling encounter is refused.
- In Free mode, execute phases are off, or on from the start with the "Execute range" setting.

What stays as in the request:

- Raid hits (the enh request has 2 hits a minute on you) and a tank assignment. They feed Lightning Shield orbs and
  Water Shield mana, so they keep the room's DPS in line with the normal sim. The HUD shows "Raid hits 2/min".
- Party cooldowns on your unit (Power Infusion, Innervate, Mana Tide) are another player's casts. The session fires
  them on the sim's own rule (`ShouldActivate`), as the APL's `autocastOtherCooldowns` does for them.

**The steps:**

1. `sim.RegisterAll()`, then `NewSim` from that request. The fight length is the set length in Encounter mode, and a
   60 minute cap in Free mode.
2. A nonzero seed picked by the session and stored with the attempt. Seed 0 would mean "pick from the clock", and the
   picked seed isn't readable from outside (`sim.go:30`).
3. `Reset()`, then `PrePull()`. Only the class code's own prepull actions are left, since the APL is empty.
4. Set `CurrentTime` to the earlier of its value and minus 10 minutes, the start of the prepull window. This comes
   after `PrePull`, because `PrePull` sets the clock when prepull actions exist (`sim.go:686`, a druid talent
   registers one).
5. Set `sim.Log`, since only the sim's own run sets it (`sim.go:373`). Write your distance and facing, since `Reset`
   clears them (`unit.go:553`).
6. Your unit gets `ManualControl`, so it never acts on its own, and its auto attacks get `SkipPullStart`, so they start
   off. Other units (the boss, pets, later the raid bots) keep their own APL.
7. The imbue hook turns off every imbue but the active one, since each imbue aura turns itself on at `Reset`
   (`windfury_weapon.go:105`).

### Fight modes and the pull

The sim's fight always starts at time 0: its encounter length, execute phases and pull are all counted from there.
But you decide when the fight starts, so the room can't know in advance how long your prepull is. It solves that by
shifting time once it knows:

1. Before the pull, the session runs the sim live in the prepull window. You walk around, drop totems, buff, and see
   your mana and cooldowns move.
2. When the pull happens at time `t` (below 0), the session swaps in a fresh sim with the same seed. It feeds it your
   inputs so far, each moved by minus `t`, and starts its clock minus `t` earlier too. The same seed and the same
   inputs in the same order replay exactly, so the fresh sim reaches 0 in the state the old one had at `t`. The pull
   is now at exactly 0, and everything before it keeps the same spacing.
3. If you stay in the prepull window for 10 minutes, the session swaps to a new 10 minute window the same way.
   Inputs older than 40 minutes are dropped, and what they left behind (the current imbue, auto attack on) is set
   again at the window's start. No input is replayed earlier than minus 50 minutes, because before minus 60 minutes
   every cooldown and the GCD look busy (`startingCDTime`, `cooldown.go:7`).

The fresh sim is built (`NewSim`, `Reset`, `PrePull`) in idle time after the session starts and after each swap, and
built again when a fight setting changes. Go WASM has one thread, so replaying your inputs into it happens on the
frame of the pull. Spike 4 measures the worst case.

The worker tells the client with a `rebase` message, so the client moves its clock, its HUD times and the inputs it
has in flight by minus `t`, and drops the log lines of the thrown-away sim.

**What triggers the pull.** Any harmful result against an enemy before time 0: a hit, a miss, a tick or a debuff,
from your casts, your totems or your procs. A Lightning Bolt cast before the pull counts from the moment it lands,
like the in-game meters count. A Magma Totem dropped early pulls with its first pulse. Through `core.EnvHooks`, the
session puts a watcher aura on each enemy that reports the first such result. The session then stops, swaps, and
runs the fresh sim up to 0, where that same hit happens again and counts (the sim counts damage from time 0 on,
`spell_result.go:542`).

Swings are the one exception. The core holds swings to time 0 or later (`attack.go:756`), so a swing never lands
before 0. So the moment you have auto attack on, the target in melee range and in front of you, before the pull, is
the pull, and the first swing comes at once, as in game.

A cast that fails a check never pulls, because nothing harmful reached the enemy.

**Free mode** (the default):

- You're in Free mode until you click Start Encounter.
- The pull is the first harmful result.
- The fight goes on until you click Stop (or a key bound to it). Stop ends the fight at that moment. The session sets
  `sim.Duration` to the stop time and then runs `Cleanup()`. `Cleanup` moves the clock to `Duration` and DPS divides
  by it (`sim.go:716`, `metrics_aggregator.go:35`), so without that the summary would cover 60 minutes.
- The 60 minute cap ends the fight if you never click Stop.
- The dummy never enters execute range, as with a training dummy in game. The "Execute range" setting puts it there
  from the start, for classes with execute abilities.

**Encounter mode:**

- Clicking "Start Encounter" starts the countdown (5s by default). The pull is at the countdown's end, so the swap
  happens on the click.
- A harmful result during the countdown is an early pull, as in game. The fight starts right then (one more swap),
  and the HUD says so.
- The fight ends on its own at the set length, as in a sim run. The session ends it itself once its clock reaches the
  length, because the sim only notices the end when it steps an event past it. Execute phases follow the encounter's
  settings.
- Stop still works, and marks the attempt as stopped early.
- The set length starts at the imported fight length.

In both modes, auto attack only swings when you have turned it on and you're in melee range.

### Running in real time

`sim/practice` holds a stateful session:

- **Input(input)** puts the input in the session queue at its apply time, in whole nanoseconds. A queued cast runs at
  the GCD's or the cast's end instead. The session queue holds every session action (inputs and world changes),
  ordered by time and then by arrival. It keeps one pending action in the sim, at priority -100, below everything the
  core uses (`ActionPriorityLow` is -1). A live run adds inputs as they arrive and a replay adds them all up front, and
  the queue makes both run ties in the same order. It goes through `Step` like any other event, so the other units'
  idle check skipping (`apl_idle.go`) sees it.
- **Run(until)** runs events whose time is before `until` and stops. It reads the next event time from
  `NextEventTime()` and never moves the clock to `until` itself. Moving the clock to frame times would settle regen
  and expire auras in frame sized steps, and then results would depend on frame rate.
- **The end** is final: Stop, the set length in Encounter mode, or the 60 minute cap. An input at or after the end is
  dropped. `Cleanup()` adds up the metrics.

The worker keeps its own clock and pushes events as they happen, instead of waiting for one call per frame. It steps
the sim on each input message, which runs at once. For the sim's own events (swings, ticks) it sets a timer to the
next event's time. A timer can fire up to 4ms late, which is under one frame at 144Hz, and the screen only updates
once a frame. A loop that never sleeps would keep a whole CPU core busy next to WoW and Discord. If spike 4 finds
timers too late, the worker waits with `Atomics.wait` and a timeout instead, which needs COOP and COEP headers from
the practice server and a shared buffer for inputs.

The worker's clock is wall time minus the time it was paused, and one catch-up step is capped at 0.25s, as the Replay
view does (`replay.tsx:963`). The worker maps each press time to sim time: the press's wall time, minus paused time,
minus every swap's shift. The session keeps running when the tab is hidden or loses focus, like the game does (the
user, 2026-11-02: the Searing Totem stopped shooting in the background). Losing focus lets go of the movement keys,
because their keyup never comes.

Between events the HUD shows values like energy and the cast bar by computing them for the display time, without
changing sim state. Aura expiry isn't an event in the sim (it happens at the next event), so the HUD times expiries
itself. The live DPS meter sums spell metrics as `getDamageDone` does (`library.go:301`), because unit metrics only add
up in `Cleanup`.

The worker is a dedicated one running `practice.wasm`, not one from the sim's `WorkerPool` (`ui/core/sim.ts:123`). The
WASM entry points are synchronous `js.FuncOf` calls in the style of `raidSim` (`sim/wasm/main.go:133`).
`practice.wasm` registers every class (about 39 MB, 6.4 MB gzipped), so any class sim's export works and Milestone
3's bots work. One worker is kept for every attempt in a visit, so the module loads once. The live heap is about
1.3 MB, so garbage collector pauses come from per-event allocations (log formatting, proto marshalling), not from the
database.

### Casting your spells

Each cast input names a spell id. The session looks it up in its own toolkit table, never with `Unit.GetSpell`: the
Flametongue and Frostbrand proc spells share the imbue's spell id (`flametongue_weapon.go:51`), so `GetSpell` could
return the proc and land a hit on the dummy. Imbue casts are registered with their own tag, so the metrics and the
Replay view keep them apart from the procs.

Then the session runs its own checks, in this order, and answers with the game's error text when one fails:

1. Moving or in the air, for a spell with a cast time.
2. Range and facing (see [Playing like WoW](#playing-like-wow)).
3. GCD and cooldown timers, cost (`Cost.CostFailureReason`, `mana.go:443`), already casting.
4. `Spell.CanCast`.

Only then does it call `Spell.Cast`. A failed `Cast` can use up the cooldown, because the cooldown is set before the
GCD check (`cast.go:273`, `cast.go:289`), so `Cast` is never called without these checks passing. During a channel,
the session cancels the channel first (`ChanneledDot.Cancel`), since `InterruptHardcast` doesn't stop channels.

When a cast-time spell finishes, the game checks range and facing again (check). The session runs that check at the
cast's end time, before the cast completes, and interrupts the cast with the error when it fails.

### Your toolkit

You can cast anything your class knows at your level (the sim character's level), at the ranks you know. The room's
spellbook shows all of it. Each spell falls in one of three kinds:

1. **Spells the sim has** (most class spells: shocks, Lightning Bolt, Stormstrike, every totem the sim models,
   Healing Wave and the other heals, shields). The shaman class code registers all of them, whatever the rotation
   uses (`shaman.go:276`), and no class code casts them on its own (its automatic casts are all procs). This kind is
   every non-passive spell in your unit's spellbook, at the sim's ranks. They're cast as described above.
   - The sim's rank rules win over the client's. It drops ranks Forever doesn't have (the Ahn'Qiraj book ranks,
     `earth_totems.go:21`) and ranks the leveling training schedule marks untrained (`spell_ranks.go:159`). Such a
     rank shows greyed in the spellbook with the reason.
   - Imbues and Lava Burst exist in the sim only at the top rank (`HighestKnownRank`). Their lower ranks show greyed.
2. **Spells the sim keeps as a setting.** Shaman imbues are the first case: the class code builds only the imbue the
   settings pick into the character (`shaman.go:66`). The shaman hook `SwapMainHandImbue(sim, imbue)` makes any imbue
   castable:
   - At build time it registers the three imbues the settings didn't pick (`Register*Imbue`). Registering the picked
     one again would panic (`aura.go:485`).
   - After each `Reset` it turns off every imbue aura but the active one.
   - On a swap it sets the main hand's `TempEnchant` to the new imbue's enchant, which the core reads on every hit to
     decide whether an imbue blocks a totem's weapon buff (`totem_weapon_buffs.go:46`). It sets `ShamanImbue` and the
     imbue's proc mask.
   - Rockbiter's attack power is added to the weapon's stats when the character is built (`rockbiter_weapon.go:92`),
     so it can't be turned off later. The hook moves it onto Rockbiter's aura as a dynamic stat.
   - After a swap it makes each of your weapon totems still standing pick its buff again. Without that, swapping to
     Windfury Weapon under your own Windfury Totem would leave the raid's Flametongue Totem buff off, and swapping
     away would never bring the totem's buff up (`shaman.go:92`, `totem_weapon_buffs.go:204`).

   You enter the room with the settings' imbue up. Casting an imbue, at any time, replaces it with the game's GCD,
   mana cost and cast. Imbues go on the main hand, the only hand the sim models for shaman imbues (`shaman.go:78`).
   The sim's imbues never expire, while Forever's last 60 minutes. The room doesn't model that, and the HUD shows the
   time left since your last imbue cast.
3. **Spells the sim doesn't have** (Ghost Wolf, Purge, Earthbind Totem, Frost Resistance Totem, Cure Poison, Water
   Walking and the like). Only these come from the client's spell tables. The session registers them through
   `core.EnvHooks` with `RegisterSpell`, with the cost, cast time, GCD, cooldown and range from those tables. Talents
   that change them (Totemic Focus on every totem's cost, Improved Ghost Wolf, Guardian Totems, Earth's Grasp) are
   applied by the session, since the client rows don't include them. Effects that change the fight are coded in
   `sim/practice`:
   - Ghost Wolf: 40% faster runs (9.8 yd/s), a 3s cast (less with Improved Ghost Wolf) that moving stops. What a
     spell press does in wolf form follows `autoUnshift` (default 1, leave the form when needed). Whether that one
     press also casts the spell is a check.
   - A totem the sim doesn't have takes its element's slot through the shaman hook `PlaceTotem`, so Earthbind Totem
     replaces Strength of Earth, Fire Nova sees a Fire Resistance Totem, and Totemic Recall refunds it, as in game.
   - Effects that mean nothing against a dummy (Water Walking, Far Sight) only cost their GCD and mana.

**Consumables.** Flasks, elixirs, food and weapon oils come from the settings and are up from the start. Potions,
Demonic Runes, sappers and explosives are your own presses: they sit in your bags, show in the spellbook's bag tab, go
on bars as items, and are used with `useItem` by item id, which runs the consumable's cooldown spell. `/use <item>`
macro lines do the same. Other players' buffs come from the settings.

### Inputs, ping and the spell queue

Your ping splits the round trip the way the game does. An input's **apply time** is the press time plus half a ping.
The sim's answer reaches the screen half a ping after that: the page holds each update for half a ping, and draws at
the sim time half a ping back. So a cast stopped in the last ping of its bar still goes off, as in game, and the cast
bar shows that stretch in EllesmereUI's latency color. The default ping is 60ms. The fight bar shows the ping and the
frame rate.

- When the apply time is within `SpellQueueWindow` (400ms by default) of the GCD's end or the current cast's end, the
  cast fires exactly at that end and replaces any spell already queued. So a queued press costs no ping, as in game.
  Whether the window also queues a spell coming off its own cooldown (Stormstrike, shocks) is a check.
- When it's earlier than that, the cast fails with "Spell is not ready yet." (or "Ability is not ready yet.", see
  [Errors](#errors)).
- A press after the GCD has ended applies half a ping later, which is the ping cost you get in game.
- Off-GCD spells pressed during the GCD fire at their apply time, not queued.
- "On next swing" abilities (Heroic Strike, Maul, Raptor Strike) are toggles and never queue.
- An input whose apply time the sim has already passed (the worker was busy, or ping is 0) applies at the sim's
  current time.

The applied time is what the attempt stores, so replays don't depend on ping. The ping is stored too, for display.

Each input carries an id, and each answer carries the input's id. The client predicts what WoW predicts and rolls it
back when the sim says no. Whether WoW starts the GCD spin and the cast bar on the press or only when the server
answers is still a check (see [Checking the feel](#checking-the-feel)). Resources and outcomes always come from the
sim.

### Messages

The client resolves macros, `@mouseover` and modifiers at the press, and sends plain actions. Types in TypeScript
(`ui/practice/protocol.ts`) and Go (`sim/practice/protocol.go`):

```ts
// A unit in the session: you, the dummy, later bosses, bots and other players.
type UnitRef = { kind: 'player' | 'target' | 'enemy' | 'ally'; index: number };

type Input =
	| { id: number; pressedAt: number; kind: 'cast'; spellId: number; target: UnitRef }
	| { id: number; pressedAt: number; kind: 'useItem'; itemId: number; target: UnitRef }
	| { id: number; pressedAt: number; kind: 'useSlot'; slot: number; target: UnitRef } // /use 13
	| { id: number; pressedAt: number; kind: 'startAttack' | 'stopAttack' | 'clearTarget' | 'cancelCast' }
	| { id: number; pressedAt: number; kind: 'target'; target: UnitRef }
	| { id: number; pressedAt: number; kind: 'move'; velocity: [number, number]; facing: number }
	| { id: number; pressedAt: number; kind: 'jump' };

type Control =
	| { kind: 'start'; request: unknown; seed: number; mode: 'free' | 'encounter'; lengthMs: number;
		countdownMs: number; pingMs: number; executeFromStart: boolean }
	| { kind: 'startEncounter' }
	| { kind: 'stop' }
	| { kind: 'restart' }; // a new attempt with a new seed and the same settings

type EventBatch = {
	simTime: number;
	logLines: string[]; // the sim's text log, parsed by logs_parser.tsx
	answers: { inputId: number; ok: boolean; error?: string }[];
	units: { ref: UnitRef; pos: [number, number]; facing: number; airborneUntil?: number;
		nextSwingAt?: number; castEndsAt?: number }[];
	rebase?: { shiftMs: number }; // the pull swap moved sim time by this much
	pulled?: { early: boolean };
	ended?: { summary: unknown; stoppedEarly: boolean };
};
```

`pressedAt` is the key event's `timeStamp` plus `performance.timeOrigin`, the wall time of the press. The worker maps
it to sim time (see [Running in real time](#running-in-real-time)). Every other time in a message is sim time in
milliseconds, fractional. The sim works in whole nanoseconds, and the session converts.

### Replays and saved attempts

An attempt is one JSON file in `~/forever-practice/attempts/`:

```json
{
	"version": 1,
	"simCommit": "<git commit of the sim build>",
	"request": { "...": "the RaidSimRequest as proto JSON, after the session's changes" },
	"seed": 123456789,
	"pingMs": 60,
	"mode": "encounter",
	"setLengthMs": 180000,
	"countdownMs": 5000,
	"executeFromStart": false,
	"startNs": -612000000000,
	"endedAtMs": 180000,
	"stoppedEarly": false,
	"earlyPull": false,
	"inputs": [
		{
			"appliedAtNs": -8200000000,
			"input": { "kind": "cast", "spellId": 8512, "target": { "kind": "player", "index": 0 } }
		}
	],
	"result": { "...": "the fight summary" }
}
```

In Free mode, `setLengthMs` is the 60 minute cap and `countdownMs` is left out. `startNs` is where the final sim's
clock started, and input times are already shifted, so the pull is at 0.

A replay runs the same practice session without a clock, from `startNs`, and feeds it the stored inputs. It turns on
the log options the Replay view needs and gives a full `RaidSimResult`, so the attempt opens in the Replay view. A bug
report is a small file.

Exact replays hold only on the same sim build and the same platform. WASM and native Go can differ in the last bits of
`math.Exp` and `math.Log`, so a native Milestone 4 server and a browser may drift apart. Spike 5 checks how much. When
`simCommit` differs from the running build, the attempt keeps its result and its replay is marked as from another
build.

### World and movement

The world lives in `sim/practice`: ground positions with velocity, facing, an airborne flag with the jump's end time,
range, and where each totem stands. It follows the leveling world's model (`leveling_world.go`) but doesn't reuse its
code, which is tied to leveling sessions inside the core. We don't use `Unit.MoveTo`: its movement aura cancels auto
attacks (`movement.go:107`), which is wrong for strafing in melee.

Your movement goes to the session as timestamped velocity and facing changes, not as per-frame positions. The world
turns them into session actions at the times something changes (you reach or leave range, you turn away). That keeps
replays exact. The client moves your model at once and the sim agrees, as WoW trusts the client's movement.

How the world talks to the sim. Every write happens inside a session action, because `EnableAutoSwing` and
`InterruptHardcast` use the current sim time, and between steps that's the last event's time.

- **Distance:** the core knows melee range as a flat 5 yards (`IsInMeleeRange`, `pvp.go:30`). The world works out the
  real melee range (see [Range and facing](#range-and-facing)) and writes `Unit.DistanceFromTarget` as the center
  distance minus how much the real range exceeds 5 yards. Then the core's 5 yard rule agrees with ours. Other readers
  (Lightning Bolt travel time, `spell.go:678`, and explosive radii) are off by at most a yard or so, which is fine.
- **Melee range:** when you leave or enter it, the world calls the cancel and enable pair that PvP downtime uses
  (`pvp.go:68`), because auto attacks only check range when enabled (`attack.go:755`). Enabling keeps the old swing
  time (`max(swingAt, now)`), so leaving and coming back neither resets the swing nor gives a free one.
- **Facing:** it writes `PseudoStats.InFrontOfTarget` (`character.go:246`), which the core reads at each roll for
  parry and block.
- **Moving:** it calls `Unit.InterruptHardcast` when movement or a jump starts during a hard cast. `IsMoving` stays
  false, because we don't use `MoveTo`, so the session's own checks handle casting while moving.
- **Totems:** a totem stands about 2 yards from you, at a fixed angle for each element (check). Fire totems hit only
  targets in their reach, through the shaman hook `FireTotemInReach`: Searing Totem 20 yards, Magma Totem 8, and Fire
  Nova 10 yards around the totem (radii from the client spell tables). The session also checks Fire Nova's own 30
  yard cast range to the totem.

Known gaps: totem buffs don't drop when you walk away from the totem (buff reach exists only in leveling,
`shaman/totem_reach.go:17`). `DistanceFromTarget` is one number per unit, which limits fights with several enemies.
Both matter from Milestone 2 on, and the fix is decided then.

### Events out of the sim

The worker streams the sim's text log lines with each batch, and the client reuses `logs_parser.tsx`, which already
turns them into typed records (casts, damage, auras, resources). Typed events come later, only if parsing is too
slow.

The live HUD also needs the next swing time and the cast end time. The Replay view gets those by looking ahead in the
finished log, which a live stream can't do. The session reads them from the units after each batch
(`AutoAttacks.NextAttackAt`, gated by its own "attack on" flag, and `Unit.Hardcast`) and sends them in `units`.

### HUD overlay

The Replay view draws your EllesmereUI frames, your aura pack (ForeverAuras, the WeakAuras style addon the user runs)
and MSBT style combat text. Its model is
built once from a finished fight (`buildReplayModel`, `replay_model.ts:327`), and parts of it look ahead: swing timers,
cast pushback, and the bar layout chosen from cooldowns used in the fight. The draw functions are "state at time t"
lookups, so a model that only grows forward in time can feed them.

So we reuse the markup, CSS and helpers (`updateIcon` and `updateCastBar` in `replay.tsx`), and write an incremental
model fed by events in `ui/practice/`. Helpers that aren't exported are copied. Your real action bars are new code.

The Replay view's cooldown sweep is a `conic-gradient` driven by a CSS variable set every frame
(`_replay.scss:790`). At 144Hz that repaints every icon every frame on the main thread, next to Three.js. Step 7
measures it. If it's too slow, the sweeps move into one 2D canvas. Combat text animates only `transform` and
`opacity`.

### Rendering

- **Three.js.** It has no particle system, so particles come later from three.quarks or our own M2 particle code.
  Frame time with 40 characters is measured in spike 2, not guessed.
- **WoW models drawn directly**, not converted to glTF (see [Assets](#assets-local-only)). Forever's character
  models keep their skeleton in a separate `.skel` file and their animations in `.anim` files listed by FileDataID.
  Weapons and helms hang on M2 attachment points: hands 1 and 2, spell palms 21 and 22, back sheaths 26 to 28, hip
  sheaths 32 and 33. Weapons are sheathed out of combat (where, from the item's sheathe type) and drawn in combat.
- **Animations by AnimationData id:** stand 0, run 5, one hand attack 17, two hand attack 18, ready 26, precast 31,
  cast 32, jump start 37, jump 38, jump end 39, fall 40, channel 124 (directed) and 125 (omni). Attacking while
  running blends the upper body over the run. Some spells name their own animation through
  their visual kit.
- **Placeholders:** until the model renderer works, units are capsules with their race's real height and combat
  reach, so camera and range checks hold before real models arrive.
- **Loading:** `instantiateStreaming`, so Chrome caches the compiled wasm (Go's file server sends `.wasm` as
  `application/wasm`). The worker is a Vite module worker (`new Worker(new URL(...), {type: 'module'})`) that imports
  `wasm_exec.js` from GOROOT's `lib/wasm`, not the CommonJS worker build in `vite.build-workers.ts`.

## Playing like WoW

The room is only as accurate as your play in it. If the camera, movement or casting feel off, you play differently
than in game, and the result means less. So every rule here copies the game, and every number is measured against
the Forever client, not guessed. A value marked "check" is our best knowledge until it is measured.

Sources: the Forever client's CVar defaults (Ketho's BlizzardInterfaceResources, `forever` branch, build 70532), the
Forever UI source (Gethe's wow-ui-source, `forever` branch, build 70707) and the client's GlobalStrings table (build
70685). Emulator code is only a hint for formulas, never the source.

All of it lives in `sim/practice` and `ui/practice`. None of it changes the sim core.

### Movement

| Rule | Value | Status |
| --- | --- | --- |
| Run | 7 yd/s | right (warcraft.wiki.gg) |
| Walk | 2.5 yd/s | right |
| Backpedal | 4.5 yd/s | right |
| Strafe | 7 yd/s | right (the protocol has no strafe speed) |
| Forward plus strafe | 7 yd/s, not faster | check |
| Backpedal plus strafe | 4.5 yd/s | check |
| Turn with keys | 180°/s | right (engine: π rad/s) |
| Turn with right drag | at once, to where the camera looks | right |
| Turn keys while the right button is held | strafe instead of turning | right |
| Jump | about 1.64 yd high, 0.83s in the air, direction locked, you can still turn | right (engine jump speed and gravity) |

- The user turns with A and D and strafes by holding the right mouse button, so the right button rule is the one
  their movement depends on. Left drag doesn't make the turn keys strafe.
- Both mouse buttons held runs forward. Autorun (Num Lock by default) runs until you press forward, back or autorun
  again. Strafing and turning don't stop it.
- Click to move is off (`autoInteract` 0).
- Units don't collide with each other. You walk through the dummy and other players, as in game. Unit cylinders are
  only for clicks and range.
- Moving or jumping stops a hard cast or channel, and the cast bar shows "Interrupted". Turning in place doesn't.
  Instants work while moving and in the air.

### Range and facing

- **Melee range** is the larger of 5 yards and the two units' combat reach plus 4/3 yards. When both units move (not
  walking), it grows by 8/3 yards, which is why you can keep hitting a running target. Blizzard confirms this extra
  reach exists but gives no numbers. Tauren have a larger reach. Status: check.
- **Spell range** starts as the engine model: max range plus your combat reach plus the target's. The moving leeway
  is added only for melee range spells. Status: check, measured at max range on the dummy.
- **At a cast's end** the game checks range and facing again, with about 10% of max range (at most 3 yards) of extra
  room. Turning away during Lightning Bolt fails it with "Target needs to be in front of you." Status: check.
- **Facing:** spells and abilities need the target in your front 180°, auto attack in your front 120° (right, engine).
  Backstab style abilities need you outside the target's front 180°. Parry and block only happen from the front,
  which the sim already handles through `InFrontOfTarget`.
- **Out of range:** auto attack stays on. When a swing is due but you're out of range or facing away, the game tries
  again every 100ms, so the swing lands about as soon as you're back, with no timer reset. The error is sent only when
  it changes, not on each try. The built-in swing timer dims while you're out of range.
- **Height:** the melee check counts the height gap, so a jump at the edge of melee range can lose a swing near the
  top of the jump. Status: check.

### Casting and attacking

- **The spell queue** is described in [Inputs, ping and the spell queue](#inputs-ping-and-the-spell-queue). "Last
  press wins" comes from players, not Blizzard. Status: check.
- **Key down or key up** follows `ActionButtonUseKeyDown` (default 1, key down).
- **Press and hold casting** follows `ActionButtonUseKeyHeldSpell` (default 0). When on, holding a key keeps casting,
  keyboard only.
- **Auto attack** starts with a melee ability, `/startattack`, or a right click on the target. It moves to a new
  hostile target when you switch, unless `stopAutoAttackOnTargetChange` is 1 (default 0). Status: check once.
- **Casts and swings:** in Classic you don't swing during a cast, and finishing a cast resets the swing timer. Step 1
  confirms the sim does this, and the HUD shows it.
- **Swing timer:** Forever has a built-in one (`showSwingTimer`, turned on in Edit Mode). The room copies it.
- **Modifiers** for macros are read at the moment of the press.

### Errors

Errors show in red at the top center. The texts come from GlobalStrings, and the rules from Forever's UI code
(`UIErrorsFrame.lua` with its Forever overrides):

- A line stays 2s and fades over 0.5s.
- Some errors only flash the line already on screen for 0.2s, with no new line and no sound: "Not enough mana", the
  client's "Out of range.", "Item is not ready yet." and a few more. Other errors add a line and play the sound on
  every press.
- Many errors play your race and gender's voice line (`Sound_EnableErrorSpeech` 1), not only a sound.
- Forever hides "Another action is in progress" and shows "Spell is not ready yet." and "Ability is not ready yet.",
  which retail hides.
- EllesmereUI might replace the error frame. That's a check.

| When | Text |
| --- | --- |
| Cast time spell while moving | "Can't do that while moving" |
| Spell or ability, target behind you | "Target needs to be in front of you." |
| Auto attack, target behind you | "You are facing the wrong way!" |
| Backstab style ability from the front | "You must be behind your target." |
| Spell or ability out of range | "Out of range." (the client's own check) |
| Out of range when the server refuses (range changed in flight) | "Out of range" |
| Auto attack out of range | "You are too far away!" |
| GCD or cooldown | "Spell is not ready yet." or "Ability is not ready yet." (check which) |
| Cost | "Not enough mana", "Not enough energy", "Not enough rage" |
| Harmful spell with no target | "You have no target." |
| Wrong kind of target | "Invalid target" or "Target is friendly" |
| `/startattack` with no target | "There is nothing to attack." |
| Dead target | "Your target is dead" |
| `/use 13` on cooldown | "Item is not ready yet." |
| A spell in Ghost Wolf with `autoUnshift` 0 | "You are in shapeshift form" |
| Ghost Wolf indoors | "Can only use outside" |
| Line of sight (Milestone 2) | "Target not in line of sight" |

### Targeting

- Left click picks a unit by its model's bounding hull from the M2 file, and by its cylinder until models load
  (`interactOnLeftClick` 1). Right click targets and attacks.
- Tab picks the nearest enemy in front of you, then cycles (`TargetNearestUseNew` 1). When your target is in combat,
  Tab cycles only enemies in combat (`TargetPriorityCombatLock` 1).
- Esc first closes menus, then stops your cast or channel, and only clears your target when there's nothing else to
  do. So Esc during Lightning Bolt cancels the bolt and keeps the target. Clicking the ground keeps the target
  (`deselectOnClick` 0, "Sticky Targeting").
- `@mouseover` is the unit under the cursor. A right press and release with no movement is a click, and the unit is
  picked at mousedown.
- No soft targeting: it's for gamepads only.
- `autoSelfCast` is on by default.
- Enemy nameplates are off by default (`nameplateShowEnemies` 0, V toggles them), and show up to 45 yards.
- `TargetEnemyAttacker` 1 targets whatever hits you when you have no target. It matters from Milestone 2.

### Camera

- **Left drag** orbits the camera around you without turning you. **Right drag** turns you to the camera's direction,
  and pitches the camera.
- **Dragging:** the room locks the pointer on mousedown and unlocks it on mouseup, so the cursor comes back where it
  started. The lock starts a moment after mousedown, so plain mouse moves are used until then. How Chrome for Windows
  handles fast repeated drags is a check.
- **Mouse look speed** follows `cameraYawMoveSpeed` (default 180) and `cameraPitchMoveSpeed` (90).
- **Mouse input** follows `rawMouseEnable` (default 0, Windows mouse acceleration on) and
  `rawMouseAccelerationEnable`. Raw input (`unadjustedMovement`, Chrome only) is used only when `rawMouseEnable` is 1.
  `mouseSpeed` applies only when `enableMouseSpeed` is 1 (default 0).
- **Following:** when you move, the camera swings back behind you the way `cameraSmoothStyle` says: 0 never, 1 only
  horizontal when moving, 4 only when moving (the default), 2 always. The swing speed is `cameraYawSmoothSpeed` (180)
  and `cameraPitchSmoothSpeed` (45).
- **Zoom:** the max distance is 15 yards times `cameraDistanceMaxZoomFactor` (default 1.0, the settings slider goes
  to 2.0, the console to 4.0 with a 50 yard limit). The wheel steps at `cameraZoomSpeed` (20). Start distance 5.55
  and pitch 10, as the game saves them.
- **Field of view** is `cameraFov` (default 90). Whether that's the horizontal or vertical angle is a check.
- Also copied: `cameraPivot` 1, `cameraTerrainTilt` 0, `cameraBobbing` 0, `CameraKeepCharacterCentered` 1. The camera
  pulls in when the ground is in the way.

### Feedback

- **Sounds:** weapon swings and impacts, cast start and finish, proc sounds, the error sound and voice lines, and
  your ForeverAuras sounds. Those are addon files under `_forever_/Interface/AddOns` (BayovCore, FojjiCore,
  ForeverAuras), not game files, so the practice server serves them read only, with a fallback when a file is missing.
  Many players react to sounds more than to icons.
- **Combat text** in your MSBT style, and your EllesmereUI frames and ForeverAuras as in game.
- **Animations** match the sim: the attack animation lands on the sim's swing time, and the cast animation follows
  the cast bar.
- **Audio** starts on the first click or key press, since browsers block it before one. Sound files are decoded ahead
  of time.

### Input and frames

- Key and mouse events carry their own timestamps, so a press counts at the moment you pressed, not at the next frame.
- Turning uses `pointerrawupdate` or coalesced pointer events, since plain mouse moves arrive once per frame.
- The screen runs at your monitor's rate (60 to 144Hz or more), with movement on a fixed step and drawn smoothly
  between steps.

### What a browser can't match

- **Display lag:** Chrome always waits for vsync, and its compositor adds about 1 to 2 frames (7 to 14ms at 144Hz)
  over WoW with vsync off. This changes your reaction time only, never the sim's results. We try the canvas
  `desynchronized` hint, and we measure press to screen once the scene exists (step 5).
- **Browser support:** Keyboard Lock is Chrome and Edge only, so the room needs one of them.
- **Keys:** Keyboard Lock works only in fullscreen started by the page (a click, not F11). The room locks only your
  bound keys plus Esc. Locking every key would let Chrome's keyboard hook swallow Alt+Tab and the Win key, which traps
  you away from WoW and Discord. Whether locking Tab also catches Alt+Tab is a check. Ctrl+Alt+Del and Win+L are never
  caught.
- **Mouse:** buttons 4 and 5 go back and forward in history, so the room cancels that on mouseup and keeps a history
  guard. Browsers don't see mouse buttons 6 and up. With Num Lock on, Windows reports Shift+numpad without Shift.
- **Monitors and focus:** a visible cursor can leave the fullscreen tab for another monitor, where WoW can keep it
  inside its window. Losing window focus drops both locks.

### Checking the feel

1. **Measured numbers.** Each "check" above gets measured in game and in the room, the same way:
   - movement, turning and jumps from a short BayovCore probe that logs your position and facing every frame
     (`UnitPosition` works outside instances) while you run a set path
   - cast bar and GCD start timing from the same probe: it logs `UNIT_SPELLCAST_SENT`, `UNIT_SPELLCAST_START` and
     `SPELL_UPDATE_COOLDOWN` times next to `GetNetStats`, which settles whether the client predicts them
   - which error texts fire, from `UI_ERROR_MESSAGE`
   - turning away during a cast, a spell press in Ghost Wolf, and a press while a spell comes off cooldown
   - cast, GCD and swing timing from the combat log
   - camera zoom, field of view and turn speed from screen recordings of the same moves
2. **Side by side.** The same short fight on the in-game dummy and in the room, with the logs compared.
3. **A feel checklist** (`docs/practice_room_feel.md`) you play through after each change to movement, camera or
   casting, one line per rule above.
4. **Spell batching:** whether Forever's server groups spell results into time batches. One log check.

## Assets (local only)

The Forever client's data sits in the shared CASC storage (WoW's archive format) of the WoW install. Its product code
in `.build.info` is `wowf`, version 1.60.1.70707 today. Two folders matter:

- The data root (`/mnt/d/Gaming/World of Warcraft` on this PC), with `.build.info` and `Data/`. It's the one setting.
- The Forever folder inside it (`_forever_`), with `WTF`, `Interface/AddOns`, `Cache/ADB` and `Fonts`. The server
  finds it by its `.flavor.info`, which names `wowf`.

**The pipeline:**

1. The practice server extracts raw files by FileDataID (the client's file number): M2 models, `.skin`, `.skel` and
   `.anim` files, BLP textures and OGG sounds. It runs TACTTool from TACTSharp's prebuilt release, a self-contained
   single file that needs no .NET runtime. The Windows build, run through WSL with Windows paths (`wslpath -w`), reads
   `D:` directly instead of through the slow WSL mount of the 139 GB `Data` folder. CascLib with a cgo binding is the
   fallback. There's no usable Go CASC library.
   - TACTTool reloads the build's tables on every run, so the server extracts in batches (`-m list`). It plans each
     character's file list from the tables, and the asset route serves only cached files and queues a batch on a
     miss. Files are cached in `~/forever-practice/assets/`.
   - TACTTool reads decryption keys only from `WoW.txt` in its working folder (wowdev's TACTKeys list). A missing key
     gives a zero-filled file with exit code 0. So after each extraction the server checks the file's magic (`MD21`,
     `SKL1`, `BLP2`, `OggS`) and that it isn't all zeros.
2. Game tables come from wago.tools as CSV, for the exact build in `.build.info`, cached in
   `~/forever-practice/db2/<build>/`. The repo already pulls wago tables by build. Reading them from the local client
   needs a DB2 reader, table layouts for the build and a merge of the hotfix cache (`DBCache.bin`). That's a project of
   its own, so it waits for Milestone 4 and spike 6.
3. The browser runs our own Three.js model renderer. We port wow.export's file readers (M2, SKEL, ANIM and BLP) and its
   character assembly (MIT), and build Three.js meshes from them. wow.export 0.2.x has its own WebGL renderer, not
   Three.js, so the port is about 10 thousand lines. three-m2loader (MIT) can't read `.skel` files or `.anim` files by
   FileDataID, so it's only a reference for building tracks and materials.

wow.export added Forever support in 0.2.20 (2026-11-02) through the new product code, so the file formats are Classic
Era's. It has no command line or remote control, so it can't drive the pipeline, and its glTF export leaves weapons
unattached and drops particles. We use its GUI to check our output.

**A dressed character** is the biggest asset job in Milestone 1. Spike 2 sizes it.

1. ChrModel, then CreatureDisplayInfo and CreatureModelData, give the body model and its skeleton. Customization
   choices give skin textures and geosets (the body parts shown or hidden, like hair and ears).
2. Each item leads through ItemModifiedAppearance and ItemAppearance to ItemDisplayInfo. Its geoset groups switch
   body parts (gloves, boots, robe, cape, belt), and HelmetGeosetData hides hair under helms.
3. ItemDisplayInfoMaterialRes and ComponentTextureFileData give one texture per body section (arm, hand, torso, leg,
   foot), painted onto the body texture.
4. ComponentModelFileData gives the models that hang on attachment points.

**Tables we need:**

- Characters: ItemModifiedAppearance, ItemAppearance, ItemDisplayInfo, ItemDisplayInfoMaterialRes, ModelFileData,
  TextureFileData, ComponentModelFileData, ComponentTextureFileData, HelmetGeosetData, ChrModel, ChrRaceXChrModel,
  CreatureDisplayInfo, CreatureModelData, ChrModelTextureLayer, ChrModelMaterial, ChrCustomization tables (Option,
  Choice, Element, Geoset, Material), CharComponentTextureLayouts and Sections, ItemSparse (sheathe type).
- Spells for the spellbook and the spells the sim doesn't have: SpellName, SkillLineAbility, SpellLevels, SpellMisc
  (icons, cast time and range ids), SpellCastTimes, SpellCooldowns, SpellPower, SpellRange, SpellDuration. The page
  passes the rows it needs to the session at start.
- Animations, visuals and sound: AnimationData, SpellXSpellVisual, SpellVisual, SpellVisualEvent, SpellVisualKit,
  SpellVisualKitEffect, SpellVisualKitModelAttach, SpellVisualMissile, SpellChainEffects, SoundKit, SoundKitEntry,
  WeaponSwingSounds2, WeaponImpactSounds.

**Spell visuals and imbue glows** are mostly M2 particles and ribbons. Milestone 1 has animations, icons, sounds and
combat text, and no particles. Later the same model renderer gets particles, with WebWowViewerCpp's emitters as the
reference (read only, it has no license file).

**Keeping it local:** the assets live in `~/forever-practice/assets/`, outside the repo. Only the practice server
serves them. The deploy guard in `tools/dps_report/build_artifact.mjs` fails a DPS Report build that touches practice
code or paths, and `dist/forever/` never holds practice files. The deploy only ships `dist/forever`.

## Keybinds and action bars

### Bindings

`bindings-cache.wtf` holds only the keys changed from the defaults. The user's file has 150 lines and no
`MOVEFORWARD`, `MOVEBACKWARD`, strafe, jump or tab target lines. Unbound keys show as `bind 7 NONE`. The defaults live
in the client binary, not in FrameXML.

So the room starts from a table of Blizzard's default bindings (W and S move, Q and E strafe, Space jumps, Tab
targets, 1 to = are `ACTIONBUTTON1-12`, and so on) in `cmd/practice/defaults/bindings.txt`, in the same `bind KEY
COMMAND` format. It's written by hand and checked once against a fresh character's key bindings menu. Then each line
of `bindings-cache.wtf` replaces the default for its key, and `NONE` unbinds it.

Notes from the user's files:

- There is one file per account. A character file overrides it, but none of the user's characters has one.
- The user has two accounts (`405385253#1` and `#3`), each with its own bindings, macros, CVars and BayovCore file.
  The server reads records from every account, keeps the newest copy of each GUID (by `seen`), and takes the
  bindings, macros and CVars from that record's account. A character's own `macros-cache.txt` and
  `config-cache.wtf` win over the account's.
- The user turns with A and D (`TURNLEFT`, `TURNRIGHT`) and strafes by holding the right mouse button. Q and E are
  action buttons.
- WoW falls back to a key's unmodified binding when the modified key is unbound. Alt+F1 isn't bound, so it runs F1's
  binding with Alt held, and the user's `[mod:alt]` macros depend on that. Key lookup does the same.
- Click bindings and situational override bindings (BayovCore fishing, EllesmereUI party mode) are ignored.

### Buttons to slots

Bindings name a command, and each command maps to a slot. Blizzard's table, checked against the user's record:

| Command | Slots |
| --- | --- |
| `ACTIONBUTTON1-12` | 1-12 (paged by stance or form, see below) |
| `MULTIACTIONBAR3BUTTON1-12` | 25-36 |
| `MULTIACTIONBAR4BUTTON1-12` | 37-48 |
| `MULTIACTIONBAR2BUTTON1-12` | 49-60 |
| `MULTIACTIONBAR1BUTTON1-12` | 61-72 |
| `MULTIACTIONBAR5BUTTON1-12` to `MULTIACTIONBAR7BUTTON1-12` | 145-180 |
| `EUI_BAR9_BUTTON1-12` (EllesmereUI) | 13-24 |
| `EUI_BAR10_BUTTON1-12` (EllesmereUI) | 109-120 |
| `EUI_BAR11` to `EUI_BAR13` (EllesmereUI, BayovCore patch) | 73-108 |

EllesmereUI keeps Blizzard's slots and binds through the same commands, so its keys are in the bindings too. Its own
bar numbers differ from Blizzard's (its "Bar2" is `MULTIACTIONBAR1`), so the doc names commands, never "bar N".

On Forever, every class's stance or form pages bar 1 to slots 73-120. Ghost Wolf is a form, and whether it pages bar
1 on Forever is a check (retail's doesn't). Until then the room keeps page 1 in Ghost Wolf.

### What's in each slot

The server keeps what sits in each slot, and no WTF file has it. BayovCore's CharSetup module records it in
`BayovCoreDB.modules.charsetup.chars[<GUID>]` with `{name, realm, class, race, level, build, slots, macros}`:

- A spell slot is `{kind="spell", id=<spell id of the rank>, name, top}`. The spell id maps straight to a sim action.
- A macro slot has the whole body, so `macros-cache.txt` is only a fallback. Account macros are at
  `charsetup.account.macros`.
- A slot the game hid is `false`.

The file is a plain Lua table, written at logout or `/reload`. A small hand parser reads the 850 KB file in 0.1s.

Today only one record exists: Bayov Oom, a level 9 shaman on the beta realm. A character has a record only after a
login with BayovCore on. So:

- The character picker lists records, keyed by GUID, not WTF folders. The WTF folder is named
  `<realm id>/<Name-Name>` and can only be joined to a record by name. A "No record" entry uses the defaults.
- With no record, the room fills bar 1 and the bars after it with the spec's main spells, taken from the class sim's
  rotation in the settings. The rotation only picks the spells, it never casts them.
- Every class spell at your level is castable (see [Your toolkit](#your-toolkit)). Only ranks the sim drops show
  greyed, and macros with unsupported commands get a badge.
- A bag item slot is `{kind="item", id=<item id>}` and runs `useItem`.
- Buttons can also be clicked.

### Editing bars and binds in the room

- **Bars:** pick up a spell, item or macro from a spellbook or a slot, and it sits on the cursor as in WoW. Click a
  slot to place it, or drop it off the bars to clear. This uses pointer events, not the browser's drag and drop.
- **Binds:** a quick keybind mode, as in WoW: hover a button and press the key or mouse button, with modifiers.
- Edits are saved per character in `~/forever-practice/settings/<GUID or "default">.json`, on top of what the client
  files and the record say: `{ "version": 1, "slots": { "<slot>": { "kind": "spell", "id": 0 } }, "binds": { "<KEY>":
  "<COMMAND>" }, "fight": { ... } }`. A slot's kind is `spell`, `macro` or `item`. They never write back to the WoW
  folder. "Reset" puts the client's layout back.

### Macros

What the user's shaman macros use, which Milestone 1 supports:

- `/cast [@mouseover,harm,nodead][] Stormstrike`: every damage spell is a macro like this.
- Conditions: `@mouseover`, `@target`, `@player`, `harm`, `help`, `dead`, `nodead`, `mod:alt` (and shift, ctrl), and
  the empty `[]` fallback.
- `/startattack`, `/cleartarget [dead][help]`.
- Ranks: `Frost Shock(Rank 1)`, any case.
- `/use 13` and `/use 14` (trinket slots, `useSlot`), and `/use <item>` lines (`useItem`), where the first usable
  one fires.
- `#showtooltip` sets the button icon.

A macro with anything else (`/castsequence`, `/tm`, `/run`, `/click`, `/focus`, `/target`, `/bayov`) gets an
"unsupported" badge. `macros-cache.txt` blocks are `VER 3 <id> "<name>" "<icon>"`, the body, then `END`.

### Settings (CVars)

CVars live in three files: `WTF/Config.wtf`, the account's `config-cache.wtf` and the character's. They are written
only when changed, so a missing CVar means the default from the Forever CVar list. The room reads every CVar named in
[Playing like WoW](#playing-like-wow), plus `SpellQueueWindow` (the user hasn't changed it, so 400ms).

## The practice server

`cmd/practice/` (built with `-tags with_db,practice`) serves the built page from `dist/practice/` and these routes:

| Route | What |
| --- | --- |
| `POST /api/request` | a class sim link or JSON in, a `RaidSimRequest` with its item database out |
| `GET /api/records` | BayovCore records (GUID, name, realm, class, level), plus "No record" |
| `GET /api/character/<guid>` | bindings (defaults + WTF + saved edits), slots, macros, CVars |
| `PUT /api/settings/<guid>` | save bar, bind and fight setting edits |
| `GET /api/assets/<fileDataId>` | one extracted file, or a queued batch on a miss |
| `GET /api/db2/<table>` | one game table as JSON, from the cached wago CSV |
| `GET /api/addons/<path>` | one sound file from `_forever_/Interface/AddOns`, read only |
| `GET`, `POST /api/attempts` | list and save attempts, `GET /api/attempts/<id>` for one |
| `POST /api/replay` | run an attempt again natively, its summary and log |
| `GET /api/presets`, `/api/presets/<name>` | the DPS Report's builds, a quick way in |
| `GET /api/spells?ids=` | spell names and ranks, from the cached wago tables |
| `GET /api/icon/<spell or item>/<id>` | an icon name from Wowhead, cached |

It also serves `/forever/assets/` from the repo's `assets/`, because `ui/core` code fetches files like
`/forever/assets/database/db.json` (`ui/core/proto_utils/database.ts:23`).

For development, the practice Vite dev server runs on its own port with its own config (`root: ui/practice`), and
proxies `/api` to `cmd/practice`. It doesn't reuse `getBaseConfig` from `vite.config.mts`, which brings
`base: '/forever/'` and `outDir: dist/forever`. It copies the one `jsxInject` line and has its own small middleware
for `/forever/assets`.

## Fight summary

The room sims only the fight you played: one iteration, no other seeds and no averages. Everything counts from the
pull to the end (Stop, the set length in Encounter mode, or the 60 minute cap). At the end it shows:

- your DPS and damage by spell
- idle GCD time
- buff and debuff uptimes
- time out of range
- swings pushed back by casts
- time with resources capped

The session works these out from the event stream, cut to the time from the pull to the end. The core's own metrics
count some of the prepull: an aura that fades before 0 adds negative uptime (`aura.go:788`), and casts and mana spent
before 0 are counted (`spell.go:616`). Damage is counted from 0 on in both.

The whole fight opens in the Replay view.

## Where it lives

Its own page, served by the practice server. You bring a class sim's settings (character, gear, talents, buffs,
rotation, fight length) through the class sim's existing export, as a link or JSON. Presets and pasted BayovCore
exports come in the same way, since they already load into the class sim. The keybinds come from a BayovCore record
you pick separately, because the sim character and the in-game character aren't the same thing.

Fight rules for Milestone 1 are the two modes in [Fight modes and the pull](#fight-modes-and-the-pull). Before the
pull you're free to prepare for as long as you like. The HUD has the Start Encounter button with its countdown and
length settings, the Stop button, the countdown, the fight timer, and the ping and "Execute range" settings. Stop,
Restart and Leave can be bound to keys. Restart starts a new attempt with a new seed and the same settings.

The fight settings are saved per character with the bar edits, in `settings/<GUID>.json`:
`"fight": { "mode", "lengthMs", "countdownMs", "pingMs", "executeFromStart" }`.

## Multiplayer (Milestone 4)

Milestone 4 is more than a new connection. It needs:

- a Go server that runs the session natively, applies each input when it arrives, and sends events right away or in
  small batches (about 50ms)
- a fixed tick (20 to 30 a second) only for position broadcasts, which clients interpolate
- several units with player control
- a lobby, because the raid is fixed when the sim is built, so nobody joins mid-fight
- a rule for disconnects: an APL bot takes the unit over
- no pausing, since one player's hidden tab can't stop the server's clock
- each player's latency
- filtering events per player, because a 40 player fight makes thousands of aura, resource and tick events a second
- assets for everyone else's race and gear on each friend's PC
- game tables read from each friend's own client, so everyone matches their build (spike 6)

Lockstep (every client runs the sim and waits for everyone's inputs) doesn't fit 40 players: every input waits for the
slowest player, and joining mid-fight is hard. So the server decides everything, as WoW does.

**Friends' setup.** The host serves the client over a tunnel such as Tailscale, so friends need only Chrome. Each
friend's browser reads their own WoW install's assets through a small local helper or the File System Access API, so
we never send game files anywhere. The details wait for Milestone 4.

## Status (2026-11-02)

The room plays: the enh shaman against the dummy, with capsules for units. `make practice-run` builds it and serves
it at http://127.0.0.1:3434/. Real WoW models are not in yet (step 0 and steps 6 and 8, see below).

**Done and tested:**

- Step 1, the session (`sim/practice`). Go tests cover the doc's whole list: the queue, casting, auto attack, every
  pull trigger, the swap keeping state, the prepull window (a 70 minute prepull too), both modes, imbue swaps under
  Windfury Totem, Rockbiter's attack power, Ghost Wolf, Earthbind Totem, potions and sappers.
- The rotation check: the enh APL's own casts as inputs give 934.8 DPS against the APL's 933.6, over 100 seeds.
- Every DPS Report build casts its whole toolkit. The enh builds never fail. Other specs' toolkits hold spells the
  sim models only for another spec (a balance druid's Cat Form), and those end the attempt with an error instead of
  crashing the worker.
- Step 2, replays: a saved attempt replays to the live fight's log byte for byte, in both modes and after prepull
  window swaps. Opening it in the Replay view is not wired yet.
- Step 3, `practice.wasm` in its own worker. `go list -deps ./sim/wasm` shows no practice package.
- Step 4, the server, with golden file tests for the Lua reader, the WTF bindings and CVars, and the records.
- Steps 5 and 7: the scene, WoW movement and camera, targeting, bars from the record (or from the rotation when
  there's none), macros, bar and key editing, unit frames, cast bar, swing timer, auras, errors, combat text, the
  meter, the fight settings and the summary. `tools/practice/smoke.py` plays a scripted fight in headless Chrome:
  no console errors, the meter matches the summary, an imbue swap, and an Encounter mode fight that ends on its own.
- The HUD copies the user's game UI (2026-11-02). The server reads the account's EllesmereUI profile
  (`cmd/practice/ui.go`): the bar blocks, unit frames, cast bar, damage meter, spell history and tooltip spot sit
  where the profile has them, scaled by its UI scale, in its font. The Bayov class pack (maintain cells, resource
  bars, main and utility cooldown bars) follows `bayovui/auras/common/anchors/group.py`. Tooltips are Wowhead's
  Forever ones, cached in `~/forever-practice/icons.json`. Toolkit spells the bars don't hold go on the key that
  `notes/<class>-keybinds-forever.md` in the WoW folder plans for them, when that key's slot is empty.
- The core hooks and the aura hold fix give identical results on all 73 DPS Report builds.

**Spike 4 (WASM in real time):**

| What | Native | WASM (Node) | Chrome |
| --- | --- | --- | --- |
| Session start (`NewSim`, `Reset`, `PrePull`) | 2ms | 37ms | 11ms |
| Worst-case pull swap: 9 minutes of nonstop strafing and turning, 5454 inputs | 3ms | 10ms | |
| A normal pull swap | | | 3ms |
| One step of a fight, p99 | 0.04ms | 0.5ms | |
| WASM load | | | 0.2s, 38.6 MB |

Two changes got the swap there. The next swap's sim is built in idle time (`PrepareSpare`). Before the pull, with
auto attack off, the world skips its 50ms ticks, since range and facing then change nothing the sim sees. The worst
case still takes about 1.5 frames at 144Hz. Headless Chrome's press to answer time (19ms median) is its software
rendering, not the sim.

**Changed from the plan above:**

- The HUD reads typed state and hit events from each batch (`PlayerState`, `HitEvent`), not parsed log lines. The log
  lines still come with each batch.
- Icons and item names come from Wowhead's tooltip service through the server (`/api/icon`), cached in
  `~/forever-practice/icons.json`, until the asset pipeline gives the client's own.
- The spells the sim doesn't have (kind 3) are a small table of top rank values in `sim/practice/utility.go`, not
  rows from the client's tables yet.
- Spell names and ranks for macros come from the wago tables already cached in `~/forever-logs/wago`.

**Real models (steps 0, 6 and 8, 2026-11-02):**

- The server extracts client files with TACTTool's Linux build (`-tact`, in `~/forever-practice/tools/tact-linux`)
  into `~/forever-practice/assets/<client build>/`. TACTTool fails on links inside the WoW folder, so it reads a folder of its own
  (`~/forever-practice/wow`) with the install's `.build.info` and a link to `Data`. A run takes about 2s. Files asked
  for within 40ms share a run, and a model's skin, textures and skeleton are queued as soon as it comes out.
- Game tables come from wago.tools by the client's build (`cmd/practice/db2.go`).
- `appearance.go` turns race, sex, class, gear and customization choices into a list of files: body model, geosets,
  the body texture's layers, attached models (helm, shoulders, weapons with their sheath points) and cape texture.
  It follows wow.export's character code. The preset's made up Synthetic items show as real look-alikes (`look.go`).
- `visuals.go` reads spell visual kits (sounds, effect models on attachment points, the cast animations, missiles),
  imbue glows, melee swing and impact sounds, unit sounds and the totem models.
- The page (`ui/practice/wow/`) reads M2, skin, .anim and BLP files, paints the body texture, builds Three.js
  skinned meshes, plays animations with fades and upper body swings while running, runs M2 particle emitters and
  ribbons, and plays the sounds. `viewer.html?model=<fdid>` or `?char=race=2&sex=0&class=7&items=...`
  shows one model for checks.
- Checked in headless Chrome: the dressed Orc stands, runs, swings, casts, sheathes, and Lightning Bolt,
  Lightning Shield, the imbue glows and the totems show their effects. The dummy is the wooden dummy (display 3019).

**Audit and fixes (2026-11-02, second pass):** four audits (animation, effects, sound, rendering), then:

- The room loads every model, texture, animation, effect and sound behind a loading bar before the session starts.
- A spell with a cast time goes off at the cast's end (the sim reports casts at their start).
- Movement turns the body like the client: 90 degrees on a strafe, 45 on a diagonal, the opposite diagonal while
  backing up. Turning in place shuffles. The backpedal plays at the walk's pace. The landing run plays out.
- Swings: the impact sound and the dummy's flinch wait for the swing's `$HIT` event, the whoosh for `$CSS`, and
  footsteps follow `$FSD` (M2 events). Auto attack draws the weapon with its animation and sound. Swing variations
  play at random, crits flinch harder, the upper body fades in and out, and a switch mid fade doesn't pop.
- Sound: plain stereo panning (the browser's HRTF mode muffled the files), the listener at your character, your
  own sounds flat, a limiter, the kits' pitch and volume variation, cooldowns and fades, looping precast and
  missile sounds, Orc attack grunts on crits and specials (a guess: no table says when the client plays them),
  weapon draw and sheathe, and the Orc's spoken errors.
- Effects: birth, hold and decay animations instead of fixed lifetimes, impacts of spells with no damage (shields,
  imbues, heals), the client's proc spells (Windfury, Flametongue, Frostbrand: `HitEvent.visualSpellId`), the
  off hand imbue glow, the Searing Totem's bolts (precast, missile, impact, `HitEvent.nextAt`), Magma pulses and
  Fire Nova's ring on the fire totem, totem birth and death, the totems' buffs on you, and the Ghost Wolf model.
  An impact on a spell's area plays at the totem, not the dummy, and Totemic Recall takes the models away.
- Strafing turns the legs where you run and twists the upper body back to where you face (the lower spine
  turns, `Model.twist`). Holding both mouse buttons runs forward and still steers.
- Buffs from the fight settings that never run out sit on their own rows, the fight's buffs under them.
- Graphics settings with Low, Medium and High presets: render scale and shadows (off, or 1024 to 4096 pixel maps
  that follow you). Glows and see-through layers cast no shadow. Particles on a texture with no see-through part
  fade to a round edge instead of drawing squares, and wait for their texture.
- The game's cursors from the client (`Interface\Cursor\Point.blp` and `Attack.blp`).
- A performance recorder (`ui/practice/perf.ts`) times each part of every frame, notes models built and shaders
  compiled mid fight, long tasks, Chrome's long animation frames and garbage collections, and saves a report to
  `~/forever-practice/perf/` when a fight ends.
- First report (206s at 240 FPS, 75 frames over 20ms) and fixes:
  - The first Searing Totem decoded a 2048 by 2048 particle texture (116ms). Totem and effect models now load
    their particle and ribbon textures behind the loading bar.
  - Shaders compiled and textures went to the GPU the first time each effect played (up to 80ms). The room now
    draws every effect and totem model once while loading (`Room.warmUp`). Mid fight compiles went from about 15
    to 1 in a test that casts every spell.
  - Hovering the mouse tested every triangle of the skinned models (25ms per move). Only the dummy's hull and the
    ground are tested now (0.08ms).
- Combat text uses the font's only weight with a little letter spacing. The faked bold ran the digits together.
- Strafing spreads the upper body's twist up the spine (lower spine 55%, chest and neck 25%, head 20%). The head
  faces forward and the chest stays about 30 degrees toward the run. The Orc model has no sideways run of its own
  (no RunLeft or RunRight), so this is procedural, as in the client.
- The room stands on the northern Barrens (Kalimdor tile 36_30, world point 750, -2650), read from the client's
  map files (`ui/practice/wow/adt.ts`, `terrain.ts`): heights, the four blended ground textures of each chunk, vertex
  colors, and the trees, rocks and props around it within 240 yards. The ground is flat out to 45 yards from the
  dummy and blends into the real hills by 110. No doodad stands within 25 yards. Buildings (WMOs) aren't drawn, and
  the spot has none near.
- The Barrens' daytime ambience and wind loop from the start (SoundAmbience 3049: sound kits 292584 and 292700).
- Second report (with the terrain): few hitches, but 120 FPS instead of 240 in the fight, from 4 to 6ms of script a
  frame drawing up to 380 meshes. The ground is now one mesh (a texture array for the ground textures and one atlas
  for the alpha maps), and each doodad model is a plain mesh in its first pose drawn for all its copies at once
  (no bones to update every frame). Draw calls went from 134 to 54 in the same view.
- Third report: 240 FPS at the median, but a quarter of the fight frames at 8ms. The page drew at 3840x2160 (a 4K
  screen at 150%), and script took about 3ms a frame. Fixes:
  - A Resolution setting caps the drawn height (1440p by default) and stretches it to the screen.
  - The HUD forced the browser to lay the page out twice a frame (the cast history's width and the canvas's size
    read after the HUD's changes). Both are read once now, before any change.
  - The HUD remembers its elements and writes a text or style only when it changes (`ui/practice/dom.ts`).
  - Empty particle emitters and ribbons aren't drawn, and only live particles go to the GPU.
  - Bones skip Three's Euler angles, and billboards update only the bones above them.
  - In a loop benchmark the HUD went from 0.66 to about 0.1ms a frame and the character's posing from 0.25 to 0.12ms.
- Fourth report (240s): 240 FPS for 99% of frames, none over 17ms, about 1.3ms of script a frame.
- The mouse locks once, at your first click in the room, and the room draws the WoW cursor itself
  (`ui/practice/pointer.ts`). It sends the page the moves, hovers, presses, clicks and wheel a real cursor would, and
  shows dropdowns as its own lists. Chrome's "press Esc to show your cursor" notice then shows once per lock instead of
  on every camera turn. The full screen notice can't be hidden by a page.
- The wheel zooms over the HUD too, and the camera glides to the new distance.
- Leaving the room stops its frames, timers, sounds and session.
- Raid test: `practice.html?raid=40` puts 39 others in your gear (two tanks, ten melee, casters, healers and hunters
  of the four Horde races) around Ragnaros (`ui/practice/raid.ts`). They swing, cast with hand glows and bolts, heal
  the tanks and move. First numbers in headless Chrome on WSL (software GPU, so only the script and the counts
  carry over): draw calls 55 to 1006, triangles 71k to 424k, script a frame about 0.3 to 6ms (animating 2ms, the
  others' effects 0.6ms, placing the scene's objects 1.1ms), JS heap 0.5 to 1.9GB, and 42s to load.
- First raid report (your PC, 40s): 60 FPS at the median, 16ms of script a frame (drawing 9.5ms for 1410 draw calls,
  animating 6.2ms), 2.3GB heap, and two stalls of 1.3 and 2.1s with no long task in them. Fixes:
  - A unit's parts that look alike (same material, textures and animated colors) draw as one. A raider went from 21
    draw calls to about 11, and the hidden geosets drop out of its index.
  - Units have bounds (their rest pose with room to spare), so the renderer skips the ones off screen and the shadows
    of the ones outside the shadow's area. Spell effects stay always drawn.
  - Raiders animate 10 times a second off screen, 30 past 40 yards, 60 past 20 yards and every frame near you.
  - Copies of a model share its parsed file and its vertices (on the GPU too). Composed skin textures let go of their
    page copy once uploaded.
  - Scratch values for sampling animations are Float64Arrays, and particles and ribbon edges are reused. Chrome stores
    a decimal in a plain array (and in three's vectors) as its own object, so these made about 30KB of garbage per
    character a frame. Some is left (three's vectors, the track cache lookup, about 1.5MB a frame in the raid).
  - In headless Chrome: draw calls 1006 to 657, animating the raid 2.0 to 0.55ms, heap 1.9 to 1.3GB, load 42 to 29s.
  - The two stalls were Alt+Tab.
- Second raid report (your PC, 37s, after the fixes): 68 FPS on average (was 49), 12.6ms frames at the median, 12.2ms
  of script a frame (drawing 7.8ms for 943 draw calls, animating 4.0ms), 1.4GB heap, and no stall. One frame over
  50ms, at the start.
- Aiming for 240 FPS in the raid. Headless Chrome on WSL can use the real GPU (Mesa's D3D12 driver:
  `GALLIUM_DRIVER=d3d12 MESA_D3D12_DEFAULT_ADAPTER_NAME=NVIDIA LD_LIBRARY_PATH=/usr/lib/wsl/lib`, ANGLE on GL), so the
  numbers below are a 2560x1440 raid of 40 on the RTX 5090 with the frame rate uncapped. Its draw calls cost about twice
  what they cost on your PC. The perf reports now also time the GPU (a `gpu` column and average).
  - Before: 40 FPS, 22ms of script a frame (drawing 17ms for 1004 draw calls, animating 3.9ms), GPU 6.9ms.
  - Raiders that look alike draw together (`wow/crowd.ts`): one draw call per part of a look (the orc male's chest in
    this armor) for every raider wearing it. Each raider's bones go into one shared texture, and each copy reads its
    own by an offset. Composed skins, rewrapped textures and compacted indices are shared, so copies match. Copies
    off screen are left out unless they cast a shadow onto the screen. 77 FPS, 680 draw calls, GPU 1.0ms.
  - The casters' hand glows, bolts and impacts draw in the crowd too, and a model with no solid part (most spell
    effects) no longer draws into the shadows. 92 FPS, 490 draw calls.
  - Particles and ribbons draw in one draw call per look (texture and blend), filled before the frame draws. Filled
    during the draw, the GPU got the new triangles over the last frame's points. 111 FPS, 184 draw calls.
  - Bones are posed in flat arrays and are out of the scene (`wow/model.ts`). The scene holds only the sockets
    (attachment points and emitters), which follow their bone. Before, the renderer placed and walked through 9000
    bones twice a frame. 221 FPS, 3.6ms of script a frame (animating 1.7ms, drawing 1.7ms for 180 draw calls), GPU
    1.1ms. Solo is 400 FPS here.
  - Tried and dropped, since they didn't measure: near raiders animating at 120 Hz, a table for the particles' sRGB
    colors, and packed track key slots.
- Third raid report (your PC, 44s): about 155 FPS in the fight (207 in the first 5s). Frames were 4.2ms or 8.4ms, so
  half missed the 240 Hz refresh. 4.2ms of script (animating 2.3ms, drawing 1.7ms), GPU 4.6ms, 168 draw calls at the
  start and 260 in the fight, a garbage collection every 90ms (about 110MB/s) and 2043 effect models made.
  - The reports now time the GPU by pass: the shadows (with the frame's uploads before them), the solid things, and
    the see-through things with the effects. They also name the GPU and say the canvas size and MSAA. A timer started
    inside the shadow pass lost the whole WebGL context in headless Chrome, so the shadows and uploads stay one
    number. In headless runs the GPU time follows the script's drawing time, so only your PC's numbers say where the
    GPU's time goes.
  - Spell effects are reused: a done effect waits for the next one of the same model, instead of a new model with new
    materials each cast. 500 effect models made in 10s of the fight became 15.
  - Less garbage: the animation's times and the bones' moves no longer go through function arguments (Chrome boxed
    each decimal), Track.get is small enough to inline, the crowd uploads its copies whole, and the effect batches copy
    with plain loops. A garbage collection every 90ms became one every 450ms.
  - The fight's extra draw calls are your totems and effects. In the raid test they draw in the crowd too.
- The raiders wear their class's raid gear: the dungeon set, tier 1 or tier 2, with weapons for their role (tanks a
  sword and shield, rogues two daggers or swords, hunters a bow or gun and a polearm, casters a staff or a main hand and
  off hand, healers a mace and an off hand) and a cloak. Raiders of one class go through the sets, so two mages rarely
  match. Your character keeps your gear.
  - Each raider's skin is painted for its armor, so in a raid of 40 no two match. The skins go into texture arrays
    (by size, those over 1024 wide at half size), and each copy reads its layer. A body of one race and sex draws every
    part any of its raiders shows, and each copy hides the parts its raider doesn't (a mask of its submeshes).
  - Headless, fight, 1920x1080, varied gear: without the arrays and masks 80 FPS, 9.8ms of script, 575 draw calls.
    With them 103 FPS, 8.7ms (animating 3.0ms, drawing 5.4ms), 470 draw calls, 505k triangles, 1.7GB heap. This is the
    new baseline. Most of the rest is gear: each item model (helms are by race) draws on its own, with its shadow.
- Fourth raid report (your PC, 95s, varied gear): about 127 FPS in the fight, frames mostly 8.3ms. 5.8ms of script
  (animating 2.4ms, drawing 3.0ms), 460 draw calls, 506k triangles. GPU 5.75ms: shadows with the uploads before them
  3.1ms, see-through 2.0ms, solid 0.65ms. A garbage collection every 190ms.
  - The uploads: each crowd batch sent five attributes every frame and each particle batch five more, about 1900
    uploads a frame, and on Windows each is a copy of its own on the GPU. Now a batch's copies are one interleaved
    buffer that goes up only when a copy changed (a raider moved, an effect's hand glow followed the hand), and the
    particles and ribbons are interleaved too. A unit standing still settles its turn exactly, so its place stops
    changing. About 210 uploads a frame now, 2MB (the bone texture is 0.9MB of it).
  - The body batches kept the first raider's own skin as their map, only so their shader samples one. Three sent it
    to the GPU anyway (8MB each, and empty once we let go of the picture). They have a one pixel stand-in now.
  - The reports count the uploads a frame (calls and kilobytes), and on Chrome's Direct3D 11 they time the uploads
    apart from the shadows (the timer starts just before the renderer switches to the shadow map).
  - Your call: with 20 or more raiders, helms, shoulders and cloaks cast no shadow.
  - Headless, fight, 1920x1080, varied gear: 128 FPS (was 103), 6.7ms of script (animating 2.7ms, drawing 3.7ms, was
    5.4ms), 450 draw calls, 210 uploads a frame, a garbage collection every 420ms.
- Fifth raid report (your PC, 65s): about 145 FPS for 30s, then 100 to 114 once the fight started (390 to 430 draw
  calls). GPU 4.4ms: uploads 0.01ms, shadows 1.2ms, solid 1.3ms, see-through 1.8ms. So the uploads weren't the GPU's
  time, but the shadows and solid things went from 3.75 to 2.56ms. 5.6ms of script (animating 2.5ms, drawing 2.8ms),
  a garbage collection every 270ms.
  - Shadows: a model's solid parts now cast their shadow in one draw call (one per side), apart from parts cut by their
    texture's alpha. A body of one race and sex casts one for all its raiders. Those batches draw only in the shadows:
    they're on a layer the camera sees only while the shadows draw. Shadow draw calls in the fight went from 171 to 128.
  - Animation: gear with nothing that moves (most helms, shoulders and shields) keeps its first pose instead of
    animating each frame, and spell effects animate like the raiders, 60 times a second past 20 yards and 30 past 40.
  - Garbage: the pose loop no longer mixes decimals with other values (a root bone's parent pivot was undefined, a
    bone's twist share was 0 or a decimal), which Chrome boxed, 30 bytes a bone a frame. Headless, a garbage
    collection every 770ms (was 420ms). What's left is mostly Three building shader parameters when one depth
    material serves both skinned and instanced meshes in the shadows.
  - See-through: a camera facing particle is kept to about 40% of the screen's height. Only a particle a few of its
    sizes from the camera gets that big.
  - Headless, fight, 1920x1080: 125 FPS, 6.8ms of script (animating 2.4ms, drawing 3.9ms), 440 draw calls. The
    drawing didn't drop here: 330 of the draw calls are the crowd's picture (65 bodies, 155 gear, 108 see-through
    layers of gear and effects), which need different textures. Drawing those in fewer calls needs texture arrays
    for gear, or multi-draw.
- Sixth raid report (your PC, 107s): about 143 FPS in the fight (was 100 to 114), 4.9ms of script (animating 2.0ms,
  drawing 2.6ms), 420 draw calls. GPU 5.3ms, and the see-through things are most of it: 2.8 to 3.0ms in the fight,
  2.0ms before it.
  - The reports now time the see-through things by kind: the units' own see-through parts, the boss, the raiders'
    gear and bodies, the spell effects, the particles and the ribbons. Each kind draws at its own render order
    (`drawOrder` in `wow/particles.ts`), with a marker before it that starts its timer. A frame column also counts
    the particles drawn. Headless, the spell effects' layers were the biggest (0.87ms), then the gear's and the
    particles.
  - A see-through layer faded all the way out still cost its pixels. A fifth of the raid's see-through copies were
    (58 of 286), so a copy with no alpha in a blending that goes by alpha is left out. A particle faded all the way
    out (they fade in at birth and out at death) draws nothing in any blend mode, so it isn't sent either.
  - Not the cause: Ragnaros has no see-through parts, and its own 24 emitters don't run (only gear gets particles
    in the raid test). Finished effects aren't drawn: a fading one hides its model while its particles die out,
    and a pooled one is out of the scene. The raid draws about 3600 particles, most on the casters' hands and their
    bolts' trails (a Fireball's trail is 200).
  - Headless: effects 0.87 to 0.71ms, gear 0.30 to 0.19ms, particles 0.25 to 0.18ms.
- Seventh raid report (your PC): about 125 FPS, 5.8ms of script (animating 2.5ms, drawing 2.9ms), GPU 4.5ms (shadows
  1.1, solid 1.4, effects 0.95, gear 0.36, particles 0.30, ribbons 0.31), 400 draw calls. Solo holds 239 FPS.
- The pose worker (`wow/pose.ts`, `wow/poseWorker.ts`, `wow/poseClient.ts`): the raid test's crowd (raiders, their
  gear and the effects in the crowd) poses off the page's thread. Posing is one class (`Poser`) that both the page
  (your character, the boss, and everything when the page can't share memory) and the worker run. A model keeps its
  animation state (what plays, its times and fades, its events) and asks for its pose each update. Once a frame the
  page sends the worker every model's latest request, and the worker writes each model's bone matrices into memory
  shared with the page, in one of two sets: the page reads the other, and they swap when the frame's poses are done.
  So the crowd's poses are a frame behind, and what hangs off their bones moves when the new pose comes in. Your
  character still poses itself, with no lag.
  - Shared memory needs a cross-origin isolated page, so the practice server sends `Cross-Origin-Opener-Policy:
    same-origin` and `Cross-Origin-Embedder-Policy: credentialless` (credentialless, so the icons from wow.zamimg.com
    still load). Without them (an old server) the crowd poses on the page as before. `?poseworker=0` turns it off.
  - The crowd copies a skeleton into the bone texture only when its pose changed (a far raider poses now and then,
    still gear once), and the bone texture goes up only when one did.
  - Headless (WSL GPU, fight, 1920x1080): animating 2.8 to 1.6ms, but drawing 4.1 to 5.5ms, so about even. Drawing
    here is bound by the GPU process translating to Direct3D 12, so the time saved shows up as waiting.
  - A copy of a painted texture that wraps another way (`setTexture`) shares the picture, which the original let go of
    once on the GPU, so the copy went up empty. The original now keeps it.
- The merged shadow (`wow/crowd.ts`): the raiders' solid shadows (bodies and gear) are one mesh, drawn in one call per
  side. Each vertex has its bones in the bone texture and its raider's place in the world in a places texture, so
  moving raiders need no rebuild; it's rebuilt when raiders join. A body's hidden submeshes are left out when it's
  built. Parts cut by their texture's alpha (hair) and the effects' shadows still draw on their own. Shadow draw
  calls went from 128 to 83. `?mergeshadows=0` turns it off. Headless: 112 to 123 FPS, and 88 to 115 on a slow run.
- Eighth raid report (your PC, worker on): about 125 FPS, 5.0ms of script (animating 1.6ms, was 2.45), 315 draw
  calls (was 400), GPU 4.9ms. Script and GPU time move together in every 10s window, so the GPU timer seems to count
  the GPU waiting for our commands: the limit is how many WebGL calls we send, which Chrome's GPU process translates.
- The merged picture (`wow/crowd.ts`): the raiders' solid parts (bodies and gear) draw in one call per kind of part,
  like the merged shadow: one mesh of all their triangles, each vertex with its bones, its raider's place, and its
  texture's layer in a texture array. The gear's textures go in arrays by size and wrapping (and the second texture
  of the parts that mix two), the bodies' skins in theirs. A kind is a material (the same shader, side and mixing)
  with the same arrays, 45 kinds for the raid. Parts that draw on their own: see-through ones, ones cut by alpha,
  ones whose color or transparency changes (a track with keys that differ; a steady one is fine), and ones whose
  texture isn't the plain kind. `?mergepicture=0` turns it off.
  - Pixel check, one frozen frame (lineup and the fight's view), merge on, off, and on again: the same, apart from a
    few hundred pixels of effects glows that also differ between two frames with the merge on.
  - `?glcount=1` counts every WebGL call by kind (`glCallsAvg` in the report).
  - Two reports at 5 FPS turned out to be Chrome drawing in software ("Microsoft Basic Render Driver") after an
    NVIDIA driver update, not the merge. Headless, the merge's steady frames upload less than without it (120 calls
    and 1.3MB a frame, against 250 and 1.6MB). The room now shows "Software graphics · restart Chrome" when the
    WebGL renderer is a software one (not in test runs: `navigator.webdriver` or `?bench`), and the report's
    `gpuInfo.software` says so.
- Ninth raid report (your PC, RTX 5090, `?glcount=1`): about 172 FPS in the fight (frames 64% at 4.2ms, 36% at
  8.3ms), 3.7ms of script (animating 1.2ms, drawing 2.2ms), GPU 3.7ms, 330 draw calls, 2330 WebGL calls a frame, and a
  garbage collection every 190ms (more than before).
  - The garbage was the call counter's own: each counted call spread its arguments into an array, 2300 a frame. The
    counters are plain functions now, and the upload counter too. The pose worker's requests also go through
    shared memory, so a frame sends it a small message instead of a copy of every request.
  - Particles that add their light (blends 1 and 4) draw together across textures, from texture arrays by size (the
    same sampling as their own textures), each particle with its texture's layer. Added light comes out the same in
    any order, but merged they can't keep their place among the other particle batches, so they draw after them
    (light over smoke): the one change to the look, where added light overlaps smoke. `?particlearrays=0` turns it
    off. Headless: 334 to 292 draw calls, 2410 to 2160 WebGL calls a frame.
- The solid things draw grouped by shader, then texture, then vertices (`opaqueOrder` in `scene.ts`), not by
  material as Three does (every copy of a model has its own materials). Headless: 2160 to 2030 WebGL calls a frame,
  programs 211 to 165, texture binds 390 to 348. Pixel check in a frozen frame: the fight's view 3 pixels apart; a
  close lineup of all 39 raiders about 1100 (0.1%), where two solid surfaces lie at exactly one depth and the other
  one now draws last. `?sortsolid=0` turns it off.
  - The crowd's see-through layers stay as they are: most are spell effects, which come and go and fade by the
    frame, and merging gear layers across models keeps their order only if no other see-through thing draws
    between them.
  - The 50 `frontFace` calls a frame come from the shadows: a solid part's shadow draws its back faces, a double
    sided one both, so the winding flips as the shadow pass goes from one to the other in the scene's order.
  - Headless (fight, 1920x1080): 127 to 144 FPS, drawing 4.8 to 4.1ms, 420 to 326 draw calls, 3040 to 2350 WebGL
    calls a frame (texture binds 528 to 423, vertex array binds 478 to 385, programs 276 to 210, matrix uniforms 271
    to 174).
- Flat particles and see-through layers that don't write depth draw slightly in front of what's at their depth. The
  Barrens ground is at the same height as Fire Nova's ring, and hid it.
- The sim sends every spell you and your pets have (`ShownSpells`), so the effects of procs outside the toolkit
  (the Flametongue Totem proc's fire impact) load before the fight too.
- Combat text scrolls like the Replay view's (MSBT): what you do goes up from the target's frame, buffs you gain go
  down to your frame, crits pop up and stay put (`ui/practice/sct.ts`).
- Particles: packed gravity, key times, sphere emitters, spin, colors, blend modes and model scale. Rendering:
  billboard bones, second textures and env maps, the client's blend modes and lighting, sRGB colors, gear glows.

**Not done yet:**

- Models: `.skel` skeletons (the Orc model doesn't use one), texture animation, particle splines, tails and
  twinkle, Chain Lightning's arc, the character's real look (the client's files don't have it, so customization is
  the defaults and sex is a choice, `practice.modelSex` in local storage).
- To check in game: Lightning Shield's orb count (one orb model, maybe one per charge), when the Orc grunts,
  whether expiring totems sink, which Flametongue proc id the client shows, and what particle flag 0x10 means.
- The BayovCore probe and the feel pass (step 9). The feel checklist is `docs/practice_room_feel.md`.
- Opening an attempt in the Replay view, pointer raw input and the spell queue's cooldown case (check).
- Full screen locks your bound keys and Esc (Ctrl+T reaches the room). Tab is bound, so whether Alt+Tab still
  leaves is a check on Windows. Holding Esc leaves full screen.

## Milestones

Each step says when it's done. Step 0 runs next to step 1, because the dressed character is Milestone 1's biggest
risk and doesn't depend on the sim work.

**Milestone 1: a basic 3D room.** Your enh shaman against a dummy on a flat arena.

0. The asset track and the probe, in parallel with step 1:
   - Spikes 1 and 2. Done when both pass and spike 2's time estimate is written into this doc.
   - The BayovCore probe (a change in the bayovui repo): logs position and facing every frame, the spell cast events,
     error messages and `GetNetStats`. Done when one run of the set path and a short dummy fight are logged.
   - The feel checklist, `docs/practice_room_feel.md`, one line per rule in [Playing like WoW](#playing-like-wow).
1. `sim/practice` session: the request changes, start, inputs and the session queue, run, casting checks, ping and the
   spell queue, the pull and the swap, both fight modes, the world, range and facing, the toolkit, plus the hooks in
   [Core touch points](#core-touch-points-the-whole-list) and the aura hold fix. Done when:
   - Go tests pass for:
     - the queue: the last press in the window wins, a press outside the window fails, ping shifts apply times, an
       input at the fight's end is dropped
     - casting: a cast while moving fails, wrong facing fails, turning away during a cast fails it at the end, no
       swing during a cast, Esc during a cast cancels it and keeps the target
     - auto attack: no swing before you turn it on, a `/startattack` macro and a right click both turn it on, leaving
       melee range stops swings and coming back resumes them without a timer reset
     - the pull: your first harmful hit is the pull at 0 in Free mode, a Lightning Bolt cast before the pull counts
       from its landing, a Magma Totem's first pulse pulls, turning auto attack on in range pulls with a swing at 0,
       walking into range with auto attack on pulls, a failed cast doesn't pull, the state at the pull is the same
       with and without the swap (resources, auras with time left, cooldowns, pending events relative to the pull)
     - the prepull window: a totem dropped before the pull is up at the pull, a totem replaced before the pull loses
       its buff at once, waiting 70 minutes before the pull and then casting works (two window swaps), every toolkit
       spell cast at minus 9 minutes has the right aura times
     - Free mode: Stop ends the fight and the summary uses the stop time, the 60 minute cap ends it, the dummy is
       never in execute range unless the setting says so
     - Encounter mode: a hit before Start Encounter starts a Free fight, the countdown puts the pull at its end, a
       Lightning Bolt landing exactly at the countdown's end counts from 0, one landing earlier is an early pull, the
       set length ends the fight, Stop marks it stopped early
     - the toolkit: an imbue cast before or during the fight replaces the active imbue, both ways of swapping under
       your own Windfury Totem keep the right totem buff, a swap to and from Rockbiter moves its attack power, an imbue
       press never fires its proc, Ghost Wolf costs its cast and mana and speeds you up, Earthbind Totem replaces
       Strength of Earth, a Magma Totem out of reach does no damage, a potion and a sapper used from the bags fire
     - building every class at 60 with its whole toolkit doesn't panic
   - The rotation check: the enh APL's own casts with their sim times (prepull ones included, plus consumables and a
     `startAttack` at 0) are fed back as inputs, in Encounter mode with the same length, a 0 countdown and 0 ping. An
     APL casts inside the event that triggered it and an input runs after it, so single seeds drift apart. So this
     test runs with labeled random streams (`UseLabeledRands`) over 100 seeds, and the mean DPS must match the APL
     run's within the noise. It's a correctness test only, the room itself never runs other seeds.
   - `make test` passes and the `identical()` check shows no change on the DPS Report requests (the aura hold fix
     gets its own run).
2. Replays and saved attempts. Done when a replayed attempt's log matches the live final sim's log byte for byte, on
   the same build, and the attempt opens in the Replay view.
3. `practice.wasm`, in a dedicated worker with its own clock. Done when spike 4 passes and `go list -deps ./sim/wasm`
   shows no practice package.
4. The practice server: the request route, records, bindings with the default table, CVars, macros, saved edits,
   assets, game tables, attempts, and the data root setting. Done when golden file tests pass for the Lua parser, the
   WTF parser, the slot table and the macro parser.
5. The 3D scene: floor, dummy, camera, movement, jumping, targeting, with sized capsules, and keys 1 to 3 from the
   default bindings mapped to casts. Done when you can walk up to the dummy, hit it with those keys, and the movement
   and camera numbers match the probe. Press to screen is measured here.
6. The model renderer: the ported readers plus your race's body with its animations, in place of the capsule. Done
   when your character runs, jumps, attacks and casts with the right animations.
7. The HUD: the spellbook (with its bag tab), your bars (editable) with bindings, unit frames, cast bar, swing timer,
   your auras, combat text, the fight settings, the Start Encounter and Stop buttons, the countdown, the fight timer,
   the live meter and the fight summary. Done when a headless Chrome smoke test plays a scripted fight with no console
   errors, the meter's total matches the summary, and the summary shows. The sweep paint cost is measured here.
8. Your gear on the model: dressed character, attachments, sheathing, icons and sounds. Done when your character
   fights the dummy with your gear on.
9. The feel pass: every "check" in [Playing like WoW](#playing-like-wow) measured with the probe and matched, and the
   feel checklist played through.

**Milestone 2: a boss.** A boss that hits back and moves, with simple mechanics (stand still, move out, interrupt) that
change the best rotation. Line of sight and hit box sizes. Totem buff reach and several enemies (see the known gaps in
[World and movement](#world-and-movement)). Spike 3 runs before it. The boss's movement and mechanics live in
`sim/practice` and use the core's spells and auras.

**Milestone 3: raid bots.** The raid sim's members as bots around you, each on its APL, with raid frames. This uses the
raid sim's whole-raid engine (its milestone M1b in `docs/raid_sim.md`) as it is.

**Milestone 4: friends.** Spikes 5 and 6 run before it. See [Multiplayer](#multiplayer-milestone-4).

## Spikes

1. **Reading the `wowf` storage.** Run TACTTool's Windows release on the data root with wowdev's `WoW.txt` keys. Pull
   a batch of files for one dressed Orc by FileDataID, check their magic, and count how many Forever-only files are
   encrypted. Open wow.export 0.2.20 on `wowf`, load an Orc shaman with gear, and keep its view as the reference. Pass:
   a batch extracts from Go with every file's magic right and none zero-filled.
2. **A dressed character.** Port wow.export's readers and character code to Three.js for one race and one gear set.
   Measure how long the rest takes, and the frame time with 40 characters on screen. Pass: the dressed Orc draws
   right next to wow.export's view and plays Stand, Run, a one and a two hand attack and SpellCast from its `.skel` and
   `.anim` files.
3. **Spell visuals** (before Milestone 2). Size the particle and ribbon work from WebWowViewerCpp's emitters, and
   compare with three.quarks plus hand made effects for the few spells enh uses.
4. **WASM in real time.** Measure:
   - cold and warm startup (fetch, compile, Go init, `RegisterAll`, `NewSim`)
   - over a 3 minute fight at ping 0, the time from a key press to its event, and the garbage collector pauses
   - the worst-case swap at the pull: a full 10 minute prepull window with totems and regen ticks

   Pass: key to event under one frame (7ms at 144Hz) at p99, and the swap under one frame at p99. When key to event
   fails, we cut per-event allocations first, then the log lines we send. When the swap fails, a second worker holds
   the spare sim, or the prepull window gets shorter.
5. **WASM against native** (before Milestone 4). Run the same seed and inputs in the browser and in a native Go test,
   and diff the logs.
6. **Game tables from the local client** (before Milestone 4). Read one table with a DB2 reader and the build's table
   layouts, merge `DBCache.bin` hotfixes, and diff it with wago's CSV for the same build.

## Open questions

1. **A "Practice" button in the class sim.** It would open the room with the current settings, but it changes the
   regular sim UI, so it needs the user's OK. Until then, settings come in through the export.

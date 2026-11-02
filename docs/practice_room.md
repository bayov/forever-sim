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

**History:** first drafted 2026-11-02. Reviewed twice the same day by five reviewers each (the sim engine, the WoW
assets, the keybinds and bars against the real client files, the client and networking, scope, the WoW feel rules,
the doc as a spec, the browser's limits, and the build and tooling), and rewritten from their findings and the user's
answers.

**Terms:**

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
| Assets, attempts, your bar edits | `~/forever-practice/assets/`, `attempts/`, `settings/` | outside the repo |

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
  correct goes into the core as a normal sim change with its own review, never as part of practice work.
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
| Class hooks for spells the sim keeps as a setting, opt-in, listed per class as we find them (the first: shaman imbues, see [Your toolkit](#your-toolkit)) | the class code reads "the current imbue" instead of "the imbue in the settings" | none when off |

Extra spells and auras come in through `core.EnvHooks` (`environment.go:180`), which already exists for tools that
watch a sim from outside, like `tools/logcheck`. It runs before the sim is finalized, while spells and auras can still
be registered.

`ManualControl` is needed because an empty APL isn't idle. Each time it finds nothing to do, it waits 50ms through
`WaitUntil`, which sets the GCD (`apl.go:227`, `gcd.go:90`). It runs after every swing, cast, and energy or rage tick
(`attack.go:432`, `cast.go:368`, `energy.go:166`, `rage.go:440`). So the GCD would look busy most of the time, a key
press at the GCD's end would lose to it, and the log would get "Pausing GCD for 50ms" 20 times a second.

`SkipPullStart` is needed because the pull starts every unit's swings at time 0, and the sim runs a swing before any
other action at the same time (`sim.go:761`, `attack.go:650`). A cancel at 0 from the session comes after your first
swing has already landed.

### Other shared files (one time, each its own commit)

- `tsconfig.json`: exclude `ui/practice`, which gets its own `ui/practice/tsconfig.json`. Otherwise one practice type
  error breaks `make` for every session (`makefile:46` runs `tsc --noEmit` on the whole tree).
- `makefile`: filter `ui/practice` out of `UI_SRC`, so practice edits don't rebuild the regular UI. Add the targets
  `practice-wasm`, `practice` (wasm, page, server) and `practice-test`.
- `package.json`: add `three` and `@types/three` once. Changing the lock file makes every session's next `make` run
  `npm ci` again, so it happens once, on its own.
- `tools/dps_report/build_artifact.mjs`: the deploy guard. It fails when the build's module ids contain `/ui/practice/`
  or the page contains `forever-practice`. About five lines.

Tests: practice Go tests carry the build tag `practice`, so `make test` doesn't run them or wait for them.
`make practice-test` runs `go test -tags practice ./sim/practice/... ./cmd/practice/...` and the practice type-check.
"`lib.wasm` unchanged" means `go list -deps ./sim/wasm | grep practice` prints nothing.

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
are all exported:

1. `sim.RegisterAll()`, then `NewSim` from the same proto request a sim run uses, with no duration variation. The
   fight length is the mode's: the set length in Encounter mode, and a 60 minute cap in Free mode. Execute phases,
   `GetRemainingDuration` and the end check all depend on it.
2. A nonzero seed picked by the session and stored with the attempt. Seed 0 would mean "pick from the clock", and the
   picked seed isn't readable from outside (`sim.go:30`).
3. `Reset()`, then `PrePull()`.
4. Set `CurrentTime` to the earlier of its value and minus the free time window (10 minutes). This comes after
   `PrePull`, because `PrePull` sets the clock when prepull actions exist (`sim.go:686`, a druid talent registers one).
5. Set `sim.Log`, since only the sim's own run sets it (`sim.go:373`). Write your distance and facing, since `Reset`
   clears them (`unit.go:553`).
6. Your unit gets an empty APL and `ManualControl`, so it never acts on its own, and its auto attacks get
   `SkipPullStart`, so they start off. Other units (the boss, pets, later the raid bots) keep their own APL.

### Fight modes and the pull

The sim's fight always starts at time 0: its encounter length, execute phases and pull are all counted from there.
But you decide when the fight starts, so the room can't know in advance how long your prepull is. It solves that by
shifting time once it knows:

1. Before the pull, the session runs the sim live from minus 10 minutes. You walk around, drop totems, buff, and see
   your mana and cooldowns move. Nothing harmful has happened yet.
2. Once the pull time `t` is known (still below 0), the session swaps in a fresh sim (built ahead of time in the
   background, with the same seed) and feeds it your inputs so far, each moved by minus `t`. The pull is now at
   exactly 0, and everything before it keeps the same spacing. Nothing has rolled for damage yet, so the fresh sim is
   in the same state the old one was in.
3. If you stay longer than 10 minutes before the pull, the session does the same swap with a new 10 minute window.

Step 1 tests that the state at the pull is identical with and without the swap. Spike 4 measures how long the swap
takes, since it can happen on the frame you attack.

**When the pull is.** Your first harmful hit landing: a spell's damage, an ability, or your first swing. So a
Lightning Bolt cast before the pull and timed to land at it counts from the moment it lands, like the in-game meters
count. The session knows the landing time when the cast starts (cast end plus travel time) and swaps then. If the cast
is interrupted, the swap is undone the same way.

**Free mode** (the default):

- The pull is your first harmful hit.
- The fight goes on until you click Stop (or a key bound to it). Stop ends the fight at that moment. The session stops
  stepping and runs `Cleanup()` with the stop time as the fight length. Step 1 checks the metrics use it.
- The 60 minute cap ends the fight if you never click Stop.
- The dummy never enters execute range, as with a training dummy in game. A setting can put it in execute range from
  the start, for classes with execute abilities.

**Encounter mode:**

- Clicking "Start Encounter" starts the countdown (5s by default). The pull is at the countdown's end, so the swap
  happens on the click.
- If your first harmful hit lands before the countdown ends, that's an early pull, as in game. The fight starts right
  then (one more swap), and the HUD says so.
- The fight ends on its own at the set length, as in a sim run. Execute phases follow the encounter's settings.
- Stop still works, and marks the attempt as stopped early.

In both modes, auto attack only swings when you have turned it on and you're in melee range. If turning it on in
range is your first harmful action, that's the pull, and the first swing comes at once, as in game.

### Running in real time

`sim/practice` holds a stateful session:

- **Input(input)** adds a pending action for a key press at its apply time, in whole nanoseconds, with a fixed lowest
  priority. It goes through `Step` like any other event, so the other units' idle check skipping (`apl_idle.go`) sees
  it.
- **Run(until)** runs events whose time is before `until` and stops. It reads the next event time from
  `NextEventTime()` and never moves the clock to `until` itself. Moving the clock to frame times would settle regen
  and expire auras in frame sized steps, and then results would depend on frame rate.
- **The end** is final: Stop, or the set length in Encounter mode. `Step` drops the last pending action before the end
  check. `Cleanup()` adds up the metrics.

The worker keeps its own clock and pushes events as they happen, instead of waiting for one call per frame. A
`MessageChannel` loop drives it, because nested timers are held to at least 4ms, in workers too. The clock pauses
when the tab is hidden or loses focus, and one catch-up step is capped at 0.25s, as the Replay view does
(`replay.tsx:963`).

Between events the HUD shows values like energy and the cast bar by computing them for the display time, without
changing sim state. Aura expiry isn't an event in the sim (it happens at the next event), so the HUD times expiries
itself. The live DPS meter sums spell metrics as `getDamageDone` does (`library.go:282`), because unit metrics only add
up in `Cleanup`.

The worker is a dedicated one running `practice.wasm`, not one from the sim's `WorkerPool` (`ui/core/sim.ts:123`). The
WASM entry points are synchronous `js.FuncOf` calls in the style of `raidSim` (`sim/wasm/main.go:133`).
`practice.wasm` registers every class (about 36 MB), so any class sim's export works and Milestone 3's bots work. The
live heap is about 1.3 MB, so garbage collector pauses come from per-event allocations (log formatting, proto
marshalling), not from the database.

### Casting your spells

For each cast input, the session first runs its own checks, in this order, and answers with the game's error text
when one fails:

1. The first harmful spell or ability before time 0 triggers [the pull](#the-pull) first, so it lands at 0. The sim
   doesn't count damage before 0 (`spell_result.go:542`), so nothing harmful may happen before the pull.
2. Moving or in the air, for a spell with a cast time.
3. Range and facing (see [Playing like WoW](#playing-like-wow)).
4. GCD and cooldown timers, cost (`Cost.CostFailureReason`, `mana.go:443`), already casting.
5. `Spell.CanCast`.

Only then does it call `Spell.Cast`. A failed `Cast` can use up the cooldown, because the cooldown is set before the
GCD check (`cast.go:273`, `cast.go:289`), so `Cast` is never called without these checks passing. During a channel,
the session cancels the channel first (`ChanneledDot.Cancel`), since `InterruptHardcast` doesn't stop channels.

### Your toolkit

You can cast anything your class knows at your level, at the ranks you know. The list comes from the client's
spell tables for the class and level, and the room's spellbook shows all of it. Each spell falls in one of three
kinds:

1. **Spells the sim has** (most class spells: shocks, Lightning Bolt, Stormstrike, every totem the sim models,
   Healing Wave and the other heals, shields). The shaman class code registers all of them, whatever the rotation
   uses (`shaman.go:276`). They're cast as described above.
2. **Spells the sim keeps as a setting.** Shaman imbues are the first case: the class code registers and turns on
   only the imbue the settings pick (`shaman.go:66`, `windfury_weapon.go:105`). So:
   - Through `core.EnvHooks`, the session registers all four imbues with the class code's own exported functions
     (`ApplyWindfuryImbue` and the others), and keeps only the active one on.
   - You enter the room with the settings' imbue up. Casting an imbue, at any time, turns the old one off and the new
     one on for that weapon, with the game's GCD, mana cost and duration.
   - Where the class code reads the settings' imbue during the fight (for example the rule that a Windfury Weapon
     imbue turns off the Windfury Totem buff, `shaman.go:92`), an opt-in class hook makes it read the current imbue.
     It's listed in the core touch points.
   - Other classes' setting-only spells (blessings, auras, aspects, poisons) get the same treatment when we add their
     class.
3. **Spells the sim doesn't have** (Ghost Wolf, Purge, Earthbind Totem, Frost Resistance Totem, Cure Poison, Water
   Walking and the like). The session registers them through `core.EnvHooks` with `RegisterSpell`, with the cost,
   cast time, GCD, cooldown and range from the client's spell tables. Effects that change the fight are coded in
   `sim/practice`:
   - Ghost Wolf: its run speed, and it ends when you cast.
   - A totem the sim doesn't have still takes its element's slot, so Earthbind Totem replaces Strength of Earth, as
     in game. When that needs the shaman's slot code, it's an opt-in class hook.
   - Effects that mean nothing against a dummy (Water Walking, Far Sight) only cost their GCD and mana.

Consumables and other players' buffs come from the settings.

### Inputs, ping and the spell queue

Your ping splits the round trip the way the game does. An input's **apply time** is the press time plus half a ping.
The sim's answer reaches the screen half a ping after that.

- When the apply time is within `SpellQueueWindow` (400ms by default) of the GCD's end or the current cast's end, the
  cast fires exactly at that end and replaces any spell already queued. So a queued press costs no ping, as in game.
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
type Input =
	| { id: number; pressedAt: number; kind: 'cast'; spellId: number; target: UnitRef }
	| { id: number; pressedAt: number; kind: 'useItem'; slot: number; target: UnitRef }
	| { id: number; pressedAt: number; kind: 'startAttack' | 'stopAttack' | 'clearTarget' | 'cancelCast' }
	| { id: number; pressedAt: number; kind: 'target'; target: UnitRef }
	| { id: number; pressedAt: number; kind: 'move'; velocity: [number, number]; facing: number }
	| { id: number; pressedAt: number; kind: 'jump' };

type EventBatch = {
	simTime: number; // nanoseconds
	logLines: string[]; // the sim's text log, parsed by logs_parser.tsx
	answers: { inputId: number; ok: boolean; error?: string }[];
	units: { ref: UnitRef; pos: [number, number]; facing: number; airborneUntil?: number;
		nextSwingAt?: number; castEndsAt?: number }[];
};
```

`pressedAt` is the key event's `timeStamp` mapped to the worker's clock through `performance.timeOrigin`. Times are in
milliseconds on the client and whole nanoseconds in the sim.

### Replays and saved attempts

An attempt is one JSON file in `~/forever-practice/attempts/`:

```json
{
	"version": 1,
	"simCommit": "<git commit of the sim build>",
	"request": { "...": "the RaidSimRequest as proto JSON" },
	"seed": 123456789,
	"pingMs": 60,
	"mode": "free",
	"lengthMs": 184300,
	"countdownMs": 5000,
	"earlyPull": false,
	"inputs": [{ "appliedAt": -8200000000, "input": { "kind": "cast", "spellId": 8512 } }],
	"result": { "...": "the fight summary" }
}
```

A replay mode feeds the inputs into a normal run at their applied times (already shifted, so the pull is at 0),
starting a little before the first one. It turns
on the same log options the Replay view needs and gives a full `RaidSimResult`, so the attempt opens in the Replay
view. A bug report is a small file.

Exact replays hold only on the same sim build and the same platform. WASM and native Go can differ in the last bits of
`math.Exp` and `math.Log`, so a native Milestone 4 server and a browser may drift apart. Spike 5 checks how much. When
`simCommit` differs from the running build, the attempt keeps its result and its replay is marked as from another
build.

### World and movement

The world lives in `sim/practice`: ground positions with velocity, facing, an airborne flag with the jump's end time,
and range. It follows the leveling world's model (`leveling_world.go`) but doesn't reuse its code, which is tied to
leveling sessions inside the core. We don't use `Unit.MoveTo`: its movement aura cancels auto attacks
(`movement.go:107`), which is wrong for strafing in melee.

Your movement goes to the session as timestamped velocity and facing changes, not as per-frame positions. The world
turns them into pending actions at the times something changes (you reach or leave range, you turn away). That keeps
replays exact. The client moves your model at once and the sim agrees, as WoW trusts the client's movement.

How the world talks to the sim. Every write happens inside a pending action, because `EnableAutoSwing` and
`InterruptHardcast` use the current sim time, and between steps that's the last event's time.

- **Distance:** the core knows melee range as a flat 5 yards (`IsInMeleeRange`, `pvp.go:30`). The world works out the
  real melee range (see [Range and facing](#range-and-facing)) and writes `Unit.DistanceFromTarget` as the center
  distance minus how much the real range exceeds 5 yards. Then the core's 5 yard rule agrees with ours. Other readers
  (Lightning Bolt travel time, `spell.go:678`, and explosive radii) are off by at most a yard or so, which is fine.
- **Melee range:** when you leave or enter it, the world calls the cancel and enable pair that PvP downtime uses
  (`pvp.go:68`), because auto attacks only check range when enabled (`attack.go:755`).
- **Facing:** it writes `PseudoStats.InFrontOfTarget` (`character.go:246`), which the core reads at each roll for
  parry and block.
- **Moving:** it calls `Unit.InterruptHardcast` when movement or a jump starts during a hard cast. `IsMoving` stays
  false, because we don't use `MoveTo`, so the session's own checks handle casting while moving.

Known gaps: totem buffs don't drop when you walk away from the totem (totem reach exists only in leveling,
`shaman/totem_reach.go:17`). `DistanceFromTarget` is one number per unit, which limits fights with several enemies.
Both matter from Milestone 2 on, and the fix is decided then.

### Events out of the sim

The worker streams the sim's text log lines with each batch, and the client reuses `logs_parser.tsx`, which already
turns them into typed records (casts, damage, auras, resources). Typed events come later, only if parsing is too
slow.

The live HUD also needs the next swing time and the cast end time. The Replay view gets those by looking ahead in the
finished log, which a live stream can't do. The session reads them from the units after each batch
(`AutoAttacks.NextAttackAt`, `Unit.Hardcast`) and sends them in `units`.

### HUD overlay

The Replay view draws your EllesmereUI frames, the Bayov WeakAuras class pack and MSBT style combat text. Its model is
built once from a finished fight (`buildReplayModel`, `replay_model.ts:327`), and parts of it look ahead: swing timers,
cast pushback, and the bar layout chosen from cooldowns used in the fight. The draw functions are "state at time t"
lookups, so a model that only grows forward in time can feed them.

So we reuse the markup, CSS and helpers (`updateIcon` and `updateCastBar` in `replay.tsx`), and write an incremental
model fed by events in `ui/practice/`. Helpers that aren't exported are copied. Your real action bars are new code.

The Replay view's cooldown sweep is a `conic-gradient` driven by a CSS variable set every frame
(`_replay.scss:790`). At 144Hz that repaints every icon every frame on the main thread, next to Three.js. Step 8
measures it. If it's too slow, the sweeps move into one 2D canvas. Combat text animates only `transform` and
`opacity`.

### Rendering

- **Three.js.** It has no particle system, so particles come later from three.quarks or our own M2 particle code.
  Frame time with 40 characters is measured in spike 2, not guessed.
- **WoW models drawn directly**, not converted to glTF (see [Assets](#assets-local-only)). Weapons and helms hang on
  M2 attachment points: hands 1 and 2, spell palms 21 and 22, back sheaths 26 to 28, hip sheaths 32 and 33. Weapons
  are sheathed out of combat (where, from the item's sheathe type) and drawn in combat.
- **Animations by AnimationData id:** stand 0, run 5, one hand attack 17, ready 26, precast 31, cast 32, channel 76
  and 77. Attacking while running blends the upper body over the run. Some spells name their own animation through
  their visual kit.
- **Placeholders:** until the model renderer works, units are capsules with their race's real height and combat
  reach, so camera and range checks hold before real models arrive.
- **Loading:** `instantiateStreaming`, so Chrome caches the compiled wasm. The worker needs `wasm_exec.js` from
  GOROOT's `lib/wasm`, the way `vite.build-workers.ts` gets it.

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
| Strafe | 7 yd/s | check |
| Forward plus strafe | 7 yd/s, not faster | check |
| Backpedal plus strafe | 4.5 yd/s | check |
| Turn with keys | 180°/s | check (no Blizzard source) |
| Turn with right drag | at once, to where the camera looks | right |
| Jump | about 1.64 yd high, 0.83s in the air, direction locked, you can still turn | check |

- Both mouse buttons held runs forward. Autorun (Num Lock by default) runs until you press a movement key.
- Click to move is off (`autoInteract` 0).
- Units don't collide with each other. You walk through the dummy and other players, as in game. Unit cylinders are
  only for clicks and range.
- Moving or jumping stops a hard cast or channel, and the cast bar shows "Interrupted". Turning in place doesn't.
  Instants work while moving and in the air.

### Range and facing

- **Melee range** is the larger of 5 yards and the two units' combat reach plus 4/3 yards. When both units move (not
  walking), it grows by 8/3 yards, which is why you can keep hitting a running target. Blizzard confirms this extra
  reach exists but gives no numbers. Tauren have a larger reach. Status: check.
- **Spell range** counts the target's combat reach. Whether your own reach (about 1.5 yards) is added too is a check,
  measured at max range on the dummy.
- **Facing:** spells and abilities need the target in your front 180°. Auto attack may use a narrower 120°. Backstab
  style abilities need you outside the target's front 180°. Status: check. Parry and block only happen from the front,
  which the sim already handles through `InFrontOfTarget`.
- **Out of range:** auto attack stays on and stops swinging. It resumes when you're back in range.

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

Errors show in red at the top center with the game's error sound. The texts come from GlobalStrings:

| When | Text |
| --- | --- |
| Cast time spell while moving | "Can't do that while moving" |
| Spell or ability, target behind you | "Target needs to be in front of you." |
| Auto attack, target behind you | "You are facing the wrong way!" |
| Backstab style ability from the front | "You must be behind your target." |
| Spell or ability out of range | "Out of range" or "Out of range." (which one shows is a check) |
| Auto attack out of range | "You are too far away!" |
| GCD or cooldown | "Spell is not ready yet." or "Ability is not ready yet." (check which) |
| Already casting | "Another action is in progress" |
| Cost | "Not enough mana", "Not enough energy", "Not enough rage" |

How often the same error repeats while you hold or mash a key is a check.

### Targeting

- Left click picks a unit by its cylinder, not its mesh (`interactOnLeftClick` 1). Right click targets and attacks.
- Tab picks the nearest enemy in front of you, then cycles (`TargetNearestUseNew` 1). When your target is in combat,
  Tab cycles only enemies in combat (`TargetPriorityCombatLock` 1).
- Esc clears the target. Clicking the ground keeps it (`deselectOnClick` 0, "Sticky Targeting").
- `@mouseover` is the unit under the cursor. A right press and release with no movement is a click, and the unit is
  picked at mousedown.
- No soft targeting: it's for gamepads only.
- `autoSelfCast` is on by default.
- Enemy nameplates are off by default (`nameplateShowEnemies` 0, V toggles them), and show up to 45 yards.

### Camera

- **Left drag** orbits the camera around you without turning you. **Right drag** turns you to the camera's direction,
  and pitches the camera.
- **The drag cycle:** the room locks the pointer on mousedown and unlocks it on mouseup, and the cursor comes back
  where it started (check on Chrome for Windows). Chrome blocks a re-lock for about a second only when you leave the
  lock with Esc yourself, and after the first "Press Esc" bubble the later locks are silent. A short Esc goes to the
  page, so it clears your target mid-drag as in WoW. The lock starts a moment after mousedown, so the room uses plain
  mouse moves until it's active.
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

- **Sounds:** weapon swings and impacts, cast start and finish, proc sounds, the error sound, and your WeakAuras'
  sounds. Many players react to sounds more than to icons.
- **Combat text** in your MSBT style, and your EllesmereUI frames and WeakAuras as in game.
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
  `desynchronized` hint, and spike 4 measures press to screen.
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
   - cast, GCD and swing timing from the combat log
   - camera zoom, field of view and turn speed from screen recordings of the same moves
2. **Side by side.** The same short fight on the in-game dummy and in the room, with the logs compared.
3. **A feel checklist** you play through after each change to movement, camera or casting, one line per rule above.

## Assets (local only)

The Forever client's data sits in the shared CASC storage (WoW's archive format) of the WoW install. Its product code
in `.build.info` is `wowf`, version 1.60.1.70707 today. The install folder is a setting. On this PC the practice
server reads it at `/mnt/d/Gaming/World of Warcraft`.

**The pipeline:**

1. The practice server extracts raw files by FileDataID (the client's file number): M2 models, `.skin`, `.skel` and
   `.anim` files, BLP textures, OGG sounds and DB2 tables. It runs TACTTool from TACTSharp's prebuilt release as a
   command-line tool. The Windows build, run through WSL, reads `D:` directly instead of through the slow WSL mount of
   the 139 GB `Data` folder. Whether it needs the .NET runtime is part of spike 1 (dotnet isn't installed). CascLib
   with a cgo binding is the fallback. There's no usable Go CASC library. Files are cached in
   `~/forever-practice/assets/`.
2. DB2 tables come from the local client, plus its hotfix cache (`DBCache.bin`), not from wago.tools. That way each
   friend matches their own client build and needs no internet.
3. The browser runs our own Three.js model renderer. It starts from three-m2loader (MIT, reads the format, skins and
   animations) and ports wow.export's character assembly (MIT, about 4 to 5 thousand lines of browser JavaScript).

wow.export added Forever support in 0.2.20 (2026-11-02). It has no command line or remote control, so it can't drive
the pipeline, and its glTF export leaves weapons unattached and drops particles. We use its GUI to check our output.

**A dressed character** is the biggest asset job in Milestone 1 (days of porting, one to two weeks from scratch):

1. ChrModel gives the body model. Customization choices give skin textures and geosets (the body parts shown or
   hidden, like hair and ears).
2. Each item leads through ItemModifiedAppearance and ItemAppearance to ItemDisplayInfo. Its geoset groups switch
   body parts (gloves, boots, robe, cape, belt), and HelmetGeosetData hides hair under helms.
3. ItemDisplayInfoMaterialRes and ComponentTextureFileData give one texture per body section (arm, hand, torso, leg,
   foot), painted onto the body texture.
4. ComponentModelFileData gives the models that hang on attachment points.

**Tables we need:**

- Characters: ItemModifiedAppearance, ItemAppearance, ItemDisplayInfo, ItemDisplayInfoMaterialRes, ModelFileData,
  TextureFileData, ComponentModelFileData, ComponentTextureFileData, HelmetGeosetData, ChrModel, ChrRaceXChrModel,
  ChrCustomization tables, CharComponentTextureLayouts and Sections, ItemSparse (sheathe type).
- Animations, spells and sound: AnimationData, SpellXSpellVisual, SpellVisual, SpellVisualEvent, SpellVisualKit,
  SpellVisualKitEffect, SpellVisualKitModelAttach, SpellVisualMissile, SpellChainEffects, SoundKit, SoundKitEntry,
  WeaponSwingSounds2, WeaponImpactSounds.

**Spell visuals and imbue glows** are mostly M2 particles and ribbons. Milestone 1 has animations, icons, sounds and
combat text, and no particles. Later the same model renderer gets particles, with WebWowViewerCpp's emitters as the
reference (read only, it has no license file).

**Keeping it local:** the assets live in `~/forever-practice/assets/`, outside the repo. Only the practice server
serves them. The deploy guard in `tools/dps_report/build_artifact.mjs` fails a DPS Report build that touches practice
code or paths, and `dist/forever/` never holds practice files.

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

On Forever, every class's stance or form pages bar 1 to slots 73-120. Milestone 1 assumes page 1.

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
- Bar spells the sim doesn't have (Ghost Wolf, heals) show greyed out.
- Buttons can also be clicked.

### Editing bars and binds in the room

- **Bars:** pick up a spell, item or macro from a spellbook or a slot, and it sits on the cursor as in WoW. Click a
  slot to place it, or drop it off the bars to clear. This uses pointer events, not the browser's drag and drop.
- **Binds:** a quick keybind mode, as in WoW: hover a button and press the key or mouse button, with modifiers.
- Edits are saved per character in `~/forever-practice/settings/<GUID or "default">.json`, on top of what the client
  files and the record say: `{ "version": 1, "slots": { "<slot>": { "kind": "spell", "id": 0 } }, "binds": { "<KEY>":
  "<COMMAND>" } }`. They never write back to the WoW folder. "Reset" puts the client's layout back.

### Macros

What the user's shaman macros use, which Milestone 1 supports:

- `/cast [@mouseover,harm,nodead][] Stormstrike`: every damage spell is a macro like this.
- Conditions: `@mouseover`, `@target`, `@player`, `harm`, `help`, `dead`, `nodead`, `mod:alt` (and shift, ctrl), and
  the empty `[]` fallback.
- `/startattack`, `/cleartarget [dead][help]`.
- Ranks: `Frost Shock(Rank 1)`, any case.
- `/use 13` and `/use 14` (trinket slots), and `/use` item lines where the first usable one fires.
- `#showtooltip` sets the button icon.

A macro with anything else (`/castsequence`, `/tm`, `/run`, `/click`, `/focus`, `/target`, `/bayov`) gets an
"unsupported" badge. `macros-cache.txt` blocks are `VER 3 <id> "<name>" "<icon>"`, the body, then `END`.

### Settings (CVars)

CVars live in three files: `WTF/Config.wtf`, the account's `config-cache.wtf` and the character's. They are written
only when changed, so a missing CVar means the default from the Forever CVar list. The room reads every CVar named in
[Playing like WoW](#playing-like-wow), plus `SpellQueueWindow` (the user hasn't changed it, so 400ms).

## The practice server

`cmd/practice/` serves the built page from `dist/practice/` and these routes:

| Route | What |
| --- | --- |
| `GET /api/records` | BayovCore records (GUID, name, realm, class, level), plus "No record" |
| `GET /api/character/<guid>` | bindings (defaults + WTF + saved edits), slots, macros, CVars |
| `PUT /api/settings/<guid>` | save bar and bind edits |
| `GET /api/assets/<fileDataId>` | one extracted file |
| `GET /api/db2/<table>` | one DB2 table as JSON |
| `GET`, `POST /api/attempts` | list and save attempts, `GET /api/attempts/<id>` for one |

For development, the practice Vite dev server runs on its own port, proxies `/api` to `cmd/practice`, and imports the
`jsxInject` setting and the `/forever/assets` middleware from `vite.config.mts` without editing it.

## Fight summary

The room sims only the fight you played: one iteration, no other seeds and no averages. Everything counts from the pull to the end
(Stop, or the set length in Encounter mode). At the end it shows:

- your DPS and damage by spell
- idle GCD time
- buff and debuff uptimes
- time out of range
- swings pushed back by casts
- time with resources capped

The whole fight opens in the Replay view.

## Where it lives

Its own page, served by the practice server. You bring a class sim's settings (character, gear, talents, buffs,
rotation, fight length) through the class sim's existing export, as a link or JSON. Presets and pasted BayovCore
exports come in the same way, since they already load into the class sim. The keybinds come from a BayovCore record
you pick separately, because the sim character and the in-game character aren't the same thing.

Fight rules for Milestone 1 are the two modes in [Fight modes and the pull](#fight-modes-and-the-pull). Before the
pull you're free to prepare for as long as you like. The HUD has the Start Encounter button with its countdown and
length settings, the Stop button, the countdown, and the fight timer. Keys for Stop, restart and leave.

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

Lockstep (every client runs the sim and waits for everyone's inputs) doesn't fit 40 players: every input waits for the
slowest player, and joining mid-fight is hard. So the server decides everything, as WoW does.

**Friends' setup.** The host serves the client over a tunnel such as Tailscale, so friends need only Chrome. Each
friend's browser reads their own WoW install's assets through a small local helper or the File System Access API, so
we never send game files anywhere. The details wait for Milestone 4.

## Milestones

Each step says when it's done. Spike 1 runs first, next to step 1, since it's the riskiest.

**Milestone 1: a basic 3D room.** Your enh shaman against a dummy on a flat arena.

1. `sim/practice` session: start, inputs, run, casting checks, ping and the spell queue, your prepull, the world,
   range and facing, plus the three core hooks. Done when:
   - Go tests pass for: the last press in the queue wins, a press outside the window fails, a cast while moving
     fails, leaving melee range stops swings and coming back resumes them, wrong facing fails, ping shifts apply
     times, a totem dropped before the pull is up at the pull, the state at the pull is identical with and without the
     time swap, the first harmful hit is the pull at 0 in Free mode, a precast Lightning Bolt timed to land at the
     pull counts from its landing, Stop ends the fight and the metrics use the stop time, the Encounter mode countdown
     puts the pull at its end, an early pull starts the fight at once, the set length ends the fight, no swing before you turn auto attack on, a `/startattack`
     macro and a right click both turn it on, an imbue cast before or during the fight replaces the active
     imbue and the Windfury Totem rule follows the current imbue, Ghost Wolf costs its GCD and mana and ends on a
     cast, Earthbind Totem replaces Strength of Earth, an input at the fight's end is dropped, and no swing during a
     cast
   - for each of 100 seeds, the APL's own casts with their sim times (prepull ones included), fed back as inputs with
     the same seed, give the same DPS, or each difference is explained (a correctness test only, the room itself never
     runs other seeds)
   - `make test` passes and the `identical()` check shows no change on the DPS Report requests
2. Replays and saved attempts. Done when a replayed attempt gives an identical log on the same build, and opens in
   the Replay view.
3. `practice.wasm`, in a dedicated worker with its own clock. Done when spike 4 passes and `go list -deps ./sim/wasm`
   shows no practice package.
4. The practice server: records, bindings with the default table, CVars, macros, saved edits, assets, attempts, and
   the install folder setting. Done when golden file tests pass for the Lua parser, the WTF parser, the slot table and
   the macro parser.
5. Spikes 1 and 2. Done when we can extract any Forever file by FileDataID, and spike 2's time estimate is written
   into this doc.
6. The 3D scene: floor, dummy, camera, movement, jumping, targeting, with sized capsules. Done when you can walk up to
   the dummy, fight it, and the camera and movement checklist items feel right to you.
7. The model renderer: three-m2loader plus your race's body with its animations, in place of the capsule. Done when
   your character runs, jumps, attacks and casts with the right animations.
8. The HUD: your bars (editable) with bindings, unit frames, cast bar, swing timer, WeakAuras, combat text, the
   Start Encounter and Stop buttons, the countdown, the fight timer, the live meter and the fight summary. Done when a headless Chrome smoke test plays a scripted fight with
   no console errors, the meter's total matches the summary, and the summary shows. The sweep paint cost is measured
   here.
9. Your gear on the model: dressed character, attachments, sheathing, icons and sounds. Done when your character
   fights the dummy with your gear on.
10. The BayovCore probe (a change in the bayovui repo): logs position and facing every frame, the spell cast events,
    error messages and `GetNetStats`. Done when one run of the set path and a short dummy fight are logged. It can
    start as soon as step 6 is in.
11. The feel pass: every "check" in [Playing like WoW](#playing-like-wow) measured with the probe and matched, and the
    feel checklist played through. Movement and camera get measured after step 6, not only at the end.

**Milestone 2: a boss.** A boss that hits back and moves, with simple mechanics (stand still, move out, interrupt) that
change the best rotation. Line of sight and hit box sizes. Totem reach and several enemies (see the known gaps in
[World and movement](#world-and-movement)). Spike 3 runs before it. The boss's movement and mechanics live in
`sim/practice` and use the core's spells and auras.

**Milestone 3: raid bots.** The raid sim's members as bots around you, each on its APL, with raid frames. This uses the
raid sim's whole-raid engine (its milestone M1b in `docs/raid_sim.md`) as it is.

**Milestone 4: friends.** Spike 5 runs before it. See [Multiplayer](#multiplayer-milestone-4).

## Spikes

1. **Reading the `wowf` storage.** Start from TACTTool's prebuilt Windows and Linux releases. Check whether they need
   the .NET runtime, and that Forever-only files (by FileDataID) come out. Open wow.export 0.2.20 on `wowf`, load an
   Orc shaman with gear, and keep its view as the reference. Pass: any file extracts by FileDataID from Go.
2. **A dressed character.** Port wow.export's character code onto three-m2loader for one race and one gear set.
   Measure how long the rest takes, and the frame time with 40 characters on screen. Pass: one dressed character draws
   right next to wow.export's view.
3. **Spell visuals** (before Milestone 2). Size the particle and ribbon work from WebWowViewerCpp's emitters, and
   compare with three.quarks plus hand made effects for the few spells enh uses.
4. **WASM in real time.** Over a 3 minute fight, measure the time from a key press to its event, and from a press to
   the next frame on screen (Event Timing API), the garbage collector pauses, and the time swap at the pull. Pass: key to event under one frame
   (7ms at 144Hz) at p99. On a fail, we cut per-event allocations first, then the log lines we send.
5. **WASM against native** (before Milestone 4). Run the same seed and inputs in the browser and in a native Go test,
   and diff the logs.

## Open questions

None right now.

# Practice room

A realtime 3D room where you play your character against an encounter, driven by the same sim that makes the DPS
Report. You press your own keybinds, and the sim decides every outcome exactly as it does in a sim run.

It's a sim tool first. The point is to see how close real play gets to the sim's optimal rotation, to try a rotation
or a talent choice by hand before using it in game, and later to practice a raid fight with friends. It is not a game
we ship.

**How to read this doc:**

- Decisions marked "(the user, 2026-11-02)" are settled, and [Decisions](#decisions) lists them.
- Everything else is a proposal.
- An implementer starts at [Milestones](#milestones). [Spikes](#spikes) lists what we must check before building on
  it.

## Decisions

1. **Browser app, not a native client** (the user, 2026-11-02). It runs at a local URL from our dev server.
2. **Stack** (the user, 2026-11-02): TypeScript + Vite for the client, Three.js for 3D, our Go sim as the game logic.
   The sim runs as WASM in a Web Worker now, and natively in a Go server once we add multiplayer.
3. **Local only, with real WoW assets** (the user, 2026-11-02). Character models, items, spell icons, sounds and
   visuals come from the WoW Forever client on this PC. They are never deployed, committed or put in `dist/`.
4. **Keybinds and action bars come from the Forever client** (the user, 2026-11-02). Later, other players import
   their own.
5. **First milestone** (the user, 2026-11-02): enh shaman, target dummy, flat arena, your keybinds, a live DPS meter,
   and an end of fight comparison with the sim's optimal rotation.

## Architecture

```
 Browser tab                                          Web Worker
 ┌─────────────────────────────────────┐              ┌──────────────────────────────┐
 │ Input: keys, mouse ──► commands ────┼── commands ─►│ lib.wasm (Go sim)            │
 │                                     │              │  practice session:           │
 │ World: positions, facing, movement ─┼── positions ►│   player unit (your input)   │
 │                                     │              │   other units (their APLs)   │
 │ Three.js scene ◄── events ◄─────────┼── events ◄───│   advance to wall time       │
 │ HTML overlay: bars, frames, MSBT    │              └──────────────────────────────┘
 └─────────────────────────────────────┘
        ▲ assets, keybinds, bars from the dev server (reads the Forever client on D:)
```

Later the worker's box moves into a Go server, and the arrows become a WebSocket. Nothing else changes.

### The sim in real time

The sim is a discrete event loop (`Simulation.Step` in `sim/core/sim.go`). It jumps from one event to the next and
doesn't care about wall time. To run it live, we add a stateful practice session in a new package (`sim/practice`):

- **Start(config)** builds the sim from the same proto config a raid sim uses (player, gear, buffs, encounter), runs
  the prepull and stops at time zero.
- **Input(command)** takes a command from the player at the current sim time: cast a spell, use an item, start or stop
  auto attack, target a unit.
- **AdvanceTo(t)** runs `Step` until the next pending event is later than `t`, then returns the events that happened.
  The client calls it every frame with the wall clock.

Your character gets a player rotation in place of its APL. When the GCD or a cast ends, it casts what you queued, if
anything. We honor the client's `SpellQueueWindow` CVar (read from `config-cache.wtf`), so a key pressed up to that
long before the GCD ends fires the moment it ends, as it does in game. Every other unit (pets, totems, the boss, raid
bots) keeps its APL.

The sim's labeled random streams are seeded, so the seed plus the timed inputs replay a whole attempt exactly. That
gives us three things for free: the Replay view can show any attempt, a bug report is a small file, and multiplayer
can start from the same idea.

The sim only knows one distance per unit (`Unit.DistanceFromTarget`). The world layer owns 3D positions and facing.
Each frame it sends each unit's distance to its target and whether it faces it, and the sim's range checks read
those. Facing and "behind the target" checks need adding to casts and melee.

### Events out of the sim

The Replay and Events views read the sim's text log today. The practice room needs typed events instead: cast start,
cast success, cast interrupted, damage and healing (with hit, crit, miss, glance), aura applied and removed, resource
changes, swing timers and cooldown starts. We add a small event proto, and the practice session collects them per
`AdvanceTo` call. The text log stays for the existing views.

### World and movement

- WoW run speed is 7 yd/s, backpedal 4.5 yd/s. Turning by keys is 180°/s.
- Flat ground, units are cylinders for collision. No physics engine.
- Melee range and spell range checks use the hit box rules the sim already uses.
- Line of sight and pillars come later, with boss mechanics.

### Rendering

- Three.js scene with a WoW style camera: left drag orbits, right drag turns the character, wheel zooms.
- glTF models with their animations: stand, run, attack, cast, channel, hit, death.
- The target dummy and the arena floor are client models too.
- Floating combat text, cast bars and nameplates as HTML over the canvas.

### HTML overlay

The Replay view already draws your EllesmereUI frames, the Bayov WeakAuras class pack and MSBT style combat text. We
reuse those components on top of the canvas: unit frames, cast bar, action bars with cooldown sweeps, buff rows and
the WeakAuras. Plus a live DPS meter and a fight timer.

## Assets (local only)

The Forever client's data sits in the shared install at `D:\Gaming\World of Warcraft\Data` (a CASC archive, the same
storage every modern WoW client uses). The `_forever_` folder holds only the executable and settings.

| What | Source | Format we use |
| --- | --- | --- |
| Race and sex models with your gear on | M2 models, ItemDisplayInfo (wago.tools tables we already pull) | glTF + PNG |
| Weapons, shields | M2 models via ItemDisplayInfo | glTF + PNG |
| Target dummy, arena floor | M2 or WMO models | glTF + PNG |
| Spell and item icons | BLP textures | PNG |
| Spell and swing sounds | SoundKit tables + OGG files | OGG |
| Spell visuals | SpellVisualKit tables + M2 particle models | later (see [Spikes](#spikes)) |

**Extraction tool:** wow.export (open source, github.com/Kruithne/wow.export). It reads a local CASC install and
exports M2 models to glTF with animations, and BLP textures to PNG. A script in `tools/practice/assets/` drives the
export for the list of models a session needs, from the race, sex, gear and spells in its config.

**Where they go:** `practice_assets/` at the repo root, in `.gitignore`. The dev server serves it at `/practice/assets/`.
Nothing under `dist/` or the DPS Report build reads it, and the deploy script refuses to run if it finds a path from it
in the output.

**Friends:** since we never send assets anywhere, each friend runs the same extraction against their own WoW install.

## Keybinds and action bars

The client splits this across three places:

1. **Key to button:** `bindings-cache.wtf`, lines like `bind SHIFT-R ACTIONBUTTON7`. There is one per account
   (`_forever_\WTF\Account\<account>\`) and an optional one per character
   (`...\<account>\<realm id>\<character>\`). The character's file wins when it exists.
2. **Button to spell:** the server keeps what sits in each action slot, and no WTF file has it. BayovCore's CharSetup
   module already records every slot of every character (spell by name and rank, macro by name and body, item by id)
   in BayovCoreDB's SavedVariables. We read that record. Other players will need a small export, either BayovCore's
   or a tiny standalone addon.
3. **Macros:** `macros-cache.txt`, one per account and one per character. Milestone 1 supports `/cast`,
   `/castsequence`, `/startattack` and the `[mod:shift]`, `[@target]` style conditions. A macro with anything else
   shows an "unsupported" badge on its button.

Button names map to slots with the client's table, which isn't in order: `ACTIONBUTTON` is slots 1-12,
`MULTIACTIONBAR1` is 61-72, bar 2 is 49-60, bar 3 is 25-36, bar 4 is 37-48 and bars 5 to 7 are 145-180. Spike 5
checks this against BayovCore's record. Movement bindings (`MOVEFORWARD`, `TURNLEFT`, `STRAFELEFT`, `JUMP`)
drive the world layer. Bindings for things the room doesn't have (bags, map, `CLICK BayovFishButton`) are ignored.

The browser can't read `D:\`, so the Go dev server reads these files and serves them as one JSON. A character picker
lists what it finds under `WTF\Account`. It never assumes a character name or a fixed game folder.

**Browser limits:** the browser keeps some keys for itself (Ctrl+W, Ctrl+T, Ctrl+N, F5). In fullscreen, Chrome's
Keyboard Lock API hands us almost all of them, and Pointer Lock keeps the mouse inside the window while you turn.
Mouse buttons 4 and 5 normally go back and forward in history, so we block that.

## Comparing with the sim

At the end of a fight the room shows:

- **Your DPS against the sim's distribution** for the same character, gear, buffs and fight length. You see where you
  land (for example "71st percentile"), because procs make one exact number meaningless.
- **Where the gap came from:** GCD time spent idle, casts the APL would have made that you didn't, buffs and debuffs
  that fell off, time out of melee range, and resource capped (energy, mana) time.
- **A replay** of your attempt in the Replay view, next to a sim iteration with the same seed.

## Milestones

**M1: you against a dummy.**

1. Spikes 1 and 2 below.
2. `sim/practice` session with Start, Input and AdvanceTo, a player rotation and the typed events. Go tests drive it
   with scripted inputs and check that a script of the APL's own casts gives the APL's DPS.
3. WASM entry points for the session, run in a Web Worker.
4. Dev server endpoints: keybinds, bars and macros for a picked character, and the asset folder.
5. Asset script for the enh shaman's race, gear, dummy, icons and sounds.
6. Client: Three.js scene, camera, movement, models with animations, HTML overlay with your bars, frames and DPS meter.
7. End of fight comparison.

**M2: a boss.** A boss that hits back and moves, with simple mechanics (stand still, move out, interrupt) that change
what the best rotation is. Line of sight.

**M3: raid bots.** The raid sim's members as bots around you, each on its APL, with raid frames. Solo raid practice,
no network.

**M4: friends.** A Go server runs the session natively and decides every outcome. Clients send their inputs and get
events and positions back over WebSocket, at a fixed tick of about 20 to 30 per second. The server runs on this PC,
and friends reach it through a tunnel such as Tailscale. Each friend runs the client and the asset extraction on their
own PC.

## Spikes

Small checks to run before building on the answer:

1. **Does wow.export open the Forever product?** The shared install lists several products in `.build.info`. If it
   can't, fallbacks are CascLib with our own M2 to glTF step (a big job) or Wowhead's online model viewer (no local
   data, but needs the internet and has only what Wowhead has).
2. **Can we get a character model with gear on?** A dressed character is a base body plus geosets (the parts shown or
   hidden by each item) plus textures painted onto the body per armor slot. If wow.export can't export a dressed
   character, we assemble it ourselves in the client from the parts. That could be most of M1's asset work, so we
   size it first.
3. **Spell visuals.** M2 particle emitters don't survive a glTF export. For M1 we show animations, icons, sounds and
   combat text, and no particles. Later we either write a small particle player for the M2 data or hand make a few
   simple effects.
4. **WASM in real time.** `lib.wasm` is 39 MB with every item in it. That's fine to load from a local server, but we
   check that one `AdvanceTo` per frame stays well under a millisecond and that Go's garbage collector doesn't cause
   frame stutter.
5. **Bar slot numbers.** Check the button to slot table above against BayovCore's slot record and your bindings, on
   one character.

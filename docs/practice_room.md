# Practice room

A realtime room where you play your character against an encounter, driven by the same sim that makes the DPS Report.
You press your own keybinds, and the sim decides every outcome exactly as it does in a sim run.

It's a sim tool first. The point is to see how close real play gets to the sim's rotation, to try a rotation or a
talent choice by hand before using it in game, and later to practice a raid fight with friends. It is not a game we
ship.

**How to read this doc:**

- Decisions marked "(the user, 2026-11-02)" are settled, and [Decisions](#decisions) lists them.
- Everything else is a proposal. [Open questions](#open-questions) lists what the user still has to decide.
- An implementer starts at [Milestones](#milestones). [Spikes](#spikes) lists what we check before building on it.

**History:** first drafted 2026-11-02. Reviewed the same day by five reviewers (the sim engine, the WoW assets, the
keybinds and bars against the real client files, the client and networking, and scope), and rewritten from their
findings.

## What it answers

- **How much of the sim's DPS you really get.** Your attempt next to the rotation's, with the gap split into causes.
- **Whether a rotation is worth it at your reaction time and ping.** For example, whether the full enh priority beats
  a simpler one when you react 300ms late.
- **How a talent, gear or rotation choice feels by hand** before you respec in game.
- **Later: a raid fight**, with bots or friends in the other spots.

## Decisions

1. **Browser app, not a native client** (the user, 2026-11-02). It runs at a local URL from our dev server.
2. **Stack** (the user, 2026-11-02): TypeScript + Vite for the client, Three.js for 3D, our Go sim as the game logic.
   The sim runs as WASM in a Web Worker now, and natively in a Go server once we add multiplayer.
3. **Local only, with real WoW assets** (the user, 2026-11-02). Character models, items, spell icons, sounds and
   visuals come from the WoW Forever client on this PC. They are never deployed, committed or put in `dist/`.
4. **Keybinds and action bars come from the Forever client** (the user, 2026-11-02). Later, other players import
   their own.
5. **First milestone** (the user, 2026-11-02): enh shaman, target dummy, flat arena, your keybinds, a live DPS meter,
   and an end of fight comparison with the sim's rotation. [Milestones](#milestones) proposes splitting it in two.

## Architecture

```
 Browser tab (main thread)                           Practice worker
 ┌──────────────────────────────────────┐            ┌─────────────────────────────────┐
 │ Input: keydown ──► input {id, time} ─┼── input ──►│ lib.wasm (Go)                   │
 │                                      │            │  practice session               │
 │ HUD state store ◄── events ◄─────────┼── events ◄─│   sim: your unit reads inputs,  │
 │   (predicts GCD and cast bar)        │            │        every other unit its APL │
 │ Three.js scene, HTML overlay         │◄─ units ───│   world: positions, facing,     │
 └──────────────────────────────────────┘            │          range, movement        │
   ▲ assets, keybinds, bars from the dev server      │   own clock, pushes events      │
                                                     └─────────────────────────────────┘
```

The sim and the world model both live in Go. The browser sends only your inputs and your own movement, and gets every
unit's state back. That keeps one owner for positions when the boss moves (M2), bots stand around you (M3) and a
server judges range for several players (M4).

### Your unit's rotation

The engine already has `SimOptions.Interactive` (`proto/api.proto:168`) and a session API in `sim/lib/library.go`.
We don't use it. The flag is global: it stops the rotation of every character, pets included
(`sim/core/character.go:525`), and several places call the rotation without a nil check (`cast.go:368`,
`attack.go:362`, `energy.go:166`, `rage.go:440`).

Instead your unit gets an APL with one custom action that reads the input queue, in the pattern of
`APLActionCustomRotation` (`sim/core/apl_actions_misc.go:283`). The APL's prepull list stays, so totems and buffs
before the pull work as they do in a sim run. Every other unit keeps its own APL.

### The sim in real time

The sim is a discrete event loop (`Simulation.Step` in `sim/core/sim.go:756`). A new package `sim/practice` holds a
stateful session:

- **Start(config)** builds the sim from the same proto config a sim run uses, with a fixed fight length and no
  duration variation, runs the prepull, and stops at time 0. Execute phases, `GetRemainingDuration` and the end check
  all depend on that fixed length. Auto attacks start on their own at 0, as in a sim run.
- **Input(input)** queues a key press as a pending action at an exact sim time in whole nanoseconds, with a fixed
  lowest priority. It goes through `Step` like any other event, so `sim.events` counts it and the other units' idle
  check skipping (`apl_idle.go`) sees it. A key pressed while the GCD is idle, or an off-GCD spell, fires at its own
  time, not at the APL's next 50ms poll.
- **Run(until)** runs events whose time is before `until` and stops. It never moves the clock to `until` itself.
  Moving the clock to frame times would settle regen and expire auras in frame sized steps, and then results depend on
  frame rate.

`Step` handles one event and has no way to peek, so `Run` reads the next event time as the smallest of the top pending
action, `minWeaponAttackTime` and `minTaskTime`.

The worker keeps its own clock and pushes events as they happen, instead of waiting for one call per frame. The clock
pauses when the tab is hidden or loses focus, and one catch-up step is capped at 0.25s, as the Replay view does
(`ui/core/components/detailed_results/replay.tsx:963`). The end of the fight is final: `Step` drops the last pending
action before the end check.

Between events the HUD shows values like energy and the cast bar by computing them for the display time, without
changing sim state. The live DPS meter sums spell metrics as `getDamageDone` does (`sim/lib/library.go:282`), because
unit metrics only add up in `Cleanup`.

The worker is its own dedicated worker, not one from the sim's `WorkerPool` (`ui/core/sim.ts:123`). The WASM entry
points are synchronous `js.FuncOf` calls like `raidSim` (`sim/wasm/main.go:133`). The live heap is about 1.3 MB, so
garbage collector pauses come from per-event allocations (log formatting, proto marshalling), not from the database.

### Inputs and prediction

Each input carries an id, the key's `event.timeStamp` mapped to the worker's clock through `performance.timeOrigin`,
and the action. Each event carries its sim time and, when it answers an input, that input's id.

The client predicts what WoW predicts: the button press, the GCD spin and the cast bar start. A cast the sim rejects
rolls them back with the game's error text ("Not enough mana", "Out of range"). Resources and outcomes always come
from the sim. In M1 the worker answers within a frame, but building it this way now means M4 isn't a rewrite.

A latency setting (your in-game ping) delays inputs before they reach the sim. Without it the room plays better than
the game does.

### Replays

An attempt is the config, the seed actually used (seed 0 means "pick from the clock", `sim.go:236`), and every input
with the sim time it was applied, movement included. A replay mode feeds those inputs into a normal run, which also
gives the text log the Replay view reads. That makes a bug report a small file.

Exact replays hold only on the same sim build and the same platform. WASM and native Go can differ in the last bits of
`math.Exp` and `math.Log`, so a native M4 server and a browser may drift apart. Spike 5 checks how much.

### World and movement

We copy the leveling world (`sim/core/leveling_world.go`), not `Unit.MoveTo`. The leveling world has 2D positions
with velocity, keeps auto attacks going while moving, interrupts a hard cast when movement starts
(`leveling_world.go:197`), and has a distance function for range. `MoveTo`'s movement aura cancels auto attacks
(`movement.go:107`), which is wrong for strafing in melee.

Your movement goes to the sim as timestamped velocity and facing changes, not as per-frame positions. That keeps
replays exact and lets idle check skipping work out when a unit comes into range. The client moves your model at once
and the sim agrees, as WoW trusts the client's movement.

What the engine has today and what we add:

- **Melee range** is a flat 5 yards (`constants.go:15`). We keep that for M1 and add hit box sizes with bosses.
- **Spell range** is only enforced in leveling (`spell.go:548`). We turn it on for practice.
- **Auto attack range** is only checked when a swing is enabled (`attack.go:755`). The world calls the same cancel
  and enable pair that PvP downtime uses (`pvp.go:68`) when you leave or enter range.
- **Facing** exists as `PseudoStats.InFrontOfTarget` (`character.go:246`), read at each roll for parry and block. We
  add the facing check for casts and melee, and the behind check.
- Run speed 7 yd/s, backpedal 4.5 yd/s. Line of sight comes with bosses in M2.

### Events out of the sim

The worker streams the sim's text log lines with each batch, and the client reuses `logs_parser.tsx`, which already
turns them into typed records (casts, damage, auras, resources). One source of truth, and a finished attempt opens in
the Replay view as is. Typed events come later, only if parsing turns out too slow.

The log needs a few more lines for the live HUD: the next swing time and the cast end time. The Replay view gets those
by looking ahead in the finished log, which a live stream can't do.

### HUD overlay

The Replay view draws your EllesmereUI frames, the Bayov WeakAuras class pack and MSBT style combat text. Its model is
built once from a finished fight (`buildReplayModel`, `replay_model.ts:327`), and parts of it look ahead: swing timers,
cast pushback, and the bar layout chosen from cooldowns used in the fight. The draw functions are "state at time t"
lookups, so a model that only grows forward in time can feed them.

So we reuse the markup, CSS and helpers (`updateIcon` and `updateCastBar` in `replay.tsx`), and write an incremental
model fed by events. Your real action bars are new code.

### Rendering

- **Three.js.** It handles 40 skinned characters (about 800 draw calls) on a desktop. It has no particle system, so
  particles come later from three.quarks or our own M2 particle code.
- **WoW models drawn directly**, not converted to glTF (see [Assets](#assets-local-only)). Weapons and helms hang on
  M2 attachment points: hands 1 and 2, spell palms 21 and 22, back sheaths 26 to 28, hip sheaths 32 and 33. Weapons
  are sheathed out of combat (where, from the item's sheathe type) and drawn in combat.
- **Animations by AnimationData id:** stand 0, run 5, one hand attack 17, ready 26, precast 31, cast 32, channel 76
  and 77. Attacking while running blends the upper body over the run. Some spells name their own animation through
  their visual kit.
- **Camera from your CVars:** turn speed (`cameraYawMoveSpeed`), max zoom, smoothing. Right drag turns you, left drag
  orbits, both buttons run forward.
- **Targeting:** click picks a unit's cylinder, not its mesh. Tab picks the nearest enemy in front, then cycles.
  `@mouseover` is the unit under the cursor. Nameplates are HTML placed from 3D positions.
- **Frame pacing:** movement runs at a fixed step and the screen interpolates at 60 to 144Hz.
- **Audio** starts on the "Enter fullscreen" click, since browsers block it before a user gesture. Sound files are
  decoded ahead of time.
- **Loading:** `instantiateStreaming` so Chrome caches the compiled 39 MB wasm.

## Assets (local only)

The Forever client's data sits in the shared CASC storage (WoW's archive format) of the WoW install. Its product code
in `.build.info` is `wowf`, version 1.60.1.70707 today. The install folder is a setting. On this PC the dev server
reads it from WSL at `/mnt/d/Gaming/World of Warcraft`.

**The pipeline:**

1. The dev server extracts raw files by FileDataID (the client's file number): M2 models, `.skin`, `.skel` and `.anim`
   files, BLP textures, OGG sounds and DB2 tables. It uses a command-line CASC reader (TACTTool from TACTSharp, or
   CascLib, picked in spike 1) and caches the files in `practice_assets/`.
2. DB2 tables come from the local client, plus its hotfix cache (`DBCache.bin`), not from wago.tools. That way each
   friend matches their own client build and needs no internet. The tables we pull from wago today are only the spell
   and talent ones, and they are two builds behind the client.
3. The browser runs our own Three.js M2 renderer. It starts from three-m2loader (MIT, reads the format, skins and
   animations) and ports wow.export's character assembly (MIT, about 4 to 5 thousand lines of browser JavaScript).

wow.export added Forever support in 0.2.20 (2026-11-02). It has no command line or remote control, so it can't drive
the pipeline, and its glTF export leaves weapons unattached and drops particles. We use its GUI to check our output.

**A dressed character**, which is the biggest asset job in M1 (days of porting, one to two weeks from scratch):

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

**Spell visuals and imbue glows** are mostly M2 particles and ribbons. M1 has animations, icons, sounds and combat
text, and no particles. Later the same M2 renderer gets particles, with WebWowViewerCpp's emitters as the reference
(read only, it has no license file).

**Keeping it local:** `practice_assets/` is in `.gitignore`. The dev server serves it at `/practice/assets/`. Nothing
under `dist/` or the DPS Report build reads it, and the deploy script refuses to run if it finds a path from it in the
output.

## Keybinds and action bars

### Bindings

`bindings-cache.wtf` holds only the keys changed from the defaults. The user's file has 150 lines and no
`MOVEFORWARD`, `MOVEBACKWARD`, strafe, jump or tab target lines. Unbound keys show as `bind 7 NONE`. The defaults live
in the client binary, not in FrameXML.

So BayovCore exports the live binding set (`GetNumBindings` and `GetBinding`) into its character record, and that is
the source. The WTF file plus a table of defaults is the fallback for characters with no record.

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

On Forever, every class's stance or form pages bar 1 to slots 73-120. M1 assumes page 1.

### What's in each slot

The server keeps what sits in each slot, and no WTF file has it. BayovCore's CharSetup module records it in
`BayovCoreDB.modules.charsetup.chars[<GUID>]` with `{name, realm, class, race, level, build, slots, macros}`:

- A spell slot is `{kind="spell", id=<spell id of the rank>, name, top}`. The spell id maps straight to a sim action.
- A macro slot has the whole body, so `macros-cache.txt` is only a fallback. Account macros are at
  `charsetup.account.macros`.
- A slot the game hid is `false`.

The file is a plain Lua table, written at logout or `/reload`. A small hand parser reads the 850 KB file in 0.1s.

**Today only one record exists:** Bayov Oom, a level 9 shaman on the beta realm. A character has a record only after
a login with BayovCore on. So:

- The character picker lists records, keyed by GUID, not WTF folders. The WTF folder is named
  `<realm id>/<Name-Name>` and can only be joined to a record by name.
- Spells the sim's rotation uses that sit on no bar go on a generated bar, which you can bind in the room. Bar spells
  the sim doesn't have (Ghost Wolf, heals) show greyed out.
- Buttons can also be clicked.

### Macros

What the user's shaman macros use, which M1 supports:

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
only when changed, so a missing CVar means the default. We read:

- `SpellQueueWindow` (default 400ms, the user hasn't changed it) and `ActionButtonUseKeyDown` (whether a cast fires on
  key down or key up).
- `cameraYawMoveSpeed`, `cameraPitchMoveSpeed`, `mouseSpeed`, `cameraDistanceMaxZoomFactor`, `cameraSmoothStyle`.
- `deselectOnClick`, `autoSelfCast`.

### Browser keys

The browser keeps some keys, and the user binds many of them: Ctrl+T (Fire Nova, opens a tab), F1 to F3, F5 to F7,
F11, F12, and the middle mouse button.

- Chrome or Edge in fullscreen with Keyboard Lock gets all of them. With Esc locked, leaving fullscreen takes a long
  press of Esc, which suits WoW.
- Firefox and Safari don't support Keyboard Lock, so this layout needs Chrome or Edge. The room says so on entry.
- Keys match by `event.code` plus modifiers, and `event.repeat` is ignored.
- Middle click and mouse buttons 4 and 5 have their browser defaults blocked.
- Pointer Lock keeps the mouse inside while you turn. Chrome waits about a second before it can lock again after Esc.

## Comparing with the sim

Proc luck in one 2 to 3 minute attempt is as big as most skill gaps. So the comparison takes luck out:

- **Your plan over many seeds.** Your timed cast list runs as a script over many seeds. Casts that become impossible
  (no mana, on cooldown) are skipped. Its DPS spread next to the rotation's spread is the skill gap without the luck.
- **Your reaction time.** The room measures your delay from GCD ready, and from a proc, to your press. Then it re-sims
  the rotation with `reaction_time_ms` (`proto/api.proto:74`) set to yours, and with your ping. That answers "is the
  full rotation worth it for me".
- **Where the gap came from:** idle GCD time, buff and debuff uptimes, time out of range, swings pushed back by casts,
  and capped resource time.
- **The headline number:** where your DPS lands in the rotation's distribution, for example "71st percentile".

Side by side with "the same seed" means nothing here. Normal runs share one random stream (`sim.go:267`), so once your
casts differ from the rotation's, every later roll differs. A per-GCD "what the rotation would have cast" needs a
re-run from the start up to each press, since the sim can't copy itself mid-fight. That comes after M1.

The end of fight sims run in the normal sim worker pool, not in the practice worker.

### Saving attempts

Each attempt (config, seed, inputs, sim build, result) is saved by the dev server to `practice_attempts/`, which is in
`.gitignore`. The room lists them with a trend over time. When the sim build changes, old attempts keep their result,
and their replay is marked as from another build.

## Where it lives

Its own page, like `ui/raid_sim/`. A "Practice" button in a class sim opens it with that sim's current settings
(character, gear, talents, buffs, rotation, fight length). Presets and pasted BayovCore exports come in the same way,
since they already load into the class sim. The keybinds come from a BayovCore record you pick separately, because the
sim character and the in-game character aren't the same thing.

Fight rules for M1: a fixed length you set, with no length variation, the prepull done by the rotation's prepull list,
a 3 second countdown, and keys for restart and leave.

## Multiplayer (M4)

M4 is more than a new connection. It needs:

- a Go server that runs the session natively, applies each input when it arrives, and sends events right away or in
  small batches (about 50ms)
- a fixed tick (20 to 30 a second) only for position broadcasts, which clients interpolate
- several units with player rotations
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
we never send game files anywhere. The details wait for M4.

## Milestones

Each step says when it's done.

**M1a: the sim loop, no 3D.** Delivers the sim value first.

1. `sim/practice` session: Start, Input, Run, your unit's input APL action, the world copied from leveling, range and
   facing. Done when Go tests with scripted inputs pass, and a script of the rotation's own casts, fed at the
   rotation's exact sim times, gives the rotation's mean DPS over 1000 seeds within noise.
2. Replays and saved attempts. Done when a replayed attempt gives an identical log on the same build.
3. The comparison: your plan over many seeds, reaction time, gap causes. Done when a Go test with a scripted slow
   player shows the expected gap.
4. WASM entry points in a dedicated worker with its own clock. Done when spike 4's numbers are in.
5. Dev server endpoints: BayovCore records, bindings, CVars, macros, and the install folder setting. Done when golden
   file tests pass for the Lua parser, the WTF parser, the slot table and the macro parser.
6. The page: your bars with bindings, unit frames, cast bar, WeakAuras and combat text as an incremental HUD over a 2D
   top-down view (`replay_map.ts`), the live meter, the comparison screen. Done when a headless Chrome smoke test plays
   a scripted fight.
7. BayovCore exports the live bindings into its record (a bayovui change).

**M1b: 3D.** Spikes 1 to 3 first. Then the M2 renderer, a dressed enh shaman with attachments and sheathing,
animations, the dummy, the arena floor, icons and sounds, the camera and targeting. Done when your character fights
the dummy in 3D with your gear on.

**M2: a boss.** A boss that hits back and moves, with simple mechanics (stand still, move out, interrupt) that change
the best rotation. Line of sight and hit box sizes.

**M3: raid bots.** The raid sim's members as bots around you, each on its APL, with raid frames. This needs the raid
sim's whole-raid engine (its milestone M1b in `docs/raid_sim.md`).

**M4: friends.** See [Multiplayer](#multiplayer-m4).

## Spikes

1. **Reading the `wowf` storage from the command line.** Try TACTTool and CascLib on the local install. Check that
   Forever-only files (by FileDataID) come out. Open wow.export 0.2.20 on `wowf`, load an Orc shaman with gear, and
   keep its view as the reference.
2. **A dressed character.** Port wow.export's character code onto three-m2loader for one race and one gear set.
   Measure how long the rest takes before planning M1b.
3. **Spell visuals.** Size the particle and ribbon work from WebWowViewerCpp's emitters, and compare with three.quarks
   plus hand made effects for the few spells enh uses.
4. **WASM in real time.** Measure the time from a key press to its event at p50 and p99, and the garbage collector
   pauses, over a 3 minute fight.
5. **WASM against native.** Run the same seed and inputs in the browser and in a native Go test, and diff the logs.

## Open questions

1. **Split M1 into M1a (sim loop, 2D) and M1b (3D)?** M1a gives the sim answers in a fraction of the time. 3D and real
   models follow. Recommended: yes.
2. **Which comparison is the headline?** Your plan over many seeds (no luck), your reaction time re-sim, or the
   percentile. Recommended: your plan over many seeds, with the percentile as the one number on top.
3. **Model your ping?** Recommended: yes, a setting with your in-game latency.
4. **Who does the prepull?** The rotation's prepull list (simple, what M1 assumes), or you (needs the sim to take
   actions before time 0).
5. **Fixed length or a dummy with health?** Recommended: fixed length for M1, matching the DPS Report.
6. **BayovCore changes:** export the live bindings (needed), and should it record bars for all characters on login?
   Today only Bayov Oom has a record.
7. **When to start,** given Forever launches 2026-11-04 and the raid sim milestones are in flight.

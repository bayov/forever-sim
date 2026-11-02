# Practice room feel checklist

Play through this after each change to movement, the camera or casting (docs/practice_room.md,
"Checking the feel"). One line per rule. Mark each line `ok`, `off` (with what felt wrong) or `check`
(not measured in game yet). The numbers come from "Playing like WoW".

Run the room with `make practice-run` and open http://127.0.0.1:3434/.

## Movement

- [ ] Running forward feels like 7 yd/s: from the start spot to the dummy (15 yards) takes about 2 seconds.
- [ ] Backpedal is slower (4.5 yd/s).
- [ ] Strafing is as fast as running, and forward plus strafe is no faster.
- [ ] A and D turn at 180°/s: a full turn takes 2 seconds.
- [ ] Holding the right mouse button makes A and D strafe.
- [ ] Both mouse buttons held runs forward.
- [ ] Autorun (Num Lock) runs until forward, back or autorun again. Strafing and turning don't stop it.
- [ ] A jump is about 0.83s in the air, keeps its direction, and you can still turn.
- [ ] You walk through the dummy.
- [ ] Ghost Wolf runs 40% faster.

## Casting and attacking

- [ ] Moving or jumping stops a cast. Turning in place doesn't.
- [ ] A cast time spell on the move fails with "Can't do that while moving".
- [ ] A spell with the dummy behind you fails with "Target needs to be in front of you."
- [ ] Turning away during Lightning Bolt fails it at the end.
- [ ] Out of range shows "Out of range." and the button turns red.
- [ ] A press in the last 400ms of the GCD goes off right at the GCD's end. The last press wins.
- [ ] A press earlier than that fails with "Spell is not ready yet."
- [ ] Auto attack starts off. A right click on the dummy, a melee ability or a `/startattack` macro turns it on.
- [ ] Swings stop when you leave melee range, and the overdue swing lands at once when you come back.
- [ ] No swing lands during a cast, and the cast's end resets the swing.
- [ ] Esc during a cast cancels the cast and keeps the target. A second Esc clears the target.
- [ ] Clicking the ground keeps the target.

## The pull and the fight

- [ ] Nothing happens on its own before you act: no swing, no cast.
- [ ] Your first hit (or Magma Totem's first pulse, or a bolt landing) is 0:00 on the fight timer.
- [ ] Auto attack on in melee range before the pull is the pull, with a swing at once.
- [ ] Start Encounter counts down, and the fight starts at the countdown's end.
- [ ] A hit during the countdown is an early pull, and the HUD says so.
- [ ] Encounter mode ends on its own at the set length. Stop ends Free mode.
- [ ] The meter's damage at the end matches the summary's.

## Camera

- [ ] Left drag orbits the camera without turning you.
- [ ] Right drag turns you to where the camera looks.
- [ ] The cursor comes back where it started after a drag.
- [ ] The wheel zooms. The max distance follows `cameraDistanceMaxZoomFactor`.
- [ ] While you move, the camera swings back behind you (`cameraSmoothStyle`).
- [ ] The field of view looks like the game's at the same window size.

## Keys and bars

- [ ] Your bars and keys are the ones from your BayovCore record and WTF files.
- [ ] Alt, Ctrl and Shift fall back to the key's own binding when the modified key is unbound.
- [ ] `[@mouseover,harm,nodead][]` macros cast at the dummy under the cursor.
- [ ] Edit bars: pick up a spell from the spellbook, click a slot to place it. It's saved.
- [ ] Bind keys: hover a button and press a key. It's saved.

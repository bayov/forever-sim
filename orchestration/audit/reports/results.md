# Results audit, pass 56 (master 06a898751, 2026-10-27)

Full pass on the shared run (58 requests, 20k, seed 1) and the drift check.

- Dashboard vs run: builds.json matches the run on all 58 within 0.1 (report re-sim 4273ee51d: racial pass 10-26, Spellstone Shadow-only,
  Arcane Missiles no longer proc Missile Barrage, Ignite ticks without Arcane Instability/Arcane Power).
- Metas vs run: all 58 metas still carry the old published.dps (finding results-20261027-1, low).
- Moves vs the previous reference: all 58 moved, explained by 500e3a147 / d6fcdd3ef / a52cdcf53 / db6e43f98. Largest: Arcane -83.5 (-8.5%,
  Missile Barrage and Ignite), Frost -35.5 (-3.8%), warlock/balance/hunter about -1 to -1.7%, rogues and Ret about -0.5 to -1%.
- Drift: same as the reference, plus `paladin__level_60_shock_ret` primarySeal Command vs Righteousness (standing paladin drift, now on the
  renamed Shock build); regen sim 925.49 = committed 925.49, no effect. regen.py now strips spaces from the race ("Night Elf").
- Unread fields: only notes.
- Open finding results-20261025-1 (consumes drift) unchanged: Ret Shock regen drift no longer shows potion/flask lines for the shock builds in the drift report; not re-measured.

Top defaults (DPS): Beast Mastery 1133.7, Fire 1089.0, Demonology 1050.2, Fury 975.2, Shadow 949.2, Ret sub Holy 941.8, Ret Shock 925.5,
Enhancement 880.5, Feral 866.8, Shock (deep Holy) 863.6, Balance 841.5, Elemental 772.2, Combat 765.6, Subtlety/Mutilate 714.9.

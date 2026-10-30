# Consumables pass 54 (master d50909a37)

Findings: none.

Checked (334c7dec..d50909a37, 5 watched files):
- presets.ts: only feral Faerie Fire duty comments/tooltips changed (cat casts caster Faerie Fire, no cooldown). No Consumes/raid/debuff objects changed; no mksettings, consumes.go, enchant, gear-set or sim/common changes.
- buffs.go: Blood Pact from party buffs, Furious Howl (24597, 83) in Battle Shout slot, camp buffs, Dark Inspiration (uptime 0.999), per-member Mana Tide, keyed Battle Shout/Howl labels. Raid-sim features, no preset effect.
- debuffs.go: Sunder ramp, Faerie Fire/Dust Cloud rank labels (65/175/285/395/505 at 12/24/36/48/60), JoW/JotC per-caster. Dust Cloud joined the minor armor slot, so scan.sh's "expect 2" count is now 3 (docs/raid_sim.md says 3 members): stale scan expectation, not a finding.
- totem_weapon_buffs.go: cast-totem slot resolver, no value changes.
- assumptions_def.py: ALL_BASE skips pet variants; no buff assumption change.
- scan.sh: all sections above counts empty; totems as ruled. wago latest is 70685, already audited; no new build.

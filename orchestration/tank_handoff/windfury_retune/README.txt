Warrior and bear tank presets re-tuned on the main-hand Windfury Totem rule (2026-11-12, standards 17). Results:
docs/tank_warrior.md and docs/tank_druid.md, "Re-tuned on the main-hand Windfury rule".

Scripts (W in them is the session's scratch dir; set it to where the rotopt/abtest binaries and outputs live):
  opt.sh <warrior|druid> <preset> <out> [flags]   rotopt -template warrior_tank|druid_tank on the six scoring bosses, own weight, 1,000 / 3,000
  kv.sh <class> <preset> <name> <knobs> [talents] a variant request: knobs/<class>_<preset> plus <knobs>
  ab.sh <class> <preset> <name> <iters> <seed>    paired abtest, A = tools/tankbench/req, B = the variant
  apply.sh <class> <preset> [talents]             writes the UI APL from knobs/<class>_<preset> and puts it in the request
knobs/: the final knob strings (later duplicates win); rotopt writes the shipped APLs from them byte for byte.
rotopt/: k1_* knob searches from the shipped knobs, t1_* talent searches on their knobs, final_* the 12 tank presets at
  10,000 (seed 1, rotopt -eval -apl on each request's rotation).
ab/: *_i3000_s1 screens, *_i10000_s{1,2} confirmations (B - A against the request before the change); druid_bestVfr80*
  B = FR 80 + Swipe 80 + Enrage 50 against FR 80 (+ Swipe 80); druid_oldVnew_ms_* the final Max survival against the old.

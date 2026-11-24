Prot warrior racials (2026-11-24, branch prot-racials-1124). Results: docs/tank_warrior.md, "Racials (2026-11-24)".
Scripts (W=/tmp/claude-0/pr held the rotopt binaries: rotopt = before Elune's Light / Eureka! lines, new/rotopt after;
k/warrior_<preset> = orchestration/tank_handoff/windfury_retune/knobs/warrior_<preset>):
  opt.sh / opt2.sh <class> <preset> <out> [flags]  rotopt -template warrior_tank, six scoring bosses, own weight (old / new binary)
  apply.sh <class> <preset>                         writes the UI APL from the knobs and puts it in the tankbench request
  grid.sh <iters> <seed>                            the racial knob grid for Orc and Troll -> grid_i<iters>_s<seed>.tsv
  races.sh <iters> <seed>                           every warrior race at each preset's knobs -> races_i<iters>_s<seed>.tsv
ne_gnome.tsv: Night Elf and Gnome before (old: no Elune's Light / Eureka! line), on cooldown, held 60 s, delayed 5 s.

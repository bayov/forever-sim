Round 3b (2026-11-04): paired A/B of the Protection warrior shield enchant, 1886 (Enchant Shield - Greater Stamina, +9 Stamina; A, the
final presets) vs 863 (Lesser Block; B), at each tank preset. Same method as round3 (README.txt, ab.sh, mkvar.py copied here).
Binaries: go build -tags=with_db -o $BIN/ ./tools/rotopt ./tools/tankbench ./tools/audit/tank_rotations/abtest
(proto Go code from grpcio-tools' protoc 1.85.1 + protoc-gen-go v1.36.6).
Variants: python3 mkvar.py ../round3/req/final/F_<preset>.json <scratch>/lb863_<preset>.json 15=:863   (off hand, index 15; only that line differs)
Runs (B - A, six scoring bosses + Ragnaros row as abtest prints; scoring_mean is abtest's own):
  BIN=<bin> ./ab.sh ab10_lb863_<preset>_s<N>.txt 10000 <N> <preset> ../round3/req/final/F_<preset>.json <scratch>/lb863_<preset>.json
  presets max_threat balanced max_survival, seeds 1 and 2 (disjoint iterations).

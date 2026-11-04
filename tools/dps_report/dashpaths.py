"""dashpaths.py: where the dashboard pipeline reads and writes. Every script imports these.

DASH_DIR   base directory: req/ (inputs), forever_tt_cache/, and the outputs out/, rv/, abl/, weights/,
           names/, icons/, data/. Default: this directory (tools/dps_report).
DASH_BIN   directory holding the binaries built from the merged tree: wowsimcli, dashstats (and
           wowsimcli_r, optional). Default: $DASH_DIR/bin.
DASH_TREE  the merged local tree every sim runs in (cwd for the binaries, source of db.json, talents,
           racials, ui/ presets ...). Default: a `w-dash` checkout next to this repository.
"""
import os

HERE = os.path.dirname(os.path.abspath(__file__))
D = os.path.abspath(os.environ.get('DASH_DIR') or HERE)
BIN = os.path.abspath(os.environ.get('DASH_BIN') or os.path.join(D, 'bin'))
# The sim options every report sim adds to its simOptions. useLabeledRands gives each kind of roll (a hit, a crit, a
# proc, ...) its own random stream, so two sims of the same seeds that differ in one setting (a race, a racial, a stat,
# a buff, an assumption) still share every other roll. Measured 2026-11-04 on five classes: the paired spread of a race
# against race comparison fell about 3x (N for the same error about 9x) for about 24% more time per fight. Fight i is
# still seed + i, so a top-up run from seed + N joins a shorter run bit for bit.
RNG = {'useLabeledRands': True}
TREE = os.path.abspath(os.environ.get('DASH_TREE') or os.path.join(HERE, '..', '..', '..', 'w-dash'))

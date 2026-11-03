# Healing priest rotation tuning (work in progress)

Scripts the priest worker tunes its three rotations with (run from the repo root):

- `apl.py`: a builder for the healing APL JSON. `rots.py`: the Holy tank, Discipline tank and Holy raid
  rotations with knobs. `mk.py OUT ROT holy|disc [knob=value ...]` writes a healbench request on the
  synthetic set; `mkmany.py` several. `bench.sh ROLE ITERS "bosses" reqs...` runs healbench -carry on each
  boss (extra flags in `$HB`, e.g. `HB="-dur 300"`). `hbpatch.py` adds a per-spell `-breakdown` to a private
  copy of tools/healbench/main.go.
- Talents: holy `00500503113-03505103030121531`, disc `005005031330011531-0350510303`, Human.

Results so far (500 iterations, tank role at the files' lengths; carry):

| Request | Onyxia | Broodlord | Magmadar | Ragnaros | Onyxia 180 s | Onyxia 300 s |
|---|---|---|---|---|---|---|
| Holy tank, base (Renew, PoM, GH 5/4/2, FH 7) | 1.150 | 0.828 | 1.719 | 0.534 | 1.081 | 0.956 |
| Holy tank `fhfirst=1 pws=1 fh=7,4 emergency=0.5 gate=0.7` (pick) | 1.242 | 0.866 | 2.075 | 0.578 | 1.095 | 0.904 |
| Disc tank, base (PW:S, Penance, Renew, GH 5/4/2, FH 7) | 1.172 | 0.803 | 1.781 | 0.547 | - | 0.781 |
| Disc tank `fhfirst=1 fh=7,4 emergency=0.5 gate=0.7` | 1.203 | 0.876 | 1.938 | 0.563 | - | 0.696 |
| Restoration shaman tank healer | 1.017 | 0.619 | 1.531 | - | - | 0.958 |

Disc runs out of mana at 300 s (PW:S on every Weakened Soul gap): next, a mana gate on PW:S (`pwsgate`,
`pwsneed` knobs in rots.py, not yet benched). Holy raid (300 iterations, Onyxia / Broodlord):
`pws=0.6 order=fh,poh,renew,gh renewshare=0.6 fh=7,5,3` 0.730 / 0.906 against the shaman's 0.703 / 0.828.

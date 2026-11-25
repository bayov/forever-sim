# Results pass 104 (master 5536db4bc)

Range 226ada86b..5536db4bc. Watched paths: only sim/compsearch/drops.json changed (regenerated after the trinket-line fix,
86fbcf456). It is the compsearch cache's drop table (embedded in normalize.go, read by the gear search), not read by the
DPS report requests or the sim run. Other commits: DPS Report assumption data, Phase 0 gold brief, healer fairness ruling
(docs only for this pass).

Shared run (68 builds, 20k, seed 1) vs published: 0 of 68 moved more than 0.5. No sim, proto, mksettings, database or
request change, so drift and unread fields are unchanged. Quiet.

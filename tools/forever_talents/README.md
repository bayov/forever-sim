# Forever talent data

Talent trees for all nine classes, used to generate `ui/core/talents/trees/*.json` and the
talents messages in `proto/*.proto`.

## Where it comes from

Two sources. `client/` is the beta client's trees, generated from the client's own game
tables at one build by `import_db2.py` (see below), and `data/` is what the importer reads.
For Rogue, Shaman and Paladin `data/` is generated from `client/` with `import_client.py`, so
every rank is the client's. The other six classes still carry the
BlizzCon transcription described below, and converting one of them means running
`import_client.py <class>`, diffing the result against the sim code (the ranks 2+ guesses
the class was implemented with are usually wrong somewhere), then `import_talents.py`.

### The client trees

`client/<class>.json` is one class's tree as the client has it: each talent's place, rank
count and prerequisite, and one tooltip per rank with the numbers filled in, plus the cost
line for actives (`tooltip.source` is `client`, with the build number and spell id per
talent). A few values the tooltips ask for aren't in the client tables and show as `?`, with
`tooltip.source` set to `client-partial` and a note saying what is missing (the Riposte
disarm duration, an effect Amplify Curse lost in 1.60.1.70450). `overrides/` keeps Classic's
number for the ones the sim needs.

These files used to be put together by hand from hyjal.cc's talent calculator,
foreverchanges.pro and wago.tools tables at several builds. They are generated now.

### Regenerating client/ from the client tables

	tools/forever_talents/import_db2.py 1.60.1.70450                 # what would change, per class
	tools/forever_talents/import_db2.py 1.60.1.70450 shaman --write  # rewrite client/shaman.json

It reads the trait tables (TraitTree, TraitNode, TraitNodeEntry, TraitDefinition,
TraitDefinitionEffectPoints and CurvePoint for the per-rank values, TraitEdge for
prerequisites) and TalentTab and SkillLine for the tree names, backgrounds and icons, through
`tools/sources/wago.py`, which downloads any table the shared cache doesn't have yet (the
`talents` and `spelltext` sets). The tooltips, cost lines and icons come from
`tools/sources/spelltext.py`, which fills the spell description templates the way the game
does, for a level 60 character with no bonuses, and notes every value it couldn't fill.

A server hotfix the client tables don't have yet is carried in `client/` by hand, with the
hotfix named in the file's `source` and in the talents' `tooltip.notes` (`tooltip.source` is
`hotfix`). The mage's Frost tree carried the 2026-10-20 hotfix to 70450 that way (Piercing Ice
removed, four talents moved, Elemental Precision 3 ranks), from wago.tools' hotfix table, until
1.60.1.70707 put it in the client tables. Running `import_db2.py <class> --write` on a build
without a hotfix the file carries puts the old tree back, so check the diff.

Without `--write` it prints each class's differences from the current file in three kinds:
structure (talents added, removed, moved, renamed, or given another rank count, prerequisite
or node ID), numbers (a rank's numbers differ), and wording (the same numbers in other words).
Read that before writing, since a structure change usually means the sim's tree needs work.

The tables don't know a few things the files carry, and the importer keeps them from the
current file: each tree's `removed` list (talents earlier builds had) and the hand-written
notes in `tooltip.notes`. The file's `source` says when it did. Writing `client/` changes
nothing in the sim by itself: `import_client.py` and `import_talents.py` still have to be run
for a class, which is a separate decision.

### The BlizzCon transcription

`data/` for the unconverted classes and `overrides/` are vendored from
[Deradon/wow-forever-talent-calc](https://github.com/Deradon/wow-forever-talent-calc) (MIT,
copy of its licence in `LICENSE.upstream`). That project extracts every talent tooltip from
the BlizzCon 2026 day 1 stream by hovering each talent on camera, reading the frame with a
vision model and reviewing the result by hand.

This is pre-beta data and it shows. Read it with the caveats:

- Only rank 1 of each talent was ever displayed. Every other rank is either copied from the
  Classic talent of the same name or extrapolated linearly. `ranksSource` on each talent says
  which: `observed`, `classic-prior`, `extrapolated` or `manual`.
- Tooltip text is machine-read. The upstream audit found no errors in high confidence records
  but most of its low confidence queue was wrong, so check `source.confidence` before trusting
  a description. `overrides/` holds the hand corrections and the importer applies them.
- Rows, columns and prerequisites come from the tree layout rather than the tooltip text and
  have held up against independent transcription of the Warlock tree.


## Talents whose per-rank scaling is guesswork

58 of the 469 talents, most of them carrying `ranksSource: "manual"`, list rank 1's numbers again
for every other rank. The extra points are not free in game, so whatever the sim does with
them is invented: implementing them flat makes points 2+ inert, implementing them linearly
assumes a scale nobody observed. Either way it is a guess, and the classes were converted
before this was measured, so they do not all guess the same way.

	tools/forever_talents/import_talents.py --unranked

prints the list, per class. The importer also warns on stderr about the class it is importing,
so regenerating a tree after the beta datamine shows immediately whether the gap has closed.

The count by class today is Warlock 11, Druid 8, Hunter 8, Mage 5, Priest 5, Warrior 4
(Rogue, Shaman and Paladin are on the client data). These are the first thing to re-check against the beta client,
because a five rank talent read from one rank is the largest single source of error in the
data.

Some of them can be closed before the beta. The write-ups that went up after BlizzCon quote
full rank lists for talents the stream only showed at rank 1, and where one of those agrees
with what the sim already assumed it is worth recording: put the ranks in `overrides/` with
the source in `reason`, and the talent drops off the list above. Bloodthrill and Dual Wield
Specialization went that way. Treat a single write-up as corroboration of an assumption, not
as a datamine, and leave the rest alone until there is a client to read.

## Regenerating a class

	tools/forever_talents/import_client.py rogue                # client/ to data/, client classes only
	tools/forever_talents/import_talents.py warlock            # print the proto message
	tools/forever_talents/import_talents.py warlock --write    # also rewrite the tree json

The tree json and the proto message have to stay in the same order, because
`FillTalentsProto` maps the nth character of a talent string to proto field number n. The
importer emits both in (row, column) order so they line up by construction. Changing the
order invalidates every saved talent string, so regenerate both together and rerun the tests.

Spell ids for talents that already exist in the tree json are preserved, so regenerating
doesn't churn the icons. New talents that have no Classic equivalent get a placeholder and
show the wrong tooltip until the beta ids are known.

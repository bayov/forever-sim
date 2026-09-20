# Forever talent data

Talent trees for all nine classes, used to generate `ui/core/talents/trees/*.json` and the
talents messages in `proto/*.proto`.

## Where it comes from

Two sources. `client/` is the beta client's trees (build 1.60.1.69876), and `data/` is what
the importer reads. For Rogue, Shaman and Paladin `data/` is generated from `client/` with
`import_client.py`, so every rank is the client's. The other six classes still carry the
BlizzCon transcription described below, and converting one of them means running
`import_client.py <class>`, diffing the result against the sim code (the ranks 2+ guesses
the class was implemented with are usually wrong somewhere), then `import_talents.py`.

### The client trees

`client/<class>.json` is the tree hyjal.cc's talent calculator ships in its page bundle,
read out of the client tables (its `tooltip.source` is `client`, with the build number and
spell id per talent). One tooltip per rank with the numbers filled in, plus the cost line
for actives. Two values were unavailable in the client data and show as `?` (the Riposte
disarm duration, Berserking's duration), `overrides/` keeps Classic's number for those.

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

# Forever talent data

Talent trees for all nine classes, used to generate `ui/core/talents/trees/*.json` and the
talents messages in `proto/*.proto`.

## Where it comes from

`data/` and `overrides/` are vendored from
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

Replace `data/` with the datamined trees once the beta client is out and rerun the importer.

## Regenerating a class

	tools/forever_talents/import_talents.py warlock            # print the proto message
	tools/forever_talents/import_talents.py warlock --write    # also rewrite the tree json

The tree json and the proto message have to stay in the same order, because
`FillTalentsProto` maps the nth character of a talent string to proto field number n. The
importer emits both in (row, column) order so they line up by construction. Changing the
order invalidates every saved talent string, so regenerate both together and rerun the tests.

Spell ids for talents that already exist in the tree json are preserved, so regenerating
doesn't churn the icons. New talents that have no Classic equivalent get a placeholder and
show the wrong tooltip until the beta ids are known.

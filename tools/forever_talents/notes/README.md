# Notes-only talent changes

Talent changes Blizzard announced (the development notes) that neither the client's tables nor
the launch realms' hotfix rows carry yet. `import_db2.py` lays `notes/<class>.json` over the
tree it generates from the client on every run, so a later client import keeps them instead of
wiping them. A class with no file here is the client's tree as it is.

Standards 37 (user ruling of 2026-11-03) says announced values the client and hotfixes don't
carry take the notes' numbers. That is what this folder is for: a class worker who builds a
rework from the notes writes it here, then regenerates the class:

	tools/forever_talents/import_db2.py 1.60.1.70753 rogue            # read what changes
	tools/forever_talents/import_db2.py 1.60.1.70753 rogue --write    # client/rogue.json
	tools/forever_talents/import_client.py rogue                      # data/rogue.json
	tools/forever_talents/import_talents.py rogue --write             # tree json; prints the proto message

## The file

```json
{
  "class": "rogue",
  "entries": [
    {"tree": "Assassination", "op": "remove", "name": "Murder",
     "source": "development notes 2026-11-02 (post 9): Murder has been removed"},
    {"tree": "Assassination", "op": "set", "name": "Lethality",
     "fields": {"max": 2, "desc": {"1": "...10%...", "2": "...20%..."}},
     "source": "development notes 2026-11-02 (post 9): Lethality 2 points, 10/20%"},
    {"tree": "Assassination", "op": "add",
     "talent": {"name": "Blood Splatter", "max": 1, "row": 3, "col": 2, "passive": true,
                "icon": "inv_misc_questionmark",
                "desc": {"1": "Increases the duration of your Rupture ability by 6 sec and your Garrote ability by 9 sec."}},
     "source": "development notes 2026-11-02 (post 9): new 1-point talent in row 3; column guessed"}
  ]
}
```

- `tree` is the tree's name as `client/<class>.json` has it.
- `add` takes a whole talent in the `client/` shape: `name`, `max`, `row` and `col` (1-based, as
  in `client/`), `passive`, `icon`, `desc` (one tooltip per rank, keyed "1", "2", ...), and
  `req` (the prerequisite's name), `reqText` and `cost` when it has them. It gets no node id,
  and its tooltip says source `notes` with the entry's `source` as its note.
- `remove` takes the talent out and adds it to the tree's `removed` list.
- `set` changes the fields named in `fields` (the same keys as `add`). Moving a talent is a `set`
  of `row` and `col`.
- `source` says where the change was announced. Say in it what you had to guess, like a column
  the notes don't give.

Two talents on one square stop the import. The talents are put back in (row, col) order, which
is the order the talent strings read them in, so a moved, added or removed talent changes the
strings of every preset after it in that tree: remap them, as for a client change.

## When the client catches up

Each run prints `notes: ... the entry can go` for an `add` whose talent the client tree now has
(the client's version wins), a `remove` whose talent the client already took out, and a `set`
whose values the client already has. Delete those entries. A `set` the client disagrees with
stays applied and shows up in the importer's diff, which is the signal to check which one is
right.

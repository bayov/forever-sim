---
name: cm
description: Commit all local changes with jj, in commits grouped by topic, and push master to origin (github.com/bayov/forever-sim). Use it when the user runs /cm or says "commit" or "push".
---

# Commit and push all local changes

We work in one jj working copy, often with more than one agent at a time. So the
working copy usually holds several unrelated pieces of work: a UI change, a regear with
its new presets, a DB regen, notes. We split them into commits by topic, then push
`master` to origin.

The workspace root is `/home/bayov/projects/forever-sim`. Run every command from there.
Use jj, not git.

## 1. See what changed

```bash
cd /home/bayov/projects/forever-sim
jj diff --stat | tail -60
jj diff --summary
```

Look for added files that don't belong in the repo: build output, binaries, big logs,
rotopt result dumps or scratch scripts. jj tracks every file that isn't ignored. Ask the
user before committing one of those. They usually want it deleted or added to
`.gitignore` instead.

## 2. Commit by topic

Give each piece of work its own commit with a real message, by listing its paths:

```bash
jj commit -m "Move custom presets into a Custom folder with a save modal" \
	ui/core/components/saved_data_manager.tsx ui/core/components/preset_tree.tsx \
	ui/scss/core/components/_saved_data_manager.scss
```

Some files go with the work that produced them:
- The generated item DB (`assets/database/*`) goes with the `assets/db_inputs/*` or
  `tools/database` change that regenerated it.
- A regear's `ui/*/gear_sets/*.json`, `ui/*/apls/*.apl.json` and `presets.ts` changes go
  together, with the `forever-wiki/` or `notes.md` lines that record the result.

When one file holds two topics, don't try to split it. Put it in the commit it fits
best.

Then commit everything that's left, but only when something is left. When
`jj diff --stat` prints "0 files changed", skip this step, or we push an empty commit.

```bash
jj commit -m "Update notes and wiki"
```

Messages are one short line in the imperative, with plain ASCII punctuation (no em
dashes or semicolons). Say what changed in domain words, like "Give the Level 30 PvP
preset the High HP gear" or "Fix Healing Stream ticks under Forever". Look at
`jj log --limit 15` for the style. Never use a placeholder like `c` or `wip`.

## 3. Push

```bash
jj bookmark set master -r @-
jj git push --remote origin --bookmark master
jj bookmark list --all master
```

The push is done when `master` and `master@origin` show the same commit. Ignore
`master@upstream`, that's the wowsims repo we forked, and we never push there.

When jj refuses to move `master` backwards or sideways, stop and tell the user. Don't
pass `--allow-backwards`. When GitHub rejects the push (a file over 100 MB, for
example), tell the user what it said.

## 4. Report

Tell the user each commit's ID and message, and that origin's `master` matches.

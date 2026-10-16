#!/bin/sh
# Copy the WoW combat logs into ~/forever-logs/<flavor>/, so they outlive the game install.
#
# The client never deletes its combat logs, but uninstalling the beta (or the launcher removing
# _classic_beta_ when the beta closes) takes the whole Logs folder with it. We copy each
# WoWCombatLog-*.txt that is new or has grown since the last run, and keep our copy even when the
# game's copy is gone.
#
# logcheck -audit compresses a log once it's a day old and fully audited. The game still has the plain one, so we skip a log we hold as .gz.
#
# Only logcheck runs this, before it reads a log, so the newest lines are in the archive when
# we look at them. The log-audit skill's runs (by hand and hourly) are what keep the archive up
# to date.
#
# It also copies BayovCore's character history out of the SavedVariables (history.py), which
# logcheck uses to know the gear, talents and skills at each moment of a log.
#
# FLAVORS lists the game folders we archive, and we skip one that doesn't exist. The beta is
# _classic_beta_. The release client will get a folder of its own, and we don't know its name
# yet. _forever_ is our guess. When the release is out, check the folders in the WoW folder and
# put the real one here (history.py takes the same list). logcheck reads every folder in the
# archive, so it needs no change.

set -eu

GAME="${WOW_DIR:-/mnt/d/Gaming/World of Warcraft}"
ARCHIVE="${FOREVER_LOGS:-$HOME/forever-logs}"
FLAVORS="${FLAVORS:-_classic_beta_ _forever_}"

for flavor in $FLAVORS; do
    src="$GAME/$flavor/Logs"
    [ -d "$src" ] || continue
    mkdir -p "$ARCHIVE/$flavor"
    # -t keeps the game's write time, so the next run only copies a log again when it has
    # changed.
    skip=""
    for gz in "$ARCHIVE/$flavor"/WoWCombatLog-*.txt.gz; do
        [ -e "$gz" ] || continue
        skip="$skip --exclude=$(basename "$gz" .gz)"
    done
    # shellcheck disable=SC2086 # one word per --exclude, and log names have no spaces
    rsync -rt $skip --include 'WoWCombatLog-*.txt' --exclude '*' "$src/" "$ARCHIVE/$flavor/"
done

FLAVORS="$FLAVORS" WOW_DIR="$GAME" FOREVER_LOGS="$ARCHIVE" \
    python3 -I "$(dirname "$0")/history.py"

# The audit loop routine

One scheduled routine runs the loop (`README.md`, "A run"). Its prompt stays thin on purpose:
everything it does is in the repo, so changing the loop is a commit, not a routine edit.

## Settings

- **Name:** Forever audit loop
- **Repository:** bayov/forever-sim, with push access (the run commits its state to master).
  A routine created from a chat session (`create_trigger`) gets no repository and can't push, so
  after creating it that way, open it in the claude.ai routines page and add the repository there.
- **Schedule:** every 6 hours (`0 */6 * * *`).
- **New session each run:** yes. Never bind it to one long-lived session; that is what this
  replaced.
- **Model:** Sonnet (the run is mechanical; it starts Opus subagents for the passes that need judgment).
- **Notifications:** push, so the run's 2 to 4 line reply reaches the phone.

## Prompt

```
Run the forever-sim audit loop: in the bayov/forever-sim checkout, follow
orchestration/audit/README.md, section "A run", exactly. If the repository isn't checked out or
pushing to master is refused, reply with one line saying which and stop.
```

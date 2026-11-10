# Keyring hang / session wipe repro

Reproduces a customer report against the CLI on Linux:
- `session.lock` was held for hours.
- Every command hung, including `auth whoami`.
- A later refresh deleted `session.json`.

```sh
scripts/repro/keyring-hang/run.sh            # old = origin/main
scripts/repro/keyring-hang/run.sh <git-ref>  # compare against another ref
```

The script needs Docker. `fake_secret_service.py` acts as a locked login keyring: every `Unlock` returns a prompt.

Scenario 1, the hang: the prompt never completes, as on a headless or SSH host where nobody can answer the unlock dialog.
- The script prints goroutine stacks (`kill -QUIT`) and `/proc/locks`.
- Old binary:
  - The lock holder and `whoami` are both stuck in `go-keyring`'s `handlePrompt`.
  - A second command waits on the flock.
  - Killing the holder passes both the lock and the hang to the next process.
- New binary: every command fails within `BOLTZ_API_KEYRING_TIMEOUT` with `keyring_unavailable`.

Scenario 2, the wipe: the prompt is dismissed, which looks the same to the CLI as an unreachable keyring.
- Old binary: deletes `session.json`.
- New binary: keeps it.

`TestAuthValidateSubprocessKeepsSessionWhenKeyringUnreachable` covers scenario 2 in CI on every OS.

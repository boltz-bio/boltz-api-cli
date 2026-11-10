#!/bin/bash
# Runs inside the container started by run.sh. Usage: repro.sh old|new
set -u
BIN=/repro/bin/boltz-api-$1
export HOME=/home/u XDG_CONFIG_HOME=/home/u/.config XDG_CACHE_HOME=/home/u/.cache
export BOLTZ_API_AUTH_ISSUER_URL=https://issuer.example.com BOLTZ_API_AUTH_CLIENT_ID=client-123
export BOLTZ_API_AUTH_SCOPE=openid BOLTZ_API_AUTH_AUDIENCE=boltz-compute-api BOLTZ_API_NO_UPDATE_CHECK=1
export BOLTZ_BASE_URL=http://127.0.0.1:1
export BOLTZ_API_KEYRING_TIMEOUT=3s BOLTZ_API_LOCK_TIMEOUT=8s   # unknown to the old binary
mkdir -p $XDG_CACHE_HOME/boltz-api
SESSION=$XDG_CACHE_HOME/boltz-api/session.json
LOCK=$XDG_CACHE_HOME/boltz-api/session.lock
export DBUS_SESSION_BUS_ADDRESS=$(dbus-daemon --session --fork --print-address=1 | head -1)
python3 /repro/fake_secret_service.py > /tmp/ss.log 2>&1 &
sleep 1

write_session() {   # expired access token; refresh token kept in the keyring
cat > $SESSION <<J
{"issuer_url":"https://issuer.example.com","client_id":"client-123","audience":"boltz-compute-api",
 "scopes":["openid"],"access_token":"expired","token_type":"Bearer","expiry":"2020-01-01T00:00:00Z",
 "token_url":"http://127.0.0.1:1/token","jwks_url":"https://issuer.example.com/jwks","algorithms":["RS256"],
 "storage_backend":"keyring"}
J
}
lock_owner() { local ino; ino=$(stat -c %i $LOCK 2>/dev/null) || { echo "  (no lock file)"; return; }
  awk -v ino="$ino" '{split($6,a,":"); if (a[3]==ino) print "  /proc/locks: " $2 " " $4 " held by pid " $5}' /proc/locks; }
stuck_in() { kill -QUIT $1 2>/dev/null; sleep 0.5
  grep -m1 -oE "(secret_service\.\(\*SecretService\)\.handlePrompt|flock\.tryCtx|authstore\.WithLock)" $2 | sed 's/^/stuck in /' || echo "(exited)"; }
code_of() { grep -oE '"code": *"[a-z_]+"' "$@" 2>/dev/null | head -1 | sed 's/.*: *//'; }

echo "=================== $1 binary ==================="
echo "--- Scenario 1: keyring unlock prompt that nobody can answer ---"
echo hang > /tmp/ss-mode; write_session
LIST="predictions:structure-and-binding list --format json"
timeout 60 $BIN $LIST >/tmp/a.out 2>/tmp/a.err & A=$!; sleep 1
echo "A = '$LIST' (token expired -> refresh), pid $A. Lock:"; lock_owner
timeout 60 $BIN $LIST >/tmp/b.out 2>/tmp/b.err & B=$!
timeout 60 $BIN auth whoami --format json >/tmp/w.out 2>/tmp/w.err & W=$!
sleep 1.5
echo "2.5s in: A alive=$(kill -0 $A 2>/dev/null && echo yes || echo no), B (same command, pid $B) alive=$(kill -0 $B 2>/dev/null && echo yes || echo no), whoami alive=$(kill -0 $W 2>/dev/null && echo yes || echo no)"
if [ "$1" = old ]; then
  sleep 3
  echo "goroutine stacks at 5.5s (kill -QUIT; each dump ends that process):"
  echo "  whoami         : $(stuck_in $W /tmp/w.err)"
  timeout 60 $BIN $LIST >/tmp/d.out 2>/tmp/d.err & D=$!; sleep 1
  echo "  B (waiter)     : $(stuck_in $B /tmp/b.err)"
  echo "  A (lock holder): $(stuck_in $A /tmp/a.err)"
  sleep 1
  echo "  A is gone. D (another waiter) takes the lock:"; lock_owner
  echo "  D 1s later     : $(stuck_in $D /tmp/d.err)   <-- the hang moves to the next process"
else
  S=$SECONDS
  wait $A; echo "  A done after ~$((SECONDS-S+2))s  code=$(code_of /tmp/a.out /tmp/a.err)"
  wait $B; echo "  B done after ~$((SECONDS-S+2))s  code=$(code_of /tmp/b.out /tmp/b.err)"
  wait $W; echo "  whoami done: warning=\"$(grep -o 'Could not read[^"]*' /tmp/w.out)\""
fi
echo "session.json after scenario 1: $([ -f $SESSION ] && echo present || echo DELETED)"

echo "--- Scenario 2: the unlock prompt is dismissed (same as: keyring unreachable) ---"
echo dismiss > /tmp/ss-mode; write_session
timeout 60 $BIN $LIST >/tmp/c.out 2>&1
echo "  result: $(grep -oE '"(code|message)": *"[^"]*"' /tmp/c.out | head -2 | tr '\n' ' ')"
echo "session.json after scenario 2: $([ -f $SESSION ] && echo present || echo 'DELETED  <-- every other in-flight run now fails')"

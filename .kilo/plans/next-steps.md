# Next Steps Plan

## Current State
- Juvia v0.1.5 is installed and running on `192.168.0.211:18473`
- Core panel APIs work: health, login, auth/me, metrics
- Agent socket exists and agent is responsive
- Website creation returns empty response — needs investigation/fix
- `test-e2e.sh` has issues parsing first-run token when setup is already complete

## Recommended Priority

### 1. Fix Website Creation (Highest Priority)

The panel's primary purpose is website hosting. Website creation currently returns empty/no data. Need to:

1. Check `internal/api/websites.go` create handler
2. Check `internal/agent/website.go` agent method
3. Verify JSON-RPC call to agent succeeds
4. Verify Linux user creation, document root, and permissions
5. Verify Nginx config generation in `internal/agent/nginx/generator.go`
6. Verify PHP-FPM pool generation
7. Test creating a website and confirm config files appear in `/etc/juvia/nginx/sites/`

### 2. Fix E2E Test Script

Update `scripts/test-e2e.sh` to:

1. Check `/api/v1/setup/status` first
2. If setup is already completed, skip first-run and use existing admin credentials
3. If setup is required, run first-run and capture token properly
4. Add better error output for failed requests
5. Add timeout to curl commands

### 3. Run Full E2E Test

After fixes, run `test-e2e.sh` and verify all 15 tests pass.

### 4. Test Other Agent-Managed Features

Once website creation works, test in order:

1. DNS record creation (BIND9)
2. Database creation (MySQL/PostgreSQL)
3. Mailbox creation (Postfix/Dovecot)
4. Firewall rule (UFW)
5. File manager
6. Terminal session
7. SSL certificate (requires real domain + DNS)

### 5. Fix Agent Methods as Needed

Any feature that fails during testing needs its agent method fixed. Common issues:

- Missing error handling
- Incorrect command paths
- Permission issues
- Missing config directories

### 6. Release v0.1.6

After core features work, tag and release v0.1.6.

## Implementation Order

1. Debug website creation end-to-end
2. Fix agent method for website creation
3. Fix E2E test script
4. Run full E2E test
5. Test and fix remaining agent features
6. Tag v0.1.6

## Success Criteria

- Website creation returns success with website data
- Nginx config and PHP-FPM pool are generated
- `test-e2e.sh` passes all 15 tests
- No 500 errors on core API endpoints

## Notes

- Focus on website creation first — it's the foundation for most other features
- The agent runs as root, so permission issues should be minimal
- Check agent logs at `/var/log/juvia/agent.log` for errors

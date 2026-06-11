# Remaining Work — Post Phase 5 Completion Plan (Revised)

**Current State:** All 5 roadmap phases are structurally implemented. Backend builds (`go build ./...`), tests pass (`go test ./...`), and frontend compiles (`npm run build`). The git repo at https://github.com/marufnwu/juvia.git is empty and needs code pushed.

**Goal:** Fix all script bugs, push code to git, and verify install/uninstall/update flows end-to-end.

---

## Rule
Follow categories in order (A → B → C → D → E). Each task is a single atomic commit that passes `go build ./...`, `go test ./...`, and `npm run build`.

---

## Category A — Fix Script Bugs

### A.1: Fix Service File Paths
- **Problem:** `build-deb.sh` puts binaries in `/usr/bin/` but `juvia.service` and `juvia-agent.service` reference `/usr/local/bin/`
- **Fix:** Change service files to use `/usr/bin/` (matches FHS for distro packages)
- **Files:** `scripts/juvia.service`, `scripts/juvia-agent.service`

### A.2: Fix Install Script Repo URL
- **Problem:** `install.sh` references `https://github.com/juvia-io/juvia` but the actual repo is `https://github.com/marufnwu/juvia`
- **Fix:** Replace all `juvia-io/juvia` with `marufnwu/juvia` in `install.sh`
- **Files:** `scripts/install.sh`

### A.3: Create Uninstall Script
- **Problem:** No standalone `scripts/uninstall.sh` exists
- **Action:** Create `scripts/uninstall.sh` that:
  - Stops and disables `juvia` and `juvia-agent` services
  - Removes `/usr/bin/juvia` and `/usr/bin/juvia-agent`
  - Removes systemd service files
  - Optionally removes `/var/lib/juvia/` and `/var/log/juvia/` (with confirmation)
  - Removes `/etc/juvia/` config directory
  - Optionally removes the `juvia` user (with confirmation)
  - Reloads systemd
- **Files:** New `scripts/uninstall.sh`

### A.4: Create Update Script
- **Problem:** No standalone `scripts/update.sh` exists for updating via CLI
- **Action:** Create `scripts/update.sh` that:
  - Detects current installed version
  - Fetches latest release from GitHub API
  - Downloads new .deb if version is newer
  - Runs `dpkg -i` to update
  - Restarts services
  - Shows changelog or release notes
- **Files:** New `scripts/update.sh`

### A.5: Add postrm to DEB Package
- **Problem:** `build-deb.sh` has `prerm` but no `postrm`
- **Fix:** Add `postrm` script that cleans up remaining files after package removal
- **Files:** Modify `scripts/build-deb.sh`

### A.6: Add preinst to DEB Package
- **Problem:** `build-deb.sh` has no `preinst` script
- **Fix:** Add `preinst` that checks for conflicts (e.g., port 8080 already in use, existing juvia user)
- **Files:** Modify `scripts/build-deb.sh`

### A.7: Fix build.sh to Copy Dist Properly
- **Problem:** `build.sh` copies `web/dist` to `internal/web/dist` but may fail if `internal/web` doesn't exist
- **Fix:** Ensure `internal/web` directory exists before copying, and add `mkdir -p`
- **Files:** Modify `scripts/build.sh`

### A.8: Add Windows Build Script
- **Problem:** No Windows build script exists for development on Windows
- **Action:** Create `scripts/build.ps1` that builds Go binaries and frontend on Windows
- **Files:** New `scripts/build.ps1`

### A.9: Install System Dependencies in install.sh
- **Problem:** `install.sh` only downloads and installs the `.deb` package. It does NOT install the system services that Juvia's agent needs to manage: MySQL, PostgreSQL, Postfix, Dovecot, PHP-FPM, BIND9, Rspamd, etc.
- **Impact:** On a fresh Ubuntu server, installing Juvia would appear to succeed, but creating a website would fail (no PHP-FPM), creating a database would fail (no MySQL), creating an email would fail (no Postfix), etc.
- **Fix:** Add `apt-get install` calls in `install.sh` to install ALL required system packages before installing the `.deb`:
  - **Web Server:** nginx, apache2
  - **PHP:** php-fpm, php-cli + extensions (php-mysql, php-pgsql, php-curl, php-gd, php-mbstring, php-xml, php-zip, php-intl, php-bcmath, php-soap)
  - **Databases:** mysql-server (or mariadb-server), postgresql
  - **Email:** postfix, dovecot-imapd, dovecot-pop3d, rspamd
  - **DNS:** bind9
  - **SSL:** certbot, python3-certbot-nginx
  - **Firewall:** ufw
  - **Metrics:** victoria-metrics (or embed a Go metrics library instead)
  - **Utilities:** git, curl, wget, rsync, unzip, sqlite3, htop, tree, net-tools
- **Reference:** See `docs/DEPLOYMENT.md` lines 172-183 and `docs/MODULES.md` for the full dependency list
- **Files:** Modify `scripts/install.sh`
- **Note:** Also update the `.deb` control file's `Depends:` line in `build-deb.sh` to match

---

## Category B — Git Setup and Push

### B.1: Initialize Git Repo
- **Action:** Initialize git repo, add remote `origin` pointing to `https://github.com/marufnwu/juvia.git`
- **Files:** `.git/config`

### B.2: Create .gitignore
- **Problem:** No `.gitignore` exists — `node_modules`, `dist`, binaries, `.db` files will be committed
- **Action:** Create `.gitignore` with:
  - `node_modules/`
  - `web/dist/`
  - `internal/web/dist/`
  - `juvia`, `juvia-agent` (binaries)
  - `*.db`, `*.db-shm`, `*.db-wal`
  - `.env`
  - `tmp/`, `temp/`
- **Files:** New `.gitignore`

### B.3: Create GitHub Actions CI/CD
- **Action:** Create `.github/workflows/ci.yml` that:
  - Runs on push to `main` and pull requests
  - Sets up Go and Node.js
  - Runs `go test ./...`
  - Runs `npm run build` in `web/`
  - Builds Go binaries
  - Runs `golangci-lint` if configured
- **Files:** New `.github/workflows/ci.yml`

### B.4: Create GitHub Actions Release
- **Action:** Create `.github/workflows/release.yml` that:
  - Triggers on tag push (`v*`)
  - Builds .deb package via `scripts/build-deb.sh`
  - Creates GitHub release with .deb asset
- **Files:** New `.github/workflows/release.yml`

### B.5: Initial Commit and Push
- **Action:**
  1. `git init` (if not already)
  2. `git remote add origin https://github.com/marufnwu/juvia.git`
  3. Stage all files except ignored
  4. Create initial commit with message: "Initial commit — Juvia server control panel"
  5. Push to `main` branch
  6. Create and push tag `v0.1.0`
- **Note:** Must verify git is configured properly before pushing

---

## Category C — Documentation

### C.1: Update README.md
- **Problem:** No README.md exists for the repo
- **Action:** Create `README.md` with:
  - Project description and screenshot placeholder
  - Features list
  - Installation instructions (one-liner from install.sh)
  - Build instructions
  - Architecture overview
  - Contributing guidelines
  - License
- **Files:** New `README.md`

### C.2: Update AGENTS.md
- **Action:** Add build and test commands to AGENTS.md:
  - `go build ./...`
  - `go test ./...`
  - `npm run build`
  - `scripts/build.sh`
  - `scripts/build-deb.sh`
  - `scripts/install.sh`
  - `scripts/uninstall.sh`
- **Files:** Modify `AGENTS.md`

### C.3: Add CHANGELOG.md
- **Action:** Create `CHANGELOG.md` with initial v0.1.0 release notes listing all implemented features
- **Files:** New `CHANGELOG.md`

---

## Category D — Integration Testing

### D.1: Test Build Scripts
- Run `scripts/build.sh` on Windows — verify frontend and Go binaries build
- Run `scripts/build-deb.sh` on Linux (or WSL) — verify .deb package is created
- Verify service files are valid: `systemd-analyze verify scripts/juvia.service`

### D.2: Test Install Script
- Run `scripts/install.sh` on a clean Ubuntu VM
- Verify panel starts and responds on port 8080
- Verify agent starts and socket is created

### D.3: Test Uninstall Script
- Run `scripts/uninstall.sh` after install
- Verify services are stopped and removed
- Verify binaries are removed
- Verify no orphaned processes

### D.4: Test Update Script
- Install v0.1.0, then run `scripts/update.sh`
- Verify new version is installed
- Verify services restart properly

### D.5: End-to-End Panel Test
- Access panel at `http://localhost:8080`
- Complete setup wizard
- Create a website
- Install WordPress
- Verify all sidebar links work

---

## Category E — Final Verification

### E.1: Build Verification
- `go build ./...` — passes
- `go test ./...` — passes
- `npm run build` — passes (no TypeScript errors)
- `scripts/build.sh` — produces working binaries

### E.2: Git Verification
- Repo at https://github.com/marufnwu/juvia.git is not empty
- `main` branch has initial commit
- Tag `v0.1.0` exists
- `.gitignore` is working (no `node_modules` in repo)

### E.3: Package Verification
- `dpkg -I juvia_0.1.0_amd64.deb` shows correct metadata
- `dpkg -c juvia_0.1.0_amd64.deb` shows correct file layout
- Package installs cleanly on Ubuntu 22.04+

---

## Implementation Priority Order

1. **A.1** — Fix service paths (blocks package install)
2. **A.2** — Fix repo URLs (blocks install script)
3. **A.9** — Install system dependencies (CRITICAL — without this, Juvia cannot function)
4. **B.2** — Create .gitignore (blocks commit)
5. **B.1** — Git init + remote (blocks push)
6. **B.5** — Initial commit and push
7. **A.3** — Create uninstall script
8. **A.4** — Create update script
9. **A.5, A.6** — DEB postrm/preinst
10. **B.3, B.4** — GitHub Actions
11. **C.1, C.2, C.3** — Documentation
12. **D.1–D.5** — Integration testing
13. **E.1–E.3** — Final verification

---

## File Summary

| Category | New Files | Modified Files |
|----------|-----------|----------------|
| A.1 | — | `scripts/juvia.service`, `scripts/juvia-agent.service` |
| A.2 | — | `scripts/install.sh` |
| A.3 | `scripts/uninstall.sh` | — |
| A.4 | `scripts/update.sh` | — |
| A.5, A.6 | — | `scripts/build-deb.sh` |
| A.7 | — | `scripts/build.sh` |
| A.8 | `scripts/build.ps1` | — |
| A.9 | — | `scripts/install.sh`, `scripts/build-deb.sh` |
| B.2 | `.gitignore` | — |
| B.3 | `.github/workflows/ci.yml` | — |
| B.4 | `.github/workflows/release.yml` | — |
| B.5 | — | `.git/config` (local only, not committed) |
| C.1 | `README.md` | — |
| C.2 | — | `AGENTS.md` |
| C.3 | `CHANGELOG.md` | — |

---

## Success Criteria

All work is complete when:
1. `go build ./...`, `go test ./...`, and `npm run build` all pass
2. Service file paths match build-deb.sh output paths
3. Install script uses correct repo URL (`marufnwu/juvia`)
4. Git repo at https://github.com/marufnwu/juvia.git has code pushed to `main`
5. Tag `v0.1.0` exists on GitHub
6. `.gitignore` prevents `node_modules`, `dist`, binaries, and `.db` files from being committed
7. Uninstall script cleanly removes all installed components
8. README.md explains installation and features
9. GitHub Actions CI runs on every push
10. `install.sh` installs all system dependencies (nginx, php-fpm, mysql, postgresql, postfix, dovecot, bind9, certbot, ufw, etc.)

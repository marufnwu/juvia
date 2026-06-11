# Fix Plan: `.github/workflows/release.yml`

## Issues Found

### 1. Deprecated GitHub Actions
- `actions/create-release@v1` — archived, unsupported, causes "Resource not accessible by integration"
- `actions/upload-release-asset@v1` — archived, unsupported

### 2. VERSION/ARCH Passed Incorrectly to build-deb.sh
`build-deb.sh` expects VERSION as positional arg `$1` and ARCH as `$2`:
```bash
VERSION=${1:-0.1.0}
ARCH=${2:-amd64}
```
Current workflow passes them as **environment variables** instead of positional arguments:
```yaml
ARCH=amd64 VERSION="${VERSION_STR}" bash scripts/build-deb.sh
```
This means the script always uses defaults (0.1.0, amd64) regardless of the tag.

### 3. Overcomplicated Job Split
`create_release` + `build` as separate jobs with `upload_url` passing is fragile. Better to use a single job or an action that handles both release creation and artifact upload in one step.

### 4. Go Cache Disabled
Already fixed — `cache: false` avoids the "File exists" tar extraction errors.

## Fix Plan

### Step 1: Replace with `ncipollo/release-action@v1`
This action handles both release creation and artifact upload in a single step. Eliminates the two-job split and deprecated actions.

```yaml
- name: Release
  uses: ncipollo/release-action@v1
  with:
    artifacts: "juvia_*.deb"
    artifactContentType: application/vnd.debian.binary-package
    generateReleaseNotes: true
```

### Step 2: Fix build-deb.sh invocation
Pass VERSION and ARCH as positional arguments, not env vars:
```yaml
- name: Build DEB package
  run: |
    chmod +x scripts/build-deb.sh
    VERSION_STR="${GITHUB_REF_NAME#v}"
    bash scripts/build-deb.sh "${VERSION_STR}" amd64
```

### Step 3: Simplify to single job
Collapse into one job:
1. Checkout
2. Setup Go (cache: false)
3. Setup Node.js
4. Build frontend
5. Copy to embed dir
6. Build .deb
7. Release (create + upload in one step)

### Step 4: Add idempotency check
If the release already exists (e.g., re-tagging), the action should update it instead of failing. `ncipollo/release-action@v1` supports `allowUpdates: true`.

## Target File
`.github/workflows/release.yml`

# Phase 2 — Web Hosting Implementation Plan

**Goal:** Website CRUD with SSL and DNS. Success = can create a website, issue SSL, and see it in a browser.

**Rule:** Follow steps in order. Do not skip ahead. Each step is a single atomic commit that passes `go build ./...` and `go test ./...`.

---

## Step 2.1: Website API + Agent Methods

### Backend

**`internal/db/website_queries.go`** — SQLite operations for websites:
- `CreateWebsite(ctx, domain, documentRoot, phpVersion, webServer, userID) (*Website, error)`
- `GetWebsiteByID(ctx, id) (*Website, error)`
- `GetWebsiteByDomain(ctx, domain) (*Website, error)`
- `ListWebsites(ctx, status string, page, limit int) ([]Website, int, error)` — returns items + total count
- `UpdateWebsite(ctx, id, updates) error`
- `SoftDeleteWebsite(ctx, id) error` — sets `deleted_at` and `status = 'deleted'`
- `RestoreWebsite(ctx, id) error` — clears `deleted_at`, sets `status = 'active'`
- `PermanentlyDeleteWebsite(ctx, id) error`
- `SuspendWebsite(ctx, id) error` — sets `status = 'suspended'`
- `ListTrashedWebsites(ctx) ([]Website, error)`
- `CreateDomain(ctx, websiteID, domain, domainType) error`
- `ListDomainsByWebsite(ctx, websiteID) ([]Domain, error)`

Use standard `database/sql` with positional parameters (`?`). Wrap in transactions where needed.

**`internal/api/websites.go`** — REST handlers:
- `GET /api/v1/websites` — list with pagination, sorting, filtering (`?status=active&search=foo&sort=domain&order=asc`)
- `POST /api/v1/websites` — create website
  - Validate domain format (regex + IDN support via `golang.org/x/net/idna`)
  - Check domain uniqueness in DB
  - Create task via `tasks.Runner.CreateTask(ctx, "create_website", userID)`
  - Return 202 with task ID
  - Call agent `website.create` via socket client
  - Agent returns, task completes
  - Insert website record + primary domain record in transaction
  - Audit log entry
- `GET /api/v1/websites/:id` — get website with domains
- `PUT /api/v1/websites/:id` — update (PHP version, web server)
- `DELETE /api/v1/websites/:id` — soft-delete to trash
- `POST /api/v1/websites/:id/suspend` — suspend
- `POST /api/v1/websites/:id/restore` — restore from trash
- `DELETE /api/v1/websites/:id/permanent` — permanent delete
- `GET /api/v1/websites/trash` — list trashed
- `DELETE /api/v1/websites/trash/purge` — purge all trashed
- `PUT /api/v1/websites/:id/php` — change PHP version

Validation rules:
- Domain: valid hostname, not empty, max 255 chars
- PHP version: one of `8.1`, `8.2`, `8.3`
- Web server: `nginx` or `apache`

**`internal/agent/website.go`** — Agent-side handlers (registered in `cmd/agent/main.go`):
- `website.create`: params `{domain, php_version, web_server}`
  - Create Linux user: `useradd -r -s /usr/sbin/nologin -d /home/{domain} -M {domain}`
  - Create directory structure:
    ```
    /home/{domain}/
      public_html/
      logs/
      ssl/
      tmp/
    ```
  - Set ownership: `chown -R {domain}:{domain} /home/{domain}`
  - Set permissions: `chmod 750 /home/{domain}`, `chmod 755 /home/{domain}/public_html`
  - Return `{document_root: "/home/{domain}/public_html", linux_user: "{domain}"}`
- `website.delete`: params `{domain, document_root, force}`
  - If `force=true`: `userdel {domain}` and `rm -rf /home/{domain}`
  - If `force=false`: just move to suspended state (keep files)
- `website.suspend`: params `{domain, document_root}`
  - Stop PHP-FPM pool if running
  - Rename nginx config to `.suspended`
  - Reload nginx

**`cmd/agent/main.go`** — Register new agent methods:
```go
server.RegisterMethod("website.create", agent.HandleWebsiteCreate)
server.RegisterMethod("website.delete", agent.HandleWebsiteDelete)
server.RegisterMethod("website.suspend", agent.HandleWebsiteSuspend)
```

**`internal/api/router.go`** — Add website routes under auth middleware:
```go
websites := api.Group("/websites", authMiddleware(cfg.JWT, cfg.Sessions))
{
    websites.GET("", listWebsitesHandler(cfg))
    websites.POST("", createWebsiteHandler(cfg))
    websites.GET("/trash", listTrashedWebsitesHandler(cfg))
    websites.DELETE("/trash/purge", purgeTrashHandler(cfg))
    websites.GET("/:id", getWebsiteHandler(cfg))
    websites.PUT("/:id", updateWebsiteHandler(cfg))
    websites.DELETE("/:id", deleteWebsiteHandler(cfg))
    websites.POST("/:id/suspend", suspendWebsiteHandler(cfg))
    websites.POST("/:id/restore", restoreWebsiteHandler(cfg))
    websites.DELETE("/:id/permanent", permanentDeleteWebsiteHandler(cfg))
    websites.PUT("/:id/php", changePHPHandler(cfg))
}
```

**Tests:**
- `internal/db/website_queries_test.go` — test all CRUD operations with in-memory DB
- `internal/api/websites_test.go` — test handlers with mocked DB and agent client
- `internal/agent/website_test.go` — test handler logic (mock system calls via injected exec wrapper)

---

## Step 2.2: Nginx Config Generation

### Backend

**`internal/agent/nginx/generator.go`**:

```go
package nginx

type SiteConfig struct {
    Domain       string
    DocumentRoot string
    PHPVersion   string
    SSLEnabled   bool
    SSLCertPath  string
    SSLKeyPath   string
    ServerName   string
}

func GenerateSiteConfig(cfg SiteConfig) (string, error)
func WriteSiteConfig(domain, content string) error   // atomic swap
func ValidateConfig(path string) error               // runs nginx -t
func ReloadNginx() error
func RemoveSiteConfig(domain string) error
```

Nginx template (`internal/agent/nginx/templates/site.conf.tmpl`):
```nginx
server {
    listen 80;
    server_name {{.Domain}} www.{{.Domain}};
    root {{.DocumentRoot}};
    index index.php index.html;

    access_log /home/{{.Domain}}/logs/access.log;
    error_log /home/{{.Domain}}/logs/error.log;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/run/php/php{{.PHPVersion}}-fpm-{{.Domain}}.sock;
    }

    location ~ /\.ht {
        deny all;
    }
}
```

PHP-FPM pool template (`internal/agent/nginx/templates/pool.conf.tmpl`):
```ini
[{{.Domain}}]
user = {{.LinuxUser}}
group = {{.LinuxUser}}
listen = /run/php/php{{.PHPVersion}}-fpm-{{.Domain}}.sock
listen.owner = www-data
listen.group = www-data
pm = dynamic
pm.max_children = 5
pm.start_servers = 2
pm.min_spare_servers = 1
pm.max_spare_servers = 3
chdir = /
php_admin_value[open_basedir] = {{.DocumentRoot}}:/tmp
```

Atomic swap pattern for all config writes:
1. Write to `/etc/juvia/nginx/sites/{domain}.conf.tmp`
2. Run `nginx -t -c /etc/nginx/nginx.conf` (or equivalent path validation)
3. If valid: `os.Rename()` to `.conf`
4. If invalid: delete `.tmp`, return error
5. Run `systemctl reload nginx` or `nginx -s reload`

Config paths:
- Nginx sites: `/etc/juvia/nginx/sites/{domain}.conf`
- PHP-FPM pools: `/etc/juvia/php-fpm/pools/{domain}.conf`
- These paths must be included by the main nginx/php-fpm configs (managed by install script, not in this phase)

**`internal/agent/website.go`** — Update `website.create` handler:
- After creating user/dirs, call `nginx.GenerateSiteConfig()` + `WriteSiteConfig()`
- Call `phpfpm.GeneratePoolConfig()` + `WritePoolConfig()`
- Validate both configs
- Reload nginx and php-fpm

**`internal/agent/website.go`** — Update `website.delete` handler:
- Remove nginx site config
- Remove PHP-FPM pool config
- Reload both services

**`internal/agent/website.go`** — Update `website.suspend` handler:
- Replace active nginx config with suspended page config
- Stop PHP-FPM pool
- Reload nginx

**Tests:**
- `internal/agent/nginx/generator_test.go` — test template rendering, atomic swap, validation errors
- Use `os.CreateTemp()` for test config paths

---

## Step 2.3: SSL (Certbot)

### Backend

**`internal/agent/ssl/certbot.go`**:
```go
package ssl

func IssueCertificate(domain string) (certPath, keyPath string, expiry time.Time, error)
func RenewCertificate(domain string) error
func CheckDNS(domain string) (ready bool, error)
func GetCertificateExpiry(domain string) (*time.Time, error)
func RemoveCertificate(domain string) error
```

Implementation:
- `IssueCertificate`: run `certbot certonly --nginx -d {domain} -d www.{domain} --non-interactive --agree-tos -m admin@{domain}`
- Parse cert path from output: `/etc/letsencrypt/live/{domain}/fullchain.pem`, `/etc/letsencrypt/live/{domain}/privkey.pem`
- Parse expiry via `openssl x509 -enddate -noout -in {certPath}`
- `RenewCertificate`: `certbot renew --cert-name {domain} --non-interactive`
- `CheckDNS`: resolve domain A record and compare to server public IP (read from `settings` table key `server_ip`, fallback to `curl ifconfig.me`)

**`internal/api/ssl.go`** — REST handlers:
- `GET /api/v1/websites/:id/ssl/check` — check DNS readiness
  - Call agent `ssl.check`
  - Return `{ready: true/false, message: "..."}`
- `POST /api/v1/websites/:id/ssl` — issue SSL
  - Create task `issue_ssl`
  - Call agent `ssl.issue`
  - On success: update `websites.ssl_enabled = true`, `ssl_expiry = {expiry}`
  - Update nginx config with SSL server block + HTTP→HTTPS redirect
- `POST /api/v1/websites/:id/ssl-renew` — renew SSL
  - Create task `renew_ssl`
  - Call agent `ssl.renew`
  - Update `ssl_expiry`
- `DELETE /api/v1/websites/:id/ssl` — remove SSL
  - Call agent `ssl.remove`
  - Update nginx config to remove SSL block
  - Set `ssl_enabled = false`, `ssl_expiry = NULL`

**Agent methods** (register in `cmd/agent/main.go`):
```go
server.RegisterMethod("ssl.issue", ssl.HandleIssue)
server.RegisterMethod("ssl.renew", ssl.HandleRenew)
server.RegisterMethod("ssl.check", ssl.HandleCheck)
server.RegisterMethod("ssl.remove", ssl.HandleRemove)
```

**Tests:**
- `internal/agent/ssl/certbot_test.go` — mock certbot commands, test parsing logic
- `internal/api/ssl_test.go` — test handlers with mocked agent client

---

## Step 2.4: DNS (BIND9)

### Backend

**`internal/agent/dns/bind.go`**:
```go
package dns

func GenerateZoneFile(domain, serverIP string, records []DNSRecord) (string, error)
func ValidateZoneFile(domain, path string) error   // named-checkzone
func WriteZoneFile(domain, content string) error   // atomic swap
func ReloadBind() error
func RemoveZoneFile(domain string) error
```

Zone template:
```
$TTL 3600
@    IN    SOA    ns1.{domain}.    admin.{domain}. (
              {serial}    ; Serial
              3600        ; Refresh
              1800        ; Retry
              604800      ; Expire
              86400 )     ; Minimum TTL

@    IN    NS     ns1.{domain}.
@    IN    NS     ns2.{domain}.
@    IN    A      {serverIP}
www  IN    CNAME  @
mail IN    A      {serverIP}
@    IN    MX     10    mail.{domain}.
@    IN    TXT    "v=spf1 mx ~all"
_dmarc   IN    TXT    "v=DMARC1; p=quarantine; rua=mailto:dmarc@{domain}"
```

Auto-create records on website creation:
- When `website.create` succeeds, panel calls `dns.zone.create` agent method
- Agent writes zone file with default records
- Panel inserts zone + dns_records into SQLite

**`internal/api/dns.go`** — REST handlers:
- `GET /api/v1/websites/:id/dns/records` — list records for website's zone
- `POST /api/v1/websites/:id/dns/records` — create record
  - Validate record type (A, AAAA, CNAME, MX, TXT, NS, CAA, SRV)
  - Validate name/value format per type
  - Insert into DB
  - Call agent `dns.zone.update` to regenerate zone file
- `PUT /api/v1/dns/records/:id` — update record
- `DELETE /api/v1/dns/records/:id` — delete record
- `POST /api/v1/websites/:id/dns/suggest` — suggest missing records
  - Check for DKIM, CAA, SPF, DMARC
  - Return suggestions array with plain-English descriptions

**Agent methods**:
```go
server.RegisterMethod("dns.zone.create", dns.HandleZoneCreate)
server.RegisterMethod("dns.zone.update", dns.HandleZoneUpdate)
server.RegisterMethod("dns.zone.delete", dns.HandleZoneDelete)
```

Zone file path: `/etc/juvia/bind/zones/{domain}.zone`
Use atomic swap + `named-checkzone {domain} {path}` before activating.

**Tests:**
- `internal/agent/dns/bind_test.go` — test zone generation, validation, atomic swap
- `internal/api/dns_test.go` — test CRUD with mocked agent

---

## Step 2.5: Website Frontend

### React Components

**`web/src/pages/Websites/WebsiteList.tsx`**:
- Table with columns: Domain, PHP Version, Web Server, SSL Status, Status, Actions
- Pagination, sorting, search
- Action buttons: View, Suspend, Delete (soft)
- "Create Website" button → opens create modal or navigates to create page
- Plain-English explanations: "PHP Version · The programming language your website's code runs on"

**`web/src/pages/Websites/CreateWebsite.tsx`**:
- Form with fields: Domain, PHP Version (select: 8.1/8.2/8.3), Web Server (select: nginx/apache)
- Domain validation feedback (real-time)
- Submit creates website, shows task progress via WebSocket
- On completion: redirect to website detail page

**`web/src/pages/Websites/WebsiteDetail.tsx`**:
- Tabs: Overview, SSL, DNS, Files, Databases, Logs, Settings
- Overview: domain, document root, PHP version, web server, status, created date
- SSL tab: show certificate status, expiry date, renew button, DNS check button
- DNS tab: record table, add/edit/delete records, suggest button
- Plain-English labels everywhere

**`web/src/pages/Websites/Trash.tsx`**:
- List of trashed websites
- Actions: Restore, Permanently Delete
- "Purge All" button with confirmation

**`web/src/App.tsx`** — Add routes:
```tsx
<Route path="websites" element={<WebsiteList />} />
<Route path="websites/create" element={<CreateWebsite />} />
<Route path="websites/:id" element={<WebsiteDetail />} />
<Route path="websites/trash" element={<Trash />} />
```

**`web/src/components/layout/Sidebar.tsx`** — Add navigation:
- Websites (with sub-items: All Sites, Trash)

**API hooks** (`web/src/lib/api.ts` or new `web/src/hooks/useWebsites.ts`):
- Use React Query for caching:
  ```ts
  export function useWebsites(params) { return useQuery({ queryKey: ['websites', params], queryFn: ... }) }
  export function useCreateWebsite() { return useMutation({ mutationFn: ..., onSuccess: () => queryClient.invalidateQueries(['websites']) }) }
  ```

**Tests:**
- `web/src/pages/Websites/WebsiteList.test.tsx`
- `web/src/pages/Websites/CreateWebsite.test.tsx`
- Use React Testing Library + MSW for API mocking

---

## Step 2.6: File Manager API

### Backend

**`internal/api/files.go`** — REST handlers:
- `GET /api/v1/websites/:id/files` — list files in document root
  - Query params: `?path=relative/path`
  - Returns: `{items: [{name, type, size, modified_at, permissions}]}`
- `POST /api/v1/websites/:id/files/upload` — multipart upload
  - Max size: 512MB (configurable)
  - Save to `document_root + safePath(path)`
- `GET /api/v1/websites/:id/files/download` — download file or folder as zip
  - Query: `?path=relative/path`
- `PUT /api/v1/websites/:id/files/rename` — rename file/folder
  - Body: `{old_path, new_path}`
- `DELETE /api/v1/websites/:id/files/delete` — delete file/folder
- `POST /api/v1/websites/:id/files/extract` — extract archive
  - Supported: zip, tar.gz
  - Max extracted size: 1GB
  - Max entries: 10,000
- `GET /api/v1/websites/:id/files/edit` — get file content
  - Returns: `{content, encoding: "utf-8"}`
- `PUT /api/v1/websites/:id/files/edit` — save file content
  - Body: `{content, encoding}`

All paths validated with `utils.SafePath(documentRoot, userInput)` before any operation.

**`internal/agent/files/operations.go`**:
```go
package files

func ListFiles(basePath, relPath string) ([]FileInfo, error)
func UploadFile(basePath, relPath string, data io.Reader, size int64) error
func DownloadFile(basePath, relPath string) (io.ReadCloser, int64, error)
func DeleteFile(basePath, relPath string) error
func RenameFile(basePath, oldRelPath, newRelPath string) error
func ExtractArchive(basePath, relPath string) error
```

Security:
- Every path must pass `utils.SafePath(base, relPath)`
- No absolute paths in archive extraction
- Reject paths containing `..` after cleaning
- Uploaded files in `public_html` cannot be executed unless site type allows it (chmod enforcement)

**Agent methods**:
```go
server.RegisterMethod("files.list", files.HandleList)
server.RegisterMethod("files.upload", files.HandleUpload)
server.RegisterMethod("files.download", files.HandleDownload)
server.RegisterMethod("files.delete", files.HandleDelete)
server.RegisterMethod("files.rename", files.HandleRename)
server.RegisterMethod("files.extract", files.HandleExtract)
```

**Tests:**
- `internal/agent/files/operations_test.go` — test all operations with temp directories
- `internal/api/files_test.go` — test handlers with mocked agent
- Test path traversal attempts are rejected

---

## Database Notes

The `websites`, `domains`, `zones`, and `dns_records` tables already exist in `001_initial_schema.sql`. No new migration is needed for Phase 2 unless schema changes are discovered during implementation.

If schema changes are needed:
1. Create `internal/db/migrations/002_phase2_webhosting.sql`
2. Run on startup via `db.ApplyMigrations()`

---

## Audit Log

Every mutating operation in Phase 2 must write to `audit_log`:
- Create website: `"Created website {domain}"`
- Delete website: `"Moved website {domain} to trash"`
- Restore: `"Restored website {domain} from trash"`
- Suspend: `"Suspended website {domain}"`
- SSL issue: `"Issued SSL certificate for {domain}"`
- DNS record create: `"Added {type} record to {domain}"`

Use `internal/api/audit.go` helper (create if not exists):
```go
func logAudit(ctx context.Context, db *sql.DB, userID int64, action, ip, ua, details string) error
```

---

## WebSocket Task Progress

All async operations (create website, issue SSL, renew SSL) must:
1. Create task → return 202 + task ID
2. Call agent method
3. Update task progress via `tasks.Runner.UpdateTaskProgress()`
4. Publish event via `eventBus.Publish(events.Event{Type: events.EventTaskProgress, Data: ...})`
5. On completion: `CompleteTask()` + `Publish(EventTaskComplete)`
6. On failure: `FailTask()` + `Publish(EventTaskComplete)`

Frontend connects to `/ws/v1/tasks/{taskId}` to show real-time progress.

---

## Security Checklist

- [ ] All file paths use `utils.SafePath()`
- [ ] All config writes use atomic swap
- [ ] Domain validation prevents invalid hostnames
- [ ] User can only access their own websites (check `user_id` or `user_sites` junction)
- [ ] Admin can access all websites
- [ ] No secrets in nginx/php-fpm configs
- [ ] Upload limits enforced
- [ ] Archive extraction has size/entry limits

---

## Testing Checklist

Every step must pass before commit:
- [ ] `go build ./...`
- [ ] `go test ./...`
- [ ] Frontend build: `cd web && npm run build` (no TypeScript errors)
- [ ] Manual test: create website → verify dirs created → verify nginx config → verify site responds

---

## File Summary

| Step | New Files | Modified Files |
|------|-----------|----------------|
| 2.1 | `internal/db/website_queries.go`, `internal/api/websites.go`, `internal/agent/website.go`, `internal/db/website_queries_test.go`, `internal/api/websites_test.go`, `internal/agent/website_test.go` | `cmd/agent/main.go`, `internal/api/router.go` |
| 2.2 | `internal/agent/nginx/generator.go`, `internal/agent/nginx/templates/site.conf.tmpl`, `internal/agent/nginx/templates/pool.conf.tmpl`, `internal/agent/phpfpm/generator.go`, `internal/agent/nginx/generator_test.go` | `internal/agent/website.go`, `cmd/agent/main.go` |
| 2.3 | `internal/agent/ssl/certbot.go`, `internal/api/ssl.go`, `internal/agent/ssl/certbot_test.go`, `internal/api/ssl_test.go` | `cmd/agent/main.go`, `internal/api/router.go` |
| 2.4 | `internal/agent/dns/bind.go`, `internal/api/dns.go`, `internal/agent/dns/bind_test.go`, `internal/api/dns_test.go` | `cmd/agent/main.go`, `internal/api/router.go` |
| 2.5 | `web/src/pages/Websites/WebsiteList.tsx`, `web/src/pages/Websites/CreateWebsite.tsx`, `web/src/pages/Websites/WebsiteDetail.tsx`, `web/src/pages/Websites/Trash.tsx`, `web/src/hooks/useWebsites.ts`, `web/src/hooks/useWebsiteDetail.ts`, `web/src/pages/Websites/WebsiteList.test.tsx`, `web/src/pages/Websites/CreateWebsite.test.tsx` | `web/src/App.tsx`, `web/src/components/layout/Sidebar.tsx` |
| 2.6 | `internal/api/files.go`, `internal/agent/files/operations.go`, `internal/api/files_test.go`, `internal/agent/files/operations_test.go` | `cmd/agent/main.go`, `internal/api/router.go` |

---

## Success Criteria

Phase 2 is complete when:
1. `POST /api/v1/websites` creates a website with Linux user, dirs, nginx config, PHP-FPM pool
2. `GET /api/v1/websites` returns paginated list
3. Website detail page shows all tabs (Overview, SSL, DNS, Files)
4. SSL can be issued via API and frontend
5. DNS records can be managed via API and frontend
6. File manager can list/upload/download files
7. All tests pass (`go test ./...`)
8. Frontend builds without errors
9. Can visit created website in browser (on target server)

# Phase 3 — Services Implementation Plan

**Goal:** Email, databases, file manager UI. Success = can create a mailbox, database, and browse files.

**Rule:** Follow steps in order. Do not skip ahead. Each step is a single atomic commit that passes `go build ./...` and `go test ./...`.

---

## Step 3.1: Email API + Agent Methods

### Backend

**`internal/db/email_queries.go`** — SQLite operations for email:
- `CreateMailbox(ctx, email, passwordHash, quota, displayName) (*Mailbox, error)`
- `GetMailboxByID(ctx, id) (*Mailbox, error)`
- `GetMailboxByEmail(ctx, email) (*Mailbox, error)`
- `ListMailboxes(ctx, search, page, limit) ([]Mailbox, int, error)`
- `UpdateMailbox(ctx, id, displayName, forwardTo, quota) error`
- `DeleteMailbox(ctx, id) error`
- `CreateAlias(ctx, domain, source, destination) (*EmailAlias, error)`
- `ListAliasesByDomain(ctx, domain) ([]EmailAlias, error)`
- `DeleteAlias(ctx, id) error`
- `CreateForwarder(ctx, domain, source, destination) (*EmailForwarder, error)`
- `ListForwardersByDomain(ctx, domain) ([]EmailForwarder, error)`
- `DeleteForwarder(ctx, id) error`
- `GetCatchAll(ctx, domain) (*EmailCatchAll, error)`
- `SetCatchAll(ctx, domain, forwardTo) error`

**`internal/api/email.go`** — REST handlers:
- `GET /api/v1/email/mailboxes` — list with pagination, search
- `POST /api/v1/email/mailboxes` — create mailbox
  - Validate email format (RFC 5322 simple regex)
  - Hash password with bcrypt
  - Insert into DB
  - Call agent `email.create` to create Postfix/Dovecot user
  - Audit log entry
- `GET /api/v1/email/mailboxes/:id` — get mailbox
- `DELETE /api/v1/email/mailboxes/:id` — delete mailbox
  - Call agent `email.delete` to remove Postfix/Dovecot user
- `PUT /api/v1/email/mailboxes/:id` — update (display_name, forward_to, quota)
- `GET /api/v1/email/aliases` — list aliases
- `POST /api/v1/email/aliases` — create alias
- `DELETE /api/v1/email/aliases/:id` — delete alias
- `GET /api/v1/email/forwarders` — list forwarders
- `POST /api/v1/email/forwarders` — create forwarder
- `DELETE /api/v1/email/forwarders/:id` — delete forwarder
- `GET /api/v1/email/catch-all` — get catch-all for domain
- `PUT /api/v1/email/catch-all` — set catch-all for domain
- `GET /api/v1/email/deliverability` — check domain deliverability
  - Returns SPF/DKIM/DMARC/PTR status with plain-English descriptions

**`internal/agent/email/postfix.go`**:
```go
func HandleMailboxCreate(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleMailboxDelete(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleMailboxUpdate(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleAliasCreate(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleAliasDelete(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleForwarderCreate(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleForwarderDelete(ctx context.Context, params json.RawMessage) (interface{}, error)
```

Implementation details:
- **Create mailbox**: Update Postfix `virtual_mailbox_maps` and Dovecot user DB
  - Use `doveadm pw -s bcrypt` to hash password for Dovecot
  - Create mailbox directory at `/var/mail/{domain}/{user}/`
  - Set ownership `dovecot:dovecot`
  - Reload Postfix + Dovecot
- **Delete mailbox**: Remove from maps, delete maildir
- **Alias**: Update `virtual_alias_maps`
- **Forwarder**: Update `virtual_alias_maps`
- **Catch-all**: Update `virtual_alias_maps` with `@domain`
- **Deliverability**: Read DNS records via `dig`/`nslookup`, check blacklist via DNS query

**Register agent methods in `cmd/agent/main.go`**:
```go
server.RegisterMethod("email.create", email.HandleMailboxCreate)
server.RegisterMethod("email.delete", email.HandleMailboxDelete)
server.RegisterMethod("email.update", email.HandleMailboxUpdate)
server.RegisterMethod("email.alias.create", email.HandleAliasCreate)
server.RegisterMethod("email.alias.delete", email.HandleAliasDelete)
server.RegisterMethod("email.forwarder.create", email.HandleForwarderCreate)
server.RegisterMethod("email.forwarder.delete", email.HandleForwarderDelete)
```

**Update `internal/api/router.go`** with email routes.

---

## Step 3.2: Database API + Agent Methods

### Backend

**`internal/db/database_queries.go`** — SQLite operations:
- `CreateDatabase(ctx, name, engine, userID) (*Database, error)`
- `GetDatabaseByID(ctx, id) (*Database, error)`
- `ListDatabases(ctx, engine, page, limit) ([]Database, int, error)`
- `DeleteDatabase(ctx, id) error`
- `CreateDBUser(ctx, databaseID, username, passwordHash, host) (*DBUser, error)`
- `ListDBUsers(ctx, databaseID) ([]DBUser, error)`
- `DeleteDBUser(ctx, id) error`

**`internal/api/databases.go`** — REST handlers:
- `GET /api/v1/databases` — list databases
- `POST /api/v1/databases` — create database
  - Validate name: alphanumeric + underscore, max 64 chars
  - Engine: `mysql` or `postgresql`
  - Call agent `database.create`
  - Insert DB record, create initial DB user with bcrypt hash
  - Audit log
- `GET /api/v1/databases/:id` — get database with users
- `DELETE /api/v1/databases/:id` — delete database
  - Call agent `database.delete`
  - Delete DB record and users
- `POST /api/v1/databases/:id/users` — create DB user
  - Generate random password, hash with bcrypt
  - Call agent `database.user.create`
- `DELETE /api/v1/databases/:id/users/:user_id` — delete DB user
- `POST /api/v1/databases/:id/export` — export database
  - Call agent `database.export`
  - Return download URL or stream SQL dump
- `POST /api/v1/databases/:id/import` — import SQL file
  - Accept multipart upload, call agent `database.import`
- `GET /api/v1/databases/:id/tables` — list tables
  - Call agent `database.tables`
- `GET /api/v1/databases/:id/tables/:table/rows` — get table rows (paginated)
  - Query params: `?page=1&limit=50&sort=id&order=asc`
- `POST /api/v1/databases/:id/query` — run SQL query
  - Body: `{ "query": "SELECT * FROM users" }`
  - Read-only for SELECT, show warning for mutating queries

**`internal/agent/database/mysql.go`**:
```go
func HandleDatabaseCreate(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleDatabaseDelete(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleDBUserCreate(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleDBUserDelete(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleExport(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleImport(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleListTables(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleGetRows(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleQuery(ctx context.Context, params json.RawMessage) (interface{}, error)
```

Implementation details:
- **Create database**: `mysql -e "CREATE DATABASE IF NOT EXISTS {name} CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"`
- **Create user**: `mysql -e "CREATE USER '{user}'@'localhost' IDENTIFIED WITH caching_sha2_password BY '{password}'; GRANT ALL PRIVILEGES ON {db}.* TO '{user}'@'localhost'; FLUSH PRIVILEGES;"`
- **Delete database**: `mysql -e "DROP DATABASE IF EXISTS {name};"` + drop users
- **Export**: `mysqldump {name} > {tmp_path}` — return file path
- **Import**: `mysql {name} < {sql_path}`
- **List tables**: Query `information_schema.tables`
- **Get rows**: Execute SELECT with LIMIT/OFFSET
- **Query**: Execute arbitrary SQL (SELECT for panel, warn on others)

**`internal/agent/database/postgresql.go`** — Same interface as MySQL but using `psql`, `createdb`, `dropdb`, `pg_dump`, `pg_restore` commands.

**Register agent methods in `cmd/agent/main.go`**:
```go
server.RegisterMethod("database.create", database.HandleDatabaseCreate)
server.RegisterMethod("database.delete", database.HandleDatabaseDelete)
server.RegisterMethod("database.user.create", database.HandleDBUserCreate)
server.RegisterMethod("database.user.delete", database.HandleDBUserDelete)
server.RegisterMethod("database.export", database.HandleExport)
server.RegisterMethod("database.import", database.HandleImport)
server.RegisterMethod("database.tables", database.HandleListTables)
server.RegisterMethod("database.rows", database.HandleGetRows)
server.RegisterMethod("database.query", database.HandleQuery)
```

**Update `internal/api/router.go`** with database routes.

---

## Step 3.3: File Manager Frontend

### React Components

**`web/src/pages/Files/FileManager.tsx`**:
- Route: `/files/:websiteId`
- Features:
  - Breadcrumb navigation showing current path
  - File/folder list with icons (directory vs file)
  - Double-click to navigate into directories
  - Drag-and-drop upload zone
  - Right-click context menu (rename, delete, download, extract)
  - Monaco Editor modal for editing files inline
  - Hidden files toggle
  - Bulk select with checkboxes
  - "Upload" button and "New Folder" button
  - Plain-English tooltips: "Upload · Send files to your website's server"

State management:
```ts
interface FileItem {
  name: string
  type: 'file' | 'directory'
  size: number
  modified_at: number
  permissions: string
}
```

API calls:
- `GET /api/v1/websites/{id}/files?path={relPath}` — list files
- `POST /api/v1/websites/{id}/files/upload` — upload files
- `GET /api/v1/websites/{id}/files/download?path={relPath}` — download
- `PUT /api/v1/websites/{id}/files/rename` — rename
- `DELETE /api/v1/websites/{id}/files/delete` — delete
- `POST /api/v1/websites/{id}/files/extract` — extract archive
- `GET /api/v1/websites/{id}/files/edit?path={relPath}` — get content
- `PUT /api/v1/websites/{id}/files/edit` — save content

**`web/src/App.tsx`** — Add routes:
```tsx
<Route path="files/:websiteId" element={<FileManager />} />
<Route path="files" element={<FileManager />} />
```

**Monaco Editor integration**:
- Use `@monaco-editor/react` package
- Lazy load for performance
- Syntax highlighting based on file extension
- Auto-save debounce (2 seconds)

**Tests**:
- `web/src/pages/Files/FileManager.test.tsx` — test rendering, navigation, upload

---

## Step 3.4: Visual Database Manager

### React Components

**`web/src/pages/Databases/DatabaseManager.tsx`**:
- Route: `/databases`
- Two-pane layout: database list (left), detail view (right)

**Database list view:**
- Table with columns: Name, Engine, Users, Created
- "Create Database" button → modal with form

**Database detail view (when selected):**
- Tabs: Tables, SQL Editor, Users, Import/Export

**Tables tab:**
- List tables with row count
- Click table → show rows with pagination
- Inline row editing (click cell → edit → save)
- "Add Row" button with form

**SQL Editor tab:**
- Monaco Editor for SQL
- "Run Query" button
- Results displayed in table below
- Syntax highlighting for SQL

**Users tab:**
- List DB users with permissions
- "Add User" / "Remove User" buttons

**Import/Export tab:**
- Upload SQL dump for import
- Download SQL dump for export
- Progress indicator

**`web/src/pages/Databases/CreateDatabase.tsx`**:
- Form: name, engine (MySQL/PostgreSQL radio)
- Validation: alphanumeric + underscore

**`web/src/App.tsx`** — Add routes:
```tsx
<Route path="databases" element={<DatabaseManager />} />
<Route path="databases/create" element={<CreateDatabase />} />
```

**Tests**:
- `web/src/pages/Databases/DatabaseManager.test.tsx`

---

## Step 3.5: Email Frontend

### React Components

**`web/src/pages/Email/EmailDashboard.tsx`**:
- Route: `/email`
- Tabs: Mailboxes, Aliases, Forwarders, Catch-All, Deliverability

**Mailboxes tab:**
- Table: Email, Quota, Display Name, Status, Actions
- "Create Mailbox" button → modal with form (email, password, quota)
- Edit/Delete actions per row
- Plain-English: "Mailbox · An email account that stores messages on your server"

**Aliases tab:**
- Table: Source → Destination
- "Create Alias" form

**Forwarders tab:**
- Table: Source → Destination
- "Create Forwarder" form

**Catch-All tab:**
- Domain selector
- Forward-to input
- Plain-English: "Catch-All · Receives email sent to any address at your domain that doesn't exist"

**Deliverability tab:**
- Domain input
- Check button
- Results cards:
  - SPF: pass/fail with explanation
  - DKIM: pass/fail with explanation
  - DMARC: pass/fail with explanation
  - PTR: pass/fail with explanation
  - Blacklist status (Spamhaus, Barracuda)
- All plain-English: "SPF Record · Authorizes which servers can send email for your domain"

**`web/src/App.tsx`** — Add route:
```tsx
<Route path="email" element={<EmailDashboard />} />
```

**Tests**:
- `web/src/pages/Email/EmailDashboard.test.tsx`

---

## Sidebar Updates

**`web/src/components/layout/Sidebar.tsx`**:
- Ensure routes exist: `/databases`, `/email`, `/files`
- These are already in the nav array from Phase 2

---

## Testing Strategy

### Go Tests
- `internal/db/email_queries_test.go` — Test CRUD operations
- `internal/db/database_queries_test.go` — Test CRUD operations
- `internal/api/email_test.go` — Test API handlers (mock agent client)
- `internal/api/databases_test.go` — Test API handlers (mock agent client)
- `internal/agent/email/postfix_test.go` — Test agent handlers (mock shell commands)
- `internal/agent/database/mysql_test.go` — Test MySQL operations (mock shell commands)
- `internal/agent/database/postgresql_test.go` — Test PostgreSQL operations (mock shell commands)

### Frontend Tests
- `web/src/pages/Email/EmailDashboard.test.tsx`
- `web/src/pages/Databases/DatabaseManager.test.tsx`
- `web/src/pages/Files/FileManager.test.tsx`

---

## File Summary

| Step | New Files | Modified Files |
|------|-----------|----------------|
| 3.1 | `internal/db/email_queries.go`, `internal/api/email.go`, `internal/agent/email/postfix.go`, `internal/db/email_queries_test.go`, `internal/api/email_test.go`, `internal/agent/email/postfix_test.go` | `cmd/agent/main.go`, `internal/api/router.go` |
| 3.2 | `internal/db/database_queries.go`, `internal/api/databases.go`, `internal/agent/database/mysql.go`, `internal/agent/database/postgresql.go`, `internal/db/database_queries_test.go`, `internal/api/databases_test.go`, `internal/agent/database/mysql_test.go`, `internal/agent/database/postgresql_test.go` | `cmd/agent/main.go`, `internal/api/router.go` |
| 3.3 | `web/src/pages/Files/FileManager.tsx`, `web/src/pages/Files/FileManager.test.tsx` | `web/src/App.tsx` |
| 3.4 | `web/src/pages/Databases/DatabaseManager.tsx`, `web/src/pages/Databases/CreateDatabase.tsx`, `web/src/pages/Databases/DatabaseManager.test.tsx` | `web/src/App.tsx` |
| 3.5 | `web/src/pages/Email/EmailDashboard.tsx`, `web/src/pages/Email/EmailDashboard.test.tsx` | `web/src/App.tsx` |

---

## Success Criteria

Phase 3 is complete when:
1. `POST /api/v1/email/mailboxes` creates a mailbox with Postfix/Dovecot integration
2. `POST /api/v1/databases` creates a MySQL/PostgreSQL database with users
3. File manager can browse, upload, and edit files via the web UI
4. Database manager can browse tables and run SQL queries
5. Email dashboard shows mailbox list, aliases, forwarders, and deliverability checks
6. All tests pass (`go test ./...`)
7. Frontend builds without errors (`npm run build`)

---

## Important Notes

- **No schema changes needed** — all tables (mailboxes, email_aliases, email_forwarders, email_catchall, databases, db_users) already exist in `001_initial_schema.sql`.
- **Agent commands must handle absence gracefully** — if `postfix`, `dovecot`, `mysql`, or `psql` are not installed, return a clear error message.
- **Password generation** — DB user passwords should be auto-generated (20 chars, alphanumeric + symbols) and shown once to the user.
- **Security** — Database SQL queries must be parameterized. Never interpolate user input into SQL strings.
- **Plain-English UI** — Every technical term must have a plain-English explanation in the frontend.

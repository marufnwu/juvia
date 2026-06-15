# DNS / Nameserver / Domain Implementation — Gap Analysis & Fix Plan

## Executive Summary

The DNS stack is **partially implemented and functionally fragile**. The database schema, API endpoints, agent handlers, and frontend UI all exist, but several critical gaps mean BIND zone files can get out of sync with the database, failures are silently ignored, the serial number does not increment correctly, and the frontend shows misleading or hard-coded data. This plan proposes a focused refactor to make DNS authoritative, reliable, and observable.

---

## 1. Current Architecture (as found)

| Layer | Files |
|-------|-------|
| Database / Models | `internal/db/models.go`, `internal/db/website_queries.go`, `internal/db/migrations/001_initial_schema.sql`, `004_add_domains_unique_index.sql`, `005_add_domain_ssl.sql` |
| API / Backend | `internal/api/dns.go`, `internal/api/domains.go`, `internal/api/websites.go`, `internal/api/settings.go`, `internal/api/router.go` |
| Agent (privileged) | `internal/agent/dns.go`, `cmd/agent/main.go` |
| Frontend | `web/src/pages/DNS/DnsManagement.tsx`, `web/src/pages/Websites/WebsiteDetail.tsx`, `web/src/pages/Websites/WebsiteList.tsx`, `web/src/pages/Websites/CreateWebsite.tsx`, `web/src/pages/Settings/Settings.tsx` |

### Flow today
1. Creating a website creates a `Zone` row and one `Domain` row (type `primary`).
2. Extra domains can be added under a website.
3. DNS records live under a zone.
4. Any record/domain change calls `dns.zone.update` / `dns.zone.sync_records` on the agent **fire-and-forget**.
5. The agent writes `/etc/juvia/bind/zones/<domain>.zone`, updates `/etc/bind/named.conf.local`, validates, and reloads BIND.

---

## 2. Bugs & Gaps

### 2.1 Backend / API

| # | Issue | Severity | File:Line |
|---|-------|----------|-----------|
| 1 | **Silent agent failures** — `syncZoneRecords`, domain create/delete, and website create call `cfg.AgentClient.Call(...)` but discard the returned `*Response` and `error`. A zone can be saved in SQLite but never written to disk. | High | `internal/api/dns.go:378`, `internal/api/domains.go:103`, `internal/api/websites.go:192` |
| 2 | **Non-incrementing serial** — serial is computed as `time.Now().Unix() / 86400`, so all changes within the same day share the same serial. BIND caches zone data by serial; clients will not refresh. | High | `internal/api/dns.go:386`, `internal/agent/dns.go:325` |
| 3 | **Zone not updated on restore** — restoring a website from trash re-creates the web config but does not recreate the DNS zone. | High | `internal/api/websites.go:427` |
| 4 | **Alias domains are invisible to DNS** — adding an alias domain only updates the web server `extra_domains`; it does not add A/AAAA records for the alias in the primary zone, nor create a separate zone. | High | `internal/api/domains.go:87` |
| 5 | **No atomic DB + filesystem consistency** — if agent fails after DB commit, the DB and filesystem diverge with no retry/reconciliation. | High | multiple |
| 6 | **Permission leakage** — zone files are written with `0644` (world-readable). BIND accepts `0640` or `0600`. | Med | `internal/agent/dns.go:340`, `514` |
| 7 | **Hard-coded email DNS** — every generated zone always creates `mail.<domain>` A + MX + SPF + DMARC records, even when the mail module is not installed or the user disabled email. | Med | `internal/agent/dns.go:372-469` |
| 8 | `digRecord` returns only the first line of `dig` output, so multi-value NS/MX/TXT checks are incomplete. | Med | `internal/agent/dns.go:217` |
| 9 | `HandleDNSVerify` checks only the first NS returned by `dig`, and expects `www` to be a CNAME even if the user changed it to an A record. | Med | `internal/agent/dns.go:98-108` |
| 10 | **Missing distros handling** — `named` vs `bind9` service names are partially handled in service restart but not in `reloadBind()` or `isBindActive()`. | Med | `internal/agent/dns.go:289`, `614` |
| 11 | `getServerIP()` in `agent/dns.go` shells out to `curl ifconfig.me` instead of reusing `discoverPublicIP()` which has fallbacks. | Low | `internal/agent/dns.go:628` |
| 12 | `validateNamedConf()` and `reloadBind()` assume BIND is installed; on a fresh dev/test box the agent panics instead of returning a clean error. | Med | `internal/agent/dns.go:605`, `614` |
| 13 | `named.conf.local` is edited in-place without backup. A malformed write can break the system BIND config. | Med | `internal/agent/dns.go:748` |
| 14 | `CreateWebsite` creates a zone row but does not verify the agent actually wrote the zone file before reporting success. | High | `internal/api/websites.go:144-195` |
| 15 | `deleteWebsiteHandler` (soft delete) does not remove the zone; only permanent delete does. A suspended/trashed site still resolves. | Med | `internal/api/websites.go:342` |
| 16 | `suggestDNSHandler` suggests `_dmarc` with a hard-coded `example.com` email address. | Low | `internal/api/dns.go:316` |
| 17 | Settings `validateSetting` forbids `_` in `ns1_hostname`/`ns2_hostname`, which is technically correct for hostnames, but then the UI auto-generates `ns1.<brand>` which is fine. However it also forbids hyphens at start/end but not in middle — acceptable. | Low | `internal/api/settings.go:124` |

### 2.2 Database / Schema

| # | Issue | Severity |
|---|-------|----------|
| 18 | `zones.serial` defaults to `1` but agent writes Unix day numbers; mixing formats. | High |
| 19 | `dns_records` has no `ttl` column, yet the frontend renders a TTL column. | Med |
| 20 | Missing unique constraint on `(zone_id, type, name, value)` allowing duplicate records. | Med |
| 21 | Missing index on `zones.website_id` (foreign key without index). | Low |
| 22 | Domain conflict check only prevents exact duplicate domains; it does not prevent an alias domain from colliding with another website's primary domain or wildcard coverage. | Med |

### 2.3 Frontend

| # | Issue | Severity | File:Line |
|---|-------|----------|-----------|
| 23 | **Hard-coded domain in UI** — DNS health card shows literal `ns1.12121997.xyz` as an example. | Med | `web/src/pages/DNS/DnsManagement.tsx:266` |
| 24 | **TTL column shows nothing** — `DNSRecord` interface includes `ttl` but API never returns it. | Med | `web/src/pages/Websites/WebsiteDetail.tsx:1086` |
| 25 | Add-record modal requires `name` to be non-empty, but `@` for root is allowed and common. (Currently works only if user types `@`; empty string is blocked.) | Low | `web/src/pages/Websites/WebsiteDetail.tsx:1168` |
| 26 | Edit modal allows editing the record `name`, which can silently change a `www` record into a different name while keeping the same `id`. Backend accepts this, but UX is confusing. | Low | `web/src/pages/Websites/WebsiteDetail.tsx:1176` |
| 27 | No in-UI indication when a zone sync failed on the agent; the record appears saved in the table. | High | `web/src/pages/Websites/WebsiteDetail.tsx:981` |
| 28 | DNS Management page lists zones but has no way to view, edit, or delete a zone's records inline. | Med | `web/src/pages/DNS/DnsManagement.tsx` |
| 29 | CreateWebsite fetches nameservers but never uses them to pre-fill DNS guidance. | Low | `web/src/pages/Websites/CreateWebsite.tsx` |

### 2.4 Testing

| # | Issue | Severity |
|---|-------|----------|
| 30 | No unit tests for agent DNS handlers (`HandleDNSVerify`, `HandleZoneSyncRecords`, etc.). | High |
| 31 | No unit/integration tests for API DNS/domain handlers. | High |
| 32 | No tests verifying zone file output format or serial increment behavior. | High |

---

## 3. Proposed Implementation Plan

### Phase A — Harden the Core Sync (backend + agent)

1. **Introduce a zone-sync result type**
   - Add `internal/agent/dns.go` response struct: `ZoneSyncResult{ZonePath, Serial, Reloaded bool, Message string}`.
   - Return it from `HandleZoneSyncRecords`, `HandleZoneCreate`, `HandleZoneUpdate`.

2. **Make every caller check agent errors**
   - Change `syncZoneRecords` to return `error`.
   - In `createDNSRecordHandler`, `updateDNSRecordHandler`, `deleteDNSRecordHandler`, `createDomainHandler`, `deleteDomainHandler`, `createWebsiteHandler`: if sync fails, return `500` and do **not** commit record changes (do DB write inside a single logical unit or rollback-friendly pattern).
   - For website creation, perform zone sync **after** `website.create` agent call succeeds and before returning `201`.

3. **Correct serial generation**
   - Store serial as `YYYYMMDDNN` (RFC 1912 style) in `zones.serial`.
   - On every sync, compute new serial: if today > current serial's date part, use `YYYYMMDD01`; else increment `NN`.
   - Persist the new serial to `zones.serial` before calling agent.

4. **Zone file hygiene**
   - Write tmp files with `0600`, final zone files with `0640` owner `root:bind`.
   - Back up `named.conf.local` to `named.conf.local.juvia-backup` before first modification.
   - Detect missing BIND (`named-checkzone`, `named-checkconf`, `systemctl`) and return a clear error instead of failing the whole agent.

5. **Service-name portability**
   - Add `resolveBindServiceName()` that checks for `bind9` then `named`.
   - Use it in `reloadBind`, `isBindActive`, and health checks.

### Phase B — Domain Model Improvements

6. **Alias / subdomain DNS handling**
   - When `createDomainHandler` classifies a domain as `alias` (different root), create a new zone for that alias and call `dns.zone.update` for it.
   - When classifying as `subdomain` (under primary), add an A/CNAME record to the primary zone.
   - On delete, remove the matching zone or record.

7. **Domain conflict expansion**
   - `checkDomainConflict` must also reject a domain that is the primary domain of any website, and reject wildcard overlaps across all websites.

8. **Schema updates** (new migration `006_dns_hardening.sql`)
   - Add `ttl INTEGER DEFAULT 3600` to `dns_records`.
   - Add unique index `idx_dns_records_unique` on `(zone_id, type, name, value)` (allow NULL priority).
   - Add index `idx_zones_website` on `zones(website_id)`.

### Phase C — Frontend Fixes

9. **Remove hard-coded values**
   - Replace `ns1.12121997.xyz` in `DnsManagement.tsx` with dynamic `ns1.${nsBrandDomain}`.

10. **Display TTL from API**
    - Update `DNSRecord` interface to include `ttl`, render it, and default to `3600`.

11. **Surface sync failures**
    - After add/edit/delete, if API returns an agent error, show a persistent alert with the error message and a "Retry sync" action.

12. **Add inline zone management to DNS page**
    - Click a zone row to open a modal showing/editing its records (reuse `DNSTab` logic or create `ZoneDetailModal`).
    - Add "Delete zone" action with confirmation.

13. **CreateWebsite DNS guidance**
    - Show nameserver hostnames and A-record instructions on the review step.

### Phase D — Verification & Observability

14. **Improve `HandleDNSVerify`**
    - Parse all NS/MX/TXT answers from `dig`.
    - Accept either CNAME or A for `www`.
    - Support wildcard queries.
    - Return per-record status plus propagation advice.

15. **Optional mail DNS**
    - Add a setting `dns_auto_mail_records` default `true`.
    - Only emit `mail`, `MX`, `SPF`, `DMARC` defaults when enabled.

16. **Reconciliation endpoint**
    - `POST /dns/reconcile` (admin) iterates all zones, regenerates zone files from DB records, and reloads BIND. Useful after agent failure or manual recovery.

### Phase E — Testing

17. **Agent DNS unit tests**
    - Test `generateZoneFromRecords` output for all record types.
    - Test serial increment logic.
    - Test `HandleZoneSyncRecords` with mocked `named-checkzone`/`named-checkconf` by injecting command runners.

18. **API integration tests**
    - Create website → zone created → record added → zone file sync succeeds.
    - Duplicate domain rejected.
    - Agent failure returns proper error and does not leave inconsistent DB state.

19. **Frontend component tests**
    - `DnsManagement` renders nameservers from API.
    - `DNSTab` validates record input and calls correct endpoints.

---

## 4. Success Criteria

- `go test ./...` passes, including new DNS tests.
- `go build ./...` succeeds.
- Creating/editing/deleting a DNS record always either succeeds in both DB and BIND or returns an explicit error with no divergence.
- Zone serial increments monotonically on every change.
- Alias domains get their own zone or a subdomain record, depending on classification.
- Frontend no longer contains hard-coded example domains and displays TTL.

---

## 5. Suggested Execution Order

1. Phase A (sync hardening + serial fix) — highest impact, unblocks reliability.
2. Phase D verification improvements (better `dig` parsing, optional mail records).
3. Phase B schema + domain model changes.
4. Phase C frontend polish.
5. Phase E tests throughout.

---

## 6. Open Questions for Stakeholder

1. Should **alias domains** (completely different root, e.g. `another.com` on `site1.com`) create a separate BIND zone, or should Juvia only support subdomains/wildcards under the primary domain?
2. Should trashed/suspended websites continue to resolve via DNS, or should the zone be removed on soft-delete and restored on restore?
3. Do we need to support **slave/secondary DNS** (AXFR/NOTIFY) or only single-master BIND?
4. Should the panel operate as a **hidden master** (NS records point to external provider) or as the public authoritative nameserver?

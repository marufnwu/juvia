# Frontend DNS UI — Remaining Limitations & Improvement Plan

## Context

The backend DNS implementation (agent zone sync, serial numbers, TTL support, health checks) is now deployed and functional. The frontend has basic DNS record management per website and a DNS health overview page, but several UX gaps remain that limit usability and observability.

---

## Remaining Limitations

| # | Limitation | Severity | Where |
|---|------------|----------|-------|
| 1 | **DNSTab empty state is misleading** when no zone exists. It shows "No DNS records found" instead of explaining that the zone must be created first. | Medium | `WebsiteDetail.tsx` DNSTab |
| 2 | **No manual zone creation UI.** Zones are only created automatically when a website is created; orphaned or manual zones cannot be added from the DNS page. | Medium | `DnsManagement.tsx` |
| 3 | **DNS Management page is read-only for zones.** The zones table has only a Copy button; there is no View Records, Edit Records, or Delete Zone action. | Medium | `DnsManagement.tsx` |
| 4 | **Sync failures are transient.** When a record add/edit/delete fails on the agent (e.g. BIND not running), only a toast appears. There is no persistent banner with a Retry action. | High | `WebsiteDetail.tsx` DNSTab |
| 5 | **DNS health does not auto-run.** Users must click "Check Health" every time they open the page. | Low | `DnsManagement.tsx` |
| 6 | **No DNSSEC support in UI.** No toggle for DNSSEC, no DS/DNSKEY record display, no key rollover. | Medium | Entire DNS stack |
| 7 | **No secondary/slave DNS UI.** Cannot configure external secondary nameservers or AXFR/NOTIFY. | Low | Entire DNS stack |
| 8 | **No reconciliation action.** Backend could support `POST /dns/reconcile`, but there is no UI button to rebuild all zone files from the DB. | Medium | `DnsManagement.tsx` |
| 9 | **No nameserver setup wizard.** New users must manually navigate to Settings → Nameservers to configure NS hostnames and brand domain. | Medium | Onboarding flow |
| 10 | **Frontend record validation is basic.** IPv6 only checks for a colon; CNAME/MX/NS do not validate trailing-dot conventions. | Low | `WebsiteDetail.tsx` |
| 11 | **No zone serial/SOA visibility in DNSTab.** Users cannot see the current serial or zone metadata. | Low | `WebsiteDetail.tsx` |
| 12 | **No bulk import/export.** Cannot upload a BIND zone file or export zone records. | Low | `DnsManagement.tsx`, DNSTab |

---

## Proposed Implementation Order

### Phase 1 — Reliability & Clarity (High Value, Low Risk)
1. **Improve DNSTab empty state**
   - Detect whether a zone exists for the website.
   - If no zone: show "No DNS zone exists for this website. Create one to manage records." with a "Create Zone" button.
   - If zone exists but no records: show "No DNS records yet. Add an A record pointing to the server IP."
2. **Surface sync failures persistently**
   - Add a `syncError` state to DNSTab.
   - On add/edit/delete failure, show a warning alert with the agent error message and a "Retry Sync" button that re-submits the last operation.
3. **Add zone metadata header in DNSTab**
   - Display zone domain, serial, and last-updated time above the records table.

### Phase 2 — DNS Management Page Actions
4. **Manual zone creation**
   - Add a "Create Zone" button on `DnsManagement.tsx`.
   - Modal with domain input; calls a new backend endpoint `POST /dns/zones`.
   - Requires a corresponding backend handler that inserts a `zones` row and calls `dns.zone.update` on the agent.
5. **Inline zone actions**
   - Add "View Records" button per zone row that opens a modal listing/editing records (reuse DNSTab component or a new `ZoneRecordsModal`).
   - Add "Delete Zone" button with confirmation; calls `DELETE /dns/zones/:id` or agent `dns.zone.delete`.
6. **Reconcile zones**
   - Add a "Rebuild All Zones" button (admin only) that calls `POST /dns/reconcile`.
   - Requires backend endpoint that iterates zones and calls `dns.zone.sync_records`.

### Phase 3 — Onboarding & Automation
7. **Nameserver setup wizard**
   - If `ns1_hostname`/`ns2_hostname` are not set, show a prominent banner on the DNS page linking to a setup wizard.
   - Wizard collects brand domain, generates `ns1/ns2.<brand>`, sets `server_ip`, and validates with `dns.verify_nameservers`.
8. **Auto-run DNS health**
   - Run `checkHealth()` automatically when `DnsManagement.tsx` mounts if nameservers are configured.
9. **Strengthen frontend validation**
   - Use a proper IPv6 regex; validate CNAME/MX/NS values end with a dot when they are absolute, warn when not.

### Phase 4 — Advanced Features
10. **DNSSEC UI**
    - Add setting to enable DNSSEC per zone; display DS record; manage KSK/ZSK lifecycle.
    - Requires agent support for `dnssec-keygen`, `dnssec-signzone`, and updated `named.conf.local`.
11. **Secondary DNS UI**
    - Allow adding secondary NS hostnames and TSIG keys; configure `also-notify` and `allow-transfer`.
12. **Bulk import/export**
    - Export zone as BIND file; import via textarea/file upload.

---

## Open Questions

1. Should **Phase 2 (manual zone create/delete)** be allowed for zones that are not linked to a website, or only as a safety net for website-linked zones?
2. Is **DNSSEC** a requirement for this deployment, or can it remain a future enhancement?
3. Should the **reconcile action** be available to all admins or restricted to a "super admin" role?
4. Do we want to support **bulk zone import** from a BIND `named.conf.local` file, or only per-zone import?

---

## Success Criteria

- A user can create a DNS zone manually from the DNS page when needed.
- DNSTab clearly explains the difference between "no zone" and "zone with no records".
- Agent sync failures show a persistent, retryable alert.
- DNS Management page supports viewing and deleting zones.
- Health check runs automatically on page load.


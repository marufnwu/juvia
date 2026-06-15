-- Add unique index on domains.domain to prevent duplicate domain assignments
CREATE UNIQUE INDEX IF NOT EXISTS idx_domains_domain_unique ON domains(domain);

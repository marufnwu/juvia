-- DNS hardening: TTL column, unique record index, zone website index
ALTER TABLE dns_records ADD COLUMN ttl INTEGER DEFAULT 3600;

CREATE UNIQUE INDEX IF NOT EXISTS idx_dns_records_unique ON dns_records(zone_id, type, name, value);

CREATE INDEX IF NOT EXISTS idx_zones_website ON zones(website_id);

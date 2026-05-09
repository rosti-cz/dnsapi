CREATE TABLE IF NOT EXISTS zones (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME,
    updated_at DATETIME,
    "delete" BOOLEAN NOT NULL DEFAULT 0,
    domain TEXT,
    serial TEXT,
    tags TEXT,
    abuse_email TEXT
);

CREATE TABLE IF NOT EXISTS records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME,
    updated_at DATETIME,
    zone_id INTEGER,
    name TEXT,
    ttl INTEGER,
    type TEXT,
    prio INTEGER,
    value TEXT,
    FOREIGN KEY(zone_id) REFERENCES zones(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_zones_domain ON zones(domain);
CREATE INDEX IF NOT EXISTS idx_records_zone_id ON records(zone_id);
CREATE INDEX IF NOT EXISTS idx_records_zone_record ON records(zone_id, id);

UPDATE zones
SET domain = lower(trim(domain))
WHERE domain IS NOT NULL;

UPDATE zones
SET tags = trim(replace(replace(tags, ', ', ','), ' ,', ','), ',')
WHERE tags IS NOT NULL;

UPDATE zones
SET serial = strftime('%Y%m%d', 'now') || '01'
WHERE serial IS NULL OR length(serial) != 10;

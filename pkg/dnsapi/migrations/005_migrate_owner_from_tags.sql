-- If owner is empty and tags contain an "owner:NNNN" entry, set owner to "companyNNNN".
UPDATE zones
SET owner = 'company' ||
  CASE
    WHEN INSTR(SUBSTR(tags, INSTR(tags, 'owner:') + 6), ',') > 0
    THEN SUBSTR(SUBSTR(tags, INSTR(tags, 'owner:') + 6), 1, INSTR(SUBSTR(tags, INSTR(tags, 'owner:') + 6), ',') - 1)
    ELSE TRIM(SUBSTR(tags, INSTR(tags, 'owner:') + 6))
  END
WHERE owner = '' AND INSTR(tags, 'owner:') > 0;

DROP INDEX IF EXISTS idx_properties_status;
DROP INDEX IF EXISTS idx_properties_origin;
DROP INDEX IF EXISTS idx_properties_last_sync_id;

ALTER TABLE properties DROP COLUMN IF EXISTS status;
ALTER TABLE properties DROP COLUMN IF EXISTS origin;
ALTER TABLE properties DROP COLUMN IF EXISTS last_sync_id;

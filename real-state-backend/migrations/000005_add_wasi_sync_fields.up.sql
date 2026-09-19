ALTER TABLE properties ADD COLUMN last_sync_id VARCHAR(50);
ALTER TABLE properties ADD COLUMN origin VARCHAR(20) DEFAULT 'LOCAL';
ALTER TABLE properties ADD COLUMN status VARCHAR(20) DEFAULT 'ACTIVE';

CREATE INDEX idx_properties_last_sync_id ON properties(last_sync_id);
CREATE INDEX idx_properties_origin ON properties(origin);
CREATE INDEX idx_properties_status ON properties(status);

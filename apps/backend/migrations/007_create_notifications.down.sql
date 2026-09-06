DROP TABLE IF EXISTS notifications;
ALTER TABLE incidents DROP COLUMN IF EXISTS closed_by;
ALTER TABLE incidents DROP COLUMN IF EXISTS closed_at;

-- This data migration reverts 'almost private' posts back to the old
-- 'private' status. This is the rollback for the 000015 up migration.
UPDATE posts SET privacy = 'private' WHERE privacy = 'almost private';

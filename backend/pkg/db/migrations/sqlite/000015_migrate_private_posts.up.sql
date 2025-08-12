-- This data migration reclassifies old 'private' (followers-only) posts
-- to the new 'almost private' status to maintain backward compatibility.
UPDATE posts SET privacy = 'almost private' WHERE privacy = 'private';

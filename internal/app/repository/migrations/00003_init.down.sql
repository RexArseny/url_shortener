START TRANSACTION;

ALTER TABLE users RENAME COLUMN deleted TO __deleted;

COMMIT;
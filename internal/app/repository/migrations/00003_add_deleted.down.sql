START TRANSACTION;

ALTER TABLE users RENAME COLUMN deleted TO __deleted;

DROP TABLE urls_for_delete;

COMMIT;
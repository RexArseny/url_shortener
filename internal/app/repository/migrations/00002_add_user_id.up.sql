START TRANSACTION;

ALTER TABLE urls ADD user_id uuid NOT NULL;

COMMIT;
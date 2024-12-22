START TRANSACTION;

ALTER TABLE urls ADD deleted bool NOT NULL DEFAULT false;

CREATE TABLE
  IF NOT EXISTS urls_for_delete (
    id SERIAL PRIMARY KEY,
    urls text[] NOT NULL,
    user_id uuid NOT NULL
  );

COMMIT;
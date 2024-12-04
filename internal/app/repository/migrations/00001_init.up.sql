START TRANSACTION;

CREATE TABLE
  IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    short_url text NOT NULL,
    original_url text NOT NULL,
    CONSTRAINT short_url_constraint UNIQUE(short_url),
    CONSTRAINT original_url_constraint UNIQUE(original_url)
  );

COMMIT;
CREATE SCHEMA palakit;

CREATE TABLE palakit.database_variable (
  name TEXT PRIMARY KEY,
  value TEXT NULL
);

CREATE FUNCTION palakit.auth(key TEXT)
RETURNS TEXT
LANGUAGE plpgsql
AS $$
  DECLARE
    value TEXT;
  BEGIN
    IF EXISTS (
      SELECT
      FROM information_schema.tables 
      WHERE table_type = 'LOCAL TEMPORARY'
      AND table_name = 'authorization_variable'
    ) THEN
      SELECT authorization_variable.value
      INTO value
      FROM pg_temp.authorization_variable
      WHERE name = key;
      RETURN value;
    ELSE
      return NULL;
    END IF;
  END;
$$;

GRANT USAGE ON SCHEMA palakit TO public;
GRANT EXECUTE ON FUNCTION palakit.auth TO PUBLIC;
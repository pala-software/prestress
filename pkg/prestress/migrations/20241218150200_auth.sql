CREATE FUNCTION prestress.auth(key TEXT)
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
      SELECT var.value
      INTO value
      FROM pg_temp.authorization_variable AS var
      WHERE name = key;
      RETURN value;
    ELSE
      return NULL;
    END IF;
  END;
$$;

CREATE FUNCTION prestress.begin_authorized(variables jsonb)
RETURNS VOID
LANGUAGE plpgsql
AS $$
  BEGIN
    CREATE TEMPORARY TABLE pg_temp.authorization_variable
      (name TEXT PRIMARY KEY, value TEXT)
    ON COMMIT DROP;

    INSERT INTO pg_temp.authorization_variable
    SELECT key AS name, value
    FROM jsonb_each_text(variables);
  END;
$$;

CREATE FUNCTION prestress.end_authorized()
RETURNS VOID
LANGUAGE plpgsql
AS $$
  BEGIN
    DROP TABLE pg_temp.authorization_variable;
  END;
$$;

GRANT USAGE ON SCHEMA prestress TO public;
GRANT EXECUTE ON FUNCTION prestress.auth TO public;
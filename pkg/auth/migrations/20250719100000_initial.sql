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
      AND table_name = 'prestress_auth'
    ) THEN
      SELECT var.value
      INTO value
      FROM pg_temp.prestress_auth AS var
      WHERE var.key = auth.key;
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
    CREATE TEMPORARY TABLE pg_temp.prestress_auth
      (key TEXT PRIMARY KEY, value TEXT)
    ON COMMIT DROP;

    INSERT INTO pg_temp.prestress_auth
    SELECT key, value
    FROM jsonb_each_text(variables);
  END;
$$;

CREATE FUNCTION prestress.end_authorized()
RETURNS VOID
LANGUAGE plpgsql
AS $$
  BEGIN
    DROP TABLE pg_temp.prestress_auth;
  END;
$$;

CREATE FUNCTION prestress.dump_authorization()
RETURNS jsonb
LANGUAGE plpgsql
AS $$
  DECLARE
    variables jsonb;
  BEGIN
    IF EXISTS (
      SELECT
      FROM information_schema.tables
      WHERE table_type = 'LOCAL TEMPORARY'
      AND table_name = 'prestress_auth'
    ) THEN
      SELECT jsonb_object(array_agg(t.key), array_agg(t.value))
        FROM pg_temp.prestress_auth AS t
        INTO variables;
    END IF;
    RETURN variables;
  END;
$$;

GRANT USAGE ON SCHEMA prestress TO public;
GRANT EXECUTE ON FUNCTION prestress.auth TO public;

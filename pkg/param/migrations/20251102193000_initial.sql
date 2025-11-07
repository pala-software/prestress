CREATE FUNCTION prestress.param(key TEXT, def TEXT)
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
      AND table_name = 'prestress_param'
    ) THEN
      IF EXISTS(
        SELECT
        FROM pg_temp.prestress_param AS var
        WHERE var.key = param.key
      ) THEN
        SELECT var.value
        INTO value
        FROM pg_temp.prestress_param AS var
        WHERE var.key = param.key;
        RETURN value;
      ELSE
        return def;
      END IF;
    ELSE
      return def;
    END IF;
  END;
$$;

CREATE FUNCTION prestress.set_params(params jsonb)
RETURNS VOID
LANGUAGE plpgsql
AS $$
  BEGIN
    CREATE TEMPORARY TABLE pg_temp.prestress_param
      (key TEXT PRIMARY KEY, value TEXT)
    ON COMMIT DROP;

    INSERT INTO pg_temp.prestress_param
    SELECT key, value
    FROM jsonb_each_text(params);
  END;
$$;

GRANT USAGE ON SCHEMA prestress TO public;
GRANT EXECUTE ON FUNCTION prestress.param TO public;

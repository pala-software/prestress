CREATE SCHEMA auth;

CREATE FUNCTION auth.uid()
RETURNS TEXT
LANGUAGE plpgsql
AS $$
  DECLARE
    uid TEXT;
  BEGIN
    IF EXISTS (
      SELECT
      FROM information_schema.tables 
      WHERE table_type = 'LOCAL TEMPORARY'
      AND table_name = 'variable'
    ) THEN
      SELECT value
      INTO uid
      FROM pg_temp.variable
      WHERE name = 'uid';
      RETURN uid;
    ELSE
      return NULL;
    END IF;
  END;
$$;
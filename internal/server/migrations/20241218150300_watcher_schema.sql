CREATE SCHEMA watcher;

CREATE TYPE watcher.operation
AS ENUM ('INSERT', 'UPDATE', 'DELETE');

CREATE TABLE watcher.subscription (
  id SERIAL PRIMARY KEY,
  role_name NAME NOT NULL,
  table_schema NAME NOT NULL,
  table_name NAME NOT NULL
);

CREATE TABLE watcher.change (
  subscription_id INTEGER NOT NULL REFERENCES watcher.subscription(id) ON DELETE CASCADE,
  row_key JSONB NULL,
  row_data JSONB NOT NULL,
  row_operation watcher.operation NOT NULL
);

CREATE FUNCTION watcher.get_primary_key(
  table_schema NAME,
  table_name NAME,
  row_data JSONB)
RETURNS JSONB
LANGUAGE plpgsql
AS $$
  DECLARE
    table_id OID;
    key_columns NAME[];
    primary_key JSONB;
  BEGIN
    SELECT pg_class.oid
      INTO table_id
      FROM pg_class
      JOIN pg_namespace
      ON pg_namespace.oid = pg_class.relnamespace
      WHERE
        pg_namespace.nspname = get_primary_key.table_schema AND
        pg_class.relname = get_primary_key.table_name;
    SELECT array_agg(pg_attribute.attname)
      INTO key_columns
      FROM pg_index
      JOIN pg_attribute
      ON pg_attribute.attnum = ANY (pg_index.indkey)
      WHERE
        pg_index.indisprimary = TRUE AND
        pg_index.indrelid = table_id AND
        pg_attribute.attrelid = table_id;
    EXECUTE format('SELECT jsonb_build_array(%s)
      FROM (SELECT $1 AS row_data)',
      (SELECT string_agg('row_data->' || quote_literal(column_name), ',')
        FROM unnest(key_columns) AS column_name))
    INTO primary_key
    USING row_data;
    RETURN primary_key;
  END;
$$;

CREATE FUNCTION watcher.get_related_tables(source_schema NAME, source_table NAME)
RETURNS TABLE (table_schema NAME, table_name NAME)
LANGUAGE sql
AS $$
  WITH RECURSIVE tab(table_id) AS (
    SELECT pg_class.oid AS table_id
    FROM pg_class
    JOIN pg_namespace ON pg_namespace.oid = pg_class.relnamespace
    WHERE
      pg_class.relname = source_table AND
      pg_namespace.nspname = source_schema
    UNION
    SELECT pg_class.oid AS table_id
    FROM tab
    JOIN pg_depend ON pg_depend.refobjid = table_id
    LEFT JOIN pg_policy ON pg_policy.oid = pg_depend.objid
    LEFT JOIN pg_rewrite ON pg_rewrite.oid = pg_depend.objid
    JOIN pg_class ON
      pg_class.oid = pg_policy.polrelid OR
      pg_class.oid = pg_rewrite.ev_class)
  SELECT
    pg_namespace.nspname AS table_schema,
    pg_class.relname AS table_name
  FROM tab
  JOIN pg_class ON pg_class.oid = tab.table_id
  JOIN pg_namespace ON pg_namespace.oid = pg_class.relnamespace;
$$;

CREATE FUNCTION watcher.extract_change(
  subscription_id INTEGER,
  table_schema NAME,
  table_name NAME,
  operation watcher.operation)
RETURNS SETOF watcher.change
LANGUAGE plpgsql
AS $$
  DECLARE
    changed_rows JSONB[];
    diff_query TEXT := 'SELECT array_agg(to_jsonb(a))
      FROM %s AS a
      WHERE a NOT IN (SELECT b FROM %s AS b)';
    state_table_name NAME := 'ws_' || subscription_id;
  BEGIN
    IF operation = 'INSERT' THEN
      EXECUTE format(
        diff_query,
        quote_ident(table_schema) || '.' ||
        quote_ident(table_name),
        quote_ident(state_table_name))
      INTO changed_rows;
    ELSIF operation = 'UPDATE' THEN
      RAISE EXCEPTION 'Operation UPDATE for watcher.extract_change without
        key_columns parameter is not implemented';
    ELSIF operation = 'DELETE' THEN
      EXECUTE format(
        diff_query,
        quote_ident(state_table_name),
        quote_ident(table_schema) || '.' ||
        quote_ident(table_name))
      INTO changed_rows;
    END IF;
    RETURN QUERY SELECT
      subscription_id,
      watcher.get_primary_key(
        table_schema,
        table_name,
        row_data)
        AS row_key,
      row_data,
      operation AS row_operation
    FROM unnest(changed_rows)
    AS row_data;
  END;
$$;

CREATE FUNCTION watcher.extract_change(
  subscription_id INTEGER,
  table_schema NAME,
  table_name NAME,
  operation watcher.operation,
  key_columns NAME[])
RETURNS SETOF watcher.change
LANGUAGE plpgsql
AS $$
  DECLARE
    changed_rows JSONB[];
    diff_query TEXT := 'SELECT array_agg(to_jsonb(a))
      FROM %s AS a
      WHERE jsonb_build_array(%s) %s (
        SELECT jsonb_build_array(%s) FROM %s AS b)';
    state_table_name NAME := 'ws_' || subscription_id;
  BEGIN
    IF operation = 'INSERT' THEN
      EXECUTE format(
        diff_query,
        quote_ident(table_schema) || '.' ||
        quote_ident(table_name),
        (SELECT string_agg('a.'|| quote_ident(column_name), ',')
          FROM unnest(key_columns) AS column_name),
        'NOT IN',
        (SELECT string_agg('b.'|| quote_ident(column_name), ',')
          FROM unnest(key_columns) AS column_name),
        quote_ident(state_table_name))
      INTO changed_rows;
    ELSIF operation = 'UPDATE' THEN
      EXECUTE format(
        diff_query,
        quote_ident(table_schema) || '.' ||
        quote_ident(table_name),
        (SELECT string_agg('a.'|| quote_ident(column_name), ',')
          FROM unnest(key_columns) AS column_name),
        'IN',
        (SELECT string_agg('b.'|| quote_ident(column_name), ',')
          FROM unnest(key_columns) AS column_name),
        quote_ident(state_table_name))
      INTO changed_rows;
    ELSIF operation = 'DELETE' THEN
      EXECUTE format(
        diff_query,
        quote_ident(state_table_name),
        (SELECT string_agg('a.'|| quote_ident(column_name), ',')
          FROM unnest(key_columns) AS column_name),
        'NOT IN',
        (SELECT string_agg('b.'|| quote_ident(column_name), ',')
          FROM unnest(key_columns) AS column_name),
        quote_ident(table_schema) || '.' ||
        quote_ident(table_name))
      INTO changed_rows;
    END IF;
    RETURN QUERY SELECT
      subscription_id,
      watcher.get_primary_key(
        table_schema,
        table_name,
        row_data)
        AS row_key,
      row_data,
      operation AS row_operation
    FROM unnest(changed_rows)
    AS row_data;
  END;
$$;

CREATE FUNCTION watcher.record_state(
  subscription_id INTEGER,
  table_schema NAME,
  table_name NAME)
RETURNS VOID
LANGUAGE plpgsql
AS $$
  BEGIN
    EXECUTE format('CREATE TEMPORARY TABLE %I
      ON COMMIT DROP
      AS SELECT * FROM %I.%I',
      'ws_' || subscription_id,
      table_schema,
      table_name);
  END;
$$;

CREATE FUNCTION watcher.record_changes(
  subscription_id INTEGER,
  table_schema NAME,
  table_name NAME)
RETURNS VOID
LANGUAGE plpgsql
AS $$
  DECLARE
    table_id OID;
    key_columns NAME[];
  BEGIN
    SELECT pg_class.oid
      INTO table_id
      FROM pg_class
      JOIN pg_namespace
      ON pg_namespace.oid = pg_class.relnamespace
      WHERE
        pg_namespace.nspname = table_schema AND
        pg_class.relname = table_name;
    SELECT array_agg(pg_attribute.attname)
      INTO key_columns
      FROM pg_index
      JOIN pg_attribute
      ON pg_attribute.attnum = ANY (pg_index.indkey)
      WHERE
        pg_index.indisprimary = TRUE AND
        pg_index.indrelid = table_id AND
        pg_attribute.attrelid = table_id;

    IF array_length(key_columns, 1) > 0 THEN
      INSERT INTO watcher.change
      SELECT *
      FROM watcher.extract_change(
        subscription_id,
        table_schema,
        table_name,
        'INSERT',
        key_columns)
      UNION
      SELECT *
      FROM watcher.extract_change(
        subscription_id,
        table_schema,
        table_name,
        'UPDATE',
        key_columns)
      UNION
      SELECT *
      FROM watcher.extract_change(
        subscription_id,
        table_schema,
        table_name,
        'DELETE',
        key_columns);
    ELSE
      INSERT INTO watcher.change
      SELECT *
      FROM watcher.extract_change(
        subscription_id,
        table_schema,
        table_name,
        'INSERT')
      UNION
      SELECT *
      FROM watcher.extract_change(
        subscription_id,
        table_schema,
        table_name,
        'DELETE');
    END IF;
    NOTIFY change;
  END;
$$;

CREATE FUNCTION watcher.setup_subscription()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
  DECLARE
    original_role NAME := CURRENT_USER;
    table_schema NAME;
    table_name NAME;
  BEGIN
    EXECUTE format('SET LOCAL ROLE TO %I', NEW.role_name);

    EXECUTE format('CREATE FUNCTION pg_temp.%I()
      RETURNS TRIGGER
      LANGUAGE plpgsql
      SECURITY DEFINER
      AS $s$
        BEGIN
          PERFORM watcher.record_state(%L, %L, %L);
          RETURN NULL;
        END;
      $s$;',
      'wb_' || NEW.id,
      NEW.id,
      NEW.table_schema,
      NEW.table_name);

    EXECUTE format('CREATE FUNCTION pg_temp.%I()
      RETURNS TRIGGER
      LANGUAGE plpgsql
      SECURITY DEFINER
      AS $s$
        BEGIN
          PERFORM watcher.record_changes(%L, %L, %L);
          RETURN NULL;
        END;
      $s$;',
      'wa_' || NEW.id,
      NEW.id,
      NEW.table_schema,
      NEW.table_name);

    EXECUTE format('SET LOCAL ROLE TO %I', original_role);

    FOR table_schema, table_name IN
      SELECT related_table.table_schema, related_table.table_name
      FROM watcher.get_related_tables(NEW.table_schema, NEW.table_name)
      AS related_table
    LOOP
      EXECUTE format('CREATE TRIGGER %I
        BEFORE INSERT OR UPDATE OR DELETE ON %I.%I
        FOR EACH STATEMENT
        EXECUTE FUNCTION pg_temp.%I();',
        'wb_' || NEW.id,
        NEW.table_schema,
        NEW.table_name,
        'wb_' || NEW.id);

      EXECUTE format('CREATE TRIGGER %I
        AFTER INSERT OR UPDATE OR DELETE ON %I.%I
        FOR EACH STATEMENT
        EXECUTE FUNCTION pg_temp.%I();',
        'wa_' || NEW.id,
        NEW.table_schema,
        NEW.table_name,
        'wa_' || NEW.id);
    END LOOP;

    RETURN NEW;
  END;
$$;

CREATE FUNCTION watcher.teardown_subscription()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
  BEGIN
    EXECUTE format('DROP FUNCTION pg_temp.%I() CASCADE;',
      'wb_' || OLD.id);
    EXECUTE format('DROP FUNCTION pg_temp.%I() CASCADE;',
      'wa_' || OLD.id);
    RETURN NEW;
  END;
$$;

CREATE TRIGGER setup_subscription
BEFORE INSERT ON watcher.subscription
FOR EACH ROW
EXECUTE FUNCTION watcher.setup_subscription();

CREATE TRIGGER teardown_subscription
BEFORE DELETE ON watcher.subscription
FOR EACH ROW
EXECUTE FUNCTION watcher.teardown_subscription();

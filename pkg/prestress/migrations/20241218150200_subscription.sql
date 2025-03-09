CREATE TYPE prestress.operation
AS ENUM ('INSERT', 'UPDATE', 'DELETE');

CREATE TYPE prestress.change AS (
  subscription_id BIGINT,
  row_key JSONB,
  row_data JSONB,
  row_operation prestress.operation
);

CREATE SEQUENCE prestress.subscription_id;

CREATE FUNCTION prestress.get_primary_key(
  table_schema NAME,
  table_name NAME,
  row_data JSONB
)
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
    EXECUTE format(
      'SELECT jsonb_build_array(%s)
      FROM (SELECT $1 AS row_data)',
      (SELECT string_agg('row_data->' || quote_literal(column_name), ',')
        FROM unnest(key_columns) AS column_name)
    )
    INTO primary_key
    USING row_data;
    RETURN primary_key;
  END;
$$;

CREATE FUNCTION prestress.get_related_tables(
  source_schema NAME,
  source_table NAME
)
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

CREATE FUNCTION prestress.extract_change(
  subscription_id BIGINT,
  table_schema NAME,
  table_name NAME,
  operation prestress.operation
)
RETURNS SETOF prestress.change
LANGUAGE plpgsql
AS $$
  DECLARE
    changed_rows JSONB[];
    diff_query TEXT := 'SELECT array_agg(to_jsonb(a))
      FROM %s AS a
      WHERE a NOT IN (SELECT b FROM %s AS b)';
    state_table_name NAME := 'prestress_state_' || subscription_id;
  BEGIN
    IF operation = 'INSERT' THEN
      EXECUTE format(
        diff_query,
        quote_ident(table_schema) || '.' ||
        quote_ident(table_name),
        quote_ident(state_table_name)
      )
      INTO changed_rows;
    ELSIF operation = 'UPDATE' THEN
      RAISE EXCEPTION 'Operation UPDATE for prestress.extract_change without
        key_columns parameter is not implemented';
    ELSIF operation = 'DELETE' THEN
      EXECUTE format(
        diff_query,
        quote_ident(state_table_name),
        quote_ident(table_schema) || '.' ||
        quote_ident(table_name)
      )
      INTO changed_rows;
    END IF;
    RETURN QUERY SELECT
      subscription_id,
      prestress.get_primary_key(
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

CREATE FUNCTION prestress.extract_change(
  subscription_id BIGINT,
  table_schema NAME,
  table_name NAME,
  operation prestress.operation,
  key_columns NAME[]
)
RETURNS SETOF prestress.change
LANGUAGE plpgsql
AS $$
  DECLARE
    changed_rows JSONB[];
    diff_query TEXT := 'SELECT array_agg(to_jsonb(a))
      FROM %s AS a
      WHERE jsonb_build_array(%s) %s (
        SELECT jsonb_build_array(%s) FROM %s AS b)';
    state_table_name NAME := 'prestress_state_' || subscription_id;
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
        quote_ident(state_table_name)
      )
      INTO changed_rows;
    ELSIF operation = 'UPDATE' THEN
      EXECUTE format(
        diff_query || ' AND a NOT IN (SELECT b FROM %s AS b)',
        quote_ident(table_schema) || '.' ||
        quote_ident(table_name),
        (SELECT string_agg('a.'|| quote_ident(column_name), ',')
          FROM unnest(key_columns) AS column_name),
        'IN',
        (SELECT string_agg('b.'|| quote_ident(column_name), ',')
          FROM unnest(key_columns) AS column_name),
        quote_ident(state_table_name),
        quote_ident(state_table_name)
      )
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
        quote_ident(table_name)
      )
      INTO changed_rows;
    END IF;
    RETURN QUERY SELECT
      subscription_id,
      prestress.get_primary_key(
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

CREATE FUNCTION prestress.record_state(
  subscription_id BIGINT,
  table_schema NAME,
  table_name NAME
)
RETURNS VOID
LANGUAGE plpgsql
AS $$
  BEGIN
    EXECUTE format(
      'CREATE TEMPORARY TABLE pg_temp.%I
      ON COMMIT DROP
      AS SELECT * FROM %I.%I',
      'prestress_state_' || subscription_id,
      table_schema,
      table_name
    );
  END;
$$;

CREATE FUNCTION prestress.drop_state(subscription_id BIGINT)
RETURNS VOID
LANGUAGE plpgsql
AS $$
  BEGIN
    EXECUTE format(
      'DROP TABLE pg_temp.%I',
      'prestress_state_' || subscription_id
    );
  END;
$$;

CREATE FUNCTION prestress.record_change(
  subscription_id BIGINT,
  table_schema NAME,
  table_name NAME
)
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
      INSERT INTO pg_temp.prestress_change
      SELECT *
      FROM prestress.extract_change(
        subscription_id,
        table_schema,
        table_name,
        'INSERT',
        key_columns)
      UNION
      SELECT *
      FROM prestress.extract_change(
        subscription_id,
        table_schema,
        table_name,
        'UPDATE',
        key_columns)
      UNION
      SELECT *
      FROM prestress.extract_change(
        subscription_id,
        table_schema,
        table_name,
        'DELETE',
        key_columns);
    ELSE
      INSERT INTO pg_temp.prestress_change
      SELECT *
      FROM prestress.extract_change(
        subscription_id,
        table_schema,
        table_name,
        'INSERT')
      UNION
      SELECT *
      FROM prestress.extract_change(
        subscription_id,
        table_schema,
        table_name,
        'DELETE');
    END IF;
  END;
$$;

CREATE FUNCTION prestress.setup_subscription(
  role_name NAME,
  table_schema NAME,
  table_name NAME,
  authorization_variables jsonb)
RETURNS BIGINT
LANGUAGE plpgsql
AS $$
  DECLARE
    subscription_id BIGINT := nextval('prestress.subscription_id');
    original_role NAME := CURRENT_USER;
  BEGIN
    EXECUTE format('SET LOCAL ROLE TO %I', role_name);

    EXECUTE format(
      'CREATE FUNCTION pg_temp.%I()
      RETURNS TRIGGER
      LANGUAGE plpgsql
      SECURITY DEFINER
      AS $s$
        BEGIN
          PERFORM prestress.begin_authorized(%L);
          PERFORM prestress.record_state(%L, %L, %L);
          PERFORM prestress.end_authorized();
          RETURN NULL;
        END;
      $s$;',
      'prestress_before_' || subscription_id,
      authorization_variables,
      subscription_id,
      table_schema,
      table_name
    );

    EXECUTE format(
      'CREATE FUNCTION pg_temp.%I()
      RETURNS TRIGGER
      LANGUAGE plpgsql
      SECURITY DEFINER
      AS $s$
        BEGIN
          PERFORM prestress.begin_authorized(%L);
          PERFORM prestress.record_change(%L, %L, %L);
          PERFORM prestress.drop_state(%L);
          PERFORM prestress.end_authorized();
          RETURN NULL;
        END;
      $s$;',
      'prestress_after_' || subscription_id,
      authorization_variables,
      subscription_id,
      table_schema,
      table_name,
      subscription_id
    );

    EXECUTE format('SET LOCAL ROLE TO %I', original_role);

    EXECUTE format(
      'CREATE TRIGGER %I
      BEFORE INSERT OR UPDATE OR DELETE ON %I.%I
      FOR EACH STATEMENT
      EXECUTE FUNCTION pg_temp.%I();',
      'prestress_before_' || subscription_id,
      table_schema,
      table_name,
      'prestress_before_' || subscription_id
    );

    EXECUTE format(
      'CREATE TRIGGER %I
      AFTER INSERT OR UPDATE OR DELETE ON %I.%I
      FOR EACH STATEMENT
      EXECUTE FUNCTION pg_temp.%I();',
      'prestress_after_' || subscription_id,
      table_schema,
      table_name,
      'prestress_after_' || subscription_id
    );

    RETURN subscription_id;
  END;
$$;

CREATE FUNCTION prestress.teardown_subscription(id BIGINT)
RETURNS VOID
LANGUAGE plpgsql
AS $$
  BEGIN
    EXECUTE format(
      'DROP FUNCTION pg_temp.%I() CASCADE;',
      'prestress_before_' || id
    );
    EXECUTE format(
      'DROP FUNCTION pg_temp.%I() CASCADE;',
      'prestress_after_' || id
    );
  END;
$$;

GRANT EXECUTE ON FUNCTION prestress.get_primary_key TO public;
GRANT EXECUTE ON FUNCTION prestress.get_related_tables TO public;
GRANT EXECUTE ON FUNCTION prestress.extract_change(
  BIGINT, NAME, NAME, prestress.operation
)
TO public;
GRANT EXECUTE ON FUNCTION prestress.extract_change(
  BIGINT, NAME, NAME, prestress.operation, NAME[]
)
TO public;
GRANT EXECUTE ON FUNCTION prestress.record_state TO public;
GRANT EXECUTE ON FUNCTION prestress.drop_state TO public;
GRANT EXECUTE ON FUNCTION prestress.record_change TO public;
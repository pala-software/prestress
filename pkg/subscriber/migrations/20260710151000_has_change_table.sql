CREATE FUNCTION prestress.has_change_table()
RETURNS BOOLEAN
LANGUAGE sql
RETURN EXISTS (
  SELECT
  FROM information_schema.tables
  WHERE table_type = 'LOCAL TEMPORARY'
  AND table_name = 'prestress_change'
);

CREATE OR REPLACE FUNCTION prestress.setup_subscription(
  role_name NAME,
  table_schema NAME,
  table_name NAME,
  authorization_variables jsonb)
RETURNS BIGINT
LANGUAGE plpgsql
AS $$
  DECLARE
    original_role NAME := CURRENT_USER;
    subscription_id BIGINT := nextval('prestress.subscription_id');
    target_schema NAME;
    target_table NAME;
    trigger_id BIGINT := 1;
  BEGIN
    EXECUTE format('SET LOCAL ROLE TO %I', role_name);

    EXECUTE format(
      'CREATE FUNCTION pg_temp.%I()
      RETURNS TRIGGER
      LANGUAGE plpgsql
      SECURITY DEFINER
      AS $s$
        BEGIN
          IF prestress.has_state(%L, %L, %L) THEN
            RETURN NULL;
          END IF;
          PERFORM prestress.begin_authorized(%L);
          PERFORM prestress.record_state(%L, %L, %L);
          PERFORM prestress.end_authorized();
          RETURN NULL;
        EXCEPTION
          WHEN insufficient_privilege THEN
            RAISE WARNING ''caught insufficient_privilege in prestress_before_x trigger function'';
            RETURN NULL;
        END;
      $s$;',
      'prestress_before_' || subscription_id,
      subscription_id,
      table_schema,
      table_name,
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
          IF NOT prestress.has_state(%L, %L, %L) THEN
            RETURN NULL;
          END IF;
          PERFORM prestress.begin_authorized(%L);
          PERFORM prestress.record_change(%L, %L, %L);
          PERFORM prestress.drop_state(%L, %L, %L);
          PERFORM prestress.end_authorized();
          RETURN NULL;
        EXCEPTION
          WHEN insufficient_privilege THEN
            RAISE WARNING ''caught insufficient_privilege in prestress_after_x trigger function'';
            RETURN NULL;
        END;
      $s$;',
      'prestress_after_' || subscription_id,
      subscription_id,
      table_schema,
      table_name,
      authorization_variables,
      subscription_id,
      table_schema,
      table_name,
      subscription_id,
      table_schema,
      table_name
    );

    EXECUTE format('SET LOCAL ROLE TO %I', original_role);

    EXECUTE format(
      'CREATE FUNCTION pg_temp.%I()
      RETURNS TRIGGER
      LANGUAGE plpgsql
      SECURITY INVOKER
      AS $s$
        DECLARE
          original_authorization jsonb;
        BEGIN
          SELECT prestress.dump_authorization() INTO original_authorization;
          CREATE TEMPORARY TABLE pg_temp.%I ON COMMIT DROP AS SELECT original_authorization AS value;
          IF original_authorization IS NOT NULL THEN
            PERFORM prestress.end_authorized();
          END IF;
          RETURN NULL;
        END;
      $s$;',
      'prestress_begin_' || subscription_id,
      'prestress_auth_' || subscription_id
    );

    EXECUTE format(
      'CREATE FUNCTION pg_temp.%I()
      RETURNS TRIGGER
      LANGUAGE plpgsql
      SECURITY INVOKER
      AS $s$
        DECLARE
          original_authorization jsonb;
        BEGIN
          SELECT value INTO original_authorization FROM pg_temp.%I;
          DROP TABLE pg_temp.%I;
          IF original_authorization IS NOT NULL THEN
            PERFORM prestress.begin_authorized(original_authorization);
          END IF;
          RETURN NULL;
        END;
      $s$;',
      'prestress_end_' || subscription_id,
      'prestress_auth_' || subscription_id,
      'prestress_auth_' || subscription_id
    );

    FOR target_schema, target_table IN
      SELECT related_table.table_schema, related_table.table_name
      FROM prestress.get_related_tables(table_schema, table_name)
      AS related_table
    LOOP
      EXECUTE format(
        'CREATE TRIGGER %I
        BEFORE INSERT OR UPDATE OR DELETE ON %I.%I
        FOR EACH STATEMENT
        WHEN (prestress.has_change_table())
        EXECUTE FUNCTION pg_temp.%I();',
        'prestress_before_1_' || subscription_id || '_' || trigger_id,
        target_schema,
        target_table,
        'prestress_begin_' || subscription_id
      );

      EXECUTE format(
        'CREATE TRIGGER %I
        BEFORE INSERT OR UPDATE OR DELETE ON %I.%I
        FOR EACH STATEMENT
        WHEN (prestress.has_change_table())
        EXECUTE FUNCTION pg_temp.%I();',
        'prestress_before_2_' || subscription_id || '_' || trigger_id,
        target_schema,
        target_table,
        'prestress_before_' || subscription_id
      );

      EXECUTE format(
        'CREATE TRIGGER %I
        BEFORE INSERT OR UPDATE OR DELETE ON %I.%I
        FOR EACH STATEMENT
        WHEN (prestress.has_change_table())
        EXECUTE FUNCTION pg_temp.%I();',
        'prestress_before_3_' || subscription_id || '_' || trigger_id,
        target_schema,
        target_table,
        'prestress_end_' || subscription_id
      );

      EXECUTE format(
        'CREATE TRIGGER %I
        AFTER INSERT OR UPDATE OR DELETE ON %I.%I
        FOR EACH STATEMENT
        WHEN (prestress.has_change_table())
        EXECUTE FUNCTION pg_temp.%I();',
        'prestress_after_1_' || subscription_id || '_' || trigger_id,
        target_schema,
        target_table,
        'prestress_begin_' || subscription_id
      );

      EXECUTE format(
        'CREATE TRIGGER %I
        AFTER INSERT OR UPDATE OR DELETE ON %I.%I
        FOR EACH STATEMENT
        WHEN (prestress.has_change_table())
        EXECUTE FUNCTION pg_temp.%I();',
        'prestress_after_2_' || subscription_id || '_' || trigger_id,
        target_schema,
        target_table,
        'prestress_after_' || subscription_id
      );

      EXECUTE format(
        'CREATE TRIGGER %I
        AFTER INSERT OR UPDATE OR DELETE ON %I.%I
        FOR EACH STATEMENT
        WHEN (prestress.has_change_table())
        EXECUTE FUNCTION pg_temp.%I();',
        'prestress_after_3_' || subscription_id || '_' || trigger_id,
        target_schema,
        target_table,
        'prestress_end_' || subscription_id
      );

      SELECT trigger_id + 1 INTO trigger_id;
    END LOOP;

    RETURN subscription_id;
  END;
$$;

GRANT EXECUTE ON FUNCTION prestress.has_change_table TO public;

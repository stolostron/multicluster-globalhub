CREATE OR REPLACE FUNCTION public.move_applications_to_history() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO history.applications SELECT * FROM spec.applications
  WHERE payload -> 'metadata' ->> 'name' = NEW.payload -> 'metadata' ->> 'name' AND
  (
    (
      (payload -> 'metadata' ->> 'namespace' IS NOT NULL AND NEW.payload -> 'metadata' ->> 'namespace' IS NOT NULL)
    AND payload -> 'metadata' ->> 'namespace' = NEW.payload -> 'metadata' ->> 'namespace'
    ) OR (
      payload -> 'metadata' -> 'namespace' IS NULL AND NEW.payload -> 'metadata' -> 'namespace' IS NULL
    )
  );
  DELETE FROM spec.applications
  WHERE payload -> 'metadata' ->> 'name' = NEW.payload -> 'metadata' ->> 'name' AND
  (
    (
      (payload -> 'metadata' ->> 'namespace' IS NOT NULL AND NEW.payload -> 'metadata' ->> 'namespace' IS NOT NULL)
    AND payload -> 'metadata' ->> 'namespace' = NEW.payload -> 'metadata' ->> 'namespace'
    ) OR (
      payload -> 'metadata' -> 'namespace' IS NULL AND NEW.payload -> 'metadata' -> 'namespace' IS NULL
    )
  );
  RETURN NEW;
END;
$$;


CREATE OR REPLACE FUNCTION public.move_channels_to_history() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO history.channels SELECT * FROM spec.channels
  WHERE payload -> 'metadata' ->> 'name' = NEW.payload -> 'metadata' ->> 'name' AND
  (
    (
      (payload -> 'metadata' ->> 'namespace' IS NOT NULL AND NEW.payload -> 'metadata' ->> 'namespace' IS NOT NULL)
    AND payload -> 'metadata' ->> 'namespace' = NEW.payload -> 'metadata' ->> 'namespace'
    ) OR (
      payload -> 'metadata' -> 'namespace' IS NULL AND NEW.payload -> 'metadata' -> 'namespace' IS NULL
    )
  );
  DELETE FROM spec.channels
  WHERE payload -> 'metadata' ->> 'name' = NEW.payload -> 'metadata' ->> 'name' AND
  (
    (
      (payload -> 'metadata' ->> 'namespace' IS NOT NULL AND NEW.payload -> 'metadata' ->> 'namespace' IS NOT NULL)
    AND payload -> 'metadata' ->> 'namespace' = NEW.payload -> 'metadata' ->> 'namespace'
    ) OR (
      payload -> 'metadata' -> 'namespace' IS NULL AND NEW.payload -> 'metadata' -> 'namespace' IS NULL
    )
  );
  RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION public.move_configs_to_history() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO history.configs SELECT * FROM spec.configs
  WHERE payload -> 'metadata' ->> 'name' = NEW.payload -> 'metadata' ->> 'name' AND
  (
    (
      (payload -> 'metadata' ->> 'namespace' IS NOT NULL AND NEW.payload -> 'metadata' ->> 'namespace' IS NOT NULL)
    AND payload -> 'metadata' ->> 'namespace' = NEW.payload -> 'metadata' ->> 'namespace'
    ) OR (
      payload -> 'metadata' -> 'namespace' IS NULL AND NEW.payload -> 'metadata' -> 'namespace' IS NULL
    )
  );
  DELETE FROM spec.configs
  WHERE payload -> 'metadata' ->> 'name' = NEW.payload -> 'metadata' ->> 'name' AND
  (
    (
      (payload -> 'metadata' ->> 'namespace' IS NOT NULL AND NEW.payload -> 'metadata' ->> 'namespace' IS NOT NULL)
    AND payload -> 'metadata' ->> 'namespace' = NEW.payload -> 'metadata' ->> 'namespace'
    ) OR (
      payload -> 'metadata' -> 'namespace' IS NULL AND NEW.payload -> 'metadata' -> 'namespace' IS NULL
    )
  );
  RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION public.move_managedclustersetbindings_to_history() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO history.managedclustersetbindings SELECT * FROM spec.managedclustersetbindings
  WHERE payload -> 'metadata' ->> 'name' = NEW.payload -> 'metadata' ->> 'name' AND
  (
    (
      (payload -> 'metadata' ->> 'namespace' IS NOT NULL AND NEW.payload -> 'metadata' ->> 'namespace' IS NOT NULL)
    AND payload -> 'metadata' ->> 'namespace' = NEW.payload -> 'metadata' ->> 'namespace'
    ) OR (
      payload -> 'metadata' -> 'namespace' IS NULL AND NEW.payload -> 'metadata' -> 'namespace' IS NULL
    )
  );
  DELETE FROM spec.managedclustersetbindings
  WHERE payload -> 'metadata' ->> 'name' = NEW.payload -> 'metadata' ->> 'name' AND
  (
    (
      (payload -> 'metadata' ->> 'namespace' IS NOT NULL AND NEW.payload -> 'metadata' ->> 'namespace' IS NOT NULL)
    AND payload -> 'metadata' ->> 'namespace' = NEW.payload -> 'metadata' ->> 'namespace'
    ) OR (
      payload -> 'metadata' -> 'namespace' IS NULL AND NEW.payload -> 'metadata' -> 'namespace' IS NULL
    )
  );
  RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION public.move_managedclustersets_to_history() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO history.managedclustersets SELECT * FROM spec.managedclustersets
  WHERE payload -> 'metadata' ->> 'name' = NEW.payload -> 'metadata' ->> 'name' AND
  (
    (
      (payload -> 'metadata' ->> 'namespace' IS NOT NULL AND NEW.payload -> 'metadata' ->> 'namespace' IS NOT NULL)
    AND payload -> 'metadata' ->> 'namespace' = NEW.payload -> 'metadata' ->> 'namespace'
    ) OR (
      payload -> 'metadata' -> 'namespace' IS NULL AND NEW.payload -> 'metadata' -> 'namespace' IS NULL
    )
  );
  DELETE FROM spec.managedclustersets
  WHERE payload -> 'metadata' ->> 'name' = NEW.payload -> 'metadata' ->> 'name' AND
  (
    (
      (payload -> 'metadata' ->> 'namespace' IS NOT NULL AND NEW.payload -> 'metadata' ->> 'namespace' IS NOT NULL)
    AND payload -> 'metadata' ->> 'namespace' = NEW.payload -> 'metadata' ->> 'namespace'
    ) OR (
      payload -> 'metadata' -> 'namespace' IS NULL AND NEW.payload -> 'metadata' -> 'namespace' IS NULL
    )
  );
  RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION public.move_placementbindings_to_history() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO history.placementbindings SELECT * FROM spec.placementbindings
  WHERE payload -> 'metadata' ->> 'name' = NEW.payload -> 'metadata' ->> 'name' AND
  (
    (
      (payload -> 'metadata' ->> 'namespace' IS NOT NULL AND NEW.payload -> 'metadata' ->> 'namespace' IS NOT NULL)
    AND payload -> 'metadata' ->> 'namespace' = NEW.payload -> 'metadata' ->> 'namespace'
    ) OR (
      payload -> 'metadata' -> 'namespace' IS NULL AND NEW.payload -> 'metadata' -> 'namespace' IS NULL
    )
  );
  DELETE FROM spec.placementbindings
  WHERE payload -> 'metadata' ->> 'name' = NEW.payload -> 'metadata' ->> 'name' AND
  (
    (
      (payload -> 'metadata' ->> 'namespace' IS NOT NULL AND NEW.payload -> 'metadata' ->> 'namespace' IS NOT NULL)
    AND payload -> 'metadata' ->> 'namespace' = NEW.payload -> 'metadata' ->> 'namespace'
    ) OR (
      payload -> 'metadata' -> 'namespace' IS NULL AND NEW.payload -> 'metadata' -> 'namespace' IS NULL
    )
  );
  RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION public.move_placementrules_to_history() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO history.placementrules SELECT * FROM spec.placementrules
  WHERE payload -> 'metadata' ->> 'name' = NEW.payload -> 'metadata' ->> 'name' AND
  (
    (
      (payload -> 'metadata' ->> 'namespace' IS NOT NULL AND NEW.payload -> 'metadata' ->> 'namespace' IS NOT NULL)
    AND payload -> 'metadata' ->> 'namespace' = NEW.payload -> 'metadata' ->> 'namespace'
    ) OR (
      payload -> 'metadata' -> 'namespace' IS NULL AND NEW.payload -> 'metadata' -> 'namespace' IS NULL
    )
  );
  DELETE FROM spec.placementrules
  WHERE payload -> 'metadata' ->> 'name' = NEW.payload -> 'metadata' ->> 'name' AND
  (
    (
      (payload -> 'metadata' ->> 'namespace' IS NOT NULL AND NEW.payload -> 'metadata' ->> 'namespace' IS NOT NULL)
    AND payload -> 'metadata' ->> 'namespace' = NEW.payload -> 'metadata' ->> 'namespace'
    ) OR (
      payload -> 'metadata' -> 'namespace' IS NULL AND NEW.payload -> 'metadata' -> 'namespace' IS NULL
    )
  );
  RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION public.move_placements_to_history() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO history.placements SELECT * FROM spec.placements
  WHERE payload -> 'metadata' ->> 'name' = NEW.payload -> 'metadata' ->> 'name' AND
  (
    (
      (payload -> 'metadata' ->> 'namespace' IS NOT NULL AND NEW.payload -> 'metadata' ->> 'namespace' IS NOT NULL)
    AND payload -> 'metadata' ->> 'namespace' = NEW.payload -> 'metadata' ->> 'namespace'
    ) OR (
      payload -> 'metadata' -> 'namespace' IS NULL AND NEW.payload -> 'metadata' -> 'namespace' IS NULL
    )
  );
  DELETE FROM spec.placements
  WHERE payload -> 'metadata' ->> 'name' = NEW.payload -> 'metadata' ->> 'name' AND
  (
    (
      (payload -> 'metadata' ->> 'namespace' IS NOT NULL AND NEW.payload -> 'metadata' ->> 'namespace' IS NOT NULL)
    AND payload -> 'metadata' ->> 'namespace' = NEW.payload -> 'metadata' ->> 'namespace'
    ) OR (
      payload -> 'metadata' -> 'namespace' IS NULL AND NEW.payload -> 'metadata' -> 'namespace' IS NULL
    )
  );
  RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION public.move_policies_to_history() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO history.policies SELECT * FROM spec.policies
  WHERE payload -> 'metadata' ->> 'name' = NEW.payload -> 'metadata' ->> 'name' AND
  (
    (
      (payload -> 'metadata' ->> 'namespace' IS NOT NULL AND NEW.payload -> 'metadata' ->> 'namespace' IS NOT NULL)
    AND payload -> 'metadata' ->> 'namespace' = NEW.payload -> 'metadata' ->> 'namespace'
    ) OR (
      payload -> 'metadata' -> 'namespace' IS NULL AND NEW.payload -> 'metadata' -> 'namespace' IS NULL
    )
  );
  DELETE FROM spec.policies
  WHERE payload -> 'metadata' ->> 'name' = NEW.payload -> 'metadata' ->> 'name' AND
  (
    (
      (payload -> 'metadata' ->> 'namespace' IS NOT NULL AND NEW.payload -> 'metadata' ->> 'namespace' IS NOT NULL)
    AND payload -> 'metadata' ->> 'namespace' = NEW.payload -> 'metadata' ->> 'namespace'
    ) OR (
      payload -> 'metadata' -> 'namespace' IS NULL AND NEW.payload -> 'metadata' -> 'namespace' IS NULL
    )
  );
  RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION public.move_subscriptions_to_history() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO history.subscriptions SELECT * FROM spec.subscriptions
  WHERE payload -> 'metadata' ->> 'name' = NEW.payload -> 'metadata' ->> 'name' AND
  (
    (
      (payload -> 'metadata' ->> 'namespace' IS NOT NULL AND NEW.payload -> 'metadata' ->> 'namespace' IS NOT NULL)
    AND payload -> 'metadata' ->> 'namespace' = NEW.payload -> 'metadata' ->> 'namespace'
    ) OR (
      payload -> 'metadata' -> 'namespace' IS NULL AND NEW.payload -> 'metadata' -> 'namespace' IS NULL
    )
  );
  DELETE FROM spec.subscriptions
  WHERE payload -> 'metadata' ->> 'name' = NEW.payload -> 'metadata' ->> 'name' AND
  (
    (
      (payload -> 'metadata' ->> 'namespace' IS NOT NULL AND NEW.payload -> 'metadata' ->> 'namespace' IS NOT NULL)
    AND payload -> 'metadata' ->> 'namespace' = NEW.payload -> 'metadata' ->> 'namespace'
    ) OR (
      payload -> 'metadata' -> 'namespace' IS NULL AND NEW.payload -> 'metadata' -> 'namespace' IS NULL
    )
  );
  RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION public.trigger_set_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION public.set_cluster_id_to_local_compliance() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  UPDATE local_status.compliance set cluster_id=(SELECT cluster_id FROM status.managed_clusters
  WHERE payload -> 'metadata' ->> 'name' = NEW.cluster_name AND leaf_hub_name = NEW.leaf_hub_name)
  WHERE cluster_name = NEW.cluster_name AND leaf_hub_name = NEW.leaf_hub_name AND cluster_id IS NULL;
  RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION public.set_cluster_id_to_compliance() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  UPDATE status.compliance set cluster_id=(SELECT cluster_id FROM status.managed_clusters
  WHERE payload -> 'metadata' ->> 'name' = NEW.cluster_name AND leaf_hub_name = NEW.leaf_hub_name)
  WHERE cluster_name = NEW.cluster_name AND leaf_hub_name = NEW.leaf_hub_name AND cluster_id IS NULL;
  RETURN NEW;
END;
$$;

-- -- Create history trigger function
-- CREATE OR REPLACE FUNCTION history.move_managed_cluster_to_history()
-- RETURNS TRIGGER AS $$
-- BEGIN
--     INSERT INTO history.managed_clusters (leaf_hub_name, cluster_id, payload, error)
--     VALUES (OLD.leaf_hub_name, OLD.cluster_id, OLD.payload, OLD.error);
--     RETURN OLD;
-- END;
-- $$ LANGUAGE plpgsql;

-- CREATE OR REPLACE FUNCTION history.move_local_policy_to_history()
-- RETURNS TRIGGER AS $$
-- BEGIN
--     INSERT INTO history.local_policies (leaf_hub_name, payload, created_at, updated_at)
--     VALUES (OLD.leaf_hub_name, OLD.payload, OLD.created_at, OLD.updated_at);
--     RETURN OLD;
-- END;
-- $$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION set_policies_timestamps()
RETURNS TRIGGER AS $$
DECLARE
    table_name text;
BEGIN
    table_name := TG_TABLE_NAME;
    IF TG_OP = 'INSERT' THEN
        NEW.created_at := NOW();
        NEW.updated_at := NOW();
    ELSIF TG_OP = 'UPDATE' THEN
        NEW.updated_at := NOW();
    ELSIF TG_OP = 'DELETE' THEN
        -- Soft delete: update the deleted_at timestamp
        NEW.updated_at := NOW();
        NEW.deleted_at := NOW();

        -- Update the row instead of deleting
        EXECUTE format('UPDATE %I.%I SET updated_at = $1, deleted_at = $2 WHERE policy_id = $3', TG_TABLE_SCHEMA, TG_TABLE_NAME)
        USING NEW.updated_at, NEW.deleted_at, OLD.policy_id;

        -- Prevent actual deletion by returning NULL
        RETURN NULL;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION set_managed_cluster_timestamps()
RETURNS TRIGGER AS $$
DECLARE
    table_name text;
BEGIN
    table_name := TG_TABLE_NAME;
    IF TG_OP = 'INSERT' THEN
        NEW.created_at := NOW();
        NEW.updated_at := NOW();
    ELSIF TG_OP = 'UPDATE' THEN
        NEW.updated_at := NOW();
    ELSIF TG_OP = 'DELETE' THEN
        -- Soft delete: update the deleted_at timestamp
        NEW.updated_at := NOW();
        NEW.deleted_at := NOW();

        -- Update the row instead of deleting
        EXECUTE format('UPDATE %I.%I SET updated_at = $1, deleted_at = $2 WHERE cluster_name = $3', TG_TABLE_SCHEMA, TG_TABLE_NAME)
        USING NEW.updated_at, NEW.deleted_at, OLD.cluster_name;

        -- Prevent actual deletion by returning NULL
        RETURN NULL;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
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


CREATE OR REPLACE FUNCTION local_status.set_cluster_id_to_compliance()
  RETURNS TRIGGER AS $$
DECLARE
  retry_count INT := 0;
BEGIN
  -- Retry loop with a maximum of 12 attempts = 2 min
  WHILE retry_count < 12 LOOP
    BEGIN
      NEW.cluster_id := (
        SELECT cluster_id
        FROM status.managed_clusters
        WHERE payload -> 'metadata' ->> 'name' = NEW.cluster_name
          AND leaf_hub_name = NEW.leaf_hub_name
        LIMIT 1
      );
      
      IF NEW.cluster_id IS NOT NULL THEN
        RETURN NEW;
      END IF;
      
      retry_count := retry_count + 1;
      PERFORM pg_sleep(10); -- Wait for 10 seconds before retrying
    EXCEPTION
      WHEN OTHERS THEN
        retry_count := retry_count + 1;
        PERFORM pg_sleep(10); -- Wait for 10 seconds before retrying
    END;
  END LOOP;

  RETURN NEW;
END;
$$;

-- Create the trigger function
CREATE OR REPLACE FUNCTION history.move_managed_cluster_to_history()
RETURNS TRIGGER AS $$
BEGIN
    -- Insert the deleted row into history.managed_clusters
    INSERT INTO history.managed_clusters (leaf_hub_name, cluster_id, payload, error)
    VALUES (OLD.leaf_hub_name, OLD.cluster_id, OLD.payload, OLD.error);
    RETURN OLD;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION status.set_cluster_id_to_compliance()
  RETURNS TRIGGER AS $$
DECLARE
  retry_count INT := 0;
BEGIN
  -- Retry loop with a maximum of 12 attempts = 2 min
  WHILE retry_count < 12 LOOP
    BEGIN
      NEW.cluster_id := (
        SELECT cluster_id
        FROM status.managed_clusters
        WHERE payload -> 'metadata' ->> 'name' = NEW.cluster_name
          AND leaf_hub_name = NEW.leaf_hub_name
        LIMIT 1
      );
      
      IF NEW.cluster_id IS NOT NULL THEN
        RETURN NEW;
      END IF;
      
      retry_count := retry_count + 1;
      PERFORM pg_sleep(10); 
    EXCEPTION
      WHEN OTHERS THEN
        retry_count := retry_count + 1;
        PERFORM pg_sleep(10);
    END;
  END LOOP;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
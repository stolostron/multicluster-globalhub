package hubofhubs

import (
	"context"
	"embed"
	"fmt"
	iofs "io/fs"
	"net/url"
	"strings"

	globalhubv1alpha4 "github.com/stolostron/multicluster-global-hub/operator/apis/v1alpha4"
	"github.com/stolostron/multicluster-global-hub/operator/pkg/condition"
	"github.com/stolostron/multicluster-global-hub/operator/pkg/config"
	"github.com/stolostron/multicluster-global-hub/pkg/database"
)

var DatabaseReconcileCounter = 0

//go:embed database
var databaseFS embed.FS

func (r *MulticlusterGlobalHubReconciler) reconcileDatabase(ctx context.Context,
	mgh *globalhubv1alpha4.MulticlusterGlobalHub,
) error {
	log := r.Log.WithName("database")

	if config.SkipDBInit(mgh) {
		log.Info("database initialization is skipped")
		return nil
	}

	if condition.ContainConditionStatus(mgh, condition.CONDITION_TYPE_DATABASE_INIT, condition.CONDITION_STATUS_TRUE) {
		log.Info("database has been initialized, checking the reconcile counter")
		// if the operator is restarted, reconcile the database again
		if DatabaseReconcileCounter > 0 {
			return nil
		}
	}

	conn, err := database.PostgresConnection(ctx, r.MiddlewareConfig.PgConnection.SuperuserDatabaseURI,
		r.MiddlewareConfig.PgConnection.CACert)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() {
		if err := conn.Close(ctx); err != nil {
			log.Error(err, "failed to close connection to database")
		}
	}()

	username := ""
	objURI, err := url.Parse(r.MiddlewareConfig.PgConnection.ReadonlyUserDatabaseURI)
	if err != nil {
		log.Error(err, "failed to parse database_uri_with_readonlyuser")
	} else {
		username = objURI.User.Username()
	}

	err = iofs.WalkDir(databaseFS, "database", func(file string, d iofs.DirEntry, beforeError error) error {
		if beforeError != nil {
			return beforeError
		}
		if d.IsDir() {
			return nil
		}
		sqlBytes, err := databaseFS.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}
		if file == "database/5.privileges.sql" {
			if username != "" {
				_, err = conn.Exec(ctx, strings.ReplaceAll(string(sqlBytes), "$1", username))
			}
		} else {
			_, err = conn.Exec(ctx, string(sqlBytes))
		}
		if err != nil {
			return fmt.Errorf("failed to create %s: %w", file, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to exec database sql: %w", err)
	}

	log.Info("database initialized")
	DatabaseReconcileCounter++
	err = condition.SetConditionDatabaseInit(ctx, r.Client, mgh, condition.CONDITION_STATUS_TRUE)
	if err != nil {
		return condition.FailToSetConditionError(condition.CONDITION_STATUS_TRUE, err)
	}
	return nil
}

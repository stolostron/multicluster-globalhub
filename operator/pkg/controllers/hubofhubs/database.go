package hubofhubs

import (
	"context"
	"embed"
	"fmt"
	iofs "io/fs"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1alpha2 "github.com/stolostron/multicluster-global-hub/operator/apis/v1alpha2"
	"github.com/stolostron/multicluster-global-hub/operator/pkg/condition"
	"github.com/stolostron/multicluster-global-hub/pkg/database"
)

//go:embed database
var databaseFS embed.FS

func (reconciler *MulticlusterGlobalHubReconciler) reconcileDatabase(ctx context.Context, mgh *operatorv1alpha2.MulticlusterGlobalHub,
	namespacedName types.NamespacedName,
) error {
	log := ctrllog.FromContext(ctx)
	if condition.ContainConditionStatus(mgh, condition.CONDITION_TYPE_DATABASE_INIT, condition.CONDITION_STATUS_TRUE) {
		log.Info("Database has initialized")
		return nil
	}

	log.Info("Database initializing")
	postgresSecret, err := reconciler.KubeClient.CoreV1().Secrets(namespacedName.Namespace).Get(
		ctx, namespacedName.Name, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "failed to get storage secret",
			"namespace", namespacedName.Namespace,
			"name", namespacedName.Name)
		return err
	}

	conn, err := database.PostgresConnection(ctx, string(postgresSecret.Data["database_uri"]),
		postgresSecret.Data["ca.crt"])
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() {
		if err := conn.Close(ctx); err != nil {
			log.Error(err, "failed to close connection to database")
		}
	}()

	err = iofs.WalkDir(databaseFS, "database", func(file string, d iofs.DirEntry, beforeError error) error {
		if beforeError != nil {
			return beforeError
		}
		if d.IsDir() {
			return nil
		}
		log.Info("Database executing SQL file: " + file)
		sqlBytes, err := databaseFS.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}
		_, err = conn.Exec(ctx, string(sqlBytes))
		if err != nil {
			return fmt.Errorf("failed to create %s: %w", file, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to exec database sql: %w", err)
	}

	log.Info("Database initialized")
	err = condition.SetConditionDatabaseInit(ctx, reconciler.Client, mgh, condition.CONDITION_STATUS_TRUE)
	if err != nil {
		return condition.FailToSetConditionError(condition.CONDITION_STATUS_TRUE, err)
	}
	return nil
}

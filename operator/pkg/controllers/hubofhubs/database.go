package hubofhubs

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"embed"
	"fmt"
	iofs "io/fs"

	"github.com/jackc/pgx/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1alpha2 "github.com/stolostron/multicluster-global-hub/operator/apis/v1alpha2"
	"github.com/stolostron/multicluster-global-hub/operator/pkg/condition"
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
	postgreSecret, err := reconciler.KubeClient.CoreV1().Secrets(namespacedName.Namespace).Get(
		ctx, namespacedName.Name, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "failed to get storage secret",
			"namespace", namespacedName.Namespace,
			"name", namespacedName.Name)
		return err
	}

	databaseURI := string(postgreSecret.Data["database_uri"])
	caCert := postgreSecret.Data["ca.crt"]
	tlsCert := postgreSecret.Data["tls.crt"]
	tlsKey := postgreSecret.Data["tls.key"]

	connConfig, err := pgx.ParseConfig(databaseURI)
	if err != nil {
		return fmt.Errorf("failed to parse database uri: %w", err)
	}

	if caCert != nil && tlsCert != nil && tlsKey != nil {
		// Parse TLS certs and keys. These should be stored in PEM format.
		clientTlsCert, err := tls.X509KeyPair(tlsCert, tlsKey)
		if err != nil {
			return fmt.Errorf("failed to parse tls cert and key: %w", err)
		}
		// Parse the CA certificate for the server.
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		connConfig.TLSConfig = &tls.Config{
			RootCAs:            caCertPool,
			Certificates:       []tls.Certificate{clientTlsCert},
			InsecureSkipVerify: true,
		}
	}

	conn, err := pgx.ConnectConfig(ctx, connConfig)
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

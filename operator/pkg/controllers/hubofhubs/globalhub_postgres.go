package hubofhubs

import (
	"context"
	"fmt"

	postgresv1beta1 "github.com/crunchydata/postgres-operator/pkg/apis/postgres-operator.crunchydata.com/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	globalhubv1alpha4 "github.com/stolostron/multicluster-global-hub/operator/apis/v1alpha4"
	"github.com/stolostron/multicluster-global-hub/operator/pkg/config"
	"github.com/stolostron/multicluster-global-hub/operator/pkg/constants"
	"github.com/stolostron/multicluster-global-hub/operator/pkg/postgres"
	"github.com/stolostron/multicluster-global-hub/operator/pkg/utils"
)

// ensureCrunchyPostgresSubscription verifies resources needed for Crunchy Postgres are created
func (r *MulticlusterGlobalHubReconciler) ensureCrunchyPostgresSubscription(ctx context.Context,
	mgh *globalhubv1alpha4.MulticlusterGlobalHub) error {
	postgresSub, err := utils.GetSubscriptionByName(ctx, r.Client, postgres.SubscriptionName)
	if err != nil {
		return err
	}

	// Get sub config, catalogsource, and annotation overrides
	subConfig, err := r.GenerateSubConfig(ctx)
	if err != nil {
		return err
	}

	createSub := false
	if postgresSub == nil {
		// Sub is nil so create a new one
		postgresSub = postgres.NewSubscription(mgh, subConfig, utils.IsCommunityMode())
		createSub = true
	}

	// Apply Crunchy Postgres sub
	calcSub := postgres.RenderSubscription(postgresSub, subConfig, utils.IsCommunityMode())
	if createSub {
		err = r.Client.Create(ctx, calcSub)
	} else {
		err = r.Client.Update(ctx, calcSub)
	}
	if err != nil {
		return fmt.Errorf("error updating subscription %s: %w", calcSub.Name, err)
	}

	return nil
}

// ensureCrunchyPostgres verifies PostgresCluster operand is created
func (r *MulticlusterGlobalHubReconciler) ensureCrunchyPostgres(ctx context.Context) error {

	postgresCluster := &postgresv1beta1.PostgresCluster{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      postgres.PostgresName,
		Namespace: config.GetDefaultNamespace(),
	}, postgresCluster)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Client.Create(ctx, postgres.NewPostgres())
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}
	return nil
}

// waitForPostgresReady waits for postgres to be ready and returns a postgres connection
func (r *MulticlusterGlobalHubReconciler) waitForPostgresReady(ctx context.Context) (
	*postgres.PostgresConnection, error) {
	guestPostgresSecret := &corev1.Secret{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      postgres.PostgresGuestUserSecretName,
		Namespace: config.GetDefaultNamespace(),
	}, guestPostgresSecret)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("postgres secret %s is nil", postgres.PostgresGuestUserSecretName)
		}
		return nil, err
	}
	// wait for postgres user secret to be ready
	superuserPostgresSecret := &corev1.Secret{}
	err = r.Client.Get(ctx, types.NamespacedName{
		Name:      postgres.PostgresSuperUserSecretName,
		Namespace: config.GetDefaultNamespace(),
	}, superuserPostgresSecret)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("postgres secret %s is nil", postgres.PostgresGuestUserSecretName)
		}
		return nil, err
	}
	// wait for guest user secret to be ready
	postgresCertName := &corev1.Secret{}
	err = r.Client.Get(ctx, types.NamespacedName{
		Name:      postgres.PostgresCertName,
		Namespace: config.GetDefaultNamespace(),
	}, postgresCertName)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("postgres secret %s is nil", postgres.PostgresGuestUserSecretName)
		}
		return nil, err
	}

	return &postgres.PostgresConnection{
		SuperuserDatabaseURI:    string(superuserPostgresSecret.Data["uri"]) + postgres.PostgresURIWithSslmode,
		ReadonlyUserDatabaseURI: string(guestPostgresSecret.Data["uri"]) + postgres.PostgresURIWithSslmode,
		CACert:                  postgresCertName.Data["ca.crt"],
	}, nil

}

// GeneratePGConnectionFromGHStorageSecret returns a postgres connection from the GH storage secret
func (r *MulticlusterGlobalHubReconciler) GeneratePGConnectionFromGHStorageSecret(ctx context.Context) (
	*postgres.PostgresConnection, error) {
	pgSecret := &corev1.Secret{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      constants.GHStorageSecretName,
		Namespace: config.GetDefaultNamespace(),
	}, pgSecret)
	if err != nil {
		return nil, err
	}
	return &postgres.PostgresConnection{
		SuperuserDatabaseURI:    string(pgSecret.Data["database_uri"]) + postgres.PostgresURIWithSslmode,
		ReadonlyUserDatabaseURI: string(pgSecret.Data["database_uri_with_readonlyuser"]) + postgres.PostgresURIWithSslmode,
		CACert:                  pgSecret.Data["ca.crt"],
	}, nil
}

package hubofhubs

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1alpha2 "github.com/stolostron/multicluster-global-hub/operator/apis/v1alpha2"
	"github.com/stolostron/multicluster-global-hub/operator/pkg/condition"
	"github.com/stolostron/multicluster-global-hub/operator/pkg/config"
	"github.com/stolostron/multicluster-global-hub/operator/pkg/deployer"
	"github.com/stolostron/multicluster-global-hub/operator/pkg/renderer"
	"github.com/stolostron/multicluster-global-hub/operator/pkg/utils"
	"github.com/stolostron/multicluster-global-hub/pkg/constants"
)

const (
	datasourceKey = "datasources.yaml"
)

func (r *MulticlusterGlobalHubReconciler) reconcileGrafana(ctx context.Context,
	mgh *operatorv1alpha2.MulticlusterGlobalHub,
) error {
	log := ctrllog.FromContext(ctx)
	if condition.ContainConditionStatus(mgh, condition.CONDITION_TYPE_GRAFANA_INIT, condition.CONDITION_STATUS_TRUE) {
		log.Info("Grafana has initialized")
		return nil
	}

	log.Info("Grafana initializing")
	// generate random session secret for oauth-proxy
	proxySessionSecret, err := utils.GeneratePassword(16)
	if err != nil {
		return fmt.Errorf("failed to generate random session secret for grafana oauth-proxy: %v", err)
	}

	// generate datasource secret: must before the grafana objects
	datasourceSecretName, err := r.GenerateGrafanaDataSourceSecret(ctx, r.Client, mgh, r.Scheme)
	if err != nil {
		return fmt.Errorf("failed to generate grafana datasource secret: %v", err)
	}

	// get the grafana objects
	grafanaRenderer, grafanaDeployer := renderer.NewHoHRenderer(fs), deployer.NewHoHDeployer(r.Client)
	grafanaObjects, err := grafanaRenderer.Render("manifests/grafana", "", func(profile string) (interface{}, error) {
		return struct {
			Namespace            string
			SessionSecret        string
			ProxyImage           string
			DatasourceSecretName string
		}{
			Namespace:            config.GetDefaultNamespace(),
			SessionSecret:        proxySessionSecret,
			ProxyImage:           config.GetImage("oauth_proxy"),
			DatasourceSecretName: datasourceSecretName,
		}, nil
	})
	if err != nil {
		return fmt.Errorf("failed to render grafana manifests: %w", err)
	}

	// create restmapper for deployer to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(r.Manager.GetConfig())
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	if err = r.manipulateObj(ctx, grafanaDeployer, mapper, grafanaObjects, mgh,
		condition.SetConditionDatabaseInit, log); err != nil {
		return err
	}

	log.Info("Grafana initialized")
	if err := condition.SetConditionGrafanaInit(ctx, r.Client, mgh,
		condition.CONDITION_STATUS_TRUE); err != nil {
		return condition.FailToSetConditionError(condition.CONDITION_STATUS_TRUE, err)
	}
	return nil
}

// GenerateGrafanaDataSource is used to generate the GrafanaDatasource as a secret.
// the GrafanaDatasource points to multicluster-global-hub cr
func (r *MulticlusterGlobalHubReconciler) GenerateGrafanaDataSourceSecret(
	ctx context.Context,
	c client.Client,
	mgh *operatorv1alpha2.MulticlusterGlobalHub,
	scheme *runtime.Scheme,
) (string, error) {
	log := ctrllog.FromContext(ctx)
	// get the grafana data source
	postgresSecret, err := r.KubeClient.CoreV1().Secrets(config.GetDefaultNamespace()).Get(
		ctx,
		mgh.Spec.DataLayer.LargeScale.Postgres.Name,
		metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	postgresURI := string(postgresSecret.Data["database_uri"])
	objURI, err := url.Parse(postgresURI)
	if err != nil {
		return "", err
	}
	password, ok := objURI.User.Password()
	if !ok {
		return "", fmt.Errorf("failed to get password from database_uri: %s", postgresURI)
	}

	// get the database from the objURI
	database := "hoh"
	paths := strings.Split(objURI.Path, "/")
	if len(paths) > 1 {
		database = paths[1]
	}

	grafanaDatasources, err := yaml.Marshal(GrafanaDatasources{
		APIVersion: 1,
		Datasources: []*GrafanaDatasource{
			{
				Name:      "GlobalHubDataSource",
				Type:      "postgres",
				Access:    "proxy",
				IsDefault: true,
				URL:       objURI.Host,
				User:      objURI.User.Username(),
				Database:  database,
				Editable:  false,
				JSONData: &JsonData{
					SSLMode:      "require",
					QueryTimeout: "300s",
					TimeInterval: "30s",
					TLSAuth:      true,
					TLSAuthCA:    true,
				},
				SecureJSONData: &SecureJsonData{
					Password:      password,
					TLSCACert:     string(postgresSecret.Data["ca.crt"]),
					TLSClientCert: string(postgresSecret.Data["tls.crt"]),
					TLSClientKey:  string(postgresSecret.Data["tls.key"]),
				},
			},
		},
	})
	if err != nil {
		return "", err
	}

	dsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "multicluster-global-hub-grafana-datasources",
			Namespace: config.GetDefaultNamespace(),
			Labels: map[string]string{
				"datasource/time-tarted":         time.Now().Format("2006-1-2.1504"),
				constants.GlobalHubOwnerLabelKey: constants.GHOperatorOwnerLabelVal,
			},
		},
		Type: "Opaque",
		Data: map[string][]byte{
			datasourceKey: grafanaDatasources,
		},
	}

	// Set MGH instance as the owner and controller
	if err = controllerutil.SetControllerReference(mgh, dsSecret, scheme); err != nil {
		return dsSecret.GetName(), err
	}

	// Check if this already exists
	grafanaDSFound := &corev1.Secret{}
	err = c.Get(ctx, types.NamespacedName{Name: dsSecret.Name, Namespace: dsSecret.Namespace}, grafanaDSFound)

	if err != nil && errors.IsNotFound(err) {
		log.Info("create grafana datasource secret")
		if err := c.Create(ctx, dsSecret); err != nil {
			return dsSecret.GetName(), err
		}
	} else if err != nil {
		return dsSecret.GetName(), err
	}

	return dsSecret.GetName(), nil
}

type GrafanaDatasources struct {
	APIVersion  int                  `yaml:"apiVersion,omitempty"`
	Datasources []*GrafanaDatasource `yaml:"datasources,omitempty"`
}

type GrafanaDatasource struct {
	Access            string          `yaml:"access,omitempty"`
	BasicAuth         bool            `yaml:"basicAuth,omitempty"`
	BasicAuthPassword string          `yaml:"basicAuthPassword,omitempty"`
	BasicAuthUser     string          `yaml:"basicAuthUser,omitempty"`
	Editable          bool            `yaml:"editable,omitempty"`
	IsDefault         bool            `yaml:"isDefault,omitempty"`
	Name              string          `yaml:"name,omitempty"`
	OrgID             int             `yaml:"orgId,omitempty"`
	Type              string          `yaml:"type,omitempty"`
	URL               string          `yaml:"url,omitempty"`
	Database          string          `yaml:"database,omitempty"`
	User              string          `yaml:"user,omitempty"`
	Version           int             `yaml:"version,omitempty"`
	JSONData          *JsonData       `yaml:"jsonData,omitempty"`
	SecureJSONData    *SecureJsonData `yaml:"secureJsonData,omitempty"`
}

type JsonData struct {
	SSLMode       string `yaml:"sslmode,omitempty"`
	TLSAuth       bool   `yaml:"tlsAuth,omitempty"`
	TLSAuthCA     bool   `yaml:"tlsAuthWithCACert,omitempty"`
	TLSSkipVerify bool   `yaml:"tlsSkipVerify,omitempty"`
	QueryTimeout  string `yaml:"queryTimeout,omitempty"`
	HttpMethod    string `yaml:"httpMethod,omitempty"`
	TimeInterval  string `yaml:"timeInterval,omitempty"`
}

type SecureJsonData struct {
	Password      string `yaml:"password,omitempty"`
	TLSCACert     string `yaml:"tlsCACert,omitempty"`
	TLSClientCert string `yaml:"tlsClientCert,omitempty"`
	TLSClientKey  string `yaml:"tlsClientKey,omitempty"`
}

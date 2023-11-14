package managedclusters

import (
	"fmt"

	clusterV1 "open-cluster-management.io/api/cluster/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/stolostron/multicluster-global-hub/agent/pkg/status/controller/config"
	"github.com/stolostron/multicluster-global-hub/agent/pkg/status/controller/generic"
	"github.com/stolostron/multicluster-global-hub/pkg/bundle"
	genericbundle "github.com/stolostron/multicluster-global-hub/pkg/bundle/generic"
	"github.com/stolostron/multicluster-global-hub/pkg/constants"
	"github.com/stolostron/multicluster-global-hub/pkg/transport"
	"github.com/stolostron/multicluster-global-hub/pkg/utils"
)

const (
	clusterStatusSyncLogName = "clusters-status-sync"
)

// mgr, pro, env.LeafHubID, incarnation, config, syncIntervals
// AddMangedClusterSyncer adds managed clusters status controller to the manager.
func AddMangedClusterSyncer(mgr ctrl.Manager, producer transport.Producer) error {
	createObjFunction := func() bundle.Object { return &clusterV1.ManagedCluster{} }
	leafHubName := config.GetLeafHubName()

	transportBundleKey := fmt.Sprintf("%s.%s", leafHubName, constants.ManagedClustersMsgKey)

	// update bundle object
	manipulateObjFunc := func(object bundle.Object) {
		utils.AddAnnotations(object, map[string]string{
			constants.ManagedClusterManagedByAnnotation: leafHubName,
		})
	}

	predicateFunc := func() bool {
		// return hubOfHubsConfig.Data["aggregationLevel"] == "full" ||
		// 	hubOfHubsConfig.Data["aggregationLevel"] == "minimal"
		// at this point send all managed clusters even if aggregation level is minimal
		return true
	}

	bundleCollection := []*generic.BundleEntry{ // single bundle for managed clusters
		generic.NewBundleEntry(transportBundleKey,
			genericbundle.NewStatusGenericBundle(leafHubName, manipulateObjFunc),
			predicateFunc),
	}

	return generic.NewStatusGenericSyncer(mgr, clusterStatusSyncLogName, producer, bundleCollection,
		createObjFunction, nil, config.GetManagerClusterDuration)
}

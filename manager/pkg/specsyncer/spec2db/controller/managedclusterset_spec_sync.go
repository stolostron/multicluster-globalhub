// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package controller

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1beta1 "open-cluster-management.io/api/cluster/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/stolostron/multicluster-global-hub/manager/pkg/specsyncer/db2transport/db"
	"github.com/stolostron/multicluster-global-hub/pkg/constants"
)

func AddManagedClusterSetController(mgr ctrl.Manager, specDB db.SpecDB) error {
	managedclustersetPredicate, _ := predicate.LabelSelectorPredicate(metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      constants.GlobalHubLocalResource,
				Operator: metav1.LabelSelectorOpDoesNotExist,
			},
		},
	})
	if err := ctrl.NewControllerManagedBy(mgr).
		For(&clusterv1beta1.ManagedClusterSet{}).
		WithEventFilter(managedclustersetPredicate).
		Complete(&genericSpecToDBReconciler{
			client:         mgr.GetClient(),
			specDB:         specDB,
			log:            ctrl.Log.WithName("managedclustersets-spec-syncer"),
			tableName:      "managedclustersets",
			finalizerName:  constants.GlobalHubCleanupFinalizer,
			createInstance: func() client.Object { return &clusterv1beta1.ManagedClusterSet{} },
			cleanObject:    cleanManagedClusterSetStatus,
			areEqual:       areManagedClusterSetsEqual,
		}); err != nil {
		return fmt.Errorf("failed to add managed cluster set controller to the manager: %w", err)
	}

	return nil
}

func cleanManagedClusterSetStatus(instance client.Object) {
	managedClusterSet, ok := instance.(*clusterv1beta1.ManagedClusterSet)

	if !ok {
		panic("wrong instance passed to cleanManagedClusterSetStatus: not a ManagedClusterSet")
	}

	managedClusterSet.Status = clusterv1beta1.ManagedClusterSetStatus{}
}

func areManagedClusterSetsEqual(instance1, instance2 client.Object) bool {
	managedClusterSet1, ok1 := instance1.(*clusterv1beta1.ManagedClusterSet)
	managedClusterSet2, ok2 := instance2.(*clusterv1beta1.ManagedClusterSet)

	if !ok1 || !ok2 {
		return false
	}

	specMatch := equality.Semantic.DeepEqual(managedClusterSet1.Spec, managedClusterSet2.Spec)
	annotationsMatch := equality.Semantic.DeepEqual(instance1.GetAnnotations(), instance2.GetAnnotations())
	labelsMatch := equality.Semantic.DeepEqual(instance1.GetLabels(), instance2.GetLabels())

	return specMatch && annotationsMatch && labelsMatch
}

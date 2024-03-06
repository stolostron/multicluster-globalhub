package syncers_test

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clustersv1beta1 "open-cluster-management.io/api/cluster/v1beta1"
	policyv1 "open-cluster-management.io/governance-policy-propagator/api/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/stolostron/multicluster-global-hub/manager/pkg/specsyncer/db2transport/bundle"
	"github.com/stolostron/multicluster-global-hub/pkg/constants"
	"github.com/stolostron/multicluster-global-hub/pkg/transport"
	"github.com/stolostron/multicluster-global-hub/pkg/utils"
)

var _ = Describe("GenericBundle", func() {
	It("sync placement bundle", func() {
		By("Create Bundle with placement")
		baseBundle := bundle.NewBaseObjectsBundle()
		// Upgrade the placement crd after this PR: https://github.com/open-cluster-management-io/api/pull/242
		placement := &clustersv1beta1.Placement{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Placement",
				APIVersion: "cluster.open-cluster-management.io/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-placements",
				Namespace: "default",
				Annotations: map[string]string{
					constants.OriginOwnerReferenceAnnotation: "2aa5547c-c172-47ed-b70b-db468c84d327",
				},
			},
			Spec: clustersv1beta1.PlacementSpec{
				ClusterSets:      []string{"cluster1", "cluster2"},
				DecisionStrategy: clustersv1beta1.DecisionStrategy{},
			},
		}
		baseBundle.AddObject(placement, uuid.New().String())

		By("Send Placement Bundle by transport")
		payloadBytes, err := json.Marshal(baseBundle)
		Expect(err).NotTo(HaveOccurred())

		evt := utils.ToCloudEvent("Placements", transport.Broadcast, payloadBytes)
		err = producer.SendEvent(ctx, evt)
		Expect(err).NotTo(HaveOccurred())

		By("Check the placement is synced")
		Eventually(func() error {
			syncedPlacement := &clustersv1beta1.Placement{}
			err := client.Get(ctx, runtimeclient.ObjectKeyFromObject(placement), syncedPlacement)
			if err == nil {
				fmt.Println("create spec resource:")
				utils.PrettyPrint(syncedPlacement)
			}
			return err
		}, 10*time.Second, 100*time.Millisecond).ShouldNot(HaveOccurred())
	})

	It("sync placementbinding bundle", func() {
		By("Create Bundle with placementbinding")
		baseBundle := bundle.NewBaseObjectsBundle()
		// Upgrade the placementbinding crd after this PR:
		// https://github.com/open-cluster-management-io/governance-policy-propagator/pull/110
		placementbinding := &policyv1.PlacementBinding{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PlacementBinding",
				APIVersion: "policy.open-cluster-management.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-placementbinding",
				Namespace: "default",
				Annotations: map[string]string{
					constants.OriginOwnerReferenceAnnotation: "2aa5547c-c172-47ed-b70b-db468c84d327",
				},
			},
			PlacementRef: policyv1.PlacementSubject{
				APIGroup: "cluster.open-cluster-management.io",
				Kind:     "Placement",
				Name:     "placement-policy-limitrange",
			},
			Subjects: []policyv1.Subject{
				{
					APIGroup: "policy.open-cluster-management.io",
					Kind:     "Policy",
					Name:     "policy-limitrange",
				},
			},
		}
		baseBundle.AddObject(placementbinding, uuid.New().String())

		By("Send Placementbinding Bundle by transport")
		payloadBytes, err := json.Marshal(baseBundle)
		Expect(err).NotTo(HaveOccurred())

		evt := utils.ToCloudEvent("Placementbinding", transport.Broadcast, payloadBytes)
		err = producer.SendEvent(ctx, evt)
		Expect(err).NotTo(HaveOccurred())

		By("Check the placementbinding is synced")
		Eventually(func() error {
			syncedPlacementbinding := &policyv1.PlacementBinding{}
			err := client.Get(ctx, runtimeclient.ObjectKeyFromObject(placementbinding), syncedPlacementbinding)
			if err == nil {
				if err == nil {
					fmt.Println("create spec resource:")
					utils.PrettyPrint(syncedPlacementbinding)
				}
			}
			return err
		}, 10*time.Second, 100*time.Millisecond).ShouldNot(HaveOccurred())
	})
})

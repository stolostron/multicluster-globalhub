package tests

import (
	"context"
	"fmt"
	"strings"
	"time"
	"encoding/json"

	"github.com/jackc/pgx/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stolostron/multicluster-global-hub/pkg/constants"
	"github.com/stolostron/multicluster-global-hub/pkg/database"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	clusterv1beta1 "open-cluster-management.io/api/cluster/v1beta1"
	clusterv1beta2 "open-cluster-management.io/api/cluster/v1beta2"
	policiesv1 "open-cluster-management.io/governance-policy-propagator/api/v1"
	placementrulev1 "open-cluster-management.io/multicloud-operators-subscription/pkg/apis/apps/placementrule/v1"
	appsv1 "open-cluster-management.io/multicloud-operators-subscription/pkg/apis/apps/v1"
	appsv1alpha1 "open-cluster-management.io/multicloud-operators-subscription/pkg/apis/apps/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	PLACEMENT_POLICY_YAML       = "../../resources/policy/inform-limitrange-policy-placement.yaml"
	PLACEMENT_APP_SUB_YAML      = "../../resources/app/app-helloworld-appsub-placement.yaml"
	PLACEMENT_LOCAL_POLICY_YAML = "../../resources/policy/local-inform-limitrange-policy-placement.yaml"

	// PLACEMENT_APP    = "../../resources/policy/enforce-limitrange-policy.yaml"
	CLUSTERSET_LABEL_KEY = "cluster.open-cluster-management.io/clusterset"
)

var _ = Describe("Apply policy/app with placement on the global hub", Ordered, Label("e2e-tests-placement"), func() {
	var managedCluster1 clusterv1.ManagedCluster
	var managedCluster2 clusterv1.ManagedCluster
	var globalClient, leafhubClient client.Client
	// var policyPlacementName string
	var policyName, policyNamespace, policyClusterset string
	var localPolicyName, localPolicyNamespace, localPlacementName, localPolicyLabelKey, localPolicyLabelVal string
	var postgresConn *pgx.Conn

	BeforeAll(func() {
		By("Initialize the variables")
		policyName = "policy-limitrange"
		policyNamespace = "global-placement"

		localPolicyName = "policy-limitrange" // mclset/mclsetbinding: default
		localPolicyNamespace = "local-placement"
		localPlacementName = "placement-policy-limitrange"
		localPolicyLabelKey = "local-policy-placement"
		localPolicyLabelVal = "test"

		policyClusterset = "clusterset1"
		Eventually(func() error {
			managedClusters, err := getManagedCluster(httpClient, httpToken)
			if err != nil {
				return err
			}
			managedCluster1 = managedClusters[0]
			managedCluster2 = managedClusters[1]
			return nil
		}, 1*time.Minute, 1*time.Second).ShouldNot(HaveOccurred())
		fmt.Println(clients.LeafHubClusterName())

		By("Init the client")
		scheme := runtime.NewScheme()
		Expect(policiesv1.AddToScheme(scheme))
		Expect(corev1.AddToScheme(scheme))
		Expect(placementrulev1.AddToScheme(scheme))
		Expect(clusterv1beta1.AddToScheme(scheme))
		Expect(clusterv1beta2.AddToScheme(scheme))
		Expect(clusterv1.AddToScheme(scheme))
		Expect(appsv1.SchemeBuilder.AddToScheme(scheme))
		Expect(appsv1alpha1.AddToScheme(scheme))
		var err error
		globalClient, err = clients.ControllerRuntimeClient(clients.HubClusterName(), scheme)
		Expect(err).ShouldNot(HaveOccurred())
		leafhubClient, err = clients.ControllerRuntimeClient((clients.LeafHubClusterName()), scheme)
		Expect(err).ShouldNot(HaveOccurred())

		By("Create Postgres connection")
		databaseURI := strings.Split(testOptions.HubCluster.DatabaseURI, "?")[0]
		postgresConn, err = database.PostgresConnection(context.TODO(), databaseURI, nil)
		Expect(err).Should(Succeed())
	})

	Context("When apply local policy with placement on the regional hub", func() {
		It("deploy local policy on the regional hub", func() {
			By("Add local policy test label")
			patches := []patch{
				{
					Op:    "add", // or remove
					Path:  "/metadata/labels/" + localPolicyLabelKey,
					Value: localPolicyLabelVal,
				},
			}
			Expect(updateClusterLabel(httpClient, patches, httpToken, string(managedCluster1.UID))).Should(Succeed())

			By("Deploy the placement policy to the leafhub")
			output, err := clients.Kubectl(clients.LeafHubClusterName(), "apply", "-f", PLACEMENT_LOCAL_POLICY_YAML)
			klog.V(5).Info(fmt.Sprintf("deploy inform local policy: %s", output))
			Expect(err).Should(Succeed())

			By("Verify the local policy is directly synchronized to the global hub spec table")
			policy := &policiesv1.Policy{}
			Eventually(func() error {
				rows, err := postgresConn.Query(context.TODO(), "select payload from local_spec.policies")
				if err != nil {
					return err
				}
				defer rows.Close()
				for rows.Next() {
					if err := rows.Scan(policy); err != nil {
						return err
					}
					fmt.Printf("local_spec.policies: %s/%s \n", policy.Namespace, policy.Name)
					if policy.Name == localPolicyName && policy.Namespace == localPolicyNamespace {
						return nil
					}
				}
				return fmt.Errorf("expect policy(placement) [%s/%s] but got [%s/%s]", localPolicyNamespace,
					localPolicyName, policy.Namespace, policy.Name)
			}, 1*time.Minute, 1*time.Second).Should(Succeed())

			By("Verify the local placement policy is synchronized to the global hub status table")
			Eventually(func() error {
				rows, err := postgresConn.Query(context.TODO(),
					"SELECT id,cluster_name,leaf_hub_name FROM local_status.compliance")
				if err != nil {
					return err
				}
				defer rows.Close()

				for rows.Next() {
					columnValues, _ := rows.Values()
					if len(columnValues) < 3 {
						return fmt.Errorf("the compliance record is not correct, expected 5 but got %d", len(columnValues))
					}
					policyId, cluster, leafhub := "", "", ""
					if err := rows.Scan(&policyId, &cluster, &leafhub); err != nil {
						return err
					}
					// 18303079-73a6-4f95-8863-e57ef14f5c6d kind-hub2 kind-hub2-cluster1
					// 90bfd360-c66d-4089-a507-22c8417cc793 kind-hub1 kind-hub1-cluster1
					// c48c5d3f-929a-4743-abe5-dc0941fe37b3 kind-hub1 kind-hub1-cluster1
					// kind-hub1
					// kind-hub1-cluster1
					fmt.Println(policyId, leafhub, cluster)
					fmt.Println("#########")
					fmt.Println(clients.LeafHubClusterName())
					fmt.Println(managedCluster1.Name)
					fmt.Println("*********")
					if policyId == string(policy.UID) && leafhub == clients.LeafHubClusterName() &&
						cluster == managedCluster1.Name {
						return nil
					}
				}
				return fmt.Errorf("the policy(placement) is not synchronized to the global hub status table")
			}, 1*time.Minute, 1*time.Second).Should(Succeed())
		})

		// to use the finalizer achieves deleting local resource from database:
		// finalizer(deprecated) -> delete from bundle -> transport -> database
		It("check the local policy(placement) resource isn't added the global cleanup finalizer", func() {
			By("Verify the local policy(placement) hasn't been added the global hub cleanup finalizer")
			Eventually(func() error {
				policy := &policiesv1.Policy{}
				err := leafhubClient.Get(context.TODO(), client.ObjectKey{
					Namespace: localPolicyNamespace,
					Name:      localPolicyName,
				}, policy)
				if err != nil {
					return err
				}
				// "clustername":"kind-hub1-cluster1","clusternamespace":"kind-hub1-cluster1"
				fmt.Println("policy: ")
				js, _ := json.Marshal(policy)
				fmt.Println(string(js))
				fmt.Println("$$$$$$$$$")
				for _, finalizer := range policy.Finalizers {
					if finalizer == constants.GlobalHubCleanupFinalizer {
						return fmt.Errorf("the local policy(%s) has been added the cleanup finalizer", policy.GetName())
					}
				}
				return nil
			}, 1*time.Minute, 1*time.Second).Should(Succeed())

			// placement is not be synchronized to the global hub database, so it doesn't need the finalizer
			By("Verify the local placement hasn't been added the global hub cleanup finalizer")
			Eventually(func() error {
				placement := &clusterv1beta1.Placement{}
				err := leafhubClient.Get(context.TODO(), client.ObjectKey{
					Namespace: localPolicyNamespace,
					Name:      localPlacementName,
				}, placement)
				if err != nil {
					return err
				}
				// "local-policy-placement": "test"
				fmt.Println("placement: ")
				js, _ := json.Marshal(placement)
				fmt.Println(string(js))
				fmt.Println("^^^^^^^^^")
				for _, finalizer := range placement.Finalizers {
					if finalizer == constants.GlobalHubCleanupFinalizer {
						return fmt.Errorf("the local placement(%s) has been added the cleanup finalizer",
							placement.GetName())
					}
				}
				return nil
			}, 1*time.Minute, 1*time.Second).Should(Succeed())
		})

		It("delete the local policy(placement) from the leafhub", func() {
			By("Delete the local policy from leafhub")
			output, err := clients.Kubectl(clients.LeafHubClusterName(), "delete", "-f", PLACEMENT_LOCAL_POLICY_YAML)
			fmt.Println(output)
			Expect(err).Should(Succeed())

			By("Verify the local policy(placement) is deleted from the spec table")
			Eventually(func() error {
				rows, err := postgresConn.Query(context.TODO(), "select payload from local_spec.policies")
				if err != nil {
					return err
				}
				defer rows.Close()
				policy := &policiesv1.Policy{}
				for rows.Next() {
					if err := rows.Scan(policy); err != nil {
						return err
					}
					// "uid":"18303079-73a6-4f95-8863-e57ef14f5c6d" "disabled": false, "status": {}
					// "uid":"90bfd360-c66d-4089-a507-22c8417cc793"
					fmt.Println("policy: ")
					js, _ := json.Marshal(policy)
					fmt.Println(string(js))
					fmt.Println("$$$$$$$$$")
					if policy.Name == localPolicyName && policy.Namespace == localPolicyNamespace {
						return fmt.Errorf("the policy(%s) is not deleted from local_spec.policies", policy.GetName())
					}
				}
				return nil
			}, 1*time.Minute, 1*time.Second).Should(Succeed())

			By("Verify the local policy(placement) is deleted from the global hub status table")
			Eventually(func() error {
				rows, err := postgresConn.Query(context.TODO(),
						"SELECT id,cluster_name,leaf_hub_name FROM local_status.compliance")
				if err != nil {
					return err
				}
				defer rows.Close()

				for rows.Next() {
					columnValues, _ := rows.Values()
					if len(columnValues) < 3 {
						return fmt.Errorf("the compliance record is not correct, expected 5 but got %d", len(columnValues))
					}
					policyId, cluster, leafhub := "", "", ""
					if err := rows.Scan(&policyId, &cluster, &leafhub); err != nil {
						return err
					}
					fmt.Println(policyId, leafhub, cluster)
				}
				return nil

				// origin
				// rows, err := postgresConn.Query(context.TODO(), "SELECT id,cluster_name,leaf_hub_name FROM local_status.compliance")
				// if err != nil {
				// 	return err
				// }
				// defer rows.Close()

				// for rows.Next() {
				// 	fmt.Println("in forloop")
				// 	columnValues, _ := rows.Values()
				// 	fmt.Println(columnValues)
				// 	return fmt.Errorf("the policy(%s) is not deleted from local_status.compliance", columnValues)
				// }
				// return nil
			}, 1*time.Minute, 1*time.Second).Should(Succeed())

			By("Remove local policy test label")
			patches := []patch{
				{
					Op:    "remove",
					Path:  "/metadata/labels/" + localPolicyLabelKey,
					Value: localPolicyLabelVal,
				},
			}
			Expect(updateClusterLabel(httpClient, patches, httpToken, string(managedCluster1.UID))).Should(Succeed())
		})
	})

	Context("When apply global policy with placement on the global hub", func() {
		It("add managedCluster2 to the clusterset1", func() {
			patches := []patch{
				{
					Op:    "add", // or remove
					Path:  "/metadata/labels/" + CLUSTERSET_LABEL_KEY,
					Value: policyClusterset,
				},
			}
			Expect(updateClusterLabel(httpClient, patches, httpToken, string(managedCluster2.UID))).Should(Succeed())
			Eventually(func() error {
				managedCluster, err := getManagedClusterByName(httpClient, httpToken, managedCluster2.Name)
				if err != nil {
					return err
				}
				if val, ok := managedCluster.Labels[CLUSTERSET_LABEL_KEY]; ok && val == policyClusterset {
					return nil
				}
				return fmt.Errorf("the label %s: %s is not exist on mcl2", CLUSTERSET_LABEL_KEY, policyClusterset)
			}, 1*time.Minute, 1*time.Second).ShouldNot(HaveOccurred())
		})

		It("apply policy with placement", func() {
			By("Deploy the policy to the global hub")
			output, err := clients.Kubectl(clients.HubClusterName(), "apply", "-f", PLACEMENT_POLICY_YAML)
			klog.V(5).Info(fmt.Sprintf("deploy inform policy with placement: %s", output))
			Expect(err).Should(Succeed())

			By("Check the inform policy in global hub")
			Eventually(func() error {
				status, err := getPolicyStatus(globalClient, httpClient, policyName, policyNamespace, httpToken)
				if err != nil {
					return err
				}
				if len(status.Status) == 1 && status.Status[0].ComplianceState == "NonCompliant" &&
					status.Status[0].ClusterName == managedCluster2.Name {
					return nil
				}
				return fmt.Errorf("the policy have not applied to the managed cluster %s", managedCluster2.Name)
			}, 1*time.Minute, 1*time.Second).Should(Succeed())
		})

		It("scale policy with placement", func() {
			patches := []patch{
				{
					Op:    "remove", // or remove
					Path:  "/metadata/labels/" + CLUSTERSET_LABEL_KEY,
					Value: policyClusterset,
				},
			}
			Expect(updateClusterLabel(httpClient, patches, httpToken, string(managedCluster2.UID))).Should(Succeed())

			By("Check the inform policy in global hub")
			Eventually(func() error {
				status, err := getPolicyStatus(globalClient, httpClient, policyName, policyNamespace, httpToken)
				if err != nil {
					return err
				}
				fmt.Println(status.Status)
				if len(status.Status) == 0 {
					return nil
				}
				return fmt.Errorf("the policy should removed from managed cluster %s", managedCluster2.Name)
			}, 1*time.Minute, 1*time.Second).Should(Succeed())
		})

		It("delete policy with placement", func() {
			By("Delete the policy in the global hub")
			output, err := clients.Kubectl(clients.HubClusterName(), "delete", "-f", PLACEMENT_POLICY_YAML)
			klog.V(5).Info(fmt.Sprintf("delete inform policy with placement: %s", output))
			Expect(err).Should(Succeed())

			By("Check the inform policy in global hub")
			Eventually(func() error {
				_, err := getPolicyStatus(globalClient, httpClient, policyName, policyNamespace, httpToken)
				if errors.IsNotFound(err) {
					return nil
				}
				if err != nil {
					return err
				}
				return fmt.Errorf("the policy should be removed from global hub")
			}, 1*time.Minute, 1*time.Second).Should(Succeed())
		})
	})

	Context("When apply global application with placement on the global hub", func() {
		It("deploy application with placement", func() {
			By("Add app label to the managedClusters")
			patches := []patch{
				{
					Op:    "add",
					Path:  "/metadata/labels/" + APP_LABEL_KEY,
					Value: APP_LABEL_VALUE,
				},
			}
			Expect(updateClusterLabel(httpClient, patches, httpToken, string(managedCluster1.UID))).Should(Succeed())
			Expect(updateClusterLabel(httpClient, patches, httpToken, string(managedCluster2.UID))).Should(Succeed())

			By("Apply the appsub to labeled clusters")
			Eventually(func() error {
				_, err := clients.Kubectl(clients.HubClusterName(), "apply", "-f", PLACEMENT_APP_SUB_YAML)
				if err != nil {
					return err
				}
				return nil
			}, 1*time.Minute, 1*time.Second).Should(Succeed())

			By("Check the appsub is applied to the cluster")
			Eventually(func() error {
				return checkAppsubreport(globalClient, httpClient, APP_SUB_NAME, APP_SUB_NAMESPACE, httpToken, 2,
					[]string{managedCluster1.Name, managedCluster2.Name})
			}, 3*time.Minute, 1*time.Second).Should(Succeed())
		})

		It("scale application with placement", func() {
			By("Move managedCluster2 to the clusterset1")
			patches := []patch{
				{
					Op:    "add",
					Path:  "/metadata/labels/" + CLUSTERSET_LABEL_KEY,
					Value: policyClusterset,
				},
			}
			Expect(updateClusterLabel(httpClient, patches, httpToken, string(managedCluster2.UID))).Should(Succeed())

			By("Check the appsub is applied to the cluster")
			Eventually(func() error {
				return checkAppsubreport(globalClient, httpClient, APP_SUB_NAME, APP_SUB_NAMESPACE, httpToken, 1,
					[]string{managedCluster1.Name})
			}, 3*time.Minute, 1*time.Second).Should(Succeed())
		})

		It("delete application with placement", func() {
			By("Delete the appsub")
			_, err := clients.Kubectl(clients.HubClusterName(), "delete", "-f", PLACEMENT_APP_SUB_YAML)
			Expect(err).Should(Succeed())

			By("Move managedCluster2 to the default clusterset")
			patches := []patch{
				{
					Op:    "remove",
					Path:  "/metadata/labels/" + CLUSTERSET_LABEL_KEY,
					Value: policyClusterset,
				},
			}
			Expect(updateClusterLabel(httpClient, patches, httpToken, string(managedCluster2.UID))).Should(Succeed())

			By("Remove app label")
			patches = []patch{
				{
					Op:    "remove",
					Path:  "/metadata/labels/" + APP_LABEL_KEY,
					Value: APP_LABEL_VALUE,
				},
			}
			Expect(updateClusterLabel(httpClient, patches, httpToken, string(managedCluster1.UID))).Should(Succeed())
			Expect(updateClusterLabel(httpClient, patches, httpToken, string(managedCluster2.UID))).Should(Succeed())

			By("manually remove the appsubreport in the regional hub") // TODO: remove this step after the issue is fixed
			appsubreport := &appsv1alpha1.SubscriptionReport{
				ObjectMeta: metav1.ObjectMeta{
					Name:      managedCluster1.Name,
					Namespace: managedCluster1.Name,
				},
			}
			Expect(leafhubClient.Delete(context.TODO(), appsubreport, &client.DeleteOptions{})).Should(Succeed())
			appsubreport.SetName(managedCluster2.Name)
			appsubreport.SetNamespace(managedCluster2.Name)
			Expect(leafhubClient.Delete(context.TODO(), appsubreport, &client.DeleteOptions{})).Should(Succeed())

			By("Check the appsub is deleted")
			Eventually(func() error {
				appsub := &appsv1.Subscription{}
				err := globalClient.Get(context.TODO(), types.NamespacedName{Namespace: APP_SUB_NAMESPACE, Name: APP_SUB_NAME},
					appsub)
				if err != nil && !errors.IsNotFound(err) {
					return err
				} else if err == nil {
					return fmt.Errorf("the appsub is not deleted from global hub")
				}

				rows, err := postgresConn.Query(context.TODO(), "select payload from status.subscription_reports")
				if err != nil {
					return err
				}
				defer rows.Close()
				appsubreport := &appsv1alpha1.SubscriptionReport{}
				for rows.Next() {
					if err := rows.Scan(appsubreport); err != nil {
						return err
					}
					fmt.Printf("status.subscription_reports: %s/%s \n", appsubreport.Namespace, appsubreport.Name)
					if appsubreport.Name == APP_SUB_NAME && appsubreport.Namespace == APP_SUB_NAMESPACE {
						return fmt.Errorf("the appsub is not deleted from regional hub")
					}
				}
				return nil
			}, 2*time.Minute, 1*time.Second).Should(Succeed())
		})
	})

	AfterAll(func() {
		By("Close the postgresql connection")
		Expect(postgresConn.Close(context.Background())).Should(Succeed())
	})
})

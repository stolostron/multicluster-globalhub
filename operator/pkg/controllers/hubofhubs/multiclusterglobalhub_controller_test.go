/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hubofhubs

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/kylelemons/godebug/diff"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	operatorv1alpha2 "github.com/stolostron/multicluster-global-hub/operator/apis/v1alpha2"
	"github.com/stolostron/multicluster-global-hub/operator/pkg/condition"
	"github.com/stolostron/multicluster-global-hub/operator/pkg/config"
	"github.com/stolostron/multicluster-global-hub/operator/pkg/constants"
	"github.com/stolostron/multicluster-global-hub/operator/pkg/renderer"
	"github.com/stolostron/multicluster-global-hub/operator/pkg/utils"
)

// +kubebuilder:docs-gen:collapse=Imports

var _ = Describe("MulticlusterGlobalHub controller", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		MGHName              = "test-mgh"
		StorageSecretName    = "storage-secret"
		TransportSecretName  = "transport-secret"
		kafkaCA              = "foobar"
		kafkaBootstrapServer = "https://test-kafka.example.com"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When create MGH Instance", func() {
		It("Should create Multicluster Global Hub resources when MGH instance is created", func() {
			By("By creating a new MGH instance")
			ctx := context.Background()

			By("By creating fake storage secret and transport secret")
			storageSecret := &corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Secret",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      StorageSecretName,
					Namespace: config.GetDefaultNamespace(),
				},
				Data: map[string][]byte{
					"database_uri": []byte("postgres://postgres:testpwd@hoh-primary.hoh-postgres.svc:/hoh"),
				},
				Type: corev1.SecretTypeOpaque,
			}
			Expect(k8sClient.Create(ctx, storageSecret)).Should(Succeed())

			transportSecret := &corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Secret",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      TransportSecretName,
					Namespace: config.GetDefaultNamespace(),
				},
				Data: map[string][]byte{
					"CA":               []byte(kafkaCA),
					"bootstrap_server": []byte(kafkaBootstrapServer),
				},
				Type: corev1.SecretTypeOpaque,
			}
			Expect(k8sClient.Create(ctx, transportSecret)).Should(Succeed())

			mgh := &operatorv1alpha2.MulticlusterGlobalHub{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "operator.open-cluster-management.io/v1alpha2",
					Kind:       "MulticlusterGlobalHub",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      MGHName,
					Namespace: config.GetDefaultNamespace(),
					Annotations: map[string]string{
						constants.AnnotationMGHSkipDBInit: "true",
					},
				},
				Spec: operatorv1alpha2.MulticlusterGlobalHubSpec{
					DataLayer: operatorv1alpha2.DataLayerConfig{
						Type: "largeScale",
						LargeScale: &operatorv1alpha2.LargeScaleConfig{
							Kafka: corev1.LocalObjectReference{
								Name: TransportSecretName,
							},
							Postgres: corev1.LocalObjectReference{
								Name: StorageSecretName,
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, mgh)).Should(Succeed())

			// 	After creating this MGH instance, check that the MGH instance's Spec fields are failed with default values.
			mghLookupKey := types.NamespacedName{Namespace: config.GetDefaultNamespace(), Name: MGHName}
			createdMGH := &operatorv1alpha2.MulticlusterGlobalHub{}

			// get this newly created MGH instance, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, mghLookupKey, createdMGH)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// make sure the default values are filled
			Expect(createdMGH.Spec.AggregationLevel).Should(Equal(operatorv1alpha2.Full))
			Expect(createdMGH.Spec.EnableLocalPolicies).Should(Equal(true))
			Expect(createdMGH.Spec.DataLayer.LargeScale.Kafka.Name).Should(Equal(TransportSecretName))
			Expect(createdMGH.Spec.DataLayer.LargeScale.Postgres.Name).Should(Equal(StorageSecretName))

			By("By checking the MGH CR database init conditions are created as expected")
			condition.SetConditionDatabaseInit(ctx, k8sClient, createdMGH, condition.CONDITION_STATUS_TRUE)
			Expect(condition.GetConditionStatus(createdMGH,
				condition.CONDITION_TYPE_DATABASE_INIT)).Should(Equal(metav1.ConditionTrue))
			Eventually(func() bool {
				condition.SetConditionDatabaseInit(ctx, k8sClient, createdMGH, condition.CONDITION_STATUS_FALSE)
				return condition.GetConditionStatus(createdMGH,
					condition.CONDITION_TYPE_DATABASE_INIT) == metav1.ConditionFalse
			}, timeout, interval).Should(BeTrue())
			Eventually(func() bool {
				condition.SetConditionDatabaseInit(ctx, k8sClient, createdMGH, condition.CONDITION_STATUS_UNKNOWN)
				return condition.GetConditionStatus(createdMGH,
					condition.CONDITION_TYPE_DATABASE_INIT) == metav1.ConditionUnknown
			}, timeout, interval).Should(BeTrue())

			By("By checking the MGH CR transport init conditions are created as expected")
			condition.SetConditionTransportInit(ctx, k8sClient, createdMGH, condition.CONDITION_STATUS_UNKNOWN)
			Expect(condition.GetConditionStatus(createdMGH,
				condition.CONDITION_TYPE_TRANSPORT_INIT)).Should(Equal(metav1.ConditionUnknown))
			Eventually(func() bool {
				condition.SetConditionTransportInit(ctx, k8sClient, createdMGH, condition.CONDITION_STATUS_FALSE)
				return condition.GetConditionStatus(createdMGH,
					condition.CONDITION_TYPE_TRANSPORT_INIT) == metav1.ConditionFalse
			}, timeout, interval).Should(BeTrue())
			Eventually(func() bool {
				condition.SetConditionTransportInit(ctx, k8sClient, createdMGH, condition.CONDITION_STATUS_TRUE)
				return condition.GetConditionStatus(createdMGH,
					condition.CONDITION_TYPE_TRANSPORT_INIT) == metav1.ConditionTrue
			}, timeout, interval).Should(BeTrue())

			By("By checking the MGH CR manager deployed conditions are created as expected")
			condition.SetConditionManagerDeployed(ctx, k8sClient, createdMGH, condition.CONDITION_STATUS_FALSE)
			Expect(condition.GetConditionStatus(createdMGH,
				condition.CONDITION_TYPE_MANAGER_DEPLOY)).Should(Equal(metav1.ConditionFalse))
			Eventually(func() bool {
				condition.SetConditionManagerDeployed(ctx, k8sClient, createdMGH, condition.CONDITION_STATUS_TRUE)
				return condition.GetConditionStatus(createdMGH,
					condition.CONDITION_TYPE_MANAGER_DEPLOY) == metav1.ConditionTrue
			}, timeout, interval).Should(BeTrue())
			Eventually(func() bool {
				condition.SetConditionManagerDeployed(ctx, k8sClient, createdMGH, condition.CONDITION_STATUS_UNKNOWN)
				return condition.GetConditionStatus(createdMGH,
					condition.CONDITION_TYPE_MANAGER_DEPLOY) == metav1.ConditionUnknown
			}, timeout, interval).Should(BeTrue())

			By("By checking the MGH CR regional hub deployed conditions are created as expected")
			condition.SetConditionLeafHubDeployed(ctx, k8sClient, createdMGH, "test", condition.CONDITION_STATUS_TRUE)
			Expect(condition.GetConditionStatus(createdMGH,
				condition.CONDITION_TYPE_LEAFHUB_DEPLOY)).Should(Equal(metav1.ConditionTrue))
			Eventually(func() bool {
				condition.SetConditionLeafHubDeployed(ctx, k8sClient, createdMGH,
					"test", condition.CONDITION_STATUS_FALSE)
				return condition.GetConditionStatus(createdMGH,
					condition.CONDITION_TYPE_LEAFHUB_DEPLOY) == metav1.ConditionFalse
			}, timeout, interval).Should(BeTrue())
			Eventually(func() bool {
				condition.SetConditionLeafHubDeployed(ctx, k8sClient, createdMGH,
					"test", condition.CONDITION_STATUS_UNKNOWN)
				return condition.GetConditionStatus(createdMGH,
					condition.CONDITION_TYPE_LEAFHUB_DEPLOY) == metav1.ConditionUnknown
			}, timeout, interval).Should(BeTrue())

			// create hoh render for testing
			hohRenderer := renderer.NewHoHRenderer(fs)

			checkResource := func(ctx context.Context, k8sClient client.Client, unsObj *unstructured.Unstructured,
			) error {
				objLookupKey := types.NamespacedName{Name: unsObj.GetName(), Namespace: unsObj.GetNamespace()}
				gvk := unsObj.GetObjectKind().GroupVersionKind()
				foundObj := &unstructured.Unstructured{}
				foundObj.SetGroupVersionKind(unsObj.GetObjectKind().GroupVersionKind())
				if err := k8sClient.Get(ctx, objLookupKey, foundObj); err != nil {
					return err
				}

				// delete metadata before compare
				unsObjContent, foundObjContent := unsObj.Object, foundObj.Object
				delete(unsObjContent, "metadata")
				delete(foundObjContent, "metadata")
				if !apiequality.Semantic.DeepDerivative(unsObjContent, foundObjContent) {
					unsObjYaml, err := yaml.Marshal(unsObjContent)
					if err != nil {
						return err
					}
					foundObjYaml, err := yaml.Marshal(foundObjContent)
					if err != nil {
						return err
					}

					return fmt.Errorf("desired and found %s(%s) are not equal, difference:\n%v\n",
						gvk, objLookupKey, diff.Diff(string(unsObjYaml), string(foundObjYaml)))
				}

				fmt.Printf("desired and found %s(%s) are equal\n", gvk, objLookupKey)
				return nil
			}

			By("By checking the multicluster-global-hub-manager resources are created as expected")
			managerObjects, err := hohRenderer.Render("manifests/manager", func(
				component string,
			) (interface{}, error) {
				return struct {
					Image                string
					DBSecret             string
					KafkaCA              string
					KafkaBootstrapServer string
					Namespace            string
				}{
					Image:                config.GetImage("multicluster_global_hub_manager"),
					DBSecret:             mgh.Spec.DataLayer.LargeScale.Postgres.Name,
					KafkaCA:              base64.RawStdEncoding.EncodeToString([]byte(kafkaCA)),
					KafkaBootstrapServer: kafkaBootstrapServer,
					Namespace:            config.GetDefaultNamespace(),
				}, nil
			})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() bool {
				for _, unsObj := range managerObjects {
					err := checkResource(ctx, k8sClient, unsObj)
					if err != nil {
						fmt.Printf("failed to check manager resource: %v\n", err)
						return false
					}
				}
				return true
			}, timeout, interval).Should(BeTrue())

			By("By checking the multicluster-global-hub-rbac resources are created as expected")
			hohRBACObjects, err := hohRenderer.Render("manifests/rbac", func(
				component string,
			) (interface{}, error) {
				return struct {
					Image     string
					Namespace string
				}{
					Image:     config.GetImage("multicluster_global_hub_rbac"),
					Namespace: config.GetDefaultNamespace(),
				}, nil
			})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() bool {
				for _, unsObj := range hohRBACObjects {
					if unsObj.GetName() == "opa-data" {
						// skip opa-data secret check
						continue
					}
					err := checkResource(ctx, k8sClient, unsObj)
					if err != nil {
						fmt.Printf("failed to check rbac resource: %v\n", err)
						return false
					}
				}
				return true
			}, timeout, interval).Should(BeTrue())

			By("By checking the multicluster-global-hub-config configmap is created")
			hohConfig := &corev1.ConfigMap{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Namespace: constants.HOHSystemNamespace,
					Name:      constants.HOHConfigName,
				}, hohConfig)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("By deleting the multicluster-global-hub-config configmap")
			Expect(k8sClient.Delete(ctx, hohConfig)).Should(Succeed())

			By("By checking the multicluster-global-hub-config configmap is recreated")
			Eventually(func() bool {
				hohConfig := &corev1.ConfigMap{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Namespace: constants.HOHSystemNamespace,
					Name:      constants.HOHConfigName,
				}, hohConfig)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("By checking the kafkaBootstrapServer")
			server, _, err := utils.GetKafkaConfig(ctx, kubeClient, createdMGH)
			Expect(err).NotTo(HaveOccurred())
			Expect(server).To(Equal(kafkaBootstrapServer))

			By("By checking the kafka secret is deleted")
			Expect(k8sClient.Delete(ctx, transportSecret)).Should(Succeed())
			_, _, err = utils.GetKafkaConfig(ctx, kubeClient, createdMGH)
			Expect(err).To(HaveOccurred())

			By("By checking setting the conditions")
			Expect(len(createdMGH.GetConditions())).Should(BeNumerically(">", 0))
			createdMGH.SetConditions(append(createdMGH.GetConditions(), metav1.Condition{
				Type: "Test", Status: condition.CONDITION_STATUS_UNKNOWN, Reason: "this is a test",
				Message: "this is a test", LastTransitionTime: metav1.Time{Time: time.Now()},
			}))
			conditions := createdMGH.GetConditions()
			hasTestCondition := false
			for _, cond := range conditions {
				if cond.Type == "Test" {
					Expect(cond.Status).To(Equal(metav1.ConditionUnknown))
					Expect(cond.Reason).To(Equal("this is a test"))
					Expect(cond.Message).To(Equal("this is a test"))
					hasTestCondition = true
				}
			}
			Expect(hasTestCondition).To(BeTrue())
		})
	})
})

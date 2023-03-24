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

package config

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1alpha2 "github.com/stolostron/multicluster-global-hub/operator/apis/v1alpha2"
	operatorconstants "github.com/stolostron/multicluster-global-hub/operator/pkg/constants"
	"github.com/stolostron/multicluster-global-hub/pkg/constants"
)

func TestSetImageOverrides(t *testing.T) {
	mghInstanceName := "multiclusterglobalhub"
	tests := []struct {
		desc               string
		initImageManifests map[string]string
		operandImagesEnv   map[string]string
		mghInstance        *operatorv1alpha2.MulticlusterGlobalHub
		imageOverrideCM    *corev1.ConfigMap
		wantImageManifests map[string]string
		wantErr            error
	}{
		{
			desc: "no any override",
			initImageManifests: map[string]string{
				"multicluster_global_hub_agent":   "quay.io/stolostron/multicluster-global-hub-agent:latest",
				"multicluster_global_hub_manager": "quay.io/stolostron/multicluster-global-hub-manager:latest",
				"oauth_proxy":                     "quay.io/stolostron/multicluster-global-hub-operator:latest",
			},
			operandImagesEnv: map[string]string{
				"OPERAND_IMAGE_MULTICLUSTER_GLOBAL_HUB_MANAGER": "quay.io/stolostron/multicluster-global-hub-manager:v0.6.0",
				"OPERAND_IMAGE_MULTICLUSTER_GLOBAL_HUB_AGENT":   "quay.io/stolostron/multicluster-global-hub-agent:v0.6.0",
				"OPERAND_IMAGE_OAUTH_PROXY":                     "quay.io/stolostron/origin-oauth-proxy:4.9",
			},
			mghInstance: &operatorv1alpha2.MulticlusterGlobalHub{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: constants.GHDefaultNamespace,
					Name:      mghInstanceName,
				},
				Spec: operatorv1alpha2.MulticlusterGlobalHubSpec{},
			},
			imageOverrideCM: nil,
			wantImageManifests: map[string]string{
				"multicluster_global_hub_agent":   "quay.io/stolostron/multicluster-global-hub-agent:v0.6.0",
				"multicluster_global_hub_manager": "quay.io/stolostron/multicluster-global-hub-manager:v0.6.0",
				"oauth_proxy":                     "quay.io/stolostron/origin-oauth-proxy:4.9",
			},
			wantErr: nil,
		},
		{
			desc: "override image repo",
			initImageManifests: map[string]string{
				"multicluster_global_hub_agent":   "quay.io/stolostron/multicluster-global-hub-agent:latest",
				"multicluster_global_hub_manager": "quay.io/stolostron/multicluster-global-hub-manager:latest",
				"oauth_proxy":                     "quay.io/stolostron/multicluster-global-hub-operator:latest",
			},
			operandImagesEnv: map[string]string{
				"OPERAND_IMAGE_MULTICLUSTER_GLOBAL_HUB_MANAGER": "quay.io/stolostron/multicluster-global-hub-manager:v0.6.0",
				"OPERAND_IMAGE_MULTICLUSTER_GLOBAL_HUB_AGENT":   "quay.io/stolostron/multicluster-global-hub-agent:v0.6.0",
				"OPERAND_IMAGE_OAUTH_PROXY":                     "quay.io/stolostron/origin-oauth-proxy:4.9",
			},
			mghInstance: &operatorv1alpha2.MulticlusterGlobalHub{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: constants.GHDefaultNamespace,
					Name:      mghInstanceName,
					Annotations: map[string]string{
						operatorconstants.AnnotationImageRepo: "quay.io/testing",
					},
				},
				Spec: operatorv1alpha2.MulticlusterGlobalHubSpec{},
			},
			imageOverrideCM: nil,
			wantImageManifests: map[string]string{
				"multicluster_global_hub_agent":   "quay.io/testing/multicluster-global-hub-agent:v0.6.0",
				"multicluster_global_hub_manager": "quay.io/testing/multicluster-global-hub-manager:v0.6.0",
				"oauth_proxy":                     "quay.io/testing/origin-oauth-proxy:4.9",
			},
			wantErr: nil,
		},
		{
			desc: "override image from configmap",
			initImageManifests: map[string]string{
				"multicluster_global_hub_agent":   "quay.io/stolostron/multicluster-global-hub-agent:latest",
				"multicluster_global_hub_manager": "quay.io/stolostron/multicluster-global-hub-manager:latest",
				"oauth_proxy":                     "quay.io/stolostron/multicluster-global-hub-operator:latest",
			},
			operandImagesEnv: map[string]string{
				"OPERAND_IMAGE_MULTICLUSTER_GLOBAL_HUB_MANAGER": "quay.io/stolostron/multicluster-global-hub-manager:v0.6.0",
				"OPERAND_IMAGE_MULTICLUSTER_GLOBAL_HUB_AGENT":   "quay.io/stolostron/multicluster-global-hub-agent:v0.6.0",
				"OPERAND_IMAGE_OAUTH_PROXY":                     "quay.io/stolostron/origin-oauth-proxy:4.9",
			},
			mghInstance: &operatorv1alpha2.MulticlusterGlobalHub{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: constants.GHDefaultNamespace,
					Name:      mghInstanceName,
					Annotations: map[string]string{
						operatorconstants.AnnotationImageOverridesCM: "mgh-images-config",
					},
				},
				Spec: operatorv1alpha2.MulticlusterGlobalHubSpec{},
			},
			imageOverrideCM: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: constants.GHDefaultNamespace,
					Name:      "mgh-images-config",
				},
				Data: map[string]string{
					"manifest.json": `[
  {
	"image-name": "mgh-manager",
	"image-remote": "quay.io/foo",
	"image-digest": "sha256:abcd",
	"image-key": "multicluster_global_hub_manager"
  },
  {
	"image-name": "mgh-agent",
	"image-remote": "quay.io/bar",
	"image-tag": "dev",
	"image-key": "multicluster_global_hub_agent"
  },
  {
	"image-name": "mgh-operator",
	"image-remote": "quay.io/goo",
	"image-digest": "sha256:efgh",
	"image-tag": "test",
	"image-key": "oauth_proxy"
  }
]`,
				},
			},
			wantImageManifests: map[string]string{
				"multicluster_global_hub_agent":   "quay.io/bar/mgh-agent:dev",
				"multicluster_global_hub_manager": "quay.io/foo/mgh-manager@sha256:abcd",
				"oauth_proxy":                     "quay.io/goo/mgh-operator@sha256:efgh",
			},
			wantErr: nil,
		},
		{
			desc: "override image repo image manifests from configmap",
			initImageManifests: map[string]string{
				"multicluster_global_hub_agent":   "quay.io/stolostron/multicluster-global-hub-agent:latest",
				"multicluster_global_hub_manager": "quay.io/stolostron/multicluster-global-hub-manager:latest",
				"oauth_proxy":                     "quay.io/stolostron/multicluster-global-hub-operator:latest",
			},
			operandImagesEnv: map[string]string{
				"OPERAND_IMAGE_MULTICLUSTER_GLOBAL_HUB_MANAGER": "quay.io/stolostron/multicluster-global-hub-manager:v0.6.0",
				"OPERAND_IMAGE_MULTICLUSTER_GLOBAL_HUB_AGENT":   "quay.io/stolostron/multicluster-global-hub-agent:v0.6.0",
				"OPERAND_IMAGE_OAUTH_PROXY":                     "quay.io/stolostron/origin-oauth-proxy:4.9",
			},
			mghInstance: &operatorv1alpha2.MulticlusterGlobalHub{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: constants.GHDefaultNamespace,
					Name:      mghInstanceName,
					Annotations: map[string]string{
						operatorconstants.AnnotationImageOverridesCM: "mgh-images-config",
						operatorconstants.AnnotationImageRepo:        "quay.io/testing",
					},
				},
				Spec: operatorv1alpha2.MulticlusterGlobalHubSpec{},
			},
			imageOverrideCM: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: constants.GHDefaultNamespace,
					Name:      "mgh-images-config",
				},
				Data: map[string]string{
					"manifest.json": `[
  {
	"image-name": "mgh-manager",
	"image-remote": "quay.io/foo",
	"image-digest": "sha256:abcd",
	"image-key": "multicluster_global_hub_manager"
  },
  {
	"image-name": "mgh-agent",
	"image-remote": "quay.io/bar",
	"image-tag": "dev",
	"image-key": "multicluster_global_hub_agent"
  },
  {
	"image-name": "mgh-operator",
	"image-remote": "quay.io/goo",
	"image-digest": "sha256:efgh",
	"image-tag": "test",
	"image-key": "oauth_proxy"
  }
]`,
				},
			},
			wantImageManifests: map[string]string{
				"multicluster_global_hub_agent":   "quay.io/bar/mgh-agent:dev",
				"multicluster_global_hub_manager": "quay.io/foo/mgh-manager@sha256:abcd",
				"oauth_proxy":                     "quay.io/goo/mgh-operator@sha256:efgh",
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		// set env for operand images
		for k, v := range tt.operandImagesEnv {
			t.Setenv(k, v)
		}
		// set init imageManifests
		imageOverrides = tt.initImageManifests
		t.Run(tt.desc, func(t *testing.T) {
			err := SetImageOverrides(tt.mghInstance)
			if ((err != nil && tt.wantErr == nil) || (err == nil && tt.wantErr != nil)) ||
				!reflect.DeepEqual(imageOverrides, tt.wantImageManifests) {
				t.Errorf("%s:\nwanted imageManifests & err:\n%+v\n%v\ngot imageManifests & err \n%+v\n%v",
					tt.desc, tt.wantImageManifests, tt.wantErr, imageOverrides, err)
			}
		})
	}
}

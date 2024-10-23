package config

import (
	"context"

	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/stolostron/multicluster-global-hub/operator/api/operator/v1alpha4"
)

type OperatorConfig struct {
	MetricsAddress        string
	ProbeAddress          string
	PodNamespace          string
	LeaderElection        bool
	GlobalResourceEnabled bool
	EnablePprof           bool
	LogLevel              string
}

type ControllerOption struct {
	ControllerName        string
	KubeClient            kubernetes.Interface
	OperatorConfig        *OperatorConfig
	IsGlobalhubReady      bool
	Ctx                   context.Context
	Manager               manager.Manager
	MulticlusterGlobalHub *v1alpha4.MulticlusterGlobalHub
}

type ComponentStatus struct {
	Ready  bool
	Kind   string
	Reason string
	Msg    string
}

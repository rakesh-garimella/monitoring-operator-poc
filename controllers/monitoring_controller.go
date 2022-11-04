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

package controllers

// TODO: dynamically create RBACs! Remove line below.
//+kubebuilder:rbac:groups="*",resources="*",verbs="*"

import (
	_ "context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/kyma-project/module-manager/operator/pkg/declarative"
	"github.com/kyma-project/module-manager/operator/pkg/types"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	_ "sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1alpha1 "github.tools.sap/huskies/monitoring-operator-pocs/operator/api/v1alpha1"
)

// MonitoringReconciler reconciles a Monitoring object
type MonitoringReconciler struct {
	declarative.ManifestReconciler
	client.Client
	Scheme *runtime.Scheme
	*rest.Config
	ChartPath string
}

const (
	chartNs      = "kyma-system"
	nameOverride = "custom-name-override"
	chartPath    = "./module-chart"
)

type ManifestResolver struct {
	chartPath string
}

func (m ManifestResolver) Get(obj types.BaseCustomObject, logger logr.Logger) (types.InstallationSpec, error) {
	sample, valid := obj.(*operatorv1alpha1.Monitoring)
	if !valid {
		return types.InstallationSpec{},
			fmt.Errorf("invalid type conversion for %s", client.ObjectKeyFromObject(obj))
	}
	return types.InstallationSpec{
		ChartPath:   m.chartPath,
		ReleaseName: sample.Spec.Foo,
		ChartFlags: types.ChartFlags{
			ConfigFlags: types.Flags{
				"Namespace":       chartNs,
				"CreateNamespace": true,
			},
			SetFlags: types.Flags{
				"mode":         "deployment",
				"nameOverride": nameOverride,
			},
		},
	}, nil
}

func (r *MonitoringReconciler) initReconciler(mgr ctrl.Manager) error {
	manifestResolver := &ManifestResolver{
		chartPath: chartPath,
	}
	return r.Inject(mgr, &operatorv1alpha1.Monitoring{},
		declarative.WithManifestResolver(manifestResolver),
		declarative.WithCustomResourceLabels(map[string]string{"sampleKey": "sampleValue"}),
		declarative.WithResourcesReady(true),
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *MonitoringReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Config = mgr.GetConfig()
	if err := r.initReconciler(mgr); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.Monitoring{}).
		Complete(r)
}

/*


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

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	airshipv1 "vino/api/v1"
)

// VinoReconciler reconciles a Vino object
type VinoReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=airship.airshipit.org,resources=vinoes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=airship.airshipit.org,resources=vinoes/status,verbs=get;update;patch

func (r *VinoReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("vino", req.NamespacedName)

	// your logic here

	vino := &airshipv1.Vino{}
	if err := r.Get(ctx, req.NamespacedName, vino); err != nil {
		logger.Error(err, "unable to fetch VINO object")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, err
	}


	return ctrl.Result{}, nil
}

func (r *VinoReconciler) ensureConfigMap(ctx context.Context, name types.NamespacedName, vino *airshipv1.Vino) error {
	generatedCm, err := r.buildConfigMap(ctx, name, vino)
	if err != nil {
		return err
	}

	currentCm, err := r.getCurrentConfigMap(name, vino)
	if err != nil {
		return err
	}

	if needsUpdate(generatedCm, currentCm) {
		err := r.Client.Update(ctx, generatedCm)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *VinoReconciler) buildConfigMap(ctx context.Context, name types.NamespacedName, vino *airshipv1.Vino) (*corev1.ConfigMap, error) {
	r.Log.Info("Generating new config map for vino object")

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: name.Name,
			Namespace: name.Namespace,
		},
		Data: make(map[string]string),
	}, nil
}

func (r *VinoReconciler) getCurrentConfigMap(name types.NamespacedName, vino *airshipv1.Vino) (*corev1.ConfigMap, error) {
	r.Log.Info("Getting current config map for vino object")

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: name.Name,
			Namespace: name.Namespace,
		},
		Data: make(map[string]string),
	}, nil
}

func needsUpdate(generated, current *corev1.ConfigMap) bool {
	for key, value := range generated.Data {
		if current.Data[key] != value {
			return false
		}
	}
	return true

}

func (r *VinoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&airshipv1.Vino{}).
		Complete(r)
}

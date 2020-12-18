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
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	airshipv1 "vino/api/v1"
)

const (
	DaemonSetTempalteDataKey = "template"

	ContainerNameLibvirt = "libvirt"
	ContainerNameSushy = "sushy"
	ContainerNameVinoBuilder = "builder"
	ContainerNameNodeAnnotator = "annotator"
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

	err := r.ensureConfigMap(ctx, req.NamespacedName, vino)
	if err != nil {
		vino.Status.Conditions = append([]airshipv1.Condition{
			{
				Status:  corev1.ConditionFalse,
				Reason:  "Error has occured while making sure that ConfigMap for VINO is in correct state",
				Message: err.Error(),
				Type:    airshipv1.ConditionTypeReady,
			},
		}, vino.Status.Conditions...)
		vino.Status.ConfigMapReady = false
	} else {
		vino.Status.ConfigMapReady = true
	}

	err = r.ensureDaemonSet(ctx, req.NamespacedName, vino)
	if err != nil {
		vino.Status.Conditions = append([]airshipv1.Condition{
			{
				Status:  corev1.ConditionFalse,
				Reason:  "Error has occured while making sure that VINO Daemonset is installed on kubernetes nodes",
				Message: err.Error(),
				Type:    airshipv1.ConditionTypeReady,
			},
		}, vino.Status.Conditions...)
		vino.Status.DaemonSetReady = false
	} else {
		vino.Status.DaemonSetReady = true
	}

	if err != nil {
		vino.Status.Conditions = append([]airshipv1.Condition{
			{
				Status:  corev1.ConditionFalse,
				Reason:  "Error has occured while checking if VINO networking stack is enforced on kubernetes nodes",
				Message: err.Error(),
				Type:    airshipv1.ConditionTypeReady,
			},
		}, vino.Status.Conditions...)
		vino.Status.NetworkingReady = false
	} else {
		vino.Status.NetworkingReady = true
	}

	if err != nil {
		vino.Status.Conditions = append([]airshipv1.Condition{
			{
				Status:  corev1.ConditionFalse,
				Reason:  "Error has occured while checking if virtual machines to be ready",
				Message: err.Error(),
				Type:    airshipv1.ConditionTypeReady,
			},
		}, vino.Status.Conditions...)
		vino.Status.VirtualMachinesReady = false
	} else {
		vino.Status.VirtualMachinesReady = true
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
			Name:      name.Name,
			Namespace: name.Namespace,
		},
		Data: make(map[string]string),
	}, nil
}

func (r *VinoReconciler) getCurrentConfigMap(name types.NamespacedName, vino *airshipv1.Vino) (*corev1.ConfigMap, error) {
	r.Log.Info("Getting current config map for vino object")
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
		Data: make(map[string]string),
	}, nil
}

func (r *VinoReconciler) checkNodeNetworking(name types.NamespacedName, vino *airshipv1.Vino) error {
	r.Log.Info("Checking networking stack configured on kubernetes nodes")
	return nil
}

func (r *VinoReconciler) checkVMs(name types.NamespacedName, vino *airshipv1.Vino) error {
	r.Log.Info("Checking virtual machines are configured on kubernetes nodes")
	return nil
}

func (r *VinoReconciler) setReadyStatus(vino *airshipv1.Vino) {
	if vino.Status.ConfigMapReady && vino.Status.DaemonSetReady &&
		vino.Status.NetworkingReady && vino.Status.VirtualMachinesReady {
		r.Log.Info("All VINO components are in ready state, setting VINO CR to ready state")
		vino.Status.Conditions = append([]airshipv1.Condition{
			{
				Status:  corev1.ConditionTrue,
				Reason:  "Networking, Virtual Machines, DaemonSet and ConfigMap is in ready state",
				Message: "All VINO components are in ready state, setting VINO CR to ready state",
				Type:    airshipv1.ConditionTypeReady,
			},
		}, vino.Status.Conditions...)
	}
}

func needsUpdate(generated, current *corev1.ConfigMap) bool {
	for key, value := range generated.Data {
		if current.Data[key] != value {
			return false
		}
	}
	return true
}

func (r *VinoReconciler) ensureDaemonSet(ctx context.Context, name types.NamespacedName, vino *airshipv1.Vino) error {
	ds, err := r.overrideDaemonSet(ctx, name, vino)
	if err != nil {
		return err
	}

	if ds == nil {
		ds = defaultDaemonSet(vino)
	}

	if err := applyRuntimeObject(ctx, name, ds, r.Client); err != nil {
		return err
	}
	return nil
}

func (r *VinoReconciler) setValues(ds *appsv1.DaemonSet, vino *airshipv1.Vino) error {
return nil}

func (r *VinoReconciler) overrideDaemonSet(ctx context.Context, name types.NamespacedName, vino *airshipv1.Vino) (*appsv1.DaemonSet, error) {
	dsTemplate := vino.Spec.DaemonSetOptions.Template
	logger := r.Log.WithValues("DaemonSetTemplate", dsTemplate)
	cm := &corev1.ConfigMap{}

	if dsTemplate.Name == "" || dsTemplate.Namespace == "" {
		logger.Info("user provided vino DaemonSet template is empty or missing name or namespace")
		return nil, fmt.Errorf("user provided vino DaemonSet template is empty or missing name or namespace")
	}

	err := r.Get(ctx, types.NamespacedName{
		Name: dsTemplate.Name,
		Namespace: dsTemplate.Namespace,
	}, cm)
	if err != nil {
		// TODO check error if it doesn't exist, we should requeue request and wait for the tempalte instead
		logger.Info("failed to get DaemonSet template does not exist in cluster", "error", err.Error())
		return nil, err
	}

	template, exist := cm.BinaryData[DaemonSetTempalteDataKey]
	if !exist {
		logger.Info("Malformed template provided tempalte map doesn't have key DaemonSetTempalteDataKey", "error", err.Error())
		return nil, err
	}

	ds := &appsv1.DaemonSet{}
	err = yaml.Unmarshal(template, ds)
	if err != nil {
		logger.Info("failed to unmarshal daemonset template", "error", err.Error())
		return nil, err
	}

	return ds, nil
}

func scheduleDS(ds *appsv1.DaemonSet, vino *airshipv1.Vino) {
	ds.Spec.Template.Spec.NodeSelector = vino.Spec.NodeSelector.MatchLabels


}

func (r *VinoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&airshipv1.Vino{}).
		Complete(r)
}

func defaultDaemonSet(vino *airshipv1.Vino) (ds *appsv1.DaemonSet) {
	libvirtImage := "docker.io/openstackhelm/libvirt:ubuntu_xenial-20190903"

	if vino.Spec.DaemonSetOptions.LibvirtImage != "" {
		libvirtImage = vino.Spec.DaemonSetOptions.LibvirtImage
	}

	ds = &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
				Spec: corev1.PodSpec{
					HostPID: true,
					HostNetwork: true,
					HostIPC: true,
					Volumes: []corev1.Volume{
						{
							Name: "libvirt-bin",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "libvirt-bin",
									},
									DefaultMode: &[]int32{0}[0555],
								},
							},
						},
						{
							Name: "libvirt-etc",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "libvirt-etc",
									},
									DefaultMode: &[]int32{0}[0444],
								},
							},
						},
						{
							Name: "cgroup",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/sys/fs/cgroup",
								},
							},
						},
						{
							Name: "dev",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/dev",
								},
							},
						},
						{
							Name: "run",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/run",
								},
							},
						},
						{
							Name: "var-lib-libvirt",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/libvirt",
								},
							},
						},
						{
							Name: "var-lib-libvirt-images",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "var-lib-libvirt-images",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name: ContainerNameLibvirt,
							SecurityContext: &corev1.SecurityContext{
								Privileged: &[]bool{true}[0],
								RunAsUser: &[]int64{0}[0],
							},
							Image: libvirtImage,
							Command: []string{"/tmp/libvirt.sh"},
							VolumeMounts: []corev1.VolumeMount{
								{
									MountPath: "/tmp/libvirt.sh",
									Name: "libvirt-bin",
									SubPath: "libvirt.sh",
									ReadOnly: true,
								},
								{},
							},
						},
					},
				},
			},
		},
	}
	return
}

func applyRuntimeObject(ctx context.Context, key client.ObjectKey, obj runtime.Object, c client.Client) error {
	getObj := obj.DeepCopyObject()
	switch err := c.Get(ctx, key, getObj); {
	case apierror.IsNotFound(err):
		return c.Create(ctx, obj)
	case err == nil:
		return c.Update(ctx, obj)
	default:
		return err
	}
}

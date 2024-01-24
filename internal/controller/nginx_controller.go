/*
Copyright 2024.

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

package controller

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	examplesv1 "github.com/cdivitotawela/nginx-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// NginxReconciler reconciles a Nginx object
type NginxReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=examples.example.dom,resources=nginxes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=examples.example.dom,resources=nginxes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=examples.example.dom,resources=nginxes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Nginx object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *NginxReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	l.Info("Starting reconcile", "name", req.Name, "namespace", req.Namespace)

	var nginx examplesv1.Nginx
	if err := r.Get(ctx, req.NamespacedName, &nginx); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &nginx.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deployment": req.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"deployment": req.Name}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "nginx", Image: "nginx:" + nginx.Spec.Tag}},
				},
			},
		},
	}

	err := controllerutil.SetControllerReference(&nginx, deploy, r.Scheme)
	if err != nil {
		return reconcile.Result{}, err
	}

	found := &appsv1.Deployment{}
	err = r.Get(context.TODO(), types.NamespacedName{
		Name:      deploy.Name,
		Namespace: deploy.Namespace,
	}, found)

	if err != nil && errors.IsNotFound(err) {
		err = r.Create(context.TODO(), deploy)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NginxReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&examplesv1.Nginx{}).
		Complete(r)
}

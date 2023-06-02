/*
Copyright 2023.

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
	"fmt"
	"strings"

	// 3rd party and SIG contexts
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	mbuilesv1alpha1 "github.com/manuelbuil/operator-testing/api/v1alpha1"
)

// Egressgwk3sReconciler reconciles a Egressgwk3s object
type Egressgwk3sReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=mbuil.es,resources=egressgwk3s,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mbuil.es,resources=egressgwk3s/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mbuil.es,resources=egressgwk3s/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Egressgwk3s object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *Egressgwk3sReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger := r.Log.WithValues("namespace", req.NamespacedName, "egressGw", req.Name)
	logger.Info("== Reconciling EgressGW")

	// Fetch the At instance
	instance := &mbuilesv1alpha1.Egressgwk3s{}
	err := r.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request - return and don't requeue:
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request:
		return reconcile.Result{}, err
	}

	// If no phase set, default to pending (the initial phase):
	if instance.Status.Phase == "" {
		instance.Status.Phase = mbuilesv1alpha1.PhasePending
	}

	// Make the main case distinction: implementing
	// the state diagram PENDING -> RUNNING -> DONE
	switch instance.Status.Phase {
	case mbuilesv1alpha1.PhasePending:
		logger.Info("Phase: PENDING")

		instance.Status.Phase = mbuilesv1alpha1.PhaseRunning

	case mbuilesv1alpha1.PhaseRunning:
		logger.Info("Phase: RUNNING")
		logger.Info("About to execute: echo " + instance.Spec.GwNode + " " + instance.Spec.SourceIP)
		pod := newPodForCR(instance)
		// Set At instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, pod, r.Scheme); err != nil {
			// requeue with error
			return reconcile.Result{}, err
		}
		found := &corev1.Pod{}
		err = r.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
		// Try to see if the Pod already exists and if not
		// (which we expect) then create a one-shot Pod as per spec:
		if err != nil && errors.IsNotFound(err) {
			err = r.Create(context.TODO(), pod)
			if err != nil {
				// requeue with error
				return reconcile.Result{}, err
			}
			logger.Info("Pod launched", "name", pod.Name)
		} else if err != nil {
			// requeue with error
			return reconcile.Result{}, err
		} else if found.Status.Phase == corev1.PodFailed || found.Status.Phase == corev1.PodSucceeded {
			logger.Info("Container terminated", "reason", found.Status.Reason, "message", found.Status.Message)
			instance.Status.Phase = mbuilesv1alpha1.PhaseDone
		} else {
			// don't requeue because it will happen automatically when the Pod status changes
			return reconcile.Result{}, nil
		}

	case mbuilesv1alpha1.PhaseDone:
		logger.Info("Phase: DONE")
		return reconcile.Result{}, nil

	default:
		logger.Info("NOP")
		return reconcile.Result{}, nil
	}

	myNodeList := &corev1.NodeList{}
	err = r.List(context.TODO(), myNodeList)

	if err != nil {
		logger.Info("Error. Shit")
		return reconcile.Result{}, err
	}

	logger.Info("NO ERROR! HURRAY!!")
	fmt.Printf("These are the nodes: %#v \n", myNodeList)

	// Update the At instance, setting the status to the respective phase:
	err = r.Status().Update(context.TODO(), instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Egressgwk3sReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mbuilesv1alpha1.Egressgwk3s{}).
		Owns(&mbuilesv1alpha1.Egressgwk3s{}).
		Owns(&corev1.Pod{}).
		Owns(&corev1.Node{}).
		Complete(r)
}

// newPodForCR returns a busybox Pod with same name/namespace declared in resource
func newPodForCR(cr *mbuilesv1alpha1.Egressgwk3s) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:    "busybox",
				Image:   "busybox",
				Command: strings.Split("echo ", cr.Spec.GwNode+" "+cr.Spec.SourceIP),
			}},
			RestartPolicy: corev1.RestartPolicyOnFailure,
		},
	}
}

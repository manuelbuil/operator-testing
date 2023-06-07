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

	// 3rd party and SIG contexts

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	mbuilesv1alpha1 "github.com/manuelbuil/operator-testing/api/v1alpha1"
	"github.com/sirupsen/logrus"
)

// Egressgwk3sReconciler reconciles a Egressgwk3s object
type Egressgwk3sReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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

	logrus.Info("== Reconciling EgressGW")

	logrus.Infof("This is req.Name: %v", req.Name)
	logrus.Infof("This is req.Namespace: %v", req.Namespace)
	logrus.Infof("This is req.String: %v", req.String())
	logrus.Infof("This is req.NamespacedName: %v", req.NamespacedName)

	isAPod := false
	isACRD := false

	podInstance := &corev1.Pod{}
	errPod := r.Get(ctx, req.NamespacedName, podInstance)
	if errPod != nil {
		if errors.IsNotFound(errPod) {
			logrus.Infof("This is not a pod!")
		} else {
			return reconcile.Result{}, errPod
		}
	} else {
		isAPod = true
		logrus.Info("It is a CRD")
	}

	// Fetch the egressgwk3s instance
	instance := &mbuilesv1alpha1.Egressgwk3s{}
	err := r.Get(ctx, req.NamespacedName, instance)
	logrus.Infof("This is err: %v", err)
	if err != nil {
		if errors.IsNotFound(err) {
			logrus.Info("This is not a crd!")
		} else {
			// Error reading the object - requeue the request:
			return reconcile.Result{}, err
		}
	} else {
		isACRD = true
		logrus.Info("It is a pod")
	}

	if !isAPod && !isACRD {
		return reconcile.Result{}, fmt.Errorf("This is not a pod or a CRD")
	}

	if isACRD {
		var podIPs []string
		logrus.Infof("This is sourcepods: %v", instance.Spec.SourcePods)
		for _, sourcePod := range instance.Spec.SourcePods {
			// Fetch pods based on the namespace selector
			namespaceSelector := sourcePod.NamespaceSelector
			namespaceList := &corev1.NamespaceList{}
			if len(namespaceSelector.MatchLabels) > 0 {
				err = r.Client.List(ctx, namespaceList, client.MatchingLabels(namespaceSelector.MatchLabels))
				if err != nil {
					return ctrl.Result{}, err
				}
			}
			// Fetch pods based on the pod selector
			podSelector := sourcePod.PodSelector
			podList := &corev1.PodList{}
			if len(podSelector.MatchLabels) > 0 {
				err = r.Client.List(ctx, podList, client.InNamespace(req.Namespace), client.MatchingLabels(podSelector.MatchLabels))
				if err != nil {
					return ctrl.Result{}, err
				}
			}

			logrus.Infof("This is namespaceList: %v and this is podList: %v", namespaceList, podList)

			// Collect the IP addresses of pods that match either the namespace selector or the pod selector
			for _, pod := range podList.Items {
				logrus.Infof("One pod found on the podList: %v", pod.Status.PodIP)
				podIPs = append(podIPs, pod.Status.PodIP)
			}
			for _, namespace := range namespaceList.Items {
				podsInNamespace := &corev1.PodList{}
				err := r.Client.List(ctx, podsInNamespace, client.InNamespace(namespace.Name), client.MatchingLabels(podSelector.MatchLabels))
				if err != nil {
					return ctrl.Result{}, err
				}
				for _, pod := range podsInNamespace.Items {
					logrus.Infof("One pod found on the namespaceList: %v", pod.Status.PodIP)
					podIPs = append(podIPs, pod.Status.PodIP)
				}
			}

			// Update the custom resource status with the collected IP addresses
			instance.Status.Pods = podIPs
			err = r.Client.Status().Update(ctx, instance)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	if isAPod {
		// Grab all the existing egressGW and create a new request to reconcile for each of them
		egressGwList := &mbuilesv1alpha1.Egressgwk3sList{}
		_ = r.List(ctx, egressGwList)
		for _, egressgw := range egressGwList.Items {
			logrus.Infof("This is egressgw: %v", egressgw)
		}
	}

	logrus.Info("NO ERROR! HURRAY!!")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Egressgwk3sReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mbuilesv1alpha1.Egressgwk3s{}).
		Watches(
			&source.Kind{Type: &corev1.Pod{}},
			&handler.EnqueueRequestForObject{},
		).
		Complete(r)
}

func (r *Egressgwk3sReconciler) podSelectorMatches(configMap client.Object) bool {
	//
	return true
}

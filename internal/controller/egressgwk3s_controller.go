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
	"os"

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

type myObect struct {
	isPod  bool
	isCrd  bool
	isNode bool
	myPod  corev1.Pod
	myNode corev1.Node
	myCRD  mbuilesv1alpha1.Egressgwk3s
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

	myObject, err := analyzeObject(ctx, req, r.Client)

	if !myObject.isCrd && !myObject.isPod && !myObject.isNode {
		logrus.Info("This is not a pod or a CRD or a node. Probably something got removed")
	}

	if myObject.isCrd {
		logrus.Infof("This is sourcepods: %v", myObject.myCRD.Spec.SourcePods)
		err = processEgressGW(ctx, myObject.myCRD, req, r.Client)
		if err != nil {
			logrus.Info("Error in isACRD")
			return reconcile.Result{}, err
		}
	}

	// If it is a pod or something else (node, pod, crd) got removed, run all egresscrds again
	if myObject.isPod || (!myObject.isPod && !myObject.isCrd && !myObject.isNode) {
		// Grab all the existing egressGW and create a new request to reconcile for each of them
		egressGwList := &mbuilesv1alpha1.Egressgwk3sList{}
		_ = r.List(ctx, egressGwList)
		for _, egressgw := range egressGwList.Items {
			logrus.Infof("This is egressgw: %v", egressgw)
			err = processEgressGW(ctx, egressgw, req, r.Client)
			if err != nil {
				logrus.Info("Error in isAPod")
				return reconcile.Result{}, err
			}
		}
	}

	logrus.Info("NO ERROR! HURRAY!!")

	return ctrl.Result{}, nil
}

// amIGw checks if the pod where controller is running is in the node that is gateway
func amIGw(ctx context.Context, nodeName string, k8sClient client.Client) (bool, error) {
	podName := os.Getenv("HOSTNAME")
	logrus.Infof("MANU - This is podName: %s", podName)

	podList := &corev1.PodList{}
	listOptions := &client.ListOptions{Namespace: ""}

	err := k8sClient.List(ctx, podList, listOptions)
	if err != nil {
		return false, err
	}

	// We need to list over all pods. Using Get will not work as we don't necessarily know the namespace
	for _, pod := range podList.Items {
		if pod.Name == podName {
			logrus.Infof("Pod found. This is its node-name: %s", pod.Spec.NodeName)
			if nodeName == pod.Spec.NodeName {
				return true, nil
			}
		}
	}

	return false, nil

}

func processEgressGW(ctx context.Context, egressgw mbuilesv1alpha1.Egressgwk3s, req ctrl.Request, k8sClient client.Client) error {
	var finalPodIPs []string
	var podIPs []string
	var err error

	// loop over all sourcePod rules
	for _, sourcePod := range egressgw.Spec.SourcePods {
		podIPs, err = processSourcePod(ctx, sourcePod, req, k8sClient)
		if err != nil {
			logrus.Info("There is an error in processEgressGW")
			return err
		}
		finalPodIPs = append(finalPodIPs, podIPs...)
	}

	// If there are no pods, it does not matter that there is a gw
	node := &corev1.Node{}
	if len(finalPodIPs) != 0 {
		// If there are pods, find the node with the passed name
		nodeGw := egressgw.Spec.GwNode
		if nodeGw != "" {
			err = k8sClient.Get(ctx, client.ObjectKey{Name: nodeGw}, node)
			if err != nil {
				logrus.Error("That node does not exist!")
				return err
			}
		}
		var gw bool
		gw, err = amIGw(ctx, nodeGw, k8sClient)
		if err != nil {
			logrus.Info("Error in amIGw")
			return err
		}
		if gw {
			logrus.Infof("OMG!!! I AM THE GATEWAY!!!!!")
		} else {
			logrus.Info("BUMMER! I AM NOT SPECIAL!")
		}
	}

	// Update the custom resource status with the collected IP addresses
	egressgw.Status.Pods = finalPodIPs
	egressgw.Status.NodeIP = node.Status.Addresses
	err = k8sClient.Status().Update(ctx, &egressgw)
	if err != nil {
		return err
	}

	return nil
}

// processSourcePod processes the labels that can be passed to match pods
func processSourcePod(ctx context.Context, sourcePod mbuilesv1alpha1.SourcePodsSelector, req ctrl.Request, k8sClient client.Client) ([]string, error) {
	var podIPs []string

	// Fetch pods based on the namespace selector
	namespaceSelector := sourcePod.NamespaceSelector
	namespaceList := &corev1.NamespaceList{}
	if len(namespaceSelector.MatchLabels) > 0 {
		err := k8sClient.List(ctx, namespaceList, client.MatchingLabels(namespaceSelector.MatchLabels))
		if err != nil {
			return podIPs, err
		}
	}

	// Fetch pods based on the pod selector
	podSelector := sourcePod.PodSelector
	podList := &corev1.PodList{}
	if len(podSelector.MatchLabels) > 0 {
		err := k8sClient.List(ctx, podList, client.InNamespace(req.Namespace), client.MatchingLabels(podSelector.MatchLabels))
		if err != nil {
			return podIPs, err
		}
	}

	// Collect the IP addresses of pods that match either the namespace selector or the pod selector
	for _, pod := range podList.Items {
		logrus.Infof("One pod found on the podList: %v", pod.Status.PodIP)
		podIPs = append(podIPs, pod.Status.PodIP)
	}
	for _, namespace := range namespaceList.Items {
		podsInNamespace := &corev1.PodList{}
		err := k8sClient.List(ctx, podsInNamespace, client.InNamespace(namespace.Name), client.MatchingLabels(podSelector.MatchLabels))
		if err != nil {
			return podIPs, err
		}
		for _, pod := range podsInNamespace.Items {
			logrus.Infof("One pod found on the namespaceList: %v", pod.Status.PodIP)
			podIPs = append(podIPs, pod.Status.PodIP)
		}
	}

	return podIPs, nil
}

// analyzeObject returns a MyOject struct
func analyzeObject(ctx context.Context, req ctrl.Request, k8sClient client.Client) (myObect, error) {

	var object myObect

	//TODO investigate r.event
	// Check if the update came from a pod

	podInstance := &corev1.Pod{}
	errPod := k8sClient.Get(ctx, req.NamespacedName, podInstance)
	if errPod != nil {
		if !errors.IsNotFound(errPod) {
			return myObect{}, errPod
		}
	} else {
		object.isPod = true
		object.myPod = *podInstance
	}

	//TODO investigate r.event
	// Check if the update came from a node
	nodeInstance := &corev1.Node{}
	errNode := k8sClient.Get(ctx, req.NamespacedName, nodeInstance)
	if errNode != nil {
		if !errors.IsNotFound(errNode) {
			return myObect{}, errNode
		}
	} else {
		object.isNode = true
		object.myNode = *nodeInstance
	}

	// Fetch the egressgwk3s instance
	instance := &mbuilesv1alpha1.Egressgwk3s{}
	err := k8sClient.Get(ctx, req.NamespacedName, instance)
	logrus.Infof("This is err: %v", err)
	if err != nil {
		if !errors.IsNotFound(err) {
			return myObect{}, err
		}
	} else {
		object.isCrd = true
		object.myCRD = *instance
	}

	return object, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Egressgwk3sReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mbuilesv1alpha1.Egressgwk3s{}).
		Watches(
			&source.Kind{Type: &corev1.Pod{}},
			&handler.EnqueueRequestForObject{},
		).
		Watches(
			&source.Kind{Type: &corev1.Node{}},
			&handler.EnqueueRequestForObject{},
		).
		Complete(r)
}

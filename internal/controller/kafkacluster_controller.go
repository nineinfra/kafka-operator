/*
Copyright 2024 nineinfra.

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
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kafkav1 "github.com/nineinfra/kafka-operator/api/v1"
)

// KafkaClusterReconciler reconciles a KafkaCluster object
type KafkaClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type reconcileFun func(ctx context.Context, cluster *kafkav1.KafkaCluster, logger logr.Logger) error

//+kubebuilder:rbac:groups=kafka.nineinfra.tech,resources=kafkaclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kafka.nineinfra.tech,resources=kafkaclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kafka.nineinfra.tech,resources=kafkaclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KafkaCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *KafkaClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var cluster kafkav1.KafkaCluster
	err := r.Get(ctx, req.NamespacedName, &cluster)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Object not found, it could have been deleted")
		} else {
			logger.Error(err, "Error occurred during fetching the object")
		}
		return ctrl.Result{}, err
	}
	requestArray := strings.Split(fmt.Sprint(req), "/")
	requestName := requestArray[1]
	logger.Info(fmt.Sprintf("Reconcile requestName %s,cluster.Name %s", requestName, cluster.Name))
	if requestName == cluster.Name {
		logger.Info("Create or update clusters")
		err = r.reconcileClusters(ctx, &cluster, logger)
		if err != nil {
			logger.Error(err, "Error occurred during create or update clusters")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *KafkaClusterReconciler) reconcileClusterStatus(ctx context.Context, cluster *kafkav1.KafkaCluster, logger logr.Logger) (err error) {
	cluster.Status.Init()
	existsPods := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(ClusterResourceLabels(cluster))
	listOps := &client.ListOptions{
		Namespace:     cluster.Namespace,
		LabelSelector: labelSelector,
	}
	err = r.Client.List(context.TODO(), existsPods, listOps)
	if err != nil {
		return err
	}
	var (
		readyMembers   []string
		unreadyMembers []string
	)
	for _, p := range existsPods.Items {
		ready := true
		for _, c := range p.Status.ContainerStatuses {
			if !c.Ready {
				ready = false
			}
		}
		if ready {
			readyMembers = append(readyMembers, p.Name)
		} else {
			unreadyMembers = append(unreadyMembers, p.Name)
		}
	}
	cluster.Status.Members.Ready = readyMembers
	cluster.Status.Members.Unready = unreadyMembers

	logger.Info("Updating cluster status")
	if cluster.Status.ReadyReplicas == cluster.Spec.Resource.Replicas {
		cluster.Status.SetPodsReadyConditionTrue()
	} else {
		cluster.Status.SetPodsReadyConditionFalse()
	}
	if cluster.Status.CurrentVersion == "" && cluster.Status.IsClusterInReadyState() {
		cluster.Status.CurrentVersion = cluster.Spec.Image.Tag
	}
	return r.Client.Status().Update(context.TODO(), cluster)
}

func (r *KafkaClusterReconciler) reconcileHeadlessService(ctx context.Context, cluster *kafkav1.KafkaCluster, logger logr.Logger) (err error) {
	desiredSvc, err := r.constructHeadlessService(cluster)
	if err != nil {
		return err
	}

	existsSvc := &corev1.Service{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desiredSvc.Name, Namespace: desiredSvc.Namespace}, existsSvc)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new headless service")
		err = r.Client.Create(context.TODO(), desiredSvc)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		logger.Info("Updating existing headless service")
		existsSvc.Spec.Ports = desiredSvc.Spec.Ports
		existsSvc.Spec.Type = desiredSvc.Spec.Type
		err = r.Client.Update(context.TODO(), existsSvc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *KafkaClusterReconciler) reconcileService(ctx context.Context, cluster *kafkav1.KafkaCluster, logger logr.Logger) (err error) {
	desiredSvc, err := r.constructService(cluster)
	if err != nil {
		return err
	}

	existsSvc := &corev1.Service{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desiredSvc.Name, Namespace: desiredSvc.Namespace}, existsSvc)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new service")
		err = r.Client.Create(context.TODO(), desiredSvc)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		logger.Info("Updating existing service")
		existsSvc.Spec.Ports = desiredSvc.Spec.Ports
		existsSvc.Spec.Type = desiredSvc.Spec.Type
		err = r.Client.Update(context.TODO(), existsSvc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *KafkaClusterReconciler) reconcileConfigMap(ctx context.Context, cluster *kafkav1.KafkaCluster, logger logr.Logger) (err error) {
	desiredCm, err := r.constructConfigMap(cluster)
	if err != nil {
		return err
	}
	existsCm := &corev1.ConfigMap{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desiredCm.Name, Namespace: desiredCm.Namespace}, existsCm)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new Kafka ConfigMap")
		err = r.Client.Create(context.TODO(), desiredCm)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		logger.Info("Updating existing ConfigMap")
		existsCm.Data = desiredCm.Data
		err = r.Client.Update(context.TODO(), existsCm)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *KafkaClusterReconciler) reconcileWorkload(ctx context.Context, cluster *kafkav1.KafkaCluster, logger logr.Logger) (err error) {
	desiredSts, err := r.constructKafkaWorkload(cluster)
	if err != nil {
		return err
	}
	existsSts := &appsv1.StatefulSet{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desiredSts.Name, Namespace: desiredSts.Namespace}, existsSts)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new Kafka StatefulSet")
		err = r.Client.Create(context.TODO(), desiredSts)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		//Todo
	}
	logger.Info("Creating a new KafkaCluster successfully")
	return nil
}

func (r *KafkaClusterReconciler) reconcileClusters(ctx context.Context, cluster *kafkav1.KafkaCluster, logger logr.Logger) error {
	for _, fun := range []reconcileFun{
		r.reconcileConfigMap,
		r.reconcileWorkload,
		r.reconcileService,
		r.reconcileHeadlessService,
		r.reconcileClusterStatus,
	} {
		if err := fun(ctx, cluster, logger); err != nil {
			return err
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kafkav1.KafkaCluster{}).
		Complete(r)
}

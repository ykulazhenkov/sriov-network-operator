package controllers

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sriovnetworkv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
)

// SriovNetworkNodeConfigReconciler reconciles a SriovNetworkNodeConfig object
type SriovNetworkNodeConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=sriovnetwork.openshift.io,resources=sriovnetworknodeconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=sriovnetwork.openshift.io,resources=sriovnetworknodeconfigs/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SriovNetworkNodeConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.6.4/pkg/reconcile
func (r *SriovNetworkNodeConfigReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("sriovnetworknodeconfig", req.NamespacedName)
	// Fetch SriovNetworkNodeConfigList
	configList := &sriovnetworkv1.SriovNetworkNodeConfigList{}
	err := r.List(ctx, configList, &client.ListOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	nodes := &corev1.NodeList{}
	err = r.List(ctx, nodes, &client.ListOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	for _, node := range nodes.Items {
		found := &sriovnetworkv1.SriovNetworkNodeState{}
		err = r.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: node.Name}, found)
		if err != nil {
			logger.Info("Fail to get SriovNetworkNodeState", "namespace", namespace, "name", node.Name)
			if errors.IsNotFound(err) {
				ns := &sriovnetworkv1.SriovNetworkNodeState{
					ObjectMeta: metav1.ObjectMeta{
						Name:      node.Name,
						Namespace: namespace}}
				err = r.Create(context.TODO(), ns)
				if err != nil {
					return reconcile.Result{}, fmt.Errorf("Couldn't create SriovNetworkNodeState: %v", err)
				}
				logger.Info("Created SriovNetworkNodeState for", namespace, ns.Name)
			} else {
				return reconcile.Result{}, fmt.Errorf("Failed to get SriovNetworkNodeState: %v", err)
			}
		} else {
			logger.Info("SriovNetworkNodeState already exists, updating")
			newVersion := found.DeepCopy()

			var nodeSelected bool

		Loop:
			for _, c := range configList.Items {
				for _, rdmaConf := range c.Spec.RDMAExclusiveMode {
					if isNodeSelected(&node, rdmaConf.NodeSelector) {
						nodeSelected = true
						break Loop
					}
				}
			}
			if nodeSelected != newVersion.Spec.NodeSettings.RDMAExclusiveMode {
				newVersion.Spec.NodeSettings.RDMAExclusiveMode = nodeSelected
				err = r.Update(context.TODO(), newVersion)
				if err != nil {
					return reconcile.Result{}, fmt.Errorf("Couldn't update SriovNetworkNodeState: %v", err)
				}
			}
		}
	}
	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SriovNetworkNodeConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sriovnetworkv1.SriovNetworkNodeConfig{}).
		Complete(r)
}

func isNodeSelected(node *corev1.Node, selectors map[string]string) bool {
	for k, v := range selectors {
		if nv, ok := node.Labels[k]; ok && nv == v {
			continue
		}
		return false
	}
	return true
}

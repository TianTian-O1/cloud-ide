package controllers

import (
	"context"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/mangohow/cloud-ide/pkg/notifier"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mv1 "github.com/mangohow/cloud-ide/cmd/control-plane/internal/api/v1"
)

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	notifier notifier.Notifier
}

func NewPodReconciler(c client.Client, scheme *runtime.Scheme, notifier notifier.Notifier) *PodReconciler {
	return &PodReconciler{
		Client:   c,
		Scheme:   scheme,
		notifier: notifier,
	}
}

// Reconcile 检测Pod的状态，然后：
// 1.从网关注册或注销Workspace
// 2.更新Workspace的状态
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	lgr := log.FromContext(ctx)

	// 获取pod
	var pod v1.Pod
	err := r.Client.Get(ctx, req.NamespacedName, &pod)
	// 1.Pod被删除了，更新Workspace状态
	if errors.IsNotFound(err) {
		lgr.V(5).Info("pod is terminated", "name", req.Name)

		r.updateWorkspaceStatus(ctx, req.NamespacedName, mv1.WorkspacePhaseStopped)

		return ctrl.Result{}, nil
	}
	if err != nil {
		lgr.Error(err, "get pod")
		return ctrl.Result{Requeue: true}, err
	}

	// 只处理工作空间Pod（有cloud-ide标签的Pod）
	if pod.Labels["app"] != "cloud-ide" {
		lgr.V(5).Info("skipping non-workspace pod", "name", req.Name)
		return ctrl.Result{}, nil
	}

	// 2.Pod被删除了，处于Terminating中
	// 2.1 从网关中注销Workspace
	// 2.2 更新Workspace的状态
	if pod.DeletionTimestamp != nil {
		lgr.V(5).Info("pod is terminating", "name", req.Name, "phase", pod.Status.Phase)

		r.notifier.Logout(pod.Annotations["sid"])

		r.updateWorkspaceStatus(ctx, req.NamespacedName, mv1.WorkspacePhaseStopping)

		return ctrl.Result{}, nil
	}

	// 3.Pod已经启动完成
	if pod.Status.Phase == v1.PodRunning {
		lgr.V(5).Info("pod is running", "name", req.Name, "phase", pod.Status.Phase)

		// 3.1 更新Workspace状态
		r.updateWorkspaceStatus(ctx, req.NamespacedName, mv1.WorkspacePhaseRunning)

		// 3.2 将Workspace注册到网关中
		endpoint := pod.Status.PodIP + ":" + strconv.Itoa(int(pod.Spec.Containers[0].Ports[0].ContainerPort))
		sid, ok := pod.Annotations["sid"]
		if !ok {
			lgr.Error(err, "get sid from annotations")
			return ctrl.Result{Requeue: true}, err
		}
		r.notifier.Login(sid, endpoint)

		// 3.3 通知用户Workspace可用
		r.notifier.Notify(sid)

		return ctrl.Result{}, nil
	}

	lgr.V(5).Info("pod is creating", "name", req.Name, "phase", pod.Status.Phase)
	// 4.Pod正在被创建,更新ws状态
	r.updateWorkspaceStatus(ctx, req.NamespacedName, mv1.WorkspacePhaseStaring)

	return ctrl.Result{}, nil
}

// 更新workspace的状态
func (r *PodReconciler) updateWorkspaceStatus(ctx context.Context, key client.ObjectKey, phase mv1.WorkSpacePhase) {
	lgr, _ := logr.FromContext(ctx)
	var (
		ws  mv1.WorkSpace
		err error
	)
	// 1.先查询本地缓存，如果不存在说明被删除了，直接返回
	if err = r.Client.Get(ctx, key, &ws); errors.IsNotFound(err) {
		return
	}
	if err != nil {
		lgr.Error(err, "update workspace status")
		return
	}

	// 2.如果实际状态就算期望状态，返回
	if ws.Status.Phase == phase {
		return
	}

	// 3.更新状态
	ws.Status.Phase = phase
	err = r.Status().Update(ctx, &ws)
	if err != nil {
		lgr.Error(err, "update status")
	}

	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{MaxConcurrentReconciles: 8}).
		For(&v1.Pod{}).
		Complete(r)
}

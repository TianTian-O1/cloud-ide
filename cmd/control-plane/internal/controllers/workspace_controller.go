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

package controllers

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"time"

	mv1 "github.com/mangohow/cloud-ide/cmd/control-plane/internal/api/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var Mode string

const (
	ModeRelease = "release"
	ModDev      = "dev"
)

// WorkSpaceReconciler reconciles a WorkSpace object
type WorkSpaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func NewWorkSpaceReconciler(c client.Client, scheme *runtime.Scheme) *WorkSpaceReconciler {
	return &WorkSpaceReconciler{
		Client: c,
		Scheme: scheme,
	}
}

// +kubebuilder:rbac:groups=cloud-ide.mangohow.com,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud-ide.mangohow.com,resources=workspaces/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud-ide.mangohow.com,resources=workspaces/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pod,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the WorkSpace object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *WorkSpaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	lgr := log.FromContext(ctx)

	// 1.先查询WorkSpace
	ws := mv1.WorkSpace{}
	err := r.Client.Get(ctx, req.NamespacedName, &ws)
	// case 1、没有找到Workspace,说明WorkSpace被删除了,删除对应的Pod和PVC即可
	if err != nil {
		if errors.IsNotFound(err) {
			if err := r.deletePod(ctx, req.NamespacedName); err != nil {
				lgr.Error(err, "delete pod")
				return ctrl.Result{Requeue: true}, err
			}
			if err := r.deletePVC(ctx, req.NamespacedName); err != nil {
				lgr.Error(err, "delete pvc")
				return ctrl.Result{Requeue: true}, err
			}

			return ctrl.Result{}, nil
		}

		lgr.Error(err, "get workspace")
		return ctrl.Result{Requeue: true}, err
	}

	// 2.找到了WorkSpace,根据WorkSpace的Operation字段判断要进行的操作
	switch ws.Spec.Command {
	// case2: 启动WorkSpace,检查PVC是否存在,如果不存在则创建
	case mv1.WorkSpaceStart:
		// 检查PVC是否存在,不存在则创建
		err = r.createPVC(ctx, &ws, req.NamespacedName)
		if err != nil {
			lgr.Error(err, "create pvc")
			return ctrl.Result{Requeue: true}, err
		}
		// 创建Pod
		err = r.createPod(ctx, &ws, req.NamespacedName)
		if err != nil {
			lgr.Error(err, "create pod")
			return ctrl.Result{Requeue: true}, err
		}

	// case3: 停止WorkSpace,删除Pod
	case mv1.WorkSpaceStop:
		// 删除Pod
		err = r.deletePod(ctx, req.NamespacedName)
		if err != nil {
			lgr.Error(err, "delete pod")
			return ctrl.Result{Requeue: true}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WorkSpaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// 为uid设置索引
	err := mgr.GetFieldIndexer().IndexField(context.Background(), &mv1.WorkSpace{}, "uid", func(object client.Object) []string {
		ws := object.(*mv1.WorkSpace)
		return []string{ws.Spec.UID}
	})
	if err != nil {
		mgr.GetLogger().Error(err, "set index")
		panic(err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{MaxConcurrentReconciles: 8}).
		For(&mv1.WorkSpace{}).
		Owns(&v1.Pod{}, builder.WithPredicates(predicatePod)).
		Owns(&v1.PersistentVolumeClaim{}, builder.WithPredicates(predicatePVC)).
		Complete(r)
}

func (r *WorkSpaceReconciler) createPod(ctx context.Context, space *mv1.WorkSpace, key client.ObjectKey) error {
	// 1.检查Pod是否存在
	exist, err := r.checkPodExist(ctx, key)
	if err != nil {
		return err
	}

	// Pod已存在,直接返回
	if exist {
		return nil
	}

	// 2.创建Pod
	pod := r.constructPod(space)

	// 设置控制器
	if err = controllerutil.SetControllerReference(space, pod, r.Scheme); err != nil {
		return err
	}

	ctx, cancelFunc := context.WithTimeout(ctx, time.Second*120)
	defer cancelFunc()
	err = r.Client.Create(ctx, pod)
	if err != nil {
		// 如果Pod已经存在,直接返回
		if errors.IsAlreadyExists(err) {
			return nil
		}

		return err
	}

	return nil
}

// 构造一个Pod对象
func (r *WorkSpaceReconciler) constructPod(space *mv1.WorkSpace) *v1.Pod {
	volumeName := "volume-user-workspace"
	workspaceDir := filepath.Join(space.Spec.MountPath, "/workspace")
	
	// 创建基础容器配置
	container := v1.Container{
		Name:            space.Name,
		Image:           space.Spec.Image,
		ImagePullPolicy: v1.PullIfNotPresent,
		Ports: []v1.ContainerPort{
			{
				ContainerPort: space.Spec.Port,
			},
		},
		// 容器挂载存储卷
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      volumeName,
				ReadOnly:  false,
				MountPath: space.Spec.MountPath,
			},
		},
		Env: []v1.EnvVar{
			{
				Name:  "OPEN_DIR",
				Value: workspaceDir,
			},
		},
	}
	
	// 解析环境变量配置（特别是Claude模板）
	actualGitRepo := space.Spec.GitRepository
	if strings.HasPrefix(space.Spec.GitRepository, "ENV:") {
		// 解析编码的环境变量信息
		parts := strings.SplitN(space.Spec.GitRepository, "|GIT:", 2)
		if len(parts) == 2 {
			envJSON := strings.TrimPrefix(parts[0], "ENV:")
			actualGitRepo = parts[1]
			
			// 解析环境变量JSON
			var envVars map[string]string
			if err := json.Unmarshal([]byte(envJSON), &envVars); err == nil {
				// 添加环境变量到容器
				for key, value := range envVars {
					container.Env = append(container.Env, v1.EnvVar{
						Name:  key,
						Value: value,
					})
				}
			}
		}
	}
	
	// 为codercom/code-server镜像添加正确的启动命令
	if strings.Contains(space.Spec.Image, "codercom/code-server") || strings.Contains(space.Spec.Image, "claude-code-server") {
		container.Command = []string{"/usr/bin/code-server"}
		container.Args = []string{
			"--bind-addr", "0.0.0.0:9999",
			"--auth", "none",
			"--disable-update-check",
			"--disable-telemetry",
			space.Spec.MountPath,  // 直接使用 MountPath 而不是 workspaceDir
		}
		
		// 如果是增强版Claude镜像，使用特殊的启动脚本
		if strings.Contains(space.Spec.Image, "claude-code-server") {
			container.Command = []string{"/app/start-claude-code.sh"}
			container.Args = []string{}
			// 添加环境变量
			container.Env = append(container.Env, v1.EnvVar{
				Name:  "WORKSPACE_DIR",
				Value: space.Spec.MountPath,
			})
		}
	}
	
	// 为简化版Claude AI镜像配置启动命令
	if strings.Contains(space.Spec.Image, "claude-ai-simple") {
		// 使用镜像默认的entrypoint，只添加环境变量
		container.Env = append(container.Env, v1.EnvVar{
			Name:  "WORKSPACE_DIR",
			Value: space.Spec.MountPath,
		})
	}
	
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      space.Name,
			Namespace: space.Namespace,
			Annotations: map[string]string{
				"sid": space.Spec.SID,
				"uid": space.Spec.UID,
			},
			Labels: map[string]string{
				"app": "cloud-ide",
			},
		},
		Spec: v1.PodSpec{
			Volumes: []v1.Volume{
				{
					Name: volumeName,
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: space.Name,
							ReadOnly:  false,
						},
					},
				},
			},
			Containers: []v1.Container{container},
		},
	}

	// 设置资源限制
	if Mode == ModeRelease {
		// 最小需求CPU2核、内存1Gi == 1 * 2^10
		pod.Spec.Containers[0].Resources = v1.ResourceRequirements{
			Requests: map[v1.ResourceName]resource.Quantity{
				v1.ResourceCPU:    resource.MustParse("2"),
				v1.ResourceMemory: resource.MustParse("1Gi"),
			},
			Limits: map[v1.ResourceName]resource.Quantity{
				v1.ResourceCPU:    resource.MustParse(space.Spec.Cpu),
				v1.ResourceMemory: resource.MustParse(space.Spec.Memory),
			},
		}
	}

	if actualGitRepo == "" {
		return pod
	}

	// 如果设置了git仓库，则通过init容器来clone
	idx := strings.LastIndexByte(actualGitRepo, '/') + 1

	localPath := filepath.Join(workspaceDir, strings.TrimSuffix(actualGitRepo[idx:], ".git"))
	pod.Spec.InitContainers = []v1.Container{
		{
			Name:            "git-cloner",
			Image:           GitClonerName,
			WorkingDir:      space.Spec.MountPath,
			ImagePullPolicy: v1.PullIfNotPresent,
			// 容器挂载存储卷
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      volumeName,
					ReadOnly:  false,
					MountPath: space.Spec.MountPath,
				},
			},
			Env: []v1.EnvVar{
				{
					Name:  "REPO_URL",
					Value: actualGitRepo,
				},
				{
					Name:  "LOCAL_PATH",
					Value: localPath,
				},
			},
		},
	}
	// 设置环境变量，code-server打开时使用该路径
	pod.Spec.Containers[0].Env[0].Value = localPath

	return pod
}

func (r *WorkSpaceReconciler) createPVC(ctx context.Context, space *mv1.WorkSpace, key client.ObjectKey) error {
	lgr := log.FromContext(ctx)
	// 1.先检查PVC是否已经存在
	exist, err := r.checkPVCExist(ctx, key)
	if err != nil {
		// PVC已经存在
		return err
	}

	// PVC已经存在,无需创建
	if exist {
		return nil
	}

	// 2.PVC不存在,创建PVC
	pvc, err := r.constructPVC(space)
	if err != nil {
		lgr.Error(err, "construct pvc")
		return err
	}

	// 设置了OwnerReference之后,PVC的状态发生变化,也会触发Reconcile方法
	// 但是对于PVC来说,我们不希望它触发这个方法,因此我们可以使用过滤器来进行过滤
	if err = controllerutil.SetControllerReference(space, pvc, r.Scheme); err != nil {
		return err
	}

	ctx, cancelFunc := context.WithTimeout(ctx, time.Second*30)
	defer cancelFunc()
	err = r.Client.Create(ctx, pvc)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}

		return err
	}

	return nil
}

// 构造PVC对象
func (r *WorkSpaceReconciler) constructPVC(space *mv1.WorkSpace) (*v1.PersistentVolumeClaim, error) {
	quantity, err := resource.ParseQuantity(space.Spec.Storage)
	if err != nil {
		return nil, err
	}

	pvc := &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      space.Name,
			Namespace: space.Namespace,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Resources: v1.ResourceRequirements{
				Limits:   v1.ResourceList{v1.ResourceStorage: quantity},
				Requests: v1.ResourceList{v1.ResourceStorage: quantity},
			},
		},
	}

	// 如果启用了动态卷制备，则需要指明StorageClassName
	if DynamicStorageEnabled {
		pvc.Spec.StorageClassName = &StorageClassName
	}

	return pvc, nil
}

func (r *WorkSpaceReconciler) checkPodExist(ctx context.Context, key client.ObjectKey) (bool, error) {
	lgr := log.FromContext(ctx)

	pod := &v1.Pod{}
	// 先查询一下
	err := r.Client.Get(context.Background(), key, pod)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}

		lgr.Error(err, "get pod")
		return false, err
	}

	return true, nil
}

func (r *WorkSpaceReconciler) deletePod(ctx context.Context, key client.ObjectKey) error {
	lgr := log.FromContext(ctx)

	exist, err := r.checkPodExist(ctx, key)
	if err != nil {
		return err
	}

	// Pod不存在,直接返回
	if !exist {
		return nil
	}

	pod := &v1.Pod{}
	pod.Name = key.Name
	pod.Namespace = key.Namespace

	ctx, cancelFunc := context.WithTimeout(ctx, time.Second*30)
	defer cancelFunc()
	// 删除Pod
	err = r.Client.Delete(ctx, pod)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}

		lgr.Error(err, "delete pod")
		return err
	}

	return nil
}

func (r *WorkSpaceReconciler) checkPVCExist(ctx context.Context, key client.ObjectKey) (bool, error) {
	lgr := log.FromContext(ctx)

	pvc := &v1.PersistentVolumeClaim{}
	err := r.Client.Get(context.Background(), key, pvc)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}

		lgr.Error(err, "get pvc")
		return false, err
	}

	return true, nil
}

func (r *WorkSpaceReconciler) deletePVC(ctx context.Context, key client.ObjectKey) error {
	lgr := log.FromContext(ctx)

	exist, err := r.checkPVCExist(ctx, key)
	if err != nil {
		return err
	}

	// pvc不存在,无需再删除
	if !exist {
		return nil
	}

	pvc := &v1.PersistentVolumeClaim{}
	pvc.Name = key.Name
	pvc.Namespace = key.Namespace

	ctx, cancelFunc := context.WithTimeout(ctx, time.Second*30)
	defer cancelFunc()
	err = r.Client.Delete(ctx, pvc)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}

		lgr.Error(err, "delete pvc")
		return err
	}

	return nil
}

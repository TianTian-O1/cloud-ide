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

package main

import (
	"flag"
	"os"

	"github.com/mangohow/cloud-ide/cmd/control-plane/internal/controllers"
	"github.com/mangohow/cloud-ide/cmd/control-plane/internal/rpc"
	"github.com/mangohow/cloud-ide/cmd/control-plane/internal/service"
	"github.com/mangohow/cloud-ide/pkg/notifier"
	"github.com/mangohow/cloud-ide/pkg/proc"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	cloudidev1 "github.com/mangohow/cloud-ide/cmd/control-plane/internal/api/v1"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(cloudidev1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var (
		metricsAddr          string
		enableLeaderElection bool
		probeAddr            string

		gatewayToken   string
		gatewayPath    string
		gatewayService string
	)

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	// 指定namespace
	flag.StringVar(&controllers.WorkspaceNamespace, "ns", "cloud-ide-ws", "The namespace controller listened")
	// 指定运行模式
	flag.StringVar(&controllers.Mode, "mode", controllers.ModeRelease, "The mode program running")
	// 指定gateway的token，在下发配置时需要使用
	flag.StringVar(&gatewayToken, "gateway-token", "", "specify gateway token")
	// 指定gateway的访问路径
	flag.StringVar(&gatewayPath, "gateway-path", "/internal/endpoint", "specify gateway path")
	// 指定gateway的service name
	flag.StringVar(&gatewayService, "gateway-service", "cloud-ide-gateway-svc", "specify gateway service")
	// 指定动态卷的storageClass
	flag.StringVar(&controllers.StorageClassName, "storage-class-name", "nfs-csi", "specify storage class name if dynamic-storage-enabled enabled")
	// 指定是否启用动态卷制备
	flag.BoolVar(&controllers.DynamicStorageEnabled, "dynamic-storage-enabled", false, "specify dynamic storage enabled")
	// 指定用于克隆git的初始化容器镜像
	flag.StringVar(&controllers.GitClonerName, "git-cloner-image", "git-cloner", "specify git cloner images")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)

	if gatewayToken == "" {
		logger.Error(nil, "must specify gateway token")
		os.Exit(1)
	}

	logger.Info("watched namespace", "namespace", controllers.WorkspaceNamespace)
	logger.Info("running mode", "mode", controllers.Mode)
	logger.Info("dynamic storage enabled", "value", controllers.DynamicStorageEnabled, "StorageClassName", controllers.StorageClassName)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "f4321edb.mangohow.com",
		Namespace:              controllers.WorkspaceNamespace,
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	ctx := proc.SetupSignalHandler()

	if err = controllers.NewWorkSpaceReconciler(
		mgr.GetClient(),
		mgr.GetScheme()).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create  controller", "controller", "WorkSpace")
		os.Exit(1)
	}

	ntf, err := notifier.NewWorkspaceNotifier(ctx, logger, gatewayService, gatewayPath, gatewayToken, 8)
	if err != nil {
		panic(err)
	}
	if err = controllers.NewPodReconciler(
		mgr.GetClient(),
		mgr.GetScheme(),
		ntf,
	).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Pod")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	// 将grpc交由manager管理,manager会调用Start方法启动
	if err := mgr.Add(rpc.New(":6387", logger, service.NewWorkSpaceService(mgr.GetClient(), logger, ntf, controllers.WorkspaceNamespace))); err != nil {
		setupLog.Error(err, "unable to set up grpc server")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

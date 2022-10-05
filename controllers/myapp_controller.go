/*
Copyright 2022 ll.

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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	//_ "k8s.io/client-go/applyconfigurations/apps/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1beta1 "github.com/Sion-L/operator-demo/api/v1beta1"
)

var (
	oldSpecAnnotation = "old/spec"
)

// MyAppReconciler reconciles a MyApp object
type MyAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.ll.io,resources=myapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.ll.io,resources=myapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.ll.io,resources=myapps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MyApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *MyAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// 先获取MyApp实例
	var myapp appv1beta1.MyApp
	if err := r.Client.Get(ctx, req.NamespacedName, &myapp); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err) // 没有就会重试，重新入队列，需要对应的err返回值，不能省
	}

	// 调谐，获取当前的一个状态，然后和我们期望的状态进行比对
	// CreateOrUpdate -> deployment
	var deployer appsv1.Deployment
	deployer.Name = myapp.Name
	deployer.Namespace = myapp.Namespace
	or, err := ctrl.CreateOrUpdate(ctx, r.Client, &deployer, func() error {
		// 调谐必须在这个方法中实现，or是调谐后的结果,类似之前更新用的反射
		MutateDeployment(&myapp, &deployer) // 填充deployment
		return controllerutil.SetControllerReference(&myapp, &deployer, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	logger.Info("CreateOrUpdate", "Deployment", or)

	// CreateOrUpdate -> service
	var svc corev1.Service
	svc.Name = myapp.Name
	svc.Namespace = myapp.Namespace
	or, err = ctrl.CreateOrUpdate(ctx, r.Client, &svc, func() error {
		// 调谐必须在这个方法中实现
		MutateService(&myapp, &svc) // 填充deployment
		return controllerutil.SetControllerReference(&myapp, &svc, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	logger.Info("CreateOrUpdate", "Service", or)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1beta1.MyApp{}).   // owns表示MyApp控制器下的deploy和svc等等资源
		Owns(&appsv1.Deployment{}). // 加上对应的owns，即使手动删除了deploy和svc也会将deploy和svc传递给Reconcile自动重建。加上watch可以查看监控其它类型的资源对象
		Owns(&corev1.Service{}).
		Complete(r)
}

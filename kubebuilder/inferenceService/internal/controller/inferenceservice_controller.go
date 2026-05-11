package controller

import (
	"context"
	"fmt"
	"inference-gateway/internal/resources"
	"inference-gateway/pkg/modelmanager"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	aiv1 "inference-gateway/api/v1"
)

type InferenceServiceReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	ModelManagerAddr string               // 模型服务地址
	ModelClient      *modelmanager.Client //客户端实例
}

//+kubebuilder:rbac:groups=inference.kubeai.platform.io,resources=inferenceservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=inference.kubeai.platform.io,resources=inferenceservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=inference.kubeai.platform.io,resources=inferenceservices/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deploy,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.

// Reconcile 实现协调循环
func (r *InferenceServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// 0. 获取 InferenceService 实例
	isvc := &aiv1.InferenceService{}
	if err := r.Get(ctx, req.NamespacedName, isvc); err != nil {
		log.Error(err, "unable to fetch InferenceService")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 1. 调用 Model Manager 获取模型路径
	log.Info("Fetching model metadata", "ModelName", isvc.Spec.ModelName, "Version", isvc.Spec.ModelVersion)
	var modelMeta *modelmanager.ModelMetadata

	// 如果模型是自定义镜像,否则用默认推理镜像+模型挂载
	if r.ModelClient != nil && isvc.Spec.Image == "" {
		modelMeta, err := r.ModelClient.GetModelMetadata(isvc.Spec.ModelName, isvc.Spec.ModelVersion)
		if err != nil {
			log.Error(err, "unable to fetch model metadata", "ModelName", isvc.Spec.ModelName)
			// 更新 CR 状态为 "ModelNotReady"
			_ = r.setModelNotReadyStatus(ctx, isvc, err.Error())
			// 返回错误，触发 Controller 重试（指数退避）
			return ctrl.Result{}, err
		}
		log.Info("Successfully fetched model path", "Path", modelMeta.StoragePath)
	} else {
		log.V(1).Info("Model Manager client not configured, skipping metadata fetch")
	}

	// 2. 协调 Stable Deployment
	stableDeploy := resources.NewStableDeployment(isvc, modelMeta)
	if err := r.reconcileDeployment(ctx, isvc, stableDeploy); err != nil {
		log.Error(err, "Failed to reconcile Stable Deployment")
		return ctrl.Result{}, err
	}

	// 3. 协调 Stable Service
	stableSvc := resources.NewStableService(isvc)
	if err := r.reconcileService(ctx, isvc, stableSvc); err != nil {
		log.Error(err, "Failed to reconcile Stable Service")
		return ctrl.Result{}, err
	}

	// 4. 协调 Canary 资源
	if isvc.Spec.Canary != nil && isvc.Spec.Canary.Enabled {
		// Canary Deployment
		canaryDeploy := resources.NewCanaryDeployment(isvc, modelMeta)
		if err := r.reconcileDeployment(ctx, isvc, canaryDeploy); err != nil {
			log.Error(err, "Failed to reconcile Canary Deployment")
			return ctrl.Result{}, err
		}
		// Canary Service
		canarySvc := resources.NewCanaryService(isvc)
		if err := r.reconcileService(ctx, isvc, canarySvc); err != nil {
			log.Error(err, "Failed to reconcile Canary Service")
			return ctrl.Result{}, err
		}
		// Canary Ingress
		canaryIng := resources.NewCanaryIngress(isvc)
		if canaryIng != nil {
			if err := r.reconcileIngress(ctx, isvc, canaryIng); err != nil {
				log.Error(err, "Failed to reconcile Canary Ingress")
				return ctrl.Result{}, err
			}
		}
	} else {
		// 如果未启用 Canary，尝试清理可能存在的 Canary 资源
		// 这里简化处理，实际生产中可以加上 Finalizer 逻辑
	}

	// 5. 协调主 Ingress
	mainIng := resources.NewIngress(isvc)
	if mainIng != nil {
		if err := r.reconcileIngress(ctx, isvc, mainIng); err != nil {
			log.Error(err, "Failed to reconcile main Ingress")
			return ctrl.Result{}, err
		}
	}

	// 6. 协调 HPA
	hpa := resources.NewHPA(isvc)
	if hpa != nil {
		if err := r.reconcileHPA(ctx, isvc, hpa); err != nil {
			log.Error(err, "Failed to reconcile HPA")
			return ctrl.Result{}, err
		}
	} else {
		// 如果没有配置 Autoscaling，删除已存在的 HPA
		existingHPA := &autoscalingv2.HorizontalPodAutoscaler{}
		err := r.Get(ctx, types.NamespacedName{Name: isvc.Name, Namespace: isvc.Namespace}, existingHPA)
		if err == nil {
			log.Info("Deleting existing HPA as Autoscaling is not configured")
			if err := r.Delete(ctx, existingHPA); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	// 7. 更新 Status
	if err := r.updateStatus(ctx, isvc); err != nil {
		log.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil

}

// reconcileDeployment 通用的 Deployment 调和函数
func (r *InferenceServiceReconciler) reconcileDeployment(ctx context.Context, isvc *aiv1.InferenceService, desired *appsv1.Deployment) error {
	log := log.FromContext(ctx)
	if desired == nil {
		return nil
	}
	found := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Deployment", "Deployment.Namespace", desired.Namespace, "Deployment.Name", desired.Name)
		if err := r.Create(ctx, desired); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		// 更新逻辑：这里简化处理，实际中可能需要更精细的 Patch 策略
		log.Info("Updating Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
		// 保留一些 runtime 字段
		desired.ResourceVersion = found.ResourceVersion
		desired.Spec.Replicas = found.Spec.Replicas // 保留 HPA 调整后的副本数
		if err := r.Update(ctx, desired); err != nil {
			return err
		}
	}
	return nil
}

// reconcileService 通用的 Service 调和函数
func (r *InferenceServiceReconciler) reconcileService(ctx context.Context, isvc *aiv1.InferenceService, desired *corev1.Service) error {
	log := log.FromContext(ctx)
	if desired == nil {
		return nil
	}
	found := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Service", "Service.Namespace", desired.Namespace, "Service.Name", desired.Name)
		if err := r.Create(ctx, desired); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		// Service 的 ClusterIP 是 immutable的，更新时需要注意
		log.Info("Updating Service", "Service.Namespace", found.Namespace, "Service.Name", found.Name)
		desired.ResourceVersion = found.ResourceVersion
		desired.Spec.ClusterIP = found.Spec.ClusterIP
		if err := r.Update(ctx, desired); err != nil {

		}
	}
	return nil
}

// reconcileIngress 通用的 Ingress 调和函数
func (r *InferenceServiceReconciler) reconcileIngress(ctx context.Context, isvc *aiv1.InferenceService, desired *networkingv1.Ingress) error {
	log := log.FromContext(ctx)
	if desired == nil {
		return nil
	}

	found := &networkingv1.Ingress{}
	err := r.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Ingress", "Ingress.Namespace", desired.Namespace, "Ingress.Name", desired.Name)
		if err := r.Create(ctx, desired); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		log.Info("Updating Ingress", "Ingress.Namespace", found.Namespace, "Ingress.Name", found.Name)
		desired.ResourceVersion = found.ResourceVersion
		if err := r.Update(ctx, desired); err != nil {
			return err
		}
	}
	return nil
}

// reconcileHPA 调和 HPA
func (r *InferenceServiceReconciler) reconcileHPA(ctx context.Context, isvc *aiv1.InferenceService, desired *autoscalingv2.HorizontalPodAutoscaler) error {
	log := log.FromContext(ctx)
	if desired == nil {
		return nil
	}

	found := &autoscalingv2.HorizontalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating HPA", "HPA.Namespace", desired.Namespace, "HPA.Name", desired.Name)
		if err := r.Create(ctx, desired); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		log.Info("Updating HPA", "HPA.Namespace", found.Namespace, "HPA.Name", found.Name)
		desired.ResourceVersion = found.ResourceVersion
		if err := r.Update(ctx, desired); err != nil {
			return err
		}
	}
	return nil
}

// updateStatus 更新 CR 状态
func (r *InferenceServiceReconciler) updateStatus(ctx context.Context, isvc *aiv1.InferenceService) error {
	log := log.FromContext(ctx)
	// 获取 Stable Deployment 状态
	deploy := &appsv1.Deployment{}
	if err := r.Get(ctx, types.NamespacedName{Name: isvc.Name + "-stable", Namespace: isvc.Namespace}, deploy); err != nil {
		return err
	}

	// 构造 URL
	var url string
	if isvc.Spec.Service != nil && isvc.Spec.Service.Host != "" {
		url = fmt.Sprintf("http://%s", isvc.Spec.Service.Host)
	} else {
		// 如果没有 Ingress，使用 Service 名称 (仅集群内访问)
		url = fmt.Sprintf("http://%s-stable.%s.svc.cluster.local", isvc.Name, isvc.Namespace)
	}

	// 更新 Status 字段
	isvc.Status.URL = url
	isvc.Status.ReadyReplicas = deploy.Status.ReadyReplicas
	isvc.Status.Ready = deploy.Status.ReadyReplicas > 0

	// 设置条件
	isvc.Status.Conditions = []metav1.Condition{
		{
			Type:               "Ready",
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             "DeploymentReady",
			Message:            "Inference service deployed successfully",
		},
	}

	log.Info("Updating InferenceService Status", "URL", url, "ReadyReplicas", deploy.Status.ReadyReplicas)
	return r.Status().Update(ctx, isvc)
}

// SetupWithManager sets up the controller with the Manager.
func (r *InferenceServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aiv1.InferenceService{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&networkingv1.Ingress{}).
		Owns(&autoscalingv2.HorizontalPodAutoscaler{}).
		Complete(r)
}

// 辅助函数：更新状态为模型未就绪
func (r *InferenceServiceReconciler) setModelNotReadyStatus(ctx context.Context, isvc *aiv1.InferenceService, msg string) error {
	// 这里简化处理，实际应使用 metav1.Condition
	isvc.Status.Ready = false
	// 可以在 Status 里加一个 Message 字段
	return r.Status().Update(ctx, isvc)
}

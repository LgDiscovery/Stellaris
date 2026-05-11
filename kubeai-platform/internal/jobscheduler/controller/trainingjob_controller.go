/*
Copyright 2026.

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
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	// Kubeflow PyTorchJob 依赖
	trainingv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	aiv1 "training-job/api/v1"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

// TrainingJobReconciler reconciles a TrainingJob object
type TrainingJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=training.kubeai.platform.io,resources=trainingjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=training.kubeai.platform.io,resources=trainingjobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=training.kubeai.platform.io,resources=trainingjobs/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

func (r *TrainingJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// 1.获取 TrainingJob CR
	var tj aiv1.TrainingJob
	if err := r.Get(ctx, req.NamespacedName, &tj); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2. 状态直接退出
	if tj.Status.Phase == "Succeeded" || tj.Status.Phase == "Failed" {
		return ctrl.Result{}, nil
	}

	// 3. 检查底层PyTorchJob是否存在
	var job trainingv1.PyTorchJob
	err := r.Get(ctx, client.ObjectKey{Name: tj.Status.JobName, Namespace: tj.Namespace}, &job)
	if err != nil && !errors.IsNotFound(err) {
		// 4. Job不存在 创建 分布式训练任务 jobscheduler
		job, err := r.buildDistributedPyTorchJob(&tj)
		if err != nil {
			log.Error(err, "failed to build k8s jobscheduler")
			r.updateStatus(&tj, "Failed", "JobBuildFailed", err.Error())
			return ctrl.Result{}, err
		}

		// 设置 OwnerReference ,自动GC
		if err := ctrl.SetControllerReference(&tj, job, r.Scheme); err != nil {
			log.Error(err, "failed to set controller reference")
			return ctrl.Result{}, err
		}

		// 创建 Job
		if err := r.Create(ctx, job); err != nil {
			log.Error(err, "failed to create jobscheduler")
			r.updateStatus(&tj, "Failed", "JobCreateFailed", err.Error())
			return ctrl.Result{}, err
		}

		// 更新 CR 状态
		tj.Status.JobName = job.Name
		tj.Status.StartTime = &metav1.Time{Time: time.Now()}
		r.updateStatus(&tj, "Pending", "JobCreated", "Job created, waiting for pod")
		return ctrl.Result{RequeueAfter: 3 * time.Second}, r.Status().Update(ctx, &tj)
	} else if err != nil {
		log.Error(err, "Get jobscheduler failed")
		return ctrl.Result{}, err
	}
	// 5. 同步分布式任务状态
	r.syncPyTorchJobStatus(&tj, &job)
	if err := r.Status().Update(ctx, &tj); err != nil {
		return ctrl.Result{}, err
	}

	// 6. 定期同步
	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

// syncPyTorchJobStatus
// 从 Kubeflow PyTorchJob 同步状态到我们的 TrainingJob CR
// 支持 Pending / Running / Succeeded / Failed
func (r *TrainingJobReconciler) syncPyTorchJobStatus(tj *aiv1.TrainingJob, pytorchJob *trainingv1.PyTorchJob) {
	tj.Status.Active = pytorchJob.Status.ReplicaStatuses[trainingv1.PyTorchJobReplicaTypeMaster].Active +
		pytorchJob.Status.ReplicaStatuses[trainingv1.PyTorchJobReplicaTypeWorker].Active

	tj.Status.Succeeded = pytorchJob.Status.ReplicaStatuses[trainingv1.PyTorchJobReplicaTypeMaster].Succeeded +
		pytorchJob.Status.ReplicaStatuses[trainingv1.PyTorchJobReplicaTypeWorker].Succeeded

	tj.Status.Failed = pytorchJob.Status.ReplicaStatuses[trainingv1.PyTorchJobReplicaTypeMaster].Failed +
		pytorchJob.Status.ReplicaStatuses[trainingv1.PyTorchJobReplicaTypeWorker].Failed

	switch {
	case tj.Status.Succeeded > 0:
		r.updateStatus(tj, "Succeeded", "JobCompleted", "Training success")
		tj.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	case tj.Status.Failed > 0:
		r.updateStatus(tj, "Failed", "JobFailed", "Training failed")
		tj.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	case tj.Status.Active > 0:
		r.updateStatus(tj, "Running", "JobRunning", "Training running")
	default:
		r.updateStatus(tj, "Pending", "PodPending", "Waiting for pod scheduling")
	}
}

// buildDistributedPyTorchJob 构建Kubeflow分布式训练任务
func (r *TrainingJobReconciler) buildDistributedPyTorchJob(tj *aiv1.TrainingJob) (*trainingv1.PyTorchJob, error) {
	jobName := fmt.Sprintf("%s-%s-dist", tj.Name, randString(5))
	replicaSpecs := map[trainingv1.ReplicaType]*trainingv1.ReplicaSpec{}

	backoffLimit := int32(3)
	if tj.Spec.BackoffLimit > 0 {
		backoffLimit = tj.Spec.BackoffLimit
	}

	// 构建容器
	container := corev1.Container{
		Name:         "trainer",
		Image:        tj.Spec.Image,
		Command:      tj.Spec.Command,
		Env:          tj.Spec.Env,
		VolumeMounts: tj.Spec.VolumeMounts,
		Resources:    r.buildResources(tj.Spec.Resources),
		// 自动注入监控配置
		Args: r.buildMonitorArgs(tj.Spec.EnableMonitor),
	}

	// Master节点
	replicaSpecs[trainingv1.PyTorchJobReplicaTypeMaster] = &trainingv1.ReplicaSpec{
		Replicas:      ptr[int32](1),
		RestartPolicy: trainingv1.RestartPolicyOnFailure,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels:      r.buildPodLabels(tj), // 注入日志/监控标签
				Annotations: r.buildMonitorAnnotations(tj.Spec.EnableMonitor),
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{container},
				Volumes:    tj.Spec.Volumes,
			},
		},
	}

	// Worker 节点 分布式开启
	if tj.Spec.Distributed && tj.Spec.WorkerNum > 0 {
		replicaSpecs[trainingv1.PyTorchJobReplicaTypeWorker] = &trainingv1.ReplicaSpec{
			Replicas:      &tj.Spec.WorkerNum,
			Template:      replicaSpecs[trainingv1.PyTorchJobReplicaTypeMaster].Template,
			RestartPolicy: trainingv1.RestartPolicyOnFailure,
		}
	}
	return &trainingv1.PyTorchJob{
		ObjectMeta: metav1.ObjectMeta{Name: jobName, Namespace: tj.Namespace},
		Spec: trainingv1.PyTorchJobSpec{
			RunPolicy: trainingv1.RunPolicy{
				BackoffLimit:            &backoffLimit,
				TTLSecondsAfterFinished: &tj.Spec.TTLSecondsAfterFinished,
			},
		},
	}, nil
}

// buildResources 构建资源限制
func (r *TrainingJobReconciler) buildResources(res aiv1.ResourceRequirements) corev1.ResourceRequirements {
	reqList := corev1.ResourceList{}
	limitList := corev1.ResourceList{}

	if res.CPU != "" {
		reqList[corev1.ResourceCPU] = resource.MustParse(res.CPU)
		limitList[corev1.ResourceCPU] = resource.MustParse(res.CPU)
	}
	if res.Memory != "" {
		reqList[corev1.ResourceMemory] = resource.MustParse(res.Memory)
		limitList[corev1.ResourceMemory] = resource.MustParse(res.Memory)
	}
	// GPU（只配置 limits，这是标准做法）
	if res.GPU != "" && strings.TrimSpace(res.GPU) != "0" {
		limitList["nvidia.com/gpu"] = resource.MustParse(res.GPU)
	}

	return corev1.ResourceRequirements{Requests: reqList, Limits: limitList}
}

// updateStatus
// 更新 TrainingJob CR 的 phase、condition、状态信息
// 标准化云原生状态流转
func (r *TrainingJobReconciler) updateStatus(tj *aiv1.TrainingJob, phase, reason, msg string) {
	now := metav1.Now()
	tj.Status.Phase = phase
	condition := metav1.Condition{
		Type:               "TrainingCompleted",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: now,
		Reason:             reason,
		Message:            msg,
	}
	tj.Status.Conditions = []metav1.Condition{condition}
}

// randString 随机字符串
func randString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// SetupWithManager sets up the controller with the Manager.
func (r *TrainingJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aiv1.TrainingJob{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}

// ptr工具函数： 创建int32指针
func ptr[T any](v T) *T {
	return &v
}

// buildPodLabels
// 给训练 Pod 打标签，用于：
// 1. 日志检索（Loki）
// 2. 监控指标（Prometheus）
// 3. 服务发现
func (r *TrainingJobReconciler) buildPodLabels(tj *aiv1.TrainingJob) map[string]string {
	labels := map[string]string{
		"app":                             "kube-ai-training-jobscheduler",
		"training.jobscheduler.name":      tj.Name,
		"training.jobscheduler.namespace": tj.Namespace,
		"training.framework":              tj.Spec.Framework,
		"training.distributed":            strconv.FormatBool(tj.Spec.Distributed),
	}

	if tj.Spec.ModelID != "" {
		labels["training.model.id"] = fmt.Sprintf("%d", tj.Spec.ModelID)
	}
	return labels
}

// buildMonitorAnnotations
// 自动注入 Prometheus 监控注解
// 开启后 Prometheus 自动抓取训练指标（loss/epoch/accuracy）
func (r *TrainingJobReconciler) buildMonitorAnnotations(enableMonitor bool) map[string]string {
	if !enableMonitor {
		return nil
	}

	return map[string]string{
		"prometheus.io/scrape":       "true",
		"prometheus.io/port":         "8082",
		"prometheus.io/path":         "/metrics",
		"prometheus.io/jobscheduler": "training-jobscheduler",
	}
}

// buildMonitorArgs
// 向容器注入监控相关启动参数（可选）
func (r *TrainingJobReconciler) buildMonitorArgs(enableMonitor bool) []string {
	if !enableMonitor {
		return nil
	}
	return []string{}
}

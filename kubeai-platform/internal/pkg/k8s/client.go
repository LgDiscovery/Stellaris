package k8s

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/internalversion/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"kubeai/internal/model"
	"os"
	trainingjobv1 "training-job/api/v1"
)

type K8sClient struct {
	clientSet         *kubernetes.Clientset
	trainingJobClient rest.Interface // CRD客户端：操作TrainingJob（新增）
	namespace         string
}

func NewK8sClient() (*K8sClient, error) {
	var config *rest.Config
	var err error
	// 加载K8s配置（Pod内/本地kubeconfig兼容）
	if config, err = rest.InClusterConfig(); err != nil {
		// 回退到 kubeconfig
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = clientcmd.RecommendedHomeFile
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	// 2. 注册TrainingJob CRD到Scheme（必须！）
	err = trainingjobv1.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, fmt.Errorf("register trainingjob scheme failed: %v", err)
	}

	// 3. 初始化原生K8s客户端（原有逻辑不变）
	clientSet, err := kubernetes.NewForConfig(config)

	if err != nil {
		return nil, err
	}

	namespace := os.Getenv("K8s_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	// 4. 新增：初始化TrainingJob CRD 专用REST客户端
	trainingJobCfg := *config
	trainingJobCfg.GroupVersion = &trainingjobv1.GroupVersion // CRD的Group/Version
	trainingJobCfg.APIPath = "/apis"                          // CRD固定API路径
	trainingJobCfg.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	trainingJobCfg.UserAgent = rest.DefaultKubernetesUserAgent()

	trainingJobClient, err := rest.RESTClientFor(&trainingJobCfg)
	if err != nil {
		return nil, fmt.Errorf("init trainingjob client failed: %v", err)
	}

	return &K8sClient{
		clientSet:         clientSet,
		trainingJobClient: trainingJobClient,
		namespace:         namespace,
	}, nil
}

// CreateJob 创建K8s Job, 返回job名称
func (c *K8sClient) CreateJob(ctx context.Context, name, image, command string, resources map[string]interface{}, envVars []map[string]string) (string, error) {
	// 解析资源需求
	var cpu, memory string
	var gpu int64

	if resources != nil {
		if v, ok := resources["cpu"].(string); ok {
			cpu = v
		}
		if v, ok := resources["memory"].(string); ok {
			memory = v
		}
		if v, ok := resources["gpu"].(float64); ok {
			gpu = int64(v)
		}
	}

	// 构建环境变量
	var env []corev1.EnvVar
	for _, v := range envVars {
		env = append(env, corev1.EnvVar{
			Name:  v["name"],
			Value: v["value"],
		})
	}

	// 构建容器资源需求
	resourceRequirements := corev1.ResourceRequirements{
		Limits:   corev1.ResourceList{},
		Requests: corev1.ResourceList{},
	}

	if cpu != "" {
		resourceRequirements.Limits[corev1.ResourceCPU] = resource.MustParse(cpu)
		resourceRequirements.Requests[corev1.ResourceCPU] = resource.MustParse(cpu)
	}
	if memory != "" {
		resourceRequirements.Limits[corev1.ResourceMemory] = resource.MustParse(memory)
		resourceRequirements.Requests[corev1.ResourceMemory] = resource.MustParse(memory)
	}
	if gpu > 0 {
		resourceRequirements.Limits["nvidia/gpu"] = *resource.NewQuantity(gpu, resource.DecimalSI)
		resourceRequirements.Requests["nvidia/gpu"] = *resource.NewQuantity(gpu, resource.DecimalSI)
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:      "trainer",
							Image:     image,
							Command:   []string{"/bin/sh", "-c"},
							Args:      []string{command},
							Env:       env,
							Resources: resourceRequirements,
						},
					},
				},
			},
		},
	}
	_, err := c.clientSet.BatchV1().Jobs(c.namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return "", err
	}
	return job.Name, nil
}

// GetJobStatus 获取Job状态
func (c *K8sClient) GetJobStatus(ctx context.Context, jobName string) (string, error) {
	job, err := c.clientSet.BatchV1().Jobs(c.namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	if job.Status.Succeeded > 0 {
		return "succeeded", nil
	} else if job.Status.Failed > 0 {
		return "failed", nil
	} else if job.Status.Active > 0 {
		return "running", nil
	} else {
		return "pending", nil
	}
}

// DeleteJob 删除job
func (c *K8sClient) DeleteJob(ctx context.Context, jobName string) error {
	// 删除Job会级联删除关联的Pod
	return c.clientSet.BatchV1().Jobs(c.namespace).Delete(ctx, jobName, metav1.DeleteOptions{})
}

// GetJobLogs 获取Job的Pod日志(取第一个pod)
func (c *K8sClient) GetJobLogs(ctx context.Context, jobName string, tailLines int64) (string, error) {
	// 找到该Job的Pod
	pods, err := c.clientSet.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("jobscheduler-name=%s", jobName),
	})
	if err != nil {
		return "", err
	}
	if len(pods.Items) == 0 {
		return "", fmt.Errorf("no pods found for jobscheduler %s", jobName)
	}
	// 取第一个Pod的日志
	pod := pods.Items[0]
	opts := corev1.PodLogOptions{}
	if tailLines > 0 {
		opts.TailLines = &tailLines
	}
	req := c.clientSet.CoreV1().Pods(c.namespace).GetLogs(pod.Name, &opts)
	logs, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer logs.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, logs)
	return buf.String(), err
}

// CreateTrainingJob 创建TrainingJob CR
func (c *K8sClient) CreateTrainingJob(ctx context.Context, job *trainingjobv1.TrainingJob) error {
	job.Namespace = c.namespace
	return c.trainingJobClient.Post().
		Namespace(c.namespace).
		Resource("trainingjobs"). // CRD资源名（小写复数）
		Body(job).
		Do(ctx).
		Error()
}

// GetTrainingJob 查询TrainingJob
func (c *K8sClient) GetTrainingJob(ctx context.Context, name string) (*trainingjobv1.TrainingJob, error) {
	var job trainingjobv1.TrainingJob
	err := c.trainingJobClient.Get().
		Namespace(c.namespace).
		Resource("trainingjobs").
		Name(name).
		Do(ctx).
		Into(&job)
	return &job, err
}

// DeleteTrainingJob 删除TrainingJob
func (c *K8sClient) DeleteTrainingJob(ctx context.Context, name string) error {
	return c.trainingJobClient.Delete().
		Namespace(c.namespace).
		Resource("trainingjobs").
		Name(name).
		Do(ctx).
		Error()
}

// ListTrainingJobs 列表查询TrainingJob
func (c *K8sClient) ListTrainingJobs(ctx context.Context) (*trainingjobv1.TrainingJobList, error) {
	var list trainingjobv1.TrainingJobList
	err := c.trainingJobClient.Get().
		Namespace(c.namespace).
		Resource("trainingjobs").
		Do(ctx).
		Into(&list)
	return &list, err
}

func (c *K8sClient) CheckResourceAvailable(ctx context.Context, req model.ResourceRequest) (bool, error) {
	// 1.获取集群所有节点
	nodes, err := c.clientSet.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, err
	}

	if len(nodes.Items) == 0 {
		return false, errors.New("no nodes found in cluster")
	}

	// 2. 解析需要的资源量
	requiredCPU, err := resource.ParseQuantity(req.CPU)
	if err != nil {
		return false, err
	}

	requiredMemory, err := resource.ParseQuantity(req.Memory)
	if err != nil {
		return false, err
	}

	// GPU 需求（默认0）
	requiredGPU := int64(0)
	if req.GPU != "" {
		q, err := resource.ParseQuantity(req.GPU)
		if err != nil {
			return false, err
		}
		requiredGPU = q.Value()
	}

	// 3. 遍历每个节点，判断是否有一个节点满足资源要求
	for _, node := range nodes.Items {
		// 节点不可调度 → 跳过
		if node.Spec.Unschedulable {
			continue
		}

		// 节点可分配资源
		allocatable := node.Status.Allocatable

		// 4. 计算节点当前已被占用的资源
		usedCPU, usedMemory, usedGPU, err := c.getNodeAllocatedResources(ctx, node.Name)
		if err != nil {
			continue
		}

		// 5. 计算可用资源
		nodeAvailCPU := allocatable.Cpu().DeepCopy()
		nodeAvailCPU.Sub(usedCPU)

		nodeAvailMemory := allocatable.Memory().DeepCopy()
		nodeAvailMemory.Sub(usedMemory)

		availGPU := int64(0)
		if gpuRes, ok := allocatable[corev1.ResourceName("nvidia.com/gpu")]; ok {
			availGPU = gpuRes.Value() - usedGPU
		}

		// 6. 判断是否满足
		cpuOK := nodeAvailCPU.Cmp(requiredCPU) >= 0
		memOK := nodeAvailMemory.Cmp(requiredMemory) >= 0
		gpuOK := availGPU >= requiredGPU

		if cpuOK && memOK && gpuOK {
			// 只要有一个节点满足 → 返回 true
			return true, nil
		}
	}

	// 所有节点都不够
	return false, errors.New("insufficient compute resources in cluster")

}

// getNodeAllocatedResources 计算节点上所有运行中 Pod 已占用的 CPU/Memory
func (c *K8sClient) getNodeAllocatedResources(ctx context.Context, nodeName string) (resource.Quantity, resource.Quantity, int64, error) {
	// 获取该节点上所有非完成态的 Pod
	pods, err := c.clientSet.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName + ",status.phase!=Succeeded,status.phase!=Failed",
	})
	if err != nil {
		return resource.Quantity{}, resource.Quantity{}, 0, err
	}

	totalCPU := resource.NewQuantity(0, resource.DecimalSI)
	totalMemory := resource.NewQuantity(0, resource.BinarySI)
	totalGPU := int64(0)

	// 累加所有容器的 Request 资源
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			req := container.Resources.Requests

			// CPU
			if cpu, ok := req[corev1.ResourceCPU]; ok {
				totalCPU.Add(cpu)
			}

			// Memory
			if mem, ok := req[corev1.ResourceMemory]; ok {
				totalMemory.Add(mem)
			}

			// GPU
			if gpu, ok := req[corev1.ResourceName("nvidia.com/gpu")]; ok {
				totalGPU += gpu.Value()
			}
		}
	}

	return *totalCPU, *totalMemory, totalGPU, nil
}

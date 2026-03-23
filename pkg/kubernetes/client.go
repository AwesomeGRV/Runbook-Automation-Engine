package kubernetes

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client provides Kubernetes operations
type Client interface {
	RestartDeployment(ctx context.Context, namespace, deployment string, options *RestartOptions) error
	ScaleDeployment(ctx context.Context, namespace, deployment string, replicas int32, options *ScaleOptions) error
	RollbackDeployment(ctx context.Context, namespace, deployment string, options *RollbackOptions) error
	GetDeployment(ctx context.Context, namespace, deployment string) (*appsv1.Deployment, error)
	GetPods(ctx context.Context, namespace, labelSelector string) ([]corev1.Pod, error)
	ExecuteCommand(ctx context.Context, namespace, pod, container string, command []string) (string, error)
	GetPodLogs(ctx context.Context, namespace, pod, container string, options *LogOptions) (string, error)
}

// RestartOptions contains options for deployment restart
type RestartOptions struct {
	WaitForRollout bool
	Timeout        time.Duration
	Annotations    map[string]string
}

// ScaleOptions contains options for deployment scaling
type ScaleOptions struct {
	WaitForRollout bool
	Timeout        time.Duration
	Strategy       string
}

// RollbackOptions contains options for deployment rollback
type RollbackOptions struct {
	Revision      int64
	WaitForRollout bool
	Timeout       time.Duration
}

// LogOptions contains options for getting pod logs
type LogOptions struct {
	Container    string
	Follow       bool
	Previous     bool
	SinceSeconds *int64
	TailLines    *int64
}

// client implements the Client interface
type client struct {
	clientset kubernetes.Interface
	config    *rest.Config
}

// NewClient creates a new Kubernetes client
func NewClient(config K8sConfig) (Client, error) {
	var restConfig *rest.Config
	var err error

	if config.InCluster {
		restConfig, err = rest.InClusterConfig()
	} else {
		restConfig, err = clientcmd.BuildConfigFromFlags("", config.Kubeconfig)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return &client{
		clientset: clientset,
		config:    restConfig,
	}, nil
}

// RestartDeployment restarts a Kubernetes deployment
func (c *client) RestartDeployment(ctx context.Context, namespace, deployment string, options *RestartOptions) error {
	if options == nil {
		options = &RestartOptions{
			WaitForRollout: true,
			Timeout:        5 * time.Minute,
		}
	}

	// Add restart annotation to trigger rollout
	restartAnnotation := fmt.Sprintf("kubectl.kubernetes.io/restartedAt=%s", time.Now().Format(time.RFC3339))
	
	patch := []byte(fmt.Sprintf(`{
		"spec": {
			"template": {
				"metadata": {
					"annotations": {
						"kubectl.kubernetes.io/restartedAt": "%s"
					}
				}
			}
		}
	}`, time.Now().Format(time.RFC3339)))

	_, err := c.clientset.AppsV1().Deployments(namespace).Patch(
		ctx,
		deployment,
		types.StrategicMergePatchType,
		patch,
		metav1.PatchOptions{},
	)

	if err != nil {
		return fmt.Errorf("failed to restart deployment %s/%s: %w", namespace, deployment, err)
	}

	if options.WaitForRollout {
		return c.waitForRollout(ctx, namespace, deployment, options.Timeout)
	}

	return nil
}

// ScaleDeployment scales a Kubernetes deployment
func (c *client) ScaleDeployment(ctx context.Context, namespace, deployment string, replicas int32, options *ScaleOptions) error {
	if options == nil {
		options = &ScaleOptions{
			WaitForRollout: true,
			Timeout:        5 * time.Minute,
		}
	}

	// Get current deployment
	deploy, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, deployment, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment %s/%s: %w", namespace, deployment, err)
	}

	// Create patch to scale
	scale := &appsv1.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment,
			Namespace: namespace,
		},
		Spec: appsv1.ScaleSpec{
			Replicas: replicas,
		},
	}

	_, err = c.clientset.AppsV1().Deployments(namespace).UpdateScale(ctx, deployment, scale, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale deployment %s/%s to %d replicas: %w", namespace, deployment, replicas, err)
	}

	if options.WaitForRollout {
		return c.waitForRollout(ctx, namespace, deployment, options.Timeout)
	}

	return nil
}

// RollbackDeployment rolls back a Kubernetes deployment
func (c *client) RollbackDeployment(ctx context.Context, namespace, deployment string, options *RollbackOptions) error {
	if options == nil {
		options = &RollbackOptions{
			WaitForRollout: true,
			Timeout:        5 * time.Minute,
		}
	}

	// Get deployment rollout history
	rolloutHistory, err := c.clientset.AppsV1().Deployments(namespace).GetRolloutHistory(ctx, deployment, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get rollout history for %s/%s: %w", namespace, deployment, err)
	}

	// Find the revision to rollback to
	var targetRevision *appsv1.ControllerRevision
	if options.Revision > 0 {
		// Find specific revision
		for _, rev := range rolloutHistory {
			if rev.Revision == options.Revision {
				targetRevision = rev
				break
			}
		}
	} else {
		// Rollback to previous revision
		if len(rolloutHistory) < 2 {
			return fmt.Errorf("no previous revision found for deployment %s/%s", namespace, deployment)
		}
		targetRevision = rolloutHistory[len(rolloutHistory)-2]
	}

	if targetRevision == nil {
		return fmt.Errorf("revision %d not found for deployment %s/%s", options.Revision, namespace, deployment)
	}

	// Create rollback annotation
	patch := []byte(fmt.Sprintf(`{
		"metadata": {
			"annotations": {
				"deployment.kubernetes.io/revision": "%d",
				"runbook-engine/rollback-at": "%s"
			}
		},
		"spec": {
			"template": %s
		}
	}`, targetRevision.Revision, time.Now().Format(time.RFC3339), targetRevision.Data.RawTemplate))

	_, err = c.clientset.AppsV1().Deployments(namespace).Patch(
		ctx,
		deployment,
		types.StrategicMergePatchType,
		patch,
		metav1.PatchOptions{},
	)

	if err != nil {
		return fmt.Errorf("failed to rollback deployment %s/%s: %w", namespace, deployment, err)
	}

	if options.WaitForRollout {
		return c.waitForRollout(ctx, namespace, deployment, options.Timeout)
	}

	return nil
}

// GetDeployment retrieves a deployment
func (c *client) GetDeployment(ctx context.Context, namespace, deployment string) (*appsv1.Deployment, error) {
	return c.clientset.AppsV1().Deployments(namespace).Get(ctx, deployment, metav1.GetOptions{})
}

// GetPods retrieves pods matching a label selector
func (c *client) GetPods(ctx context.Context, namespace, labelSelector string) ([]corev1.Pod, error) {
	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	return pods.Items, nil
}

// ExecuteCommand executes a command in a pod
func (c *client) ExecuteCommand(ctx context.Context, namespace, pod, container string, command []string) (string, error) {
	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(pod).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   command,
			Stdout:    true,
			Stderr:    true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return "", fmt.Errorf("failed to create executor: %w", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		return "", fmt.Errorf("command execution failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// GetPodLogs retrieves pod logs
func (c *client) GetPodLogs(ctx context.Context, namespace, pod, container string, options *LogOptions) (string, error) {
	logOptions := &corev1.PodLogOptions{
		Container: container,
	}

	if options != nil {
		logOptions.Follow = options.Follow
		logOptions.Previous = options.Previous
		logOptions.SinceSeconds = options.SinceSeconds
		logOptions.TailLines = options.TailLines
	}

	req := c.clientset.CoreV1().Pods(namespace).GetLogs(pod, logOptions)
	logs, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %w", err)
	}
	defer logs.Close()

	data, err := io.ReadAll(logs)
	if err != nil {
		return "", fmt.Errorf("failed to read logs: %w", err)
	}

	return string(data), nil
}

// waitForRollout waits for a deployment rollout to complete
func (c *client) waitForRollout(ctx context.Context, namespace, deployment string, timeout time.Duration) error {
	return wait.PollImmediate(time.Second, timeout, func() (bool, error) {
		deploy, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, deployment, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}

		// Check if deployment is rolled out
		if deploy.Generation <= deploy.Status.ObservedGeneration {
			cond := getDeploymentCondition(deploy.Status, appsv1.DeploymentProgressing)
			if cond != nil && cond.Reason == "NewReplicaSetAvailable" {
				return true, nil
			}
		}

		return false, nil
	})
}

// getDeploymentCondition gets the deployment condition by type
func getDeploymentCondition(status appsv1.DeploymentStatus, condType appsv1.DeploymentConditionType) *appsv1.DeploymentCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}

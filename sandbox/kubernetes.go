// Runs sandbox jobs as Kubernetes Jobs: an unprivileged pod with no service account token, its
// secrets in a Secret the Job owns, and egress limited by the gavel-runner-egress NetworkPolicy.
package sandbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/flanksource/commons-db/connection"
	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/types"
	"github.com/flanksource/commons/logger"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	metaapply "k8s.io/client-go/applyconfigurations/meta/v1"
	networkingapply "k8s.io/client-go/applyconfigurations/networking/v1"
	"k8s.io/client-go/kubernetes"
)

// runnerContainer is the name of the job pod's one container.
const runnerContainer = "runner"

// runnerLabels select every sandbox job pod, for the NetworkPolicy.
var runnerLabels = map[string]string{"app.kubernetes.io/name": "gavel-runner"}

// KubernetesOptions configures the Kubernetes executor.
type KubernetesOptions struct {
	// Kubeconfig is a kubeconfig path or its contents; empty uses $KUBECONFIG, ~/.kube/config or
	// the in-cluster service account.
	Kubeconfig string
	Namespace  string
	// ServePodSelector and ServePort let jobs reach gavel serve through the NetworkPolicy.
	ServePodSelector map[string]string
	ServePort        int
	// ExtraDeny adds CIDRs jobs must not reach, such as cluster pod and service ranges outside
	// RFC 1918.
	ExtraDeny []string
}

type kubernetesExecutor struct {
	client  kubernetes.Interface
	options KubernetesOptions
}

// NewKubernetes returns an executor running jobs in options.Namespace.
func NewKubernetes(ctx context.Context, options KubernetesOptions) (*kubernetesExecutor, error) {
	kubeconfig := connection.KubeconfigConnection{}
	if options.Kubeconfig != "" {
		kubeconfig.Kubeconfig = &types.EnvVar{ValueStatic: options.Kubeconfig}
	}
	client, _, err := kubeconfig.Populate(dbcontext.NewContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("connect to Kubernetes for sandbox jobs: %w", err)
	}
	logger.Warnf("sandbox jobs rely on the gavel-runner-egress NetworkPolicy, which only a CNI that enforces NetworkPolicy (such as Calico or Cilium) applies")
	return newKubernetes(client, options), nil
}

func newKubernetes(client kubernetes.Interface, options KubernetesOptions) *kubernetesExecutor {
	return &kubernetesExecutor{client: client, options: options}
}

func (e *kubernetesExecutor) Name() string { return "kubernetes" }

func (e *kubernetesExecutor) Submit(ctx context.Context, job Job) (Handle, error) {
	spec, err := e.jobSpec(job)
	if err != nil {
		return nil, err
	}
	created, err := e.client.BatchV1().Jobs(e.options.Namespace).Create(ctx, spec, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("create the Kubernetes Job for sandbox run %s: %w", job.RunID, err)
	}
	data := map[string][]byte{}
	for key, value := range job.SecretEnv {
		data[key] = []byte(value)
	}
	// Created after the Job so the Job can own it: deleting the Job deletes the Secret.
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: created.Name, Namespace: e.options.Namespace, Labels: created.Labels,
			OwnerReferences: []metav1.OwnerReference{{APIVersion: "batch/v1", Kind: "Job", Name: created.Name, UID: created.UID}},
		},
		Data: data,
	}
	if _, err := e.client.CoreV1().Secrets(e.options.Namespace).Create(ctx, secret, metav1.CreateOptions{}); err != nil {
		cancelErr := e.deleteJob(context.WithoutCancel(ctx), created.Name)
		return nil, errors.Join(fmt.Errorf("create the Secret for sandbox run %s: %w", job.RunID, err), cancelErr)
	}
	return &kubernetesHandle{executor: e, name: created.Name, run: job.RunID.String()}, nil
}

func (e *kubernetesExecutor) jobSpec(job Job) (*batchv1.Job, error) {
	if job.Image == "" {
		return nil, fmt.Errorf("sandbox job has no image")
	}
	limits := job.Limits.withDefaults()
	cpu, err := resource.ParseQuantity(limits.CPU)
	if err != nil || cpu.Sign() <= 0 {
		return nil, fmt.Errorf("sandbox job's CPU limit %q is not a positive quantity such as 2 or 500m", limits.CPU)
	}
	memory, err := resource.ParseQuantity(limits.Memory)
	if err != nil || memory.Sign() <= 0 {
		return nil, fmt.Errorf("sandbox job's memory limit %q is not a positive quantity such as 4Gi", limits.Memory)
	}
	// sortedEnv validates the names and values.
	if _, err := sortedEnv(job.Env); err != nil {
		return nil, err
	}
	env := make([]corev1.EnvVar, 0, len(job.Env))
	for _, key := range sortedKeys(job.Env) {
		env = append(env, corev1.EnvVar{Name: key, Value: job.Env[key]})
	}
	labels := map[string]string{
		"app.kubernetes.io/name": runnerLabels["app.kubernetes.io/name"],
		"gavel.dev/org":          job.Scope.OrgID.String(),
		"gavel.dev/run":          job.RunID.String(),
	}
	name := jobName(job.RunID)
	resources := corev1.ResourceList{corev1.ResourceCPU: cpu, corev1.ResourceMemory: memory}
	workspaceLimit := resource.MustParse("10Gi")
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: e.options.Namespace, Labels: labels},
		Spec: batchv1.JobSpec{
			BackoffLimit:            ptr(int32(0)),
			TTLSecondsAfterFinished: ptr(int32(3600)),
			ActiveDeadlineSeconds:   ptr(int64(limits.Timeout / time.Second)),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					RestartPolicy:                corev1.RestartPolicyNever,
					AutomountServiceAccountToken: ptr(false),
					EnableServiceLinks:           ptr(false),
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot:   ptr(true),
						RunAsUser:      ptr(int64(1000)),
						FSGroup:        ptr(int64(1000)),
						SeccompProfile: &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault},
					},
					Containers: []corev1.Container{{
						Name:    runnerContainer,
						Image:   job.Image,
						Command: job.Command,
						Env:     env,
						EnvFrom: []corev1.EnvFromSource{{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: name}}}},
						SecurityContext: &corev1.SecurityContext{
							AllowPrivilegeEscalation: ptr(false),
							ReadOnlyRootFilesystem:   ptr(true),
							Capabilities:             &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
						},
						Resources: corev1.ResourceRequirements{Requests: resources, Limits: resources},
						VolumeMounts: []corev1.VolumeMount{
							{Name: "workspace", MountPath: "/workspace"},
							{Name: "tmp", MountPath: "/tmp"},
							{Name: "home", MountPath: "/home/gavel"},
						},
					}},
					Volumes: []corev1.Volume{
						{Name: "workspace", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{SizeLimit: &workspaceLimit}}},
						{Name: "tmp", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
						{Name: "home", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
					},
				},
			},
		},
	}, nil
}

func (e *kubernetesExecutor) deleteJob(ctx context.Context, name string) error {
	background := metav1.DeletePropagationBackground
	err := e.client.BatchV1().Jobs(e.options.Namespace).Delete(ctx, name, metav1.DeleteOptions{PropagationPolicy: &background})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("delete Kubernetes Job %s: %w", name, err)
	}
	return e.deleteSecret(ctx, name)
}

func (e *kubernetesExecutor) deleteSecret(ctx context.Context, name string) error {
	err := e.client.CoreV1().Secrets(e.options.Namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("delete Secret %s: %w", name, err)
	}
	return nil
}

// EnsureNetworkPolicy applies gavel-runner-egress: no ingress to job pods, and egress only to the
// internet outside private ranges, cluster DNS, and gavel serve.
func (e *kubernetesExecutor) EnsureNetworkPolicy(ctx context.Context) error {
	except := append(append([]string{}, privateRanges...), e.options.ExtraDeny...)
	udp, tcp := corev1.ProtocolUDP, corev1.ProtocolTCP
	dnsPort, servePort := intstr.FromInt32(53), intstr.FromInt32(int32(e.options.ServePort))
	policy := networkingapply.NetworkPolicy("gavel-runner-egress", e.options.Namespace).
		WithSpec(networkingapply.NetworkPolicySpec().
			WithPodSelector(metaapply.LabelSelector().WithMatchLabels(runnerLabels)).
			WithPolicyTypes("Ingress", "Egress").
			WithEgress(
				networkingapply.NetworkPolicyEgressRule().WithTo(networkingapply.NetworkPolicyPeer().
					WithIPBlock(networkingapply.IPBlock().WithCIDR("0.0.0.0/0").WithExcept(except...))),
				networkingapply.NetworkPolicyEgressRule().
					WithTo(networkingapply.NetworkPolicyPeer().
						WithNamespaceSelector(metaapply.LabelSelector().WithMatchLabels(map[string]string{"kubernetes.io/metadata.name": "kube-system"})).
						WithPodSelector(metaapply.LabelSelector().WithMatchLabels(map[string]string{"k8s-app": "kube-dns"}))).
					WithPorts(
						networkingapply.NetworkPolicyPort().WithProtocol(udp).WithPort(dnsPort),
						networkingapply.NetworkPolicyPort().WithProtocol(tcp).WithPort(dnsPort)),
				networkingapply.NetworkPolicyEgressRule().
					WithTo(networkingapply.NetworkPolicyPeer().WithPodSelector(metaapply.LabelSelector().WithMatchLabels(e.options.ServePodSelector))).
					WithPorts(networkingapply.NetworkPolicyPort().WithProtocol(tcp).WithPort(servePort)),
			))
	if _, err := e.client.NetworkingV1().NetworkPolicies(e.options.Namespace).Apply(ctx, policy, metav1.ApplyOptions{FieldManager: "gavel", Force: true}); err != nil {
		return fmt.Errorf("apply the gavel-runner-egress NetworkPolicy in %s: %w", e.options.Namespace, err)
	}
	return nil
}

type kubernetesHandle struct {
	executor *kubernetesExecutor
	name     string
	run      string
}

func (h *kubernetesHandle) Ref() string { return h.name }

// pod waits for the job's pod to start or finish.
func (h *kubernetesHandle) pod(ctx context.Context) (*corev1.Pod, error) {
	for {
		pods, err := h.executor.client.CoreV1().Pods(h.executor.options.Namespace).List(ctx, metav1.ListOptions{LabelSelector: "gavel.dev/run=" + h.run})
		if err != nil {
			return nil, fmt.Errorf("find the pod of sandbox job %s: %w", h.name, err)
		}
		for i := range pods.Items {
			if phase := pods.Items[i].Status.Phase; phase != corev1.PodPending && phase != "" {
				return &pods.Items[i], nil
			}
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("wait for the pod of sandbox job %s: %w", h.name, ctx.Err())
		case <-time.After(time.Second):
		}
	}
}

// Logs streams the runner container's raw output, following it until the container exits.
func (h *kubernetesHandle) Logs(ctx context.Context) (io.ReadCloser, error) {
	pod, err := h.pod(ctx)
	if err != nil {
		return nil, err
	}
	stream, err := h.executor.client.CoreV1().Pods(pod.Namespace).
		GetLogs(pod.Name, &corev1.PodLogOptions{Container: runnerContainer, Follow: true}).Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("follow the logs of sandbox job %s: %w", h.name, err)
	}
	return stream, nil
}

func (h *kubernetesHandle) Wait(ctx context.Context) (Result, error) {
	jobs := h.executor.client.BatchV1().Jobs(h.executor.options.Namespace)
	for {
		job, err := jobs.Get(ctx, h.name, metav1.GetOptions{})
		if err != nil {
			return Result{}, fmt.Errorf("wait for sandbox job %s: %w", h.name, err)
		}
		if finished(job) {
			break
		}
		select {
		case <-ctx.Done():
			return Result{}, fmt.Errorf("wait for sandbox job %s: %w", h.name, ctx.Err())
		case <-time.After(time.Second):
		}
	}
	defer func() {
		if err := h.executor.deleteSecret(context.WithoutCancel(ctx), h.name); err != nil {
			logger.Warnf("%v", err)
		}
	}()
	pod, err := h.pod(ctx)
	if err != nil {
		return Result{}, err
	}
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name == runnerContainer && status.State.Terminated != nil {
			return Result{ExitCode: int(status.State.Terminated.ExitCode)}, nil
		}
	}
	return Result{}, fmt.Errorf("sandbox job %s finished, but its pod %s reports no exit code (deadline exceeded or evicted?)", h.name, pod.Name)
}

func finished(job *batchv1.Job) bool {
	for _, condition := range job.Status.Conditions {
		if (condition.Type == batchv1.JobComplete || condition.Type == batchv1.JobFailed) && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func (h *kubernetesHandle) Cancel(ctx context.Context) error {
	return h.executor.deleteJob(ctx, h.name)
}

func ptr[T any](value T) *T { return &value }

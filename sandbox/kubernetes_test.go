// Tests the Kubernetes executor's spec generation against a fake clientset: the Job and Secret a
// sandbox job becomes, waiting, cancelling and the egress NetworkPolicy. It tests our specs, not
// Kubernetes.
package sandbox

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const testNamespace = "gavel-sandbox"

func kubernetesJob() Job {
	run := uuid.MustParse("6f1c2b8e-1111-4222-8333-444455556666")
	return Job{
		RunID:     run,
		Scope:     tenant.Scope{OrgID: uuid.MustParse("0a0b0c0d-1111-4222-8333-444455556666")},
		Image:     "ghcr.io/flanksource/gavel-runner:1",
		Command:   []string{"gavel", "runner", "exec", "--run", run.String()},
		Env:       map[string]string{"GAVEL_SERVER": "https://gavel.example", "CI": "1"},
		SecretEnv: map[string]string{"GAVEL_RUN_TOKEN": "secret-token"},
		Limits:    Limits{CPU: "2", Memory: "4Gi", Timeout: time.Hour},
	}
}

func TestKubernetesSubmitCreatesAHardenedJobAndItsSecret(t *testing.T) {
	client := fake.NewClientset()
	executor := newKubernetes(client, KubernetesOptions{Namespace: testNamespace})
	ctx := context.Background()
	job := kubernetesJob()

	handle, err := executor.Submit(ctx, job)
	require.NoError(t, err)
	name := "gavel-run-" + job.RunID.String()
	assert.Equal(t, name, handle.Ref())

	created, err := client.BatchV1().Jobs(testNamespace).Get(ctx, name, metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"app.kubernetes.io/name": "gavel-runner",
		"gavel.dev/org":          job.Scope.OrgID.String(),
		"gavel.dev/run":          job.RunID.String(),
	}, created.Labels)
	spec := created.Spec
	assert.Equal(t, int32(0), *spec.BackoffLimit)
	assert.Equal(t, int32(3600), *spec.TTLSecondsAfterFinished)
	assert.Equal(t, int64(3600), *spec.ActiveDeadlineSeconds)
	assert.Equal(t, created.Labels, spec.Template.Labels, "the pod carries the labels the NetworkPolicy selects")

	pod := spec.Template.Spec
	assert.Equal(t, corev1.RestartPolicyNever, pod.RestartPolicy)
	assert.False(t, *pod.AutomountServiceAccountToken)
	assert.False(t, *pod.EnableServiceLinks)
	assert.True(t, *pod.SecurityContext.RunAsNonRoot)
	assert.Equal(t, int64(1000), *pod.SecurityContext.RunAsUser)
	assert.Equal(t, int64(1000), *pod.SecurityContext.FSGroup)
	assert.Equal(t, corev1.SeccompProfileTypeRuntimeDefault, pod.SecurityContext.SeccompProfile.Type)

	require.Len(t, pod.Containers, 1)
	container := pod.Containers[0]
	assert.Equal(t, "runner", container.Name)
	assert.Equal(t, job.Image, container.Image)
	assert.Equal(t, job.Command, container.Command)
	assert.False(t, *container.SecurityContext.AllowPrivilegeEscalation)
	assert.True(t, *container.SecurityContext.ReadOnlyRootFilesystem)
	assert.Equal(t, []corev1.Capability{"ALL"}, container.SecurityContext.Capabilities.Drop)
	want := corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("2"), corev1.ResourceMemory: resource.MustParse("4Gi")}
	assert.True(t, want.Cpu().Equal(*container.Resources.Limits.Cpu()) && want.Memory().Equal(*container.Resources.Limits.Memory()))
	assert.True(t, want.Cpu().Equal(*container.Resources.Requests.Cpu()) && want.Memory().Equal(*container.Resources.Requests.Memory()), "requests equal limits")
	assert.Equal(t, []corev1.EnvVar{{Name: "CI", Value: "1"}, {Name: "GAVEL_SERVER", Value: "https://gavel.example"}}, container.Env)
	require.Len(t, container.EnvFrom, 1)
	assert.Equal(t, name, container.EnvFrom[0].SecretRef.Name)
	for _, env := range container.Env {
		assert.NotEqual(t, "secret-token", env.Value, "secrets never appear in the pod spec")
	}

	mounts := map[string]string{}
	for _, mount := range container.VolumeMounts {
		mounts[mount.MountPath] = mount.Name
	}
	assert.Equal(t, map[string]string{"/workspace": "workspace", "/tmp": "tmp", "/home/gavel": "home"}, mounts)
	for _, volume := range pod.Volumes {
		require.NotNil(t, volume.EmptyDir, volume.Name)
		if volume.Name == "workspace" {
			assert.True(t, resource.MustParse("10Gi").Equal(*volume.EmptyDir.SizeLimit))
		}
	}

	secret, err := client.CoreV1().Secrets(testNamespace).Get(ctx, name, metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, "secret-token", string(secret.Data["GAVEL_RUN_TOKEN"]))
	require.Len(t, secret.OwnerReferences, 1, "the Job owns its Secret, so deleting the Job deletes it")
	assert.Equal(t, "Job", secret.OwnerReferences[0].Kind)
	assert.Equal(t, name, secret.OwnerReferences[0].Name)
}

func TestKubernetesWaitReadsTheExitCodeAndDeletesTheSecret(t *testing.T) {
	client := fake.NewClientset()
	executor := newKubernetes(client, KubernetesOptions{Namespace: testNamespace})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	job := kubernetesJob()
	handle, err := executor.Submit(ctx, job)
	require.NoError(t, err)

	// The job controller and kubelet, played by the test: a finished pod and a failed Job.
	_, err = client.CoreV1().Pods(testNamespace).Create(ctx, &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: handle.Ref() + "-abcde", Namespace: testNamespace, Labels: map[string]string{"gavel.dev/run": job.RunID.String()}},
		Status: corev1.PodStatus{Phase: corev1.PodFailed, ContainerStatuses: []corev1.ContainerStatus{{
			Name: "runner", State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 3}},
		}}},
	}, metav1.CreateOptions{})
	require.NoError(t, err)
	created, err := client.BatchV1().Jobs(testNamespace).Get(ctx, handle.Ref(), metav1.GetOptions{})
	require.NoError(t, err)
	created.Status.Conditions = []batchv1.JobCondition{{Type: batchv1.JobFailed, Status: corev1.ConditionTrue}}
	_, err = client.BatchV1().Jobs(testNamespace).UpdateStatus(ctx, created, metav1.UpdateOptions{})
	require.NoError(t, err)

	result, err := handle.Wait(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, result.ExitCode)
	_, err = client.CoreV1().Secrets(testNamespace).Get(ctx, handle.Ref(), metav1.GetOptions{})
	assert.Error(t, err, "Wait deletes the Secret")
}

func TestKubernetesLogsFollowTheRunnerContainer(t *testing.T) {
	client := fake.NewClientset()
	executor := newKubernetes(client, KubernetesOptions{Namespace: testNamespace})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	job := kubernetesJob()
	handle, err := executor.Submit(ctx, job)
	require.NoError(t, err)
	_, err = client.CoreV1().Pods(testNamespace).Create(ctx, &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: handle.Ref() + "-abcde", Namespace: testNamespace, Labels: map[string]string{"gavel.dev/run": job.RunID.String()}},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "runner"}}},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	logs, err := handle.Logs(ctx)
	require.NoError(t, err)
	data, err := io.ReadAll(logs)
	require.NoError(t, err)
	require.NoError(t, logs.Close())
	assert.Contains(t, string(data), "fake logs", "the fake clientset's log stream")
}

func TestKubernetesCancelDeletesTheJob(t *testing.T) {
	client := fake.NewClientset()
	executor := newKubernetes(client, KubernetesOptions{Namespace: testNamespace})
	ctx := context.Background()
	handle, err := executor.Submit(ctx, kubernetesJob())
	require.NoError(t, err)

	require.NoError(t, handle.Cancel(ctx))
	_, err = client.BatchV1().Jobs(testNamespace).Get(ctx, handle.Ref(), metav1.GetOptions{})
	assert.Error(t, err)
	_, err = client.CoreV1().Secrets(testNamespace).Get(ctx, handle.Ref(), metav1.GetOptions{})
	assert.Error(t, err)
}

func TestKubernetesNetworkPolicy(t *testing.T) {
	client := fake.NewClientset()
	executor := newKubernetes(client, KubernetesOptions{
		Namespace: testNamespace, ServePodSelector: map[string]string{"app.kubernetes.io/name": "gavel"}, ServePort: 9092,
		ExtraDeny: []string{"203.0.113.0/24"},
	})
	ctx := context.Background()
	require.NoError(t, executor.EnsureNetworkPolicy(ctx))
	require.NoError(t, executor.EnsureNetworkPolicy(ctx), "applying again is a no-op")

	policy, err := client.NetworkingV1().NetworkPolicies(testNamespace).Get(ctx, "gavel-runner-egress", metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"app.kubernetes.io/name": "gavel-runner"}, policy.Spec.PodSelector.MatchLabels)
	assert.ElementsMatch(t, []networkingv1.PolicyType{networkingv1.PolicyTypeIngress, networkingv1.PolicyTypeEgress}, policy.Spec.PolicyTypes)
	assert.Empty(t, policy.Spec.Ingress, "no ingress")

	require.Len(t, policy.Spec.Egress, 3)
	internet := policy.Spec.Egress[0].To[0].IPBlock
	assert.Equal(t, "0.0.0.0/0", internet.CIDR)
	assert.ElementsMatch(t, append(append([]string{}, privateRanges...), "203.0.113.0/24"), internet.Except)

	dns := policy.Spec.Egress[1]
	assert.Equal(t, map[string]string{"kubernetes.io/metadata.name": "kube-system"}, dns.To[0].NamespaceSelector.MatchLabels)
	assert.Equal(t, map[string]string{"k8s-app": "kube-dns"}, dns.To[0].PodSelector.MatchLabels)
	require.Len(t, dns.Ports, 2)
	assert.Equal(t, int32(53), dns.Ports[0].Port.IntVal)

	serve := policy.Spec.Egress[2]
	assert.Equal(t, map[string]string{"app.kubernetes.io/name": "gavel"}, serve.To[0].PodSelector.MatchLabels)
	assert.Equal(t, int32(9092), serve.Ports[0].Port.IntVal)
}

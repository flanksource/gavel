// Renders the gavel Helm chart with `helm template` and checks the hosted server it describes:
// Clerk settings, the bundled or brought Postgres, blob storage, the substrate sandbox and SSH.
package chart

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/yaml"
)

// clerkValues are the settings every hosted install needs.
var clerkValues = []string{
	"publicURL=https://gavel.example.com",
	"clerk.issuer=https://clerk.example.com",
	"clerk.publishableKey=pk_test_example",
	"clerk.clientId=client_example",
	"clerk.secretKey=sk_test_example",
}

// manifests are a render's objects, by kind and then name.
type manifests map[string]map[string]runtime.Object

// helmTemplate renders the chart as release gavel in namespace gavel with values set by --set,
// or passed as they are when they are flags such as --set-json, returning helm's output and its
// error.
func helmTemplate(t *testing.T, values ...string) (string, error) {
	t.Helper()
	args := []string{"template", "gavel", ".", "--namespace", "gavel"}
	for _, value := range values {
		if strings.HasPrefix(value, "--") {
			args = append(args, value)
		} else {
			args = append(args, "--set", value)
		}
	}
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("helm", args...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", &renderError{err: err, stderr: stderr.String()}
	}
	return stdout.String(), nil
}

type renderError struct {
	err    error
	stderr string
}

func (e *renderError) Error() string { return e.err.Error() + ": " + e.stderr }

// render renders the chart with clerkValues and values, decoding every object.
func render(t *testing.T, values ...string) manifests {
	t.Helper()
	out, err := helmTemplate(t, append(append([]string{}, clerkValues...), values...)...)
	require.NoError(t, err)
	objects := manifests{}
	for doc := range strings.SplitSeq(out, "\n---") {
		if strings.TrimSpace(strings.TrimPrefix(doc, "---")) == "" {
			continue
		}
		object, kind, err := scheme.Codecs.UniversalDeserializer().Decode([]byte(doc), nil, nil)
		require.NoError(t, err, doc)
		name, err := metaName(doc)
		require.NoError(t, err)
		if objects[kind.Kind] == nil {
			objects[kind.Kind] = map[string]runtime.Object{}
		}
		objects[kind.Kind][name] = object
	}
	return objects
}

func metaName(doc string) (string, error) {
	var object struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
	}
	err := yaml.Unmarshal([]byte(doc), &object)
	return object.Metadata.Name, err
}

func (m manifests) deployment(t *testing.T) *appsv1.Deployment {
	t.Helper()
	require.Contains(t, m["Deployment"], "gavel")
	return m["Deployment"]["gavel"].(*appsv1.Deployment)
}

func (m manifests) secret(t *testing.T, name string) *corev1.Secret {
	t.Helper()
	require.Contains(t, m["Secret"], name)
	return m["Secret"][name].(*corev1.Secret)
}

// secretValue is a Secret's value for key, whether written as data or stringData.
func secretValue(secret *corev1.Secret, key string) string {
	if value, ok := secret.StringData[key]; ok {
		return value
	}
	return string(secret.Data[key])
}

func serveContainer(t *testing.T, m manifests) corev1.Container {
	t.Helper()
	containers := m.deployment(t).Spec.Template.Spec.Containers
	require.Len(t, containers, 1)
	return containers[0]
}

// env is the container's environment: literal values as they are, Secret references as
// "secret:<name>/<key>".
func env(container corev1.Container) map[string]string {
	values := map[string]string{}
	for _, variable := range container.Env {
		switch {
		case variable.ValueFrom != nil && variable.ValueFrom.SecretKeyRef != nil:
			ref := variable.ValueFrom.SecretKeyRef
			values[variable.Name] = "secret:" + ref.Name + "/" + ref.Key
		default:
			values[variable.Name] = variable.Value
		}
	}
	return values
}

func mountAt(container corev1.Container, path string) (corev1.VolumeMount, bool) {
	for _, mount := range container.VolumeMounts {
		if mount.MountPath == path {
			return mount, true
		}
	}
	return corev1.VolumeMount{}, false
}

func volume(t *testing.T, m manifests, name string) corev1.Volume {
	t.Helper()
	for _, v := range m.deployment(t).Spec.Template.Spec.Volumes {
		if v.Name == name {
			return v
		}
	}
	require.Failf(t, "no volume", "the gavel pod has no volume %q", name)
	return corev1.Volume{}
}

func requireHelm(t *testing.T) {
	t.Helper()
	_, err := exec.LookPath("helm")
	require.NoError(t, err, "the chart tests render with the helm CLI")
}

func TestChartRefusesAHostedInstallWithoutItsClerkSettings(t *testing.T) {
	requireHelm(t)
	for _, missing := range []string{"publicURL", "clerk.issuer", "clerk.publishableKey", "clerk.clientId", "clerk.secretKey"} {
		var values []string
		for _, value := range clerkValues {
			if !strings.HasPrefix(value, missing+"=") {
				values = append(values, value)
			}
		}
		_, err := helmTemplate(t, values...)
		require.Error(t, err, missing)
		assert.Contains(t, err.Error(), missing+" is required", missing)
	}

	_, err := helmTemplate(t, append(clerkValues[:len(clerkValues)-1:len(clerkValues)-1], "secret.create=false")...)
	assert.NoError(t, err, "clerk.secretKey comes from the brought Secret when the chart creates none")
}

func TestChartRunsOneHostedServerInClerkMode(t *testing.T) {
	requireHelm(t)
	m := render(t)
	deployment := m.deployment(t)
	assert.EqualValues(t, 1, *deployment.Spec.Replicas, "agent-run relays and pollers live in one process")
	assert.Equal(t, appsv1.RecreateDeploymentStrategyType, deployment.Spec.Strategy.Type, "the data volume is ReadWriteOnce")
	assert.Equal(t, "gavel", deployment.Spec.Template.Labels["app.kubernetes.io/name"])

	container := serveContainer(t, m)
	assert.Equal(t, "ghcr.io/flanksource/gavel:v0.0.57", container.Image)
	assert.Equal(t, []string{
		"serve", "--auth", "clerk", "--public-url", "https://gavel.example.com",
		"--addr", "0.0.0.0", "--port", "9092", "--data-dir", "/var/lib/gavel",
		"--properties-file", "/etc/gavel/gavel.properties",
	}, container.Args)
	assert.Equal(t, map[string]string{
		"HOME":                        "/home/gavel",
		"GAVEL_CLERK_ISSUER":          "https://clerk.example.com",
		"GAVEL_CLERK_PUBLISHABLE_KEY": "pk_test_example",
		"GAVEL_CLERK_CLIENT_ID":       "client_example",
		"CLERK_SECRET_KEY":            "secret:gavel/CLERK_SECRET_KEY",
		"GAVEL_SECRET_KEY":            "secret:gavel/GAVEL_SECRET_KEY",
		"GAVEL_DB_DSN":                "secret:gavel-postgres/DB_URL",
	}, env(container))
	require.NotNil(t, container.StartupProbe)
	assert.Equal(t, "/api/auth/config", container.StartupProbe.HTTPGet.Path, "the probe Clerk lets through unauthenticated")
	assert.Equal(t, "/api/auth/config", container.ReadinessProbe.HTTPGet.Path)
	assert.Equal(t, "/api/auth/config", container.LivenessProbe.HTTPGet.Path)

	data, ok := mountAt(container, "/var/lib/gavel")
	require.True(t, ok, "repositories, mirrors and workspaces persist")
	assert.Equal(t, "gavel", volume(t, m, data.Name).PersistentVolumeClaim.ClaimName)
	require.Contains(t, m["PersistentVolumeClaim"], "gavel")
	for _, path := range []string{"/home/gavel", "/tmp"} {
		_, ok := mountAt(container, path)
		assert.True(t, ok, "%s is writable on a read-only root filesystem", path)
	}
	assert.True(t, *container.SecurityContext.ReadOnlyRootFilesystem)
	assert.True(t, *deployment.Spec.Template.Spec.SecurityContext.RunAsNonRoot)

	secret := m.secret(t, "gavel")
	assert.Equal(t, "sk_test_example", secretValue(secret, "CLERK_SECRET_KEY"))
	key, err := base64.StdEncoding.DecodeString(secretValue(secret, "GAVEL_SECRET_KEY"))
	require.NoError(t, err, "GAVEL_SECRET_KEY is base64")
	assert.GreaterOrEqual(t, len(key), 32, "GAVEL_SECRET_KEY decodes to at least 32 bytes")

	service := m["Service"]["gavel"].(*corev1.Service)
	assert.Equal(t, corev1.ServiceTypeClusterIP, service.Spec.Type)
	require.Len(t, service.Spec.Ports, 1)
	assert.EqualValues(t, 9092, service.Spec.Ports[0].Port)
	assert.Equal(t, "gavel", service.Spec.Selector["app.kubernetes.io/name"])
	assert.NotContains(t, m["Service"], "gavel-ssh", "the push server is off unless enabled")
}

func TestChartRunsItsOwnPostgres(t *testing.T) {
	requireHelm(t)
	m := render(t)
	require.Contains(t, m["StatefulSet"], "gavel-postgresql")
	postgres := m["StatefulSet"]["gavel-postgresql"].(*appsv1.StatefulSet)
	assert.EqualValues(t, 1, *postgres.Spec.Replicas)
	require.Len(t, postgres.Spec.VolumeClaimTemplates, 1)
	assert.Equal(t, "20Gi", postgres.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests.Storage().String())
	container := postgres.Spec.Template.Spec.Containers[0]
	assert.Contains(t, container.Args, "--max_connections=200")
	require.Len(t, container.EnvFrom, 1)
	assert.Equal(t, "gavel-postgres", container.EnvFrom[0].SecretRef.Name)
	assert.Equal(t, "gavel", env(container)["POSTGRES_DB"])

	require.Contains(t, m["Service"], "gavel-postgresql")
	assert.EqualValues(t, 5432, m["Service"]["gavel-postgresql"].(*corev1.Service).Spec.Ports[0].Port)

	secret := m.secret(t, "gavel-postgres")
	password := secretValue(secret, "POSTGRES_PASSWORD")
	assert.Len(t, password, 32)
	assert.Equal(t, "postgres://postgres:"+password+"@gavel-postgresql.gavel.svc.cluster.local:5432/gavel?sslmode=disable",
		secretValue(secret, "DB_URL"))
}

func TestChartUsesABroughtPostgres(t *testing.T) {
	requireHelm(t)
	m := render(t, "db.create=false", "db.secretKeyRef.name=shared-db", "db.secretKeyRef.key=GAVEL_DSN")
	assert.Empty(t, m["StatefulSet"])
	assert.NotContains(t, m["Secret"], "gavel-postgres")
	assert.NotContains(t, m["Service"], "gavel-postgresql")
	assert.Equal(t, "secret:shared-db/GAVEL_DSN", env(serveContainer(t, m))["GAVEL_DB_DSN"])
}

// gavelYAML is the ~/.gavel.yaml the chart writes for the server.
func gavelYAML(t *testing.T, m manifests) map[string]any {
	t.Helper()
	container := serveContainer(t, m)
	mount, ok := mountAt(container, "/home/gavel/.gavel.yaml")
	require.True(t, ok, "the server reads its storage from ~/.gavel.yaml")
	assert.Equal(t, ".gavel.yaml", mount.SubPath)
	secretName := volume(t, m, mount.Name).Secret.SecretName
	var config map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(secretValue(m.secret(t, secretName), ".gavel.yaml")), &config))
	return config
}

func TestChartStoresBlobsOnTheDataVolumeByDefault(t *testing.T) {
	requireHelm(t)
	assert.Equal(t, map[string]any{"storage": map[string]any{"type": "local", "path": "/var/lib/gavel/blobs"}},
		gavelYAML(t, render(t)))
}

func TestChartStoresBlobsInAnS3Bucket(t *testing.T) {
	requireHelm(t)
	m := render(t,
		"storage.type=s3", "storage.bucket=gavel-blobs", "storage.prefix=prod",
		"storage.s3.region=eu-west-1", "storage.s3.endpoint=https://minio.example.com", "storage.s3.usePathStyle=true",
		"storage.s3.credentialsSecret=gavel-s3",
	)
	assert.Equal(t, map[string]any{"storage": map[string]any{
		"type": "s3", "bucket": "gavel-blobs", "prefix": "prod",
		"s3": map[string]any{"region": "eu-west-1", "endpoint": "https://minio.example.com", "usePathStyle": true},
	}}, gavelYAML(t, m))
	values := env(serveContainer(t, m))
	assert.Equal(t, "secret:gavel-s3/AWS_ACCESS_KEY_ID", values["AWS_ACCESS_KEY_ID"])
	assert.Equal(t, "secret:gavel-s3/AWS_SECRET_ACCESS_KEY", values["AWS_SECRET_ACCESS_KEY"])

	values = env(serveContainer(t, render(t, "storage.type=s3", "storage.bucket=gavel-blobs")))
	assert.NotContains(t, values, "AWS_ACCESS_KEY_ID", "without a Secret the AWS default chain applies, such as IRSA")
}

func TestChartStoresBlobsInAGCSBucket(t *testing.T) {
	requireHelm(t)
	m := render(t, "storage.type=gcs", "storage.bucket=gavel-blobs", `--set-json=storage.gcs.credentials="{\"type\":\"service_account\"}"`)
	assert.Equal(t, map[string]any{"storage": map[string]any{
		"type": "gcs", "bucket": "gavel-blobs", "gcs": map[string]any{"credentials": `{"type":"service_account"}`},
	}}, gavelYAML(t, m))

	_, err := helmTemplate(t, append(append([]string{}, clerkValues...), "storage.type=gcs", "storage.bucket=gavel-blobs")...)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "storage.gcs.credentials is required")
}

// setJSON sets key to the string value through --set-json, which leaves template braces alone.
func setJSON(key, value string) string {
	data, _ := json.Marshal(value)
	return "--set-json=" + key + "=" + string(data)
}

func TestChartTemplatesTheStorageSettings(t *testing.T) {
	requireHelm(t)
	m := render(t,
		"storage.type=s3", setJSON("storage.bucket", "{{ .Release.Name }}-blobs"), setJSON("storage.prefix", "{{ .Release.Namespace }}"),
		setJSON("storage.s3.region", "{{ .Values.awsRegion }}"), "awsRegion=eu-west-1",
		setJSON("storage.s3.endpoint", "https://{{ .Release.Name }}-minio:9000"),
	)
	assert.Equal(t, map[string]any{"storage": map[string]any{
		"type": "s3", "bucket": "gavel-blobs", "prefix": "gavel",
		"s3": map[string]any{"region": "eu-west-1", "endpoint": "https://gavel-minio:9000"},
	}}, gavelYAML(t, m))

	m = render(t, setJSON("storage.path", "/var/lib/gavel/{{ .Release.Name }}"))
	assert.Equal(t, "/var/lib/gavel/gavel", gavelYAML(t, m)["storage"].(map[string]any)["path"])

	m = render(t, "storage.type=gcs", "storage.bucket=b", setJSON("storage.gcs.endpoint", "http://{{ .Release.Name }}-gcs"),
		setJSON("storage.gcs.credentials", "{{ .Values.gcsKey }}"), setJSON("gcsKey", `{"type":"service_account"}`))
	assert.Equal(t, map[string]any{"endpoint": "http://gavel-gcs", "credentials": `{"type":"service_account"}`},
		gavelYAML(t, m)["storage"].(map[string]any)["gcs"])
}

func TestChartReadsBlobStorageFromABroughtSecret(t *testing.T) {
	requireHelm(t)
	m := render(t, "storage.existingSecret=gavel-config")
	mount, ok := mountAt(serveContainer(t, m), "/home/gavel/.gavel.yaml")
	require.True(t, ok)
	assert.Equal(t, "gavel-config", volume(t, m, mount.Name).Secret.SecretName)
	assert.NotContains(t, m["Secret"], "gavel-config")
}

func TestChartRunsSandboxJobsOnSubstrate(t *testing.T) {
	requireHelm(t)
	m := render(t,
		append(substrateSandbox(),
			"sandbox.substrate.template=gavel-runner",
			"sandbox.substrate.egress={api.anthropic.com,api.openai.com}",
			"sandbox.substrate.trustBundle=/run/ate/egress-trust-bundle/trust-bundle.pem",
		)...,
	)
	container := serveContainer(t, m)
	args := strings.Join(container.Args, " ")
	assert.Contains(t, args, " --sandbox substrate --sandbox-image ghcr.io/flanksource/gavel-runner@sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	assert.Contains(t, args, " --substrate-env-api ate-env-api.ate-env.svc.cluster.local:7777 --substrate-control api.ate-system.svc.cluster.local:443")
	assert.Contains(t, args, " --substrate-token-file /var/run/secrets/substrate/token --substrate-atespace ate-env --substrate-template gavel-runner")
	assert.Contains(t, args, " --substrate-egress api.anthropic.com --substrate-egress api.openai.com")
	assert.Contains(t, args, " --substrate-gavel-name beta")
	assert.NotContains(t, args, "--substrate-credential")
	assert.Equal(t, "secret:gavel-credential-provider/token", env(container)["GAVEL_CREDENTIAL_PROVIDER_TOKEN"])
	assert.Contains(t, args, " --substrate-trust-bundle /run/ate/egress-trust-bundle/trust-bundle.pem")
	assert.NotContains(t, args, "--substrate-ca-file", "an unset CA is not passed")

	mount, ok := mountAt(container, "/var/run/secrets/substrate")
	require.True(t, ok)
	projected := volume(t, m, mount.Name).Projected
	require.NotNil(t, projected)
	token := projected.Sources[0].ServiceAccountToken
	assert.Equal(t, "api.ate-system.svc", token.Audience)
	assert.Equal(t, "token", token.Path)
	assert.Equal(t, "gavel", m.deployment(t).Spec.Template.Spec.ServiceAccountName)
	require.Contains(t, m["ServiceAccount"], "gavel")

	m = render(t, append(substrateSandbox(), "sandbox.substrate.ca.secretName=ate-ca", "sandbox.substrate.ca.key=ca.crt")...)
	container = serveContainer(t, m)
	assert.Contains(t, strings.Join(container.Args, " "), " --substrate-ca-file /etc/gavel/substrate-ca/ca.crt")
	mount, ok = mountAt(container, "/etc/gavel/substrate-ca")
	require.True(t, ok)
	assert.Equal(t, "ate-ca", volume(t, m, mount.Name).Secret.SecretName)

	assert.NotContains(t, strings.Join(serveContainer(t, render(t)).Args, " "), "--sandbox", "the sandbox is off unless enabled")
}

func TestChartReachesSubstrateInAnotherCluster(t *testing.T) {
	requireHelm(t)
	m := render(t, append(substrateSandbox(),
		"sandbox.substrate.envAPI=substrate-env.flanksource.com:443", "sandbox.substrate.control=substrate-api.flanksource.com:443",
		"sandbox.substrate.clientCert.secretName=substrate-client",
		"sandbox.substrate.token.secretName=substrate-token", "sandbox.substrate.token.key=token")...)
	container := serveContainer(t, m)
	args := strings.Join(container.Args, " ")
	assert.Contains(t, args, " --substrate-env-api substrate-env.flanksource.com:443 --substrate-control substrate-api.flanksource.com:443")
	assert.Contains(t, args, " --substrate-token-file /var/run/secrets/substrate/token")
	assert.Contains(t, args, " --substrate-client-cert /etc/gavel/substrate-client/tls.crt --substrate-client-key /etc/gavel/substrate-client/tls.key")

	mount, ok := mountAt(container, "/var/run/secrets/substrate")
	require.True(t, ok)
	tokenVolume := volume(t, m, mount.Name)
	assert.Nil(t, tokenVolume.Projected, "a token from substrate's cluster replaces this cluster's projected token")
	require.NotNil(t, tokenVolume.Secret)
	assert.Equal(t, "substrate-token", tokenVolume.Secret.SecretName)
	assert.Equal(t, []corev1.KeyToPath{{Key: "token", Path: "token"}}, tokenVolume.Secret.Items)

	mount, ok = mountAt(container, "/etc/gavel/substrate-client")
	require.True(t, ok)
	assert.Equal(t, "substrate-client", volume(t, m, mount.Name).Secret.SecretName)

	assert.NotContains(t, strings.Join(serveContainer(t, render(t, substrateSandbox()...)).Args, " "),
		"--substrate-client-cert", "an unset client certificate is not passed")
}

// substrateSandbox are the values every substrate sandbox needs.
func substrateSandbox() []string {
	return []string{
		"sandbox.enabled=true",
		"sandbox.image=ghcr.io/flanksource/gavel-runner@sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		"sandbox.substrate.gavelName=beta",
		"sandbox.substrate.credentialProviderToken.secretName=gavel-credential-provider",
	}
}

func TestChartRefusesASandboxWithoutItsCredentialProvider(t *testing.T) {
	requireHelm(t)
	for value, want := range map[string]string{
		"sandbox.substrate.gavelName=":                          "sandbox.substrate.gavelName is required with sandbox.enabled",
		"sandbox.substrate.credentialProviderToken.secretName=": "sandbox.substrate.credentialProviderToken.secretName is required with sandbox.enabled",
	} {
		_, err := helmTemplate(t, append(append(append([]string{}, clerkValues...), substrateSandbox()...), value)...)
		require.Error(t, err, value)
		assert.Contains(t, err.Error(), want, value)
	}
}

func TestChartRefusesASandboxWithoutAPinnedRunnerImage(t *testing.T) {
	requireHelm(t)
	for _, image := range []string{"", "ghcr.io/flanksource/gavel-runner:v1"} {
		_, err := helmTemplate(t, append(append(append([]string{}, clerkValues...), substrateSandbox()...), "sandbox.image="+image)...)
		require.Error(t, err, image)
		assert.Contains(t, err.Error(), "sandbox.image must be pinned by digest, as repo@sha256:<digest>", image)
	}
}

func TestChartConnectsTheGitHubApp(t *testing.T) {
	requireHelm(t)
	m := render(t, "githubApp.id=12345", "githubApp.slug=gavel-dev", "githubApp.privateKeySecretRef.name=gavel-github-app")
	container := serveContainer(t, m)
	assert.Contains(t, strings.Join(container.Args, " "), " --github-app-id 12345 --github-app-slug gavel-dev")
	assert.Equal(t, "/etc/gavel/github-app/private-key.pem", env(container)["GAVEL_GITHUB_APP_PRIVATE_KEY_FILE"])
	mount, ok := mountAt(container, "/etc/gavel/github-app")
	require.True(t, ok)
	secret := volume(t, m, mount.Name).Secret
	assert.Equal(t, "gavel-github-app", secret.SecretName)
	assert.Equal(t, []corev1.KeyToPath{{Key: "private-key.pem", Path: "private-key.pem"}}, secret.Items)

	_, err := helmTemplate(t, append(append([]string{}, clerkValues...), "githubApp.id=12345")...)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "githubApp.slug is required")
	assert.NotContains(t, strings.Join(serveContainer(t, render(t)).Args, " "), "--github-app", "the App is off unless its id is set")
}

func TestChartServesGitPushesOverSSH(t *testing.T) {
	requireHelm(t)
	m := render(t, "ssh.enabled=true", "ssh.service.type=LoadBalancer")
	container := serveContainer(t, m)
	assert.Contains(t, strings.Join(container.Args, " "), " --ssh-addr :2222")
	require.Contains(t, m["Service"], "gavel-ssh")
	service := m["Service"]["gavel-ssh"].(*corev1.Service)
	assert.Equal(t, corev1.ServiceTypeLoadBalancer, service.Spec.Type)
	assert.EqualValues(t, 22, service.Spec.Ports[0].Port)
	assert.Equal(t, "ssh", service.Spec.Ports[0].TargetPort.String())
}

func TestChartWritesTheServerProperties(t *testing.T) {
	requireHelm(t)
	m := render(t, "properties.retention\\.job_logs=30d")
	require.Contains(t, m["ConfigMap"], "gavel")
	assert.Equal(t, "retention.job_logs=30d\n", m["ConfigMap"]["gavel"].(*corev1.ConfigMap).Data["gavel.properties"])
	_, ok := mountAt(serveContainer(t, m), "/etc/gavel/gavel.properties")
	assert.True(t, ok)
}

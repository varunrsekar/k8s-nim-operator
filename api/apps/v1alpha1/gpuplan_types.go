/*
Copyright 2025.

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

package v1alpha1

import (
	"fmt"
	"maps"
	"os"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	"github.com/NVIDIA/k8s-nim-operator/internal/k8sutil"
	rendertypes "github.com/NVIDIA/k8s-nim-operator/internal/render/types"
	utils "github.com/NVIDIA/k8s-nim-operator/internal/utils"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	// GPUPlanConditionReady indicates that the NIM deployment is ready.
	GPUPlanConditionReady = "GPU_PLAN_READY"
	// GPUPlanConditionFailed indicates that the NIM deployment has failed.
	GPUPlanConditionFailed = "GPU_PLAN_FAILED"

	// GPUPlanStatusPending indicates that NIM deployment is in pending state.
	GPUPlanStatusPending = "Pending"
	// GPUPlanStatusNotReady indicates that NIM deployment is not ready.
	GPUPlanStatusNotReady = "NotReady"
	// GPUPlanStatusReady indicates that NIM deployment is ready.
	GPUPlanStatusReady = "Ready"
	// GPUPlanStatusFailed indicates that NIM deployment has failed.
	GPUPlanStatusFailed = "Failed"
)

// GPUPlanSpec defines the desired state of GPUPlan.
type GPUPlanSpec struct {
       // ModelRequestSpec is the specification of a model's inferencing needs.
       // +kubebuilder:validation:MinLength=1
	ModelRequests *ModelRequestSpec `json:"modelRequests"`

	GPURequest *GPURequest `json:"gpuRequest,omitempty"`

	GPUSharing *GPUSharingSpec `json:"gpuSharing,omitempty"`

	// ClaimTemplateGenerationLimit is the maximum number of DRA resource claim templates to generate for the GPUPlan.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=32
	ClaimTemplateGenerationLimit *int32 `json:"claimTemplateGenerationLimit,omitempty"`
}

type ModelRequestSpec struct {
       // ModelConfig is the configuration to access the model repository in NGC/HF/NIMCache.
       //
       // +kubebuilder:validation:Required
	   // +kubebuilder:validation:XValidation:rule="has(self.RegistryConfig) || has(self.NIMCacheConfig)",message="Only one of RegistryConfig or NIMCacheConfig can be set."
       ModelConfig ModelConfig `json:"modelConfig"`

       // Requests is the list of requests for different inferencing needs.
       // Each generated resource claim template would satisfy exactly 1 model request in this list.
       // If this is empty, then model config defaults are used to generate the claim templates.
       Requests []ModelRequest `json:"requests"`
}

type ModelRequest struct {
       // Name is used to uniquely identify the model request
       Name string `json:"name"`

	// Parallesism defines the parallelism configuration for the model inferencing.
	// This is used to determine how many GPUs to use for the model.
	Parallelism *ParallelismSpec `json:"parallelism"`

	// BatchSize is the number of requests to process concurrently by the model.
	// +kubebuilder:validation:Minimum=1
	BatchSize   *int32          `json:"batchSize,omitempty"`

	// SequenceLength is the maximum number of tokens in a single request to process by the model.
	// +kubebuilder:validation:Minimum=1
	SequenceLength *int32 `json:"sequenceLength,omitempty"`

	// MemoryOverheadPercentage is the percentage of memory overhead to account for.
	//
	//+kubebuilder:default="0%"
	//+kubebuilder:validation:XIntOrString
	//+kubebuilder:validation:Pattern="^((100|[0-9]{1,2})%|[0-9]+)$"
	MemoryOverheadPercentage intstr.IntOrString `json:"memoryOverheadPercentage"`
}

type ModelConfig struct {
      RegistryConfig *RegistryConfig `json:"ngcConfig,omitempty"`
	  NIMCacheConfig *NIMCacheConfig `json:"nimCacheConfig,omitempty"`
}

type RegistryConfig struct {
type ParallelismSpec struct {
	// TP aka tensor parallelism defines how many chunks to shard the model horizontally.
	// Each chunk of the model is loaded on a different GPU.
	// +kubebuilder:validation:Minimum=1
	TP *int32 `json:"tp"`
	// PP aka pipeline parallelism defines how many chunks to shard the model vertically.
	// Each chunk of the model is loaded on a different GPU.
	// +kubebuilder:validation:Minimum=1
	PP *int32 `json:"pp"`
}

type GPURequest struct {
	// Product is the product name of the GPU to consider.
	Product *string `json:"product,omitempty"`
	// DeviceID is the GPU device ID to consider.
	DeviceID *string `json:"deviceID,omitempty"`
	// Architecture is the architecture of the GPU.
	Architecture string `json:"architecture"`
	// MemoryBytes is the total amount of memory in bytes needed to be served by the GPU(s).
	//
	// Note: If 1 GPU cannot satisfy this request, then this may lead to multiple GPUs being used.
	//
	// +kubebuilder:validation:Minimum=1
	MemoryBytes *int64 `json:"memoryBytes,omitempty"`

      // TODO: ADD MORE GPU FILTERS AS NECESSARY.
}

type GPUSharingSpec struct {
	// TP aka tensor parallelism defines how many chunks to shard the model horizontally.
	// Each chunk of the model is loaded on a different GPU.
	// +kubebuilder:validation:Minimum=1
	TP *int32 `json:"tp"`
}
// GPUPlanStatus defines the observed state of GPUPlan.
type GPUPlanStatus struct {
	Conditions        []metav1.Condition `json:"conditions,omitempty"`
	State             string             `json:"state,omitempty"`
	ClaimTemplates    []GPUPlanResourceClaimTemplate `json:"claimTemplates,omitempty"`

}

// GPUPlanResourceClaimTemplate defines the status of the DRA resource claim template generated for the GPUPlan.
type GPUPlanResourceClaimTemplate struct {
	Name string `json:"name"`
       ModelRequest string `json:"modelRequest,omitempty"`
	CreationTimestamp metav1.Time `json:"creationTimestamp"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.state`,priority=0
// +kubebuilder:printcolumn:name="Age",type="date",format="date-time",JSONPath=".metadata.creationTimestamp",priority=0

// GPUPlan is the Schema for the nimservices API.
type GPUPlan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GPUPlanSpec   `json:"spec,omitempty"`
	Status GPUPlanStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GPUPlanList contains a list of GPUPlan.
type GPUPlanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GPUPlan `json:"items"`
}

// GetStandardSelectorLabels returns the standard selector labels for the GPUPlan deployment.
func (n *GPUPlan) GetStandardSelectorLabels() map[string]string {
	return map[string]string{
		"app": n.Name,
	}
}

// GetStandardLabels returns the standard set of labels for GPUPlan resources.
func (n *GPUPlan) GetStandardLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":             n.Name,
		"app.kubernetes.io/instance":         n.Name,
		"app.kubernetes.io/operator-version": os.Getenv("OPERATOR_VERSION"),
		"app.kubernetes.io/part-of":          "gou-plan",
		"app.kubernetes.io/managed-by":       "k8s-nim-operator",
	}
}

// GetStandardEnv returns the standard set of env variables for the GPUPlan container.
func (n *GPUPlan) GetStandardEnv() []corev1.EnvVar {
	// add standard env required for NIM service
	envVars := []corev1.EnvVar{
		{
			Name:  "NIM_CACHE_PATH",
			Value: "/model-store",
		},
		{
			Name: "NGC_API_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: n.Spec.AuthSecret,
					},
					Key: "NGC_API_KEY",
				},
			},
		},
		{
			Name:  "OUTLINES_CACHE_DIR",
			Value: "/tmp/outlines",
		},
		{
			Name:  "NIM_SERVER_PORT",
			Value: fmt.Sprintf("%d", *n.Spec.Expose.Service.Port),
		},
		{
			Name:  "NIM_HTTP_API_PORT",
			Value: fmt.Sprintf("%d", *n.Spec.Expose.Service.Port),
		},
		{
			Name:  "NIM_JSONL_LOGGING",
			Value: "1",
		},
		{
			Name:  "NIM_LOG_LEVEL",
			Value: "INFO",
		},
	}
	if n.Spec.Expose.Service.GRPCPort != nil {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "NIM_GRPC_API_PORT",
			Value: fmt.Sprintf("%d", *n.Spec.Expose.Service.GRPCPort),
		}, corev1.EnvVar{
			Name:  "NIM_TRITON_GRPC_PORT",
			Value: fmt.Sprintf("%d", *n.Spec.Expose.Service.GRPCPort),
		})
	}
	if n.Spec.Expose.Service.MetricsPort != nil {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "NIM_TRITON_METRICS_PORT",
			Value: fmt.Sprintf("%d", *n.Spec.Expose.Service.MetricsPort),
		})
	}

	return envVars
}

// GetProxySpec returns the proxy spec for the GPUPlan deployment.
func (n *GPUPlan) GetProxyEnv() []corev1.EnvVar {

	envVars := []corev1.EnvVar{
		{
			Name:  "NIM_SDK_USE_NATIVE_TLS",
			Value: "1",
		},
		{
			Name:  "HTTPS_PROXY",
			Value: n.Spec.Proxy.HttpsProxy,
		},
		{
			Name:  "HTTP_PROXY",
			Value: n.Spec.Proxy.HttpProxy,
		},
		{
			Name:  "NO_PROXY",
			Value: n.Spec.Proxy.NoProxy,
		},
		{
			Name:  "https_proxy",
			Value: n.Spec.Proxy.HttpsProxy,
		},
		{
			Name:  "http_proxy",
			Value: n.Spec.Proxy.HttpProxy,
		},
		{
			Name:  "no_proxy",
			Value: n.Spec.Proxy.NoProxy,
		},
	}

	return envVars
}

// GetStandardAnnotations returns default annotations to apply to the GPUPlan instance.
func (n *GPUPlan) GetStandardAnnotations() map[string]string {
	standardAnnotations := map[string]string{
		"openshift.io/required-scc":             "nonroot",
		utils.NvidiaAnnotationParentSpecHashKey: utils.DeepHashObject(n.Spec),
	}
	if n.GetProxySpec() != nil {
		standardAnnotations["openshift.io/required-scc"] = "anyuid"
	}
	return standardAnnotations
}

// GetGPUPlanAnnotations returns annotations to apply to the GPUPlan instance.
func (n *GPUPlan) GetGPUPlanAnnotations() map[string]string {
	standardAnnotations := n.GetStandardAnnotations()

	if n.Spec.Annotations != nil {
		return utils.MergeMaps(standardAnnotations, n.Spec.Annotations)
	}

	return standardAnnotations
}

// GetServiceLabels returns merged labels to apply to the GPUPlan instance.
func (n *GPUPlan) GetServiceLabels() map[string]string {
	standardLabels := n.GetStandardLabels()

	if n.Spec.Labels != nil {
		return utils.MergeMaps(standardLabels, n.Spec.Labels)
	}
	return standardLabels
}

// GetSelectorLabels returns standard selector labels to apply to the GPUPlan instance.
func (n *GPUPlan) GetSelectorLabels() map[string]string {
	// TODO: add custom ones
	return n.GetStandardSelectorLabels()
}

// GetNodeSelector returns node selector labels for the GPUPlan instance.
func (n *GPUPlan) GetNodeSelector() map[string]string {
	return n.Spec.NodeSelector
}

// GetTolerations returns tolerations for the GPUPlan instance.
func (n *GPUPlan) GetTolerations() []corev1.Toleration {
	return n.Spec.Tolerations
}

// GetPodAffinity returns pod affinity for the GPUPlan instance.
func (n *GPUPlan) GetPodAffinity() *corev1.PodAffinity {
	return n.Spec.PodAffinity
}

// GetContainerName returns name of the container for GPUPlan deployment.
func (n *GPUPlan) GetContainerName() string {
	return fmt.Sprintf("%s-ctr", n.Name)
}

// GetCommand return command to override for the GPUPlan container.
func (n *GPUPlan) GetCommand() []string {
	return n.Spec.Command
}

// GetArgs return arguments for the GPUPlan container.
func (n *GPUPlan) GetArgs() []string {
	return n.Spec.Args
}

// GetEnv returns merged slice of standard and user specified env variables.
func (n *GPUPlan) GetEnv() []corev1.EnvVar {
	envVarList := utils.MergeEnvVars(n.GetStandardEnv(), n.Spec.Env)
	if n.GetProxySpec() != nil {
		envVarList = utils.MergeEnvVars(envVarList, n.GetProxyEnv())
	}
	return envVarList
}

// GetImage returns container image for the GPUPlan.
func (n *GPUPlan) GetImage() string {
	return fmt.Sprintf("%s:%s", n.Spec.Image.Repository, n.Spec.Image.Tag)
}

// GetImagePullSecrets returns the image pull secrets for the NIM container.
func (n *GPUPlan) GetImagePullSecrets() []string {
	return n.Spec.Image.PullSecrets
}

// GetImagePullPolicy returns the image pull policy for the NIM container.
func (n *GPUPlan) GetImagePullPolicy() string {
	return n.Spec.Image.PullPolicy
}

// GetResources returns resources to allocate to the GPUPlan container.
func (n *GPUPlan) GetResources() *corev1.ResourceRequirements {
	return n.Spec.Resources
}

// IsProbeEnabled returns true if a given liveness/readiness/startup probe is enabled.
func IsProbeEnabled(probe Probe) bool {
	if probe.Enabled == nil {
		return true
	}
	return *probe.Enabled
}

// GetLivenessProbe returns liveness probe for the GPUPlan container.
func (n *GPUPlan) GetLivenessProbe() *corev1.Probe {
	if n.Spec.LivenessProbe.Probe == nil {
		return n.GetDefaultLivenessProbe()
	}
	return n.Spec.LivenessProbe.Probe
}

// GetDefaultLivenessProbe returns the default liveness probe for the GPUPlan container.
func (n *GPUPlan) GetDefaultLivenessProbe() *corev1.Probe {
	probe := corev1.Probe{
		InitialDelaySeconds: 15,
		TimeoutSeconds:      1,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/v1/health/live",
				Port: intstr.FromString(DefaultNamedPortAPI),
			},
		},
	}

	return &probe
}

// GetReadinessProbe returns readiness probe for the GPUPlan container.
func (n *GPUPlan) GetReadinessProbe() *corev1.Probe {
	if n.Spec.ReadinessProbe.Probe == nil {
		return n.GetDefaultReadinessProbe()
	}
	return n.Spec.ReadinessProbe.Probe
}

// GetDefaultReadinessProbe returns the default readiness probe for the GPUPlan container.
func (n *GPUPlan) GetDefaultReadinessProbe() *corev1.Probe {
	probe := corev1.Probe{
		InitialDelaySeconds: 15,
		TimeoutSeconds:      1,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/v1/health/ready",
				Port: intstr.FromString(DefaultNamedPortAPI),
			},
		},
	}

	return &probe
}

// GetStartupProbe returns startup probe for the GPUPlan container.
func (n *GPUPlan) GetStartupProbe() *corev1.Probe {
	if n.Spec.StartupProbe.Probe == nil {
		return n.GetDefaultStartupProbe()
	}
	return n.Spec.StartupProbe.Probe
}

// GetDefaultStartupProbe returns the default startup probe for the GPUPlan container.
func (n *GPUPlan) GetDefaultStartupProbe() *corev1.Probe {
	probe := corev1.Probe{
		InitialDelaySeconds: 30,
		TimeoutSeconds:      1,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    30,
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/v1/health/ready",
				Port: intstr.FromString(DefaultNamedPortAPI),
			},
		},
	}

	return &probe
}

// GetVolumesMounts returns volume mounts for the GPUPlan container.
func (n *GPUPlan) GetVolumesMounts() []corev1.Volume {
	// TODO: setup volume mounts required for NIM
	return nil
}

// GetVolumes returns volumes for the GPUPlan container.
func (n *GPUPlan) GetVolumes(modelPVC PersistentVolumeClaim) []corev1.Volume {
	// TODO: Fetch actual PVC name from associated NIMCache obj
	volumes := []corev1.Volume{
		{
			Name: "dshm",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium:    corev1.StorageMediumMemory,
					SizeLimit: n.Spec.Storage.SharedMemorySizeLimit,
				},
			},
		},
		{
			Name: "model-store",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: modelPVC.Name,
					ReadOnly:  n.GetStorageReadOnly(),
				},
			},
		},
	}

	if n.GetProxySpec() != nil {
		volumes = append(volumes, k8sutil.GetVolumesForUpdatingCaCert(n.Spec.Proxy.CertConfigMap)...)
	}

	return volumes
}

// GetVolumeMounts returns volumes for the GPUPlan container.
func (n *GPUPlan) GetVolumeMounts(modelPVC PersistentVolumeClaim) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "model-store",
			MountPath: "/model-store",
			SubPath:   modelPVC.SubPath,
		},
		{
			Name:      "dshm",
			MountPath: "/dev/shm",
		},
	}

	if n.GetProxySpec() != nil {
		volumeMounts = append(volumeMounts, k8sutil.GetVolumesMountsForUpdatingCaCert()...)
	}
	return volumeMounts
}

// GetServiceAccountName returns service account name for the GPUPlan deployment.
func (n *GPUPlan) GetServiceAccountName() string {
	return n.Name
}

// GetRuntimeClassName return the runtime class name for the GPUPlan deployment.
func (n *GPUPlan) GetRuntimeClassName() string {
	return n.Spec.RuntimeClassName
}

// GetNIMCacheName returns the NIMCache name to use for the GPUPlan deployment.
func (n *GPUPlan) GetNIMCacheName() string {
	return n.Spec.Storage.NIMCache.Name
}

// GetNIMCacheProfile returns the explicit profile to use for the GPUPlan deployment.
func (n *GPUPlan) GetNIMCacheProfile() string {
	return n.Spec.Storage.NIMCache.Profile
}

// GetHPA returns the HPA spec for the GPUPlan deployment.
func (n *GPUPlan) GetHPA() HorizontalPodAutoscalerSpec {
	return n.Spec.Scale.HPA
}

// GetServiceMonitor returns the Service Monitor details for the GPUPlan deployment.
func (n *GPUPlan) GetServiceMonitor() ServiceMonitor {
	return n.Spec.Metrics.ServiceMonitor
}

// GetReplicas returns replicas for the GPUPlan deployment.
func (n *GPUPlan) GetReplicas() int {
	if n.IsAutoScalingEnabled() {
		return 0
	}
	return n.Spec.Replicas
}

// GetDeploymentKind returns the kind of deployment for GPUPlan.
func (n *GPUPlan) GetDeploymentKind() string {
	return "Deployment"
}

// GetInitContainers returns the init containers for the GPUPlan deployment.
func (n *GPUPlan) GetInitContainers() []corev1.Container {
	if n.Spec.Proxy != nil {
		return []corev1.Container{
			{
				Name:            "update-ca-certificates",
				Image:           n.GetImage(),
				ImagePullPolicy: corev1.PullPolicy(n.GetImagePullPolicy()),
				Command:         k8sutil.GetUpdateCaCertInitContainerCommand(),
				SecurityContext: k8sutil.GetUpdateCaCertInitContainerSecurityContext(),
				VolumeMounts:    k8sutil.GetUpdateCaCertInitContainerVolumeMounts(),
			},
		}
	}
	return []corev1.Container{}
}

// IsAutoScalingEnabled returns true if autoscaling is enabled for GPUPlan deployment.
func (n *GPUPlan) IsAutoScalingEnabled() bool {
	return n.Spec.Scale.Enabled != nil && *n.Spec.Scale.Enabled
}

// IsIngressEnabled returns true if ingress is enabled for GPUPlan deployment.
func (n *GPUPlan) IsIngressEnabled() bool {
	return n.Spec.Expose.Ingress.Enabled != nil && *n.Spec.Expose.Ingress.Enabled
}

// GetIngressSpec returns the Ingress spec GPUPlan deployment.
func (n *GPUPlan) GetIngressSpec() networkingv1.IngressSpec {
	return n.Spec.Expose.Ingress.Spec
}

// IsServiceMonitorEnabled returns true if servicemonitor is enabled for GPUPlan deployment.
func (n *GPUPlan) IsServiceMonitorEnabled() bool {
	return n.Spec.Metrics.Enabled != nil && *n.Spec.Metrics.Enabled
}

// GetServicePort returns the service port for the GPUPlan deployment or default port.
func (n *GPUPlan) GetServicePort() int32 {
	if n.Spec.Expose.Service.Port == nil {
		return DefaultAPIPort
	}
	return *n.Spec.Expose.Service.Port
}

// GetServiceType returns the service type for the GPUPlan deployment.
func (n *GPUPlan) GetServiceType() string {
	return string(n.Spec.Expose.Service.Type)
}

// GetUserID returns the user ID for the GPUPlan deployment.
func (n *GPUPlan) GetUserID() *int64 {
	if n.Spec.UserID != nil {
		return n.Spec.UserID
	}
	return ptr.To[int64](1000)
}

// GetGroupID returns the group ID for the GPUPlan deployment.
func (n *GPUPlan) GetGroupID() *int64 {
	if n.Spec.GroupID != nil {
		return n.Spec.GroupID
	}
	return ptr.To[int64](2000)
}

// GetStorageReadOnly returns true if the volume have to be mounted as read-only for the GPUPlan deployment.
func (n *GPUPlan) GetStorageReadOnly() bool {
	if n.Spec.Storage.ReadOnly == nil {
		return false
	}
	return *n.Spec.Storage.ReadOnly
}

// GetServiceAccountParams return params to render ServiceAccount from templates.
func (n *GPUPlan) GetServiceAccountParams() *rendertypes.ServiceAccountParams {
	params := &rendertypes.ServiceAccountParams{}

	// Set metadata
	params.Name = n.GetName()
	params.Namespace = n.GetNamespace()
	params.Labels = n.GetServiceLabels()
	params.Annotations = n.GetGPUPlanAnnotations()
	return params
}

// GetDeploymentParams returns params to render Deployment from templates.
func (n *GPUPlan) GetDeploymentParams() *rendertypes.DeploymentParams {
	params := &rendertypes.DeploymentParams{}

	// Set metadata
	params.Name = n.GetName()
	params.Namespace = n.GetNamespace()
	params.Labels = n.GetServiceLabels()
	params.Annotations = n.GetGPUPlanAnnotations()
	params.PodAnnotations = n.GetGPUPlanAnnotations()
	delete(params.PodAnnotations, utils.NvidiaAnnotationParentSpecHashKey)

	// Set template spec
	if !n.IsAutoScalingEnabled() {
		params.Replicas = n.GetReplicas()
	}
	params.NodeSelector = n.GetNodeSelector()
	params.Tolerations = n.GetTolerations()
	params.Affinity = n.GetPodAffinity()
	params.ImagePullSecrets = n.GetImagePullSecrets()
	params.ImagePullPolicy = n.GetImagePullPolicy()

	// Set labels and selectors
	params.SelectorLabels = n.GetSelectorLabels()

	// Set container spec
	params.ContainerName = n.GetContainerName()
	params.Env = n.GetEnv()
	params.Args = n.GetArgs()
	params.Command = n.GetCommand()
	params.Resources = n.GetResources()
	params.Image = n.GetImage()

	// Set container probes
	if IsProbeEnabled(n.Spec.LivenessProbe) {
		params.LivenessProbe = n.GetLivenessProbe()
	}
	if IsProbeEnabled(n.Spec.ReadinessProbe) {
		params.ReadinessProbe = n.GetReadinessProbe()
	}
	if IsProbeEnabled(n.Spec.StartupProbe) {
		params.StartupProbe = n.GetStartupProbe()
	}
	params.UserID = n.GetUserID()
	params.GroupID = n.GetGroupID()

	// Set service account
	params.ServiceAccountName = n.GetServiceAccountName()

	// Set runtime class
	params.RuntimeClassName = n.GetRuntimeClassName()

	// Set scheduler
	params.SchedulerName = n.GetSchedulerName()

	// Setup container ports for nimservice
	params.Ports = []corev1.ContainerPort{
		{
			Name:          DefaultNamedPortAPI,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: *n.Spec.Expose.Service.Port,
		},
	}
	if n.Spec.Expose.Service.GRPCPort != nil {
		params.Ports = append(params.Ports, corev1.ContainerPort{
			Name:          DefaultNamedPortGRPC,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: *n.Spec.Expose.Service.GRPCPort,
		})
	}
	if n.Spec.Expose.Service.MetricsPort != nil {
		params.Ports = append(params.Ports, corev1.ContainerPort{
			Name:          DefaultNamedPortMetrics,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: *n.Spec.Expose.Service.MetricsPort,
		})
	}
	return params
}

// GetSchedulerName returns the scheduler name for the GPUPlan deployment.
func (n *GPUPlan) GetSchedulerName() string {
	return n.Spec.SchedulerName
}

// GetStatefulSetParams returns params to render StatefulSet from templates.
func (n *GPUPlan) GetStatefulSetParams() *rendertypes.StatefulSetParams {

	params := &rendertypes.StatefulSetParams{}

	// Set metadata
	params.Name = n.GetName()
	params.Namespace = n.GetNamespace()
	params.Labels = n.GetServiceLabels()
	params.Annotations = n.GetGPUPlanAnnotations()

	// Set template spec
	if !n.IsAutoScalingEnabled() {
		params.Replicas = n.GetReplicas()
	}
	params.ServiceName = n.GetName()
	params.NodeSelector = n.GetNodeSelector()
	params.Tolerations = n.GetTolerations()
	params.Affinity = n.GetPodAffinity()
	params.ImagePullSecrets = n.GetImagePullSecrets()
	params.ImagePullPolicy = n.GetImagePullPolicy()

	// Set labels and selectors
	params.SelectorLabels = n.GetSelectorLabels()

	// Set container spec
	params.ContainerName = n.GetContainerName()
	params.Env = n.GetEnv()
	params.Args = n.GetArgs()
	params.Command = n.GetCommand()
	params.Resources = n.GetResources()
	params.Image = n.GetImage()

	// Set container probes
	params.LivenessProbe = n.GetLivenessProbe()
	params.ReadinessProbe = n.GetReadinessProbe()
	params.StartupProbe = n.GetStartupProbe()

	// Set service account
	params.ServiceAccountName = n.GetServiceAccountName()

	// Set runtime class
	params.RuntimeClassName = n.GetRuntimeClassName()
	return params
}

// GetServiceParams returns params to render Service from templates.
func (n *GPUPlan) GetServiceParams() *rendertypes.ServiceParams {
	params := &rendertypes.ServiceParams{}

	// Set metadata
	params.Name = n.GetName()
	params.Namespace = n.GetNamespace()
	params.Labels = n.GetServiceLabels()
	params.Annotations = n.GetServiceAnnotations()

	// Set service selector labels
	params.SelectorLabels = n.GetSelectorLabels()

	// Set service type
	params.Type = n.GetServiceType()

	// Set service ports
	params.Ports = []corev1.ServicePort{
		{
			Name:       DefaultNamedPortAPI,
			Port:       n.GetServicePort(),
			TargetPort: intstr.FromString((DefaultNamedPortAPI)),
			Protocol:   corev1.ProtocolTCP,
		},
	}
	if n.Spec.Expose.Service.GRPCPort != nil {
		params.Ports = append(params.Ports, corev1.ServicePort{
			Name:       DefaultNamedPortGRPC,
			Port:       *n.Spec.Expose.Service.GRPCPort,
			TargetPort: intstr.FromString(DefaultNamedPortGRPC),
			Protocol:   corev1.ProtocolTCP,
		})
	}
	if n.Spec.Expose.Service.MetricsPort != nil {
		params.Ports = append(params.Ports, corev1.ServicePort{
			Name:       DefaultNamedPortMetrics,
			Port:       *n.Spec.Expose.Service.MetricsPort,
			TargetPort: intstr.FromString(DefaultNamedPortMetrics),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	return params
}

// GetIngressParams returns params to render Ingress from templates.
func (n *GPUPlan) GetIngressParams() *rendertypes.IngressParams {
	params := &rendertypes.IngressParams{}

	params.Enabled = n.IsIngressEnabled()
	// Set metadata
	params.Name = n.GetName()
	params.Namespace = n.GetNamespace()
	params.Labels = n.GetServiceLabels()
	params.Annotations = n.GetIngressAnnotations()
	params.Spec = n.GetIngressSpec()
	return params
}

// GetRoleParams returns params to render Role from templates.
func (n *GPUPlan) GetRoleParams() *rendertypes.RoleParams {
	params := &rendertypes.RoleParams{}

	// Set metadata
	params.Name = n.GetName()
	params.Namespace = n.GetNamespace()

	// Set rules to use SCC
	if n.GetProxySpec() != nil {
		params.Rules = []rbacv1.PolicyRule{
			{
				APIGroups:     []string{"security.openshift.io"},
				Resources:     []string{"securitycontextconstraints"},
				ResourceNames: []string{"anyuid"},
				Verbs:         []string{"use"},
			},
		}
	} else {
		params.Rules = []rbacv1.PolicyRule{
			{
				APIGroups:     []string{"security.openshift.io"},
				Resources:     []string{"securitycontextconstraints"},
				ResourceNames: []string{"nonroot"},
				Verbs:         []string{"use"},
			},
		}
	}

	return params
}

// GetRoleBindingParams returns params to render RoleBinding from templates.
func (n *GPUPlan) GetRoleBindingParams() *rendertypes.RoleBindingParams {
	params := &rendertypes.RoleBindingParams{}

	// Set metadata
	params.Name = n.GetName()
	params.Namespace = n.GetNamespace()

	params.ServiceAccountName = n.GetServiceAccountName()
	params.RoleName = n.GetName()
	return params
}

// GetHPAParams returns params to render HPA from templates.
func (n *GPUPlan) GetHPAParams() *rendertypes.HPAParams {
	params := &rendertypes.HPAParams{}

	params.Enabled = n.IsAutoScalingEnabled()

	// Set metadata
	params.Name = n.GetName()
	params.Namespace = n.GetNamespace()
	params.Labels = n.GetServiceLabels()
	params.Annotations = n.GetHPAAnnotations()

	// Set HPA spec
	hpa := n.GetHPA()
	hpaSpec := autoscalingv2.HorizontalPodAutoscalerSpec{
		ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
			Kind:       n.GetDeploymentKind(),
			Name:       n.GetName(),
			APIVersion: "apps/v1",
		},
		MinReplicas: hpa.MinReplicas,
		MaxReplicas: hpa.MaxReplicas,
		Metrics:     hpa.Metrics,
		Behavior:    hpa.Behavior,
	}
	params.HPASpec = hpaSpec
	return params
}

// GetSCCParams return params to render SCC from templates.
func (n *GPUPlan) GetSCCParams() *rendertypes.SCCParams {
	params := &rendertypes.SCCParams{}
	// Set metadata
	params.Name = "nim-service-scc"

	params.ServiceAccountName = n.GetServiceAccountName()
	return params
}

// GetServiceMonitorParams return params to render Service Monitor from templates.
func (n *GPUPlan) GetServiceMonitorParams() *rendertypes.ServiceMonitorParams {
	params := &rendertypes.ServiceMonitorParams{}
	serviceMonitor := n.GetServiceMonitor()
	params.Enabled = n.IsServiceMonitorEnabled()
	params.Name = n.GetName()
	params.Namespace = n.GetNamespace()
	svcLabels := n.GetServiceLabels()
	maps.Copy(svcLabels, serviceMonitor.AdditionalLabels)
	params.Labels = svcLabels
	params.Annotations = n.GetServiceMonitorAnnotations()

	// Set Service Monitor spec
	smSpec := monitoringv1.ServiceMonitorSpec{
		NamespaceSelector: monitoringv1.NamespaceSelector{MatchNames: []string{n.Namespace}},
		Selector:          metav1.LabelSelector{MatchLabels: n.GetServiceLabels()},
		Endpoints: []monitoringv1.Endpoint{
			{
				Path:          "/v1/metrics",
				Port:          DefaultNamedPortAPI,
				ScrapeTimeout: serviceMonitor.ScrapeTimeout,
				Interval:      serviceMonitor.Interval,
			},
		},
	}
	if n.Spec.Expose.Service.MetricsPort != nil {
		smSpec.Endpoints = append(smSpec.Endpoints, monitoringv1.Endpoint{
			Path:          "/metrics",
			Port:          DefaultNamedPortMetrics,
			ScrapeTimeout: serviceMonitor.ScrapeTimeout,
			Interval:      serviceMonitor.Interval,
		})
	}
	params.SMSpec = smSpec
	return params
}

func (n *GPUPlan) GetIngressAnnotations() map[string]string {
	nimServiceAnnotations := n.GetGPUPlanAnnotations()

	if n.Spec.Expose.Ingress.Annotations != nil {
		return utils.MergeMaps(nimServiceAnnotations, n.Spec.Expose.Ingress.Annotations)
	}
	return nimServiceAnnotations
}

func (n *GPUPlan) GetServiceAnnotations() map[string]string {
	nimServiceAnnotations := n.GetGPUPlanAnnotations()

	if n.Spec.Expose.Service.Annotations != nil {
		return utils.MergeMaps(nimServiceAnnotations, n.Spec.Expose.Service.Annotations)
	}
	return nimServiceAnnotations
}

func (n *GPUPlan) GetHPAAnnotations() map[string]string {
	nimServiceAnnotations := n.GetGPUPlanAnnotations()

	if n.Spec.Scale.Annotations != nil {
		return utils.MergeMaps(nimServiceAnnotations, n.Spec.Scale.Annotations)
	}
	return nimServiceAnnotations
}

func (n *GPUPlan) GetServiceMonitorAnnotations() map[string]string {
	nimServiceAnnotations := n.GetGPUPlanAnnotations()

	if n.Spec.Metrics.ServiceMonitor.Annotations != nil {
		return utils.MergeMaps(nimServiceAnnotations, n.Spec.Metrics.ServiceMonitor.Annotations)
	}
	return nimServiceAnnotations
}

// GetProxySpec returns the proxy spec for the GPUPlan deployment.
func (n *GPUPlan) GetProxySpec() *ProxySpec {
	return n.Spec.Proxy
}

func init() {
	SchemeBuilder.Register(&GPUPlan{}, &GPUPlanList{})
}

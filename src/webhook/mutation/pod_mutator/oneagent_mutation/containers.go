package oneagent_mutation

import (
	dynatracev1beta1 "github.com/Dynatrace/dynatrace-operator/src/api/v1beta1"
	"github.com/Dynatrace/dynatrace-operator/src/kubeobjects"
	dtwebhook "github.com/Dynatrace/dynatrace-operator/src/webhook"
	corev1 "k8s.io/api/core/v1"
)

func (mutator *OneAgentPodMutator) configureInitContainer(request *dtwebhook.MutationRequest, installer installerInfo) {
	addInstallerInitEnvs(request.InstallContainer, installer, mutator.getVolumeMode(request.DynaKube))
	addInitVolumeMounts(request.InstallContainer)
}

func (mutator *OneAgentPodMutator) mutateUserContainers(request *dtwebhook.MutationRequest) {
	for i := range request.Pod.Spec.Containers {
		container := &request.Pod.Spec.Containers[i]
		addContainerInfoInitEnv(request.InstallContainer, i+1, container.Name, container.Image)
		mutator.addOneAgentToContainer(request.Pod, container, request.DynaKube)
	}
}

// reinvokeUserContainers mutates each user container that hasn't been injected yet.
// It makes sure that the new containers will have an environment variable in the install-container
// that doesn't conflict with the previous environment variables of the originally injected containers
func (mutator *OneAgentPodMutator) reinvokeUserContainers(request *dtwebhook.ReinvocationRequest) bool {
	pod := request.Pod
	initContainer := findOneAgentInstallContainer(pod.Spec.InitContainers)
	newContainers := []*corev1.Container{}

	for i := range pod.Spec.Containers {
		currentContainer := &pod.Spec.Containers[i]
		if containerIsInjected(currentContainer) {
			continue
		}
		newContainers = append(newContainers, currentContainer)
	}

	oldContainersLen := len(pod.Spec.Containers) - len(newContainers)
	for i := range newContainers {
		currentContainer := newContainers[i]
		addContainerInfoInitEnv(initContainer, oldContainersLen+i+1, currentContainer.Name, currentContainer.Image)
		mutator.addOneAgentToContainer(request.Pod, currentContainer, request.DynaKube)
	}
	return len(newContainers) > 0
}

func (mutator *OneAgentPodMutator) addOneAgentToContainer(pod *corev1.Pod, container *corev1.Container, dynakube dynatracev1beta1.DynaKube) {
	log.Info("adding OneAgent to container", "name", container.Name)
	installPath := kubeobjects.GetField(pod.Annotations, dtwebhook.AnnotationInstallPath, dtwebhook.DefaultInstallPath)

	addOneAgentVolumeMounts(container, installPath)
	addDeploymentMetadataEnv(container, dynakube, mutator.clusterID)
	addPreloadEnv(container, installPath)

	if dynakube.HasActiveGateCaCert() {
		addCertVolumeMounts(container)
	}

	if dynakube.FeatureAgentInitialConnectRetry() > 0 {
		addCurlOptionsVolumeMount(container)
	}

	if dynakube.NeedsOneAgentProxy() {
		addProxyEnv(container)
	}

	if dynakube.Spec.NetworkZone != "" {
		addNetworkZoneEnv(container, dynakube.Spec.NetworkZone)
	}
}

func findOneAgentInstallContainer(initContainers []corev1.Container) *corev1.Container {
	for i := range initContainers {
		container := &initContainers[i]
		if container.Name == dtwebhook.InstallContainerName {
			return container
		}
	}
	return nil
}

package dataingest_mutation

import corev1 "k8s.io/api/core/v1"

func mutateUserContainers(pod *corev1.Pod) {
	for i := range pod.Spec.Containers {
		container := &pod.Spec.Containers[i]
		setupVolumeMountsForUserContainer(container)
	}
}

func reinvokeUserContainers(pod *corev1.Pod) bool {
	var updated bool
	for i := range pod.Spec.Containers {
		container := &pod.Spec.Containers[i]
		if containerIsInjected(container) {
			continue
		}
		setupVolumeMountsForUserContainer(container)
		updated = true
	}
	return updated
}

func updateInstallContainer(installContainer *corev1.Container, workload *workloadInfo) {
	addWorkloadInfoEnvs(installContainer, workload)
	addWorkloadEnrichmentVolumeMount(installContainer)
}

package oneagent_mutation

import (
	"fmt"
	"path/filepath"

	dynatracev1beta1 "github.com/Dynatrace/dynatrace-operator/src/api/v1beta1"
	dtcsi "github.com/Dynatrace/dynatrace-operator/src/controllers/csi"
	csivolumes "github.com/Dynatrace/dynatrace-operator/src/controllers/csi/driver/volumes"
	appvolumes "github.com/Dynatrace/dynatrace-operator/src/controllers/csi/driver/volumes/app"
	"github.com/Dynatrace/dynatrace-operator/src/standalone"
	dtwebhook "github.com/Dynatrace/dynatrace-operator/src/webhook"
	corev1 "k8s.io/api/core/v1"
)

func (mutator *OneAgentPodMutator) addVolumes(pod *corev1.Pod, dynakube dynatracev1beta1.DynaKube) {
	addInjectionConfigVolume(pod)
	addOneAgentVolumes(pod, dynakube)
}

func addOneAgentVolumeMounts(container *corev1.Container, installPath string) {
	container.VolumeMounts = append(container.VolumeMounts,
		corev1.VolumeMount{
			Name:      oneAgentShareVolumeName,
			MountPath: preloadPath,
			SubPath:   preloadSubPath,
		},
		corev1.VolumeMount{
			Name:      oneAgentBinVolumeName,
			MountPath: installPath,
		},
		corev1.VolumeMount{
			Name:      oneAgentShareVolumeName,
			MountPath: containerConfPath,
			SubPath:   getContainerConfSubPath(container.Name),
		})
}

func getContainerConfSubPath(containerName string) string {
	return fmt.Sprintf(standalone.ContainerConfFilenameTemplate, containerName)
}

func addCertVolumeMounts(container *corev1.Container) {
	container.VolumeMounts = append(container.VolumeMounts,
		corev1.VolumeMount{
			Name:      oneAgentShareVolumeName,
			MountPath: filepath.Join(oneAgentCustomKeysPath, customCertFileName),
			SubPath:   customCertFileName,
		})
}

func addInitVolumeMounts(initContainer *corev1.Container) {
	initContainer.VolumeMounts = append(initContainer.VolumeMounts,
		corev1.VolumeMount{Name: oneAgentBinVolumeName, MountPath: standalone.BinDirMount},
		corev1.VolumeMount{Name: oneAgentShareVolumeName, MountPath: standalone.ShareDirMount},
	)
}

func addCurlOptionsVolumeMount(container *corev1.Container) {
	container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
		Name:      oneAgentShareVolumeName,
		MountPath: filepath.Join(oneAgentCustomKeysPath, standalone.CurlOptionsFileName),
		SubPath:   standalone.CurlOptionsFileName,
	})
}

func addInjectionConfigVolume(pod *corev1.Pod) {
	pod.Spec.Volumes = append(pod.Spec.Volumes,
		corev1.Volume{
			Name: injectionConfigVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: dtwebhook.SecretConfigName,
				},
			},
		},
	)
}

func addOneAgentVolumes(pod *corev1.Pod, dynakube dynatracev1beta1.DynaKube) {
	pod.Spec.Volumes = append(pod.Spec.Volumes,
		corev1.Volume{
			Name:         oneAgentBinVolumeName,
			VolumeSource: getInstallerVolumeSource(dynakube),
		},
		corev1.Volume{
			Name: oneAgentShareVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	)
}

func getInstallerVolumeSource(dynakube dynatracev1beta1.DynaKube) corev1.VolumeSource {
	volumeSource := corev1.VolumeSource{}
	if dynakube.NeedsCSIDriver() {
		volumeSource.CSI = &corev1.CSIVolumeSource{
			Driver: dtcsi.DriverName,
			VolumeAttributes: map[string]string{
				csivolumes.CSIVolumeAttributeModeField:     appvolumes.Mode,
				csivolumes.CSIVolumeAttributeDynakubeField: dynakube.Name,
			},
		}
	} else {
		volumeSource.EmptyDir = &corev1.EmptyDirVolumeSource{}
	}
	return volumeSource
}

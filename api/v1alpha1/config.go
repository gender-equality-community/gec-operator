package v1alpha1

import (
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type ClusterApp uint8

const (
	UnknownClusterApp ClusterApp = iota
	ClusterBot
	ClusterProcessor
	ClusterSlacker
	ClusterMeta
)

const (
	cosignPubKey = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEfh98reV6VTLXq2TzyekyybK1sPiD
7ndQ+oC6kjGsQawwMUCFU7oCpW2hmjXA/Zj4x6A4zPZl/3nvRTVDsIMxHA==
-----END PUBLIC KEY-----
`
	botContainerImage       = "ghcr.io/gender-equality-community/gec-bot"
	processorContainerImage = "ghcr.io/gender-equality-community/gec-processor"
	slackerContainerImage   = "ghcr.io/gender-equality-community/gec-slacker"
)

var (
	// VolumeType is used to determine things like storage classes and
	// volume configs between cluster types.
	//
	// For instance, on GCP we want to use GCE disks, whereas locally we
	// might actually want our NFS volumes
	VolumeType = os.Getenv("VOLUME_TYPE")
	VolumeSize = resource.MustParse("100Mi")

	// Default resources are used for smaller containers, largely
	// written in go
	defaultCpu = resource.MustParse("100m")
	defaultMem = resource.MustParse("64Mi")

	// Data resources are used for containers which perform data-y
	// tasks, such as taggers and labelers
	dataCpu = resource.MustParse("1")
	dataMem = resource.MustParse("4Gi")
)

func (c ClusterApp) String() string {
	switch c {
	case ClusterBot:
		return "gec-bot"

	case ClusterProcessor:
		return "gec-processor"

	case ClusterSlacker:
		return "gec-slacker"

	case ClusterMeta:
		return "meta"

	default:
		return "unknown"
	}
}

func (c ClusterApp) Resources() corev1.ResourceList {
	if c == ClusterProcessor {
		return corev1.ResourceList{
			corev1.ResourceCPU:    dataCpu,
			corev1.ResourceMemory: dataMem,
		}
	}

	return corev1.ResourceList{
		corev1.ResourceCPU:    defaultCpu,
		corev1.ResourceMemory: defaultMem,
	}
}

func (c ClusterApp) VolumeMount(name string) []corev1.VolumeMount {
	if c != ClusterBot {
		return nil
	}

	return []corev1.VolumeMount{
		{
			Name:      name,
			MountPath: "/database/",
		},
	}
}

func (c ClusterApp) Volume(name string) []corev1.Volume {
	if c != ClusterBot {
		return nil
	}

	switch VolumeType {
	case "gce":
		return gceVolume(name)

	case "pvc":
		return pvcVolume(name)

	default:
		return defaultVolume(name)
	}
}

func defaultVolume(name string) []corev1.Volume {
	return pvcVolume(name)
}

func gceVolume(name string) []corev1.Volume {
	return []corev1.Volume{
		{
			Name: name,
			VolumeSource: corev1.VolumeSource{
				GCEPersistentDisk: &corev1.GCEPersistentDiskVolumeSource{
					PDName: name,
				},
			},
		},
	}

}

func pvcVolume(name string) []corev1.Volume {
	return []corev1.Volume{
		{
			Name: name,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: name,
				},
			},
		},
	}
}

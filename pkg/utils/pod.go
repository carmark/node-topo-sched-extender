package utils

import (
	"log"
	"strconv"

	"k8s.io/api/core/v1"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

// AssignedNonTerminatedPod selects pods that are assigned and non-terminal (scheduled and running).
func AssignedNonTerminatedPod(pod *v1.Pod) bool {
	if pod.DeletionTimestamp != nil {
		return false
	}

	if len(pod.Spec.NodeName) == 0 {
		return false
	}
	if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
		return false
	}
	return true
}

// IsCompletePod determines if the pod is complete
func IsCompletePod(pod *v1.Pod) bool {
	if pod.DeletionTimestamp != nil {
		return true
	}

	if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
		return true
	}
	return false
}

// IsGPUTopoPod determines if it's the pod for GPU topology
func IsGPUTopoPod(pod *v1.Pod) bool {
	return GetGPUTopoNum(pod) > 0
}

// GetGPUIDFromAnnotation gets GPU UUID from Annotation
func GetGPUIDFromAnnotation(pod *v1.Pod) string {
	if len(pod.ObjectMeta.Annotations) > 0 {
		value, found := pod.ObjectMeta.Annotations[ResourceName]
		if found {
			return value
		}
	}

	return ""
}

// GetGPUIDFromEnv gets GPU ID from Env
func GetGPUIDFromEnv(pod *v1.Pod) int {
	id := -1
	for _, container := range pod.Spec.Containers {
		id = getGPUIDFromContainer(container)
		if id >= 0 {
			return id
		}
	}

	return id
}

func getGPUIDFromContainer(container v1.Container) (devIdx int) {
	devIdx = -1
	var err error
loop:
	for _, env := range container.Env {
		if env.Name == EnvResourceIndex {
			devIdx, err = strconv.Atoi(env.Value)
			if err != nil {
				log.Printf("warn: Failed due to %v for %s", err, container.Name)
				devIdx = -1
			}
			break loop
		}
	}

	return devIdx
}

func GetGPUTopoNum(pod *v1.Pod) int64 {

	res := &schedulernodeinfo.Resource{}
	for _, container := range pod.Spec.Containers {
		res.Add(container.Resources.Requests)
	}

	// take max_resource(sum_pod, any_init_container)
	for _, container := range pod.Spec.InitContainers {
		res.SetMaxResource(container.Resources.Requests)
	}

	resList := res.ResourceList()
	gpuTopo := resList["nvidia.com/gpu-topo"]
	gpuTopoNum, _ := gpuTopo.AsInt64()

	return gpuTopoNum
}

package scheduler

import (
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"

	"github.com/gpucloud/node-topology-manager/pkg/cache"
	"github.com/gpucloud/node-topology-manager/pkg/utils"
)

type Priority struct {
	Name   string
	client *kubernetes.Clientset
	pcache *cache.SchedulerCache
}

// NewTopoSchedulerPriority return a new priority scheduler
func NewTopoSchedulerPriority(Name string, clientset *kubernetes.Clientset, c *cache.SchedulerCache) *Priority {
	return &Priority{
		Name:   Name,
		client: clientset,
		pcache: c,
	}
}

func (p *Priority) Handler(args schedulerapi.ExtenderArgs) *schedulerapi.HostPriorityList {
	pod := args.Pod
	nodeNames := *args.NodeNames
	result := schedulerapi.HostPriorityList{}

	gpuTopoNum := utils.GetGPUTopoNum(pod)

	for _, nodeName := range nodeNames {
		score, err := p.makeScore(pod, nodeName, gpuTopoNum)
		if err != nil {
			klog.Errorf("Failed to count the score of node[%s]: %v", nodeName, err)
			continue
		}
		result = append(result, schedulerapi.HostPriority{nodeName, score})
	}

	return &result
}

func (p *Priority) makeScore(pod *v1.Pod, nodeName string, num int64) (int, error) {
	node, err := p.pcache.GetNodeInfo(nodeName)
	if err != nil {
		return -1, err
	}

	return node.MakeScore(pod, num)
}

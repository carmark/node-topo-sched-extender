package cache

import (
	"strings"
	"sync"

	"k8s.io/api/core/v1"
	"k8s.io/klog"

	"github.com/gpucloud/node-topology-manager/pkg/utils"
)

const (
	OptimisticLockErrorMsg = "the object has been modified; please apply your changes to the latest version and try again"
)

// NodeInfo is node level aggregated information.
type NodeInfo struct {
	name     string
	node     *v1.Node
	topology *Topology
	devs     map[string]*v1.Pod
	rwmu     *sync.RWMutex
}

// NewNodeInfo Create Node Level
func NewNodeInfo(node *v1.Node) *NodeInfo {
	klog.V(2).Infof("NewNodeInfo() creates nodeInfo for %s", node.Name)

	devs := map[string]*v1.Pod{}
	// Get Node Topology information
	topo := &Topology{}

	return &NodeInfo{
		name:     node.Name,
		node:     node,
		topology: topo,
		devs:     devs,
		rwmu:     new(sync.RWMutex),
	}
}

// GetName get node name
func (n *NodeInfo) GetName() string {
	return n.name
}

// GetNode get *v1.Node
func (n *NodeInfo) GetNode() *v1.Node {
	return n.node
}

func (n *NodeInfo) removePod(pod *v1.Pod) {
	n.rwmu.Lock()
	defer n.rwmu.Unlock()

	uids := utils.GetGPUIDFromAnnotation(pod)
	if len(uids) > 0 {
		for _, uid := range strings.Split(uids, ",") {
			_, found := n.devs[uid]
			if !found {
				klog.Warningf("Pod %s in ns %s failed to find the GPU[%s] in node %s", pod.Name, pod.Namespace, uid, n.name)
			} else {
				delete(n.devs, uid)
			}
		}
	} else {
		klog.Warningf("Pod %s in ns %s is not set the GPU[%s] in node %s", pod.Name, pod.Namespace, uids, n.name)
	}
}

// addOrUpdatePod Add the Pod which has the GPU id to the node
func (n *NodeInfo) addOrUpdatePod(pod *v1.Pod) (added bool) {
	n.rwmu.Lock()
	defer n.rwmu.Unlock()

	uids := utils.GetGPUIDFromAnnotation(pod)
	klog.V(2).Infof("Pod %s in ns %s with the GPUs[%s] should be added to device map", pod.Name, pod.Namespace, uids)
	if len(uids) > 0 {
		for _, uid := range strings.Split(uids, ",") {
			_, found := n.devs[uid]
			if !found {
				klog.Warningf("Pod %s in ns %s failed to find the GPU[%s] in node %s", pod.Name, pod.Namespace, uid, n.name)
			} else {
				n.devs[uid] = pod
				added = true
			}
		}
	} else {
		klog.Warningf("Pod %s in ns %s is not set the GPU ID%v in node %s", pod.Name, pod.Namespace, uids, n.name)
	}
	return added
}

// MakeScore make the score for the pod on the node
func (n *NodeInfo) MakeScore(pod *v1.Pod, gpuTopoNum int64) (int, error) {
	// make sure we do the score for 2^n allocation
	if gpuTopoNum&(gpuTopoNum-1) == 0 {
		return -1, nil
	}
	if gpuTopoNum == 1 {
		return 1, nil
	}
	var score int
	var num int
	for i, d := range n.topology.GPUDevice {
		if _, ok := n.devs[d.UUID]; ok {
			// IN USE
			continue
		}
		num++
		for j := i + 1; j < len(d.Topology); j++ {
			score += d.Topology[j].Link.Score()
		}
	}

	return score / num, nil
}

package cache

import (
	"encoding/json"
	"sync"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog"

	"github.com/gpucloud/node-topology-manager/pkg/utils"
)

type SchedulerCache struct {

	// a map from pod key to podState.
	nodes map[string]*NodeInfo

	// nodeLister can list/get nodes from the shared informer's store.
	nodeLister corelisters.NodeLister

	// podLister can list/get pods from the shared informer's store.
	podLister corelisters.PodLister

	// record the knownPod, it will be added when annotation ALIYUN_GPU_ID is added, and will be removed when complete and deleted
	knownPods map[types.UID]*v1.Pod
	nLock     *sync.RWMutex
}

func NewSchedulerCache(nLister corelisters.NodeLister, pLister corelisters.PodLister) *SchedulerCache {
	return &SchedulerCache{
		nodes:      make(map[string]*NodeInfo),
		nodeLister: nLister,
		podLister:  pLister,
		knownPods:  make(map[types.UID]*v1.Pod),
		nLock:      new(sync.RWMutex),
	}
}

// BuildCache Build cache when initializing
func (cache *SchedulerCache) BuildCache() error {
	klog.V(2).Infof("begin to build scheduler cache")
	nodes, err := cache.nodeLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("Failed to list node list: %v", err)
		return err
	}
	for _, node := range nodes {
		var t Topology
		if val, ok := node.Annotations["nvidia.com/gpu-topo"]; !ok {
			continue
		} else {
			err = json.Unmarshal([]byte(val), &t)
			if err != nil {
				klog.Errorf("Failed to decode node's topology: %v", err)
				continue
			}
		}
		if err = cache.AddOrUpdateNode(node.Name, &t); err != nil {
			klog.Errorf("Failed to AddOrUpdateNode: %v", node.Name)
			return err
		}
	}
	pods, err := cache.podLister.List(labels.Everything())
	if err != nil {
		return err
	}
	for _, pod := range pods {
		if len(pod.Spec.NodeName) == 0 {
			continue
		}

		err = cache.AddOrUpdatePod(pod)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cache *SchedulerCache) GetPod(name, namespace string) (*v1.Pod, error) {
	return cache.podLister.Pods(namespace).Get(name)
}

// KnownPod Get known pod from the pod UID
func (cache *SchedulerCache) KnownPod(podUID types.UID) bool {
	cache.nLock.RLock()
	defer cache.nLock.RUnlock()

	_, found := cache.knownPods[podUID]
	return found
}

// AddOrUpdatePod add/update pod
func (cache *SchedulerCache) AddOrUpdatePod(pod *v1.Pod) error {
	klog.V(2).Infof("Add or update pod info: %v", pod)
	klog.V(2).Infof("Node %v", cache.nodes)
	if len(pod.Spec.NodeName) == 0 {
		klog.V(2).Infof("pod %s in ns %s is not assigned to any node, skip", pod.Name, pod.Namespace)
		return nil
	}

	n, err := cache.GetNodeInfo(pod.Spec.NodeName)
	if err != nil {
		return err
	}
	podCopy := pod.DeepCopy()
	if n.addOrUpdatePod(podCopy) {
		// put it into known pod
		cache.rememberPod(pod.UID, podCopy)
	} else {
		klog.V(2).Infof("Pod %s in ns %s's gpu id is %d, it's illegal, skip",
			pod.Name,
			pod.Namespace,
			utils.GetGPUIDFromAnnotation(pod))
	}

	return nil
}

// RemovePod remove pod from scheduler cache
// The lock is in cacheNode
func (cache *SchedulerCache) RemovePod(pod *v1.Pod) {
	klog.V(2).Infof("Remove pod info: %v", pod)
	klog.V(2).Infof("Node %v", cache.nodes)
	n, err := cache.GetNodeInfo(pod.Spec.NodeName)
	if err == nil {
		n.removePod(pod)
	} else {
		klog.V(2).Infof("debug: Failed to get node %s due to %v", pod.Spec.NodeName, err)
	}

	cache.forgetPod(pod.UID)
}

func (cache *SchedulerCache) AddOrUpdateNode(name string, t *Topology) error {
	node, err := cache.nodeLister.Get(name)
	if err != nil {
		return err
	}

	cache.nLock.Lock()
	defer cache.nLock.Unlock()

	n, ok := cache.nodes[name]
	if !ok {
		n = NewNodeInfo(node)
		n.topology = t
		cache.nodes[name] = n
	} else {
		cache.nodes[name].topology = t
	}
	return nil
}

// GetNodeInfo Get or build nodeInfo if it doesn't exist
func (cache *SchedulerCache) GetNodeInfo(name string) (*NodeInfo, error) {
	node, err := cache.nodeLister.Get(name)
	if err != nil {
		return nil, err
	}

	cache.nLock.Lock()
	defer cache.nLock.Unlock()
	n, ok := cache.nodes[name]

	if !ok {
		n = NewNodeInfo(node)
		cache.nodes[name] = n
	} else {
		n = NewNodeInfo(node)
		cache.nodes[name] = n

		klog.V(2).Infof("debug: GetNodeInfo() uses the existing nodeInfo for %s", name)
	}
	return n, nil
}

func (cache *SchedulerCache) forgetPod(uid types.UID) {
	cache.nLock.Lock()
	defer cache.nLock.Unlock()
	delete(cache.knownPods, uid)
}

func (cache *SchedulerCache) rememberPod(uid types.UID, pod *v1.Pod) {
	cache.nLock.Lock()
	defer cache.nLock.Unlock()
	cache.knownPods[pod.UID] = pod
}

package scheduler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"k8s.io/klog"

	"github.com/gpucloud/node-topology-manager/pkg/cache"
)

func (p *Priority) NodeTopoHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var (
		err error
		t   cache.Topology
	)
	defer func() {
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			errMsg := fmt.Sprintf("{'error':'%v'}", err)
			w.Write([]byte(errMsg))
		}
	}()

	if err = json.NewDecoder(r.Body).Decode(&t); err != nil {
		klog.Errorf("Failed to parse request due to error %v", err)
		return
	}
	klog.V(2).Infof("NodeTopoHandler: Topology = %v", t)

	if p.pcache == nil {
		klog.Errorf("Priority's cache is nil")
		return
	}
	var name string = ps.ByName("name")
	err = p.pcache.AddOrUpdateNode(name, &t)
	if err != nil {
		klog.Errorf("Failed to AddOrUpdatePod with node[%v]: %v", name, err)
	}
	return
}

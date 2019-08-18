package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"k8s.io/klog"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"

	"github.com/gpucloud/node-topology-manager/pkg/scheduler"
)

const (
	apiPrefix      = "/topo-scheduler"
	priorityPrefix = apiPrefix + "/priority"
)

var (
	version = "0.1.0"
)

func checkBody(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
}

func PriorityRoute(priority *scheduler.Priority) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		checkBody(w, r)

		var buf bytes.Buffer
		body := io.TeeReader(r.Body, &buf)

		var extenderArgs schedulerapi.ExtenderArgs
		var hostPriorityList *schedulerapi.HostPriorityList

		if err := json.NewDecoder(body).Decode(&extenderArgs); err != nil {
			klog.Warningf("Failed to parse request due to error %v", err)
			hostPriorityList = &schedulerapi.HostPriorityList{}
		} else {
			klog.V(2).Infof("gpu-topo-priority ExtenderArgs =%v", extenderArgs)
			hostPriorityList = priority.Handler(extenderArgs)
		}

		if resultBody, err := json.Marshal(hostPriorityList); err != nil {
			klog.Warningf("Failed due to %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			errMsg := fmt.Sprintf("{'error':'%v'}", err)
			w.Write([]byte(errMsg))
		} else {
			klog.Info(priority.Name, " extenderFilterResult = ", string(resultBody))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(resultBody)
		}
	}
}

func DebugLogging(h httprouter.Handle, path string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		klog.Info("debug: ", path, " request body = ", r.Body)
		h(w, r, p)
		klog.Info("debug: ", path, " response=", w)
	}
}

func AddPriority(router *httprouter.Router, priority *scheduler.Priority) {
	router.POST(priorityPrefix, DebugLogging(PriorityRoute(priority), priorityPrefix))
}

func AddNodeTopo(router *httprouter.Router, s *scheduler.Priority) {
	router.POST("/nodes/:name", DebugLogging(s.NodeTopoHandler, "/nodes"))
}

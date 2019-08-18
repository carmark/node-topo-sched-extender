package routes

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func AddVersion(r *httprouter.Router) {
	r.GET("/version", handleVersion)
}

// handleVersion writes the server's version information.
func handleVersion(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
}

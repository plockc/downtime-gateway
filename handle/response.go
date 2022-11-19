package handle

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func errorResponse(w http.ResponseWriter, path string, code int, err error) {
	jsonResponse(w, path, code, map[string]string{"error": err.Error()})
	log.Printf("Error %d at %s: %s\n", code, path, err.Error())
}

func jsonResponse(w http.ResponseWriter, path string, code int, data any) {
	response, err := json.Marshal(data)
	if err != nil {
		code = 500
		response = []byte(fmt.Sprintf(
			`{"error":"Failed to convert data to json for request at '%s' with code %d: %v"}`,
			path, code, err,
		))
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(response)
	if err != nil {
		log.Printf(
			"Failed to respond to request at '%s' with code %d: %v",
			path, code, err,
		)
	}
}

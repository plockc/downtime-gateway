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

func locationResponse(w http.ResponseWriter, path string, code int, location string) {
	jsonWithHeadersResponse(w, path, code, map[string]any{"Location": location}, nil)
}

func jsonWithHeadersResponse(w http.ResponseWriter, path string, code int, headers map[string]any, body []byte) {
	for h, v := range headers {
		w.Header().Add(h, fmt.Sprintf("%v", v))
		w.WriteHeader(code)
	}
	_, err := w.Write(body)
	if err != nil {
		log.Printf(
			"Failed to respond to request at '%s' with code %d: %v",
			path, code, err,
		)
	}
}

func jsonResponse(w http.ResponseWriter, path string, code int, data any) {
	var response []byte
	var err error
	response, err = json.Marshal(data)
	if err != nil {
		code = 500
		response = []byte(fmt.Sprintf(
			`{"error":"Failed to convert data to json for request at '%s' with code %d: %v"}`,
			path, code, err,
		))
	}
	// if there is no elements marshalled, still have an empty object
	if string(response) == "{}" {
		response = nil
	}
	fmt.Println("RESPONSING", path, data, string(response), fmt.Sprintf("%T", data))
	jsonWithHeadersResponse(w, path, code, map[string]any{"content-type": "application/json"}, response)
}

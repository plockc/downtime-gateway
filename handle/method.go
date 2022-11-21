package handle

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"golang.org/x/exp/maps"
)

type MethodHandlers map[string]TypedHandler
type Handler func(parts []string, data any) (int, any, error)

type TypedHandler struct {
	Handler
	T reflect.Type
}

func (mhs MethodHandlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts, ok := getURLParts(r.URL.Path)
	if !ok {
		errorResponse(w, r.URL.Path, http.StatusBadRequest, fmt.Errorf(
			"url is not acceptable, received '%s'",
			r.URL.Path,
		))
		return
	}
	methodHandler, ok := mhs[r.Method]
	if !ok {
		errorResponse(w, r.URL.Path, http.StatusMethodNotAllowed, fmt.Errorf(
			"method '%v' is not allowed, allowed: %v", r.Method, maps.Keys(mhs),
		))
		return
	}

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errorResponse(w, r.URL.Path, http.StatusInternalServerError, err)
		return
	}
	ct := r.Header.Get("content-type")
	if ct != "application/json" {
		fmt.Println(r.Header)
		errorResponse(w, r.URL.Path, http.StatusUnsupportedMediaType, fmt.Errorf(
			"content type 'application/json' required",
		))
		return
	}

	dataPtr := reflect.New(methodHandler.T).Interface()
	if len(body) > 0 {
		err = json.Unmarshal(body, dataPtr)
		if err != nil {
			errorResponse(w, r.URL.Path, http.StatusInternalServerError, fmt.Errorf(
				"must be json: %w", err),
			)
			return
		}
	}

	data := reflect.ValueOf(dataPtr).Elem().Interface()
	code, respData, err := methodHandler.Handler(parts[1:], data)
	if err != nil {
		errorResponse(w, r.URL.Path, code, err)
		return
	}

	jsonResponse(w, r.URL.Path, http.StatusOK, respData)
}

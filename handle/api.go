package handle

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/plockc/gateway/resource"
	"golang.org/x/exp/slices"
)

type Api struct {
}

func (api Api) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	ct := req.Header.Get("content-type")
	if ct != "application/json" {
		fmt.Println(req.Header)
		errorResponse(w, path, http.StatusUnsupportedMediaType, fmt.Errorf(
			"content type 'application/json' required",
		))
		return
	}

	if !strings.HasPrefix(path, "/api") {
		errorResponse(w, path, http.StatusBadRequest, fmt.Errorf(
			"url is not acceptable, received '%s'", path,
		))
		return
	}

	// remove /api then remove / from beginning and end then split on path elements
	parts := strings.Split(
		strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(
			path, "/api"), "/"), "/"), "/",
	)

	ids := []string{}
	for i, p := range parts {
		if i%2 == 0 {
			ids = append(ids, p)
		}
	}

	handler := Versions
	// inside the loop can handle i==len(parts), which is a list request
	// increment by two as relationship require path with id of parent + relationship name
	for i := 0; i <= len(parts); i += 2 {
		// factories just gathers the the number of IDs it needs to construct
		res, err := handler.Factory(nil, ids...)
		if err != nil {
			errorResponse(w, req.URL.Path, http.StatusMethodNotAllowed, err)
			return
		}
		lc := resource.Lifecycle{Resource: res}
		switch {
		// case: no ID for the requested resource, it's a GET-list() or DELETE-clear() request
		// e.g. /api/v1/ns/test/ipsets/tvs (6 parts, so %2 == 0)
		case i == len(parts) && (len(parts)%2 == 0):
			fmt.Println("---- LIST ----")
			switch req.Method {
			// handle a GET request for a list - e.g. GET /api/v1/ns/test/ipsets/tvs
			case http.MethodGet:
				if !slices.Contains(handler.Allowed, LIST_ALLOWED) {
					errorResponse(w, path, http.StatusMethodNotAllowed, fmt.Errorf(
						"method '%v' is not allowed, allowed: %v", req.Method, handler.Allowed,
					))
					return
				}
				list, err := res.List()
				if err != nil {
					errorResponse(w, path, http.StatusInternalServerError, fmt.Errorf(
						"failed to list: %w", err,
					))
					return
				}
				jsonResponse(w, path, 200, list)
			// handle a DELETE request for a list - e.g. GET /api/v1/ns/test/ipsets/tvs
			case http.MethodDelete:
				if !slices.Contains(handler.Allowed, DELETE_ALLOWED) {
					errorResponse(w, path, http.StatusMethodNotAllowed, fmt.Errorf(
						"method '%v' is not allowed, allowed: %v", req.Method, handler.Allowed,
					))
					return
				}
				cleared, err := lc.EnsureCleared()
				if err != nil {
					errorResponse(w, path, http.StatusInternalServerError, fmt.Errorf(
						"failed to clear: %w", err,
					))
					return
				}
				if cleared {
					jsonResponse(w, path, 204, nil)
				} else {
					jsonResponse(w, path, 200, nil)
				}
			default:
				errorResponse(w, path, http.StatusMethodNotAllowed, err)
			}
		// case: has ID for the requested resource
		// e.g. /api/v1/ns/test
		case i == len(parts)-1:
			fmt.Println("---- SPECIFIC RESOURCE", req.Method, "----")
			// Load the Body using the type specified in the handler
			// this will be used for PUT / PATCH / POST
			defer req.Body.Close()

			//body, err := ioutil.ReadAll(req.Body)
			//if err != nil {
			//	errorResponse(w, path, http.StatusInternalServerError, err)
			//	return
			//}
			//res, err = handler.Factory([]byte(body), ids...)
			switch req.Method {
			// handle a GET request for a resource - e.g. GET /api/v1/ns/test
			case http.MethodGet:
				if !slices.Contains(handler.Allowed, GET_ALLOWED) {
					errorResponse(w, path, http.StatusMethodNotAllowed, fmt.Errorf(
						"method '%v' is not allowed", req.Method,
					))
					return
				}
				exists, err := lc.Exists()
				if err != nil {
					errorResponse(w, path, http.StatusInternalServerError, fmt.Errorf(
						"failed to get: %w", err,
					))
					return
				}
				if exists {
					jsonResponse(w, path, 200, nil)
				} else {
					errorResponse(w, path, 404, fmt.Errorf("missing %s", path))
				}
			// handle a DELETE request for a resource - e.g. DELETE /api/v1/ns/test
			case http.MethodDelete:
				if !slices.Contains(handler.Allowed, DELETE_ALLOWED) {
					errorResponse(w, path, http.StatusMethodNotAllowed, fmt.Errorf(
						"method '%v' is not allowed", req.Method,
					))
					return
				}
				deleted, err := lc.EnsureDeleted()
				if err != nil {
					errorResponse(w, path, http.StatusInternalServerError, fmt.Errorf(
						"failed to delete: %w", err,
					))
					return
				}
				if deleted {
					jsonResponse(w, path, 204, nil)
				} else {
					jsonResponse(w, path, 200, nil)
				}
			// handle a PUT request for a new resource - e.g. PUT /api/v1/ns/test/ipsets/tvs
			case http.MethodPut:
				if !slices.Contains(handler.Allowed, UPSERT_ALLOWED) {
					errorResponse(w, path, http.StatusMethodNotAllowed, fmt.Errorf(
						"method '%v' is not allowed for %s", req.Method, handler.Name,
					))
					return
				}
				created, err := lc.Ensure()
				if err != nil {
					errorResponse(w, path, http.StatusInternalServerError, fmt.Errorf(
						"failed to ensure: %w", err,
					))
					return
				}
				if created {
					jsonResponse(w, path, 201, nil)
				} else {
					jsonResponse(w, path, 200, nil)
				}
			default:
				errorResponse(w, path, http.StatusMethodNotAllowed, fmt.Errorf(
					"Method "+req.Method+" is not allowed",
				))
			}
		// default: there is a relationship to traverse
		default:
			relation := parts[i+1]
			// paths[0] is the ID of the current handler, so next is the relationship
			var ok bool
			handler, ok = handler.Relationships[relation]
			if !ok {
				errorResponse(w, path, http.StatusBadRequest, fmt.Errorf(
					"relationship %s does not exist", relation,
				))
				return
			}
		}
	}
}

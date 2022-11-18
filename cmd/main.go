package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/plockc/gateway"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

var urlRegex *regexp.Regexp
var Runner = &gateway.Runner{}

func init() {
	urlRegex = regexp.MustCompile(`^/v1/[^/]+(/?$)|(/([^/]+))$`)
	//urlRegex = regexp.MustCompile(`^/v1/[^/]+(/?|(/[^/]+))$`)
}

// getId returns empty string if there was no Id
func getId(url string) (id string, ok bool) {
	// FindStringSubmatch returns the whole match then the group matches in return slice
	matches := urlRegex.FindStringSubmatch(url)
	if matches == nil {
		return "", false
	}
	return matches[3], true
}

func main() {
	if _, out, err := gateway.ExecLine("id -u"); err != nil {
		fmt.Println("Could not determine user id: " + err.Error())
		os.Exit(1)
	} else if out != "0" {
		fmt.Println("must be run as root")
		os.Exit(1)
	}
	handleRequests()
}

func handleRequests() {
	http.Handle("/v1/sets/", setsHandlers)
	http.Handle("/v1/sets", setsHandlers)
	fmt.Println("Listening on :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

type methodHandlers map[string]methodHandler

type methodHandler struct {
	handler func(id string, data any) (int, any, error)
	T       reflect.Type
}

func (mhs methodHandlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, ok := getId(r.URL.Path)
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
			"Method '%s' is not allowed, allowed: %v", r.Method, maps.Keys(mhs),
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

	//dataZeroVal := reflect.Zero(methodHandler.T)
	//dataPtrVal := reflect.Zero(reflect.PointerTo(methodHandler.T))
	//dataPtrVal.Set(dataZeroVal.Addr())
	//dataPtr := dataPtrVal.Interface()

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
	code, respData, err := methodHandler.handler(id, data)
	if err != nil {
		errorResponse(w, r.URL.Path, code, err)
		return
	}

	jsonResponse(w, r.URL.Path, http.StatusOK, respData)
}

var setsHandlers = methodHandlers{
	http.MethodGet: methodHandler{
		handler: setGet,
		T:       reflect.TypeOf(gateway.IPSet{}),
	},
	http.MethodPost: methodHandler{
		handler: setPost,
		T:       reflect.TypeOf(gateway.IPSet{}),
	},
	//http.MethodPatch:  setPatch,
	//http.MethodPut:    setPatch,
	http.MethodDelete: methodHandler{
		handler: setDelete,
		T:       reflect.TypeOf(gateway.IPSet{}),
	},
}

func setDelete(id string, ipSetIface any) (int, any, error) {
	err := Runner.Line("ipset destroy " + id)
	if err != nil {
		return 500, nil, fmt.Errorf("%w %s", err, Runner.LastOut())
	}
	return 200, nil, nil
}

func setPost(_ string, ipSetIface any) (int, any, error) {
	ipSet := ipSetIface.(gateway.IPSet)
	if ipSet.Name == "" {
		return 400, nil, fmt.Errorf("cannot create IPSet with empty name")
	}
	err := Runner.Line("ipset -N " + ipSet.Name + " hash:mac")
	if err != nil {
		return 500, nil, fmt.Errorf("%w %s", err, Runner.LastOut())
	}
	// TODO: handle MACs
	return 200, ipSet, nil
}

func ipSets() ([]string, error) {
	err := Runner.Line("ipset list -n")
	if err != nil {
		return nil, fmt.Errorf("failed to list ipsets: %w", err)
	}
	return strings.Split(Runner.LastOut(), "\n"), nil
}

func ipSetExists(name string) (bool, error) {
	sets, err := ipSets()
	if err != nil {
		return false, err
	}
	return slices.Contains(sets, name), nil
}

func setGet(id string, _ any) (int, any, error) {
	sets, err := ipSets()
	if err != nil {
		return 500, nil, fmt.Errorf("failed to list ipsets: %w", err)
	}
	if id == "" {
		return 200, sets, nil
	}
	if !slices.Contains(sets, id) {
		return 400, nil, fmt.Errorf("cannot list ipset " + id + " that does not exist")
	}
	err = Runner.Line("ipset save " + id)
	if err != nil {
		return 500, nil, fmt.Errorf("failed to list ipset %s: %w", id, err)
	}
	elems := gateway.Keep(strings.Split(Runner.LastOut(), "\n"), func(s string) bool {
		return strings.HasPrefix(s, "add ")
	})
	elems = gateway.Map(elems, func(s string) string {
		return strings.TrimPrefix(s, "add "+id+" ")
	})
	macs, err := gateway.MapErrable(elems, func(s string) (gateway.MAC, error) {
		return gateway.MACFromString(s)
	})
	if err != nil {
		return 500, nil, fmt.Errorf("failed to parse macs: %w", err)
	}
	return 200, gateway.IPSet{Name: id, Members: macs}, nil
}

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

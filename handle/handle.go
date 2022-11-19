package handle

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/plockc/gateway"
)

var urlRegex *regexp.Regexp
var Runner = &gateway.Runner{}

func init() {
	urlRegex = regexp.MustCompile(`^/v1/[^/]+(/?$)|(/.+)$`)
}

// getURLParts splits the url path after the object type
func getURLParts(url string) (parts []string, ok bool) {
	// FindStringSubmatch returns the whole match then the group matches in return slice
	matches := urlRegex.FindStringSubmatch(url)
	if matches == nil {
		return nil, false
	}
	return strings.Split(matches[2], "/"), true
}

func Serve() {
	http.Handle("/v1/sets/", SetsHandlers)
	http.Handle("/v1/sets", SetsHandlers)
	fmt.Println("Listening on :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

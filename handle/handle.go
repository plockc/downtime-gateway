package handle

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/plockc/gateway/resource"
)

var NS = resource.NewNS("")

func Serve() {
	http.Handle("/api/", Api{})
	fmt.Println("Listening on :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func UpdateFromJson(body []byte, target any) error {
	if len(body) > 0 {
		return json.Unmarshal(body, target)
	}
	return nil
}

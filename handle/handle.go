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

func FromJson[T any](body []byte, target *T) error {
	if len(body) > 0 {
		err := json.Unmarshal(body, target)
		if err != nil {
			return fmt.Errorf("must be json: %w", err)
		}
	}
	return nil
}

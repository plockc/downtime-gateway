package handle_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/plockc/gateway/funcs"
	"github.com/plockc/gateway/handle"
	"github.com/plockc/gateway/iptables"
	"github.com/plockc/gateway/resource"
)

var (
	testNS = resource.NewNS("test")
)

func Failf(t *testing.T, errFmt string, args ...any) {
	namespaces, err := resource.NewNS("").NSResource().List()
	if err != nil {
		panic("could not list namespaces")
	}
	for _, ns := range namespaces {
		sets, err := iptables.NewIPSet(resource.NewNS(ns), "").NSResource().List()
		if err != nil {
			panic("could not list ip sets for ns " + string(ns))
		}
		fmt.Printf("Namespace '%s' has IP Sets: %v\n", ns, sets)
	}
	t.Log(string(debug.Stack()))
	t.Fatalf(errFmt, args...)
}

type TestResponseWriter struct {
	Body []byte
	Code int
}

func (w *TestResponseWriter) Header() http.Header {
	return http.Header{}
}

func (w *TestResponseWriter) Write(b []byte) (int, error) {
	w.Body = b
	return len(b), nil
}

func (w *TestResponseWriter) WriteHeader(statusCode int) {
	w.Code = statusCode
}

func AssertHandler[T any](t *testing.T, method, path string, bodyObj any, expectedCode int) *T {
	resp, err := testRequest[T](t, method, path, bodyObj, expectedCode)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func testRequest[T any](t *testing.T, method, path string, bodyObj any, expectedCode int) (*T, error) {
	url, err := url.Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	reqBody, err := json.Marshal(bodyObj)
	if err != nil {
		t.Fatal(err)
	}
	header := http.Header{}
	header.Set("Content-Type", "application/json")
	req := http.Request{
		Method: method,
		URL:    url,
		Body:   ioutil.NopCloser(strings.NewReader(string(reqBody))),
		Header: header,
	}
	responseWriter := &TestResponseWriter{}
	handle.Api{}.ServeHTTP(responseWriter, &req)

	if responseWriter.Code != expectedCode {
		Failf(
			t, "expected code %d but got code %d and body: %s",
			expectedCode, responseWriter.Code, string(responseWriter.Body),
		)
	}
	var responseObj *T
	if len(responseWriter.Body) > 0 {
		if expectedCode >= 400 {
			var errData map[string]interface{}
			err := json.Unmarshal(responseWriter.Body, &errData)
			if err != nil {
				t.Fatal(err)
			}
			msg, ok := errData["error"]
			if !ok {
				t.Fatalf("missing error message from %v", errData)
			}
			if len(errData) != 1 {
				t.Fatalf("expected only error in response body, got %v", errData)
			}
			return nil, fmt.Errorf("%v", msg)
		} else {
			if err := handle.FromJson(responseWriter.Body, &responseObj); err != nil {
				t.Fatal(err)
			}
		}
	}
	return responseObj, nil
}

func AssertHandlerFail(t *testing.T, method, path string, bodyObj any, expectedCode int) {
	_, err := testRequest[any](t, method, path, bodyObj, expectedCode)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMain(m *testing.M) {
	// it is the internal client outbound that can get blocked for downtime
	exitCode := func() int {
		testRunner := testNS.Runner()
		testNSLifecycle := resource.Lifecycle{Resource: testNS.NSResource()}
		handle.NS = testNS

		testNSLifecycle.EnsureDeleted()

		// every one of these functions can error, use Do to execute, stopping if any fails
		var ignored bool
		if err := funcs.Do(
			funcs.AssignFunc(testNSLifecycle.EnsureDeleted, &ignored),
			funcs.AssignFunc(testNSLifecycle.Ensure, &ignored),
		); err != nil {
			fmt.Println(testRunner)
			fmt.Println(err)
			return 1
		}

		return m.Run()
	}()

	os.Exit(exitCode)
}

package pkg

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"testing"

	httpclient "github.com/digitalocean/do-dcgm-exporter/pkg/client"
)

const (
	statusNotAllowed     = "405 method not allowed"
	statusCodeNotAllowed = 405
)

func TestForwardMetricsToProxy(t *testing.T) {
	var testBuffer bytes.Buffer
	testRequestBody := "Test Body"
	testBuffer.Write([]byte(testRequestBody))

	returnHttpSuccess := func(request *http.Request) (*http.Response, error) {
		validateRequest(request, t, testBuffer)

		resp := &http.Response{
			Status:     "200 SUCCESSFUL",
			StatusCode: 200,
		}
		return resp, nil
	}

	returnHttpFailure := func(request *http.Request) (*http.Response, error) {
		validateRequest(request, t, testBuffer)
		resp := &http.Response{
			Status:     statusNotAllowed,
			StatusCode: statusCodeNotAllowed,
		}
		return resp, nil
	}

	const expectedErrorMessage = "internal error"
	returnError := func(request *http.Request) (*http.Response, error) {
		validateRequest(request, t, testBuffer)
		return nil, errors.New(expectedErrorMessage)
	}

	var tests = []struct {
		name        string
		clientDo    func(request *http.Request) (*http.Response, error)
		returnError bool
		errorMsg    string
	}{
		{"Expect successful request", returnHttpSuccess, false, ""},
		{"Expect http request failure", returnHttpFailure, true, fmt.Sprintf("failed to forward metrics to proxy. Got status: %d(%q)", statusCodeNotAllowed, statusNotAllowed)},
		{"Expect internal error", returnError, true, fmt.Sprintf("failed to forward metrics to proxy: %s", expectedErrorMessage)},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("TestForwardMetricsToProxy: %s", tt.name)
		t.Run(testname, func(t *testing.T) {
			agent := GPUMetricsAgent{
				ProxyClient: &httpclient.FakeHTTPClient{DoFunc: tt.clientDo},
			}

			err := agent.forwardMetricsToProxy(&testBuffer)
			if err != nil {
				if tt.returnError == false {
					t.Errorf("expected no error, but got: %s", err.Error())
				}

				if err.Error() != tt.errorMsg {
					t.Errorf("expected error message %q, but got: %s", tt.errorMsg, err.Error())
				}
			}
		})
	}
}

func validateRequest(request *http.Request, t *testing.T, testBuffer bytes.Buffer) {
	if request.Method != "POST" {
		t.Errorf("expected request method of type POST, but got: %s", request.Method)
	}

	if request.URL == nil {
		t.Errorf("expected request to have set an URL")
	}

	if request.URL.Path != "/v1/gpu_metrics" {
		t.Errorf("expected path to be /v1/gpu_metrics, but got: %s", request.URL.Path)
	}

	if request.URL.Host != "169.254.169.254:80" {
		t.Errorf("expected host to be %s, but got: %s", "169.254.169.254:80", request.URL.Host)
	}

	if request.ContentLength != int64(testBuffer.Len()) {
		t.Errorf("expected content length to be %d, but got: %d", testBuffer.Len(), request.ContentLength)
	}
}

package pkg

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (a GPUMetricsAgent) forwardMetricsToProxy(requestBody *bytes.Buffer) error {
	// "http://169.254.169.254:80/v1/gpu_metrics"
	url := fmt.Sprintf("%s:%d/%s", internalProxyURL, internalProxyPort, internalProxyPath)

	req, err := http.NewRequest("POST", url, bytes.NewReader(requestBody.Bytes()))
	if err != nil {
		return errors.Wrap(err, "failed to construct POST request to proxy")
	}

	resp, err := a.ProxyClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to forward metrics to proxy")
	}
	defer func(res *http.Response) {
		if res.Body != nil {
			res.Body.Close()
		}
	}(resp)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return errors.Errorf("failed to forward metrics to proxy. Got status: %d(%q)", resp.StatusCode, resp.Status)
	}

	logrus.Debugf("Response from proxy has status code: %d", resp.StatusCode)
	return nil
}

package utils

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"time"
)

func NewLogHttpClient(logger *logrus.Logger) *http.Client {
	httpClient := &http.Client{
		Timeout: time.Minute,
	}

	if logger == nil {
		logger = logrus.New()
	}

	httpClient.Transport = &loggingTransport{
		logger:    logger,
		transport: http.DefaultTransport,
	}

	return httpClient
}

type loggingTransport struct {
	transport http.RoundTripper
	logger    *logrus.Logger
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	shouldLog := t.logger != nil && t.logger.IsLevelEnabled(logrus.TraceLevel)

	if shouldLog {
		// Dump the request and log it
		requestDump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			t.logger.Errorf("Error dumping request: %v", err)
		} else {
			fmt.Println("Request:\n")
			fmt.Println(string(requestDump))
		}
	}

	// Make the HTTP request using the underlying transport
	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if shouldLog {
		// Dump the response and log it
		responseDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			t.logger.Errorf("Error dumping response: %v", err)
		} else {
			fmt.Println("Response:\n")
			fmt.Println(string(responseDump))
		}
	}

	return resp, nil
}

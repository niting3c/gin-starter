package services

import (
	"fmt"
	"io"
	"net/http"
	"starter/internal/app/utils"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	maxRetries     = 5
	initialBackoff = 1 * time.Second
	clientTimeout  = 30 * time.Second
)

// RestCaller defines the interface for making REST calls.
//
//go:generate mockery --name RestCaller
type RestCaller interface {
	Get(url string) (*http.Response, error)
	MakeRestCallToPartner(url string, params string) *utils.ErrorMessage
}

// DefaultRestCaller is the default implementation of RestCaller using http.Client.
type restCaller struct {
	client *http.Client
}

// NewDefaultRestCaller creates a new DefaultRestCaller instance.
func NewDefaultRestCaller() RestCaller {
	return &restCaller{
		client: &http.Client{
			Timeout: clientTimeout,
		},
	}
}

// Get makes a GET request using the underlying http.Client.
func (rc *restCaller) Get(url string) (*http.Response, error) {
	return rc.client.Get(url)
}

// MakeRestCallToPartner makes a REST call to Partner using the provided RestCaller.
func (rc *restCaller) MakeRestCallToPartner(url string, params string) *utils.ErrorMessage {
	retries := 0
	var body []byte
	url = url + "?" + params
	logrus.Debugf("Partner URL Formed: %v", url)
	for {
		resp, restErr := rc.Get(url)
		if restErr != nil {
			if retries >= maxRetries {
				statusCode := 0
				if resp != nil {
					statusCode = resp.StatusCode
				}
				return &utils.ErrorMessage{
					Message:    fmt.Sprintf("Failed to make GET request after retries -> %v", restErr.Error()),
					StatusCode: statusCode,
				}
			}
			retries++
			time.Sleep(initialBackoff * time.Duration(retries))
			continue
		}
		statusCode := resp.StatusCode
		logrus.Debug("trying to read response body if any")
		if resp.Body != nil {
			defer resp.Body.Close()
			body, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				return &utils.ErrorMessage{
					Message:    fmt.Sprintf("Failed to read response body -> %v", readErr.Error()),
					StatusCode: statusCode,
				}
			}
			logrus.Infof("Printing any response found from Partner: %v", string(body))
		}
		if statusCode != http.StatusOK {
			return &utils.ErrorMessage{
				Message:    "Received non-OK HTTP status: " + string(body),
				StatusCode: statusCode,
			}
		}
		logrus.Info("Successfully contacted Partner To run the scripts")
		break
	}
	return nil
}

package sentinel

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	purgeStatusUpdating = "updating"
)

type purgeStatusResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	StartTime  time.Time `json:"startTime"`
	EndTime    time.Time `json:"endTime"`
	Properties struct {
		RecordCount int    `json:"RecordCount"`
		Status      string `json:"Status"`
	} `json:"properties"`
}

func (s *Sentinel) GetPurgeStatus(l *logrus.Entry, operationURL string) (string, error) {
	logger := l.WithField("module", "purge_status")

	if operationURL == "" {
		return "", fmt.Errorf("empty operation URL")
	}

	logger.Info("checking log purge status")

	req, err := http.NewRequest(http.MethodGet, operationURL, nil)
	if err != nil {
		return "", fmt.Errorf("could not create http request: %w", err)
	}

	token, err := getToken(s.azCreds)
	if err != nil {
		return "", fmt.Errorf("could not get token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not send purge status request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not read response body (status=%d): %w", resp.StatusCode, err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to retrieve purge status (status=%d): %s", resp.StatusCode, string(body))
	}

	var statusResp purgeStatusResponse
	if err := json.Unmarshal(body, &statusResp); err != nil {
		return "", fmt.Errorf("could not parse purge status response: %w", err)
	}

	if statusResp.Status == "" {
		return "", fmt.Errorf("empty purge status response")
	}

	if l.Logger.IsLevelEnabled(logrus.DebugLevel) {
		l.Debugf("purge status response: %+v", statusResp)
	}

	return strings.ToLower(statusResp.Status), nil
}

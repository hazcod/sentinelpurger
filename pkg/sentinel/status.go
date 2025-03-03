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
	PurgeStatusPending = "pending"
)

type purgeStatusResponse struct {
	Status string `json:"status"`
}

func (s *Sentinel) GetPurgeStatus(l *logrus.Entry, subscriptionID, resourceGroup, workspaceName string, purgeID string) (string, error) {
	logger := l.WithField("module", "purge_status").WithField("id", purgeID)

	if purgeID == "" {
		return "", fmt.Errorf("empty purge ID")
	}

	logger.Info("checking log purge status")

	purgeURL := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.OperationalInsights/workspaces/%s/operations/%s?api-version=2023-09-01",
		subscriptionID, resourceGroup, workspaceName, purgeID,
	)

	req, err := http.NewRequest(http.MethodGet, purgeURL, nil)
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

	return strings.ToLower(statusResp.Status), nil
}

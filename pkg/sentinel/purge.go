package sentinel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
)

func getToken(creds *azidentity.ClientSecretCredential) (string, error) {
	token, err := creds.GetToken(context.TODO(), policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})
	if err != nil {
		return "", err
	}
	return token.Token, nil
}

type purgeRequest struct {
	Filters []struct {
		Column   string `json:"column"`
		Operator string `json:"operator"`
		Value    string `json:"value"`
	} `json:"filters"`
}

func createPurgeRequest(treshold time.Time) ([]byte, error) {
	request := purgeRequest{
		Filters: []struct {
			Column   string `json:"column"`
			Operator string `json:"operator"`
			Value    string `json:"value"`
		}{
			{
				Column:   "TimeGenerated",
				Operator: "<",
				Value:    treshold.UTC().Format("2006-01-02T15:04:05Z"),
			},
		},
	}
	return json.Marshal(request)
}

func (s *Sentinel) PurgeLogs(l *logrus.Entry, subscriptionID, resourceGroup, workspaceName string, tableName string, treshold time.Time) error {
	logger := l.WithField("module", "purger")

	logger.Info("purging logs")

	purgeURL := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.OperationalInsights/workspaces/%s/tables/%s/deleteData?api-version=2023-09-01",
		subscriptionID, resourceGroup, workspaceName, tableName,
	)

	purgePayload, err := createPurgeRequest(treshold)
	if err != nil {
		return fmt.Errorf("could not create purge request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, purgeURL, bytes.NewBuffer(purgePayload))
	if err != nil {
		return fmt.Errorf("could not create http request: %w", err)
	}

	token, err := getToken(s.azCreds)
	if err != nil {
		return fmt.Errorf("could not get token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not send purge request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read response body: %w", err)
	}

	if resp.StatusCode != http.StatusAccepted {
		if l.Logger.IsLevelEnabled(logrus.DebugLevel) {
			b, _ := io.ReadAll(resp.Body)
			l.Logger.Debugf("purge response: %s", string(b))
		}
		return fmt.Errorf("failed to purge data (status=%d): %s", resp.StatusCode, string(body))
	}

	asyncOperationURL := resp.Header.Get("Azure-Asyncoperation")
	if asyncOperationURL == "" {
		return fmt.Errorf("could not parse purge response: missing 'Azure-Asyncoperation' URL header")
	}

	purgeStatus, err := s.GetPurgeStatus(l, asyncOperationURL)
	if err != nil {
		return fmt.Errorf("could not get purge status: %w", err)
	}

	logger.WithField("status", purgeStatus).Info("purge job registered")

	if purgeStatus != purgeStatusUpdating {
		return fmt.Errorf("unknown purge job status: %s", purgeStatus)
	}

	return nil
}

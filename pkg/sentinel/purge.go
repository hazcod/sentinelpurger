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
	Table   string `json:"table"`
	Filters []struct {
		Column   string `json:"column"`
		Operator string `json:"operator"`
		Value    string `json:"value"`
	} `json:"filters"`
}

func createPurgeRequest(tableName string, treshold time.Time) ([]byte, error) {
	request := purgeRequest{
		Table: tableName,
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

type purgeResponse struct {
	ID string `json:"operationId"`
}

func (s *Sentinel) PurgeLogs(l *logrus.Entry, subscriptionID, resourceGroup, workspaceName string, tableName string, treshold time.Time) error {
	logger := l.WithField("module", "purger")

	logger.Info("purging logs")

	purgeURL := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.OperationalInsights/workspaces/%s/purge?api-version=2023-09-01",
		subscriptionID, resourceGroup, workspaceName,
	)

	purgePayload, err := createPurgeRequest(tableName, treshold)
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
		return fmt.Errorf("failed to purge data (status=%d): %s", resp.StatusCode, string(body))
	}

	var purgeResponse purgeResponse
	if err := json.Unmarshal(body, &purgeResponse); err != nil {
		return fmt.Errorf("could not parse purge response: %w", err)
	}

	logger.Info("requested to purge logs")
	s.logger.Errorf("%s: %s -> %+v", tableName, purgeResponse.ID, string(purgePayload))

	purgeStatus, err := s.GetPurgeStatus(l, subscriptionID, resourceGroup, workspaceName, purgeResponse.ID)
	if err != nil {
		return fmt.Errorf("could not get purge status: %w", err)
	}

	logger.WithField("id", purgeResponse.ID).WithField("status", purgeStatus).Info("purge job registered")

	if purgeStatus != PurgeStatusPending {
		return fmt.Errorf("unknown purge job status: %s", purgeStatus)
	}

	return nil
}

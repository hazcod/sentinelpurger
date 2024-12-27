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
				Value:    treshold.UTC().Format("2006-01-02"),
			},
		},
	}
	return json.Marshal(request)
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

	if resp.StatusCode != http.StatusAccepted {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.WithError(err).Error("could not read response body")
		}
		return fmt.Errorf("failed to purge data (status=%d): %s", resp.StatusCode, string(body))
	}

	logger.Info("requested to purge logs")

	return nil
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	handler "github.com/modanisatech/grafana-webhook-handler"

	"github.com/pkg/errors"
)

func main() {
	debeziumURL := os.Getenv("DEBEZIUM_URL")
	if debeziumURL == "" {
		log.Fatal("Empty debeziumURL")
	}
	triggerPayload := os.Getenv("GRAFANA_TRIGGER_PAYLOAD")
	if triggerPayload == "" {
		log.Fatal("Empty triggerPayload")
	}

	alertingConnectors, err := getAlertingConnector(triggerPayload)
	if errors.Is(err, handler.ErrPayloadStateIsNotAlerting) {
		return
	}
	if err != nil {
		log.Fatal(err)
	}

	debeziumClient := NewDebeziumClient(debeziumURL)

	if err = alertingConnectors.applyOperation(debeziumClient); err != nil {
		log.Fatal(err)
	}
}

func (connectors AlertingConnectors) applyOperation(client DebeziumClient) error {
	operationErrors := make([]string, 0)
	for _, connector := range connectors {
		ctx := context.Background()
		taskIDs, err := client.GetConnectorFailedTaskIDs(ctx, connector)
		if err != nil {
			operationErrors = append(operationErrors, err.Error())
			continue
		}

		for _, taskID := range taskIDs {
			err = client.RestartConnectorTaskByID(ctx, connector, taskID)
			if err != nil {
				operationErrors = append(operationErrors, err.Error())
				continue
			}
		}
	}

	if len(operationErrors) != 0 {
		return fmt.Errorf(strings.Join(operationErrors, "\n"))
	}

	return nil
}

func getAlertingConnector(triggerPayload string) (AlertingConnectors, error) {
	alertingConnectors := make(AlertingConnectors, 0)

	var payload handler.WebhookPayload

	errJSONUnmarshal := json.Unmarshal([]byte(triggerPayload), &payload)
	if errJSONUnmarshal != nil {
		errJSONUnmarshal = errors.Wrapf(errJSONUnmarshal, "received payload %s", triggerPayload)
		errJSONUnmarshal = errors.Wrap(errJSONUnmarshal, "trigger payload json unmarshal error")

		return alertingConnectors, errJSONUnmarshal
	}

	if payload.State != handler.ALERTING {
		return alertingConnectors, handler.ErrPayloadStateIsNotAlerting
	}

	for _, connector := range payload.EvalMatches {
		alertingConnectors = append(alertingConnectors, connector.Metric)
	}

	return alertingConnectors, nil
}

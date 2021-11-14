package main

import (
	"strings"
	"testing"

	handler "github.com/modanisatech/grafana-webhook-handler"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func Test_GetAlertingConnector(t *testing.T) {
	genericTriggerPayload := `
		{
			"dashboardId": 10,
			"evalMatches": [
				{
					"value": 1,
					"metric": "some-connector",
					"tags": {
						"connector": "some-connector"
					}
				}
			],
			"orgId": 1,
			"panelId": 10,
			"ruleId": 7,
			"ruleName": "Overall Debezium Failed Connector Tasks Alert",
			"ruleUrl": "some url",
			"state": "",
			"tags": {},
			"title": "[Alerting] Overall Debezium Failed Connector Tasks Alert",
			"id": "18621555",
			"ref": "master",
			"variables": {}
		}
	`
	t.Run("when webhook payload is not valid, get error", func(t *testing.T) {
		triggerPayload := strings.ReplaceAll(genericTriggerPayload, `"state": ""`, `invalid json body`)
		names, err := getAlertingConnector(triggerPayload)
		assert.Len(t, names, 0)
		assert.NotNil(t, err)
	})

	t.Run("when webhook state is ok, no operation", func(t *testing.T) {
		triggerPayload := strings.ReplaceAll(genericTriggerPayload, `"state": ""`, `"state": "ok"`)

		names, err := getAlertingConnector(triggerPayload)
		assert.ErrorIs(t, err, handler.ErrPayloadStateIsNotAlerting)
		assert.Len(t, names, 0)
	})

	t.Run("when webhook state is alerting, handle", func(t *testing.T) {
		triggerPayload := strings.ReplaceAll(genericTriggerPayload, `"state": ""`, `"state": "alerting"`)

		names, err := getAlertingConnector(triggerPayload)
		assert.Nil(t, err)
		assert.Len(t, names, 1)
		assert.Equal(t, names[0], "some-connector")
	})
}

func Test_ApplyAlertingConnectorsOperation(t *testing.T) {
	ErrSomethingWentWrong := errors.New("something went wrong")

	mockController := gomock.NewController(t)

	client := NewMockDebeziumClient(mockController)

	t.Run("alerting connector not exist", func(t *testing.T) {
		client.EXPECT().GetConnectorFailedTaskIDs(gomock.Any(), gomock.Any()).Times(0)
		client.EXPECT().RestartConnectorTaskByID(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		alertingConnectors := AlertingConnectors{}
		err := alertingConnectors.applyOperation(client)
		assert.Nil(t, err)
	})

	t.Run("any alerting connector exists with connector failed task ids fetching error", func(t *testing.T) {
		connector := "test-connector"
		client.EXPECT().GetConnectorFailedTaskIDs(gomock.Any(), connector).Return(ConnectorFailedTaskIDs{}, ErrSomethingWentWrong).Times(1)
		client.EXPECT().RestartConnectorTaskByID(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		alertingConnectors := AlertingConnectors{connector}
		err := alertingConnectors.applyOperation(client)
		assert.NotNil(t, err)
	})

	t.Run("any alerting connector exists with connector task restarting error", func(t *testing.T) {
		connector := "test-connector"
		connectorFailedTaskIDs := ConnectorFailedTaskIDs{1, 2, 3}
		client.EXPECT().GetConnectorFailedTaskIDs(gomock.Any(), connector).Return(connectorFailedTaskIDs, nil).Times(1)
		for _, id := range connectorFailedTaskIDs {
			client.EXPECT().RestartConnectorTaskByID(gomock.Any(), connector, id).Return(ErrSomethingWentWrong).Times(1)
		}

		alertingConnectors := AlertingConnectors{connector}
		err := alertingConnectors.applyOperation(client)
		assert.NotNil(t, err)
	})

	t.Run("any alerting connector exists", func(t *testing.T) {
		alertingConnectors := AlertingConnectors{"test-connector-1", "test-connector-2"}
		connectorFailedTaskIDs := ConnectorFailedTaskIDs{1, 2, 3}

		for _, connector := range alertingConnectors {
			client.EXPECT().GetConnectorFailedTaskIDs(gomock.Any(), connector).Return(connectorFailedTaskIDs, nil).Times(1)
			for _, id := range connectorFailedTaskIDs {
				client.EXPECT().RestartConnectorTaskByID(gomock.Any(), connector, id).Return(nil).Times(1)
			}
		}

		err := alertingConnectors.applyOperation(client)
		assert.Nil(t, err)
	})
}

package main

import (
	"context"

	"github.com/cloudhut/connect-client"
	"github.com/pkg/errors"
)

type AlertingConnectors []string
type ConnectorFailedTaskIDs []int

type DebeziumClient interface {
	RestartConnectorTaskByID(context.Context, string, int) error
	GetConnectorFailedTaskIDs(context.Context, string) (ConnectorFailedTaskIDs, error)
}

type debeziumClient struct {
	*connect.Client
}

func NewDebeziumClient(hostURL string) DebeziumClient {
	connectClient := connect.NewClient(connect.WithHost(hostURL))

	return &debeziumClient{Client: connectClient}
}

func (dc *debeziumClient) RestartConnectorTaskByID(ctx context.Context, connectorName string, taskID int) error {
	err := dc.RestartConnectorTask(ctx, connectorName, taskID)
	if err != nil {
		return errors.Wrapf(err, "error when restarting connector %s task %d", connectorName, taskID)
	}

	return nil
}

func (dc *debeziumClient) GetConnectorFailedTaskIDs(ctx context.Context, connectorName string) (ConnectorFailedTaskIDs, error) {
	connectorFailedTaskIDs := make(ConnectorFailedTaskIDs, 0)

	stateInfo, err := dc.GetConnectorStatus(ctx, connectorName)
	if err != nil {
		return connectorFailedTaskIDs, errors.Wrapf(err, "error when getting connector %s status", connectorName)
	}

	for _, task := range stateInfo.Tasks {
		if task.State == "FAILED" {
			connectorFailedTaskIDs = append(connectorFailedTaskIDs, task.ID)
		}
	}

	return connectorFailedTaskIDs, nil
}

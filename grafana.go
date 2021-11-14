package handler

import "errors"

var ErrPayloadStateIsNotAlerting = errors.New("payload state is not alerting")

type AlertState string

const (
	OK       AlertState = "ok"
	PAUSED   AlertState = "paused"
	ALERTING AlertState = "alerting"
	PENDING  AlertState = "pending"
	NODATA   AlertState = "no_data"
)

type WebhookPayload struct {
	DashboardID int `json:"dashboardId"`
	EvalMatches []struct {
		Value  interface{} `json:"value"`
		Metric string      `json:"metric"`
		Tags   struct {
			Connector string `json:"connector"`
		} `json:"tags"`
	} `json:"evalMatches"`
	OrgID    int        `json:"orgId"`
	PanelID  int        `json:"panelId"`
	RuleID   int        `json:"ruleId"`
	RuleName string     `json:"ruleName"`
	RuleURL  string     `json:"ruleUrl"`
	State    AlertState `json:"state"`
	Tags     struct {
	} `json:"tags"`
	Title     string `json:"title"`
	ID        string `json:"id"`
	Ref       string `json:"ref"`
	Variables struct {
	} `json:"variables"`
}

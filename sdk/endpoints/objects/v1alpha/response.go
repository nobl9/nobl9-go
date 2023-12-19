package v1alpha

import "github.com/nobl9/nobl9-go/manifest/v1alpha/alert"

type GetAlertsResponse struct {
	Alerts       []alert.Alert
	TruncatedMax int
}

package slo

// PingdomMetric represents metric from Pingdom.
type PingdomMetric struct {
	CheckID   *string `json:"checkId" validate:"required,notBlank,numeric" example:"1234567"`
	CheckType *string `json:"checkType" validate:"required,pingdomCheckTypeFieldValid" example:"uptime"`
	Status    *string `json:"status,omitempty" validate:"omitempty,pingdomStatusValid" example:"up,down"`
}

const (
	PingdomTypeUptime      = "uptime"
	PingdomTypeTransaction = "transaction"
)

package slostatusapi

type Links struct {
	Self   string `json:"self"`
	Next   string `json:"next"`
	Cursor string
}

type Objective struct {
	Name                           string   `json:"name"`
	DisplayName                    string   `json:"displayName"`
	Target                         float64  `json:"target"`
	BurnRate                       *float64 `json:"burnRate,omitempty"`
	ErrorBudgetRemaining           *float64 `json:"errorBudgetRemaining,omitempty"`
	ErrorBudgetRemainingPercentage *float64 `json:"errorBudgetRemainingPercentage,omitempty"`
	Reliability                    *float64 `json:"reliability,omitempty"`
	Counts                         *Counts  `json:"counts,omitempty"`
	SLIType                        string   `json:"sliType"`
}

type Composite struct {
	Target            *float64           `json:"target,omitempty"`
	BurnRateCondition *BurnRateCondition `json:"burnRateCondition,omitempty"`
	CompositeObjective
}

type CompositeObjective struct {
	BurnRate                       *float64 `json:"burnRate,omitempty"`
	ErrorBudgetRemaining           *float64 `json:"errorBudgetRemaining,omitempty"`
	ErrorBudgetRemainingPercentage *float64 `json:"errorBudgetRemainingPercentage,omitempty"`
	Reliability                    *float64 `json:"reliability,omitempty"`
}

type BurnRateCondition struct {
	Value    float64 `json:"value"`
	Operator string  `json:"op"`
}

type Counts struct {
	Good  *float64 `json:"good,omitempty"`
	Total *float64 `json:"total,omitempty"`
}

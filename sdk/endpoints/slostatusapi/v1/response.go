package v1

type SLOListResponse struct {
	Data  []SLODetails `json:"data"`
	Links Links        `json:"links"`
}

type Links struct {
	Self   string `json:"self"`
	Next   string `json:"next"`
	Cursor string
}

type SLODetails struct {
	Name            string              `json:"name"`
	DisplayName     string              `json:"displayName"`
	Description     string              `json:"description"`
	Project         string              `json:"project"`
	Service         string              `json:"service"`
	CreatedAt       string              `json:"createdAt"`
	Objectives      []Objective         `json:"objectives"`
	Composite       *Composite          `json:"composite,omitempty"`
	Labels          map[string][]string `json:"labels,omitempty"`
	Annotations     map[string]string   `json:"annotations,omitempty"`
	BudgetingMethod string              `json:"budgetingMethod"`
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

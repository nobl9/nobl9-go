package models

// IAMRoleIDs struct which is used for exposing AWS IAM role auth data
type IAMRoleIDs struct {
	ExternalID string `json:"externalID"`
	AccountID  string `json:"accountID"`
}

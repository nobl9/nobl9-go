package models

// AWSIAMRoleAuthExternalIDs struct which is used for exposing AWS IAM role auth data
type AWSIAMRoleAuthExternalIDs struct {
	ExternalID string `json:"externalID"`
	AccountID  string `json:"accountID"`
}

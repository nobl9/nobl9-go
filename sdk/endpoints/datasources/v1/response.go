package v1

type M2MAppCredentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type IAMRoleIDs struct {
	ExternalID string `json:"externalID"`
	AccountID  string `json:"accountID"`
}

package v2

// GetUsersRequest defines filters for fetching users.
type GetUsersRequest struct {
	IDs   []string
	Limit uint
}

type getUsersRequest struct {
	Phrase string
	GetUsersRequest
}

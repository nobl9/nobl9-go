package v2

// GetUsersRequest defines filters for fetching users.
type GetUsersRequest struct {
	IDs []string
}

type getUsersRequest struct {
	Phrase string
	IDs    []string
}

package v2

type GetUsersRequest struct {
	IDs   []string
	Limit uint
}

type getUsersRequest struct {
	Phrase string
	GetUsersRequest
}

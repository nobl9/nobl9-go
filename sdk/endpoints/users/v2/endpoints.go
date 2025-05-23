package v2

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	endpointsHelpers "github.com/nobl9/nobl9-go/internal/endpoints"
)

const (
	baseAPIPath = "usrmgmt/v2/users"
)

//go:generate ../../../../bin/ifacemaker -y " " -f ./*.go -s endpoints -i Endpoints -o endpoints_interface.go -p "$GOPACKAGE"

func NewEndpoints(client endpointsHelpers.Client) Endpoints {
	return endpoints{client: client}
}

type endpoints struct {
	client endpointsHelpers.Client
}

// GetUsers fetches a list of [User] filtered by the provided search phrase.
func (e endpoints) GetUsers(ctx context.Context, params GetUsersRequest) ([]*User, error) {
	q := url.Values{"phrase": []string{params.Phrase}}
	req, err := e.client.CreateRequest(ctx, http.MethodGet, baseAPIPath, nil, q, nil)
	if err != nil {
		return nil, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var users struct {
		Users []*User `json:"users"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}
	return users.Users, nil
}

// GetUser fetches a user by the id.
// It returns nil if the user was not found.
func (e endpoints) GetUser(ctx context.Context, id string) (*User, error) {
	users, err := e.GetUsers(ctx, GetUsersRequest{Phrase: id})
	if err != nil {
		return nil, err
	}
	switch len(users) {
	case 1:
		return users[0], nil
	case 0:
		return nil, nil
	default:
		return nil, errors.Errorf("unexpected number of users returned: %d", len(users))
	}
}

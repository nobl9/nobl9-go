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

// GetUser fetches a user by a unique identifier, this can be either:
//   - external id (e.g. 00u2y4e4atkzaYkXP4x8)
//   - email (e.g. foo.bar@nobl9.com)
//
// It returns nil if the user was not found.
func (e endpoints) GetUser(ctx context.Context, id string) (*User, error) {
	users, err := e.getUsers(ctx, getUsersRequest{Phrase: id})
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

// getUsers fetches a list of [User] filtered by the provided search phrase.
func (e endpoints) getUsers(ctx context.Context, params getUsersRequest) ([]*User, error) {
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

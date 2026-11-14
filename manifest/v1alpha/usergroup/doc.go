// Package usergroup defines UserGroup object definitions.
//
// Organization administrators can create and manage manual groups. SCIM and
// system groups remain managed by their respective owners. Applying a manual
// group replaces its complete member list and does not invite or provision users.
// A UserGroup and its RoleBindings can be applied together; groups are created
// before bindings are resolved, regardless of their order in the request.
// Before deleting a group, remove its RoleBindings, either separately or in the
// same delete request. A group with remaining bindings cannot be deleted.
package usergroup

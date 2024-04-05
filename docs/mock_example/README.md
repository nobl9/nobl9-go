# Mocking the SDK Client

In order to mock the instance of `sdk.Client` in your project,
while possible to do manually, we recommend using [mockgen](https://github.com/uber-go/mock).

Define the interface you want to accept for `sdk.Client` within your project.
If you're using any of the API endpoint's methods, like `sdk.Client.Objects`,
you'll need to mock these interfaces as well.

The example Go module within this directory demonstrates how to use `mockgen`
to generate mocks for `sdk.Client` and verify a function for fetching Projects.
package v1

type Label struct {
	ID        string            `json:"id"`
	Key       string            `json:"key"`
	Value     string            `json:"value"`
	Resources []LabeledResource `json:"resources"`
}

type LabeledResource struct {
	Kind  string `json:"kind"`
	Count int    `json:"count"`
}

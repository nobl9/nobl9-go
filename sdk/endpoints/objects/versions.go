package objects

import "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1alpha"

type Versions struct{}

func (v Versions) V1alpha() v1alpha.Endpoints {

}

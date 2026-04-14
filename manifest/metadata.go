package manifest

// MetadataProvider is implemented by objects that expose typed metadata.
//
// The metadata type is object-kind specific, e.g. v1alphaSLO.Metadata.
type MetadataProvider[M any] interface {
	GetMetadata() M
}

// MapMetadata returns metadata for each provided object in the same order.
func MapMetadata[M any, O MetadataProvider[M]](objects []O) []M {
	if objects == nil {
		return nil
	}
	metadata := make([]M, 0, len(objects))
	for _, object := range objects {
		metadata = append(metadata, object.GetMetadata())
	}
	return metadata
}

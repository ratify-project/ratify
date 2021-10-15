package types

type ArtifactTypeVerifyPolicy string

const (
	AnyVerifySuccess ArtifactTypeVerifyPolicy = "any"
	AllVerifySuccess                          = "all"
)

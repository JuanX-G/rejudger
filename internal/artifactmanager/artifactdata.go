package artifactmanager

import "time"

type ArtifactFileType int

//go:generate stringer -type=ArtifactFileType
const (
	ArtifactUnknown ArtifactFileType = iota
	ArtifactPNG
	ArtifactJPG
)

func ParseArtifactFileType(ft string) ArtifactFileType {
	switch ft {
	case ArtifactJPG.String():
		return ArtifactJPG
	case ArtifactPNG.String():
		return ArtifactPNG
	default:
		return ArtifactUnknown
	}
}

type ArtifactMetadata struct {
	FileType   ArtifactFileType
	BaseHash   string
	UploadDate time.Time
	Name       string
}

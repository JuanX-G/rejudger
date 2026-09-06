package artifactmanager

import (
	"context"
	"io"
)

type ArtifactStore interface {
	Put(context.Context, ArtifactMetadata, io.ReadCloser) error
	Get(ctx context.Context, fileType ArtifactFileType, hash string, rev int32) (io.ReadCloser, error)
	GetData(ctx context.Context, fileType ArtifactFileType, hash string) (ArtifactMetadata, []int32, error)
	Delete(ctx context.Context, fileType ArtifactFileType, hash string, name string, rev int32) error
	Exists(ctx context.Context, fileType ArtifactFileType, hash string, name string, rev int32) error
}

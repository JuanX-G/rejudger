package testingservices

import (
	"bytes"
	"context"
	"io"
	"revit/internal/artifactmanager"
)

type ExpectedMsg struct {
	Hashes         []string
	SubmissionHash string
	SubmissionId   int64
}

type MockArtifactService struct {
	store             *MockArtifactStore
	ExpectedArtifacts chan ExpectedMsg
}

func (as *MockArtifactService) ExpectArtifacts(hashes []string, submissionHash string, submissionId int64) {
	if as.ExpectedArtifacts != nil {
		as.ExpectedArtifacts <- ExpectedMsg{Hashes: hashes, SubmissionHash: submissionHash, SubmissionId: submissionId}
	}
}

func NewMockArtifactService(st *MockArtifactStore) *MockArtifactService {
	return &MockArtifactService{store: st}
}

type MockArtifact struct {
	Meta  artifactmanager.ArtifactMetadata
	Bytes []byte
}

type MockArtifactStore struct {
	art  MockArtifact
	full bool
}

func (as *MockArtifactStore) Put(ctx context.Context, meta artifactmanager.ArtifactMetadata, r io.ReadCloser) error {
	as.art.Meta = meta
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	as.art.Bytes = b
	return nil
}

const DEFAULT_OBJECT = "object"
const DEFAULT_OBJECT_HASH = "ohash"
const DEFAULT_OBJECT_NAME = "oname"

var DEFAULT_OBJECT_METADATA = artifactmanager.ArtifactMetadata{FileType: artifactmanager.ArtifactJPG, BaseHash: DEFAULT_OBJECT_HASH, Name: DEFAULT_OBJECT_NAME}

func (as *MockArtifactStore) Get(ctx context.Context, fileType artifactmanager.ArtifactFileType, hash string, rev int32) (io.ReadCloser, error) {
	if as.full {
		return io.NopCloser(bytes.NewReader(as.art.Bytes)), nil
	} else {
		return io.NopCloser(bytes.NewReader([]byte(DEFAULT_OBJECT))), nil
	}
}

func (as *MockArtifactStore) GetData(ctx context.Context, fileType artifactmanager.ArtifactFileType, hash string) (artifactmanager.ArtifactMetadata, []int32, error) {
	if as.full {
		return as.art.Meta, []int32{1, 2, 3, 4, 5, 6}, nil
	} else {
		return DEFAULT_OBJECT_METADATA, []int32{1, 2}, nil
	}
}

func (as *MockArtifactStore) Delete(ctx context.Context, fileType artifactmanager.ArtifactFileType, hash, name string, rev int32) error {
	as.art = MockArtifact{}
	return nil
}

func (as *MockArtifactStore) Exists(ctx context.Context, fileType artifactmanager.ArtifactFileType, hash string, name string, rev int32) error {
	if as.full {
		return nil
	} else {
		return artifactmanager.ErrArtifactDoesNotExist
	}
}

func (as *MockArtifactStore) GetRawContents() MockArtifact {
	return as.art
}

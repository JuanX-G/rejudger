package artifactmanager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"revit/internal/db"
	"revit/internal/store"
)

func MakeFileExtension(fileType ArtifactFileType) string {
	switch fileType {
	case ArtifactJPG:
		return "jpeg"
	case ArtifactPNG:
		return "png"
	default:
		return ""
	}
}
func MakeFilePath(hash string, name string, generation int32, fileType ArtifactFileType) string {
	ext := MakeFileExtension(fileType)
	return fmt.Sprintf("%s/%s/%d-%s.%s", fileType.String(), hash, generation, name, ext)
}

func MakeFolderPath(hash string, name string, fileType ArtifactFileType) string {
	return fmt.Sprintf("%s/%s/%s", fileType.String(), hash, name)
}

type ArtifactManager struct {
	store    store.Store
	basePath string
}

func NewArtifactManager(store store.Store) *ArtifactManager {
	return &ArtifactManager{store: store}
}

var ErrArtifactDoesNotExist = errors.New("error artifacts does not exist")
var ErrInvalidFileType = errors.New("error invalid file type")
var ErrInvalidRevision = errors.New("error invalid revision")

func (am *ArtifactManager) Put(ctx context.Context, data ArtifactMetadata, submissionId int64, r io.ReadCloser) error {
	defer r.Close()
	queryFn := func(q db.Querier) error {
		ok, err := q.ExistsArtifact(ctx, db.ExistsArtifactParams{Hash: data.BaseHash, FileType: data.FileType.String()})
		if err != nil {
			return err
		}
		if ok {
			artifact, err := q.GetNewestArtifact(ctx, db.GetNewestArtifactParams{Hash: data.BaseHash, FileType: data.FileType.String()})
			if err != nil {
				return err
			}

			err = q.InsertArtifact(ctx, db.InsertArtifactParams{
				Hash:         data.BaseHash,
				Name:         data.Name,
				SubmissionID: submissionId,
				FileType:     data.FileType.String()})
			if err != nil {
				return err
			}
			gen := artifact.Generation + 1

			path := MakeFilePath(data.BaseHash, data.Name, gen, data.FileType)
			path = am.basePath + path
			f, err := os.OpenFile(path, os.O_TRUNC, 0755)
			if err != nil {
				return err
			}
			defer f.Close()
			io.Copy(f, r)
			return nil
		}

		err = q.InsertArtifact(ctx, db.InsertArtifactParams{Hash: data.BaseHash, FileType: data.FileType.String(), SubmissionID: submissionId})
		if err != nil {
			return err
		}

		path := MakeFilePath(data.BaseHash, data.Name, 1, data.FileType)
		path = am.basePath + path
		f, err := os.OpenFile(path, os.O_TRUNC, 0755)
		if err != nil {
			return err
		}
		defer f.Close()
		io.Copy(f, r)
		return nil
	}
	return am.store.ExecTx(ctx, queryFn)
}

func (am *ArtifactManager) Get(ctx context.Context, fileType ArtifactFileType, hash string, rev int32) (io.ReadCloser, error) {
	var reader io.ReadCloser
	queryFn := func(q db.Querier) error {
		ok, err := q.ExistsArtifact(ctx, db.ExistsArtifactParams{Hash: hash, FileType: fileType.String()})
		if err != nil {
			return err
		}
		if !ok {
			return ErrArtifactDoesNotExist
		}
		artifact, err := q.GetArtifactByGeneration(ctx, db.GetArtifactByGenerationParams{FileType: fileType.String(), Hash: hash, Generation: rev})
		if err != nil {
			return err
		}
		fileType := ParseArtifactFileType(artifact.FileType)
		if fileType == ArtifactUnknown {
			return ErrInvalidFileType
		}
		path := MakeFilePath(artifact.Hash, artifact.Name, rev, fileType)
		path = am.basePath + path
		f, err := os.OpenFile(path, os.O_RDONLY, 0755)
		if err != nil {
			return err
		}
		reader = f
		return nil
	}
	err := am.store.ExecTx(ctx, queryFn)
	if err != nil {
		if reader != nil {
			reader.Close()
		}
		return nil, err
	}
	return reader, nil
}

func (am *ArtifactManager) Delete(ctx context.Context, fileType ArtifactFileType, hash, name string, rev int32) error {
	queryFn := func(q db.Querier) error {
		ok, err := q.ExistsArtifact(ctx, db.ExistsArtifactParams{Hash: hash, FileType: fileType.String()})
		if err != nil {
			return err
		}
		if !ok {
			return ErrArtifactDoesNotExist
		}
		if rev == 0 {
			err = q.DeleteArtifact(ctx, db.DeleteArtifactParams{Hash: hash, FileType: fileType.String()})
			if err != nil {
				return err
			}

			path := MakeFolderPath(hash, name, fileType)
			err = os.RemoveAll(am.basePath + path)
			if err != nil {
				return err
			}

			return nil
		} else {
			err := q.DeleteArtifactByGen(ctx, db.DeleteArtifactByGenParams{Hash: hash, FileType: fileType.String(), Generation: rev})
			if err != nil {
				return err
			}

			path := MakeFilePath(hash, name, rev, fileType)
			err = os.Remove(am.basePath + path)
			if err != nil {
				return err
			}

			return nil
		}
	}
	return am.store.ExecTx(ctx, queryFn)
}

func (am *ArtifactManager) Exists(ctx context.Context, fileType ArtifactFileType, hash, name string, rev int32) error {
	queryFn := func(q db.Querier) error {
		if rev < 1 {
			return ErrInvalidRevision
		}
		ok, err := q.ExistsArtifactRev(ctx, db.ExistsArtifactRevParams{Hash: hash, FileType: fileType.String(), Generation: rev})
		if err != nil {
			return err
		}
		if !ok {
			return ErrArtifactDoesNotExist
		}
		return nil
	}
	return am.store.ExecTx(ctx, queryFn)
}

func (am *ArtifactManager) GetData(ctx context.Context, fileType ArtifactFileType, hash string) (ArtifactMetadata, []int32, error) {
	artifacts, err := am.store.GetQueries().GetArtifacts(ctx, db.GetArtifactsParams{Hash: hash, FileType: fileType.String()})
	if err != nil {
		return ArtifactMetadata{}, []int32{}, err
	}

	out := make([]int32, 0, len(artifacts))
	for _, a := range artifacts {
		out = append(out, a.Generation)
	}

	return ArtifactMetadata{
		FileType:   ParseArtifactFileType(artifacts[0].FileType),
		BaseHash:   artifacts[0].Hash,
		UploadDate: artifacts[0].UploadedAt.Time,
		Name:       artifacts[0].Name,
	}, out, nil
}

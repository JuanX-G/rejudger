package artifactservice

import (
	"revit/internal/artifactmanager"
	"sync"
)

type ArtifactService interface {
	ExpectArtifacts(hashes []string, submissionHash string, submissionId int64)
}

type ArtifactRef struct {
	submisionHash string
	hash          string
}

type BaseArtifactService struct {
	mu       sync.RWMutex
	expected map[ArtifactRef]int64
	mgr      *artifactmanager.ArtifactManager
}

func (as *BaseArtifactService) ExpectArtifacts(hashes []string, subHash string, submissionId int64) {
	as.mu.Lock()
	defer as.mu.Unlock()
	for _, h := range hashes {
		if _, ok := as.expected[ArtifactRef{hash: h, submisionHash: subHash}]; ok {
			continue
		}
		as.expected[ArtifactRef{hash: h, submisionHash: subHash}] = submissionId
	}
}

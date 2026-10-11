package deadlines

import (
	"context"
	"log"
	"revit/internal/db"
	"sync"
	"time"
)

type StageTask struct {
	StageID  int64
	Interval time.Duration
}

type DeadlineMgr struct {
	store   DeadlineStore
	mu      sync.RWMutex
	workers map[int64]context.CancelFunc
}

func NewDeadlineMgr(store DeadlineStore) *DeadlineMgr {
	return &DeadlineMgr{
		store:   store,
		workers: make(map[int64]context.CancelFunc),
	}
}

func (dm *DeadlineMgr) PastDeadline(ctx context.Context, stageID int64) (bool, error) {
	deadline, err := dm.store.GetDeadlineByStageId(ctx, stageID)
	if err != nil {
		return false, err
	}

	if deadline.LastRan.Valid {
		if !deadline.LastRan.Time.Add(time.Duration(deadline.Duration)).After(time.Now()) {
			return true, nil
		}
		return false, nil
	}
	return false, ErrInvalidDeadlineData
}

// StartSchedulers initializes background workers for a slice of tasks.
func (dm *DeadlineMgr) StartSchedulers(ctx context.Context, tasks []StageTask, wg *sync.WaitGroup) {
	for _, task := range tasks {
		dm.StartStageScheduler(ctx, task.StageID, task.Interval, wg)
	}
}

// StartStageScheduler launches a background goroutine for a specific stage ID and duration interval.
func (dm *DeadlineMgr) StartStageScheduler(ctx context.Context, stageID int64, interval time.Duration, wg *sync.WaitGroup) {
	dm.mu.Lock()
	if cancel, exists := dm.workers[stageID]; exists {
		cancel()
	}

	workerCtx, cancel := context.WithCancel(ctx)
	dm.workers[stageID] = cancel
	dm.mu.Unlock()

	if wg != nil {
		wg.Add(1)
	}

	go func(ctx context.Context, stageID int64, interval time.Duration) {
		if wg != nil {
			defer wg.Done()
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				dm.executeTask(ctx, stageID)
			}
		}
	}(workerCtx, stageID, interval)
}

// StopStageScheduler cancels a running task goroutine for a specific stage ID.
func (dm *DeadlineMgr) StopStageScheduler(stageID int64) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if cancel, exists := dm.workers[stageID]; exists {
		cancel()
		delete(dm.workers, stageID)
	}
}

func (dm *DeadlineMgr) executeTask(ctx context.Context, stageID int64) {
	stage, err := dm.store.GetStageById(ctx, stageID)
	if err != nil {
		log.Printf("[deadlines] failed to get stage %d: %v", stageID, err)
		return
	}

	submissions, err := dm.store.GetSubmissionsByStageId(ctx, stageID)
	if err != nil {
		log.Printf("[deadlines] failed to get submissions for stage %d: %v", stageID, err)
		return
	}

	for _, submission := range submissions {
		if ctx.Err() != nil {
			return
		}

		vetos, err := dm.store.GetVetosBySubmissionId(ctx, submission.ID)
		if err != nil {
			log.Printf(
				"[deadlines] failed to get vetos for submission %d: %v",
				submission.ID, err,
			)
			continue
		}

		// Only process submissions that have at least one veto,
		// and whose vetos are all soft.
		if len(vetos) == 0 {
			continue
		}

		hasHardVeto := false
		for _, veto := range vetos {
			if veto.Hard {
				hasHardVeto = true
				break
			}
		}
		if hasHardVeto {
			continue
		}

		count, err := dm.store.CountPlusOnes(
			ctx,
			db.CountPlusOnesParams{
				SubmissionID:  submission.ID,
				PipelineStage: submission.PipelineStage,
			},
		)
		if err != nil {
			log.Printf(
				"[deadlines] failed to count plus-ones for submission %d: %v",
				submission.ID, err,
			)
			continue
		}

		if count < int64(stage.PlusOneRequired) {
			continue
		}

		rows, err := dm.store.AdvanceSubmissionToNextStage(ctx, submission.ID)
		if err != nil {
			log.Printf(
				"[deadlines] failed to advance submission %d: %v",
				submission.ID, err,
			)
			continue
		}

		if rows == 0 {
			// No next stage exists, or the submission no longer
			// matches the stage it was read from.
			continue
		}

		log.Printf(
			"[deadlines] advanced soft-vetoed submission %d from stage %d",
			submission.ID, stageID,
		)
	}

	if err := dm.store.SetDeadlineLastRanToNowByStageId(ctx, stageID); err != nil {
		log.Printf(
			"[deadlines] failed to update last_ran for stage %d: %v",
			stageID, err,
		)
	}
}

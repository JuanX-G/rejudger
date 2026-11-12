package dbpublishingservice

import (
	"context"
	"revit/internal/db"
	"revit/internal/sharedmodels"

	"github.com/jackc/pgx/v5/pgtype"
)

type DBPublishingService struct {
	store dbPublishingServiceStore
}

func (dps *DBPublishingService) Publish(ctx context.Context, submission sharedmodels.Submission) error {
	return dps.store.InsertIntoPublishedSubmissions(ctx,
		db.InsertIntoPublishedSubmissionsParams{
			CreatedAt: pgtype.Timestamp{
				Time:             submission.DateCrated,
				InfinityModifier: pgtype.Finite,
				Valid:            true,
			},
		})

}

func (dps *DBPublishingService) GetPublishedByHash(ctx context.Context, hash string) (sharedmodels.Submission, error) {
	submission, err := dps.store.GetPublishedByHash(ctx, hash)
	if err != nil {
		return sharedmodels.Submission{}, err
	}

	user, err := dps.store.GetUserById(ctx, submission.Author)
	if err != nil {
		return sharedmodels.Submission{}, err
	}

	userData := sharedmodels.MapUserDataFromDBModel(user)

	return sharedmodels.Submission{
		DateCrated: submission.CreatedAt.Time,
		Author:     userData,
		Content:    submission.Content,
		Hash:       submission.Hash,
	}, nil
}

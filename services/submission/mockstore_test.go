package submission

import (
	"context"
	"revit/internal/db"
	services "revit/internal/testingservices"

	"github.com/jackc/pgx/v5/pgtype"
)

type InsertionNotification struct {
	db.InsertSubmissionParams
	submissionId int64
}

type mockSubmissionStore struct {
	insertNotify chan InsertionNotification
}

func (ss *mockSubmissionStore) GetUserById(ctx context.Context, id int64) (db.User, error) {
	return db.User{
		ID:    services.DEFAULT_USERID,
		Name:  services.DEFAULT_USERNAME,
		Email: pgtype.Text{String: services.DEFAULT_EMAIL, Valid: true},
	}, nil
}

func (ss *mockSubmissionStore) InsertSubmission(ctx context.Context, arg db.InsertSubmissionParams) error {
	ss.insertNotify <- InsertionNotification{arg, 123}
	return nil
}

func (ss *mockSubmissionStore) GetSubmissionsByHash(ctx context.Context, hash string) (db.Submission, error) {
	return db.Submission{
		ID: 123,
	}, nil
}

func (ss *mockSubmissionStore) GetSubmissionsByAuthor(ctx context.Context, author int64) ([]db.Submission, error) {
	return DEFAULT_SUBMISSIONS, nil
}

func (ss *mockSubmissionStore) GetSubmissionsByAuthorOffset(ctx context.Context, arg db.GetSubmissionsByAuthorOffsetParams) ([]db.Submission, error) {
	if arg.Offset == 0 {
		return DEFAULT_SUBMISSIONS, nil
	} else if arg.Offset == 1 {
		return []db.Submission{DEFAULT_SUBMISSION_2}, nil
	} else {
		return []db.Submission{}, nil // TODO: add error here
	}
}

func (ss *mockSubmissionStore) GetUserByEmail(ctx context.Context, email pgtype.Text) (db.User, error) {
	return db.User{
		ID:    services.DEFAULT_USERID,
		Name:  services.DEFAULT_USERNAME,
		Email: pgtype.Text{String: services.DEFAULT_EMAIL, Valid: true},
	}, nil
}

func (ss *mockSubmissionStore) GetUserByName(ctx context.Context, name string) (db.User, error) {
	return db.User{
		ID:    services.DEFAULT_USERID,
		Name:  services.DEFAULT_USERNAME,
		Email: pgtype.Text{String: services.DEFAULT_EMAIL, Valid: true},
	}, nil
}

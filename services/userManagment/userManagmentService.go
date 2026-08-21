package userManagmentService

import (
	"context"
	"net/http"
	"time"

	db "revit/internal/db"
	"revit/internal/jsonHelpers"
	"revit/internal/store"
	"revit/services/auth"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const DEFAULT_USER_MANAGMENT_TIMEOUT = 10

type UserManagementService struct {
	guard   *auth.EndpointGuard
	store   store.Store
	timeout time.Duration
}

func (s *UserManagementService) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /magament/users/get", s.guard.LockRead(s.getUserHandler()))
	mux.HandleFunc("POST /managment/users/add", s.guard.LockWrite(s.addUserHandler()))
	mux.HandleFunc("POST /magament/users/delete", s.guard.LockWrite(s.deleteUserHandler()))
	mux.HandleFunc("POST /magament/users/add_role", s.guard.LockWrite(s.addToRoleHandler()))
}

func NewUserManagementService(pool *pgxpool.Pool, authMgr *auth.AuthManager, appContext string, timeout uint) *UserManagementService {
	guard := auth.NewEndpointGuard(authMgr, appContext)
	if timeout != 0 {
		return &UserManagementService{guard: guard, store: store.NewStore(pool), timeout: time.Duration(timeout) * time.Second}
	} else {
		return &UserManagementService{guard: guard, store: store.NewStore(pool), timeout: DEFAULT_USER_MANAGMENT_TIMEOUT * time.Second}
	}
}

func (s *UserManagementService) requestCtx(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), s.timeout)
}

func writeUserManagmentFail(w http.ResponseWriter, status int, msg string) {
	jsonHelpers.WriteJSON(w, status, UserManagmentResponse{Success: false, Message: msg})
}

type UserActionQuery struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	InternaId string `json:"internal_id"`
}

func getUserDataFromQuery(ctx context.Context, w http.ResponseWriter, queries *db.Queries, query UserActionQuery) (db.User, error) {
	var err error
	var user db.User
	if query.Email != "" && query.Name == "" {
		user, err = queries.GetUserByEmail(ctx, pgtype.Text{String: query.Email, Valid: true})
		if err != nil {
			writeUserManagmentFail(w, http.StatusNotFound, "user not found")
			return db.User{}, err
		}
	} else if query.InternaId != "" && query.Name == "" {
		user, err = queries.GetUserByInternalId(ctx, pgtype.Text{String: query.InternaId, Valid: true})
		if err != nil {
			writeUserManagmentFail(w, http.StatusNotFound, "user not found")
			return db.User{}, err
		}
	} else if query.Name != "" {
		user, err = queries.GetUserByName(ctx, query.Name)
		if err != nil {
			writeUserManagmentFail(w, http.StatusNotFound, "user not found")
			return db.User{}, err
		}
	} else {
		writeUserManagmentFail(w, http.StatusBadRequest, "no data to find the user by in the request")
		return db.User{}, err
	}
	return user, nil
}

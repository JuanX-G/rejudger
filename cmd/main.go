package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"revit/internal/artifactmanager"
	"revit/internal/logger"
	"revit/internal/permissions"
	"revit/internal/store"
	"revit/services/artifactservice"
	"revit/services/auth"
	"revit/services/login"
	"revit/services/submission"
	userManagmentService "revit/services/userManagment"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "")
	if err != nil {
		fmt.Println("Err: ", err)
	}

	loggerMgr := logger.NewMultiLogger()
	authMgr := auth.NewAuthManager()
	store := store.NewStore(pool)
	permissionMgr, err := permissions.NewPermissionManager(ctx, store)
	if err != nil {
		fmt.Println("Err: ", err)
	}

	artifactMgr := artifactmanager.NewArtifactManager(store)
	artifactSvc := artifactservice.NewBaseArtifactService(artifactMgr)

	loginMgr, err := login.NewLoginManager(ctx, store, authMgr, permissionMgr)
	if err != nil {
		fmt.Println("Err: ", err)
	}

	mux := http.NewServeMux()

	userManagmentSvc := userManagmentService.NewUserManagementService(pool, authMgr, "userManagment", 12)

	submissionSvc := submission.NewSubmissionService(store, authMgr, artifactSvc, loggerMgr, "tech", 0)

	userManagmentSvc.RegisterRoutes(mux)
	submissionSvc.RegisterRoutes(mux)
	loginMgr.RegisterRoutes(mux)

	err = http.ListenAndServe(":6619", mux)
	if err != nil {
		log.Fatalf("Error serving: %s", err)
	}
}

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"revit/internal/artifactmanager"
	"revit/internal/config"
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
	loggerMgr, err := logger.MakeDefaultMultiLogger(ctx, os.Stdout)
	if err != nil {
		fmt.Println("Err: ", err)
		os.Exit(1)
	}

	pool, err := pgxpool.New(ctx, "user=macia host=localhost port=5432 dbname=rejudger pool_max_conns=10")
	if err != nil {
		panic(err)
	}

	cfgMgr, results, err := config.NewConfigMgr(pool, "./config/config.yaml")
	if err != nil {
		panic(err)
	}

	for _, v := range results {
		loggerMgr.Log(ctx, logger.LogMessage{
			Level:  logger.LogLevelError,
			Source: "main",
			Msg:    fmt.Sprintf("error loading config file: %s, error reported: %s", v.FileName, v.Err.Error()),
		})
	}

	err = cfgMgr.SyncAll(ctx)
	if err != nil {
		panic(err)
	}

	authMgr := auth.NewAuthManager()
	store := store.NewStore(pool)
	permissionMgr := permissions.NewPermissionManager(ctx, store)

	artifactMgr := artifactmanager.NewArtifactManager(store)
	artifactSvc := artifactservice.NewBaseArtifactService(artifactMgr)

	loginMgr := login.NewLoginManager(ctx, store, authMgr, permissionMgr)

	mux := http.NewServeMux()

	userManagmentSvc := userManagmentService.NewUserManagementService(pool, authMgr, "userManagment", 12)

	submissionSvc := submission.NewSubmissionService(submission.SubmissionServiceParams{
		Store:       store,
		AuthMgr:     authMgr,
		ArtifactSvc: artifactSvc,
		BaseLogger:  loggerMgr,
		AppContext:  "tech",
		Timeout:     0,
	})

	userManagmentSvc.RegisterRoutes(mux)
	submissionSvc.RegisterRoutes(mux)
	loginMgr.RegisterRoutes(mux)

	err = loggerMgr.Log(ctx, logger.LogMessage{
		Level:  logger.LogLevelInfo,
		Source: "main",
		Msg:    fmt.Sprintf("HTTP server listeing on: %d", 6619),
	})

	err = http.ListenAndServe(":6619", mux)
	if err != nil {
		log.Fatalf("Error serving: %s", err)
	}
}

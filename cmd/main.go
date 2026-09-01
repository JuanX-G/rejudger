package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"revit/internal/db"
	"revit/internal/permissions"
	"revit/internal/store"
	"revit/services/auth"
	"revit/services/login"
	userManagmentService "revit/services/userManagment"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "")
	if err != nil {
		fmt.Println("Err: ", err)
	}

	authMgr := auth.NewAuthManager()
	store := store.NewStore(pool)
	permissionMgr, err := permissions.NewPermissionManager(ctx, store)
	if err != nil {
		fmt.Println("Err: ", err)
	}

	loginMgr, err := login.NewLoginManager(ctx, db.New(pool), authMgr, permissionMgr)
	if err != nil {
		fmt.Println("Err: ", err)
	}

	mux := http.NewServeMux()

	userManagmentSvc := userManagmentService.NewUserManagementService(pool, authMgr, "userManagmentCtx", 12)

	userManagmentSvc.RegisterRoutes(mux)
	loginMgr.RegisterRoutes(mux)

	err = http.ListenAndServe(":6619", mux)
	if err != nil {
		log.Fatalf("Error serving: %s", err)
	}
}

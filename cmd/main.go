package main

import (
	"context"
	"fmt"
	"net/http"
	"revit/internal/permissions"
	"revit/services/auth"
	"revit/services/login"
	"revit/services/userManagment"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "")
	if err != nil {
		fmt.Println("Err: ", err)
	}

	authMgr := auth.NewAuthManager()
	permissionMgr, err := permissions.NewPermissionManager(ctx, pool)
	if err != nil {
		fmt.Println("Err: ", err)
	}

	loginMgr, err := login.NewLoginManager(ctx, pool, authMgr, permissionMgr)
	if err != nil {
		fmt.Println("Err: ", err)
	}
	fmt.Println(loginMgr)

	mux := http.NewServeMux()

	userManagmentSvc := userManagmentService.NewUserManagementService(pool, authMgr, "userManagmentCtx", 12)

	userManagmentSvc.RegisterRoutes(mux)
	loginMgr.RegisterRoutes(mux)
}

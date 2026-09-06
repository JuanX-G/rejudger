package permissions

import (
	"context"
	"revit/internal/db"
)

func QueriesInsertPermissionSet(ctx context.Context, q db.Querier, set PermissionSet) error {
	for k := range set.Permissions {
		err := q.InsertPermission(ctx, db.InsertPermissionParams{Context: set.Context, Action: k.String()})
		if err != nil {
			return err
		}
	}
	return nil
}

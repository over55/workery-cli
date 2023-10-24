package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"

	"github.com/over55/workery-cli/adapter/storage/mongodb"
	"github.com/over55/workery-cli/adapter/storage/postgres"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	user_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importUserRoleCmd)
}

var importUserRoleCmd = &cobra.Command{
	Use:   "import_user_role",
	Short: "Import the user role from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		defaultLogger := slog.Default()
		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		userStorer := user_ds.NewDatastore(cfg, defaultLogger, mc)
		RunImportUserRole(cfg, ppc, lpc, tenantStorer, userStorer)
	},
}

type OldUserGroup struct {
	Id      uint64 `json:"id"`
	UserId  uint64 `json:"shareduser_id"`
	GroupId uint64 `json:"group_id"`
}

// Function returns a paginated list of all type element items.
func ListAllUserGroups(db *sql.DB) ([]*OldUserGroup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
	    id, shareduser_id, group_id
	FROM
	    workery_users_groups
	ORDER BY
		id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldUserGroup
	defer rows.Close()
	for rows.Next() {
		m := new(OldUserGroup)
		err = rows.Scan(
			&m.Id,
			&m.UserId,
			&m.GroupId,
		)
		if err != nil {
			panic(err)
		}
		arr = append(arr, m)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return arr, err
}

func RunImportUserRole(cfg *config.Conf, public *sql.DB, london *sql.DB, tenantStorer tenant_ds.TenantStorer, userStorer user_ds.UserStorer) {
	fmt.Println("Beginning importing user roles")
	data, err := ListAllUserGroups(public)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importUserRole(context.Background(), tenantStorer, userStorer, datum)
	}
	fmt.Println("Finished importing user roles")
}

func importUserRole(ctx context.Context, ts tenant_ds.TenantStorer, us user_ds.UserStorer, ou *OldUserGroup) {
	user, err := us.GetByOldID(ctx, ou.UserId)
	if err != nil {
		log.Fatal(err)
	}
	if user == nil {
		log.Println("missing user", ou.UserId)
		return
	}
	user.Role = int8(ou.GroupId)
	if err := us.UpdateByID(ctx, user); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Imported user role ID#", user.ID, "role", user.Role)
}

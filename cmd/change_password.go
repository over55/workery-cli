package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"

	"github.com/over55/workery-cli/adapter/storage/mongodb"
	"github.com/over55/workery-cli/adapter/storage/postgres"
	user_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
	"github.com/over55/workery-cli/provider/password"
	p "github.com/over55/workery-cli/provider/password"
	"github.com/spf13/cobra"
)

// ex:
// $ go run main.go change_password --email="b@b.com" --password="123"

var (
	changePassEmail    string
	changePassPassword string
)

func init() {
	changePasswordCmd.Flags().StringVarP(&changePassEmail, "email", "e", "", "Email of the user account")
	changePasswordCmd.MarkFlagRequired("email")
	changePasswordCmd.Flags().StringVarP(&changePassPassword, "password", "p", "", "Password of the user account")
	changePasswordCmd.MarkFlagRequired("password")
	rootCmd.AddCommand(changePasswordCmd)
}

var changePasswordCmd = &cobra.Command{
	Use:   "change_password",
	Short: "Change user password",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.New()
		pass := password.NewProvider()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		defaultLogger := slog.Default()
		userStorer := user_ds.NewDatastore(cfg, defaultLogger, mc)
		runChangePassword(cfg, ppc, lpc, pass, userStorer)
	},
}

func runChangePassword(cfg *config.Conf, public *sql.DB, london *sql.DB, pass p.Provider, us user_ds.UserStorer) {
	ctx := context.Background()

	user, err := us.GetByEmail(context.Background(), changePassEmail)
	if err != nil {
		log.Fatal(err)
	}
	if user == nil {
		log.Fatal("User D.N.E.")
	}

	passwordHash, err := pass.GenerateHashFromPassword(changePassPassword)
	if err != nil {
		log.Fatal("HashPassword:", err)
	}
	user.PasswordHash = passwordHash

	us.UpdateByID(ctx, user)

	fmt.Print("\033[H\033[2J")
	fmt.Println("Password successfully changed")
}

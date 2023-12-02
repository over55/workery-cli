package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"log/slog"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/over55/workery-cli/adapter/storage/mongodb"
	"github.com/over55/workery-cli/adapter/storage/postgres"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	user_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importUserCmd)
}

var importUserCmd = &cobra.Command{
	Use:   "import_user",
	Short: "Import the user from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		defaultLogger := slog.Default()
		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		userStorer := user_ds.NewDatastore(cfg, defaultLogger, mc)
		RunImportUser(cfg, ppc, lpc, tenantStorer, userStorer)
	},
}

type OldUser struct {
	ID       uint64        `json:"id"`
	TenantID sql.NullInt64 `json:"franchise_id"`
	// password character varying(128) COLLATE pg_catalog."default" NOT NULL,
	// last_login timestamp with time zone,
	// is_superuser boolean NOT NULL,
	Email      string    `json:"email"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	DateJoined time.Time `json:"date_joined"`
	IsActive   bool      `json:"is_active"`
	// avatar character varying(100) COLLATE pg_catalog."default",
	LastModified time.Time `json:"last_modified"`
	// salt character varying(127) COLLATE pg_catalog."default",
	WasEmailActivated bool `json:"was_email_activated"`
	// pr_access_code character varying(127) COLLATE pg_catalog."default" NOT NULL,
	// pr_expiry_date timestamp with time zone NOT NULL,

	// IsArchived              bool   `json:"is_archived"`
}

// Function returns a paginated list of all type element items.
func ListAllUsers(db *sql.DB) ([]*OldUser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
	    id, email, first_name, last_name, date_joined, is_active, last_modified, was_email_activated, franchise_id
	FROM
	    workery_users
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Println("failled querring old database")
		return nil, err
	}

	var arr []*OldUser
	defer rows.Close()
	for rows.Next() {
		m := new(OldUser)
		err = rows.Scan(
			&m.ID,
			&m.Email,
			&m.FirstName,
			&m.LastName,
			&m.DateJoined,
			&m.IsActive,
			&m.LastModified,
			&m.WasEmailActivated,
			&m.TenantID,
		)
		if err != nil {
			log.Println("failled querring2 old database")
			panic(err)
		}
		arr = append(arr, m)
	}
	err = rows.Err()
	if err != nil {
		log.Println("failled querring3 old database")
		panic(err)
	}
	return arr, err
}

func RunImportUser(cfg *config.Conf, public *sql.DB, london *sql.DB, tenantStorer tenant_ds.TenantStorer, userStorer user_ds.UserStorer) {
	fmt.Println("Beginning importing users")
	data, err := ListAllUsers(public)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importUser(context.Background(), tenantStorer, userStorer, datum)
	}
	fmt.Println("Finished importing users")
}

const (
	UserActiveState          = 1
	UserInactiveState        = 0
	UserExecutiveRoleId      = 1
	UserManagementRoleId     = 2
	UserFrontlineStaffRoleId = 3
	UserStaffRoleId          = 3
	UserAssociateRoleId      = 4
	UserCustomerRoleId       = 5
)

type User struct {
	ID                primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	TenantID          primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	TenantName        string             `bson:"tenant_name" json:"tenant_name,omitempty"`
	Email             string             `bson:"email" json:"email,omitempty"`
	FirstName         string             `bson:"first_name" json:"first_name,omitempty"`
	LastName          string             `bson:"last_name" json:"last_name,omitempty"`
	Name              string             `bson:"name" json:"name,omitempty"`
	LexicalName       string             `bson:"lexical_name" json:"lexical_name,omitempty"`
	PasswordAlgorithm string             `bson:"password_algorithm" json:"password_algorithm,omitempty"`
	PasswordHash      string             `bson:"password_hash" json:"password_hash,omitempty"`
	State             int8               `bson:"state" json:"state,omitempty"`
	RoleId            int8               `bson:"role_id" json:"role_id,omitempty"`
	Timezone          string             `bson:"timezone" json:"timezone,omitempty"`
	CreatedTime       time.Time          `bson:"created_time" json:"created_time,omitempty"`
	ModifiedTime      time.Time          `bson:"modified_time" json:"modified_time,omitempty"`
	JoinedTime        time.Time          `bson:"joined_time" json:"joined_time,omitempty"`
	Salt              string             `bson:"salt" json:"salt,omitempty"`
	WasEmailActivated bool               `bson:"was_email_activated" json:"was_email_activated,omitempty"`
	PrAccessCode      string             `bson:"pr_access_code" json:"pr_access_code,omitempty"`
	PrExpiryTime      time.Time          `bson:"pr_expiry_time" json:"pr_expiry_time,omitempty"`
	OldID             uint64             `bson:"old_id" json:"old_id,omitempty"`
	AccessToken       string             `bson:"access_token" json:"access_token,omitempty"`
	RefreshToken      string             `bson:"refresh_token" json:"refresh_token,omitempty"`
}

func importUser(ctx context.Context, ts tenant_ds.TenantStorer, us user_ds.UserStorer, ou *OldUser) {
	var state int8 = UserInactiveState
	if ou.IsActive == true {
		state = UserActiveState
	}

	// BUGFIX: If no user tenant account associated with the account then
	//         assign it to london. This is why id=2.
	tenantID := sql.NullInt64{Int64: 2, Valid: true}
	if ou.TenantID.Valid == true {
		tenantID = sql.NullInt64{Int64: ou.TenantID.Int64, Valid: true}
	}

	tenant, err := ts.GetByOldID(ctx, uint64(tenantID.Int64))
	if err != nil {
		log.Fatal(err)
	}
	if tenant == nil {
		log.Fatal("missing tenant", tenantID)
	}

	name := strings.Replace(ou.FirstName+" "+ou.LastName, "   ", "", 0)
	name = strings.Replace(name, "  ", "", 0)
	lexicalName := ou.LastName + ", " + ou.FirstName
	lexicalName = strings.Replace(lexicalName, ", ,", ",", 0)
	lexicalName = strings.Replace(lexicalName, "  ", " ", 0)
	lexicalName = strings.Replace(lexicalName, ", , ", ", ", 0)
	lexicalName = strings.Replace(lexicalName, "   ", "", 0)

	m := &user_ds.User{
		OldID:            ou.ID,
		ID:               primitive.NewObjectID(),
		FirstName:        ou.FirstName,
		LastName:         ou.LastName,
		Name:             name,
		LexicalName:      lexicalName,
		Email:            ou.Email,
		JoinedTime:       ou.DateJoined,
		Status:           state,
		Timezone:         "America/Toronto",
		CreatedAt:        ou.DateJoined,
		ModifiedAt:       ou.LastModified,
		Salt:             "",
		WasEmailVerified: ou.WasEmailActivated,
		PrAccessCode:     "",
		PrExpiryTime:     time.Now(),
		TenantID:         tenant.ID,
	}
	if err := us.Create(ctx, m); err != nil {
		log.Panic(err)
	}
	fmt.Println("Imported user ID#", m.ID)
}

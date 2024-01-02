package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"log/slog"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/guregu/null.v4"

	"github.com/over55/workery-cli/adapter/storage/mongodb"
	"github.com/over55/workery-cli/adapter/storage/postgres"
	a_ds "github.com/over55/workery-cli/app/associate/datastore"
	aal_ds "github.com/over55/workery-cli/app/associateawaylog/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	u_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importAssociateAwayLogCmd)
}

var importAssociateAwayLogCmd = &cobra.Command{
	Use:   "import_associate_away_log",
	Short: "Import the associate away log from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabaseLondonSchemaName)

		defaultLogger := slog.Default()

		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		uStorer := u_ds.NewDatastore(cfg, defaultLogger, mc)
		aStorer := a_ds.NewDatastore(cfg, defaultLogger, mc)
		aalStorer := aal_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(ctx, cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportAssociateAwayLog(cfg, ppc, lpc, uStorer, aStorer, aalStorer, tenant)
	},
}

func RunImportAssociateAwayLog(cfg *config.Conf, public *sql.DB, london *sql.DB, uStorer u_ds.UserStorer, aStorer a_ds.AssociateStorer, aalStorer aal_ds.AssociateAwayLogStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing associate away logs")
	data, err := ListAllAssociateAwayLogs(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importAssociateAwayLog(context.Background(), uStorer, aStorer, aalStorer, tenant, datum)
	}
	fmt.Println("Finished importing associate away logs")
}

type OldAssociateAwayLog struct {
	ID                 uint64      `json:"id"`
	AssociateID        uint64      `json:"associate_id"`
	Reason             int8        `json:"reason"`
	ReasonOther        null.String `json:"reason_other"`
	UntilFurtherNotice bool        `json:"until_further_notice"`
	UntilDate          null.Time   `json:"until_date"`
	StartDate          null.Time   `json:"start_date"`
	WasDeleted         bool        `json:"was_deleted"`
	CreatedTime        time.Time   `json:"created"`
	CreatedByID        null.Int    `json:"created_by_id"`
	LastModifiedTime   time.Time   `json:"last_modified"`
	LastModifiedByID   null.Int    `json:"last_modified_by_id"`
}

func ListAllAssociateAwayLogs(db *sql.DB) ([]*OldAssociateAwayLog, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
        id, associate_id, reason, reason_other, until_further_notice, until_date,
		start_date, was_deleted, created, created_by_id,
		last_modified, last_modified_by_id
	FROM
        london.workery_away_logs
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldAssociateAwayLog
	defer rows.Close()
	for rows.Next() {
		m := new(OldAssociateAwayLog)
		err = rows.Scan(
			&m.ID,
			&m.AssociateID,
			&m.Reason,
			&m.ReasonOther,
			&m.UntilFurtherNotice,
			&m.UntilDate,
			&m.StartDate,
			&m.WasDeleted,
			&m.CreatedTime,
			&m.CreatedByID,
			&m.LastModifiedTime,
			&m.LastModifiedByID,
		)
		if err != nil {
			log.Fatal("rows.Scan", err)
		}
		arr = append(arr, m)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal("rows.Err", err)
	}
	return arr, err
}

func importAssociateAwayLog(ctx context.Context, uStorer u_ds.UserStorer, aStorer a_ds.AssociateStorer, aalStorer aal_ds.AssociateAwayLogStorer, tenant *tenant_ds.Tenant, aal *OldAssociateAwayLog) {

	//
	// Lookup related.
	//

	a, err := aStorer.GetByPublicID(ctx, aal.AssociateID)
	if err != nil {
		log.Fatal(err)
	}
	if a == nil {
		log.Fatal("associate does not exist")
	}

	//
	// Set the `state`.
	//

	var state int8 = 1
	if aal.WasDeleted == true {
		state = 0
	}

	//
	// Get `createdByID` and `createdByName` values.
	//

	var createdByID primitive.ObjectID = primitive.NilObjectID
	var createdByName string
	if aal.CreatedByID.ValueOrZero() > 0 {
		user, err := uStorer.GetByPublicID(ctx, uint64(aal.CreatedByID.ValueOrZero()))
		if err != nil {
			log.Fatal("ur.GetByPublicID", err)
		}
		if user != nil {
			createdByID = user.ID
			createdByName = user.Name
		}

		// // For debugging purposes only.
		// log.Println("createdByID:", createdByID)
		// log.Println("createdByName:", createdByName)
	}

	//
	// Get `modifiedByID` and `modifiedByName` values.
	//

	var modifiedByID primitive.ObjectID = primitive.NilObjectID
	var modifiedByName string
	if aal.CreatedByID.ValueOrZero() > 0 {
		user, err := uStorer.GetByPublicID(ctx, uint64(aal.CreatedByID.ValueOrZero()))
		if err != nil {
			log.Fatal("ur.GetByPublicID", err)
		}
		if user != nil {
			modifiedByID = user.ID
			modifiedByName = user.Name
		}

		// // For debugging purposes only.
		// log.Println("modifiedByID:", modifiedByID)
		// log.Println("modifiedByName:", modifiedByName)
	}

	// Convert

	var ufn int8 = int8(aal_ds.UntilFurtherNoticeUnspecified)
	if aal.UntilFurtherNotice {
		ufn = int8(aal_ds.UntilFurtherNoticeYes)
	} else {
		ufn = int8(aal_ds.UntilFurtherNoticeNo)
	}

	//
	// Insert `AssociateAwayLog` record.
	//

	m := &aal_ds.AssociateAwayLog{
		PublicID:              aal.ID,
		ID:                    primitive.NewObjectID(),
		TenantID:              tenant.ID,
		AssociateID:           a.ID,
		AssociateName:         a.Name,
		AssociateLexicalName:  a.LexicalName,
		Reason:                aal.Reason,
		Status:                state,
		ReasonOther:           aal.ReasonOther.ValueOrZero(),
		UntilFurtherNotice:    ufn,
		UntilDate:             aal.UntilDate.ValueOrZero(),
		StartDate:             aal.StartDate.ValueOrZero(),
		CreatedAt:             aal.CreatedTime,
		CreatedByUserID:       createdByID,
		CreatedByUserName:     createdByName,
		CreatedFromIPAddress:  "",
		ModifiedAt:            aal.LastModifiedTime,
		ModifiedByUserID:      modifiedByID,
		ModifiedByUserName:    modifiedByName,
		ModifiedFromIPAddress: "",
	}

	if err := aalStorer.Create(ctx, m); err != nil {
		log.Panic(err)
	}

	//
	// Update the associate record.
	//

	m2 := &a_ds.AssociateAwayLog{
		PublicID:              aal.ID,
		ID:                    primitive.NewObjectID(),
		TenantID:              tenant.ID,
		AssociateID:           a.ID,
		AssociateName:         a.Name,
		AssociateLexicalName:  a.LexicalName,
		Reason:                aal.Reason,
		Status:                state,
		ReasonOther:           aal.ReasonOther.ValueOrZero(),
		UntilFurtherNotice:    ufn,
		UntilDate:             aal.UntilDate.ValueOrZero(),
		StartDate:             aal.StartDate.ValueOrZero(),
		CreatedAt:             aal.CreatedTime,
		CreatedByUserID:       createdByID,
		CreatedByUserName:     createdByName,
		CreatedFromIPAddress:  "",
		ModifiedAt:            aal.LastModifiedTime,
		ModifiedByUserID:      modifiedByID,
		ModifiedByUserName:    modifiedByName,
		ModifiedFromIPAddress: "",
	}
	a.AwayLogs = append(a.AwayLogs, m2)
	if err := aStorer.UpdateByID(ctx, a); err != nil {
		log.Panic(err)
	}

	fmt.Println("Imported AssociateAwayLog ID#", m2.ID, "for Associate ID", a.ID)
}

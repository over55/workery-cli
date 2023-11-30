package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"log/slog"

	"github.com/over55/workery-cli/adapter/storage/mongodb"
	"github.com/over55/workery-cli/adapter/storage/postgres"
	comm_ds "github.com/over55/workery-cli/app/comment/datastore"
	hh_ds "github.com/over55/workery-cli/app/howhear/datastore"
	cust_ds "github.com/over55/workery-cli/app/staff/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	user_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importStaffCommentCmd)
}

var importStaffCommentCmd = &cobra.Command{
	Use:   "import_staff_comment",
	Short: "Import the staff comments from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		defaultLogger := slog.Default()
		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		userStorer := user_ds.NewDatastore(cfg, defaultLogger, mc)
		custStorer := cust_ds.NewDatastore(cfg, defaultLogger, mc)
		hhStorer := hh_ds.NewDatastore(cfg, defaultLogger, mc)
		comStorer := comm_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(context.Background(), cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportStaffComment(cfg, ppc, lpc, tenantStorer, userStorer, custStorer, hhStorer, comStorer, tenant)
	},
}

type OldStaffComment struct {
	Id        uint64    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	StaffId   uint64    `json:"staff_id"`
	CommentId uint64    `json:"comment_id"`
}

func ListAllStaffComments(db *sql.DB) ([]*OldStaffComment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
	    id, created_at, about_id, comment_id
	FROM
	    london.workery_staff_comments
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldStaffComment
	defer rows.Close()
	for rows.Next() {
		m := new(OldStaffComment)
		err = rows.Scan(
			&m.Id,
			&m.CreatedAt,
			&m.StaffId,
			&m.CommentId,
		)
		if err != nil {
			log.Panic("ListAllStaffComments | Next | err", err)
		}
		arr = append(arr, m)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return arr, err
}

func RunImportStaffComment(cfg *config.Conf, public *sql.DB, london *sql.DB, tenantStorer tenant_ds.TenantStorer, userStorer user_ds.UserStorer, custStorer cust_ds.StaffStorer, hhStorer hh_ds.HowHearAboutUsItemStorer, comStorer comm_ds.CommentStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing staffs")
	data, err := ListAllStaffComments(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importStaffComment(context.Background(), tenantStorer, userStorer, custStorer, hhStorer, comStorer, tenant, datum)
	}
	fmt.Println("Finished importing staffs")
}

func importStaffComment(ctx context.Context, ts tenant_ds.TenantStorer, us user_ds.UserStorer, custStorer cust_ds.StaffStorer, hhStorer hh_ds.HowHearAboutUsItemStorer, comStorer comm_ds.CommentStorer, tenant *tenant_ds.Tenant, ou *OldStaffComment) {

	//
	// Lookup related.
	//

	staff, err := custStorer.GetByOldID(ctx, ou.StaffId)
	if err != nil {
		log.Fatal(err)
	}
	if staff == nil {
		log.Fatal("staff does not exist")
	}
	comment, err := comStorer.GetByOldID(ctx, ou.CommentId)
	if err != nil {
		log.Fatal(err)
	}
	if comment == nil {
		log.Fatal("comment does not exist")
	}

	//
	// Create the staff comment.
	//

	cc := &cust_ds.StaffComment{
		ID:                    comment.ID,
		TenantID:              comment.TenantID,
		CreatedAt:             comment.CreatedAt,
		CreatedByUserID:       comment.CreatedByUserID,
		CreatedByUserName:     comment.CreatedByUserName,
		CreatedFromIPAddress:  comment.CreatedFromIPAddress,
		ModifiedAt:            comment.ModifiedAt,
		ModifiedByUserID:      comment.ModifiedByUserID,
		ModifiedByUserName:    comment.ModifiedByUserName,
		ModifiedFromIPAddress: comment.ModifiedFromIPAddress,
		Content:               comment.Content,
		Status:                comment.Status,
		OldID:                 comment.OldID,
	}

	// Append comments to staff details.
	staff.Comments = append(staff.Comments, cc)

	if err := custStorer.UpdateByID(ctx, staff); err != nil {
		log.Fatal(err)
	}

	//
	// Update the comment.
	//

	comment.BelongsTo = comm_ds.BelongsToStaff
	comment.StaffID = staff.ID
	comment.StaffName = staff.Name
	if err := comStorer.UpdateByID(ctx, comment); err != nil {
		log.Fatal(err)
	}

	//
	// For debugging purposes only.
	//

	fmt.Println("Imported Staff Comment ID#", cc.ID, "for StaffID", staff.ID)
}

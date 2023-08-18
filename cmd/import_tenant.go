package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slog"

	"github.com/over55/workery-cli/adapter/storage/mongodb"
	"github.com/over55/workery-cli/adapter/storage/postgres"
	"github.com/over55/workery-cli/app/tenant/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importTenantCmd)
}

var importTenantCmd = &cobra.Command{
	Use:   "import_tenant",
	Short: "Import the franchise from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)

		defaultLogger := slog.Default()

		tenantStorer := datastore.NewDatastore(cfg, defaultLogger, mc)
		RunImportTenant(cfg, ppc, lpc, tenantStorer)
	},
}

type OldTenant struct {
	Id                      uint64          `json:"id"`
	SchemaName              string          `json:"schema_name"`
	Created                 time.Time       `json:"created"`
	LastModified            time.Time       `json:"last_modified"`
	AlternateName           string          `json:"alternate_name"`
	Description             string          `json:"description"`
	Name                    string          `json:"name"`
	Url                     sql.NullString  `json:"url"`
	AreaServed              sql.NullString  `json:"area_served"`
	AvailableLanguage       sql.NullString  `json:"available_language"`
	ContactType             sql.NullString  `json:"contact_type"`
	Email                   sql.NullString  `json:"email"`
	FaxNumber               sql.NullString  `json:"fax_number"`
	Telephone               sql.NullString  `json:"telephone"`
	TelephoneTypeOf         int8            `json:"telephone_type_of"`
	TelephoneExtension      sql.NullString  `json:"telephone_extension"`
	OtherTelephone          sql.NullString  `json:"other_telephone"`
	OtherTelephoneExtension sql.NullString  `json:"other_telephone_extension"`
	OtherTelephoneTypeOf    int8            `json:"other_telephone_type_of"`
	AddressCountry          string          `json:"address_country"`
	AddressRegion           string          `json:"address_region"`
	AddressLocality         string          `json:"address_locality"`
	PostOfficeBoxNumber     string          `json:"post_office_box_number"`
	PostalCode              string          `json:"postal_code"`
	StreetAddress           string          `json:"street_address"`
	StreetAddressExtra      string          `json:"street_address_extra"`
	Elevation               sql.NullFloat64 `json:"elevation"`
	Latitude                sql.NullFloat64 `json:"latitude"`
	Longitude               sql.NullFloat64 `json:"longitude"`
	TimezoneName            string          `json:"timestamp_name"`
	IsArchived              bool            `json:"is_archived"`
}

var (
	TenantActiveState  = int8(1)
	TenantArchiveState = int8(2)
)

type Tenant struct {
	ID                      primitive.ObjectID `bson:"_id" json:"id"`
	Uuid                    string             `bson:"uuid" json:"uuid"`
	SchemaName              string             `bson:"schema_name" json:"schema_name"`
	AlternateName           string             `bson:"alternate_name" json:"alternate_name"`
	Description             string             `bson:"description" json:"description"`
	Name                    string             `bson:"name" json:"name"`
	Url                     string             `bson:"url" json:"url"`
	State                   int8               `bson:"state" json:"state"`
	Timezone                string             `bson:"timestamp" json:"timestamp"`
	CreatedTime             time.Time          `bson:"created_time" json:"created_time"`
	ModifiedTime            time.Time          `bson:"modified_time" json:"modified_time"`
	AddressCountry          string             `bson:"address_country" json:"address_country"`
	AddressRegion           string             `bson:"address_region" json:"address_region"`
	AddressLocality         string             `bson:"address_locality" json:"address_locality"`
	PostOfficeBoxNumber     string             `bson:"post_office_box_number" json:"post_office_box_number"`
	PostalCode              string             `bson:"postal_code" json:"postal_code"`
	StreetAddress           string             `bson:"street_address" json:"street_address"`
	StreetAddressExtra      string             `bson:"street_address_extra" json:"street_address_extra"`
	Elevation               float64            `bson:"elevation" json:"elevation"`
	Latitude                float64            `bson:"latitude" json:"latitude"`
	Longitude               float64            `bson:"longitude" json:"longitude"`
	AreaServed              string             `bson:"area_served" json:"area_served"`
	AvailableLanguage       string             `bson:"available_language" json:"available_language"`
	ContactType             string             `bson:"contact_type" json:"contact_type"`
	Email                   string             `bson:"email" json:"email"`
	FaxNumber               string             `bson:"fax_number" json:"fax_number"`
	Telephone               string             `bson:"telephone" json:"telephone"`
	TelephoneTypeOf         int8               `bson:"telephone_type_of" json:"telephone_type_of"`
	TelephoneExtension      string             `bson:"telephone_extension" json:"telephone_extension"`
	OtherTelephone          string             `bson:"other_telephone" json:"other_telephone"`
	OtherTelephoneExtension string             `bson:"other_telephone_extension" json:"other_telephone_extension"`
	OtherTelephoneTypeOf    int8               `bson:"other_telephone_type_of" json:"other_telephone_type_of"`
	OldId                   uint64             `bson:"old_id" json:"old_id"`
}

// Function returns a paginated list of all type element items.
func ListAllTenants(db *sql.DB) ([]*OldTenant, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
	    id, schema_name, created, last_modified, alternate_name, description,
		name, url, area_served, available_language, contact_type, email,
		fax_number, telephone, telephone_type_of, telephone_extension,
		other_telephone, other_telephone_extension, other_telephone_type_of,
		address_country, address_region, address_locality, post_office_box_number,
		postal_code, street_address, street_address_extra, elevation, latitude,
		longitude, timezone_name, is_archived
	FROM
	    workery_franchises
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Println("failled querring old database")
		return nil, err
	}

	var arr []*OldTenant
	defer rows.Close()
	for rows.Next() {
		m := new(OldTenant)
		err = rows.Scan(
			&m.Id,
			&m.SchemaName,
			&m.Created,
			&m.LastModified,
			&m.AlternateName,
			&m.Description,
			&m.Name,
			&m.Url,
			&m.AreaServed,
			&m.AvailableLanguage,
			&m.ContactType,
			&m.Email,
			&m.FaxNumber,
			&m.Telephone,
			&m.TelephoneTypeOf,
			&m.TelephoneExtension,
			&m.OtherTelephone,
			&m.OtherTelephoneExtension,
			&m.OtherTelephoneTypeOf,
			&m.AddressCountry,
			&m.AddressRegion,
			&m.AddressLocality,
			&m.PostOfficeBoxNumber,
			&m.PostalCode,
			&m.StreetAddress,
			&m.StreetAddressExtra,
			&m.Elevation,
			&m.Latitude,
			&m.Longitude,
			&m.TimezoneName,
			&m.IsArchived,
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

func RunImportTenant(cfg *config.Conf, public *sql.DB, london *sql.DB, tenantStorer datastore.TenantStorer) {
	fmt.Println("Beginning importing tenants")
	tt, err := ListAllTenants(public)
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range tt {
		importTenant(context.Background(), tenantStorer, t)
		// runTenantInsert(v, r)
	}
	fmt.Println("Finished importing tenants")
}

func importTenant(ctx context.Context, tenantStorer datastore.TenantStorer, t *OldTenant) {
	m := &datastore.Tenant{
		OldId:              t.Id,
		ID:                 primitive.NewObjectID(),
		AlternateName:      t.AlternateName,
		Description:        t.Description,
		Name:               t.Name,
		Url:                t.Url.String,
		Status:             1,
		Timezone:           "America/Toronto",
		CreatedTime:        t.Created,
		ModifiedTime:       t.LastModified,
		AddressCountry:     t.AddressCountry,
		AddressRegion:      t.AddressRegion,
		AddressLocality:    t.AddressLocality,
		PostalCode:         t.PostalCode,
		StreetAddress:      t.StreetAddress,
		StreetAddressExtra: t.StreetAddressExtra,
		SchemaName:         t.SchemaName,
	}
	if err := tenantStorer.Create(ctx, m); err != nil {
		log.Panic(err)
	}
	fmt.Println("Imported tenant ID#", m.ID)
}

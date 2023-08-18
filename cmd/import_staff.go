package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slog"
	"gopkg.in/guregu/null.v4"

	"github.com/over55/workery-cli/adapter/storage/mongodb"
	"github.com/over55/workery-cli/adapter/storage/postgres"
	hh_ds "github.com/over55/workery-cli/app/howhear/datastore"
	s_ds "github.com/over55/workery-cli/app/staff/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	user_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importStaffCmd)
}

var importStaffCmd = &cobra.Command{
	Use:   "import_staff",
	Short: "Import the staff from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		defaultLogger := slog.Default()
		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		userStorer := user_ds.NewDatastore(cfg, defaultLogger, mc)
		sStorer := s_ds.NewDatastore(cfg, defaultLogger, mc)
		hhStorer := hh_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(context.Background(), cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportStaff(cfg, ppc, lpc, tenantStorer, userStorer, sStorer, hhStorer, tenant)
	},
}

type OldStaff struct {
	Created                              time.Time   `json:"created"`
	LastModified                         time.Time   `json:"last_modified"`
	AvailableLanguage                    null.String `json:"available_language"`
	ContactType                          null.String `json:"contact_type"`
	Email                                null.String `json:"email"`
	FaxNumber                            null.String `json:"fax_number"`
	Telephone                            null.String `json:"telephone"`
	TelephoneTypeOf                      int8        `json:"telephone_type_of"`
	TelephoneExtension                   null.String `json:"telephone_extension"`
	OtherTelephone                       null.String `json:"other_telephone"`
	OtherTelephoneExtension              null.String `json:"other_telephone_extension"`
	OtherTelephoneTypeOf                 int8        `json:"other_telephone_type_of"`
	AddressCountry                       string      `json:"address_country"`
	AddressLocality                      string      `json:"address_locality"`
	AddressRegion                        string      `json:"address_region"`
	PostOfficeBoxNumber                  null.String `json:"post_office_box_number"`
	PostalCode                           null.String `json:"postal_code"`
	StreetAddress                        string      `json:"street_address"`
	StreetAddressExtra                   null.String `json:"street_address_extra"`
	Elevation                            null.Float  `json:"elevation"`
	Latitude                             null.Float  `json:"latitude"`
	Longitude                            null.Float  `json:"longitude"`
	GivenName                            null.String `json:"given_name"`
	MiddleName                           null.String `json:"middle_name"`
	LastName                             null.String `json:"last_name"`
	Birthdate                            null.Time   `json:"birthdate"`
	JoinDate                             null.Time   `json:"join_date"`
	Nationality                          null.String `json:"nationality"`
	Gender                               null.String `json:"gender"`
	TaxID                                null.String `json:"tax_id"`
	ID                                   uint64      `json:"id"`
	IndexedText                          null.String `json:"indexed_text"`
	CreatedFrom                          null.String `json:"created_from"`
	CreatedFromIsPublic                  bool        `json:"created_from_is_public"`
	LastModifiedFrom                     null.String `json:"last_modified_from"`
	LastModifiedFromIsPublic             bool        `json:"last_modified_from_is_public"`
	IsArchived                           bool        `json:"is_archived"`
	CreatedByID                          null.Int    `json:"created_by_id"`
	LastModifiedByID                     null.Int    `json:"last_modified_by_id"`
	OwnerID                              null.Int    `json:"owner_id"`
	HowHearOther                         null.String `json:"how_hear_other"`
	HowHearID                            null.Int    `json:"how_hear_id"`
	AvatarImageID                        null.Int    `json:"avatar_image_id"`
	PersonalEmail                        null.String `json:"personal_email"`
	EmergencyContactAlternativeTelephone null.String `json:"emergency_contact_alternative_telephone"`
	EmergencyContactName                 null.String `json:"emergency_contact_name"`
	EmergencyContactRelationship         null.String `json:"emergency_contact_relationship"`
	EmergencyContactTelephone            null.String `json:"emergency_contact_telephone"`
	PoliceCheck                          null.Time   `json:"police_check"`
	Description                          null.String `json:"description"`
	// TypeOf                   int8            `json:"type_of"`
	// IsOkToEmail              bool            `json:"is_ok_to_email"`
	// IsOkToText               bool            `json:"is_ok_to_text"`
	// IsBusiness               bool            `json:"is_business"`
	// IsSenior                 bool            `json:"is_senior"`
	// IsSupport                bool            `json:"is_support"`
	// JobInfoRead              null.String  `json:"job_info_read"`
	// OrganizationID           null.Int   `json:"organization_id"`
	// IsBlacklisted            bool            `json:"is_blacklisted"`
	// DeactivationReason       int8            `json:"deactivation_reason"`
	// DeactivationReasonOther  string          `json:"deactivation_reason_other"`
	// State                    string          `json:"state"`
	// HowHearOld               int8            `json:"how_hear_old"`
	// OrganizationName         null.String  `json:"organization_name"`
	// OrganizationTypeOf       int8            `json:"organization_type_of"`
}

func ListAllStaffs(db *sql.DB) ([]*OldStaff, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
	    id, created, last_modified, available_language, contact_type, email, fax_number,
		telephone, telephone_type_of, telephone_extension,
		other_telephone, other_telephone_extension, other_telephone_type_of,
		address_country, address_locality, address_region, post_office_box_number,
		postal_code, street_address, street_address_extra, elevation, latitude,
		longitude, given_name, middle_name, last_name, birthdate, join_date,
		nationality, gender, tax_id, id, indexed_text,
		created_from, created_from_is_public, last_modified_from,
		last_modified_from_is_public, is_archived, created_by_id, last_modified_by_id,
		owner_id, how_hear_other, how_hear_id, avatar_image_id, personal_email,
		emergency_contact_alternative_telephone, emergency_contact_name,
		emergency_contact_relationship, emergency_contact_telephone, police_check,
		description
	FROM
	    london.workery_staff
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldStaff
	defer rows.Close()
	for rows.Next() {
		m := new(OldStaff)
		err = rows.Scan(
			&m.ID, &m.Created, &m.LastModified, &m.AvailableLanguage, &m.ContactType, &m.Email, &m.FaxNumber,
			&m.Telephone, &m.TelephoneTypeOf, &m.TelephoneExtension,
			&m.OtherTelephone, &m.OtherTelephoneExtension, &m.OtherTelephoneTypeOf,
			&m.AddressCountry, &m.AddressLocality, &m.AddressRegion, &m.PostOfficeBoxNumber,
			&m.PostalCode, &m.StreetAddress, &m.StreetAddressExtra, &m.Elevation,
			&m.Latitude, &m.Longitude, &m.GivenName, &m.MiddleName, &m.LastName,
			&m.Birthdate, &m.JoinDate, &m.Nationality, &m.Gender, &m.TaxID, &m.ID, &m.IndexedText,
			&m.CreatedFrom, &m.CreatedFromIsPublic, &m.LastModifiedFrom,
			&m.LastModifiedFromIsPublic, &m.IsArchived, &m.CreatedByID, &m.LastModifiedByID,
			&m.OwnerID, &m.HowHearOther, &m.HowHearID, &m.AvatarImageID, &m.PersonalEmail,
			&m.EmergencyContactAlternativeTelephone, &m.EmergencyContactName,
			&m.EmergencyContactRelationship, &m.EmergencyContactTelephone, &m.PoliceCheck,
			&m.Description,
		)
		if err != nil {
			log.Fatal("(AA)", err)
		}
		arr = append(arr, m)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal("(BB)", err)
	}
	return arr, err
}

func RunImportStaff(cfg *config.Conf, public *sql.DB, london *sql.DB, tenantStorer tenant_ds.TenantStorer, userStorer user_ds.UserStorer, sStorer s_ds.StaffStorer, hhStorer hh_ds.HowHearAboutUsItemStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing staffs")
	data, err := ListAllStaffs(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importStaff(context.Background(), tenantStorer, userStorer, sStorer, hhStorer, tenant, datum)
	}
	fmt.Println("Finished importing staffs")
}

func importStaff(ctx context.Context, ts tenant_ds.TenantStorer, us user_ds.UserStorer, sStorer s_ds.StaffStorer, hhStorer hh_ds.HowHearAboutUsItemStorer, tenant *tenant_ds.Tenant, ou *OldStaff) {
	var status int8 = s_ds.StaffStatusArchived
	if ou.IsArchived == true {
		status = s_ds.StaffStatusActive
	}

	// // BUGFIX: If no user tenant account staffd with the account then
	// //         assign it to london. This is why id=2.
	// tenantID := sql.NullInt64{Int64: 2, Valid: true}
	// if ou.TenantID.Valid == true {
	// 	tenantID = sql.NullInt64{Int64: ou.TenantID.Int64, Valid: true}
	// }
	//
	// tenant, err := ts.GetByOldID(ctx, uint64(tenantID.Int64))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if tenant == nil {
	// 	log.Fatal("missing tenant", tenantID)
	// }

	lexicalName := ou.LastName.ValueOrZero() + ", " + ou.GivenName.ValueOrZero()
	lexicalName = strings.Replace(lexicalName, ", ,", ",", 0)
	lexicalName = strings.Replace(lexicalName, "  ", " ", 0)
	lexicalName = strings.Replace(lexicalName, ", , ", ", ", 0)

	//
	// Compile the `full address` and `address url`.
	//

	address := ""
	if ou.StreetAddress != "" && ou.StreetAddress != "-" {
		address += ou.StreetAddress
	}
	if ou.StreetAddressExtra.IsZero() != false && ou.StreetAddressExtra.ValueOrZero() != "" {
		address += ou.StreetAddressExtra.ValueOrZero()
	}
	if ou.StreetAddress != "" && ou.StreetAddress != "-" {
		address += ", "
	}
	address += ou.AddressLocality
	address += ", " + ou.AddressRegion
	address += ", " + ou.AddressCountry
	fullAddressWithoutPostalCode := address
	fullAddressWithPostalCode := "-"
	fullAddressUrl := ""
	if ou.PostalCode.String != "" {
		fullAddressWithPostalCode = address + ", " + ou.PostalCode.String
		fullAddressUrl = "https://www.google.com/maps/place/" + fullAddressWithPostalCode
	} else {
		fullAddressUrl = "https://www.google.com/maps/place/" + fullAddressWithoutPostalCode
	}

	//
	// Compile the `how hear` text.
	//

	howHearID := uint64(ou.HowHearID.Int64)
	howHearText := ""
	howHear, err := hhStorer.GetByOldID(ctx, uint64(howHearID))
	if err != nil {
		log.Fatal(err)
		return
	}
	if howHearID == 1 {
		// if ou.HowHearOther == "" {
		// 	howHearText = "-"
		// } else {
		// 	howHearText = ou.HowHearOther
		// }
	} else {
		howHearText = howHear.Text
	}

	//
	// Get created by
	//

	var createdByUserID primitive.ObjectID = primitive.NilObjectID
	var createdByUserName string
	createdByUser, _ := us.GetByOldID(ctx, uint64(ou.CreatedByID.ValueOrZero()))
	if createdByUser != nil {
		createdByUserID = createdByUser.ID
		createdByUserName = createdByUser.Name
	}

	//
	// Get modified by
	//

	var modifiedByUserID primitive.ObjectID = primitive.NilObjectID
	var modifiedByUserName string
	modifiedByUser, _ := us.GetByOldID(ctx, uint64(ou.LastModifiedByID.ValueOrZero()))
	if modifiedByUser != nil {
		modifiedByUserID = modifiedByUser.ID
		modifiedByUserName = modifiedByUser.Name
	}

	//
	// Empty arrays
	//

	cc := []*s_ds.StaffComment{}
	sss := []*s_ds.StaffSkillSet{}
	irs := []*s_ds.StaffInsuranceRequirement{}
	vts := []*s_ds.StaffVehicleType{}
	al := []*s_ds.StaffAwayLog{}
	at := []*s_ds.StaffTag{}

	//
	// Insert our `Staff` data.
	//

	m := &s_ds.Staff{
		OldID:                        ou.ID,
		ID:                           primitive.NewObjectID(),
		TenantID:                     tenant.ID,
		FirstName:                    ou.GivenName.ValueOrZero(),
		LastName:                     ou.LastName.ValueOrZero(),
		Name:                         ou.GivenName.ValueOrZero() + " " + ou.LastName.ValueOrZero(),
		LexicalName:                  lexicalName,
		Email:                        ou.Email.ValueOrZero(),
		Phone:                        ou.Telephone.ValueOrZero(),
		PhoneTypeOf:                  ou.TelephoneTypeOf,
		PhoneExtension:               ou.TelephoneExtension.ValueOrZero(),
		FaxNumber:                    ou.FaxNumber.ValueOrZero(),
		OtherPhone:                   ou.OtherTelephone.ValueOrZero(),
		OtherPhoneTypeOf:             ou.OtherTelephoneTypeOf,
		OtherPhoneExtension:          ou.OtherTelephoneExtension.ValueOrZero(),
		Country:                      ou.AddressCountry,
		Region:                       ou.AddressRegion,
		City:                         ou.AddressLocality,
		PostalCode:                   ou.PostalCode.ValueOrZero(),
		AddressLine1:                 ou.StreetAddress,
		AddressLine2:                 ou.StreetAddressExtra.ValueOrZero(),
		PostOfficeBoxNumber:          ou.PostOfficeBoxNumber.ValueOrZero(),
		FullAddressWithoutPostalCode: fullAddressWithoutPostalCode,
		FullAddressWithPostalCode:    fullAddressWithPostalCode,
		FullAddressURL:               fullAddressUrl,
		HowDidYouHearAboutUsID:       howHear.ID,
		HowDidYouHearAboutUsOther:    howHearText,
		HowDidYouHearAboutUsValue:    howHear.Text,
		AgreeTOS:                     true,
		CreatedAt:                    ou.Created,
		CreatedByUserID:              createdByUserID,
		CreatedByUserName:            createdByUserName,
		CreatedFromIPAddress:         ou.CreatedFrom.String,
		ModifiedAt:                   ou.LastModified,
		ModifiedByUserID:             modifiedByUserID,
		ModifiedByUserName:           modifiedByUserName,
		ModifiedFromIPAddress:        ou.LastModifiedFrom.String,
		Status:                       status,
		// Comments:              Comments // SKIP
		JoinedTime:     ou.JoinDate.ValueOrZero(),
		Timezone:       "American/Toronto",
		HasUserAccount: false,
		UserID:         primitive.NilObjectID,
		// TypeOf:         ou.TypeOf, //TODO:IMPL.
		IsOkToEmail: true,
		IsOkToText:  true,
		// IsBusiness:     ou.IsBusiness,
		// IsSenior:                ou.IsSenior,
		// IsSupport:               ou.IsSupport,
		// DeactivationReason:      ou.DeactivationReason,
		// DeactivationReasonOther: ou.DeactivationReasonOther,
		Description: ou.Description.ValueOrZero(),
		// AvatarObjectKey                      string             `bson:"avatar_object_key" json:"avatar_object_key"`
		// AvatarFileType                       string             `bson:"avatar_file_type" json:"avatar_file_type"`
		// AvatarFileName                       string             `bson:"avatar_file_name" json:"avatar_file_name"`
		Birthdate:   ou.Birthdate.ValueOrZero(),
		JoinDate:    ou.JoinDate.ValueOrZero(),
		Nationality: ou.Nationality.ValueOrZero(),
		Gender:      ou.Gender.ValueOrZero(),
		TaxID:       ou.TaxID.ValueOrZero(),
		Elevation:   ou.Elevation.ValueOrZero(),
		Latitude:    ou.Elevation.ValueOrZero(),
		Longitude:   ou.Longitude.ValueOrZero(),
		// AreaServed:            ou.AreaServed.ValueOrZero(),
		AvailableLanguage:     ou.AvailableLanguage.ValueOrZero(),
		ContactType:           ou.ContactType.ValueOrZero(),
		Tags:                  at,
		Comments:              cc,
		SkillSets:             sss,
		InsuranceRequirements: irs,
		VehicleTypes:          vts,
		AwayLogs:              al,
	}
	if err := sStorer.Create(ctx, m); err != nil {
		log.Panic(err)
	}
	fmt.Println("Imported user ID#", m.ID)
}

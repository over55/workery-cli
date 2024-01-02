package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func importStaff(
	ctx context.Context,
	ts tenant_ds.TenantStorer,
	us user_ds.UserStorer,
	sStorer s_ds.StaffStorer,
	hhStorer hh_ds.HowHearAboutUsItemStorer,
	tenant *tenant_ds.Tenant,
	ou *OldStaff,
) {
	//
	// Set the `state`.
	//

	var status int8 = s_ds.StaffStatusArchived
	if ou.IsArchived == true {
		status = s_ds.StaffStatusActive
	}

	//
	// Variable used to keep the ID of the user record in our database.
	//

	ownerUserID := uint64(ou.OwnerID.Int64)
	var userRoleID int8

	//
	// Generate our full name / lexical full name.
	//

	var name string
	var lexicalName string
	if ou.MiddleName.Valid {
		name = ou.GivenName.String + " " + ou.MiddleName.String + " " + ou.LastName.String
		lexicalName = ou.LastName.String + ", " + ou.MiddleName.String + ", " + ou.GivenName.String
	} else {
		name = ou.GivenName.String + " " + ou.LastName.String
		lexicalName = ou.LastName.String + ", " + ou.GivenName.String
	}
	lexicalName = strings.Replace(lexicalName, ", ,", ",", 0)
	lexicalName = strings.Replace(lexicalName, "  ", " ", 0)
	lexicalName = strings.Replace(lexicalName, ", , ", ", ", 0)

	//
	// Get user.
	//

	var ownerUser *user_ds.User

	// CASE 1: User record exists in our database.
	if ou.OwnerID.Valid {
		user, err := us.GetByPublicID(ctx, ownerUserID)
		if err != nil {
			log.Fatal("(A)", err)
		}
		if user == nil {
			log.Fatal("(B) User is null")
		}
		ownerUser = user

		// CASE 2: Record D.N.E.
	} else {
		var email string

		// CASE 2A: Email specified
		if ou.Email.Valid {
			email = ou.Email.String

			// CASE 2B: Email is not specified
		} else {
			staffIdStr := strconv.FormatUint(ou.ID, 10)
			email = "staff+" + staffIdStr + "@workery.ca"
		}

		user, err := us.GetByEmail(ctx, email)
		if err != nil {
			log.Panic("(C)", err)
		}

		if user == nil {
			um := &user_ds.User{
				ID:          primitive.NewObjectID(),
				FirstName:   ou.GivenName.String,
				LastName:    ou.LastName.String,
				Name:        name,
				LexicalName: lexicalName,
				Email:       email,
				// JoinedTime:        ou.DateJoined,
				Status:   status,
				Timezone: "America/Toronto",
				// CreatedTime:       ou.DateJoined,
				// ModifiedTime:      ou.LastModified,
				Salt:             "",
				WasEmailVerified: false,
				PrAccessCode:     "",
				PrExpiryTime:     time.Now(),
				TenantID:         tenant.ID,
				Role:             5, // Staff
			}
			err = us.UpsertByEmail(ctx, um)
			if err != nil {
				log.Panic("(D)", err)
			}
			user = um
		}

		ownerUser = user
	}

	userRoleID = ownerUser.Role

	// //
	// // Get `createdByID` and `createdByName` values.
	// //
	//
	// var createdByID primitive.ObjectID = primitive.NilObjectID
	// var createdByName string
	// if ou.CreatedByID.ValueOrZero() > 0 {
	// 	user, err := us.GetByPublicID(ctx, uint64(ou.CreatedByID.ValueOrZero()))
	// 	if err != nil {
	// 		log.Fatal("ur.GetByPublicID", err)
	// 	}
	// 	if user != nil {
	// 		createdByID = user.ID
	// 		createdByName = user.Name
	// 	}
	// }
	//
	// //
	// // Get `lastModifiedById` and `lastModifiedByName` values.
	// //
	//
	// var lastModifiedById null.Int
	// var lastModifiedByName null.String
	// if ou.LastModifiedByID.ValueOrZero() > 0 {
	// 	userId, err := ur.GetIdByPublicID(ctx, uint64(ou.LastModifiedByID.ValueOrZero()))
	// 	if err != nil {
	// 		log.Panic("ur.GetIdByOldId", err)
	// 	}
	// 	user, err := ur.GetById(ctx, tid, userId)
	// 	if err != nil {
	// 		log.Panic("ur.GetById", err)
	// 	}
	//
	// 	if user != nil {
	// 		lastModifiedById = null.IntFrom(int64(userId))
	// 		lastModifiedByName = null.StringFrom(user.Name)
	// 	} else {
	// 		log.Println("WARNING: D.N.E.")
	// 	}
	// }

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

	howHearId := uint64(ou.HowHearID.Int64)
	howHearText := ""
	isHowHearOther := false
	howHear, err := hhStorer.GetByPublicID(ctx, uint64(howHearId))
	if err != nil {
		log.Fatal(err)
		return
	}
	if howHear != nil {
		if howHearId == 1 {
			if ou.HowHearOther.ValueOrZero() == "" {
				howHearText = "-"
			} else {
				howHearText = ou.HowHearOther.ValueOrZero()
				isHowHearOther = true
			}
		} else {
			howHearText = howHear.Text
		}
	} else {
		howHear, err = hhStorer.GetByText(ctx, "Other")
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	// Defensive code.
	if howHear == nil {
		log.Fatal("how hear does not exist")
	}

	//
	// Get created by
	//

	var createdByUserID primitive.ObjectID = primitive.NilObjectID
	var createdByUserName string
	createdByUser, _ := us.GetByPublicID(ctx, uint64(ou.CreatedByID.ValueOrZero()))
	if createdByUser != nil {
		createdByUserID = createdByUser.ID
		createdByUserName = createdByUser.Name
	}

	//
	// Get modified by
	//

	var modifiedByUserID primitive.ObjectID = primitive.NilObjectID
	var modifiedByUserName string
	modifiedByUser, _ := us.GetByPublicID(ctx, uint64(ou.LastModifiedByID.ValueOrZero()))
	if modifiedByUser != nil {
		modifiedByUserID = modifiedByUser.ID
		modifiedByUserName = modifiedByUser.Name
	}

	//
	// Empty arrays
	//

	cc := make([]*s_ds.StaffComment, 0)
	sss := make([]*s_ds.StaffSkillSet, 0)
	irs := make([]*s_ds.StaffInsuranceRequirement, 0)
	vts := make([]*s_ds.StaffVehicleType, 0)
	al := make([]*s_ds.StaffAwayLog, 0)
	at := make([]*s_ds.StaffTag, 0)

	//
	// Gender
	//

	var gender int8
	if ou.Gender.ValueOrZero() == "male" {
		gender = s_ds.StaffGenderMan
	} else if ou.Gender.ValueOrZero() == "female" {
		gender = s_ds.StaffGenderWoman
	} else if ou.Gender.ValueOrZero() == "prefer not to say" {
		gender = s_ds.StaffGenderPreferNotToSay
	}

	//
	// Insert our `Staff` data.
	//

	m := &s_ds.Staff{
		ID:                           primitive.NewObjectID(),
		TenantID:                     tenant.ID,
		FirstName:                    ou.GivenName.ValueOrZero(),
		LastName:                     ou.LastName.ValueOrZero(),
		Name:                         name,
		LexicalName:                  lexicalName,
		Email:                        ou.Email.ValueOrZero(),
		Phone:                        ou.Telephone.ValueOrZero(),
		PhoneType:                    ou.TelephoneTypeOf,
		PhoneExtension:               ou.TelephoneExtension.ValueOrZero(),
		FaxNumber:                    ou.FaxNumber.ValueOrZero(),
		OtherPhone:                   ou.OtherTelephone.ValueOrZero(),
		OtherPhoneType:               ou.OtherTelephoneTypeOf,
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
		IsHowDidYouHearAboutUsOther:  isHowHearOther,
		HowDidYouHearAboutUsText:     howHear.Text,
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
		Timezone:       "American/Toronto",
		HasUserAccount: false,
		UserID:         ownerUser.ID,
		Type:           userRoleID,
		IsOkToEmail:    true,
		IsOkToText:     true,
		// IsBusiness:     ou.IsBusiness,
		// IsSenior:                ou.IsSenior,
		// IsSupport:               ou.IsSupport,
		// DeactivationReason:      ou.DeactivationReason,
		// DeactivationReasonOther: ou.DeactivationReasonOther,
		Description: ou.Description.ValueOrZero(),
		// AvatarObjectKey                      string             `bson:"avatar_object_key" json:"avatar_object_key"`
		// AvatarFileType                       string             `bson:"avatar_file_type" json:"avatar_file_type"`
		// AvatarFileName                       string             `bson:"avatar_file_name" json:"avatar_file_name"`
		BirthDate:   ou.Birthdate.ValueOrZero(),
		JoinDate:    ou.JoinDate.ValueOrZero(),
		Nationality: ou.Nationality.ValueOrZero(),
		Gender:      gender,
		TaxID:       ou.TaxID.ValueOrZero(),
		Elevation:   ou.Elevation.ValueOrZero(),
		Latitude:    ou.Elevation.ValueOrZero(),
		Longitude:   ou.Longitude.ValueOrZero(),
		// AreaServed:            ou.AreaServed.ValueOrZero(),
		PreferredLanguage:                    ou.AvailableLanguage.ValueOrZero(),
		ContactType:                          ou.ContactType.ValueOrZero(),
		Tags:                                 at,
		Comments:                             cc,
		SkillSets:                            sss,
		InsuranceRequirements:                irs,
		VehicleTypes:                         vts,
		AwayLogs:                             al,
		PersonalEmail:                        ou.PersonalEmail.ValueOrZero(),
		EmergencyContactAlternativeTelephone: ou.EmergencyContactAlternativeTelephone.ValueOrZero(),
		EmergencyContactName:                 ou.EmergencyContactName.ValueOrZero(),
		EmergencyContactRelationship:         ou.EmergencyContactRelationship.ValueOrZero(),
		EmergencyContactTelephone:            ou.EmergencyContactTelephone.ValueOrZero(),
		PoliceCheck:                          ou.PoliceCheck.ValueOrZero(),
	}
	if err := sStorer.Create(ctx, m); err != nil {
		log.Panic(err)
	}
	fmt.Println("Imported staff ID#", m.ID)
}

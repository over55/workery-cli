package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/guregu/null.v4"

	"github.com/over55/workery-cli/adapter/storage/mongodb"
	"github.com/over55/workery-cli/adapter/storage/postgres"
	a_ds "github.com/over55/workery-cli/app/associate/datastore"
	hh_ds "github.com/over55/workery-cli/app/howhear/datastore"
	sf_ds "github.com/over55/workery-cli/app/servicefee/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	user_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importAssociateCmd)
}

var importAssociateCmd = &cobra.Command{
	Use:   "import_associate",
	Short: "Import the associate from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		defaultLogger := slog.Default()
		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		userStorer := user_ds.NewDatastore(cfg, defaultLogger, mc)
		aStorer := a_ds.NewDatastore(cfg, defaultLogger, mc)
		hhStorer := hh_ds.NewDatastore(cfg, defaultLogger, mc)
		sfStorer := sf_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(context.Background(), cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportAssociate(cfg, ppc, lpc, tenantStorer, userStorer, aStorer, hhStorer, sfStorer, tenant)
	},
}

type OldAssociate struct {
	Created                              time.Time   `json:"created"`
	LastModified                         time.Time   `json:"last_modified"`
	AlternateName                        null.String `json:"alternate_name"`
	Description                          null.String `json:"description"`
	Name                                 null.String `json:"name"`
	Url                                  null.String `json:"url"`
	AreaServed                           null.String `json:"area_served"`
	AvailableLanguage                    null.String `json:"available_language"`
	ContactType                          null.String `json:"contact_type"`
	Email                                null.String `json:"email"`
	FaxNumber                            null.String `json:"fax_number"`
	ProductSupported                     null.String `json:"product_supported"`
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
	TypeOf                               int8        `json:"type_of"`
	IsOkToEmail                          bool        `json:"is_ok_to_email"`
	IsOkToText                           bool        `json:"is_ok_to_text"`
	CreatedFrom                          null.String `json:"created_from"`
	CreatedFromIsPublic                  bool        `json:"created_from_is_public"`
	LastModifiedFrom                     null.String `json:"last_modified_from"`
	LastModifiedFromIsPublic             bool        `json:"last_modified_from_is_public"`
	CreatedByID                          null.Int    `json:"created_by_id"`
	LastModifiedByID                     null.Int    `json:"last_modified_by_id"`
	OwnerID                              null.Int    `json:"owner_id"`
	HowHearOther                         string      `json:"how_hear_other"`
	IsArchived                           bool        `json:"is_archived"`
	HourlySalaryDesired                  null.Int    `json:"hourly_salary_desired"`
	LimitSpecial                         null.String `json:"limit_special"`
	DuesDate                             null.Time   `json:"dues_date"`
	CommercialInsuranceExpiryDate        null.Time   `json:"commercial_insurance_expiry_date"`
	AutoInsuranceExpiryDate              null.Time   `json:"auto_insurance_expiry_date"`
	WsibNumber                           null.String `json:"wsib_number"`
	WsibInsuranceDate                    null.Time   `json:"wsib_insurance_date"`
	PoliceCheck                          null.Time   `json:"police_check"`
	DriversLicenseClass                  null.String `json:"drivers_license_class"`
	HowHearID                            null.Int    `json:"how_hear_id"`
	HowHearOld                           int8        `json:"how_hear_old"`
	OrganizationName                     null.String `json:"organization_name"`
	OrganizationTypeOf                   int8        `json:"organization_type_of"`
	AvatarImageID                        null.Int    `json:"avatar_image_id"`
	ServiceFeeID                         null.Int    `json:"service_fee_id"`
	EmergencyContactName                 null.String `json:"emergency_contact_name"`
	EmergencyContactRelationship         null.String `json:"emergency_contact_relationship"`
	EmergencyContactTelephone            null.String `json:"emergency_contact_telephone"`
	EmergencyContactAlternativeTelephone null.String `json:"emergency_contact_alternative_telephone"`
	BalanceOwingAmount                   float64     `json:"balance_owing_amount"`
}

func ListAllAssociates(db *sql.DB) ([]*OldAssociate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
	    id, created, last_modified, alternate_name, description, name, url,
		area_served, available_language, contact_type, email, fax_number,
		product_supported, telephone, telephone_type_of, telephone_extension,
		other_telephone, other_telephone_extension, other_telephone_type_of,
		address_country, address_locality, address_region, post_office_box_number,
		postal_code, street_address, street_address_extra, elevation, latitude,
		longitude, given_name, middle_name, last_name, birthdate, join_date,
		nationality, gender, tax_id, id, indexed_text, type_of, is_ok_to_email,
		is_ok_to_text,created_from, created_from_is_public, last_modified_from,
		last_modified_from_is_public, is_archived, created_by_id, last_modified_by_id,
		owner_id, how_hear_other,

		hourly_salary_desired, limit_special, dues_date, commercial_insurance_expiry_date,
		auto_insurance_expiry_date, wsib_number, wsib_insurance_date, police_check, drivers_license_class,

		how_hear_id, how_hear_old, organization_name,
		organization_type_of, avatar_image_id, service_fee_id,
		emergency_contact_name, emergency_contact_relationship,
		emergency_contact_telephone, emergency_contact_alternative_telephone,
		balance_owing_amount
	FROM
	    london.workery_associates
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldAssociate
	defer rows.Close()
	for rows.Next() {
		m := new(OldAssociate)
		err = rows.Scan(
			&m.ID, &m.Created, &m.LastModified, &m.AlternateName, &m.Description, &m.Name, &m.Url,
			&m.AreaServed, &m.AvailableLanguage, &m.ContactType, &m.Email, &m.FaxNumber,
			&m.ProductSupported, &m.Telephone, &m.TelephoneTypeOf, &m.TelephoneExtension,
			&m.OtherTelephone, &m.OtherTelephoneExtension, &m.OtherTelephoneTypeOf,
			&m.AddressCountry, &m.AddressLocality, &m.AddressRegion, &m.PostOfficeBoxNumber,
			&m.PostalCode, &m.StreetAddress, &m.StreetAddressExtra, &m.Elevation,
			&m.Latitude, &m.Longitude, &m.GivenName, &m.MiddleName, &m.LastName,
			&m.Birthdate, &m.JoinDate, &m.Nationality, &m.Gender, &m.TaxID, &m.ID, &m.IndexedText, &m.TypeOf,
			&m.IsOkToEmail, &m.IsOkToText,
			&m.CreatedFrom, &m.CreatedFromIsPublic, &m.LastModifiedFrom,
			&m.LastModifiedFromIsPublic, &m.IsArchived, &m.CreatedByID, &m.LastModifiedByID,
			&m.OwnerID, &m.HowHearOther,

			&m.HourlySalaryDesired, &m.LimitSpecial, &m.DuesDate, &m.CommercialInsuranceExpiryDate,
			&m.AutoInsuranceExpiryDate, &m.WsibNumber, &m.WsibInsuranceDate, &m.PoliceCheck, &m.DriversLicenseClass,

			&m.HowHearID, &m.HowHearOld, &m.OrganizationName,
			&m.OrganizationTypeOf, &m.AvatarImageID, &m.ServiceFeeID,
			&m.EmergencyContactName, &m.EmergencyContactRelationship, &m.EmergencyContactTelephone,
			&m.EmergencyContactAlternativeTelephone, &m.BalanceOwingAmount,
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

func RunImportAssociate(cfg *config.Conf, public *sql.DB, london *sql.DB, tenantStorer tenant_ds.TenantStorer, userStorer user_ds.UserStorer, aStorer a_ds.AssociateStorer, hhStorer hh_ds.HowHearAboutUsItemStorer, sfStorer sf_ds.ServiceFeeStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing associates")
	data, err := ListAllAssociates(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importAssociate(context.Background(), tenantStorer, userStorer, aStorer, hhStorer, sfStorer, tenant, datum)
	}
	fmt.Println("Finished importing associates")
}

func importAssociate(ctx context.Context, ts tenant_ds.TenantStorer, us user_ds.UserStorer, aStorer a_ds.AssociateStorer, hhStorer hh_ds.HowHearAboutUsItemStorer, sfStorer sf_ds.ServiceFeeStorer, tenant *tenant_ds.Tenant, ou *OldAssociate) {
	var status int8 = a_ds.AssociateStatusActive
	if ou.IsArchived == true {
		status = a_ds.AssociateStatusArchived
	}

	// // BUGFIX: If no user tenant account associated with the account then
	// //         assign it to london. This is why id=2.
	// tenantID := sql.NullInt64{Int64: 2, Valid: true}
	// if ou.TenantID.Valid == true {
	// 	tenantID = sql.NullInt64{Int64: ou.TenantID.Int64, Valid: true}
	// }
	//
	// tenant, err := ts.GetByPublicID(ctx, uint64(tenantID.Int64))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if tenant == nil {
	// 	log.Fatal("missing tenant", tenantID)
	// }

	name := strings.Replace(ou.GivenName.ValueOrZero()+" "+ou.LastName.ValueOrZero(), "   ", "", 0)
	name = strings.Replace(name, "  ", "", 0)
	lexicalName := ou.LastName.ValueOrZero() + ", " + ou.GivenName.ValueOrZero()
	lexicalName = strings.Replace(lexicalName, ", ,", ",", 0)
	lexicalName = strings.Replace(lexicalName, "  ", " ", 0)
	lexicalName = strings.Replace(lexicalName, ", , ", ", ", 0)
	lexicalName = strings.Replace(lexicalName, "   ", "", 0)

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
	isHowHearOther := false
	howHear, err := hhStorer.GetByPublicID(ctx, uint64(howHearID))
	if err != nil {
		log.Fatal(err)
		return
	}
	if howHearID == 1 {
		if ou.HowHearOther == "" {
			howHearText = "-"
		} else {
			howHearText = ou.HowHearOther
			isHowHearOther = true
		}
	} else {
		howHearText = howHear.Text
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

	cc := make([]*a_ds.AssociateComment, 0)
	sss := make([]*a_ds.AssociateSkillSet, 0)
	irs := make([]*a_ds.AssociateInsuranceRequirement, 0)
	vts := make([]*a_ds.AssociateVehicleType, 0)
	al := make([]*a_ds.AssociateAwayLog, 0)
	at := make([]*a_ds.AssociateTag, 0)

	//
	// Gender
	//

	var gender int8
	if ou.Gender.ValueOrZero() == "male" {
		gender = a_ds.AssociateGenderMan
	} else if ou.Gender.ValueOrZero() == "female" {
		gender = a_ds.AssociateGenderWoman
	} else if ou.Gender.ValueOrZero() == "prefer not to say" {
		gender = a_ds.AssociateGenderPreferNotToSay
	}

	//
	// Generate associate ID.
	//

	associateID := primitive.NewObjectID()

	//
	// Check for unique email.
	//

	u := &user_ds.User{}

	emailExists, err := us.CheckIfExistsByEmail(ctx, ou.Email.ValueOrZero())
	if err != nil {
		log.Panic(err)
	}
	if emailExists {
		u, err = us.GetByEmail(ctx, ou.Email.ValueOrZero())
		if err != nil {
			log.Panic(err)
		}
		u.Role = user_ds.UserRoleAssociate
		u.ReferenceID = associateID // Important!
		if err := us.UpdateByID(ctx, u); err != nil {
			log.Panic(err)
		}
	} else {
		//
		// Insert our `User` data.
		//

		u.ID = primitive.NewObjectID()
		u.TenantID = tenant.ID
		u.FirstName = ou.GivenName.ValueOrZero()
		u.LastName = ou.LastName.ValueOrZero()
		u.Name = fmt.Sprintf("%s %s", ou.GivenName.ValueOrZero(), ou.LastName.ValueOrZero())
		u.LexicalName = fmt.Sprintf("%s, %s", ou.LastName.ValueOrZero(), ou.GivenName.ValueOrZero())
		u.OrganizationName = ou.OrganizationName.ValueOrZero()
		u.OrganizationType = ou.OrganizationTypeOf
		if ou.Email.IsZero() {
			u.Email = ou.Email.ValueOrZero()
			u.Status = status
		} else {
			u.Email = fmt.Sprintf("associate_%s@workery.ca", u.ID.Hex())
			u.Status = user_ds.UserStatusArchived
		}
		u.PasswordHashAlgorithm = primitive.NewObjectID().Hex()
		u.PasswordHash = "MongoDB Primitive"
		u.Role = user_ds.UserRoleAssociate
		u.ReferenceID = associateID // Important!
		u.WasEmailVerified = true
		u.EmailVerificationCode = ""
		u.EmailVerificationExpiry = time.Now()
		u.Phone = ou.Telephone.ValueOrZero()
		u.Country = ou.AddressCountry
		u.Region = ou.AddressRegion
		u.City = ou.AddressLocality
		u.AgreeTOS = true
		u.AgreePromotionsEmail = true
		u.CreatedAt = time.Now()
		// u.CreatedByUserID = userID
		// u.CreatedByUserName = userName
		// u.CreatedFromIPAddress = ipAddress
		u.ModifiedAt = time.Now()
		// u.ModifiedByUserID = userID
		// u.ModifiedByUserName = userName
		// u.ModifiedFromIPAddress = ipAddress
		u.Status = user_ds.UserStatusActive
		u.Comments = make([]*user_ds.UserComment, 0)
		u.Salt = ""
		// u.JoinedTime = req.JoinDateDT
		u.PrAccessCode = ""
		u.PrExpiryTime = time.Now()
		u.PublicID = 0
		u.Timezone = "American/Toronto"
		u.OTPEnabled = false // impl.Config.AppServer.Enable2FAOnRegistration
		u.OTPVerified = false
		u.OTPValidated = false
		u.OTPSecret = ""
		u.OTPAuthURL = ""
		if err := us.Create(ctx, u); err != nil {
			log.Panic(err)
		}
	}

	//
	// Insert our `Associate` data.
	//

	m := &a_ds.Associate{
		ID:                           associateID,
		PublicID:                     ou.ID,
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
		Timezone:                     "American/Toronto",
		HasUserAccount:               false,
		UserID:                       u.ID,
		Type:                         ou.TypeOf,
		IsOkToEmail:                  ou.IsOkToEmail,
		IsOkToText:                   ou.IsOkToText,
		// IsBusiness:     ou.IsBusiness,
		// IsSenior:                ou.IsSenior,
		// IsSupport:               ou.IsSupport,
		// DeactivationReason:      ou.DeactivationReason,
		// DeactivationReasonOther: ou.DeactivationReasonOther,
		Description: ou.Description.ValueOrZero(),
		// AvatarObjectKey                      string             `bson:"avatar_object_key" json:"avatar_object_key"`
		// AvatarFileType                       string             `bson:"avatar_file_type" json:"avatar_file_type"`
		// AvatarFileName                       string             `bson:"avatar_file_name" json:"avatar_file_name"`
		BirthDate:                     ou.Birthdate.ValueOrZero(),
		JoinDate:                      ou.JoinDate.ValueOrZero(),
		Nationality:                   ou.Nationality.ValueOrZero(),
		Gender:                        gender,
		TaxID:                         ou.TaxID.ValueOrZero(),
		Elevation:                     ou.Elevation.ValueOrZero(),
		Latitude:                      ou.Elevation.ValueOrZero(),
		Longitude:                     ou.Longitude.ValueOrZero(),
		AreaServed:                    ou.AreaServed.ValueOrZero(),
		PreferredLanguage:             ou.AvailableLanguage.ValueOrZero(),
		ContactType:                   ou.ContactType.ValueOrZero(),
		OrganizationName:              ou.OrganizationName.ValueOrZero(),
		OrganizationType:              ou.OrganizationTypeOf,
		HourlySalaryDesired:           ou.HourlySalaryDesired.ValueOrZero(),
		LimitSpecial:                  ou.LimitSpecial.ValueOrZero(),
		DuesDate:                      ou.DuesDate.ValueOrZero(),
		CommercialInsuranceExpiryDate: ou.CommercialInsuranceExpiryDate.ValueOrZero(),
		AutoInsuranceExpiryDate:       ou.AutoInsuranceExpiryDate.ValueOrZero(),
		WsibNumber:                    ou.WsibNumber.ValueOrZero(),
		WsibInsuranceDate:             ou.WsibInsuranceDate.ValueOrZero(),
		PoliceCheck:                   ou.PoliceCheck.ValueOrZero(),
		DriversLicenseClass:           ou.DriversLicenseClass.ValueOrZero(),
		// ServiceFeeID:                   ou.ServiceFeeID,
		BalanceOwingAmount:                   ou.BalanceOwingAmount,
		EmergencyContactName:                 ou.EmergencyContactName.ValueOrZero(),
		EmergencyContactRelationship:         ou.EmergencyContactRelationship.ValueOrZero(),
		EmergencyContactTelephone:            ou.EmergencyContactTelephone.ValueOrZero(),
		EmergencyContactAlternativeTelephone: ou.EmergencyContactAlternativeTelephone.ValueOrZero(),
		Tags:                                 at,
		Comments:                             cc,
		SkillSets:                            sss,
		InsuranceRequirements:                irs,
		VehicleTypes:                         vts,
		AwayLogs:                             al,
	}

	//
	// Get service fee
	//

	if !ou.ServiceFeeID.IsZero() {
		sf, err := sfStorer.GetByPublicID(ctx, uint64(ou.ServiceFeeID.ValueOrZero()))
		if err != nil {
			log.Panic(err)
		}
		if sf != nil {
			m.ServiceFeeID = sf.ID
			m.ServiceFeeName = sf.Name
			m.ServiceFeePercentage = sf.Percentage
		}
	}

	//
	// Save the update.
	//

	if err := aStorer.Create(ctx, m); err != nil {
		log.Panic(err)
	}
	fmt.Println("Imported associate ID#", m.ID)
}

package constants

type key int

const (
	SessionIsAuthorized key = iota
	SessionSkipAuthorization
	SessionID
	SessionIPAddress
	SessionUser
	SessionUserCompanyName
	SessionUserRole
	SessionUserID
	SessionUserUUID
	SessionUserTimezone
	SessionUserName
	SessionUserFirstName
	SessionUserLastName
	SessionUserOrganizationID
	SessionUserOrganizationName
)

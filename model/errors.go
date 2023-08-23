package model

// General errors.
const (
	ErrUnauthorized           = Error("unauthorized")
	ErrInternal               = Error("internal error")
	ErrUserNotFound           = Error("user not found")
	ErrMovieNotFound          = Error("movie not found")
	ErrUserExists             = Error("user already exists")
	ErrUserIDRequired         = Error("user id required")
	ErrUserNameRequired       = Error("user's username required")
	ErrInvalidJSON            = Error("invalid json")
	ErrUserRequired           = Error("user required")
	ErrInvalidEntry           = Error("invalid Entry")
	ErrUserNotCached          = Error("User not saved")
	ErrUserNameEmpty          = Error("Username is Emptyername")
	ErrUserPasswordEmpty      = Error("User Password is Empty")
	ErrUsrDbUnreachable       = Error("Unable to get the UserDB")
	ErrSessionCookieSaveError = Error("could not save cookie")
	ErrIvalidRedirect         = Error("invalid redirect URL")
	ErrSessionCookieError     = Error("could not create cookie")

	ErrAppPasswordEmpty  = Error("App Password is Empty")
	ErrAppNotFound       = Error("app not found")
	ErrAppExists         = Error("app already exists")
	ErrAppNameRequired   = Error("app's appname required")
	ErrAppRequired       = Error("app required")
	ErrAppNotCached      = Error("App not saved")
	ErrAppNameEmpty      = Error("Appname is Empty")

	ErrSessionPasswordEmpty = Error("Session Password is Empty")
	ErrSessDbUnreachable    = Error("Unable to get the SessionDB")
	ErrSessionNotFound      = Error("session not found")
	ErrSessionExists        = Error("session already exists")
	ErrSessionIDRequired    = Error("session id required")
	ErrSessionNameRequired  = Error("Session name required")
	ErrSessionRequired      = Error("session required")
	ErrSessionNotCached     = Error("Session not saved")
	ErrSessionNameEmpty     = Error("Sessionname is Empty")

	ErrAccountDataPasswordEmpty = Error("Account Password is Empty")
	ErrAcctDbUnreachable    = Error("Unable to get the AccountDataDB")
	ErrAccountDataNotFound      = Error("AccountData not found")
	ErrAccountDataExists        = Error("AccountData already exists")
	ErrEmailNotVerified    = Error("Email not yet verified")
	ErrAccountDataNameRequired  = Error("AccountData name required")
	ErrAccountDataRequired    = Error("AccountData required")
	ErrReachedLimit      = Error("Cannot Exceed Limit")
	ErrPasswordConflict     = Error("Password not same as verify Password")
	ErrAccountDataNameEmpty     = Error("AccountDataname is Empty")

	ErrDatabaseNotFound = Error("Database not Available")
	ErrHistoryPasswordEmpty = Error("History Password is Empty")
	ErrHistDbUnreachable    = Error("Unable to get the HistoryDB")
	ErrHistoryNotFound      = Error("History not found")
	ErrHistoryRequired      = Error("History required")
	ErrHistoryExists        = Error("History already exists")
	ErrSomethingWentWrong   = Error("Something Went Wrong")
	ErrEmailNotFound   = Error("Email Address is Empty")
	ErrPasswordExpired     = Error("Change Password Link Expired")
	ErrHistoryNameEmpty     = Error("Historyname is Empty")
	ErrEmailExpired     = Error("Email verification link expired")
	ErrReferralNotActive = Error("Referral Deactivated")
	ErrReferralEmpty = Error("Referral is Empty")
)

// Error represents a User error.
type Error string

// Error returns the error message.
func (e Error) Error() string { return string(e) }
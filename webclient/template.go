package webclient

import (
	"encoding/json"
	"log"
	"net/http"
	"coinbitly.com/model"
)

// WebService is a user login-aware wrapper for a html/template.
type WebService struct {
}

// parseTemplate applies a given file to the body of the base template.
func NewWebService() *WebService {
	return &WebService{}
}

// Execute writes the template using the provided data, adding login and user
// information to the base template.
func (resp *WebService) Execute(w http.ResponseWriter, r *http.Request, data interface{}, usr interface{}, msg interface{}, code model.Error) error {
	d := struct {
		Data      interface{} `json:"data"`
		LogoutURL string      `json:"logoutUrl"`
		User      interface{} `json:"user"`
		Msg       interface{} `json:"msg"`
	}{
		Data:      data,
		LogoutURL: "/logout?redirect=" + r.URL.RequestURI(),
		User:      usr,
		Msg:       msg,
	}
	
	w.WriteHeader(ErrorStatusCode(string(code)))
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(d)
	if err != nil {
		log.Printf("mobileWarning %v\n", err)
	}
	return err
}

// LogError logs an error with the HTTP route information.
func LogError(r *http.Request, err error) {
	log.Printf("[http] error: %s %s: %s", r.Method, r.URL.Path, err)
}

// lookup of application error codes to HTTP status codes.
var codes = map[string]int{
	string(model.ErrUnauthorized):             http.StatusUnauthorized,
	string(model.ErrInternal):                 http.StatusInternalServerError,
	string(model.ErrUserNotFound):             http.StatusNotFound,
	string(model.ErrUserExists):               http.StatusBadRequest,
	string(model.ErrUserIDRequired):           http.StatusInternalServerError,
	string(model.ErrUserNameRequired):         http.StatusBadRequest,
	string(model.ErrInvalidJSON):              http.StatusInternalServerError,
	string(model.ErrUserRequired):             http.StatusBadRequest,
	string(model.ErrInvalidEntry):             http.StatusConflict,
	string(model.ErrUserNotCached):            http.StatusInternalServerError,
	string(model.ErrUserNameEmpty):            http.StatusBadRequest,
	string(model.ErrUserPasswordEmpty):        http.StatusBadRequest,
	string(model.ErrUsrDbUnreachable):         http.StatusInternalServerError,
	string(model.ErrSessionCookieSaveError):   http.StatusConflict,
	string(model.ErrIvalidRedirect):           http.StatusConflict,
	string(model.ErrSessionCookieError):       http.StatusConflict,
	string(model.ErrAppPasswordEmpty):         http.StatusBadRequest,
	string(model.ErrAppNotFound):              http.StatusNotFound,
	string(model.ErrAppExists):                http.StatusBadRequest,
	string(model.ErrAppNameRequired):          http.StatusBadRequest,
	string(model.ErrAppRequired):              http.StatusBadRequest,
	string(model.ErrAppNotCached):             http.StatusBadRequest,
	string(model.ErrAppNameEmpty):             http.StatusBadRequest,
	string(model.ErrSessionPasswordEmpty):     http.StatusInternalServerError,
	string(model.ErrSessDbUnreachable):        http.StatusInternalServerError,
	string(model.ErrSessionNotFound):          http.StatusNotFound,
	string(model.ErrSessionExists):            http.StatusInternalServerError,
	string(model.ErrSessionIDRequired):        http.StatusInternalServerError,
	string(model.ErrSessionNameRequired):      http.StatusInternalServerError,
	string(model.ErrSessionRequired):          http.StatusInternalServerError,
	string(model.ErrSessionNotCached):         http.StatusInternalServerError,
	string(model.ErrSessionNameEmpty):         http.StatusInternalServerError,
	string(model.ErrAccountDataPasswordEmpty): http.StatusBadRequest,
	string(model.ErrAcctDbUnreachable):        http.StatusInternalServerError,
	string(model.ErrAccountDataNotFound):      http.StatusNotFound,
	string(model.ErrAccountDataExists):        http.StatusBadRequest,
	string(model.ErrEmailNotVerified):         http.StatusBadRequest,
	string(model.ErrAccountDataNameRequired):  http.StatusBadRequest,
	string(model.ErrReachedLimit):             http.StatusBadRequest,
	string(model.ErrPasswordConflict):         http.StatusBadRequest,
	string(model.ErrAccountDataNameEmpty):     http.StatusBadRequest,
	string(model.ErrHistoryPasswordEmpty):     http.StatusBadRequest,
	string(model.ErrHistDbUnreachable):        http.StatusInternalServerError,
	string(model.ErrHistoryNotFound):          http.StatusInternalServerError,
	string(model.ErrSomethingWentWrong):       http.StatusConflict,
	string(model.ErrEmailNotFound):            http.StatusBadRequest,
	string(model.ErrPasswordExpired):          http.StatusBadRequest,
	string(model.ErrHistoryNameEmpty):         http.StatusBadRequest,
	string(model.ErrEmailExpired):             http.StatusBadRequest,
	string(model.ErrReferralNotActive):        http.StatusBadRequest,
	string(model.ErrReferralEmpty):            http.StatusBadRequest,
}

// ErrorStatusCode returns the associated HTTP status code for a model error code.
func ErrorStatusCode(code string) int {
	if v, ok := codes[code]; ok {
		return v
	}
	return http.StatusOK
}

// FromErrorStatusCode returns the associated WTF code for an HTTP status code.
func FromErrorStatusCode(code int) string {
	for k, v := range codes {
		if v == code {
			return k
		}
	}
	return string(model.ErrInternal)
}
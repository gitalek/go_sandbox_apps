package auth

import (
	"fmt"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/common"
	"github.com/stretchr/objx"
	"net/http"
	"strings"
)

type authHandler struct {
	next http.Handler
}

func MustAuth(handler http.Handler) http.Handler {
	return &authHandler{next: handler}
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("auth")
	// ? error handling
	if err == http.ErrNoCookie || ((cookie != nil) && (cookie.Value == "")) {
		// no authenticated
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	if err != nil {
		// some other error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// success - call the next handler
	h.next.ServeHTTP(w, r)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	segs := strings.Split(r.URL.Path, "/")
	action := segs[2]
	provider := segs[3]
	switch action {
	case "login":
		provider, err := gomniauth.Provider(provider)
		if err != nil {
			getProviderErrorHelper(w, provider, err)
			return
		}
		loginUrl, err := provider.GetBeginAuthURL(nil, nil)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("Error when trying to GetBeginAuthURL for %s: %s", provider, err),
				http.StatusInternalServerError,
			)
			return
		}
		w.Header().Set("Location", loginUrl)
		w.WriteHeader(http.StatusTemporaryRedirect)
	case "callback":
		provider, err := gomniauth.Provider(provider)
		if err != nil {
			getProviderErrorHelper(w, provider, err)
			return
		}
		creds, err := provider.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("Error when trying to complete auth for %s: %s", provider, err),
				http.StatusInternalServerError,
			)
			return
		}
		user, err := provider.GetUser(creds)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("Error when triyng to get user from %s: %s", provider, err),
				http.StatusInternalServerError,
			)
			return
		}
		authCookieValue := objx.New(map[string]interface{}{
			"name":       user.Nickname(),
			"avatar_url": user.AvatarURL(),
			"email":      user.Email(),
		}).MustBase64()
		http.SetCookie(w, &http.Cookie{
			Name:  "auth",
			Value: authCookieValue,
			Path:  "/",
		})
		w.Header().Set("Location", "/chat")
		w.WriteHeader(http.StatusTemporaryRedirect)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Auth action %s not supported", action)
	}
}

func getProviderErrorHelper(w http.ResponseWriter, provider common.Provider, err error) {
	http.Error(
		w,
		fmt.Sprintf("Error when trying to get provider %s: %s", provider, err),
		http.StatusBadRequest,
	)
}

package handlers

import "net/http"

type authHandler struct {
	* ChainHandler
}


func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, err := r.Cookie("auth"); err == http.ErrNoCookie {
		//not authenticated
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else if err != nil {
		panic(err.Error())
	} else {
		//success - call next Handler
		h.Next().ServeHTTP(w, r)
	}
}

func MustAuth(handler http.Handler) http.Handler{
	return &authHandler{NewChainedHandler(handler)}
}
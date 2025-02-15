package source

import (
	"fmt"
	"net/http"
	"net/url"
)

const (
	MarkCookieUser = "tokenU"
	MarkCookieID   = "tokenID"
)

func ReadCookie(r *http.Request, mark string) (string, error) {
	if len(mark) < 1 {
		return "", fmt.Errorf("empty cookie")
	}

	cookie, errC := r.Cookie(mark)
	if errC != nil {
		return "", errC
	}

	value, errV := url.QueryUnescape(cookie.Value)
	if errV != nil {
		return "", errV
	}

	return value, nil
}

func CleanCookie(w http.ResponseWriter, r *http.Request) {
	for _, v := range r.Cookies() {
		c := http.Cookie{
			Name:   v.Name,
			MaxAge: -1,
		}
		http.SetCookie(w, &c)
	}
}

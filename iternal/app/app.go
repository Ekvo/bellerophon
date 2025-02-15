package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Ekvo/bellerophon/iternal/source"
)

type Application struct {
	source *source.SqlSource
	cashe  map[string]time.Time
}

func NewApplication(s *source.SqlSource) *Application {
	return &Application{
		source: s,
		cashe:  make(map[string]time.Time),
	}
}

const (
	pathLogin  = "/bellerophon/login"
	pathLogout = "/bellerophon/logout"
	pathSignUp = "/bellerophon/signup"
	pathMain   = "/bellerophon/my/main"
	pathUserID = "/bellerophon/ownid"
)

func (a Application) Routes(r *mux.Router) {
	r.HandleFunc(pathSignUp, a.SignUp).Methods("POST")
	r.HandleFunc(pathLogin, a.LogIn).Methods("GET", "POST")
	r.HandleFunc(pathLogout, a.LogOut).Methods("GET")

	r.HandleFunc(pathMain, a.authorization(a.Main)).Methods("GET", "PUT")
	r.HandleFunc(pathUserID, a.authorization(a.OwnID)).Methods("GET", "PUT")
}

const livingTime = 60 * time.Minute

func (a Application) LogIn(w http.ResponseWriter, r *http.Request) {
	log.Printf("handle task: LogIn on url:%s", r.URL.Path)

	if r.Method == http.MethodGet {
		msg := source.Message{Msg: "Enter login and Password"}
		_ = encode(w, &msg, http.StatusOK)

		return
	}

	if r.Method == http.MethodPost {
		var u source.UserSourceData
		httpStatus, errDec := decode(r, &u)
		if errDec != nil {
			http.Error(w, errDec.Error(), httpStatus)

			return
		}

		if u.Direct != source.UserConnect {
			http.Error(w, source.IncorrectDirectUserStruct.Error(), http.StatusBadRequest)

			return
		}

		if errHash := u.HashPassword(); errHash != nil {
			http.Error(w, errHash.Error(), http.StatusBadRequest)

			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		user, errUser := a.source.UserLogin(ctx, &u)
		if errUser != nil {
			http.Error(w, errUser.Error(), http.StatusInternalServerError)

			return
		}

		token := u.Login + u.PasswordOne
		tokenHash := source.HashData(token)

		startTime := time.Now()
		exploration := startTime.Add(livingTime)

		a.cashe[tokenHash] = exploration

		cookieU := http.Cookie{
			Name:    source.MarkCookieUser,
			Value:   url.QueryEscape(tokenHash),
			Expires: exploration,
		}
		http.SetCookie(w, &cookieU)

		cookieID := http.Cookie{
			Name:    source.MarkCookieID,
			Value:   url.QueryEscape(strconv.Itoa(user.ID)),
			Expires: exploration,
		}
		http.SetCookie(w, &cookieID)

		http.Redirect(w, r, pathMain, http.StatusSeeOther)

		return
	}

	http.Error(w, fmt.Sprintf("unexepted Metod - %s on url - %s", r.Method, r.URL.Path), http.StatusMethodNotAllowed)
}

func (a Application) LogOut(w http.ResponseWriter, r *http.Request) {
	log.Printf("handle tsk: LogOut ou url:%s", r.URL.Path)

	tokenU, _ := source.ReadCookie(r, source.MarkCookieUser)
	if len(tokenU) > 0 {
		if _, ex := a.cashe[tokenU]; !ex {
			http.Error(w, fmt.Sprintf("cashe - empty, cookie - not empty:%s", tokenU), http.StatusInternalServerError)

			return
		}

		delete(a.cashe, tokenU)
	}

	source.CleanCookie(w, r)
	http.Redirect(w, r, pathLogin, http.StatusSeeOther)
}

func (a Application) SignUp(w http.ResponseWriter, r *http.Request) {
	log.Printf("handle task: SignUp on url:%s", r.URL.Path)

	var u source.UserSourceData
	httpStatus, errDec := decode(r, &u)
	if errDec != nil {
		http.Error(w, errDec.Error(), httpStatus)

		return
	}

	if u.Direct != source.UserCreate {
		http.Error(w, source.IncorrectDirectUserStruct.Error(), http.StatusBadRequest)

		return
	}

	if errHash := u.HashPassword(); errHash != nil {
		http.Error(w, errHash.Error(), http.StatusBadRequest)

		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	id, errDBUser := a.source.UserCreate(ctx, &u)
	if errDBUser != nil {
		http.Error(w, errDBUser.Error(), http.StatusInternalServerError)

		return
	}

	msg := source.Message{Msg: fmt.Sprintf("new user ID=%d", id)}
	_ = encode(w, &msg, http.StatusCreated)
}

func (a Application) Main(w http.ResponseWriter, r *http.Request) {
	log.Printf("handle task: Main on url:%s with Metod:%s", r.URL.Path, r.Method)

	tokenID, errT := source.ReadCookie(r, source.MarkCookieID)
	if errT != nil {
		http.Error(w, errT.Error(), http.StatusInternalServerError)

		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 300*time.Second)
	defer cancel()

	if r.Method == http.MethodGet {
		secret, errDB := a.source.InfoByID(ctx, tokenID)
		if errDB != nil {
			http.Error(w, errDB.Error(), http.StatusNoContent)

			return
		}

		msg := &source.Message{Msg: secret}
		_ = encode(w, &msg, http.StatusOK)

		return
	}

	if r.Method == http.MethodPut {
		var secret source.Message
		httpStatus, errDec := decode(r, &secret)
		if errDec != nil {
			http.Error(w, errDec.Error(), httpStatus)

			return
		}

		errDB := a.source.InfoChangeByID(ctx, tokenID, secret.Msg)
		if errDB != nil {
			http.Error(w, errDB.Error(), http.StatusInternalServerError)

			return
		}

		msg := source.Message{Msg: fmt.Sprint("upload secret")}
		_ = encode(w, &msg, http.StatusCreated)

		return
	}

	http.Error(w, fmt.Sprintf("unexepted Metod - %s on url - %s", r.Method, r.URL.Path), http.StatusMethodNotAllowed)
}

func (a Application) OwnID(w http.ResponseWriter, r *http.Request) {
	log.Printf("handle task: OwnID on url:%s with Metod^%s", r.URL.Path, r.Method)

	tokenID, errT := source.ReadCookie(r, source.MarkCookieID)
	if errT != nil {
		http.Error(w, errT.Error(), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	if r.Method == http.MethodGet {
		user, errUser := a.source.UserData(ctx, tokenID)
		if errUser != nil {
			http.Error(w, errUser.Error(), http.StatusInternalServerError)

			return
		}

		_ = encode(w, &user, http.StatusOK)

		return
	}

	if r.Method == http.MethodPut {
		var u source.UserSourceData
		httpStatus, errDec := decode(r, &u)
		if errDec != nil {
			http.Error(w, errDec.Error(), httpStatus)

			return
		}

		id, errID := strconv.Atoi(tokenID)
		if errID != nil {
			http.Error(w, errID.Error(), http.StatusInternalServerError)

			return
		}

		u.ID = id
		m := source.Message{}
		httpStatus = http.StatusCreated

		switch u.Direct {

		case source.NewLogin:
			if len(u.Login) < 1 {
				http.Error(w, fmt.Sprint("empty login"), http.StatusBadRequest)

				return
			}

			errUpdate := a.source.UserDataLoginUpdate(ctx, &u)
			if errUpdate != nil {
				http.Error(w, errUpdate.Error(), http.StatusInternalServerError)

				return
			}

			m.Msg = fmt.Sprintf("user login with id=%d updated", u.ID)

		case source.NewPassword:
			if len(u.PasswordOne) < 1 {
				http.Error(w, fmt.Sprint("empty password"), http.StatusBadRequest)

				return
			}
			if u.PasswordOne != u.PasswordTwo {
				http.Error(w, fmt.Sprint("passwords not rqual"), http.StatusBadRequest)

				return
			}

			if errStatusHash := u.HashPassword(); errStatusHash != nil {
				http.Error(w, errStatusHash.Error(), http.StatusBadRequest)

				return
			}

			errUpdate := a.source.UserDataPasswordUpdate(ctx, &u)
			if errUpdate != nil {
				http.Error(w, errUpdate.Error(), http.StatusInternalServerError)

				return
			}

			m.Msg = fmt.Sprintf("user password with id=%d updated", u.ID)

		case source.NewName:
			if len(u.Name) < 1 {
				http.Error(w, fmt.Sprint("empty name"), http.StatusBadRequest)

				return
			}

			errUpdate := a.source.UserDataNameUpdate(ctx, &u)
			if errUpdate != nil {
				http.Error(w, errUpdate.Error(), http.StatusInternalServerError)

				return
			}

			m.Msg = fmt.Sprintf("user Name and Surname with id=%d updated", u.ID)

		case source.NewEmail:
			if len(u.Email) < 1 {
				http.Error(w, fmt.Sprint("empty emal"), http.StatusBadRequest)

				return
			}

			errUpdate := a.source.UserDataEmailUpdate(ctx, &u)
			if errUpdate != nil {
				http.Error(w, errUpdate.Error(), http.StatusInternalServerError)

				return
			}

			m.Msg = fmt.Sprintf("user email with id=%d updated", u.ID)

		case source.UserDelete:
			errDelete := a.source.UserDataDelete(ctx, strconv.Itoa(u.ID))
			if errDelete != nil {
				http.Error(w, errDelete.Error(), http.StatusInternalServerError)

				return
			}

			m.Msg = fmt.Sprintf("user with id=%d deleted", u.ID)
			httpStatus = http.StatusOK

		default:
			http.Error(w, fmt.Sprintf("unsupported direction=%d", u.Direct), http.StatusBadRequest)

			return
		}

		tokenU, _ := source.ReadCookie(r, source.MarkCookieUser)
		delete(a.cashe, tokenU)
		source.CleanCookie(w, r)

		_ = encode(w, &m, httpStatus)

		return
	}

	http.Error(w, fmt.Sprintf("unexepted Metod - %s on url - %s", r.Method, r.URL.Path), http.StatusMethodNotAllowed)
}

func (a Application) authorization(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("handle task: authorization on url:%s with Metod:%s", r.URL.Path, r.Method)

		tokenU, err := source.ReadCookie(r, source.MarkCookieUser)
		if err != nil {
			http.Redirect(w, r, pathLogin, http.StatusSeeOther)

			return
		}

		if life, ex := a.cashe[tokenU]; !ex || life.Before(time.Now()) {
			delete(a.cashe, tokenU)
			http.Redirect(w, r, pathLogin, http.StatusSeeOther)

			return
		}

		next(w, r)
	}
}

func decode(r *http.Request, obj any) (int, error) {
	media := r.Header.Get("Content-Type")

	parse, _, errMed := mime.ParseMediaType(media)
	if errMed != nil {
		return http.StatusBadRequest, errMed
	}
	if parse != "application/json" {
		return http.StatusUnsupportedMediaType, fmt.Errorf("need - 'app/json'; get - %s", parse)
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	errDec := dec.Decode(obj)
	if errDec != nil {
		return http.StatusBadRequest, errDec
	}

	return http.StatusCreated, nil
}

func encode(w http.ResponseWriter, obj any, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errWrite := json.NewEncoder(w).Encode(obj)
	if errWrite != nil {
		log.Printf("json.Encode error - %v", errWrite)

		return errWrite
	}

	return nil
}

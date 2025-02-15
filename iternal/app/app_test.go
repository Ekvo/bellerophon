package app

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Ekvo/bellerophon/iternal/connect"
	"github.com/Ekvo/bellerophon/iternal/source"
)

var (
	db     *sql.DB           = nil
	s      *source.SqlSource = nil
	a      *Application      = nil
	r      *mux.Router       = nil
	srv    *http.Server      = nil
	client *http.Client      = nil
	jar    *cookiejar.Jar    = nil
)

func startBaseAndServAndClient() error {
	conn, errCon := connect.NewConnect("../connect/connectData.json")
	if errCon != nil {
		return errCon
	}

	var errDB error = nil

	db, errDB = sql.Open("postgres", conn.String())
	if errDB != nil {
		return errDB
	}

	s = source.NewSqlSource(db)
	a = NewApplication(s)
	r = mux.NewRouter()

	a.Routes(r)

	srv = &http.Server{
		Addr:         "127.0.0.1:8000",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	var errJar error = nil

	jar, errJar = cookiejar.New(nil)
	if errJar != nil {
		return errJar
	}

	client = &http.Client{
		Jar: jar,
	}

	return nil
}

func newUser(direct int) *source.UserSourceData {
	hashed := source.HashData("qwert1234")

	return &source.UserSourceData{
		Direct: direct,
		ID:     0,
		ChangeLogin: source.ChangeLogin{
			Login: "Loko",
		},
		ChangePassword: source.ChangePassword{
			Hashed:      source.Hashed,
			PasswordOne: hashed,
			PasswordTwo: hashed,
		},
		ChangeName: source.ChangeName{
			Name:    "Pavel",
			Surname: "",
		},
		ChangeEmail: source.ChangeEmail{
			Email: "genus1991@gmail.com",
		},
	}
}

func getNumberFromBody(data string) (string, error) {
	arr := []byte{}

	for _, ch := range data {
		if 47 < ch && ch < 58 {
			arr = append(arr, byte(ch))
			continue
		}

		if len(arr) != 0 {
			break
		}
	}
	if len(arr) == 0 {
		return "", fmt.Errorf("unexpected res.Body")
	}

	return string(arr), nil
}

func TestSignUpApprove(t *testing.T) {
	errStart := startBaseAndServAndClient()
	require.NoError(t, errStart)
	defer db.Close()

	go func() {
		if errListen := srv.ListenAndServe(); errListen != nil {
			assert.ErrorIs(t, errListen, http.ErrServerClosed)
		}
	}()
	defer func() {
		if err := srv.Close(); err != nil {
			assert.NoError(t, err)
		}
	}()

	time.Sleep(1 * time.Second)

	user := newUser(source.UserCreate)

	data, errMar := json.Marshal(&user)
	require.NoError(t, errMar)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	req, errReq := http.NewRequestWithContext(ctx, http.MethodPost, "http://127.0.0.1:8000"+pathSignUp, bytes.NewReader(data))
	require.NoError(t, errReq)
	req.Header.Set("Content-Type", "application/json")

	res, errRes := client.Do(req)
	require.NoError(t, errRes)
	defer res.Body.Close()

	require.Equal(t, http.StatusCreated, res.StatusCode)
	require.NotEmpty(t, res.Body)

	arrData, errData := io.ReadAll(res.Body)
	require.NoError(t, errData)

	need := `{"message":"new user ID=`
	get := strings.TrimSpace(string(arrData))

	require.Greater(t, len(get), len(need))
	assert.Equal(t, need, get[:len(need)])

	strID, errNum := getNumberFromBody(get)
	require.NoError(t, errNum)

	_ = s.UserDataDelete(ctx, strID)

}

func TestLoginAndAutorizationAndMainApprove(t *testing.T) {
	errStart := startBaseAndServAndClient()
	require.NoError(t, errStart)
	defer db.Close()

	go func() {
		if errListen := srv.ListenAndServe(); errListen != nil {
			assert.ErrorIs(t, errListen, http.ErrServerClosed)
		}
	}()
	defer func() {
		if err := srv.Close(); err != nil {
			assert.NoError(t, err)
		}
	}()

	time.Sleep(1 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	user := newUser(source.UserCreate)

	id, errCreate := s.UserCreate(ctx, user)
	require.NoError(t, errCreate)

	strID := strconv.Itoa(id)

	errSecret := s.InfoChangeByID(ctx, strID, "new secret Loko")
	require.NoError(t, errSecret)

	userLogin := newUser(source.UserConnect)

	data, errMar := json.Marshal(userLogin)
	require.NoError(t, errMar)

	req, errReq := http.NewRequestWithContext(ctx, http.MethodPost, "http://127.0.0.1:8000"+pathLogin, bytes.NewReader(data))
	require.NoError(t, errReq)
	req.Header.Set("Content-Type", "application/json")

	res, errRes := client.Do(req)
	require.NoError(t, errRes)
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.NotEmpty(t, res.Body)

	arrData, errRead := io.ReadAll(res.Body)
	require.NoError(t, errRead)

	wont := `{"message":"new secret Loko"}`
	get := strings.TrimSpace(string(arrData))

	assert.Equal(t, wont, get)

	_ = s.UserDataDelete(ctx, strID)
}

func TestLoginAndOwnIDAndNewLogin(t *testing.T) {
	errStart := startBaseAndServAndClient()
	require.NoError(t, errStart)
	defer db.Close()

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			require.ErrorIs(t, err, http.ErrServerClosed)
		}
	}()
	defer func() {
		if err := srv.Close(); err != nil {
			assert.NoError(t, err)
		}
	}()

	time.Sleep(1 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	user := newUser(source.UserCreate)

	id, errCreate := s.UserCreate(ctx, user)
	require.NoError(t, errCreate)

	strID := strconv.Itoa(id)

	userNewLogin := newUser(source.NewLogin)
	userNewLogin.Login = "L"

	data, errMar := json.Marshal(userNewLogin)
	require.NoError(t, errMar)

	const tokU = "a1271fc76143dbce8cccce94e3ac04b55eb381fcacc377d355fd10a87ace401e"

	req, errReq := http.NewRequestWithContext(ctx, http.MethodPut, "http://127.0.0.1:8000"+pathUserID, bytes.NewReader(data))
	require.NoError(t, errReq)
	req.Header.Set("Cookie", fmt.Sprintf("tokenU=%s; tokenID=%s", tokU, strID))
	req.Header.Set("Content-Type", "application/json")
	a.cashe[tokU] = time.Now().Add(livingTime)

	res, errRes := client.Do(req)
	require.NoError(t, errRes)
	defer res.Body.Close()

	require.Equal(t, http.StatusCreated, res.StatusCode)

	arrData, errData := io.ReadAll(res.Body)
	require.NoError(t, errData)

	wont := fmt.Sprintf(`{"message":"user login with id=%s updated"}`, strID)
	get := strings.TrimSpace(string(arrData))

	assert.Equal(t, wont, get)

	life, ex := a.cashe[tokU]

	assert.False(t, ex)
	assert.Empty(t, life)

	cookies := res.Cookies()
	assert.NotNil(t, cookies)

	for _, c := range cookies {
		assert.Empty(t, c.Value)
	}

	_ = s.UserDataDelete(ctx, strID)
}

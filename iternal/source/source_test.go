package source

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"

	"github.com/Ekvo/bellerophon/iternal/connect"
)

var (
	db    *sql.DB    = nil
	store *SqlSource = nil
)

func startBase() error {
	coon, errCon := connect.NewConnect("../connect/connectData.json")
	if errCon != nil {
		return errCon
	}

	var errDB error = nil

	db, errDB = sql.Open("postgres", coon.String())
	if errDB != nil {
		return errDB
	}

	store = NewSqlSource(db)

	return nil
}

func NewUser() *UserSourceData {
	hashedPas := HashData("qwert1234")

	return &UserSourceData{
		Direct: 1,
		ID:     0,
		ChangeLogin: ChangeLogin{
			Login: "Loko",
		},
		ChangePassword: ChangePassword{
			Hashed:      Hashed,
			PasswordOne: hashedPas,
			PasswordTwo: hashedPas,
		},
		ChangeName: ChangeName{
			Name:    "Pavel",
			Surname: "",
		},
		ChangeEmail: ChangeEmail{
			Email: "genus1991@gmail.com",
		},
	}
}

func TestUserCreateGetDelete(t *testing.T) {
	errStart := startBase()
	require.NoError(t, errStart)
	defer db.Close()

	ctx := context.Background()

	userSign := NewUser()

	id, errCreate := store.UserCreate(ctx, userSign)
	require.NoError(t, errCreate)
	require.NotZero(t, id)

	strID := strconv.Itoa(id)

	u, errData := store.UserData(ctx, strID)
	require.NoError(t, errData)
	require.NotEmpty(t, u)

	assert.Empty(t, u.HashPassword)

	assert.Equal(t, id, u.ID)
	assert.Equal(t, userSign.Login, u.Login)
	assert.Equal(t, userSign.Name, u.Name)
	assert.Equal(t, userSign.Surname, u.Surname)
	assert.Equal(t, userSign.Email, u.Email)

	errDelete := store.UserDataDelete(ctx, strID)

	require.NoError(t, errDelete)

	u, errData = store.UserData(ctx, strID)
	assert.ErrorIs(t, errData, sql.ErrNoRows)

	assert.Empty(t, u)
}

func newChangeUser(id int) *UserSourceData {
	hashedPas := HashData("qwert4321")

	return &UserSourceData{
		Direct: 0,
		ID:     id,
		ChangeLogin: ChangeLogin{
			Login: "L",
		},
		ChangePassword: ChangePassword{
			Hashed:      Hashed,
			PasswordOne: hashedPas,
			PasswordTwo: hashedPas,
		},
		ChangeName: ChangeName{
			Name:    "Egor",
			Surname: "Morozov",
		},
		ChangeEmail: ChangeEmail{
			Email: "genus1991@yandex.ru",
		},
	}
}

func TestUserNewLoginPasswordNameEmail(t *testing.T) {
	errStart := startBase()
	require.NoError(t, errStart)
	defer db.Close()

	ctx := context.Background()

	userSign := NewUser()

	id, errCreate := store.UserCreate(ctx, userSign)
	require.NoError(t, errCreate)

	userChange := newChangeUser(id)

	userChange.Direct = NewLogin
	errNewLog := store.UserDataLoginUpdate(ctx, userChange)
	assert.NoError(t, errNewLog)

	userChange.Direct = NewPassword
	errNewPas := store.UserDataPasswordUpdate(ctx, userChange)
	assert.NoError(t, errNewPas)

	userChange.Direct = NewName
	errNewName := store.UserDataNameUpdate(ctx, userChange)
	assert.NoError(t, errNewName)

	userChange.Direct = NewEmail
	errNewEmail := store.UserDataEmailUpdate(ctx, userChange)
	assert.NoError(t, errNewEmail)

	strID := strconv.Itoa(id)

	u, errData := store.UserData(ctx, strID)
	require.NoError(t, errData)

	assert.Equal(t, u.ID, id)
	assert.Equal(t, u.Login, userChange.Login)
	assert.Equal(t, u.Name, userChange.Name)
	assert.Equal(t, u.Surname, userChange.Surname)
	assert.Equal(t, u.Email, userChange.Email)

	errDel := store.UserDataDelete(ctx, strID)
	require.NoError(t, errDel)
}

func TestGetInfoNewInfo(t *testing.T) {
	errStart := startBase()
	require.NoError(t, errStart)
	defer db.Close()

	ctx := context.Background()

	userSign := NewUser()

	id, errCreate := store.UserCreate(ctx, userSign)
	require.NoError(t, errCreate)

	strID := strconv.Itoa(id)

	secret, errInfo := store.InfoByID(ctx, strID)
	require.Error(t, errInfo)
	assert.Empty(t, secret)

	newSecret := "so big secret"

	errNewInfo := store.InfoChangeByID(ctx, strID, newSecret)
	require.NoError(t, errNewInfo)

	secret, errInfo = store.InfoByID(ctx, strID)
	require.NoError(t, errInfo)
	assert.Equal(t, secret, newSecret)

	errDel := store.UserDataDelete(ctx, strID)
	require.NoError(t, errDel)
}

package source

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type SqlSource struct {
	source *sql.DB
}

func NewSqlSource(source *sql.DB) *SqlSource {
	return &SqlSource{source: source}
}

func (s *SqlSource) UserCreate(ctx context.Context, u *UserSourceData) (int, error) {
	row := s.source.QueryRowContext(ctx, `
WITH ins_1 AS (
	INSERT INTO users (login,
    						 hashed_password,
    						 name,
   				   	 surname,
    						 email)
	VALUES ($1,$2,$3,$4,$5)
	RETURNING  id)
INSERT INTO info (id) 
SELECT id FROM ins_1
RETURNING  id;`,
		u.Login, u.PasswordOne, u.Name, u.Surname, u.Email)

	id := 0
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}
	// return ID
	return id, nil
}

func (s *SqlSource) UserLogin(ctx context.Context, u *UserSourceData) (User, error) {
	row := s.source.QueryRowContext(ctx, `
SELECT id,
       login,
       name,
       surname
FROM users
WHERE login = $1
		AND hashed_password = $2;`, u.Login, u.PasswordOne)

	var user User
	err := row.Scan(&user.ID, &user.Login, &user.Name, &user.Surname)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (s *SqlSource) UserData(ctx context.Context, id string) (User, error) {
	row := s.source.QueryRowContext(ctx, `
SELECT id,
       login,
       name,
       surname,
		 email
FROM users
WHERE id = $1;`, id)

	var u User
	err := row.Scan(&u.ID, &u.Login, &u.Name, &u.Surname, &u.Email)
	if err != nil {
		return User{}, err
	}

	return u, nil
}

func (s *SqlSource) UserDataLoginUpdate(ctx context.Context, u *UserSourceData) error {
	_, err := s.source.ExecContext(ctx, `
UPDATE users
SET login=$1
WHERE id = $2;`, u.Login, u.ID)

	return err
}

func (s *SqlSource) UserDataPasswordUpdate(ctx context.Context, u *UserSourceData) error {
	_, err := s.source.ExecContext(ctx, `
UPDATE users
SET hashed_password=$2
WHERE id = $1;`, u.ID, u.PasswordOne)

	return err
}

func (s *SqlSource) UserDataNameUpdate(ctx context.Context, u *UserSourceData) error {
	_, err := s.source.ExecContext(ctx, `
UPDATE users
SET  name=$1,
	  surname=$2
WHERE id = $3;`, u.Name, u.Surname, u.ID)

	return err
}

func (s *SqlSource) UserDataEmailUpdate(ctx context.Context, u *UserSourceData) error {
	_, err := s.source.ExecContext(ctx, `
UPDATE users
SET email=$1
WHERE id = $2;`, u.Email, u.ID)

	return err
}

func (s *SqlSource) UserDataDelete(ctx context.Context, id string) error {
	tx, err := s.source.Begin()
	if err != nil {
		return err
	}

	var errBack error = nil

	_, err = s.source.ExecContext(ctx, `DELETE FROM info WHERE id = $1;`, id)
	if err != nil {
		errBack = tx.Rollback()
		if errBack != nil {
			return fmt.Errorf("UserDataDelete first error - %w, second error -%w", err, errBack)
		}

		return err
	}

	_, err = s.source.ExecContext(ctx, `DELETE FROM users WHERE id = $1;`, id)
	if err != nil {
		errBack = tx.Rollback()
		if errBack != nil {
			return fmt.Errorf("UserDataDelete first error - %w, second error -%w", err, errBack)
		}

		return err
	}

	return tx.Commit()
}

func (s *SqlSource) InfoCreate(ctx context.Context, id string) error {
	_, err := s.source.ExecContext(ctx, `
INSERT INTO info (id) VALUES ($1)`, id)

	return err
}

func (s *SqlSource) InfoByID(ctx context.Context, id string) (string, error) {
	row := s.source.QueryRowContext(ctx, `
SELECT secret
FROM info
WHERE id = $1;`, id)

	secret := ""
	err := row.Scan(&secret)
	if err != nil {
		return "", err
	}

	return secret, nil
}

func (s *SqlSource) InfoChangeByID(ctx context.Context, id, secret string) error {
	_, err := s.source.ExecContext(ctx, `
UPDATE info
SET secret = $1
WHERE id = $2;`, secret, id)

	return err
}

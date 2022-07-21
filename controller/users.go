package controller

import (
	"repgen/core"
	"time"
)

type User struct {
	Id       int
	Email    string
	Password string
	Name     string
	Created  time.Time
}

const (
	UserEmailMaxLength    = 100
	UserPasswordMaxLength = 20
	UserNameMaxLength     = 100
)

func CreateUser(user *User) error {
	rows, err := core.Database.Query("INSERT INTO users (email, password, name, created) VALUES($1, $2, $3, $4) RETURNING id",
		user.Email, user.Password, user.Name, user.Created)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&user.Id)
		if err != nil {
			return err
		}

	}
	return nil
}

func UpdateUser(user User) (int64, error) {
	result, err := core.Database.Exec("UPDATE users SET name = $1 WHERE id = $2", user.Name, user.Id)
	if err != nil {
		return 0, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rows, nil
}

func UpdateUserPassword(user User) (int64, error) {
	result, err := core.Database.Exec("UPDATE users SET password = $1 WHERE id = $2", user.Password, user.Id)
	if err != nil {
		return 0, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rows, nil
}

func GetUserByEmail(email string) (*User, error) {
	rows, err := core.Database.Query("SELECT id, email, password, name, created FROM users WHERE email = $1", email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var user *User
	for rows.Next() {
		user = &User{}
		err := rows.Scan(&user.Id, &user.Email, &user.Password, &user.Name, &user.Created)
		if err != nil {
			return nil, err
		}
	}
	return user, nil
}

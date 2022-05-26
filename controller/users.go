package controller

import (
	"repgen/core"
	"time"
)

type Users struct {
	Id       int
	Email    string
	Password string
	Name     string
	Created  time.Time
}

func CreateUsers(user *Users) error {
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
	// commit?
	return nil
}

func GetUsersByEmail(email string) (*Users, error) {
	rows, err := core.Database.Query("SELECT id, email, password, name, created FROM users WHERE email = $1", email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var user *Users
	for rows.Next() {
		user = &Users{}
		err := rows.Scan(&user.Id, &user.Email, &user.Password, &user.Name, &user.Created)
		if err != nil {
			return nil, err
		}
	}
	return user, nil
}

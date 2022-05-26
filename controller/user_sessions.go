package controller

import (
	"repgen/core"
	"time"
)

type UserSession struct {
	Id      int
	UserId  int
	Session string
	Created time.Time
}

func CreateUserSession(userSession UserSession) error {
	rows, err := core.Database.Query("INSERT INTO user_session (user_id, session, created) VALUES($1, $2, $3)",
		userSession.UserId, userSession.Session, userSession.Created)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}

func DeleteUserSession(id int) error {
	rows, err := core.Database.Query("DELETE FROM user_session WHERE id = $1", id)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}

func GetUserSession(session string) (*UserSession, error) {
	rows, err := core.Database.Query("SELECT id, user_id, session, created FROM user_session WHERE session = $1", session)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userSession *UserSession
	for rows.Next() {
		userSession = &UserSession{}
		err := rows.Scan(&userSession.Id, &userSession.UserId, &userSession.Session, &userSession.Created)
		if err != nil {
			return nil, err
		}
	}
	return userSession, nil
}

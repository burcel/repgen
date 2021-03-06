package controller

import (
	"repgen/core"
	"time"
)

type Project struct {
	Id            int       `json:"id"`
	Name          string    `json:"name"`
	Created       time.Time `json:"created"`
	CreatedUserId int       `json:"-"`
}

const (
	ProjectNameMaxLength = 200
	ProjectPageLimit     = 10
)

func CreateProject(project *Project) error {
	rows, err := core.Database.Query("INSERT INTO project (name, created, created_user_id) VALUES($1, $2, $3) RETURNING id",
		project.Name, project.Created, project.CreatedUserId)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&project.Id)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateProject(project *Project) (int64, error) {
	result, err := core.Database.Exec("UPDATE project SET name = $1 WHERE id = $2", project.Name, project.Id)
	if err != nil {
		return 0, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rows, nil
}

func SelectProject(page int) ([]Project, error) {
	rows, err := core.Database.Query("SELECT id, name, created, created_user_id FROM project ORDER BY id ASC LIMIT $1 OFFSET $2 ",
		ProjectPageLimit, ProjectPageLimit*page)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	projects := []Project{}
	for rows.Next() {
		var project Project
		err := rows.Scan(&project.Id, &project.Name, &project.Created, &project.CreatedUserId)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, nil
}

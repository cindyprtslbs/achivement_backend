package models

import "time"

type Lecturer struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	LecturerID string     `json:"lecturer_id"`
	Department string     `json:"department"`
	CreatedAt  time.Time  `json:"created_at"`
}

type CreateLecturerRequest struct {
	UserID     string `json:"user_id"`
	LecturerID string `json:"lecturer_id"`
	Department string `json:"department"`
}

type UpdateLecturerRequest struct {
	UserID     string `json:"user_id"`
	LecturerID string `json:"lecturer_id"`
	Department string `json:"department"`
}

type SetLecturerProfileRequest struct {
	LecturerID string `json:"lecturer_id"`
	Department string `json:"department"`
}
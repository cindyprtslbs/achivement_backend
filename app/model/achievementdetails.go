package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ===============================================================
// ACHIEVEMENT DETAILS (Flexible untuk banyak jenis prestasi)
// ===============================================================
type AchievementDetails struct {
	CompetitionName  *string    `bson:"competitionName,omitempty" json:"competition_name,omitempty"`
	CompetitionLevel *string    `bson:"competitionLevel,omitempty" json:"competition_level,omitempty"`
	Rank             *int       `bson:"rank,omitempty" json:"rank,omitempty"`
	MedalType        *string    `bson:"medal_type,omitempty" json:"medal_type,omitempty"`

	PublicationType  *string    `bson:"publicationType,omitempty" json:"publication_type,omitempty"`
	PublicationTitle *string    `bson:"publicationTitle,omitempty" json:"publication_title,omitempty"`
	Authors          []string   `bson:"authors,omitempty" json:"authors,omitempty"`
	Publisher        *string    `bson:"publisher,omitempty" json:"publisher,omitempty"`
	ISSN             *string    `bson:"issn,omitempty" json:"issn,omitempty"`

	OrganizationName *string    `bson:"organizationName,omitempty" json:"organization_name,omitempty"`
	Position         *string    `bson:"position,omitempty" json:"position,omitempty"`

	Period *struct {
		Start *time.Time `bson:"start,omitempty" json:"start,omitempty"`
		End   *time.Time `bson:"end,omitempty" json:"end,omitempty"`
	} `bson:"period,omitempty" json:"period,omitempty"`

	CertificationName   *string    `bson:"certificationName,omitempty" json:"certification_name,omitempty"`
	IssuedBy            *string    `bson:"issuedBy,omitempty" json:"issued_by,omitempty"`
	CertificationNumber *string    `bson:"certificationNumber,omitempty" json:"certification_number,omitempty"`
	ValidUntil          *time.Time `bson:"validUntil,omitempty" json:"valid_until,omitempty"`

	EventDate *time.Time `bson:"eventDate,omitempty" json:"event_date,omitempty"`
	Location  *string    `bson:"location,omitempty" json:"location,omitempty"`
	Organizer *string    `bson:"organizer,omitempty" json:"organizer,omitempty"`

	Score        *float64       `bson:"score,omitempty" json:"score,omitempty"`
	CustomFields map[string]any `bson:"customFields,omitempty" json:"custom_fields,omitempty"`
}

// ===============================================================
// FILE ATTACHMENT
// ===============================================================
type Attachment struct {
	FileName   string    `json:"file_name" bson:"file_name"`
	FileURL    string    `json:"file_url" bson:"file_url"`
	FileType   string    `json:"file_type" bson:"file_type"`
	UploadedAt time.Time `json:"uploaded_at" bson:"uploaded_at"`
}

// ===============================================================
// ACHIEVEMENT (MongoDB Document)
// ===============================================================
type Achievement struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	StudentID       string             `bson:"studentId" json:"student_id"`
	AchievementType string             `bson:"achievementType" json:"achievement_type"`

	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Details     AchievementDetails `bson:"details" json:"details"`

	Attachments []Attachment `bson:"attachments" json:"attachments"`
	Tags        []string     `bson:"tags" json:"tags"`
	Points      *float64     `bson:"points,omitempty" json:"points,omitempty"`

	Status    string `bson:"status" json:"status"`          // draft / deleted (FR-005)
	IsDeleted bool   `bson:"isDeleted" json:"is_deleted"`   // soft delete flag

	CreatedAt time.Time `bson:"createdAt" json:"created_at"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updated_at"`
}

// ===============================================================
// REQUEST: CREATE ACHIEVEMENT
// ===============================================================
type CreateAchievementRequest struct {
	StudentID       string             `json:"student_id"`
	AchievementType string             `json:"achievement_type"`
	Title           string             `json:"title"`
	Description     string             `json:"description"`
	Details         AchievementDetails `json:"details"`
	Attachments     []Attachment       `json:"attachments"`
	Tags            []string           `json:"tags"`
	Points          *float64           `json:"points,omitempty"`
}

// ===============================================================
// REQUEST: UPDATE ACHIEVEMENT (Hanya untuk DRAFT)
// ===============================================================
type UpdateAchievementRequest struct {
	AchievementType string             `json:"achievement_type"`
	Title           string             `json:"title"`
	Description     string             `json:"description"`
	Details         AchievementDetails `json:"details"`
	Attachments     []Attachment       `json:"attachments"`
	Tags            []string           `json:"tags"`
	Points          *float64           `json:"points,omitempty"`
}

// ===============================================================
// REQUEST: UPDATE ATTACHMENTS ONLY
// ===============================================================
type UpdateAchievementAttachmentsRequest struct {
	Attachments []Attachment `json:"attachments"`
}

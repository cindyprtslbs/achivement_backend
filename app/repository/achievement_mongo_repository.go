package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	models "achievement_backend/app/model"
)

// ================= INTERFACE =================

type MongoAchievementRepository interface {
	CreateDraft(ctx context.Context, req *models.CreateAchievementRequest) (*models.Achievement, error)
	GetByID(ctx context.Context, id string) (*models.Achievement, error)
	GetByStudentID(ctx context.Context, studentID string) ([]models.Achievement, error)

	UpdateDraft(ctx context.Context, id string, req *models.UpdateAchievementRequest) (*models.Achievement, error)
	UpdateAttachments(ctx context.Context, id string, attachments []models.Attachment) (*models.Achievement, error)

	SoftDelete(ctx context.Context, id string) error
}

// ================= STRUCT =================

type mongoAchievementRepository struct {
	collection *mongo.Collection
}

// ================= CONSTRUCTOR =================

func NewMongoAchievementRepository(db *mongo.Database) MongoAchievementRepository {
	return &mongoAchievementRepository{
		collection: db.Collection("achievements"),
	}
}

// ================= CREATE DRAFT (FR-001) =================

func (r *mongoAchievementRepository) CreateDraft(ctx context.Context, req *models.CreateAchievementRequest) (*models.Achievement, error) {
	achievement := models.Achievement{
		ID:              primitive.NewObjectID(),
		StudentID:       req.StudentID,
		AchievementType: req.AchievementType,
		Title:           req.Title,
		Description:     req.Description,
		Details:         req.Details,
		Attachments:     req.Attachments,
		Tags:            req.Tags,
		Points:          req.Points,

		Status:    "draft",
		IsDeleted: false,

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := r.collection.InsertOne(ctx, achievement)
	if err != nil {
		return nil, err
	}

	return &achievement, nil
}

// ================= GET BY ID =================

func (r *mongoAchievementRepository) GetByID(ctx context.Context, id string) (*models.Achievement, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var achievement models.Achievement
	err = r.collection.FindOne(ctx, bson.M{
		"_id":       objID,
		"isDeleted": false,
	}).Decode(&achievement)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &achievement, err
}

// ================= GET BY STUDENT ID =================

func (r *mongoAchievementRepository) GetByStudentID(ctx context.Context, studentID string) ([]models.Achievement, error) {
	filter := bson.M{
		"studentId": studentID,
		"isDeleted": false,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []models.Achievement
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// ================= UPDATE DRAFT (FR-002) =================

func (r *mongoAchievementRepository) UpdateDraft(ctx context.Context, id string, req *models.UpdateAchievementRequest) (*models.Achievement, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	// Pastikan status draft (SRS requirement)
	var existing models.Achievement
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&existing)
	if err != nil {
		return nil, err
	}

	if existing.Status != "draft" {
		return nil, errors.New("prestasi tidak dapat diubah karena bukan status draft")
	}

	update := bson.M{
		"$set": bson.M{
			"achievementType": req.AchievementType,
			"title":           req.Title,
			"description":     req.Description,
			"details":         req.Details,
			"attachments":     req.Attachments,
			"tags":            req.Tags,
			"points":          req.Points,
			"updatedAt":       time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

// ================= UPDATE ATTACHMENTS (FR-006) =================

func (r *mongoAchievementRepository) UpdateAttachments(ctx context.Context, id string, attachments []models.Attachment) (*models.Achievement, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var existing models.Achievement
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&existing)
	if err != nil {
		return nil, err
	}

	if existing.Status != "draft" {
		return nil, errors.New("files hanya dapat diubah pada status draft")
	}

	update := bson.M{
		"$set": bson.M{
			"attachments": attachments,
			"updatedAt":   time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

// ================= SOFT DELETE DRAFT (FR-005) =================

func (r *mongoAchievementRepository) SoftDelete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	var existing models.Achievement
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&existing)
	if err != nil {
		return err
	}

	if existing.Status != "draft" {
		return errors.New("prestasi hanya dapat dihapus jika masih draft")
	}

	update := bson.M{
		"$set": bson.M{
			"status":    "deleted",
			"isDeleted": true,
			"updatedAt": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

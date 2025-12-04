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
	GetAll(ctx context.Context) ([]models.Achievement, error)
	GetByAdvisor(ctx context.Context, studentIDs []string) ([]models.Achievement, error)

	CreateDraft(ctx context.Context, studentID string, req *models.CreateAchievementRequest) (*models.Achievement, error)
	GetByID(ctx context.Context, id string) (*models.Achievement, error)
	GetByStudentID(ctx context.Context, studentID string) ([]models.Achievement, error)

	UpdateDraft(ctx context.Context, id string, req *models.UpdateAchievementRequest) (*models.Achievement, error)
	UpdateAttachments(ctx context.Context, id string, attachments []models.Attachment) (*models.Achievement, error)

	SoftDelete(ctx context.Context, id string) error

	GetManyByIDs(ctx context.Context, ids []string) (map[string]models.Achievement, error)
	UpdateStatus(ctx context.Context, id string, status string) error
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

// ================= LIST ALL (Admin) =================

func (r *mongoAchievementRepository) GetAll(ctx context.Context) ([]models.Achievement, error) {
	filter := bson.M{"isDeleted": false}

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

// ================= LIST BY ADVISOR =================
// mahasiswa bimbingan (studentID array)

func (r *mongoAchievementRepository) GetByAdvisor(ctx context.Context, studentIDs []string) ([]models.Achievement, error) {
	filter := bson.M{
		"studentId": bson.M{"$in": studentIDs},
		"isDeleted": false,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.Achievement
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// ================= CREATE DRAFT =================
func (r *mongoAchievementRepository) CreateDraft(ctx context.Context, studentID string, req *models.CreateAchievementRequest,) (*models.Achievement, error) {

    achievement := models.Achievement{
        ID:              primitive.NewObjectID(),
        StudentID:       studentID, // <-- PENTING: bukan dari req
        AchievementType: req.AchievementType,
        Title:           req.Title,
        Description:     req.Description,
        Details:         req.Details,
        Attachments:     req.Attachments,
        Tags:            req.Tags,
        Points:          req.Points,
        Status:          models.StatusDraft,
        IsDeleted:       false,
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }

    _, err := r.collection.InsertOne(ctx, achievement)
    if err != nil {
        return nil, err
    }

    return &achievement, nil
}


// ================= GET BY ID =================

func (r *mongoAchievementRepository) GetByID(ctx context.Context, id string) (*models.Achievement, error) {
    var result models.Achievement

    // ======================================
    // 1. Coba sebagai ObjectID
    // ======================================
    if oid, err := primitive.ObjectIDFromHex(id); err == nil {
        err = r.collection.FindOne(ctx, bson.M{
            "_id":       oid,
            "isDeleted": false,
        }).Decode(&result)

        if err == nil {
            return &result, nil
        }

        // Kalau err == ErrNoDocuments → lanjut ke string mode
        if !errors.Is(err, mongo.ErrNoDocuments) {
            return nil, err
        }
    }

    // ======================================
    // 2. Fallback: cari _id STRING
    // ======================================
    err := r.collection.FindOne(ctx, bson.M{
        "_id":       id,
        "isDeleted": false,
    }).Decode(&result)

    if errors.Is(err, mongo.ErrNoDocuments) {
        return nil, nil
    }

    return &result, err
}


// ================= LIST BY STUDENT =================

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

// ================= UPDATE DRAFT =================

func (r *mongoAchievementRepository) UpdateDraft(ctx context.Context, id string, req *models.UpdateAchievementRequest) (*models.Achievement, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"_id":       objID,
		"status":    models.StatusDraft,
		"isDeleted": false,
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

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	if result.ModifiedCount == 0 {
		return nil, errors.New("prestasi hanya dapat diubah jika status masih draft")
	}

	return r.GetByID(ctx, id)
}

// ================= UPDATE ATTACHMENTS =================

func (r *mongoAchievementRepository) UpdateAttachments(ctx context.Context, id string, attachments []models.Attachment) (*models.Achievement, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var existing models.Achievement
	err = r.collection.FindOne(ctx, bson.M{
		"_id":       objID,
		"isDeleted": false,
	}).Decode(&existing)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.New("achievement not found")
	}
	if err != nil {
		return nil, err
	}

	// Draft only
	if existing.Status != models.StatusDraft {
		return nil, errors.New("attachments hanya dapat diubah jika masih draft")
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{
		"$set": bson.M{
			"attachments": attachments,
			"updatedAt":   time.Now(),
		},
	})
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

// ================= SOFT DELETE =================

func (r *mongoAchievementRepository) SoftDelete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	var existing models.Achievement
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&existing)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return errors.New("achievement not found")
	}
	if err != nil {
		return err
	}

	if existing.Status != models.StatusDraft {
		return errors.New("prestasi hanya dapat dihapus jika masih draft")
	}

	update := bson.M{
		"$set": bson.M{
			"status":    models.StatusDeleted,
			"isDeleted": true,
			"updatedAt": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

// ================= GET MANY BY IDS =================

func (r *mongoAchievementRepository) GetManyByIDs(ctx context.Context, ids []string) (map[string]models.Achievement, error) {
	objIDs := []primitive.ObjectID{}
	for _, id := range ids {
		objID, e := primitive.ObjectIDFromHex(id)
		if e == nil {
			objIDs = append(objIDs, objID)
		}
	}

	filter := bson.M{
		"_id":       bson.M{"$in": objIDs},
		"isDeleted": false,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string]models.Achievement)
	for cursor.Next(ctx) {
		var ach models.Achievement
		if err := cursor.Decode(&ach); err == nil {
			result[ach.ID.Hex()] = ach
		}
	}

	return result, nil
}

// ================= UPDATE STATUS =================

func (r *mongoAchievementRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
	}

	res, err := r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return errors.New("achievement not found")
	}

	return nil
}

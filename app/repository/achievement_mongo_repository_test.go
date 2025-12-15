package repository

import (
	"context"
	"testing"
	"time"

	models "achievement_backend/app/model"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	testDB   *mongo.Database
	testRepo MongoAchievementRepository
)

// =======================================================
// SETUP & TEARDOWN
// =======================================================

func setupMongoTest(t *testing.T) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(
		"mongodb://localhost:27017",
	))
	assert.NoError(t, err)

	testDB = client.Database("achievement_test_db")
	testRepo = NewMongoAchievementRepository(testDB)

	// clean collection
	_ = testDB.Collection("achievements").Drop(context.Background())
}

func teardownMongoTest() {
	_ = testDB.Collection("achievements").Drop(context.Background())
}

// =======================================================
// TEST: CreateDraft
// =======================================================

func TestMongoAchievement_CreateDraft(t *testing.T) {
	setupMongoTest(t)
	defer teardownMongoTest()

	req := &models.CreateAchievementRequest{
		AchievementType: "competition",
		Title:           "Lomba Nasional",
		Description:     "Juara 1",
		Tags:            []string{"nasional"},
	}

	result, err := testRepo.CreateDraft(
		context.Background(),
		"student-1",
		req,
		100,
	)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "student-1", result.StudentID)
	assert.Equal(t, models.StatusDraft, result.Status)
	assert.NotNil(t, result.Points)
}

// =======================================================
// TEST: GetByID
// =======================================================

func TestMongoAchievement_GetByID(t *testing.T) {
	setupMongoTest(t)
	defer teardownMongoTest()

	req := &models.CreateAchievementRequest{
		AchievementType: "competition",
		Title:           "Test Prestasi",
		Description:     "Deskripsi",
	}

	created, _ := testRepo.CreateDraft(
		context.Background(),
		"student-1",
		req,
		50,
	)

	found, err := testRepo.GetByID(context.Background(), created.ID.Hex())

	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, created.Title, found.Title)
}

// =======================================================
// TEST: UpdateStatus
// =======================================================

func TestMongoAchievement_UpdateStatus(t *testing.T) {
	setupMongoTest(t)
	defer teardownMongoTest()

	req := &models.CreateAchievementRequest{
		AchievementType: "competition",
		Title:           "Update Status",
		Description:     "Test",
	}

	created, _ := testRepo.CreateDraft(
		context.Background(),
		"student-1",
		req,
		80,
	)

	err := testRepo.UpdateStatus(
		context.Background(),
		created.ID.Hex(),
		models.StatusSubmitted,
	)

	assert.NoError(t, err)

	updated, _ := testRepo.GetByID(context.Background(), created.ID.Hex())
	assert.Equal(t, models.StatusSubmitted, updated.Status)
}

// =======================================================
// TEST: GetManyByIDs
// =======================================================

func TestMongoAchievement_GetManyByIDs(t *testing.T) {
	setupMongoTest(t)
	defer teardownMongoTest()

	ids := []string{}

	for i := 0; i < 2; i++ {
		req := &models.CreateAchievementRequest{
			AchievementType: "competition",
			Title:           "Bulk Test",
			Description:     "Bulk",
		}

		a, _ := testRepo.CreateDraft(
			context.Background(),
			"student-1",
			req,
			10,
		)
		ids = append(ids, a.ID.Hex())
	}

	result, err := testRepo.GetManyByIDs(context.Background(), ids)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// =======================================================
// TEST: SoftDelete
// =======================================================

func TestMongoAchievement_SoftDelete(t *testing.T) {
	setupMongoTest(t)
	defer teardownMongoTest()

	req := &models.CreateAchievementRequest{
		AchievementType: "competition",
		Title:           "Delete Test",
		Description:     "Delete",
	}

	created, _ := testRepo.CreateDraft(
		context.Background(),
		"student-1",
		req,
		30,
	)

	err := testRepo.SoftDelete(context.Background(), created.ID.Hex())
	assert.NoError(t, err)

	// verify deleted
	col := testDB.Collection("achievements")
	var raw bson.M
	_ = col.FindOne(context.Background(), bson.M{"_id": created.ID}).Decode(&raw)

	assert.Equal(t, true, raw["isDeleted"])
	assert.Equal(t, models.StatusDeleted, raw["status"])
}

// =======================================================
// TEST: UpdateDraft
// =======================================================

func TestMongoAchievement_UpdateDraft(t *testing.T) {
	setupMongoTest(t)
	defer teardownMongoTest()

	req := &models.CreateAchievementRequest{
		AchievementType: "competition",
		Title:           "Before",
		Description:     "Before",
	}

	created, _ := testRepo.CreateDraft(
		context.Background(),
		"student-1",
		req,
		20,
	)

	updateReq := &models.UpdateAchievementRequest{
		AchievementType: "competition",
		Title:           "After",
		Description:     "After",
	}

	updated, err := testRepo.UpdateDraft(
		context.Background(),
		created.ID.Hex(),
		updateReq,
		99,
	)

	assert.NoError(t, err)
	assert.Equal(t, "After", updated.Title)
	assert.WithinDuration(t, time.Now(), updated.UpdatedAt, time.Second)
}

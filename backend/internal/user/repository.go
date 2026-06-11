package user

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserProfile defines how a user document looks inside MongoDB
type UserProfile struct {
	FirebaseUID string    `bson:"_id" json:"uid"` // Using Firebase UID as the unique document ID
	DisplayName string    `bson:"display_name" json:"display_name"`
	Email       string    `bson:"email" json:"email"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
}

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(client *mongo.Client) *UserRepository {
	// Selects the database "booking_db" and collection "users"
	db := client.Database("booking_db")
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

// UpsertProfile inserts a new profile or updates an existing one if the UID matches
func (r *UserRepository) UpsertProfile(ctx context.Context, profile UserProfile) error {
	filter := bson.M{"_id": profile.FirebaseUID}

	// If the profile already exists, we overwrite everything except the creation date
	update := bson.M{
		"$set": bson.M{
			"display_name": profile.DisplayName,
			"email":        profile.Email,
			"updated_at":   time.Now(),
		},
		"$setOnInsert": bson.M{
			"created_at": time.Now(),
		},
	}

	// upsert: true makes it handle both creations and updates automatically
	opts := options.Update().SetUpsert(true)

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// GetProfileByID searches MongoDB for a profile by its Firebase UID string
func (r *UserRepository) GetProfileByID(ctx context.Context, uid string) (*UserProfile, error) {
	var profile UserProfile
	filter := bson.M{"_id": uid}

	err := r.collection.FindOne(ctx, filter).Decode(&profile)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

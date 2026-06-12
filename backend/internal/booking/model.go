package booking

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Booking struct {
	ID        int       `bson:"_id" json:"id"`
	UserID    string    `bson:"user_id" json:"user_id"`
	MovieID   int       `bson:"movie_id" json:"movie_id"`
	Seats     []string  `bson:"seats" json:"seats"`
	Showtime  string    `bson:"showtime" json:"showtime"`
	Status    string    `bson:"status" json:"status"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

type SystemLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	EventType string             `bson:"event_type" json:"event_type"`
	Message   string             `bson:"message" json:"message"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

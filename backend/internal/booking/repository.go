package booking

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BookingRepository struct {
	collection *mongo.Collection
}

func NewBookingRepository(client *mongo.Client) *BookingRepository {
	return &BookingRepository{
		collection: client.Database("booking_db").Collection("bookings"),
	}
}

func (r *BookingRepository) getNextSequence(ctx context.Context, name string) (int, error) {
	countersColl := r.collection.Database().Collection("counters")
	filter := bson.M{"_id": name}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result struct {
		Seq int `bson:"seq"`
	}
	err := countersColl.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	return result.Seq, err
}

func (r *BookingRepository) Create(ctx context.Context, b *Booking) error {
	if b.ID == 0 {
		seq, err := r.getNextSequence(ctx, "bookings")
		if err != nil {
			return err
		}
		b.ID = seq
	}
	b.CreatedAt = time.Now()
	if b.Status == "" {
		b.Status = "confirmed"
	}
	_, err := r.collection.InsertOne(ctx, b)
	return err
}

func (r *BookingRepository) GetByUserID(ctx context.Context, userID string) ([]Booking, error) {
	bookings := []Booking{}
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return bookings, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &bookings); err != nil {
		return bookings, err
	}
	return bookings, nil
}

func (r *BookingRepository) GetAll(ctx context.Context) ([]Booking, error) {
	bookings := []Booking{}
	opts := options.Find().SetSort(bson.M{"created_at": -1}) // Newest first
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return bookings, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &bookings); err != nil {
		return bookings, err
	}
	return bookings, nil
}

func (r *BookingRepository) GetOccupiedSeats(ctx context.Context, movieID int, showtime string) ([]string, error) {
	filter := bson.M{
		"movie_id": movieID,
		"showtime": showtime,
		"status":   "confirmed",
	}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return []string{}, err
	}
	defer cursor.Close(ctx)

	occupied := []string{}
	for cursor.Next(ctx) {
		var b Booking
		if err := cursor.Decode(&b); err == nil {
			occupied = append(occupied, b.Seats...)
		}
	}
	return occupied, nil
}

func (r *BookingRepository) GetOccupiedSeat(ctx context.Context, movieID int) ([]string, error) {
	filter := bson.M{
		"movie_id": movieID,
		"status":   "confirmed",
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var occupiedSeats []string
	for cursor.Next(ctx) {
		var b Booking
		if err = cursor.Decode(&b); err != nil {
			occupiedSeats = append(occupiedSeats, b.Seats...)
		}
	}
	return occupiedSeats, nil
}

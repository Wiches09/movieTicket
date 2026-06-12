package booking

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BookingRepository struct {
	collection *mongo.Collection
	redis      *redis.Client
}

func NewBookingRepository(client *mongo.Client, rdb *redis.Client) *BookingRepository {
	return &BookingRepository{
		collection: client.Database("booking_db").Collection("bookings"),
		redis:      rdb,
	}
}

func (r *BookingRepository) LockSeat(ctx context.Context, movieID int, showtime, seat, userID string) (bool, error) {
	key := fmt.Sprintf("lock:%d:%s:%s", movieID, showtime, seat)
	// NX: Set only if the key does not exist
	success, err := r.redis.SetNX(ctx, key, userID, 5*time.Minute).Result()
	return success, err
}

func (r *BookingRepository) UnlockSeat(ctx context.Context, movieID int, showtime, seat, userID string) error {
	key := fmt.Sprintf("lock:%d:%s:%s", movieID, showtime, seat)

	// Only unlock if it was locked by the same user
	val, err := r.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return err
	}

	if val == userID {
		return r.redis.Del(ctx, key).Err()
	}
	return nil
}

func (r *BookingRepository) GetLockedSeats(ctx context.Context, movieID int, showtime string) (map[string]string, error) {
	pattern := fmt.Sprintf("lock:%d:%s:*", movieID, showtime)
	keys, err := r.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	locks := make(map[string]string)
	for _, key := range keys {
		userID, _ := r.redis.Get(ctx, key).Result()
		// key format is lock:movieID:showtime:seat
		parts := fmt.Sprintf("lock:%d:%s:", movieID, showtime)
		seat := key[len(parts):]
		locks[seat] = userID
	}
	return locks, nil
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

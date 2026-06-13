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
	logsColl   *mongo.Collection
	redis      *redis.Client
}

func NewBookingRepository(client *mongo.Client, rdb *redis.Client) *BookingRepository {
	db := client.Database("booking_db")
	return &BookingRepository{
		collection: db.Collection("bookings"),
		logsColl:   db.Collection("system_logs"),
		redis:      rdb,
	}
}

func (r *BookingRepository) InsertLog(ctx context.Context, eventType, message string) error {
	logEntry := SystemLog{
		EventType: eventType,
		Message:   message,
		CreatedAt: time.Now(),
	}
	_, err := r.logsColl.InsertOne(ctx, logEntry)
	return err
}

func (r *BookingRepository) GetSystemLogs(ctx context.Context) ([]SystemLog, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(100)
	cursor, err := r.logsColl.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	var logs []SystemLog
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *BookingRepository) LockSeat(ctx context.Context, movieID int, showtime, seat, userID string) (bool, error) {
	key := fmt.Sprintf("lock:%d:%s:%s", movieID, showtime, seat)
	fmt.Printf("[DEBUG] Locking seat: %s for user: %s (TTL: 10s)\n", key, userID)
	// NX: Set only if the key does not exist.
	// Set TTL slightly longer than the application timeout (5m) to allow manual timeout handling.
	success, err := r.redis.SetNX(ctx, key, userID, 6*time.Minute).Result()
	return success, err
}

func (r *BookingRepository) UnlockSeat(ctx context.Context, movieID int, showtime, seat, userID string) (bool, error) {
	key := fmt.Sprintf("lock:%d:%s:%s", movieID, showtime, seat)
	fmt.Printf("[DEBUG] Attempting to unlock seat: %s for user: %s\n", key, userID)

	// Only unlock if it was locked by the same user
	val, err := r.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		fmt.Printf("[DEBUG] Unlock failed: Key %s not found in Redis\n", key)
		return false, nil
	}
	if err != nil {
		fmt.Printf("[DEBUG] Unlock error: %v\n", err)
		return false, err
	}

	if val == userID {
		deleted, err := r.redis.Del(ctx, key).Result()
		fmt.Printf("[DEBUG] Unlock result for %s: %v (deleted: %d)\n", key, deleted > 0, deleted)
		return deleted > 0, err
	}
	fmt.Printf("[DEBUG] Unlock failed: User mismatch. Expected %s, found %s\n", userID, val)
	return false, nil
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

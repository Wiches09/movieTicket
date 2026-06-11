package movie

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MovieRepository struct {
	collection *mongo.Collection
}

func NewMovieRepository(client *mongo.Client) *MovieRepository {
	return &MovieRepository{
		collection: client.Database("booking_db").Collection("movies"),
	}
}

func (r *MovieRepository) getNextSequence(ctx context.Context, name string) (int, error) {
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

func (r *MovieRepository) GetAll(ctx context.Context) ([]Movie, error) {
	var movies []Movie
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &movies); err != nil {
		return nil, err
	}
	return movies, nil
}

func (r *MovieRepository) GetByID(ctx context.Context, id int) (*Movie, error) {
	var movie Movie
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&movie)
	if err != nil {
		return nil, err
	}
	return &movie, nil
}

func (r *MovieRepository) Create(ctx context.Context, movie Movie) error {
	if movie.ID == 0 {
		seq, err := r.getNextSequence(ctx, "movies")
		if err != nil {
			return err
		}
		movie.ID = seq
	}
	_, err := r.collection.InsertOne(ctx, movie)
	return err
}

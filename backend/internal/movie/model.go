package movie

type Movie struct {
	ID    int    `bson:"_id" json:"id"`
	Title string `bson:"title" json:"title"`
	Year  int    `bson:"year" json:"year"`
}

package post

import "time"

type Post struct {
	ID         string    `bson:"_id,omitempty"`
	AuthorID   int64     `bson:"author_id"`
	Content    string    `bson:"content"`
	Images     []string  `bson:"images"`
	DiseaseTag string    `bson:"disease_tag"`
	Status     string    `bson:"status"`
	CreatedAt  time.Time `bson:"created_at"`
}

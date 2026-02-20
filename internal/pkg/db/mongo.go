package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

func InitMongo(uri string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return err
	}
	MongoClient = client
	return nil
}

// 更新示例：在集合 users 中更新一个字段
func UpdateMongoUser(ctx context.Context, dbName, coll string, filter bson.M, update bson.M) (*mongo.UpdateResult, error) {
	collection := MongoClient.Database(dbName).Collection(coll)
	return collection.UpdateOne(ctx, filter, bson.M{"$set": update})
}

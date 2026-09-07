package greeting

import (
	"context"
	"errors"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const mongoDocID = "greeting"

type mongoDoc struct {
	ID        string    `bson:"_id"`
	Name      string    `bson:"name"`
	UpdatedAt time.Time `bson:"updatedAt"`
	UpdatedBy string    `bson:"updatedBy"`
	Version   int64     `bson:"version"`
}

// MongoView is the CQRS read model: one document with the current greeting.
type MongoView struct {
	col *mongo.Collection
}

func OpenMongoView(ctx context.Context, url, dbName string) (*MongoView, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}
	return &MongoView{col: client.Database(dbName).Collection("greeting")}, nil
}

func (v *MongoView) Get() (Record, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var doc mongoDoc
	err := v.col.FindOne(ctx, bson.M{"_id": mongoDocID}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return Record{}, err
	}
	if err != nil {
		return Record{}, err
	}
	return Record{
		Name:      doc.Name,
		UpdatedAt: doc.UpdatedAt.UTC(),
		UpdatedBy: doc.UpdatedBy,
		Version:   doc.Version,
	}, nil
}

func (v *MongoView) Apply(evt Event) error {
	rec, err := Fold([]Event{evt})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	filter := bson.M{
		"_id": mongoDocID,
		"$or": bson.A{
			bson.M{"version": bson.M{"$lte": rec.Version}},
			bson.M{"version": bson.M{"$exists": false}},
		},
	}
	_, err = v.col.UpdateOne(ctx, filter, bson.M{
		"$set": bson.M{
			"name":      rec.Name,
			"updatedAt": rec.UpdatedAt,
			"updatedBy": rec.UpdatedBy,
			"version":   rec.Version,
		},
	}, options.Update().SetUpsert(true))
	return err
}

func (v *MongoView) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return v.col.Database().Client().Ping(ctx, nil)
}

func MongoURL() string {
	if v := os.Getenv("MONGO_URL"); v != "" {
		return v
	}
	return "mongodb://localhost:27017"
}

func MongoDBName() string {
	if v := os.Getenv("MONGO_DB"); v != "" {
		return v
	}
	return "hola"
}

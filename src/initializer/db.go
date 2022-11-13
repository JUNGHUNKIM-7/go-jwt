package initializer

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mgo struct {
	Client     *mongo.Client
	Collection *mongo.Collection
}

var Mongos *Mgo = new(Mgo)

func Db() error {
	if err := godotenv.Load(); err != nil {
		return errors.New(strings.ToLower("No .env file found"))
	}
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		return errors.New(strings.ToLower("You must set your 'MONGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable"))
	}
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	Mongos.Client = client
	if err != nil {
		panic(err)
	}
	Mongos.Collection = client.Database(os.Getenv("DB")).Collection(os.Getenv("COLL"))
	return nil
}
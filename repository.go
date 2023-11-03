package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const database_name = "ts3bot"
const whitelist_collection = "whitelist"
const subscribers_collection = "subscribers"

type Repository struct {
	Client *mongo.Client
}

func CreateRepository(config *Config) (*Repository, error) {
	clientOptions := options.Client().ApplyURI(config.Bot.MongodbUri)

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	log.Println("Connected to MongoDB!")

	return &Repository{
		Client: client,
	}, nil
}

type WhiteListEntry struct {
	Id    string `bson:"_id"`
	Alias string `bson:"alias"`
}

func (repository *Repository) AddWhiteListEntry(telegram_id int64, alias string) error {
	collection := repository.Client.Database(database_name).Collection(whitelist_collection)
	_, err := collection.InsertOne(context.Background(), WhiteListEntry{Id: fmt.Sprintf("%d", telegram_id), Alias: alias})
	return err
}

func (repository *Repository) RemoveWhiteListEntry(telegram_id int64) error {
	collection := repository.Client.Database(database_name).Collection(whitelist_collection)
	_, err := collection.DeleteOne(context.Background(), bson.M{"_id": fmt.Sprintf("%d", telegram_id)})
	return err
}

func (repository *Repository) IsOnWhitelist(telegram_id int64) (bool, error) {
	collection := repository.Client.Database(database_name).Collection(whitelist_collection)
	count, err := collection.CountDocuments(context.Background(), bson.M{"_id": fmt.Sprintf("%d", telegram_id)})
	return count > 0, err
}

func (repository *Repository) GetWhiteList() ([]WhiteListEntry, error) {
	collection := repository.Client.Database(database_name).Collection(whitelist_collection)
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	var results []WhiteListEntry
	if err = cursor.All(context.Background(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

type UserSubscribers struct {
	Id                  string  `bson:"_id"`
	Name                string  `bson:"name"`
	TelegramSubscribers []int64 `bson:"subscribers"`
}

func (repository *Repository) AddSubscriber(subscriber_telegram_id int64, subscribed_teamspeak_id string, subscribed_teamspeak_name string) error {
	collection := repository.Client.Database(database_name).Collection(subscribers_collection)
	_, err := collection.UpdateOne(context.Background(), bson.M{"_id": subscribed_teamspeak_id}, bson.M{"$addToSet": bson.M{"subscribers": subscriber_telegram_id}, "$set": bson.M{"name": subscribed_teamspeak_name}}, options.Update().SetUpsert(true))
	return err
}

func (repository *Repository) RemoveSubscriber(subscriber_telegram_id int64, subscribed_teamspeak_id string) error {
  collection := repository.Client.Database(database_name).Collection(subscribers_collection)
  _, err := collection.UpdateOne(context.Background(), bson.M{"_id": subscribed_teamspeak_id}, bson.M{"$pull": bson.M{"subscribers": subscriber_telegram_id}})
  return err
}

func (repository *Repository) GetSubscribers(subscribed_teamspeak_id string) (UserSubscribers, error) {
  collection := repository.Client.Database(database_name).Collection(subscribers_collection)
  var result UserSubscribers
  err := collection.FindOne(context.Background(), bson.M{"_id": subscribed_teamspeak_id}).Decode(&result)
  return result, err
}

func (repository *Repository) GetSubscribedTeamspeaks(subscriber_telegram_id int64) ([]UserSubscribers, error) {
  collection := repository.Client.Database(database_name).Collection(subscribers_collection)
  cursor, err := collection.Find(context.Background(), bson.M{"subscribers": subscriber_telegram_id})
  if err != nil {
    return nil, err
  }
  var results []UserSubscribers
  if err = cursor.All(context.Background(), &results); err != nil {
    return nil, err
  }
  return results, nil
}

type Quote struct {
	UUID      string `bson:"_id"`
	Author    string `bson:"author"`
	Content   string `bson:"content"`
  CreatedBy int64 `bson:"created_by"`
}

func (repository *Repository) AddQuote(quote Quote) error {
	collection := repository.Client.Database(database_name).Collection("quotes")
	_, err := collection.InsertOne(context.Background(), quote)
	return err
}

func (repository *Repository) GetAllQuotes() ([]Quote, error) {
	collection := repository.Client.Database(database_name).Collection("quotes")
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}

	var quotes []Quote
	if err = cursor.All(context.Background(), &quotes); err != nil {
		return nil, err
	}
	return quotes, nil
}

func (repository *Repository) DeleteQuote(uuid string) error {
	collection := repository.Client.Database(database_name).Collection("quotes")
	_, err := collection.DeleteOne(context.Background(), bson.M{"_id": uuid})
	return err
}

type Property struct {
	ID    string `bson:"_id"`
	Value string `bson:"value"`
}

func (repository *Repository) SetProperty(property Property) error {
	collection := repository.Client.Database(database_name).Collection("properties")
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"_id": property.ID}
	update := bson.M{"$set": property}
	_, err := collection.UpdateOne(context.Background(), filter, update, opts)
	return err
}

func (repository *Repository) GetProperty(id string, defaultValueFunc func() string) (Property, error) {
	collection := repository.Client.Database(database_name).Collection("properties")
	var property Property
	err := collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&property)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return Property{ID: id, Value: defaultValueFunc()}, nil
		}
		return Property{}, err
	}
	return property, nil
}

func (repository *Repository) SetQuotesChannel(value string) (error) {
  property := Property{ID: "quotes_channel", Value: value}
  err := repository.SetProperty(property)
  return err
}

func (repository *Repository) GetQuotesChannel() (string, bool) {
  property, err := repository.GetProperty("quotes_channel", func() string { return "" })
  if err != nil || property.Value == "" {
    return "", false 
  }
  return property.Value, true
}

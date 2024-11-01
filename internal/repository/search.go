package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type SearchStorage struct {
	postgres *sql.DB
	mongo    *mongo.Database
	logger   *zap.SugaredLogger
}

func NewSearchStorage(p *sql.DB, m *mongo.Database, logger *zap.SugaredLogger) *SearchStorage {
	return &SearchStorage{
		postgres: p,
		mongo:    m,
		logger:   logger,
	}
}

func (s *SearchStorage) AddJsonFileToDB(pathToFile string) {
	jsonArr := s.loadJSONArrayFromFile(pathToFile)
	collection := s.mongo.Collection("items")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertMany(ctx, jsonArr)
	if err != nil {
		s.logger.Error(err)
		return
	}
	s.logger.Info("JSON массив успешно добавлен в MongoDB!")
}

func (s *SearchStorage) loadJSONArrayFromFile(path string) []interface{} {
	file, err := os.Open(path)
	if err != nil {
		s.logger.Error(err)
		return nil
	}
	defer file.Close()
	var jsonArray []interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&jsonArray); err != nil {
		s.logger.Error(err)
		return nil
	}

	return jsonArray
}

func (s *SearchStorage) CreateSearchIndex() {
	collection := s.mongo.Collection("items")
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "question", Value: "text"}, {Key: "answer", Value: "text"}},
		//Options: options.Index().SetDefaultLanguage("english"),
	}
	_, err := collection.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		log.Fatal(err)
	}

}

func (s *SearchStorage) Search(userMessage *domain.Message) *domain.Message {
	mongoSearchResults := s.FindRelevantContext(userMessage.TranslatedMessageText)
	if mongoSearchResults != nil {
		userMessage.Context = mongoSearchResults
	}
	return userMessage
}

func (s *SearchStorage) FindRelevantContext(userMessage string) []bson.M {
	collection := s.mongo.Collection("items")
	filter := bson.M{"$text": bson.M{"$search": userMessage}}

	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil
	}
	defer cursor.Close(context.TODO())

	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil
	}

	return results
}

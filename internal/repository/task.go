package repository

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/re-tofl/tofl-gpt-chat/internal/bootstrap"
	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
)

type TaskStorage struct {
	postgres *sql.DB
	mongo    *mongo.Database
	logger   *zap.SugaredLogger
	cfg      *bootstrap.Config
}

type TranslateRequest struct {
	Messages           []string `json:"texts"`
	FolderID           string   `json:"folderId"`
	TargetLanguageCode string   `json:"targetLanguageCode"`
}

type TranslateResponse struct {
	Translations []Translations `json:"translations"`
}

type Translations struct {
	Text                 string `json:"text"`
	DetectedLanguageCode string `json:"detectedLanguageCode"`
}

func NewTaskStorage(p *sql.DB, m *mongo.Database, logger *zap.SugaredLogger, cfg *bootstrap.Config) *TaskStorage {
	return &TaskStorage{
		postgres: p,
		mongo:    m,
		logger:   logger,
		cfg:      cfg,
	}
}

func (ts *TaskStorage) Solve(message *domain.Message) {

}

func (ts *TaskStorage) Answer(message *domain.Message) {

}

func (ts *TaskStorage) Translate(message *domain.Message) *domain.Message {
	translateRequest := TranslateRequest{
		Messages:           make([]string, 0),
		FolderID:           ts.cfg.YandexTranslateFolderId,
		TargetLanguageCode: "en",
	}

	translateRequest.Messages = append(translateRequest.Messages, message.OriginalMessageText)
	translateResponse := ts.SendToYandex(translateRequest)
	message.TranslatedMessageText = translateResponse.Translations[0].Text // TODO fix this hardcode
	return message
}

func (ts *TaskStorage) SendToYandex(request TranslateRequest) (response TranslateResponse) {
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		ts.logger.Error(err)
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", ts.cfg.YandexTranslateUrl, bytes.NewBuffer(jsonRequest))
	if err != nil {
		ts.logger.Error(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", ts.cfg.YandexKey)
	resp, err := client.Do(req)
	if err != nil {
		ts.logger.Error(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ts.logger.Error(err)
	}
	response = ts.ParseTranslateResponse(body)
	return response
}

func (ts *TaskStorage) ParseTranslateResponse(jsonResponse []byte) (response TranslateResponse) {
	err := json.Unmarshal(jsonResponse, &response)
	if err != nil {
		ts.logger.Error(err)
	}
	return response
}

func (ts *TaskStorage) AddJsonFileToDB(pathToFile string) {
	jsonArr := ts.loadJSONArrayFromFile(pathToFile)
	collection := ts.mongo.Collection("items")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertMany(ctx, jsonArr)
	if err != nil {
		ts.logger.Error(err)
		return
	}
	ts.logger.Info("JSON массив успешно добавлен в MongoDB!")
}

func (ts *TaskStorage) loadJSONArrayFromFile(path string) []interface{} {
	file, err := os.Open(path)
	if err != nil {
		ts.logger.Error(err)
		return nil
	}
	defer file.Close()
	var jsonArray []interface{}
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&jsonArray); err != nil {
		ts.logger.Error(err)
		return nil
	}

	return jsonArray
}

func (ts *TaskStorage) CreateSearchIndex() {
	collection := ts.mongo.Collection("items")
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "question", Value: "text"}, {Key: "answer", Value: "text"}},
		//Options: options.Index().SetDefaultLanguage("english"),
	}
	_, err := collection.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		log.Fatal(err)
	}
}

func (ts *TaskStorage) Search(userMessage *domain.Message) *domain.Message {
	mongoSearchResults := ts.FindRelevantContext(userMessage.TranslatedMessageText)
	if mongoSearchResults != nil {
		userMessage.Context = mongoSearchResults
	}
	return userMessage
}

func (ts *TaskStorage) FindRelevantContext(userMessage string) []bson.M {
	collection := ts.mongo.Collection("items")
	filter := bson.M{"$text": bson.M{"$search": userMessage}}

	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		ts.logger.Error(err)
		return nil
	}
	defer cursor.Close(context.TODO())

	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		ts.logger.Error(err)
		return nil
	}

	return results
}

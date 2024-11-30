package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/re-tofl/tofl-gpt-chat/internal/adapters"
	"go.uber.org/zap"

	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
)

type SearchStorage struct {
	postgres *adapters.AdapterPG
	logger   *zap.SugaredLogger
}

func NewSearchStorage(p *adapters.AdapterPG, logger *zap.SugaredLogger) *SearchStorage {
	return &SearchStorage{
		logger:   logger,
		postgres: p,
	}
}

/*func (s *SearchStorage) AddJsonFileToDB(pathToFile string) {
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
}*/

func (s *SearchStorage) LoadJSONArrayFromFile(path string) []domain.DatabaseItem {
	file, err := os.Open(path)
	if err != nil {
		s.logger.Error(err)
		return nil
	}
	defer file.Close()
	var jsonArray []domain.DatabaseItem
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&jsonArray); err != nil {
		s.logger.Error(err)
		return nil
	}

	return jsonArray
}

func (s *SearchStorage) DoDatabaseEmbedding(ctx context.Context, items []domain.DatabaseItem) {
	/*for _, item := range items {
		vector := s.SendEmbeddingReq(item)

	}*/
}

func (s *SearchStorage) PushVectorToMatrix(vector domain.EmbeddingResp) {

}

func (s *SearchStorage) SendEmbeddingReq(item domain.DatabaseItem) domain.EmbeddingResp {
	jsonReq, err := json.Marshal(item)
	if err != nil {
		s.logger.Error(err)
	}

	req, err := http.NewRequest("POST", "http://127.0.0.1:8000/embed", bytes.NewBuffer(jsonReq))
	if err != nil {
		s.logger.Error(err)
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error(err)
	}

	defer resp.Body.Close()
	byteResp, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error(err)
	}

	embeddingResponse := domain.EmbeddingResp{}
	err = json.Unmarshal(byteResp, &embeddingResponse)
	if err != nil {
		s.logger.Error(err)
	}
	return embeddingResponse
}

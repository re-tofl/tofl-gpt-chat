package adapters

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/re-tofl/tofl-gpt-chat/internal/bootstrap"
)

type AdapterMongo struct {
	*mongo.Database
	cfg *bootstrap.Config
}

func NewAdapterMongo(cfg *bootstrap.Config) *AdapterMongo {
	return &AdapterMongo{
		cfg: cfg,
	}
}

func (a *AdapterMongo) Init(ctx context.Context) error {
	uri := "mongodb://root:Artem557@localhost:27017/tofl_gpt_chat?authSource=admin"

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Ошибка подключения к MongoDB: %v", err)
	}

	// Проверка подключения
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Не удалось пропинговать MongoDB: %v", err)
	}

	a.Database = client.Database("tofl_gpt_chat")

	// Вставка тестового документа (для проверки)
	testCollection := a.Database.Collection("test_collection")
	_, err = testCollection.InsertOne(ctx, bson.M{"name": "test", "value": "Это тестовый документ"})
	if err != nil {
		log.Fatalf("Ошибка вставки тестового документа: %v", err)
	}

	log.Println("Успешно подключено к MongoDB и вставлен тестовый документ")
	return nil
}

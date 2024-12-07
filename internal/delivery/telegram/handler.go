package telegram

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/re-tofl/tofl-gpt-chat/internal/adapters"
	"github.com/re-tofl/tofl-gpt-chat/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/re-tofl/tofl-gpt-chat/internal/bootstrap"
	"go.uber.org/zap"
)

type OpenAiUsecase interface {
	SaveMedia(ctx context.Context, message *tgbotapi.Message, bot *tgbotapi.BotAPI) string
	SendToGpt(ctx context.Context, message *tgbotapi.Message) string
}

type SpeechUsecase interface {
	ConvertSpeechToText(ctx context.Context, filePath string) string
}

type TaskUsecase interface {
	RateTheory(ctx context.Context, message *tgbotapi.Message) error
}

type SearchUsecase interface {
	DatabaseToVector(ctx context.Context)
}

type Handler struct {
	cfg            *bootstrap.Config
	log            *zap.SugaredLogger
	bot            *tgbotapi.BotAPI
	openAiUC       OpenAiUsecase
	speechUC       SpeechUsecase
	taskUC         TaskUsecase
	mu             sync.Mutex
	userStates     map[int64]int
	userContextIDs map[int64]string
	mongo          *adapters.AdapterMongo
	achs           map[int64]domain.Achievement
	postgres       *adapters.AdapterPG
	searchUC       SearchUsecase
}

func NewHandler(cfg *bootstrap.Config, log *zap.SugaredLogger,
	o OpenAiUsecase, s SpeechUsecase, t TaskUsecase, m *adapters.AdapterMongo, search SearchUsecase) *Handler {
	return &Handler{
		cfg:        cfg,
		log:        log,
		openAiUC:   o,
		speechUC:   s,
		taskUC:     t,
		userStates: make(map[int64]int),
		mongo:      m,
		achs:       CreateAchMap(),
		searchUC:   search,
	}
}

func (h *Handler) Listen(ctx context.Context) error {
	bot, err := tgbotapi.NewBotAPI(h.cfg.TGBotToken)
	if err != nil {
		h.log.Errorw("tgbotapi.NewBotAPI", zap.Error(err))
		return fmt.Errorf("tgbotapi.NewBotAPI: %w", err)
	}

	h.bot = bot
	h.log.Infof("logged in TG with username %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		h.log.Errorw("bot.GetUpdatesChan", zap.Error(err))
		return fmt.Errorf("bot.GetUpdatesChan: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			bot.StopReceivingUpdates()
			return nil
		case update := <-updates:
			// TODO: сделать многопоточным
			h.processUpdate(ctx, &update)
		}
	}
}

func (h *Handler) Send(c tgbotapi.Chattable) *tgbotapi.Message {
	send, err := h.bot.Send(c)
	if err != nil {
		h.log.Errorw("bot.Send", zap.Error(err))
		return nil
	}

	return &send
}

func (h *Handler) processUpdate(ctx context.Context, update *tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	messageText := update.Message.Text
	if messageText == "" {
		messageText = update.Message.Caption
		update.Message.Text = messageText
	}

	h.log.Infow("received message",
		"username", update.Message.From.UserName,
		"message", messageText)

	if update.Message.IsCommand() {
		h.processCommand(ctx, update.Message)
		return
	}

	if update.Message.Photo != nil || update.Message.Document != nil {
		h.log.Infow("received photo or file from ",
			"username", update.Message.From.UserName)

		h.handleProblemWithImages(ctx, update.Message)
		return
	}

	if update.Message.Voice != nil {
		h.handleVoice(ctx, update.Message, h.bot)
	}

	h.handleMessage(ctx, update.Message)
}

func (h *Handler) handleProblemWithImages(ctx context.Context, message *tgbotapi.Message) {
	gptFileResponse := h.openAiUC.SaveMedia(ctx, message, h.bot)
	reply := tgbotapi.NewMessage(message.Chat.ID, gptFileResponse)
	h.Send(reply)
}

func (h *Handler) handleVoice(ctx context.Context, message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	fileId := message.Voice.FileID
	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileId})
	if err != nil {
		h.log.Error(err)
	}

	filePath, err := h.SaveAndDownloadVoice(file.FilePath, file.FileID)
	textFromSpeech := h.speechUC.ConvertSpeechToText(ctx, filePath)

	message.Text = textFromSpeech

	h.handleGptTextMessage(ctx, message)
}

func (h *Handler) SaveAndDownloadVoice(tgFilePath string, fileName string) (string, error) {
	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", h.cfg.TGBotToken, tgFilePath)

	resp, err := http.Get(fileURL)
	if err != nil {
		return "", fmt.Errorf("ошибка при загрузке файла: %w", err)
	}
	defer resp.Body.Close()

	filePath := filepath.Join("upload/voices", fileName+".ogg")
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("ошибка при создании файла: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка при сохранении файла: %w", err)
	}

	return filePath, nil
}
func (h *Handler) handleGptTextMessage(ctx context.Context, message *tgbotapi.Message) {
	if CheckUserExists(ctx, message.Chat.ID, h.mongo.Database) {
		AddUserReqIntoMongo(ctx, message.Chat.ID, h.mongo.Database)
		fmt.Println("Пользователь существует, выполняем обновление")
	} else {
		fmt.Println("Пользователь не найден, выполняем вставку")
		AddUserReqIntoMongo(ctx, message.Chat.ID, h.mongo.Database)
	}

	countOfReq := CheckCountOfReq(ctx, message.Chat.ID, h.mongo.Database)
	if ach, exists := h.achs[countOfReq]; exists {
		InsertAchIntoMongo(ctx, ach, message.Chat.ID, h.mongo.Database)
		achtext := fmt.Sprintf("🎉 Вы забрали новую ачивку! 🎉\n\n🏆 *%s*\n\n📜 %s\n\n⭐ Оценка вашего гигачадства: *%s* ТФЯ",
			ach.Title, ach.Desc, ach.Grade)

		achReply := tgbotapi.NewMessage(message.Chat.ID, achtext)
		achReply.ParseMode = "Markdown" // Используем Markdown для форматирования

		h.Send(achReply)
	}

	gptResponse := h.openAiUC.SendToGpt(ctx, message)
	reply := tgbotapi.NewMessage(message.Chat.ID, gptResponse)
	h.Send(reply)
}

func (h *Handler) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	reply := tgbotapi.NewMessage(message.Chat.ID, "")

	state, ok := h.userStates[message.Chat.ID]
	if !ok {
		h.mu.Lock()
		h.userStates[message.Chat.ID] = domain.StartState
		h.mu.Unlock()
	}

	switch state {
	case domain.GPTInputState:
		question := message.Text
		reply.Text = fmt.Sprintf("Ты задал вопрос: %s. Отправляю в GPT...", question)
		h.Send(reply)
		h.handleGptTextMessage(ctx, message)

	case domain.ProblemInputState:
		err := h.Problem(ctx, message)
		if err != nil {
			h.log.Errorw("h.Problem", zap.Error(err))
			h.Send(tgbotapi.NewMessage(message.Chat.ID, "Произошла ошибка"))
		}

	case domain.TheoryInputState:
		err := h.Theory(ctx, message)
		if err != nil {
			h.log.Errorw("h.Theory", zap.Error(err))
			h.Send(tgbotapi.NewMessage(message.Chat.ID, "Произошла ошибка"))
		} else {
			h.Send(tgbotapi.NewMessage(message.Chat.ID, "Оцените ответ от 1 до 10"))

			h.mu.Lock()
			h.userStates[message.Chat.ID] = domain.TheoryRateState
			h.mu.Unlock()

			return
		}

	case domain.TheoryRateState:
		err := h.RateTheory(ctx, message)
		if err != nil {
			h.log.Errorw("h.RateTheory", zap.Error(err))
			h.Send(tgbotapi.NewMessage(message.Chat.ID, "Произошла ошибка: "+err.Error()))
		} else {
			h.Send(tgbotapi.NewMessage(message.Chat.ID, "Спасибо за оценку!"))
		}

	default:
		h.Send(tgbotapi.NewMessage(message.Chat.ID, "Введите команду"))
	}

	h.mu.Lock()
	h.userStates[message.Chat.ID] = domain.StartState
	h.mu.Unlock()
}

func (h *Handler) processCommand(ctx context.Context, message *tgbotapi.Message) {
	reply := tgbotapi.NewMessage(message.Chat.ID, "")

	switch message.Command() {
	case "gpt":
		reply.Text = "Введи вопрос к gpt:"

		h.mu.Lock()
		h.userStates[message.Chat.ID] = domain.GPTInputState
		h.mu.Unlock()

		h.Send(reply)
	case "theory":
		reply.Text = "Введите вопрос по теории"

		h.mu.Lock()
		h.userStates[message.Chat.ID] = domain.TheoryInputState
		h.mu.Unlock()

		h.Send(reply)
	case "problem":
		reply.Text = "Введите текст задачи"

		h.mu.Lock()
		h.userStates[message.Chat.ID] = domain.ProblemInputState
		h.mu.Unlock()

		h.Send(reply)

	case "developers":
		reply.Text = "тут будут контакты разработчиков"
		h.Send(reply)

	default:
		reply.Text = "Неизвестная команда"
		h.Send(reply)
	}
}

func (h *Handler) HandleCreateMatrix() {

}

func InsertAchIntoMongo(ctx context.Context, ach interface{}, chatId int64, db *mongo.Database) {
	collection := db.Collection("user_achievement")
	filter := map[string]interface{}{"chat_id": chatId}
	update := map[string]interface{}{
		"$push": map[string]interface{}{
			"Achivments": ach,
		},
	}

	updateResult, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Fatalf("Ошибка обновления документа: %v", err)
	}

	if updateResult.MatchedCount == 0 {
		newDoc := map[string]interface{}{
			"chat_id":    chatId,
			"Achivments": []interface{}{ach},
		}
		insertResult, err := collection.InsertOne(ctx, newDoc)
		if err != nil {
			log.Fatalf("Ошибка вставки нового документа: %v", err)
		}
		fmt.Printf("Новый документ вставлен с ID: %v\n", insertResult.InsertedID)
	} else {
		fmt.Printf("Достижение добавлено для chatId: %v\n", chatId)
	}
}

func AddUserReqIntoMongo(ctx context.Context, chatId int64, db *mongo.Database) {
	fmt.Println("in mongo")
	collection := db.Collection("user_achievement")
	filter := bson.M{"chat_id": chatId}
	update := bson.M{
		"$inc": bson.M{
			"count_of_req": 1,
		},
	}

	// Настройка опций с upsert: true
	opts := options.Update().SetUpsert(true)

	result, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Fatalf("Ошибка обновления документа: %v", err)
	}

	if result.MatchedCount > 0 {
		fmt.Printf("Обновлено %d документа(ов)\n", result.ModifiedCount)
	} else if result.UpsertedCount > 0 {
		fmt.Printf("Вставлен новый документ с _id: %v\n", result.UpsertedID)
	} else {
		fmt.Println("Никаких изменений не было внесено")
	}
}

func CheckUserExists(ctx context.Context, chatId int64, db *mongo.Database) bool {
	collection := db.Collection("user_achievement")
	filter := bson.M{"chat_id": chatId}

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		log.Fatalf("Ошибка при подсчёте документов: %v", err)
	}

	return count > 0
}

func CheckCountOfReq(ctx context.Context, chatId int64, db *mongo.Database) int64 {
	collection := db.Collection("user_achievement")
	filter := map[string]interface{}{"chat_id": chatId}
	var result struct {
		CountOfReq int64 `bson:"count_of_req"`
	}
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0
		}
		log.Fatalf("Ошибка при поиске документа: %v", err)
	}

	return result.CountOfReq
}

func CreateAchMap() map[int64]domain.Achievement {
	achs := map[int64]domain.Achievement{}

	ach1 := domain.Achievement{
		Title: "Пробник",
		Desc:  "Привет, друг! Будем часто общаться!",
		Grade: "Салага", // салага
	}
	achs[5] = ach1

	ach2 := domain.Achievement{
		Title: "Начинаешь познавать ТФЯ!",
		Desc:  "На правильном пути, стремишься познать!",
		Grade: "Прапорщик", // прапорщик
	}
	achs[10] = ach2

	ach3 := domain.Achievement{
		Title: "Тигр ТФЯ",
		Desc:  "Крутой тип, вероятность РК на максимум стремится к единице",
		Grade: "Старшина", // старшина
	}
	achs[20] = ach3

	ach4 := domain.Achievement{
		Title: "Охранный пёс в Переяславле",
		Desc:  "Тут без комментариев",
		Grade: "Сержант", // сержант
	}
	achs[50] = ach4

	ach5 := domain.Achievement{
		Title: "Антонина Николаевна – Королева Лекций",
		Desc:  "Прослушал все лекции Антонины Николаевны. не заскучал ни разу и хочешь разобраться ещё лучше!",
		Grade: "Лейтенант", // лейтенант
	}
	achs[100] = ach5

	ach6 := domain.Achievement{
		Title: "Исследователь ИПС РАН",
		Desc:  "Провёл день в лаборатории ИПС РАН вместе с Антониной Николаевной",
		Grade: "Капитан", // капитан
	}
	achs[200] = ach6

	ach7 := domain.Achievement{
		Title: "Переяславский выживальщик",
		Desc:  "Выжил в суровых условиях Переяславля Залесского и сохранил учебу.",
		Grade: "Майор", // майор
	}
	achs[500] = ach7

	ach8 := domain.Achievement{
		Title: "Легенда ИПС РАН и Переяславля",
		Desc:  "Синхронизировал гены с Антониной Николаевной, освоил все тайны ТФЯ и стал бессмертным студентом!",
		Grade: "Генерал", // генерал
	}
	achs[1000] = ach8

	return achs
}

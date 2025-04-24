package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"

	"aidicti.top/pkg/logging"
	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/pkg/utils"
	"aidicti.top/tlg/internal/model"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type privateKey struct {
	Key string `json:"private_key"`
}

type gateway struct {
	bot       *bot.Bot
	requests  chan Req
	responses chan Resp
}

// TODO what a fuck???
type ModifyData struct {
	Id   int64
	Data string
}

type UIObject struct {
	Id        uint64
	Text      string
	IsModied  bool
	MessageId uint64
}

type UICallback struct {
	Id        uint64
	MessageId uint64
}

type Req struct {
	pkgmodel.ReqData
	model.Message
	MData  *ModifyData
	UIData []UIObject
}

type Resp struct {
	pkgmodel.ReqData
	model.Message
	MData  *ModifyData
	UICall *UICallback
}

func (g *gateway) Requests() chan<- Req {
	return g.requests
}

func (g *gateway) Responses() <-chan Resp {
	return g.responses
}

func New() *gateway {
	const TelegramCredPathEnv = "TELEGRAM_APPLICATION_CREDENTIALS"
	value := os.Getenv(TelegramCredPathEnv)
	utils.Assert(value != "", "TELEGRAM_APPLICATION_CREDENTIALS env variable not set")

	file, err := os.Open(value)
	utils.Assert(err == nil, fmt.Sprintf("TELEGRAM_APPLICATION_CREDENTIALS path got read with an error [%s]", value))
	defer file.Close()

	var privateKey privateKey
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&privateKey)
	utils.Assert(err == nil, "TELEGRAM_APPLICATION_CREDENTIALS json file got read with an error")

	gtw := &gateway{bot: nil, requests: make(chan Req, 10), responses: make(chan Resp, 10)}

	opts := []bot.Option{
		bot.WithDefaultHandler(func(ctx context.Context, bot *bot.Bot, update *models.Update) {
			defaultHandler(gtw, ctx, bot, update)
		}),
		// bot.WithMessageTextHandler("/add", bot.MatchTypeExact, addHandler),
		// bot.WithMessageTextHandler("/practice", bot.MatchTypeExact, practiceHandler),
		// bot.WithDefaultHandler(defaultHandlerBtn),
		// bot.WithCallbackQueryDataHandler("button", bot.MatchTypePrefix, callbackHandlerBtn),
		// bot.WithDefaultHandler(handlerInline),
		// bot.WithMessageTextHandler("/select", bot.MatchTypeExact, commandHandler),
		bot.WithCallbackQueryDataHandler("button", bot.MatchTypePrefix,
			func(ctx context.Context, bot *bot.Bot, update *models.Update) {
				callbackHandler(gtw, ctx, bot, update)
			}),
	}

	gtw.bot, err = bot.New(privateKey.Key, opts...)
	utils.Assert(err == nil, fmt.Sprintf("telegramBot started with an error [%s]", err))

	return gtw
}

func (g *gateway) Run(ctx context.Context) {
	go g.bot.Start(ctx)

	go func() {
		for {
			select {
			case r := <-g.requests:

				if len(r.UIData) != 0 {

					markup := &models.InlineKeyboardMarkup{
						InlineKeyboard: [][]models.InlineKeyboardButton{
							{},
						},
					}

					for _, ui := range r.UIData {
						markup.InlineKeyboard = append(markup.InlineKeyboard,
							[]models.InlineKeyboardButton{
								{Text: ui.Text, CallbackData: fmt.Sprintf("button_%d", ui.Id)},
							})
					}

					if r.UIData[0].IsModied {
						_, err := g.bot.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
							ChatID:      r.DlgId,
							MessageID:   int(r.UIData[0].MessageId),
							ReplyMarkup: markup,
						})

						if err != nil {
							logging.Warn("error occured", "err", err)
						}

						break
					}

					_, err := g.bot.SendMessage(ctx, &bot.SendMessageParams{
						ChatID:      r.DlgId,
						Text:        r.Message.Message,
						ParseMode:   models.ParseModeHTML,
						ReplyMarkup: markup,
					})

					if err != nil {
						logging.Debug("error SendMessage", "err", err)
					}

					break

				}

				if r.Action != nil {

					if r.MData != nil {
						_, err := g.bot.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
							ChatID:      r.DlgId,
							MessageID:   int(r.MData.Id),
							ReplyMarkup: buildKeyboard(uint64(r.ReqData.ReqId), r.Action)})

						if err != nil {
							logging.Warn("error occured", "err", err)
						}

						break
					}

					_, err := g.bot.SendMessage(ctx, &bot.SendMessageParams{
						ChatID:      r.Id,
						Text:        r.Message.Message,
						ParseMode:   models.ParseModeHTML,
						ReplyMarkup: buildKeyboard(uint64(r.ReqData.ReqId), r.Action),
					})

					if err != nil {
						logging.Debug("error SendMessage", "err", err)
					}

					break
				}

				g.bot.SendMessage(ctx, &bot.SendMessageParams{
					ChatID:    r.Id,
					Text:      r.Message.Message,
					ParseMode: models.ParseModeHTML,
				})
			case <-ctx.Done():
				return
			}
		}
	}()

}

func defaultHandler(g *gateway, ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message.Voice != nil {
		audioData, err := getAudio(ctx, b, update.Message.Voice)

		if err != nil {
			logging.Info("error")
		}

		logging.Info("not error")

		g.responses <- Resp{
			pkgmodel.ReqData{},
			model.Message{model.UserID(update.Message.From.ID), &audioData, update.Message.Text, nil, []uint64{}},
			nil, nil,
		}
	} else {
		g.responses <- Resp{
			pkgmodel.ReqData{},
			model.Message{model.UserID(update.Message.From.ID), nil, update.Message.Text, nil, []uint64{}},
			nil, nil,
		}
	}
}

func getAudio(ctx context.Context, b *bot.Bot, voice *models.Voice) (model.AudioData, error) {
	utils.Assert(voice != nil, "models.Voice == nil")

	if voice.FileSize > int64(2*math.Pow(2, 20)) {
		logging.Info("file is too big")
		return model.AudioData{}, fmt.Errorf("file is too big")
	}

	// Get the file path of the voice message
	getFilePath := func(fileID string) (string, error) {
		url := fmt.Sprintf("https://api.telegram.org/bot%s/getFile?file_id=%s", b.Token(), fileID)
		resp, err := http.Get(url)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		var result struct {
			Ok     bool `json:"ok"`
			Result struct {
				FilePath string `json:"file_path"`
			} `json:"result"`
		}

		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil || !result.Ok {
			return "", fmt.Errorf("failed to get file path")
		}

		return result.Result.FilePath, nil
	}

	downloadFile := func(filePath string) ([]byte, error) {
		url := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", b.Token(), filePath)
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)

		return data, err
	}

	path, err := getFilePath(voice.FileID)
	if err != nil {
		logging.Info("reading data is finished with an error", err)
		return model.AudioData{}, err
	}

	data, err := downloadFile(path)
	if err != nil {
		logging.Info("reading data is finished with an error", err)
		return model.AudioData{}, err
	}

	logging.Info("model.AudioData is creating")
	return model.AudioData{Data: data}, nil
}

// func callbackHandlerBtn(ctx context.Context, b *bot.Bot, update *models.Update) {
// 	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
// 		CallbackQueryID: update.CallbackQuery.ID,
// 		ShowAlert:       false,
// 	})

// 	b.SendMessage(ctx, &bot.SendMessageParams{
// 		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
// 		Text:   "You selected the button: " + update.CallbackQuery.Data,
// 	})

// 	b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
// 		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
// 		MessageID:   update.CallbackQuery.Message.Message.ID,
// 		ReplyMarkup: buildKeyboard(),
// 	})

// }

// func defaultHandlerBtn(ctx context.Context, b *bot.Bot, update *models.Update) {
// 	kb := &models.InlineKeyboardMarkup{
// 		InlineKeyboard: [][]models.InlineKeyboardButton{
// 			{
// 				{Text: "Button 1", CallbackData: "button_1"},
// 				{Text: "Button 2", CallbackData: "button_2"},
// 			}, {
// 				{Text: "Button 3", CallbackData: "button_3"},
// 			},
// 			{
// 				{Text: "Button 4", CallbackData: "button_4"},
// 				{Text: "Button 5", CallbackData: "button_5"},
// 			},
// 		},
// 	}

// 	b.SendMessage(ctx, &bot.SendMessageParams{
// 		ChatID:      update.Message.Chat.ID,
// 		Text:        "Click by button",
// 		ReplyMarkup: kb,
// 	})
// }

// func handlerInline(ctx context.Context, b *bot.Bot, update *models.Update) {
// 	if update.InlineQuery == nil {
// 		return
// 	}

// 	results := []models.InlineQueryResult{
// 		&models.InlineQueryResultArticle{ID: "1", Title: "Foo 1", InputMessageContent: &models.InputTextMessageContent{MessageText: "foo 1"}},
// 		&models.InlineQueryResultArticle{ID: "2", Title: "Foo 2", InputMessageContent: &models.InputTextMessageContent{MessageText: "foo 2"}},
// 		&models.InlineQueryResultArticle{ID: "3", Title: "Foo 3", InputMessageContent: &models.InputTextMessageContent{MessageText: "foo 3"}},
// 	}

// 	b.AnswerInlineQuery(ctx, &bot.AnswerInlineQueryParams{
// 		InlineQueryID: update.InlineQuery.ID,
// 		Results:       results,
// 	})
// }

func callbackHandler(g *gateway, ctx context.Context, b *bot.Bot, update *models.Update) {
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	// g.responses <- Resp{
	// 	pkgmodel.ReqData{DlgId: pkgmodel.DlgID(update.CallbackQuery.Message.Message.Chat.ID)},
	// 	model.Message{},
	// 	&ModifyData{Id: int64(update.CallbackQuery.Message.Message.ID), Data: update.CallbackQuery.Data},
	// }

	id := utils.Must(strconv.Atoi(update.CallbackQuery.Data[7:]))

	g.responses <- Resp{
		pkgmodel.ReqData{DlgId: pkgmodel.DlgID(update.CallbackQuery.Message.Message.Chat.ID)},
		model.Message{},
		nil, &UICallback{Id: uint64(id), MessageId: uint64(update.CallbackQuery.Message.Message.ID)},
	}

	// b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
	// 	ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
	// 	MessageID:   update.CallbackQuery.Message.Message.ID,
	// 	ReplyMarkup: buildKeyboard(),
	// })
}

var currentOptions = []bool{false, false, false}

func buildKeyboard(id uint64, action *model.Action) models.ReplyMarkup {
	utils.Assert(action != nil, "")

	// id, err := uuid.NewUUID()
	// utils.Assert(err == nil, "UUID creating problem")

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: action.Message, CallbackData: fmt.Sprintf("button_%d", id)},
			},
		},
	}

	return kb
}

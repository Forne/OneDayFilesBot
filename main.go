package main

import (
	"encoding/json"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/ncw/swift"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"time"
)

var (
	bot         *tgbotapi.BotAPI
	config      Configuration
	sw          swift.Connection
	letterRunes = []rune("abcdefghkmnopqrstuvwxyz0123456789")
)

func init() {
	// Read configuration
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatalf("[Configuration] Reading error: %+v", err)
	} else {
		log.Println("[Configuration] Read successfully.")
	}
	// Decode configuration
	if err = json.Unmarshal(file, &config); err != nil {
		log.Fatalf("[Configuration] Decoding error: %+v", err)
	}

	// Initialize swift client.
	c := swift.Connection{
		UserName:    config.Swift.UserName,
		ApiKey:      config.Swift.ApiKey,
		AuthUrl:     config.Swift.AuthUrl,
		AuthVersion: config.Swift.AuthVersion,
	}
	err = c.Authenticate()
	if err != nil {
		log.Fatalf("[Swift] Initialize error: %+v", err)
	} else {
		log.Println("[Swift] Initialization successfully.")
		sw = c
	}

	// Initialize bot
	newBot, err := tgbotapi.NewBotAPI(config.Telegram.Token)
	if err != nil {
		log.Panicf("[Bot] Initialize error: %+v", err)
	} else {
		bot = newBot
		log.Printf("[Bot] Authorized as @%s", bot.Self.UserName)
	}
}

func main() {
	updates := make(<-chan tgbotapi.Update)

	// Get updates
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Printf("[Bot] Getting updates error: %+v", err)
	}

	for update := range updates {
		if update.Message != nil {
			go queryMsg(update.Message)
		}
	}
}

func queryMsg(u *tgbotapi.Message) {
	switch true {
	case u.Command() == "start":
		message := tgbotapi.NewMessage(u.Chat.ID, "Привет! Отправь мне любой файл, фотографию или видео и получи ссылку на загруженный файл.")
		if _, err := bot.Send(message); err != nil {
			log.Printf("[Bot] Sending message error: %+v", err)
		}
	case u.Document != nil:
		bot.Send(tgbotapi.NewChatAction(u.Chat.ID, tgbotapi.ChatTyping))
		tgfile, err := bot.GetFileDirectURL(u.Document.FileID)
		var msg string
		if err != nil {
			log.Printf("[Bot] File get error: %+v", err)
			msg = "Internal service error :("
		} else {
			msg = swUpload(tgfile)
		}
		message := tgbotapi.NewMessage(u.Chat.ID, msg)
		if _, err := bot.Send(message); err != nil {
			log.Printf("[Bot] Sending message error: %+v", err)
		}
	case u.Photo != nil:
		message := tgbotapi.NewMessage(u.Chat.ID, "Для загрузке фотографий, пожалуйста, не используйте сжатие! Я пока не умею работать с такими изображениями :(")
		if _, err := bot.Send(message); err != nil {
			log.Printf("[Bot] Sending message error: %+v", err)
		}
	case u.Audio != nil:
		bot.Send(tgbotapi.NewChatAction(u.Chat.ID, tgbotapi.ChatTyping))
		tgfile, err := bot.GetFileDirectURL(u.Audio.FileID)
		var msg string
		if err != nil {
			log.Printf("[Bot] Audio get error: %+v", err)
			msg = "Internal service error :("
		} else {
			msg = swUpload(tgfile)
		}
		message := tgbotapi.NewMessage(u.Chat.ID, msg)
		if _, err := bot.Send(message); err != nil {
			log.Printf("[Bot] Sending message error: %+v", err)
		}
	}
}

func swUpload(url string) string {
	var msg string
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		log.Printf("[HTTP] Download error: %+v", err)
		msg = "Internal service error :("
	} else {
		filebyte, err := ioutil.ReadAll(resp.Body)
		var extension = filepath.Ext(url)
		pathonswift := randStringRunes(8) + extension
		err = sw.ObjectPutBytes(config.Swift.Container, config.Swift.PathToFile+pathonswift, filebyte, "")
		if err != nil {
			log.Printf("[Swift] Upload error: %+v", err)
			msg = "Internal service error :("
		} else {
			msg = config.Swift.FrontendUrl + pathonswift
		}
	}
	return msg
}

func randStringRunes(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

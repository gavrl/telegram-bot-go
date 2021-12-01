package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type wallet map[string]float64

type bResponse struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price,string"`
}

var db = map[int]wallet{}

func main() {
	bot, err := tgbotapi.NewBotAPI("")
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		command := strings.Split(update.Message.Text, " ")
		userId := update.Message.From.ID

		switch command[0] {
		case "ADD":
			if len(command) != 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверные аргументы"))
				continue
			}

			_, err := getPrice(command[1])
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "неверная валюта"))
				continue
			}

			money, err := strconv.ParseFloat(command[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
			}

			if _, ok := db[userId]; !ok {
				db[userId] = make(wallet)
			}

			db[userId][command[1]] += money

		case "SUB":
			if len(command) != 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверные аргументы"))
				continue
			}

			money, err := strconv.ParseFloat(command[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				continue
			}

			if _, ok := db[userId]; !ok {
				db[userId] = make(wallet)
			}

			db[userId][command[1]] -= money
		case "DEL":
			if len(command) != 2 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверные аргументы"))
				continue
			}

			delete(db[userId], command[1])
		case "SHOW":
			resp := ""
			for key, value := range db[userId] {
				usdPrice, err := getPrice(key)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
					continue
				}
				resp += fmt.Sprintf("%s: %.2f rub\n", key, value*usdPrice)
			}
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, resp))
		default:
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда"))
		}
	}
}

func getPrice(symbol string) (float64, error) {
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sRUB", symbol)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}

	var bRes bResponse
	err = json.NewDecoder(resp.Body).Decode(&bRes)
	if err != nil {
		return 0, err
	}

	if bRes.Symbol == "" {
		return 0, errors.New("неверная валюта")
	}

	return bRes.Price, nil
}

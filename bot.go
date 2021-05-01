package goTelegram

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

func NewBot(s string) (Bot, error) {

	var newBot Bot

	newBot.APIURL = "https://api.telegram.org/bot" + s

	newBot.Keyboard = keyboard{Keyboard: [][]inlineKeyboard{}}

	resp, err := http.Get(newBot.APIURL + "/getMe")

	if err != nil {
		log.Println("Fetch Bot Details Failed, Check Internet Connection")
		log.Println(err)
		return newBot, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Invalid Token Provided")
		return newBot, errors.New("Invalid Bot Token Provided")
	}

	err = json.NewDecoder(resp.Body).Decode(&newBot)

	if err != nil {
		log.Println("Couldn't Marshal Response")
		log.Println(err)
		return newBot, err
	}

	return newBot, nil

	//return Bot{APIURL: "https://api.telegram.org/bot" + s, Keyboard: keyboard{Keyboard: [][]inlineKeyboard{}}}
}

func (b *Bot) AddButton(text, callback string) {
	b.Keyboard.Buttons = append(b.Keyboard.Buttons, inlineKeyboard{Text: text, Data: callback})
}

func (b *Bot) MakeKeyboardRow() {
	b.Keyboard.Keyboard = append(b.Keyboard.Keyboard, b.Keyboard.Buttons)
	b.Keyboard.Buttons = nil
}

func (b *Bot) DeleteKeyboard() {
	b.Keyboard.Keyboard = nil
	b.Keyboard.Buttons = nil
}

func (b *Bot) SetHandler(fn interface{}) {
	b.HandlerSet = false
	b.Handler = reflect.ValueOf(fn)
	if b.Handler.Kind() != reflect.Func {
		log.Println("Argument Is Not Of Type Function")
		return
	}

	b.HandlerSet = true
}

func (b *Bot) UpdateHandler(w http.ResponseWriter, r *http.Request) {

	if b.HandlerSet {

		var update Update

		err := json.NewDecoder(r.Body).Decode(&update)

		if err != nil {
			log.Println("Couldn't Parse Incoming Message")
			return
		}

		text := strings.Fields(update.Message.Text)

		if len(text) == 0 {
			return
		}

		if strings.HasPrefix(text[0], "/") {

			update.Command = text[0]

			if strings.HasSuffix(text[0], b.Me.Username) {
				update.Command = strings.Split(text[0], "@")[0]
			}
		}

		rarg := make([]reflect.Value, 1)

		rarg[0] = reflect.ValueOf(update)

		go b.Handler.Call(rarg)
	} else {
		log.Println("Please Set A Function To Be Called Upon New Updates")
		return
	}

}

func (b *Bot) AnswerCallback(callbackID string) {
	link := b.APIURL + "/answerCallbackQuery"

	answer := answerCallback{
		ID: callbackID,
	}

	jsonBody, err := json.Marshal(answer)

	if err != nil {
		log.Println(err)
		return
	}

	resp, err := http.Post(link, "application/json", bytes.NewBuffer(jsonBody))

	if err != nil {
		log.Println("Couldn't Answer CallBack Successfully, Check Internet Source")
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Couldn't Answer CallBack Successfully, Status Code Not OK")
	}
}

func (b *Bot) SendMessage(s string, c chat) {

	link := b.APIURL + "/sendMessage"

	reply := replyBody{
		ChatID: strconv.Itoa(c.ID),
		Text:   s,
	}

	if len(b.Keyboard.Keyboard) > 0 {
		reply.ReplyMarkup.InlineKeyboard = b.Keyboard.Keyboard
	}

	jsonBody, err := json.Marshal(reply)

	if err != nil {
		log.Println("Couldn't Marshal Response")
		return
	}

	resp, err := http.Post(link, "application/json", bytes.NewBuffer(jsonBody))

	if err != nil {
		log.Println("Couldn't Make Request Successfully, Please Check Internet Source")
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Message Wasn't Sent Successfully, Please Try Again")
	}
}

func (b *Bot) EditMessage(message message, text string) {
	link := b.APIURL + "/editMessageText"

	updatedText := editBody{
		ChatID:    strconv.Itoa(message.Chat.ID),
		MessageID: message.MessageID,
		Text:      text,
	}

	if len(b.Keyboard.Keyboard) > 0 {
		updatedText.ReplyMarkup.InlineKeyboard = b.Keyboard.Keyboard
	}

	jsonBody, err := json.Marshal(updatedText)

	if err != nil {
		log.Println("There Was An Error Marshalling The Message")
		log.Println(err)
		return
	}

	resp, err := http.Post(link, "application/json", bytes.NewBuffer(jsonBody))

	if err != nil {
		log.Println("Couldn't Communicate With Telegram Servers, Please Check Internet Source")
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Message Wasn't Edited Successfully, Please Try Again")
		log.Println(err)
	}
}

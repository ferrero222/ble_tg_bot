package internble

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type bleInlineButton struct {
	Title string
	Data  string
}

// bleMainKeyboardHandler /*******************************************************************************
func bleMainKeyboardHandler(update tgbotapi.Update, bot *tgbotapi.BotAPI, command string) {
	mainMenuMap := map[string]func(update tgbotapi.Update, bot *tgbotapi.BotAPI){
		"start":    bleStartCmd,
		"scan":     bleScanCmd,
		"stopScan": bleStopScanCmd,
		"exit":     bleExitCmd,
	}
	if handler, exists := mainMenuMap[command]; exists {
		handler(update, bot)
	}
}

// bleKeyboardConfig /*******************************************************************************
func bleInlineKeyboardHandler(update tgbotapi.Update, bot *tgbotapi.BotAPI, command string) {
	inlineMenu := map[string]func(update tgbotapi.Update, bot *tgbotapi.BotAPI){
		"connect":    bleConnectSubCmd,
		"disconnect": bleDisconnectSubCmd,
		//------------------------------------------1lvl menu
		"getServices": bleGetServicesSubCmd,
		/*		//"disconnect":   bleDisconnectSubCmd,
				//------------------------------------------2lvl menu
				"prevService":  blePrevServiceCmd,
				"getChars":     bleGetCharCmd,
				"backToSecond": bleBackToSecond,
				"nextService":  bleNextService,
				//------------------------------------------3lvl menu
				"prevChar":     blePrevCharCmd,
				"readChar":     bleReadCharCmd,
				"writeChar":    bleWriteCharCmd,
				"backToThird":  bleBackToThird,
				"notification": bleNotification,
				"nextChar":     bleNextChar,*/
	}
	if handler, exists := inlineMenu[command]; exists {
		handler(update, bot)
	}
}

// bleKeyboardConfig /*******************************************************************************
func bleKeyboardConfig(statusButton string, buttons []string) tgbotapi.ReplyKeyboardMarkup {
	var keyboard [][]tgbotapi.KeyboardButton
	keyboard = append(keyboard, []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton("status: " + statusButton)})
	var row []tgbotapi.KeyboardButton
	for _, button := range buttons {
		row = append(row, tgbotapi.NewKeyboardButton(button))
	}
	if len(row) > 0 {
		keyboard = append(keyboard, row)
	}
	return tgbotapi.NewReplyKeyboard(keyboard...)
}

// bleInlineKeyboardConfig /*******************************************************************************
func bleInlineKeyboardConfig(buttons []bleInlineButton) tgbotapi.InlineKeyboardMarkup {
	var inlineKeyboard [][]tgbotapi.InlineKeyboardButton
	for _, btn := range buttons {
		inlineKeyboard = append(inlineKeyboard, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(btn.Title, btn.Data),
		})
	}
	return tgbotapi.NewInlineKeyboardMarkup(inlineKeyboard...)
}

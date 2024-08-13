package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

/*******************************************************************************
 ** \brief  parsing and request proc
 ** \param  ...
 ** \retval ...
 ******************************************************************************/
func ble_start_cmd(update tgbotapi.Update) {
	if global_status == not_ready_status {
		chat_id = update.Message.Chat.ID
		err := adapter.Enable()
		if err != nil {
			log.Printf("error: BLE adapter issue")
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "BOT: Error trying to enable BLE adapter")
			set_status_menu(not_ready_status, &msg)
			_, err = bot.Send(msg)
		} else {
			log.Printf("BLE adapter started")
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "BOT: BLE adapter started successfully")
			set_status_menu(ready_status, &msg)
			_, err = bot.Send(msg)
		}
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "BOT: Already started")
		bot.Send(msg)
	}
}

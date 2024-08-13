package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"tinygo.org/x/bluetooth"
)

var err error

var user_struct struct{
	adapter       = bluetooth.DefaultAdapter
	advertiser    = bluetooth.Advertisement{}
	bot           *(tgbotapi.BotAPI)
	chat_id       int64
	scaned_amount int
	scaned_id     int
	global_status = not_ready_status
}

var user_scan_message [40]struct {
	scaned_address bluetooth.Address
	message_id     int
}

const (
	not_ready_status   = "not ready"
	ready_status       = "ready"
	connected_status   = "connected"
	advertising_status = "advertising"
	scanning_status    = "scanning"
)

/*******************************************************************************
 ** \brief  parsing and request proc
 ** \param  ...
 ** \retval ...
 ******************************************************************************/
func ble_init() {
	bot, err = tgbotapi.NewBotAPI("7462877109:AAHmfC6iMYwvUVDPZeV-IslXfP5HRHr3JBw")
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	//----------------------------------------------------------------------------------------
	for cur_update := range updates {
		if cur_update.Message != nil {
			switch cur_update.Message.Text {
			case "Start":
				ble_start_cmd(cur_update)
			case "Scan":
				ble_scan_cmd(cur_update)
			case "Advert":
				ble_adver_cmd(cur_update)
			case "Stop":
				ble_stop_cmd(cur_update)
			case "Exit":
				ble_exit_cmd(cur_update)
			}
			if cur_update.CallbackQuery != nil {
				data := update.CallbackQuery.Data
				switch data {
				case "Connect":
					ble_connect_subcmd(cur_update)
				case "Disconnect":
					ble_disconnect_subcmd(cur_update)
				case "Get Chars":
					ble_get_chars_subcmd(cur_update)
				case "Read Char":
					ble_read_char_subcmd(cur_update)
				case "Write Char":
					ble_write_char_subcmd(cur_update)
				}
			}
		}
	}
}

/*******************************************************************************
 ** \brief  parsing and request proc
 ** \param  ...
 ** \retval ...
 ******************************************************************************/
func set_status_menu(status_button_temp string, msg *tgbotapi.MessageConfig) {
	global_status = status_button_temp
	var menu_keyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Status: "+global_status),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Scan"),
			tgbotapi.NewKeyboardButton("Advert"),
			tgbotapi.NewKeyboardButton("Stop"),
			tgbotapi.NewKeyboardButton("Exit"),
		),
	)
	msg.ReplyMarkup = menu_keyboard
}

/*******************************************************************************
 ** \brief  parsing and request proc
 ** \param  ...
 ** \retval ...
 ******************************************************************************/
func onScan(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
	var message_temp tgbotapi.Message
	var elem_inline_keyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Connect", "connect"),
		),
	)
	if scaned_amount != 0 {
		temp := scaned_amount - 1
		for temp >= 0 {
			if user_scan_message[temp].scaned_address == device.Address {
				return
			}
			temp--
		}
	}
	log.Printf("found new BLE device %d", scaned_id)
	deviceInfo := fmt.Sprintf("BOT found device: \n<b>Number</b>: %d\n<b>MAC</b>: %s\n<b>RSSI</b>: %d\n<b>Local Name</b>: %s\n<b>Advertisement Payload</b>: ", scaned_id, device.Address.String(), device.RSSI, device.LocalName())
	msg := tgbotapi.NewMessage(chat_id, deviceInfo)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = elem_inline_keyboard
	message_temp, err = bot.Send(msg)
	user_scan_message[scaned_amount].scaned_address = device.Address
	user_scan_message[scaned_amount].message_id = message_temp.MessageID
	scaned_amount++
	if scaned_amount >= 40 {
		scaned_amount = 0
	}
	if err != nil {
		log.Printf("Cant send scanned BLE device: %s, %s", device.Address.String(), device.LocalName())
	}
	scaned_id++
}

/*******************************************************************************
 ** \brief  parsing and request proc
 ** \param  ...
 ** \retval ...
 ******************************************************************************/
func ble_check_chat_id(update tgbotapi.Update) int {
	if chat_id != update.Message.Chat.ID && chat_id != 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "BOT: Server is busy for another client, wait and try again later")
		bot.Send(msg)
		return 0
	} else {
		return 1
	}
}

/*******************************************************************************
 ** \brief  parsing and request proc
 ** \param  ...
 ** \retval ...
 ******************************************************************************/
func ble_whole_stop(update tgbotapi.Update) int {
	if chat_id != update.Message.Chat.ID && chat_id != 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "BOT: Server is busy for another client, wait and try again later")
		bot.Send(msg)
		return 0
	} else {
		return 1
	}
}
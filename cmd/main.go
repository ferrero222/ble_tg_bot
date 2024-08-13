package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"tinygo.org/x/bluetooth"
)

var err error

var user_struct struct{
	adapter       bluetooth.DefaultAdapter
	advertiser    bluetooth.Advertisement{}
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


/*******************************************************************************
 ** \brief  parsing and request proc
 ** \param  ...
 ** \retval ...
 ******************************************************************************/
func main() {
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
	for update := range updates {
		if update.Message != nil {
			switch update.Message.Text {
			case "Start":
				if chat_id != update.Message.Chat.ID && chat_id != 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "BOT: Server is busy for another client, wait and try again later")
					bot.Send(msg)
				} else {
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
			//----------------------------------------------------------------------------------------
			case "Scan":
				if chat_id != update.Message.Chat.ID && chat_id != 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "BOT: Server is busy for another client, wait and try again later")
					bot.Send(msg)
				} else if chat_id == update.Message.Chat.ID {
					if global_status == ready_status || global_status == advertising_status {
						if chat_id != update.Message.Chat.ID {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "BOT: Server is busy for another client, wait and try again later")
							bot.Send(msg)
						} else {
							err = advertiser.Stop()
							scaned_amount = 0
							scaned_id = 0
							go adapter.Scan(onScan)
							if err != nil {
								log.Printf("error: cant start scanning")
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, "BOT: Failed to register scan callback")
								set_status_menu(ready_status, &msg)
								bot.Send(msg)
							} else {
								log.Printf("start scanning")
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, "BOT: Start scanning.......")
								set_status_menu(scanning_status, &msg)
								bot.Send(msg)
							}
						}
					} else {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "BOT: BLE not started/Scan already started")
						bot.Send(msg)
					}
				}
			//----------------------------------------------------------------------------------------
			case "Advert":
				if chat_id != update.Message.Chat.ID && chat_id != 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "BOT: Server is busy for another client, wait and try again later")
					bot.Send(msg)
				} else if chat_id == update.Message.Chat.ID {
					if global_status == ready_status || global_status == scanning_status {
						if chat_id != update.Message.Chat.ID {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "BOT: Server is busy for another client, wait and try again later")
							bot.Send(msg)
						} else {
							log.Printf("start advertising")
							adapter.StopScan()
							go advertiser.Start()
							scaned_amount = 0
							scaned_id = 0
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "BOT: Start advertising......")
							set_status_menu(advertising_status, &msg)
							bot.Send(msg)
						}
					} else {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "BOT: BLE not started/Advert already started")
						bot.Send(msg)
					}
				}
			//----------------------------------------------------------------------------------------
			case "Stop":
				if chat_id != update.Message.Chat.ID && chat_id != 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "BOT: Server is busy for another client, wait and try again later")
					bot.Send(msg)
				} else if chat_id == update.Message.Chat.ID {
					if global_status != ready_status {
						log.Printf("stoped whole process")
						adapter.StopScan()
						advertiser.Stop()
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "BOT: BLE proc was stoped")
						set_status_menu(ready_status, &msg)
						bot.Send(msg)
					} else {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "BOT: No process to stop")
						bot.Send(msg)
					}
				}
				//----------------------------------------------------------------------------------------
			case "Exit":
				if chat_id != update.Message.Chat.ID && chat_id != 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "BOT: Server is busy for another client, wait and try again later")
					bot.Send(msg)
				} else if chat_id == update.Message.Chat.ID {
					if global_status != not_ready_status {
						log.Printf("bot stoped")
						global_status = not_ready_status
						chat_id = 0
						adapter.StopScan()
						advertiser.Stop()
						scaned_amount = 0
						scaned_id = 0
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "BOT: Bot stoped")
						msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
						bot.Send(msg)
					}
				}
			}
		}
		//----------------------------------------------------------------------------------------
		//----------------------------------------------------------------------------------------
		if update.CallbackQuery != nil {
			data := update.CallbackQuery.Data
			switch data {
			case "connect":
				for _, v := range user_scan_message {
					if scaned_amount != 0 {
						if v.message_id == update.CallbackQuery.Message.MessageID {
							adapter.StopScan()
							advertiser.Stop()
							scaned_amount = 0
							scaned_id = 0
							var params bluetooth.ConnectionParams
							_, err = adapter.Connect(v.scaned_address, params)
							if err != nil {
								log.Printf("Cant connect to device: %s", v.scaned_address)
							}
							msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "BOT: Connected succesful")
							set_status_menu(connected_status+" to "+v.scaned_address.String(), &msg)
							bot.Send(msg)
							break
						}
					}
				}
			}
		}
	}
}

package internble

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
	"tinygo.org/x/bluetooth"
)

// bleStartCmd /*******************************************************************************
func bleStartCmd(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	var err error
	bleChatId = update.Message.Chat.ID
	var adapter bluetooth.Adapter
	err = adapter.Enable()
	if err != nil {
		log.Printf("ERROR: Failed to start ble adapter")
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, bleMessageHeader+"FAILED TO START BLE ADAPTER")
		bleStatus = bleNotReadyStatus
		setStatusMenu(bleStatus, &msg)
		_, err = bot.Send(msg)
		if err != nil {
			log.Printf("ERROR %v: Cant send message about failed to start ble adapter", err)
		}
	} else {
		bleTimerProc(update, bot)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, bleMessageHeader+"BLE ADAPTER STARTED SUCCESSFULLY")
		bleStatus = bleReadyStatus
		setStatusMenu(bleStatus, &msg)
		_, err = bot.Send(msg)
		if err != nil {
			log.Printf("ERROR %v: Cant send message about successfully started adapter", err)
		}
	}
}

// bleScanCmd /*******************************************************************************
func bleScanCmd(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	var err error
	bleScannedAmount = 0
	go bleAdapter.Scan(onScan)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, bleMessageHeader+"START SCANNING")
	bleStatus = bleScanningStatus
	setStatusMenu(bleStatus, &msg)
	_, err = bot.Send(msg)
	if err != nil {
		log.Printf("ERROR %v: Cant send message about successfully started scanning", err)
	}
}

// bleAdvertCmd /*******************************************************************************
func bleAdvertCmd(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	var err error
	go bleAdvertiser.Start()
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, bleMessageHeader+"START ADVERTISING")
	bleStatus = bleAdvertisingStatus
	setStatusMenu(bleStatus, &msg)
	_, err = bot.Send(msg)
	if err != nil {
		log.Printf("ERROR %v: Cant send message about successfully started advertising", err)
	}
}

// bleStopCmd /*******************************************************************************
func bleStopCmd(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	var err error
	err = bleAdapter.StopScan()
	if err != nil {
		log.Printf("ERROR %v: Cant stop scanning", err)
	}
	err = bleAdvertiser.Stop()
	if err != nil {
		log.Printf("ERROR %v: Cant stop advert", err)
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, bleMessageHeader+"BLE PROCESS WAS STOPPED")
	bleStatus = bleReadyStatus
	setStatusMenu(bleStatus, &msg)
	_, err = bot.Send(msg)
	if err != nil {
		log.Printf("ERROR %v: Cant send message about successfully stoped proccess", err)
	}
}

// bleExitCmd /*******************************************************************************
func bleExitCmd(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	var err error
	bleChatId = 0
	err = bleAdapter.StopScan()
	if err != nil {
		log.Printf("ERROR %v: Cant stop scanning", err)
	}
	bleTimerStop()
	bleScannedAmount = 0
	bleStatus = bleNotReadyStatus
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, bleMessageHeader+"BOT STOPPED")
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, err = bot.Send(msg)
	if err != nil {
		log.Printf("ERROR %v: Cant send message about successfully stoped bot", err)
	}
}

// bleConnectSubCmd /*******************************************************************************
func bleConnectSubCmd(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	var err error
	for _, v := range bleScanBotList {
		if bleScannedAmount != 0 {
			if v.messageId == update.CallbackQuery.Message.MessageID {
				var params bluetooth.ConnectionParams
				var elemInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Characteristics", "getChars"),
						tgbotapi.NewInlineKeyboardButtonData("Disconnect", "disconnect"),
					),
				)
				//----------------------------------------------------------------------------------------
				bleConnDevice, err = bleAdapter.Connect(v.scannedAddress, params)
				if err != nil {
					log.Printf("ERROR %v: Cant connect to device: %s", err, v.scannedAddress)
					lines := strings.Split(update.CallbackQuery.Message.Text, "\n")
					lines[0] = "Status: connection failed\n"
					msg1 := tgbotapi.NewEditMessageText(
						update.CallbackQuery.Message.Chat.ID,
						update.CallbackQuery.Message.MessageID,
						strings.Join(lines, "\n"),
					)
					_, err = bot.Send(msg1)
					if err != nil {
						log.Printf("ERROR %v: Cant send message about failed connection", err)
					}
					return
				}
				//----------------------------------------------------------------------------------------
				lines := strings.Split(update.CallbackQuery.Message.Text, "\n")
				lines[0] = "Status: connected"
				msg2 := tgbotapi.NewEditMessageTextAndMarkup(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					strings.Join(lines, "\n"),
					elemInlineKeyboard,
				)
				_, err = bot.Send(msg2)
				if err != nil {
					log.Printf("ERROR %v: Cant send message about successefull connection", err)
				}
				//----------------------------------------------------------------------------------------
				msg3 := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, bleMessageHeader+"STATUS UPDATE")
				bleStatus = bleConnectedStatus
				setStatusMenu(bleStatus+" to "+v.scannedAddress.String(), &msg3)
				_, err = bot.Send(msg3)
				if err != nil {
					log.Printf("ERROR %v: Cant send message about status updating", err)
				}
				break
			}
		}
	}
}

// bleDisconnectSubCmd /*******************************************************************************
func bleDisconnectSubCmd(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	var err error
	var elemInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Connect", "connect"),
		),
	)
	//----------------------------------------------------------------------------------------
	err = bleConnDevice.Disconnect()
	if err != nil {
		log.Printf("ERROR %v: Cant disconnect", err)
		return
	}
	lines := strings.Split(update.CallbackQuery.Message.Text, "\n")
	lines[0] = "Status: not connected"
	msg1 := tgbotapi.NewEditMessageTextAndMarkup(
		update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		strings.Join(lines, "\n"),
		elemInlineKeyboard,
	)
	_, err = bot.Send(msg1)
	if err != nil {
		log.Printf("ERROR %v: Cant send message about successfully disconnect", err)
	}
	//----------------------------------------------------------------------------------------
	msg2 := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, bleMessageHeader+"STATUS UPDATE")
	bleStatus = bleReadyStatus
	setStatusMenu(bleStatus, &msg2)
	_, err = bot.Send(msg2)
	if err != nil {
		log.Printf("ERROR %v: Cant send message about status updating", err)
	}
}

// bleGetCharsSubCmd /*******************************************************************************
func bleGetCharsSubCmd(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	/*	var elemInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Read", "/read"),
				tgbotapi.NewInlineKeyboardButtonData("Write", "/write"),
				tgbotapi.NewInlineKeyboardButtonData("Disconnect", "/disconnect"),
			),
		)
		charList, err := bleConnDevice.DiscoverServices(nil)
		if err != nil {
			log.Printf("ERROR %v: Cant send message about status updating", err)
		}
		var charString string
		for _, v := range charList {

		}
		msg1 := tgbotapi.NewEditMessageTextAndMarkup(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			update.CallbackQuery.Message.Text,
			elemInlineKeyboard,
		)
		_, err = bot.Send(msg1)
		if err != nil {
			log.Printf("ERROR %v: Cant send message about successfully disconnect", err)
		}*/
	//----------------------------------------------------------------------------------------

}

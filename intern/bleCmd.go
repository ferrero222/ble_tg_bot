package internble

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"tinygo.org/x/bluetooth"
)

// BleBotInit ************************************************
// Brief:   None
// Param:   None
// Return:  None
// ***********************************************************
func bleStartCmd(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	var adapter bluetooth.Adapter
	var err error
	if bleChatId != 0 {
		log.Printf("ERROR: BLE adater already started")
		bleMessageSend(update, bot, "", nil, "ALREADY STARTED")
		return
	}
	//--------------------------------------------------------
	bleChatId = update.Message.Chat.ID
	err = adapter.Enable()
	if err != nil {
		log.Printf("ERROR: Failed to start ble adapter")
		bleMessageSend(update, bot, bleNotReadyStatus, []string{"/exit"}, "FAILED TO START BLE ADAPTER")
	} else {
		bleMessageSend(update, bot, bleReadyStatus, []string{"/scan", "/exit"}, "BLE ADAPTER STARTED SUCCESSFULLY")
	}
}

// BleBotInit ************************************************
// Brief:   None
// Param:   None
// Return:  None
// ***********************************************************
func bleScanCmd(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if ble.scan.scanStatus {
		bleMessageSend(update, bot, "", nil, "BLE SCANNING ALREADY STARTED")
		return
	}
	ble.scan.bleScannedAmount = 0
	ble.scan.scanList = nil
	go bleAdapter.Scan(bleOnScan)
	ble.scan.scanStatus = true
	bleMessageSend(update, bot, bleScanningStatus, []string{"/stopScan", "/exit"}, "START SCANNING")
}

// BleBotInit ************************************************
// Brief:   None
// Param:   None
// Return:  None
// ***********************************************************
func bleStopScanCmd(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	var err error
	if !ble.scan.scanStatus {
		bleMessageSend(update, bot, "", nil, "BLE SCANNING ALREADY STOPPED")
		return
	}
	err = bleAdapter.StopScan()
	if err != nil {
		bleMessageSend(update, bot, "", nil, "CANT STOP SCANNING")
		return
	}
	ble.scan.scanStatus = false
	bleMessageSend(update, bot, bleReadyStatus, []string{"/scan", "/exit"}, "BLE SCANNING WAS STOPPED")
}

// BleBotInit ************************************************
// Brief:   None
// Param:   None
// Return:  None
// ***********************************************************

func bleExitCmd(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	var err error
	if ble.scan.scanStatus {
		err = bleAdapter.StopScan()
		if err != nil {
			bleMessageSend(update, bot, "", nil, "CANT STOP SCANNING")
			return
		} else {
			ble.scan.scanStatus = false
			ble.scan.bleScannedAmount = 0
			ble.scan.scanList = nil
		}
	}
	//-------------------------------------------------------
	if ble.connect.connStatus {
		err = ble.connect.device.Disconnect()
		if err != nil {
			bleMessageSend(update, bot, "", nil, "CANT DISCONNECT")
			return
		} else {
			inlineButtons := []bleInlineButton{{Title: "Connect", Data: "connect"}}
			msg := tgbotapi.NewEditMessageTextAndMarkup(
				ble.connect.message.Chat.ID,
				ble.connect.message.MessageID,
				bleMessageEdit(ble.connect.message.Text, "Status: not connected", bleMsgDescrFormatStatus),
				bleInlineKeyboardConfig(inlineButtons),
			)
			_, err = bot.Send(msg)
			if err != nil {
				log.Printf("ERROR %v: Cant send message about successfully disconnect", err)
			}
			ble.connect.connStatus = false
		}
	}
	//-------------------------------------------------------
	bleChatId = 0
	bleTimerStop()
	bleMessageSend(update, bot, "", []string{"REMOVE"}, "BOT STOPPED")
}

// BleBotInit ************************************************
// Brief:   None
// Param:   None
// Return:  None
// ***********************************************************
func bleConnectSubCmd(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	var err error
	if ble.connect.connStatus {
		bleMessageSend(update, bot, "", nil, "BLE ALREADY CONNECTED TO SOME DEVICE")
		return
	}
	if ble.scan.bleScannedAmount != 0 {
		for _, v := range ble.scan.scanList {
			if v.message.MessageID == update.CallbackQuery.Message.MessageID {
				var params bluetooth.ConnectionParams
				//--------------------------------------------
				ble.connect.device, err = bleAdapter.Connect(v.device.Address, params)
				if err != nil {
					ble.connect.connStatus = false
					log.Printf("ERROR %v: Cant connect to device: %s", err, v.device.Address)
					msg := tgbotapi.NewEditMessageText(
						update.CallbackQuery.Message.Chat.ID,
						update.CallbackQuery.Message.MessageID,
						bleMessageEdit(ble.connect.message.Text, "Status: connection failed", bleMsgDescrFormatStatus),
					)
					_, err = bot.Send(msg)
					if err != nil {
						log.Printf("ERROR %v: Cant send message about failed connection", err)
					}
					return
				}
				//----------------------------------------------
				inlineButtons := []bleInlineButton{{Title: "Services", Data: "getServices"}, {Title: "Disconnect", Data: "disconnect"}}
				msg1 := tgbotapi.NewEditMessageTextAndMarkup(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					bleMessageEdit(ble.connect.message.Text, "Status: connected", bleMsgDescrFormatStatus),
					bleInlineKeyboardConfig(inlineButtons),
				)
				ble.connect.connStatus = true
				ble.connect.message = *(update.CallbackQuery.Message)
				_, err = bot.Send(msg1)
				if err != nil {
					log.Printf("ERROR %v: Cant send message about successefull connection", err)
				}
				//----------------------------------------------
				msg2 := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, bleMessageHeader+"STATUS UPDATE")
				keyboard := []string{"/scan", "/exit"}
				msg2.ReplyMarkup = bleKeyboardConfig(bleConnectedStatus+" to "+v.device.Address.String(), keyboard)
				_, err = bot.Send(msg2)
				if err != nil {
					log.Printf("ERROR %v: Cant send message about status updating", err)
				}
				break
			}
		}
	}
}

// BleBotInit ************************************************
// Brief:   None
// Param:   None
// Return:  None
// ***********************************************************
func bleDisconnectSubCmd(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	var err error
	//--------------------------------------------------------
	if !ble.connect.connStatus {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, bleMessageHeader+"BLE ALREADY NOT CONNECTED")
		_, err = bot.Send(msg)
		if err != nil {
			log.Printf("ERROR %v: Cant send message about already not connected", err)
		}
		return
	}
	ble.connect.connStatus = false
	err = ble.connect.device.Disconnect()
	if err != nil {
		log.Printf("ERROR %v: Cant disconnect", err)
		return
	}
	inlineButtons := []bleInlineButton{{Title: "Connect", Data: "connect"}}
	msg1 := tgbotapi.NewEditMessageTextAndMarkup(
		update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		bleMessageEdit(ble.connect.message.Text, "Status: not connected", bleMsgDescrFormatStatus),
		bleInlineKeyboardConfig(inlineButtons),
	)
	_, err = bot.Send(msg1)
	if err != nil {
		log.Printf("ERROR %v: Cant send message about successfully disconnect", err)
	}
	//---------------------------------------------------------
	msg2 := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, bleMessageHeader+"STATUS UPDATE")
	keyboard := []string{"/scan", "/exit"}
	msg2.ReplyMarkup = bleKeyboardConfig(bleReadyStatus, keyboard)
	_, err = bot.Send(msg2)
	if err != nil {
		log.Printf("ERROR %v: Cant send message about status updating", err)
	}
}

// BleBotInit ************************************************
// Brief:   None
// Param:   None
// Return:  None
// ***********************************************************
func bleGetServicesSubCmd(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	/*	var err error
		//----------------------------------------------------
		serviceList, err := ble.connect.device.DiscoverServices(nil)
		var sb strings.Builder
		sb.WriteString("Data:\n")
		for _, service := range serviceList {
			sb.WriteString(fmt.Sprintf("UUID: %s\n", service.UUID))
		}
		charList, err := serviceList[1].DiscoverCharacteristics(nil)

		lines := strings.Split(update.CallbackQuery.Message.Text, "\n")
		lines[0] = "Status: not connected"
		inlineButtons := []bleInlineButton{{Title: "Connect", Data: "connect"}}
		msg1 := tgbotapi.NewEditMessageTextAndMarkup(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			strings.Join(lines, "\n"),
			bleInlineKeyboardConfig(inlineButtons),
		)
		_, err = bot.Send(msg1)
		if err != nil {
			log.Printf("ERROR %v: Cant send message about successfully disconnect", err)
		}*/
	//--------------------------------------------------------
}

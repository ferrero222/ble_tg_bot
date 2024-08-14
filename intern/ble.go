package internble

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"time"
	"tinygo.org/x/bluetooth"
)

var (
	bleConnDevice    bluetooth.Device
	bleTimer         *time.Timer
	bleBot           *tgbotapi.BotAPI
	bleAdapter       bluetooth.Adapter
	bleAdvertiser    bluetooth.Advertisement
	bleChatId        int64
	bleScannedAmount int
	bleStatus        string
)

var bleScanBotList [40]struct {
	scannedAddress bluetooth.Address
	messageId      int
}

const (
	bleBotToken = "7328834082:AAGm8ygR2gr9oTlsdSUuULDwkHQWsUW-2G0"
)

const (
	bleMessageHeader = "BOT:"
)

const (
	bleNotReadyStatus    = "not ready"
	bleReadyStatus       = "ready"
	bleConnectedStatus   = "connected"
	bleAdvertisingStatus = "advertising"
	bleScanningStatus    = "scanning"
)

// BleBotInit /*******************************************************************************
func BleBotInit() {
	var err error
	bleBot, err = tgbotapi.NewBotAPI(bleBotToken)
	if err != nil {
		log.Panicf("PANIC %v: Cant init bot token", err)
	}
	bleBot.Debug = false
}

// BleUpdate /*******************************************************************************
func BleUpdate() {
	updatesConfig := tgbotapi.NewUpdate(0)
	updatesConfig.Timeout = 60
	updates := bleBot.GetUpdatesChan(updatesConfig)
	//----------------------------------------------------------------------------------------
	for v := range updates {
		if v.Message != nil {
			if bleCheckChatId(v.Message, bleBot) != 0 {
				bleDeleteUserMessage(v.Message.MessageID, bleBot)
				switch v.Message.Command() {
				case "start":
					bleStartCmd(v, bleBot)
				case "scan":
					bleScanCmd(v, bleBot)
				case "advert":
					bleAdvertCmd(v, bleBot)
				case "stop":
					bleStopCmd(v, bleBot)
				case "exit":
					bleExitCmd(v, bleBot)
				}
			}
		}
		if v.CallbackQuery != nil {
			if bleCheckChatId(v.CallbackQuery.Message, bleBot) != 0 {
				data := v.CallbackQuery.Data
				switch data {
				case "connect":
					bleConnectSubCmd(v, bleBot)
				case "disconnect":
					bleDisconnectSubCmd(v, bleBot)
				case "getChars":
					bleGetCharsSubCmd(v, bleBot)
				case "readChar":
					//bleReadCharSubCmd(v, bleBot)
				case "writeChar":
					//bleWriteCharSubCmd(v, bleBot)
				}
			}
		}
	}
}

// setStatusMenu /*******************************************************************************
func setStatusMenu(statusButtonTemp string, msg *tgbotapi.MessageConfig) {
	var menuKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("status: "+statusButtonTemp),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/scan"),
			tgbotapi.NewKeyboardButton("/advert"),
			tgbotapi.NewKeyboardButton("/stop"),
			tgbotapi.NewKeyboardButton("/exit"),
		),
	)
	msg.ReplyMarkup = menuKeyboard
}

// onScan /*******************************************************************************
func onScan(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
	var elemInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Connect", "connect"),
		),
	)
	if bleScannedAmount != 0 {
		temp := bleScannedAmount - 1
		for temp >= 0 {
			if bleScanBotList[temp].scannedAddress == device.Address {
				return
			}
			temp--
		}
	}
	deviceInfo := fmt.Sprintf(
		"Status: not connected\n"+
			"Local Name: %s\n"+
			"MAC: %s\n"+
			"RSSI: %d\n"+
			"Advertisement Payload: ",
		device.LocalName(),
		device.Address.String(),
		device.RSSI)
	msg := tgbotapi.NewMessage(bleChatId, deviceInfo)
	msg.ReplyMarkup = elemInlineKeyboard
	messageTemp, err := bleBot.Send(msg)
	if err != nil {
		log.Printf("ERROR %v: Cant send message about scanned device: %s, %s", err, device.Address.String(), device.LocalName())
	}
	bleScanBotList[bleScannedAmount].scannedAddress = device.Address
	bleScanBotList[bleScannedAmount].messageId = messageTemp.MessageID
	bleScannedAmount++
	if bleScannedAmount >= 40 {
		bleScannedAmount = 0
	}
}

// bleCheckChatId /*******************************************************************************
func bleCheckChatId(message *tgbotapi.Message, bot *tgbotapi.BotAPI) int {
	var err error
	if bleChatId != message.Chat.ID && bleChatId != 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, bleMessageHeader+"SERVER IS BUSY FOR ANOTHER CLIENT, WAIT AND TRY AGAIN LATER")
		_, err = bot.Send(msg)
		if err != nil {
			log.Printf("ERROR %v: Cant send message about busy server", err)
		}
		return 0
	} else {
		return 1
	}
}

// bleTimerProc /*******************************************************************************
func bleTimerProc(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	var err error
	bleTimer = time.NewTimer(5 * time.Minute)
	go func() {
		<-bleTimer.C
		bleChatId = 0
		err = bleAdapter.StopScan()
		if err != nil {
			log.Printf("ERROR %v: Cant stop scanning", err)
		}
		err = bleAdvertiser.Stop()
		if err != nil {
			log.Printf("ERROR %v: Cant stop advert", err)
		}
		bleScannedAmount = 0
		bleStatus = bleNotReadyStatus
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, bleMessageHeader+"AFK 5MIN TIMEOUT")
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		_, err = bot.Send(msg)
		if err != nil {
			log.Printf("ERROR %v: Cant send message about timeout", err)
		}
	}()
}

// bleTimerStop /*******************************************************************************
func bleTimerStop() {
	if bleTimer != nil {
		stop := bleTimer.Stop()
		if !stop {
			log.Panicf("PANIC: Cant stop timer")
		}
	}
}

// bleDeleteUserMessage /*******************************************************************************
func bleDeleteUserMessage(messageID int, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewDeleteMessage(bleChatId, messageID)
	_, _ = bot.Send(msg)
}

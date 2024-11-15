package internble

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/saltosystems/winrt-go/windows/devices/bluetooth/genericattributeprofile"
	"log"
	"time"
	"tinygo.org/x/bluetooth"
)

var (
	ble        bleStruct
	bleTimer   *time.Timer
	bleBot     *tgbotapi.BotAPI
	bleAdapter bluetooth.Adapter
	bleChatId  int64
)

type bleStruct struct {
	scan    bleScan
	connect bleConnDevice
}

type bleScan struct {
	scanStatus       bool
	bleScannedAmount int
	scanList         []bleScannedDev
}

type bleScannedDev struct {
	device  bluetooth.ScanResult
	message tgbotapi.Message
}

type bleConnDevice struct {
	connStatus     bool
	message        tgbotapi.Message
	device         bluetooth.Device
	servicesAmount int
	services       []bleConnDevService
}

type bleConnDevService struct {
	uuid                  bluetooth.UUID
	name                  string
	service               genericattributeprofile.GattDeviceService
	characteristicsAmount int
	characteristics       []bleConnDecChars
}
type bleConnDecChars struct {
	uuid           bluetooth.UUID
	name           string
	characteristic genericattributeprofile.GattCharacteristic
	properties     genericattributeprofile.GattCharacteristicProperties
}

const (
	bleBotToken = "7328834082:AAGm8ygR2gr9oTlsdSUuULDwkHQWsUW-2G0"
)

const (
	bleNotReadyStatus  = "not ready"
	bleReadyStatus     = "ready"
	bleConnectedStatus = "connected"
	bleScanningStatus  = "scanning"
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
		bleTimerStop()
		bleTimerStart(v, bleBot)
		if v.Message != nil {
			if bleCheckChatId(v.Message, bleBot) != 0 {
				bleDeleteUserMessage(v.Message.MessageID, bleBot)
				bleMainKeyboardHandler(v, bleBot, v.Message.Command())
			}
		}
		//----------------------------------------------------------------------------------------
		if v.CallbackQuery != nil {
			if bleCheckChatId(v.CallbackQuery.Message, bleBot) != 0 {
				bleInlineKeyboardHandler(v, bleBot, v.CallbackQuery.Data)
			}
		}
	}
}

// bleOnScan /*******************************************************************************
// Когда передаешь стёроку в отправку он ее форматирует как то странно и возвращает другую, тоже
// самое происходит и с join
func bleOnScan(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
	var err error
	var curMessageText string
	var sendMessage tgbotapi.Message
	description := fmt.Sprintf(
		"[DEVICE DESCRIPTION]\n"+
			"Status: not connected\n"+
			"Local Name: %s\n"+
			"MAC: %s\n"+
			"RSSI: %d\n"+
			"Advertisement Payload:\n",
		device.LocalName(),
		device.Address.String(),
		device.RSSI)
	//----------------------------------------------------------------------------------------
	if ble.scan.bleScannedAmount != 0 {
		for _, v := range ble.scan.scanList {
			if v.device.Address == device.Address {
				curMessageText = v.message.Text
				if v.device.LocalName() == "" && device.LocalName() != "" {
					curMessageText = bleMessageEdit(curMessageText, fmt.Sprintf("Local Name: %s", device.LocalName()), bleMsgDescrFormatLocalName)
					msg := tgbotapi.NewEditMessageText(v.message.Chat.ID, v.message.MessageID, curMessageText)
					msg.ReplyMarkup = v.message.ReplyMarkup
					v.message, err = bleBot.Send(msg)
					if err != nil {
						log.Printf("ERROR %v: Cant send message about scanned device: %s, %s", err, device.Address.String(), device.LocalName())
					}
				}
				if v.device.RSSI != device.RSSI {
					curMessageText = bleMessageEdit(curMessageText, fmt.Sprintf("RSSI: %d", device.RSSI), bleMsgDescrFormatRSSI)
					msg := tgbotapi.NewEditMessageText(v.message.Chat.ID, v.message.MessageID, curMessageText)
					msg.ReplyMarkup = v.message.ReplyMarkup
					v.message, err = bleBot.Send(msg)
					if err != nil {
						log.Printf("ERROR %v: Cant send message about scanned device: %s, %s", err, device.Address.String(), device.LocalName())
					}
				}
				return
			}
		}
	}
	//----------------------------------------------------------------------------------------
	msg := tgbotapi.NewMessage(bleChatId, description)
	elemInlineKeyboard := []bleInlineButton{{Title: "Connect", Data: "connect"}}
	msg.ReplyMarkup = bleInlineKeyboardConfig(elemInlineKeyboard)
	sendMessage, err = bleBot.Send(msg)
	if err != nil {
		log.Printf("ERROR %v: Cant send message about scanned device: %s, %s", err, device.Address.String(), device.LocalName())
	}
	//----------------------------------------------------------------------------------------
	ble.scan.scanList = append(ble.scan.scanList, bleScannedDev{device, sendMessage})
	ble.scan.bleScannedAmount++
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
func bleTimerStart(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	bleTimer = time.AfterFunc(5*time.Minute, func() {
		bleExitCmd(update, bot)
	})
}

// bleTimerStop /*******************************************************************************
func bleTimerStop() {
	if bleTimer != nil {
		if bleTimer.C != nil {
			stop := bleTimer.Stop()
			if !stop {
				log.Printf("ERROR: Cant stop timer")
			}
		}
	}
}

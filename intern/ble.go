package internble

import (
	"fmt"
	"log"
	"math"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/saltosystems/winrt-go/windows/devices/bluetooth/genericattributeprofile"
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
	//updateTime *time.Timer
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
	characteristics       []bleConnDevChars
}

type bleConnDevChars struct {
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

// BleBotInit ************************************************
// Brief:   None
// Param:   None
// Return:  None
// ************************************************************
func BleBotInit() {
	var err error
	bleBot, err = tgbotapi.NewBotAPI(bleBotToken)
	if err != nil {
		log.Panicf("PANIC %v: Cant init bot token", err)
	}
	bleBot.Debug = false
}

// BleUpdate ************************************************
// Brief:   None
// Param:   None
// Return:  None
// ************************************************************
func BleUpdate() {
	updatesConfig := tgbotapi.NewUpdate(0)
	updatesConfig.Timeout = 60
	updates := bleBot.GetUpdatesChan(updatesConfig)
	//--------------------------------------------------------
	for v := range updates {
		bleTimerStop()
		bleTimerStart(v, bleBot)
		if v.Message != nil {
			if bleCheckChatId(v.Message, bleBot) != 0 { ////////////////////////////////////////////////////////////////////////////////
				bleDeleteUserMessage(v.Message.MessageID, bleBot)
				bleMainKeyboardHandler(v, bleBot, v.Message.Command())
			}
		}
		//----------------------------------------------------
		if v.CallbackQuery != nil {
			if bleCheckChatId(v.CallbackQuery.Message, bleBot) != 0 {
				bleInlineKeyboardHandler(v, bleBot, v.CallbackQuery.Data)
			}
		}
	}
}

// BleBotInit ************************************************
// Brief:   None
// Param:   None
// Return:  None
// ************************************************************
func bleOnScan(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
	if ble.scan.bleScannedAmount != 0 {
		for _, v := range ble.scan.scanList {
			if v.device.Address == device.Address {
				bleOnScanUpdate(device, v)
				return
			}
		}
	}
	bleOnScanNew(device)
}

// BleBotInit ************************************************
// Brief:   None
// Param:   None
// Return:  None
// ************************************************************
func bleOnScanUpdate(device bluetooth.ScanResult, v bleScannedDev) {
	var err error
	// if v.updateTime != nil {
	// 	if v.updateTime.C != nil {
	// 		stop := v.updateTime.Stop()
	// 		if !stop {
	// 			log.Printf("ERROR: Cant stop timer")
	// 		}
	// 		v.updateTime = time.Tick(5 * time.Second)
	// 	} else {
	// 		return
	// 	}
	// }
	if v.device.LocalName() == "" && device.LocalName() != "" {
		v.message.Text = bleMessageEdit(v.message.Text, fmt.Sprintf("Local Name: %s", device.LocalName()), bleMsgDescrFormatLocalName)
		msg := tgbotapi.NewEditMessageText(v.message.Chat.ID, v.message.MessageID, v.message.Text)
		msg.ReplyMarkup = v.message.ReplyMarkup
		v.message, err = bleBot.Send(msg)
		if err != nil {
			log.Printf("ERROR %v: Cant send message about scanned device: %s, %s", err, device.Address.String(), device.LocalName())
		}
	}
	if math.Abs(float64(v.device.RSSI-device.RSSI)) >= 5 {
		v.message.Text = bleMessageEdit(v.message.Text, fmt.Sprintf("RSSI: %d", device.RSSI), bleMsgDescrFormatRSSI)
		msg := tgbotapi.NewEditMessageText(v.message.Chat.ID, v.message.MessageID, v.message.Text)
		msg.ReplyMarkup = v.message.ReplyMarkup
		v.message, err = bleBot.Send(msg)
		if err != nil {
			log.Printf("ERROR %v: Cant send message about scanned device: %s, %s", err, device.Address.String(), device.LocalName())
		}
	}
}

// BleBotInit ************************************************
// Brief:   None
// Param:   None
// Return:  None
// ************************************************************
func bleOnScanNew(device bluetooth.ScanResult) {
	var err error
	var sendMessage tgbotapi.Message
	description := fmt.Sprintf(
		"[DEVICE DESCRIPTION]\n"+"Status: not connected\n"+"Local Name: %s\n"+"MAC: %s\n"+"RSSI: %d\n"+"Advertisement Payload:",
		device.LocalName(), device.Address.String(), device.RSSI)
	msg := tgbotapi.NewMessage(bleChatId, description)
	elemInlineKeyboard := []bleInlineButton{{Title: "Connect", Data: "connect"}}
	msg.ReplyMarkup = bleInlineKeyboardConfig(elemInlineKeyboard)
	sendMessage, err = bleBot.Send(msg)
	if err != nil {
		log.Printf("ERROR %v: Cant send message about scanned device: %s, %s", err, device.Address.String(), device.LocalName())
	}
	//updateTime := time.NewTimer(5 * time.Second)
	ble.scan.scanList = append(ble.scan.scanList, bleScannedDev{device, sendMessage})
	ble.scan.bleScannedAmount++

}

// BleBotInit ************************************************
// Brief:   None
// Param:   None
// Return:  None
// ************************************************************
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

// BleBotInit ************************************************
// Brief:   None
// Param:   None
// Return:  None
// ************************************************************
func bleTimerStart(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	bleTimer = time.AfterFunc(5*time.Minute, func() {
		bleExitCmd(update, bot)
	})
}

// BleBotInit ************************************************
// Brief:   None
// Param:   None
// Return:  None
// ************************************************************
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

package internble

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"runtime"
	"strings"
)

const (
	bleMessageHeader = "BOT: "
)

const (
	bleMsgDescrFormatHeader               = iota //DEVICE DESCRIPTION
	bleMsgDescrFormatStatus                      //Status:
	bleMsgDescrFormatLocalName                   //Local name:
	bleMsgDescrFormatMAC                         //MAC:
	bleMsgDescrFormatRSSI                        //RSSI:
	bleMsgStringFormatAdvertPayloadString        //Advertisement payload:
)

const (
	bleMsgServiceFormatHeader     = iota //SERVICE
	bleMsgServiceServiceLocalName        //Local name:
	bleMsgServiceServiceUUID             //UUID:
	bleMsgServiceServiceData             //data:
)

const (
	bleMsgStringCharsHeader     = iota //CHARACTERISTIC
	bleMsgStringCharsLocalName         //Local name:
	bleMsgStringCharsUUID              //UUID:
	bleMsgStringCharsData              //data:
	bleMsgStringCharsProperties        //properties:
)

// BleBotInit ************************************************
// Brief:   None
// Param:   None
// Return:  None
// ***********************************************************
func bleDeleteUserMessage(messageID int, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewDeleteMessage(bleChatId, messageID)
	_, _ = bot.Send(msg)
}

// BleBotInit ************************************************
// Brief:   None
// Param:   None
// Return:  None
// ***********************************************************
func bleMessageEdit(message, new string, stringNum int) string {
	stringParts := strings.Split(message, "\n")
	if len(stringParts)-1 < stringNum {
		return message
	}
	if new == "" {
		return message
	}
	stringParts[stringNum] = new
	return strings.Join(stringParts, "\n")
}

// BleBotInit ************************************************
// Brief:   None
// Param:   None
// Return:  None
// ***********************************************************
func bleMessageSend(update tgbotapi.Update, bot *tgbotapi.BotAPI, status string, keyboard []string, message string) error {
	var err error
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, bleMessageHeader+message)
	if keyboard != nil {
		if keyboard[0] == "REMOVE" {
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		} else {
			msg.ReplyMarkup = bleKeyboardConfig(status, keyboard)
		}
	}
	_, err = bot.Send(msg)
	if err != nil {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			log.Printf("ERROR %v: Cant send message! file: %s, line: %d", err, file, line)
		} else {
			log.Printf("ERROR %v: Cant send message! file: UNKNOWN, line: UNKNOWN", err)
		}
	}
	return err
}

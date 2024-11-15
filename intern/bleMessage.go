package internble

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

// bleDeleteUserMessage /*******************************************************************************
func bleDeleteUserMessage(messageID int, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewDeleteMessage(bleChatId, messageID)
	_, _ = bot.Send(msg)
}

// bleMessageEdit /*******************************************************************************
func bleMessageEdit(message, new string, stringNum int) string {
	stringParts := strings.Split(message, "\n")
	if len(stringParts)-1 < stringNum {
		return message
	}
	if new == "" {
		return message
	}
	stringParts[stringNum] = new
	return strings.Join(stringParts, "\n") + "\n"
}

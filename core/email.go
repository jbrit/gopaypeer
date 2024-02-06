package core

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/mailjet/mailjet-apiv3-go/v4"
)

var godotenvErr = godotenv.Load()

func SendMail(recipient, subject, message string) error {
	mailjetClient := mailjet.NewMailjetClient(os.Getenv("MJ_APIKEY_PUBLIC"), os.Getenv("MJ_APIKEY_PRIVATE"))
	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: "noreply@ajibolaojo.com",
				Name:  "Ajibola from Paypeer",
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: recipient,
				},
			},
			Subject:  subject,
			TextPart: message,
		},
	}
	messages := mailjet.MessagesV31{Info: messagesInfo}
	_, err := mailjetClient.SendMailV31(&messages)
	return err
}

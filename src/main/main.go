package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slack"
	"slacker"
)

func postMessagesUsingWebhooks() {
	//Creating a slack instance
	slackInst := slack.Slack{
		//Webhook: "https://hooks.slack.com/services/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",
		Webhook: "https://hooks.slack.com/services/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",
	}

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter text: ")
		text, _ := reader.ReadString('\n')

		fmt.Println("Input read is  : ", text)

		//Sending the message
		newMessage := slack.Message{
			Text: fmt.Sprintf("%s", text),
		}

		err1 := slackInst.SendMessage(&newMessage)

		if err1 != nil {
			log.Println("Error occurend in sending the message :", err1)
			return
		}

	}
	fmt.Println("Ending the communicator ")
	return
}

func workInRTMClientMode() {

	// Open connection
	slackApp, err := slack.NewRTMClient("xoxp-163166927445-163077586066-163077793426-15cd90f16e06ce562ff8155655521d20")
	if err != nil {
		log.Fatalf("Error in creating rtm client : %s", err)
	}

	for {
		msg, err := slackApp.Receive()
		if err != nil {
			log.Fatal(err)
		}

		if len(msg) != 0 {
			quote := "hi this is slack app responding. I can listen you "
			if err != nil {
				continue
			}

			// Post the quote
			slackApp.Send(quote)
		}
	}

}

func main() {

	//postMessagesUsingWebhooks()
	//return

	addr := "localhost:8081"
	token := "xoxp-163166927445-163077586066-163077793426-15cd90f16e06ce562ff8155655521d20"

	log.Printf("[info] starting slacker on port : %s", addr)
	slack := slacker.New()

	slack.HandleFunc("hello", token, func(w io.Writer, cmd *slacker.Command) error {
		fmt.Fprint(w, "Hello")
		fmt.Fprint(w, " World")
		return nil
	})

	log.Fatal(http.ListenAndServe(addr, slack))

}

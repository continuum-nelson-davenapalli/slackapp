package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const (
	slackApiUrl = string("https://api.slack.com/")
	rtmstart    = string("https://slack.com/api/rtm.start")
	messageType = string("message")
)

type attachment struct {
	Title     string `json:"title"`
	TitleLink string `json:"title_link"`
	Text      string `json:"text"`
}

type Slack struct {
	Webhook string
}

type Message struct {
	Username    string       `json:"user_name"`
	IconEmoji   string       `json:"icon_emoji"`
	Text        string       `json:"text"`
	Attachments []attachment `json:"attachments"`
}

// Standard slack message format
type messageRx struct {
	Type    string `json:"type"`
	Channel string `json:"channel"`
	User    string `json:"user"`
	Text    string `json:"text"`
	TS      string `json:"ts"`
	// TODO does not include edits or subtypes
}

// Standard slack message format
type messageTx struct {
	Id      uint64 `json:"id"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

// SlackRTMClient is a real time messaging client for slack
type SlackRTMClient struct {
	ws           *websocket.Conn
	id           string
	msg          messageRx
	send         messageTx
	sendSequence uint64
}

// NewRTMClient returns a new real time messaging client for a given token
func NewRTMClient(token string) (*SlackRTMClient, error) {

	// Request an rtm session
	socketUrl, id, err := rtmStart(token)
	if err != nil {
		return nil, err
	}

	// Connect to the rtm socket
	ws, err := websocket.Dial(socketUrl, "", slackApiUrl)
	if err != nil {
		return nil, err
	}

	return &SlackRTMClient{ws, id, messageRx{}, messageTx{}, 0}, nil
}

func rtmStart(tok string) (socket string, id string, err error) {

	url := fmt.Sprintf("%v?token=%v&no_unreads=true&scope=bot,incoming-webhook", rtmstart, tok)
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("Request Failed: %v", resp.StatusCode)
	}

	defer resp.Body.Close()
	jsonData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	// Json format
	type self struct {
		Id string `json:"id"`
	}
	type rtmStartResp struct {
		Ok    bool   `json:"ok"`
		Error string `json:"error"`
		Url   string `json:"url"`
		Self  self   `json:"self"`
	}

	// Parse the response
	var data rtmStartResp
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return "", "", err
	}

	if !data.Ok {
		return "", "", fmt.Errorf(data.Error)
	}

	return data.Url, data.Self.Id, nil
}

// Receive receives data until a message for the requested id is obtained
func (s *SlackRTMClient) Receive() (string, error) {

	checkMsg := func() error {
		if err := websocket.JSON.Receive(s.ws, &s.msg); err != nil {
			return err
		}
		return nil
	}

	var err error
	directMsg := fmt.Sprintf("<@%v>: ", s.id)
	for err = checkMsg(); err != nil || !strings.HasPrefix(s.msg.Text, directMsg) || s.msg.Type != messageType; {
		err = checkMsg()
	}

	text := s.msg.Text[len(s.id)+5:]
	return text, nil
}

// Send a response to the same channel as previous received message
func (s *SlackRTMClient) Send(m string) error {

	// Setup the send message
	s.send.Channel = s.msg.Channel
	s.sendSequence++
	s.send.Id = s.sendSequence
	s.send.Text = m
	s.send.Type = messageType

	return websocket.JSON.Send(s.ws, &s.send)
}

func (s *Slack) SendMessage(newMessage *Message) error {
	log.Printf("Sending message: %v", newMessage)
	messageBytes, err := json.Marshal(*newMessage)
	if err != nil {
		return err
	}
	resp, err := http.Post(s.Webhook, "application/json", bytes.NewBuffer(messageBytes))
	log.Printf("Got status code: %d from slack, err: %v", resp.StatusCode, err)
	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("Got invalid status code %d", resp.StatusCode))
	}
	return err
}

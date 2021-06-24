package main

import (
	"encoding/gob"
	"io"
	"sync"
)

type DownloadButtonMessage struct {
	ChannelID string `json:"channelId"`
	MessageID string `json:"messageId"`
}

type DownloadButtonMessageCollection struct {
	mutex    sync.Mutex
	messages []*DownloadButtonMessage
}

func NewDownloadButtonMessageCollection() *DownloadButtonMessageCollection {
	return &DownloadButtonMessageCollection{
		mutex:    sync.Mutex{},
		messages: make([]*DownloadButtonMessage, 0),
	}
}

func (mc *DownloadButtonMessageCollection) Add(m DownloadButtonMessage) {
	mc.mutex.Lock()
	mc.messages = append(mc.messages, &m)
	mc.mutex.Unlock()
}

func (mc *DownloadButtonMessageCollection) Remove(m DownloadButtonMessage) {
	mc.mutex.Lock()
	for i, message := range mc.messages {
		if message.ChannelID == m.ChannelID && message.MessageID == m.MessageID {
			mc.messages[i] = mc.messages[len(mc.messages)-1]
			mc.messages = mc.messages[:len(mc.messages)-1]
		}
	}
	mc.mutex.Unlock()
}

func (mc *DownloadButtonMessageCollection) Messages() []DownloadButtonMessage {
	mc.mutex.Lock()
	messages := make([]DownloadButtonMessage, len(mc.messages))
	for i, m := range mc.messages {
		messages[i] = *m
	}
	mc.mutex.Unlock()

	return messages
}

func (mc *DownloadButtonMessageCollection) Write(w io.Writer) error {
	messages := mc.Messages()
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(messages)
	if err != nil {
		return err
	}
	return nil
}

func (mc *DownloadButtonMessageCollection) Read(r io.Reader) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&mc.messages)
	return err
}

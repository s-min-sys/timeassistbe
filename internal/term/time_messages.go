package term

import (
	"sync"
	"time"
)

type TimeMessage struct {
	message  string
	expireAt time.Time
}

type TimeMessages struct {
	messagesLock sync.Mutex
	messages     []*TimeMessage
}

func NewTimeMessages() *TimeMessages {
	return &TimeMessages{
		messages: make([]*TimeMessage, 0, 100),
	}
}

func (tms *TimeMessages) Add(message string, expireAt time.Time) {
	tms.messagesLock.Lock()
	defer tms.messagesLock.Unlock()

	tms.messages = append(tms.messages, &TimeMessage{
		message:  message,
		expireAt: expireAt,
	})
}

func (tms *TimeMessages) GetMessages() (messages []string) {
	tms.messagesLock.Lock()
	defer tms.messagesLock.Unlock()

	timeNow := time.Now()

	for i := 0; i < len(tms.messages); {
		if timeNow.After(tms.messages[i].expireAt) {
			tms.messages = append(tms.messages[:i], tms.messages[i+1:]...)
		} else {
			messages = append(messages, tms.messages[i].message)

			i++
		}
	}

	return
}

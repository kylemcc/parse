package parse

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// Interface representing a Parse Push notification and the various
// options for sending a push notification. This API is chainable for
// conveniently building push notifications:
//
// parse.NewPushNotification().Channels("chan1", "chan2").Where(parse.NewPushQuery().EqualTo("deviceType", "ios")).Data(map[string]interface{}{"alert": "hello"}).Send()
type PushNotification interface {
	// Set the query for advanced targeting
	//
	// use parse.NewPushQuery to create a new query
	Where(q Query) PushNotification

	// Set the channels to target
	Channels(c ...string) PushNotification

	// Specify a specific time to send this push
	PushTime(t time.Time) PushNotification

	// Set the time this push notification should expire if it can't be immediately sent
	ExpirationTime(t time.Time) PushNotification

	// Set the duration after which this push notification should expire if it can't be immediately sent
	ExpirationInterval(d time.Duration) PushNotification

	// Set the payload for this push notification
	Data(d map[string]interface{}) PushNotification

	// Send the push notification
	Send() error
}

type pushT struct {
	shouldUseMasterKey bool
	channels           []string
	expirationInterval int64
	expirationTime     *Date
	pushTime           *Date
	where              map[string]interface{}
	data               map[string]interface{}
}

func (p *pushT) method() string {
	return "POST"
}

func (p *pushT) endpoint() (string, error) {
	u := url.URL{}
	u.Scheme = "https"
	u.Host = parseHost
	u.Path = "/1/push"

	return u.String(), nil
}

func (p *pushT) body() (string, error) {
	if p.expirationTime != nil && p.expirationInterval > 0 {
		return "", errors.New("cannot use both expiration_time and expiration_interval")
	}

	payload, err := json.Marshal(&struct {
		Channels           []string               `json:"channels,omitempty"`
		ExpirationTime     *Date                  `json:"expiration_time,omitempty"`
		ExpirationInterval int64                  `json:"expiration_interval,omitempty"`
		PushTime           *Date                  `json:"push_time,omitempty"`
		Data               map[string]interface{} `json:"data,omitempty"`
		Where              map[string]interface{} `json:"where,omitempty"`
	}{
		Channels:           p.channels,
		ExpirationTime:     p.expirationTime,
		PushTime:           p.pushTime,
		ExpirationInterval: p.expirationInterval,
		Data:               p.data,
		Where:              p.where,
	})

	fmt.Printf("body: %s\n", payload)
	return string(payload), err
}

func (p *pushT) useMasterKey() bool {
	return p.shouldUseMasterKey
}

func (p *pushT) session() *sessionT {
	return nil
}

func (p *pushT) contentType() string {
	return "application/json"
}

// Convenience function for creating a new query for use in SendPush.
func NewPushQuery() Query {
	q, _ := NewQuery(&Installation{})
	return q
}

// Create a new Push Notifaction
//
// See the Push Notification Guide for more details: https://www.parse.com/docs/push_guide#sending/REST
func NewPushNotification() PushNotification {
	return &pushT{}
}

func (p *pushT) Where(q Query) PushNotification {
	p.where = q.(*queryT).where
	return p
}

func (p *pushT) Channels(c ...string) PushNotification {
	p.channels = c
	return p
}

func (p *pushT) PushTime(t time.Time) PushNotification {
	d := Date(t)
	p.pushTime = &d
	return p
}

func (p *pushT) ExpirationTime(t time.Time) PushNotification {
	d := Date(t)
	p.expirationTime = &d
	return p
}

func (p *pushT) ExpirationInterval(d time.Duration) PushNotification {
	p.expirationInterval = int64(d.Seconds())
	return p
}

func (p *pushT) Data(d map[string]interface{}) PushNotification {
	p.data = d
	return p
}

func (p *pushT) Send() error {
	b, err := defaultClient.doRequest(p)
	data := map[string]interface{}{}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	fmt.Printf("data: %v\n", data)
	return err
}

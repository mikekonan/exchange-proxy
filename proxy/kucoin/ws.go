package kucoin

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dgrr/websocket"
	"github.com/google/uuid"
	"github.com/mailru/easyjson"
	"github.com/mikekonan/exchange-proxy/proxy"
	"github.com/mikekonan/exchange-proxy/store"
	"github.com/sirupsen/logrus"
	"go.uber.org/ratelimit"
)

const (
	bulletPublicPath = "/api/v1/bullet-public"

	welcomeMessageType   = "welcome"
	messageMessageType   = "message"
	subscribeMessageType = "subscribe"

	ping = "ping"
	pong = "pong"

	marketCandlesTopicPrefix = "/market/candles:"
)

type subscriber struct {
	pool []*ws

	wsRl   ratelimit.Limiter
	l      *sync.Mutex
	subs   map[string]struct{}
	client *proxy.Client
	config *Config
	store  *store.Store
	httpRl ratelimit.Limiter
}

func (s *subscriber) subscribeKLines(pair string, tf string) {
	s.l.Lock()
	defer s.l.Unlock()

	topic := wsTopic(pair, tf)
	if _, ok := s.subs[topic]; ok {
		return
	}

	s.subs[topic] = struct{}{}

	for i, c := range s.pool {
		if c.subsCount == s.config.KucoinTopicsPerWs {
			continue
		}

		c.subsCount += 1
		s.wsRl.Take()
		if err := c.subscribeKLines(topic); err != nil {
			logrus.Fatal(err)
		}

		logrus.Infof("#%d-%d topic: '%s' subscribing...", i+1, c.subsCount, topic)

		return
	}

	wsConn := &ws{
		subsCount:  0,
		client:     s.client,
		httpRl:     s.httpRl,
		wsRl:       s.wsRl,
		retryCount: 15,
		store:      s.store,
		config:     s.config,
		id:         uuid.New(),
		conn:       nil,
		pingPongCh: make(chan uuid.UUID, 1),

		writeLock: new(sync.Mutex),
	}

	wsConn.connect()
	go wsConn.serve()

	wsConn.subsCount += 1
	if err := wsConn.subscribeKLines(topic); err != nil {
		logrus.Fatal(err)
	}

	s.pool = append(s.pool, wsConn)
	logrus.Infof("#%d-%d topic: '%s' subscribing...", len(s.pool), 1, topic)
}

type ws struct {
	id uuid.UUID

	subsCount int

	client     *proxy.Client
	httpRl     ratelimit.Limiter
	retryCount int
	store      *store.Store
	config     *Config

	conn *websocket.Client
	wsRl ratelimit.Limiter

	pingInterval *time.Ticker
	pingTimeout  time.Duration
	pingPongCh   chan uuid.UUID

	writeLock *sync.Mutex
}

func (w *ws) executeBulletPublicRequest() (int, *bulletPublicResponse, error) {
	w.httpRl.Take()

	statusCode, data, err := w.client.Post(nil, fmt.Sprintf("%s/%s", w.config.KucoinApiURL, bulletPublicPath), nil)

	if err != nil {
		return statusCode, nil, err
	}

	bulletPublicResponse := &bulletPublicResponse{}
	if err := easyjson.Unmarshal(data, bulletPublicResponse); err != nil {
		return statusCode, nil, err
	}

	return statusCode, bulletPublicResponse, nil
}

func (w *ws) getBulletPublic() (int, *bulletPublicResponse, error) {
	for i := 1; i <= w.retryCount; i++ {
		w.httpRl.Take()

		if statusCode, bulletPublicResponse, err := w.executeBulletPublicRequest(); statusCode == 200 {
			return statusCode, bulletPublicResponse, nil
		} else {
			if i == w.retryCount {
				return statusCode, bulletPublicResponse, fmt.Errorf("get bullet public exceeded retry '%d' attemts: %w", w.retryCount, err)
			}

			time.Sleep(time.Second)
		}
	}

	return 500, nil, fmt.Errorf("retry count is zero")
}

func (w *ws) connect() {
	_, bulletResp, err := w.getBulletPublic()
	if err != nil {
		logrus.Fatal(err)
	}

	w.pingInterval = time.NewTicker(time.Millisecond * time.Duration(bulletResp.Data.InstanceServers[0].PingInterval))
	w.pingTimeout = time.Millisecond * time.Duration(bulletResp.Data.InstanceServers[0].PingTimeout)

	path := fmt.Sprintf("%s?token=%s&connectId=%s", bulletResp.Data.InstanceServers[0].Endpoint, bulletResp.Data.Token, w.id.String())

	conn, err := websocket.Dial(path)
	if err != nil {
		logrus.Fatal(err)
	}

	w.conn = conn

	w.readWelcomeMsg()
	go w.pingPongRoutine()
}

func (w *ws) handlePongResponse(waitForID uuid.UUID) {
	logrus.Debugf("handling pong message with id '%s'", waitForID.String())

	for {
		select {
		case <-time.After(w.pingTimeout):
			logrus.Fatal("pong timeout violation")
		case receivedID := <-w.pingPongCh:
			if waitForID == receivedID {
				return
			} else {
				logrus.Warnf("ping/pong id mismatch: sent '%s', received '%s'", waitForID.String(), receivedID.String())
				w.pingPongCh <- receivedID
			}
		}
	}
}

func (w *ws) pingPongRoutine() {
	w.pingPongCh <- w.writePing()

	for {
		select {
		case <-w.pingInterval.C:
			w.pingPongCh <- w.writePing()
		case waitForID := <-w.pingPongCh:
			w.handlePongResponse(waitForID)
		}
	}
}

func (w *ws) writePing() uuid.UUID {
	id := uuid.New()

	logrus.Debugf("writing ping message with id '%s'", id.String())

	data, err := easyjson.Marshal(pingMessageRequest{ID: id, Type: ping})

	if err != nil {
		logrus.Fatal(err)
	}

	w.writeLock.Lock()
	defer w.writeLock.Unlock()

	if _, err := w.conn.Write(data); err != nil {
		logrus.Fatal(err)
	}

	return id
}

func (w *ws) readWelcomeMsg() {
	frame := websocket.AcquireFrame()
	defer websocket.ReleaseFrame(frame)

	_, err := w.conn.ReadFrame(frame)
	if err != nil {
		logrus.Fatalf("failed getting welcome message: %v", err)
	}

	welcomeMsg := &welcomeMessageResponse{}
	if err := easyjson.Unmarshal(frame.Payload(), welcomeMsg); err != nil {
		logrus.Fatalf("failed parsing welcome message: %v", err)
	}

	if welcomeMsg.ID != w.id && welcomeMsg.Type != welcomeMessageType {
		logrus.Fatal("failed establishing ws connection: id or message is incorrect")
	}
}

func (w *ws) subscribeKLines(topic string) error {
	topic = fmt.Sprintf("%s%s", marketCandlesTopicPrefix, topic)

	logrus.Debugf("subscribing to '%s'...", topic)

	message := subscribeMessageRequest{
		ID:             uuid.New(),
		Type:           subscribeMessageType,
		Topic:          topic,
		PrivateChannel: false,
		Response:       false,
	}

	data, err := easyjson.Marshal(message)
	if err != nil {
		logrus.Fatal(err)
	}

	w.writeLock.Lock()
	defer w.writeLock.Unlock()

	if _, err := w.conn.Write(data); err != nil {
		logrus.Fatal(err)
	}

	return nil
}

func (w *ws) processFrame(frame *websocket.Frame) {
	message := &genericMessageResponse{}
	if err := easyjson.Unmarshal(frame.Payload(), message); err != nil {
		logrus.Fatalf("failed parsing generic message: %v. message is : '%s'", err, string(frame.Payload()))

		return
	}

	logrus.Tracef("received message '%s'-'%s'-'%s'", message.Topic, message.Subject, message.Type)

	switch message.Type {
	case pong:
		w.pingPongCh <- message.ID
		return

	case messageMessageType:
		if strings.HasPrefix(message.Topic, marketCandlesTopicPrefix) {
			pairTf := message.Topic[len(marketCandlesTopicPrefix):]
			pair := strings.Split(pairTf, "_")[0]
			tf := strings.Split(pairTf, "_")[1]

			entry := &kLineUpdateMessageEntry{}
			if err := easyjson.Unmarshal(message.Data, entry); err != nil {
				logrus.Fatal(err)
			}

			w.store.Store(storeKey(pair, tf), timeframeToDuration(tf), parseCandle(entry.Candles))

			return
		}
	}
}

func (w *ws) serve() {
	for {
		frame := websocket.AcquireFrame()

		if _, err := w.conn.ReadFrame(frame); err != nil {
			logrus.Fatal(err)
		}

		w.processFrame(frame)

		websocket.ReleaseFrame(frame)
	}
}

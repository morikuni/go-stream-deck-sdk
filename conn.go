package streamdeck

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"golang.org/x/net/websocket"
)

type Conn struct {
	conn *websocket.Conn
}

func DialAuto() (*Conn, error) {
	// TODO: fix name
	fs := flag.NewFlagSet("plugin", flag.ContinueOnError)

	port := fs.String("port", "", "port to bind websocket server")
	uuid := fs.String("pluginUUID", "", "the ID of the plugin")
	event := fs.String("registerEvent", "", "the event type to register websocket connection")
	info := fs.String("info", "", "the context of the plugin")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		return nil, fmt.Errorf("invalid parameter: %w", err)
	}

	return Dial(*port, *event, *uuid, *info)
}

func Dial(port, event, uuid, info string) (*Conn, error) {
	if port == "" {
		return nil, errors.New("port is emtpy")
	}
	if uuid == "" {
		return nil, errors.New("uuid is emtpy")
	}
	if event == "" {
		return nil, errors.New("event is emtpy")
	}
	if info == "" {
		return nil, errors.New("info is emtpy")
	}

	conn, err := websocket.Dial("ws://localhost:"+port, "", "http://localhost:"+port)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the server: %w", err)
	}

	err = websocket.JSON.Send(conn, map[string]string{
		"event": event,
		"uuid":  uuid,
	})
	if err != nil {
		return nil, fmt.Errorf("error during registratino procedure: %w", err)
	}

	return &Conn{conn}, nil
}

func (c *Conn) Receive() (Event, error) {
	var payload eventPayload
	err := websocket.JSON.Receive(c.conn, &payload)
	if err != nil {
		return nil, fmt.Errorf("failed to receive an event: %w", err)
	}

	ev, err := payload.Typed()
	if err != nil {
		return nil, fmt.Errorf("failed to parse an event: %w: %v", err, payload)
	}

	return ev, nil
}

func (c *Conn) Send(cmd Command) error {
	payload, err := newCommandPayload(cmd)
	if err != nil {
		return err
	}

	err = websocket.JSON.Send(c.conn, payload)
	if err != nil {
		return fmt.Errorf("failed to send a command: %w: %v", err, cmd)
	}

	return nil
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) Logger() *Logger {
	return &Logger{c}
}
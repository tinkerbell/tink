package informers

import (
	sha "crypto/sha256"
	"encoding/base64"
	"encoding/json"

	"github.com/tinkerbell/tink/protos/events"
)

// ListenerFactory is the interface for listener factory
type ListenerFactory interface {
	ListenerFor(req *events.WatchRequest) (Listener, error)
}

type listenerFactory struct {
	listeners map[string]Listener
}

var factory ListenerFactory

// Factory returns an instance of the listener factory
func Factory() ListenerFactory {
	if factory == nil {
		factory = &listenerFactory{
			listeners: make(map[string]Listener),
		}
	}
	return factory
}
func (f *listenerFactory) ListenerFor(req *events.WatchRequest) (Listener, error) {
	listenerKey, err := getKey(req)
	if err != nil {
		return nil, err
	}

	if listener, found := f.listeners[listenerKey]; found {
		return listener, nil
	}
	listener, err := listenerFor(req)
	if err != nil {
		return nil, err
	}
	f.listeners[listenerKey] = listener
	return listener, nil
}

func getKey(req *events.WatchRequest) (string, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	h := sha.New()
	return base64.StdEncoding.EncodeToString(h.Sum(reqBytes)), nil
}

func listenerFor(req *events.WatchRequest) (Listener, error) {
	listener, err := newPQListener()
	if err != nil {
		return nil, err
	}
	return &sharedListener{
		req:        req,
		pqListener: listener,
	}, nil
}

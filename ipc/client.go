package ipc

import (
	"github.com/Ericwyn/EzeTranslate/log"
)

func SendMessage(message string) error {
	us := NewUnixSocket(UnixSocketAddress)
	res, err := us.ClientSendContext(message)
	if err != nil {
		return err
	}
	log.D("ipc response:", res)
	return nil
}

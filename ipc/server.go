package ipc

import "github.com/Ericwyn/EzeTranslate/log"

const MessageNewSelection string = "NEW_SELECTION\n"

type MessageHandler func(message string)

const UnixSocketAddress = "/tmp/trans_utils.socket"

var PONG string = "PONG\n"

func StartUnixSocketListener(messageHandler MessageHandler) {
	log.D("开始监听 IPC 消息")
	us := NewUnixSocket(UnixSocketAddress)
	us.SetContextHandler(func(message string) string {
		messageHandler(message)
		return PONG
	})

	us.StartServer()
}

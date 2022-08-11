package ipc

import "github.com/Ericwyn/EzeTranslate/log"

type IpcMessage string

const IpcMessageNewSelection IpcMessage = "NEW_SELECTION\n"
const IpcMessageOcr IpcMessage = "OCR\n"
const IpcMessageOcrAndTrans IpcMessage = "OCR_AND_TRANS\n"
const IpcMessagePing IpcMessage = "PING\n"

type MessageHandler func(message IpcMessage)

const UnixSocketAddress = "/tmp/trans_utils.socket"

var PONG = "PONG\n"

func StartUnixSocketListener(messageHandler MessageHandler) {
	log.D("开始监听 IPC 消息")
	us := NewUnixSocket(UnixSocketAddress)
	us.SetContextHandler(func(message IpcMessage) string {
		messageHandler(message)
		return PONG
	})

	us.StartServer()
}

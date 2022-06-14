package ultimate_cedar

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	jsoniter "github.com/json-iterator/go"
)

const MagicWebsocketKey = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

var cedarWebsocketHub *sync.Map

// WebsocketSwitchProtocol
// 用来扩展websocket
// 只实现了保持在线和推送
// GET /chat HTTP/1.1
// Host: example.com:8000
// Upgrade: websocket
// Connection: Upgrade
// Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
// Sec-WebSocket-Version: 13
func WebsocketSwitchProtocol(w ResponseWriter, r Request, key string, fn func(value *CedarWebSocketBuffReader)) {
	// 申请一个map
	if cedarWebsocketHub == nil {
		cedarWebsocketHub = &sync.Map{}
	}
	version := r.Header.Get("Sec-Websocket-Version")
	if debug {
		log.Println("[cedar] websocket version", version)
	}
	if version != "13" {
		w.WriteHeader(400)
		return
	}
	swKey := r.Header.Get("Sec-WebSocket-Key")
	// 计算值
	newKey := getNewKey(swKey)
	w.Header().Add("Upgrade", "websocket")
	w.Header().Add("Connection", "Upgrade")
	w.Header().Add("Sec-Websocket-Accept", newKey)
	w.WriteHeader(http.StatusSwitchingProtocols)
	_, err := w.Write(nil)
	if err != nil {
		return
	}
	hj, ok := w.writer.(http.Hijacker)
	if !ok {
		http.Error(w, "Not a Hijacker", 500)
		return
	}
	nc, _, err := hj.Hijack()
	if err != nil {
		log.Panicln(err)
	}
	cedarWebsocketHub.Store(key, nc)
	go func(nc net.Conn) {
		closeHj := make(chan bool)
		for {
			cwb, err := NewCedarWebSocketBuffReader(nc)
			if err != nil {
				if debug {
					log.Println(err)
				}
				closeHj <- true
				break
			}
			fn(cwb)
		}
		<-closeHj
		nc.Close()
		close(closeHj)
		cedarWebsocketHub.Delete(key)
		log.Println("disconnect")
	}(nc)
}

func socketReplay(op int, data []byte) []byte {
	var frame = make([]byte, 0)
	bl := len(data)
	switch {
	case bl <= 125: // Payload length 7bits.
	case bl == 126: // Payload length 7+16bits

	case bl == 127: // Payload length 7+64bits
	}
	frame = append(frame, byte(0x1<<7+op))
	var f2 byte
	f2 |= 0
	lengthFields := 0
	length := len(data)
	switch {
	case length <= 125:
		f2 |= byte(length)
	case length < 65536:
		f2 |= 126
		lengthFields = 2
	default:
		f2 |= 127
		lengthFields = 8
	}
	frame = append(frame, f2)
	for i := 0; i < lengthFields; i++ {
		j := uint((lengthFields - i - 1) * 8)
		b := byte((length >> j) & 0xff)
		frame = append(frame, b)
	}
	frame = append(frame, data...)
	return frame
}

func WebsocketSwitchPush(key string, op int, data []byte) error {
	if nc, ok := cedarWebsocketHub.Load(key); ok {
		_, err := nc.(net.Conn).Write(socketReplay(op, data))
		if err != nil {
			return err
		}
		return nil
	} else {
		return fmt.Errorf("not find this key %s", key)
	}
}

func getNewKey(key string) string {
	return base64.StdEncoding.EncodeToString(GetSha1([]byte(key+MagicWebsocketKey), nil))
}

// cedarWebsocketBuffScan 快速读取json
// Scan usage *CedarWebSocketBuffReader.Scan
type cedarWebsocketBuffScan interface {
	Scan(v interface{}) error
}

// CedarWebSocketBuffReader 读取websocket协议,这里的websocket主要针对 4086byte 的文本格式
// Data 读取的文本载荷
// Length 文本[]byte长度
type CedarWebSocketBuffReader struct {
	Data   []byte
	Length int
	cedarWebsocketBuffScan
}

func NewCedarWebSocketBuffReader(nc net.Conn) (*CedarWebSocketBuffReader, error) {
	go func() {
		if err := recover(); err != nil {
			nc.Close()
			log.Println(err)
		}
	}()
	sbr := new(CedarWebSocketBuffReader)
	goto again
again:
	buf := bufio.NewReader(nc)
	var header []byte
	var b byte
	// First byte. FIN/RSV1/RSV2/RSV3/OpCode(4bits)
	b, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	header = append(header, b)
	fin := (header[0]>>7)&1 != 0
	if debug {
		log.Println("[cedar] websocket FIN", fin)
	}
	opcode := header[0] & 0x0f
	// replay opcode for ping
	switch opcode {
	case 0x8:
		return nil, fmt.Errorf("client close")
	case 0x9:
		socketReplay(0xA, []byte("pong"))
		return nil, nil
	}
	if debug {
		log.Println("[cedar] websocket OPCODE", opcode)
	}
	// Second byte. Mask/Payload len(7bits)
	b, err = buf.ReadByte()
	if err != nil {
		return nil, err
	}
	header = append(header, b)
	mask := (b & 0x80) != 0
	b &= 0x7f
	lengthFields := 0
	var headerLength int64 = 0
	switch {
	case b <= 125: // Payload length 7bits.
		headerLength = int64(b)
	case b == 126: // Payload length 7+16bits
		lengthFields = 2
	case b == 127: // Payload length 7+64bits
		lengthFields = 8
	}
	for i := 0; i < lengthFields; i++ {
		b, err = buf.ReadByte()
		if err != nil {
			return nil, err
		}
		if lengthFields == 8 && i == 0 { // MSB must be zero when 7+64 bits
			b &= 0x7f
		}
		header = append(header, b)
		headerLength = headerLength*256 + int64(b)
	}
	if debug {
		log.Println("[cedar] websocket Payload length", headerLength)
	}
	maskKey := make([]byte, 0)
	if mask {
		// Masking key. 4 bytes.
		for i := 0; i < 4; i++ {
			b, err = buf.ReadByte()
			if err != nil {
				return nil, err
			}
			header = append(header, b)
			maskKey = append(maskKey, b)
		}
	}
	// XorDecodeStr()
	payload := make([]byte, headerLength)
	kl := len(maskKey)
	if mask {
		for i := 0; i < len(payload); i++ {
			b, err = buf.ReadByte()
			payload[i] = b ^ maskKey[i%kl]
		}
	} else {
		for i := 0; i < len(payload); i++ {
			b, err = buf.ReadByte()
			if err != nil {
				break
			}
			payload[i] = b
		}
	}
	sbr.Data = append(sbr.Data, payload...)
	sbr.Length += len(payload)
	sbr.cedarWebsocketBuffScan = nil
	if !fin {
		goto again
	}
	return sbr, nil
}
func (sbr *CedarWebSocketBuffReader) Scan(v interface{}) error {
	if sbr.Length == 0 {
		return fmt.Errorf("data length is zero")
	}
	return jsoniter.Unmarshal(sbr.Data, v)
}

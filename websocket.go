package ultimate_cedar

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const MagicWebsocketKey = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

var cedarWebsocketHub = new(sync.Map)

// MaxKeys 最多保持多少个key 超过这个的数量
// 1. 最后访问时间最远的并且为空 将被移除掉
// 2. 超过10天的key 也将被移除
var MaxKeys uint64 = 2000

var KeyPc uint64 = 0
var MaxKeysMapping = make([]*KV, MaxKeys)
var mux sync.Mutex

type KV struct {
	Key        int
	Value      int64
	KeyOutside string
}

func (t *tree) SetWebsocketMaxKey(n uint64) {
	MaxKeys = n
}

const (
	OnlyPush = iota
	ReadPush
)

var bootModel int = ReadPush
var pointer uint64 = 0

// SetWebsocketModel default it can read and push
func (t *tree) SetWebsocketModel(model int) {
	bootModel = model
}

// MaxKeysSaveOrDelete 感觉是每次都触发
// 加锁和不加锁 会导致什么结果呢
func MaxKeysSaveOrDelete(key string) {
	mux.Lock()
	defer mux.Unlock()
	var isNotInMapping bool = true
	for _, kv := range MaxKeysMapping {
		if kv != nil && kv.KeyOutside == key {
			isNotInMapping = false
		}
	}
	if KeyPc >= MaxKeys && isNotInMapping {
		// 查找并移除
		// 排序后 得到时间最长的几组数据
		MaxKeysMapping = HeapSortSpecial(MaxKeysMapping)
		// 建立最大堆 ，首次进行填满 层数为4层
		var n = 1
		var now = time.Now().Unix()
		var day10 int64 = 3600 * 24 * 10
		for i := 0; i < len(MaxKeysMapping); i++ {
			if (now - MaxKeysMapping[i].Value) > day10 {
				n = i
			}
		}
		if n == 0 {
			n = 1
		}
		for _, kv := range MaxKeysMapping[:n] {
			cedarWebsocketHub.Delete(kv.KeyOutside)
		}
		MaxKeysMapping = MaxKeysMapping[n:]
		MaxKeysMapping = append(MaxKeysMapping, &KV{
			Key:        len(MaxKeysMapping) + 1,
			Value:      time.Now().Unix(),
			KeyOutside: key,
		})
		atomic.StoreUint64(&KeyPc, uint64(KeyPc-uint64(n)+1))
	} else {
		for i, kv := range MaxKeysMapping {
			if kv == nil {
				atomic.AddUint64(&KeyPc, 1)
				MaxKeysMapping[i] = &KV{
					Key:        i,
					Value:      time.Now().Unix(),
					KeyOutside: key,
				}
				break
			}
		}
	}
}

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
	MaxKeysSaveOrDelete(key)
	// 申请一个map
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
	for !atomic.CompareAndSwapUint64(&pointer, 0, 1) {
		time.Sleep(time.Millisecond * 10)
	}
	room, ok := cedarWebsocketHub.Load(key)
	if !ok {
		room2 := make(map[string]net.Conn)
		room2[nc.RemoteAddr().String()] = nc
		cedarWebsocketHub.Store(key, room2)
		room = room2
	}
	room.(map[string]net.Conn)[nc.RemoteAddr().String()] = nc
	atomic.CompareAndSwapUint64(&pointer, 1, 0)
	// cedarWebsocketHub.Store(key, nc)
	go func(nc net.Conn) {
		closeHj := make(chan bool)
		for {
			cwb, err := NewCedarWebSocketBuffReader(nc)
			if err != nil {
				if debug {
					log.Println("[cedar] websocket", err)
				}
				break
			}
			fn(cwb)
		}
		if debug {
			log.Println("[cedar] websocket close channel")
		}
		nc.Close()
		close(closeHj)
		delete(room.(map[string]net.Conn), nc.RemoteAddr().String())
		if debug {
			log.Println("[cedar] websocket disconnect")
		}
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
		for _, conn := range nc.(map[string]net.Conn) {
			con := conn
			go func(conn *net.Conn) {
				_, err := (*conn).Write(socketReplay(op, data))
				if err != nil {
					log.Println(err)
				}
			}(&con)
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
	// go func() {
	// 	if err := recover(); err != nil {
	// 		log.Println("[cedar] websocket recover error", err)
	// 	}
	// }()
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
	if bootModel == OnlyPush {
		switch opcode {
		case 0x8:
			return nil, fmt.Errorf("client close")
		case 0x9:
			socketReplay(0xA, []byte("pong"))
			return nil, nil
		default:
			return nil, nil
		}
	}
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
	return json.Unmarshal(sbr.Data, v)
}

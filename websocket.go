package ultimate_cedar

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"log"
	"net/http"
)

const MagicWebsocketKey = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

// 用来扩展websocket
// 只实现了保持在线和推送

// GET /chat HTTP/1.1
// Host: example.com:8000
// Upgrade: websocket
// Connection: Upgrade
// Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
// Sec-WebSocket-Version: 13
func switchProtocol(w http.ResponseWriter, r *http.Request) {
	version := r.Header.Get("Sec-Websocket-Version")
	log.Println("version is", version)
	if version != "13" {
		w.WriteHeader(400)
		return
	}
	key := r.Header.Get("Sec-WebSocket-Key")
	// 计算值
	newKey := getNewKey(key)
	w.Header().Add("Upgrade", "websocket")
	w.Header().Add("Connection", "Upgrade")
	w.Header().Add("Sec-Websocket-Accept", newKey)
	w.WriteHeader(http.StatusSwitchingProtocols)
	w.Write(nil)
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Not a Hijacker", 500)
		return
	}
	nc, _, err := hj.Hijack()
	if err != nil {
		log.Println(err)
	}
	closeHj := make(chan bool)
	go func() {
		for {
			buf := bufio.NewReader(nc)
			var header []byte
			var b byte
			// First byte. FIN/RSV1/RSV2/RSV3/OpCode(4bits)
			b, err = buf.ReadByte()
			if err != nil {
				return
			}
			header = append(header, b)
			fin := ((header[0] >> 7) & 1) != 0
			log.Println("FIN", fin)
			for i := 0; i < 3; i++ {
				j := uint(6 - i)
				log.Println("RSV", i, ((header[0]>>j)&1) != 0)
			}

			// 计算机的二进制骚操作 位运算
			// & | >> <<
			log.Println("OPCODE", header[0]&0x0f)

			// Second byte. Mask/Payload len(7bits)
			b, err = buf.ReadByte()
			if err != nil {
				return
			}
			header = append(header, b)
			mask := (b & 0x80) != 0
			log.Println("MASK", mask)
			b &= 0x7f
			lengthFields := 0
			switch {
			case b <= 125: // Payload length 7bits.
				log.Println("PLAYLOAD lENGTH 7 BITS", int64(b))
			case b == 126: // Payload length 7+16bits
				log.Println("PLAYLOAD lENGTH 7+16 BITS", int64(b))
				lengthFields = 2
			case b == 127: // Payload length 7+64bits
				log.Println("PLAYLOAD lENGTH 7+64 BITS", int64(b))
				lengthFields = 8
			}
			var headerLength int64 = 0
			log.Println("LENGTH FIEDLDS", lengthFields)
			for i := 0; i < lengthFields; i++ {
				b, err = buf.ReadByte()
				if err != nil {
					return
				}
				if lengthFields == 8 && i == 0 { // MSB must be zero when 7+64 bits
					b &= 0x7f
				}
				header = append(header, b)
				headerLength = headerLength*256 + int64(b)
			}
			log.Println("HEADER LENGTH", headerLength)
			maskKey := make([]byte, 0)
			if mask {
				// Masking key. 4 bytes.
				for i := 0; i < 4; i++ {
					b, err = buf.ReadByte()
					if err != nil {
						return
					}
					header = append(header, b)
					maskKey = append(maskKey, b)
				}
			}
			out := make([]byte, 0)
			log.Println(buf.Size(), len(header))
			log.Println(buf.ReadBytes(2))
			log.Println("1", out)
			log.Println(string(bytes.NewBuffer(out).String()))
		}
	}()
	<-closeHj
	log.Println("close connection")
	nc.Close()
}

func getNewKey(key string) string {
	return base64.StdEncoding.EncodeToString(GetSha1([]byte(key+MagicWebsocketKey), nil))
}

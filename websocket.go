package ultimate_cedar

import (
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
			data := make([]byte, 1024)
			n, err := nc.Read(data)
			if err != nil {
				log.Println(err)
				closeHj <- true
				break
			}
			log.Println("read data", string(data[:n]))
		}
	}()
	<-closeHj
	log.Println("close connection")
	nc.Close()
}

func getNewKey(key string) string {
	return base64.StdEncoding.EncodeToString(GetSha1([]byte(key+MagicWebsocketKey), nil))
}

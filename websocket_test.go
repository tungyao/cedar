package ultimate_cedar

import (
	"log"
	"net/http"
	"testing"
)

// 标准测试 在
// READ 913 byte 10次 并发下表现良好
func Test_switchProtocol(t *testing.T) {
	r := NewRouter()
	// r.Debug()
	r.Get("/ws", func(writer ResponseWriter, request Request) {
		WebsocketSwitchProtocol(writer, request, "123", func(value *CedarWebSocketBuffReader) {
			log.Println(string(value.Data))
		})
	})
	r.Post("/ws/push", func(writer ResponseWriter, request Request) {
		WebsocketSwitchPush("123", 0x1, []byte("hello world"))
	})
	http.ListenAndServe(":8080", r)
}

func Test_getNewKey(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{name: "test new key", args: args{key: "dGhlIHNhbXBsZSBub25jZQ=="}, want: "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getNewKey(tt.args.key); got != tt.want {
				t.Errorf("getNewKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestByte(t *testing.T) {
	// t.Log(byte(0x1<<7 + 0x1))
	// 356
	// 0000100
	b := 356
	b &= 0x7f
	var headerLength int64 = 0
	for i := 0; i < 2; i++ {
		headerLength = headerLength*256 + int64(b)
	}
	t.Log(headerLength)
}

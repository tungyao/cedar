package ultimate_cedar

import (
	"net/http"
	"testing"
)

func Test_switchProtocol(t *testing.T) {

	http.HandleFunc("/ws", switchProtocol)
	http.ListenAndServe(":8000", nil)
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

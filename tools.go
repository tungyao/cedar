package ultimate_cedar

import (
	"crypto/sha1"
)

func inArrayString(target string, srcArr []string) bool {
	for _, v := range srcArr {
		if v == target {
			return true
		}
	}
	return false
}
func GetSha1(data []byte, mix []byte) []byte {
	sha := sha1.New()
	sha.Write(data)
	return sha.Sum(mix)
}

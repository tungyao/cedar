package ultimate_cedar

import (
	"crypto/sha1"
	"time"
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

func HeapSortSpecial(arr []*KV) []*KV {
	length := len(arr)
	buildMaxHeap(arr, length)
	for i := 0; i < len(arr)-1; i++ {
		arr[i], arr[0] = arr[0], arr[i]
		length--
		heapify(arr, 0, length)
	}
	return arr
}
func buildMaxHeap(arr []*KV, arrLen int) {
	for i := arrLen / 2; i >= 0; i-- {
		heapify(arr, i, arrLen)
	}
}
func heapify(arr []*KV, i int, leng int) {
	var ut = time.Now().Unix()
	left, right, largest := 2*i+1, 2*i+2, i
	if left < leng && ut-arr[left].Value > ut-arr[largest].Value {
		largest = left
	}
	if right < leng && ut-arr[right].Value > ut-arr[largest].Value {
		largest = right
	}
	if largest != i {
		arr[i], arr[largest] = arr[largest], arr[i]
		heapify(arr, largest, leng)
	}
}

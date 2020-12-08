package utils

import (
	"fmt"
	"github.com/shopspring/decimal"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func GetUinxMillisecond() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func TypeOf(v interface{}) string {
	switch v.(type) {
	case int:
		return "int"
	case float64:
		return "float64"
	case string:
		return "string"
	default:
		return "unknown"
	}
}

var letterRunes = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandomString(n uint) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func BytesToStr(v []byte) string {
	return string(v[:])
}

func ToStr(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

func ToDecimal(v interface{}) decimal.Decimal {
	return decimal.RequireFromString(ToStr(v))
}

func ToInt(v interface{}) int {
	vStr := ToStr(v)
	ret, err := strconv.Atoi(vStr)
	if err != nil {
		return 0
	}
	return ret
}

func ToMap(v interface{}) map[string]interface{} {
	if res, ok := v.(map[string]interface{}); ok {
		return res
	}

	return nil
}

func DelFile(path string) error {
	return os.Remove(path)
}

func AppendFile(path string, data string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		err2 := os.MkdirAll(filepath.Dir(path), os.ModePerm)
		if err2 != nil {
			log.Print(err2)
			return err
		}

		f, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
	}
	defer f.Close()

	if _, err = f.WriteString(data); err != nil {
		return err
	}

	return nil
}

func MsToTime(msInt int) time.Time {
	tm := time.Unix(0, int64(msInt)*int64(time.Millisecond))
	return tm
}

func StrToBytes(str string) (ret []byte) {
	for _, v := range str {
		ret = append(ret, byte(v))
	}

	return ret
}

func ExistsFile(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}

	return true
}

func Chunk(actions []string, batchSize int) [][]string {
	var batches [][]string

	for batchSize < len(actions) {
		actions, batches = actions[batchSize:], append(batches, actions[0:batchSize:batchSize])
	}
	batches = append(batches, actions)

	return batches
}

package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

func CurrentTimeString() string {
	return time.Now().Format(time.RFC3339)
}

func ContainsString(arr []string, str string) bool {
	for _, s := range arr {
		if str == s {
			return true
		}
	}
	return false
}

func ContainsUUID(arr []uuid.UUID, id uuid.UUID) bool {
	for _, i := range arr {
		if i == id {
			return true
		}
	}
	return false
}

func IsTruthyString(str string) bool {
	truthyStrings := []string{"1", "t", "true"}
	return ContainsString(truthyStrings, str)
}

func IsFalseyPtr(val *bool) bool {
	return val == nil || !*val
}

func FirstNonEmpty(first string, second string) string {
	if first != "" {
		return first
	}
	return second
}

func Substring(str string, start int, end int) string {
	rune := []rune(str)
	return string(rune[start:end])
}

func EnvOrDefault(varName string, defaultVal string) string {
	if val, ok := os.LookupEnv(varName); ok {
		return val
	}
	return defaultVal
}

func PrettyPrintJson(obj interface{}) {
	objJSON, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("\n%s\n", string(objJSON))
}

func NewTruePtr() *bool {
	b := true
	return &b
}

func NewFalsePtr() *bool {
	b := false
	return &b
}

func NewUIntPtr(val uint) *uint {
	i := val
	return &i
}

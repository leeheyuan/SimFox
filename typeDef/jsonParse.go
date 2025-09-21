package typeDef

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type JsonParse struct {
}

func (This *JsonParse) ParseUrl(url string, obj interface{}) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("error making GET request: %v", err)
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(obj)
	if err != nil {
		fmt.Println("Error decoding XML:", err)
		return
	}
}

func (This *JsonParse) ParseFile(url string, obj interface{}) {
	file, err := os.Open(url)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	// 创建 XML 解码器
	decoder := json.NewDecoder(file)
	err = decoder.Decode(obj)
	if err != nil {
		fmt.Println("Error decoding XML:", err)
		return
	}
}

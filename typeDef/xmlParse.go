package typeDef

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"utils"
)

type XmlParse struct {
}

func (This *XmlParse) ParseUrl(url string, obj interface{}) {

	if utils.IsValidURL(url) {
		resp, err := http.Get(url)
		if err != nil {
			log.Fatalf("error making GET request: %v", err)
		}
		defer resp.Body.Close()
		decoder := xml.NewDecoder(resp.Body)
		err = decoder.Decode(obj)
		if err != nil {
			fmt.Println("Error decoding XML:", err)
			return
		}
	} else {
		This.ParseFile(url, obj)
	}

}

func (This *XmlParse) ParseFile(url string, obj interface{}) {
	file, err := os.Open(url)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	// 创建 XML 解码器
	decoder := xml.NewDecoder(file)
	err = decoder.Decode(obj)
	if err != nil {
		fmt.Println("Error decoding XML:", err)
		return
	}
}

package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

const symbols = "23456789BCDFGHJKMNPQRTVWXY"

var base64encoding = base64.StdEncoding

func main() {
	r := setupRouter()
	err := r.Run(":8080")
	if err != nil {
		fmt.Println("Failed to start the server:", err)
	}
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.POST("/generate-steamguard-code", func(c *gin.Context) {
		sharedSecret := c.PostForm("sharedSecret")
		code, err := GenerateSteamGuardCode(sharedSecret)
		if err != nil {
			c.HTML(http.StatusBadRequest, "index.html", gin.H{"error": err.Error()})
		} else {
			c.HTML(http.StatusOK, "index.html", gin.H{"code": code})
		}
	})

	return r
}

func GenerateSteamGuardCode(sharedSecret string) (string, error) {
	key, err := base64encoding.DecodeString(sharedSecret)
	if err != nil {
		return "", err
	}

	timestamp := time.Now().Unix() / 30
	h := hmac.New(sha1.New, key)
	err = binary.Write(h, binary.BigEndian, uint64(timestamp))
	if err != nil {
		return "", err
	}

	hash := h.Sum(nil)
	offset := int(hash[19] & 0x0F)
	truncatedHash := int(hash[offset]&0x7F)<<24 |
		int(hash[offset+1]&0xFF)<<16 |
		int(hash[offset+2]&0xFF)<<8 |
		int(hash[offset+3]&0xFF)
	codeIndex := truncatedHash % len(symbols)

	code := make([]byte, 5)
	for i := 0; i < 5; i++ {
		code[i] = symbols[codeIndex]
		truncatedHash /= len(symbols)
		codeIndex = truncatedHash % len(symbols)
	}

	return string(code), nil
}

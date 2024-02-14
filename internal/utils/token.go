package utils

import (
	"encoding/hex"
	"log"
	"math/rand"
	"strconv"
	"time"
)

/*
* Random OTP Generator
 */
func GenerateToken(length int) string {
	rand.Seed(time.Now().UnixNano())
	var otp string
	for i := 0; i < length; i++ {
		otp += strconv.Itoa(rand.Intn(10))
	}
	return otp
}

func GenerateRandomHexString(byteLength int) string {
	randomBytes := make([]byte, byteLength)
	_, err := rand.Read(randomBytes)
	if err != nil {
		// Log the error and return an indicative or empty string
		log.Printf("Error generating random bytes: %v", err)
		return ""
	}

	return hex.EncodeToString(randomBytes)
}

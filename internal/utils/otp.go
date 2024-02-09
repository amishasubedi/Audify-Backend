package utils

import (
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

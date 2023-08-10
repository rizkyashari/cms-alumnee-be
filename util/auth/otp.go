package auth

import (
	"crypto/rand"
	"log"
	"os"
	"strconv"
	"time"
)

type OTP struct {
	Password  string
	ExpiresAt time.Time
}

func GenerateOTP() (*OTP, error) {
	const otpChars = "1234567890"
	strLen := os.Getenv("OTP_LEN")

	length, err := strconv.Atoi(strLen)
	if err != nil {
		return nil, err
	}

	durationMin := os.Getenv("OTP_DURATION_MIN")
	intDurationMin, err := strconv.Atoi(durationMin)
	if err != nil {
		log.Fatal(err.Error())
	}

	otpExpiresAt := time.Now().Add(time.Duration(intDurationMin) * time.Minute)

	buffer := make([]byte, length)
	_, err = rand.Read(buffer)
	if err != nil {
		return nil, err
	}

	otpCharsLength := len(otpChars)
	for i := 0; i < length; i++ {
		buffer[i] = otpChars[int(buffer[i])%otpCharsLength]
	}

	return &OTP{Password: string(buffer), ExpiresAt: otpExpiresAt}, nil
}

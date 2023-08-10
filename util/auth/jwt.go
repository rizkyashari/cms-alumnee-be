package auth

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type CustomClaim struct {
	Email string
	jwt.RegisteredClaims
}

type Jwt struct {
	Token     string
	ExpiresAt time.Time
}

func getSecret() string {
	var secret = os.Getenv("JWT_SECRET")
	if len(secret) == 0 {
		return "no_secret"
	}
	return secret
}

func GenerateJWT(email string) (*Jwt, error) {
	secret := getSecret()

	durationMin := os.Getenv("JWT_DURATION_MIN")
	intDurationMin, err := strconv.Atoi(durationMin)
	if err != nil {
		log.Fatal(err.Error())
	}

	jwtDuration := time.Duration(intDurationMin) * time.Minute

	now := time.Now()

	claims := CustomClaim{
		email,
		jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(jwtDuration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	Jwt := Jwt{Token: ss, ExpiresAt: now.Add(jwtDuration)}

	return &Jwt, nil
}

func CheckToken(input string) (string, error) {
	secret := getSecret()

	token, err := jwt.ParseWithClaims(input, &CustomClaim{}, func(tkn *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if claims, ok := token.Claims.(*CustomClaim); ok && token.Valid {
		return claims.Email, nil
	} else {
		return "", err
	}
}

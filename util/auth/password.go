package auth

import "golang.org/x/crypto/bcrypt"

func HashAndSalt(passwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	return string(hash), err
}

func ComparePassword(hashpasswd string, input string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashpasswd), []byte(input))
	return err == nil
}

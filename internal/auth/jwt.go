package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// generate token creates a signed tken that identifies one user for a specified active time
func GenerateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,                                  //who the token is for
		"exp": time.Now().Add(30 * time.Minute).Unix(), //time of expiration

	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

//after creating a token you parse the token
//parse token checks the signature and expiry, and returns the user id inside

func ParseToken(tokenStr string) (int, error) {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"HS256"})) //REFUSE TOKENS SIGNES ANOTHER WAY

	token, err := parser.Parse(tokenStr,
		func(t *jwt.Token) (any, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

	if err != nil || token == nil || !token.Valid {
		return 0, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, ErrInvalidClaims
	}

	sub, ok := claims["sub"].(float64)
	if !ok {
		return 0, ErrInvalidTokenSubject
	}
	return int(sub), nil
}

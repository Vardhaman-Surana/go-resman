package encryption

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/vds/go-resman/pkg/logger"
	"github.com/vds/go-resman/pkg/models"
	"time"
)

func CreateToken(ctx context.Context, claims *models.Claims) (string, error) {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	jwtKey := []byte("SecretKey")

	logger.LogDebug(reqId, reqUrl, "generating jwt token")

	expirationTime := time.Now().Add(120 * time.Minute).Unix()
	claims.ExpiresAt = expirationTime
	//remember to change it later
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	logger.LogInfo(reqId, reqUrl, "jwt token generated", 0)
	return tokenString, nil
}

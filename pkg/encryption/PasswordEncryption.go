package encryption

import (
	"context"
	"errors"
	"github.com/vds/go-resman/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

//errors
var errGenHash = errors.New("error in generating hash for email id")

func GenerateHash(value string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
	if err != nil {
		return "", errGenHash
	}
	return string(hash), nil
}
func ComparePasswords(ctx context.Context, phash, pass string) bool {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogDebug(reqId, reqUrl, "password verification")

	err := bcrypt.CompareHashAndPassword([]byte(phash), []byte(pass))
	if err != nil {
		logger.LogInfo(reqId, reqUrl, "password does not match", 0)
		return false
	}
	logger.LogInfo(reqId, reqUrl, "password match", 0)
	return true
}

package encryption

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/vds/go-resman/pkg/models"
	"time"
)

func CreateToken(claims *models.Claims) (string,error){
	jwtKey:=[]byte("SecretKey")
	expirationTime:=time.Now().Add(120*time.Minute).Unix()
	claims.ExpiresAt=expirationTime
	//remember to change it later
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err!=nil{
		return "",err
	}
	return tokenString,nil
}
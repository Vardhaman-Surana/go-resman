package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/vds/go-resman/pkg/database"
	"github.com/vds/go-resman/pkg/encryption"
	"github.com/vds/go-resman/pkg/middleware"
	"github.com/vds/go-resman/pkg/models"
	"log"
	"net/http"
	"regexp"
)

const(
	Success="Success"
	Fail="Fail"
)

type LogInController struct{
	database.Database
}

func NewLogInController(db database.Database)*LogInController{
	lc:=new(LogInController)
	lc.Database=db
	return lc
}
func(l *LogInController)LogIn(c *gin.Context){
	var cred models.Credentials
	err:=c.ShouldBindJSON(&cred)
	if err!=nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	isValid:=middleware.IsValidUserType(cred.Role)
	if !isValid{
		c.Status(http.StatusNotFound)
		return
	}
	if !re.MatchString(cred.Email){
		c.JSON(http.StatusBadRequest, gin.H{
			"error":"Invalid Email",
			"status":Fail,
		})
		return
	}
	userID,err:=l.LogInUser(&cred)
	if err!=nil{
		fmt.Print(err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
			"status": Fail,
		})
		return
	}
	claims:=&models.Claims{
		ID:userID,
		Role:cred.Role,
	}
	token,err:=encryption.CreateToken(claims)
	if err!=nil{
		log.Printf("%v",err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"token":token,
		"role":cred.Role,
		"msg":"Login Successful",
		"status":Success,
	})
}

func(l *LogInController)LogOut(c *gin.Context){
	tokenStr:=c.Request.Header.Get("token")
	if tokenStr==""{
		c.Status(http.StatusBadRequest)
		return
	}
	err:=l.StoreToken(tokenStr)
	if err!=nil{
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK,gin.H{
	"msg":"Logged Out Successfully",
	"status":Success,
	})
}



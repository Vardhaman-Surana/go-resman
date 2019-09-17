package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/vds/go-resman/pkg/database"
	"github.com/vds/go-resman/pkg/middleware"
	"github.com/vds/go-resman/pkg/models"
	"net/http"
	"regexp"
)
type RegisterController struct{
	database.Database
}

func NewRegisterController(db database.Database) *RegisterController{
	regController:=new(RegisterController)
	regController.Database=db
	return regController
}

func(r *RegisterController)Register(c *gin.Context){
	var user models.UserReg
	err:=c.ShouldBindJSON(&user)
	if err!=nil {
		fmt.Printf("err is %v",err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"status":Fail,
		})
		return
	}
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if user.Role!=middleware.Admin && user.Role!=middleware.SuperAdmin{
		c.Status(http.StatusNotFound)
		return
	}
	if !re.MatchString(user.Email){
		c.JSON(http.StatusBadRequest, gin.H{
			"error":"Invalid Email",
			"status":Fail,
		})
		return
	}
	err=r.CreateUser(&user)
	if err!=nil{
		if err==database.ErrDupEmail{
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
				"status":Fail,
			})
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"role":user.Role,
		"msg":"Registration Successful",
		"status":Success,
	})
}
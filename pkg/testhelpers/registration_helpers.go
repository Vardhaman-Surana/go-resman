package testhelpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vds/go-resman/pkg/database/mysql"
	"github.com/vds/go-resman/pkg/models"
	"net/http"
)

var (
	AdminToRegister = models.UserReg{
		Role:     "admin",
		Email:    "admin1@gmail.com",
		Name:     "admin1",
		Password: "pass1",
	}
	SuperAdminToRegister = models.UserReg{
		Role:     "superAdmin",
		Email:    "superAdmin1@gmail.com",
		Name:     "superAdmin1",
		Password: "superPass1",
	}
)


func NewRegisterRequest(user *models.UserReg,baseUrl string) (*http.Request,error) {
	data, err := json.Marshal(user)
	if err != nil {
		return nil,err
	}
	body:= bytes.NewBuffer(data)
	request, err := http.NewRequest("POST", baseUrl+"/register", body)
	if err != nil {
		return nil,err
	}
	return request,nil
}

func ClearRegisteredUsers(){
	_, err := Db.Exec(fmt.Sprintf("delete from %s where email_id=?", mysql.AdminTable),AdminToRegister.Email)
	if err != nil {
		panic(err)
	}
	_, err = Db.Exec(fmt.Sprintf("delete from %s where email_id=?", mysql.SuperAdminTable),SuperAdminToRegister.Email)
	if err != nil {
		panic(err)
	}
}





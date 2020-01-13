package testhelpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vds/go-resman/pkg/models"
	"net/http"
)

func AdminCredentials()*models.Credentials{
	return &models.Credentials{
		Role:     admin.Role,
		Email:    admin.Email,
		Password: admin.Password,
	}
}

func SuperAdminCredentials()*models.Credentials{
	return &models.Credentials{
		Role:     superAdmin.Role,
		Email:    superAdmin.Email,
		Password: superAdmin.Password,
	}
}

func OwnerByAdminCredentials()*models.Credentials{
	return &models.Credentials{
		Role:     ownerByAdmin.Role,
		Email:    ownerByAdmin.Email,
		Password: ownerByAdmin.Password,
	}
}
func OwnerBySuperAdminCredentials()*models.Credentials{
	return &models.Credentials{
		Role:     ownerBySuperAdmin.Role,
		Email:    ownerBySuperAdmin.Email,
		Password: ownerBySuperAdmin.Password,
	}
}

func NewLogInRequest(credentials *models.Credentials,baseUrl string) (*http.Request,error) {
	data, err := json.Marshal(credentials)
	if err != nil {
		return nil,err
	}
	body:= bytes.NewBuffer(data)
	req, err := http.NewRequest(http.MethodPost, baseUrl+"/login",body)
	if err != nil {
		return nil,err
	}
	req.Header.Add("Content-Type", "application/json")
	return req,nil
}

func NewLogOutRequest(token string,baseUrl string) (*http.Request,error) {
	req, err := http.NewRequest(http.MethodGet, baseUrl+"/logout", nil)
	if err!=nil{
		return nil,err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req,nil
}


func ClearInvalidTokens(){
	_, err := Db.Exec(fmt.Sprintf("delete from %s", InvalidTokenTable))
	if err != nil {
		panic(err)
	}
}

package testhelpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vds/go-resman/pkg/models"
	"net/http"
	"reflect"
	"testing"
)

func GetAdminForUD()*models.UserOutput{
	return &models.UserOutput{
		ID:    adminForUDId,
		Email: adminForUD.Email,
		Name:  adminForUD.Name,
	}
}

func NewGetAdminRequest(token string,baseUrl string)(*http.Request,error) {
	req, err := http.NewRequest(http.MethodGet, baseUrl+"/manage/admins", nil)
	if err != nil {
		return nil,err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req,nil
}

func NewUpdateAdminRequest(token string,admin *models.UserOutput,baseUrl string)(*http.Request,error){
	data, err := json.Marshal(admin)
	if err != nil {
		return nil,err
	}
	body:= bytes.NewBuffer(data)
	req, err := http.NewRequest(http.MethodPut, baseUrl+fmt.Sprintf("/manage/admins/%s", admin.ID),body)
	if err != nil {
		return nil,err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req,nil
}

func AssertAdmin(t *testing.T,got *models.UserOutput,want *models.UserOutput){
	t.Helper()
	if !reflect.DeepEqual(got,want){
		t.Fatalf("want %v got %v",want,got)
	}
}


func NewDeleteAdminRequest(token string, id string,baseUrl string) (*http.Request,error){
	req, err := http.NewRequest(http.MethodDelete, baseUrl+fmt.Sprintf("/manage/admins?id=%v",id), nil)
	if err != nil {
		return nil,err
	}
	req.Header.Set("token",token)
	return req,nil
}


func AssertAdmins(t *testing.T,got []models.UserOutput,want []models.UserOutput){
	t.Helper()
	if !reflect.DeepEqual(got,want){
		t.Fatalf("want %v got %v",want,got)
	}
}

func GetAdmins()(admins []models.UserOutput){
	return []models.UserOutput{
		{ID: adminId,Email: admin.Email,Name: admin.Name},
	}
}


package testhelpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vds/go-resman/pkg/models"
	"net/http"
	"reflect"
	"sort"
	"testing"
)

func NewCreateOwnerRequest(token string,owner *models.OwnerReg,baseUrl string)(*http.Request,error){
	data, err := json.Marshal(owner)
	if err != nil {
		return nil,err
	}
	body:= bytes.NewBuffer(data)
	req, err := http.NewRequest(http.MethodPost, baseUrl+"/manage/owners",body)
	if err != nil {
		return nil,err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req,nil
}

func AssertOwner(t *testing.T,got *models.UserOutput,want *models.UserOutput){
	t.Helper()
	if !reflect.DeepEqual(got,want){
		t.Fatalf("want %v got %v",want,got)
	}
}

func NewUpdateOwnerRequest(token string,owner *models.UserOutput,baseUrl string)(*http.Request,error){
	data, err := json.Marshal(owner)
	if err != nil {
		return nil,err
	}
	body:= bytes.NewBuffer(data)
	req, err := http.NewRequest(http.MethodPut, baseUrl+fmt.Sprintf("/manage/owners/%s", owner.ID),body)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req,nil
}

func NewDeleteOwnerRequest(token string,id string,baseUrl string)(*http.Request,error){
	req, err := http.NewRequest(http.MethodDelete, baseUrl+fmt.Sprintf("/manage/owners?id=%s",id),nil)
	if err!=nil{
		return nil,err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req,nil
}

func GetAllOwners()[]models.UserOutput{

	owners:= []models.UserOutput{
		{ID:ownerByAdminId,Name:ownerByAdmin.Name,Email:ownerByAdmin.Email},
		{ID:ownerBySuperAdminId,Name:ownerBySuperAdmin.Name,Email:ownerBySuperAdmin.Email},
	}
	sort.Slice(owners, func(i, j int) bool {
		return owners[i].ID < owners[j].ID
	})
	return owners
}

func GetOwnersByAdmin()[]models.UserOutput{
	return []models.UserOutput{
		{ID:ownerByAdminId,Name:ownerByAdmin.Name,Email:ownerByAdmin.Email},
	}
}

func NewGetOwnerRequest(token string,baseUrl string)(*http.Request,error){
	req, err := http.NewRequest(http.MethodGet, baseUrl+"/manage/owners",nil)
	if err!=nil{
		return nil,err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req,nil
}

func AssertOwners(t *testing.T,got []models.UserOutput,want []models.UserOutput){
	t.Helper()
	if !reflect.DeepEqual(got,want){
		t.Fatalf("want %v got %v",want,got)
	}
}

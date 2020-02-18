package main

import (
	"encoding/json"
	"github.com/vds/go-resman/pkg/middleware"
	"github.com/vds/go-resman/pkg/models"
	"github.com/vds/go-resman/pkg/testhelpers"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestAdmins(t *testing.T){
	adminForUD:= testhelpers.GetAdminForUD()
	adminForUD.Name = "nameUpdated"
	superAdminToken,err := testhelpers.GetSuperAdminToken(serverUrl)
	if err!=nil{
		t.Fatalf("error in getting superAdmin token:%v",err)
	}

	testUpdateAdmin := []struct {
		name       string
		admin      *models.UserOutput
		wantedStatus int
	}{
		{name: "update admin with empty field",admin: &models.UserOutput{Email:"email",Name:"",ID:"invalidAdminID"},wantedStatus: http.StatusBadRequest},
		{name: "update non existing admin",admin:&models.UserOutput{Email:"email123",Name:"name123",ID:"id123"},wantedStatus:http.StatusBadRequest},
		{name: "update existing admin",admin: adminForUD,wantedStatus: http.StatusOK},
	}
	for _, test := range testUpdateAdmin {
		t.Run(test.name, func(t *testing.T) {
			request,err := testhelpers.NewUpdateAdminRequest(superAdminToken,test.admin,serverUrl)
			if err!=nil{
				t.Fatalf("unable to create request:%v",err)
			}
			resp, err := http.DefaultClient.Do(request)
			if err!=nil{
				t.Fatalf("http request failed:%v",err)
			}
			testhelpers.AssertStatus(t,resp.StatusCode,test.wantedStatus)
			if test.wantedStatus == 200{
				var updatedAdmin models.UserOutput
				body,_ := ioutil.ReadAll(resp.Body)
				err:=json.Unmarshal(body,&updatedAdmin)
				if err!=nil{
					t.Fatalf("response not in correct format:%v",err)
				}
				testhelpers.AssertAdmin(t,&updatedAdmin,adminForUD)
			}
		})
	}


	testDeleteAdmin := []struct {
		name       string
		id      	string
		wantedStatus int
	}{
		{name:"tests for valid deletion",id:adminForUD.ID, wantedStatus: http.StatusOK},
		{name:"empty array of id",id:"",wantedStatus:http.StatusBadRequest},
		{name:"try to delete invalid id", id:"invalid",wantedStatus: http.StatusBadRequest},
	}
	for _, test := range testDeleteAdmin {
		t.Run(test.name, func(t *testing.T) {
			request,err := testhelpers.NewDeleteAdminRequest(superAdminToken, test.id,serverUrl)
			if err!=nil{
				t.Fatalf("http request failed:%v",err)
			}
			resp, err := http.DefaultClient.Do(request)
			if err!=nil{
				t.Fatalf("http request failed:%v",err)
			}
			testhelpers.AssertStatus(t,resp.StatusCode,test.wantedStatus)
		})
	}


	tests := []struct {
		name       string
		token 		string
		wantedStatus int
	}{
		{name:"request with invalid token",token:"token",wantedStatus:middleware.StatusTokenInvalid},
		{name:"request with valid token",token:superAdminToken,wantedStatus:http.StatusOK},
	}

	admins:= testhelpers.GetAdmins()

	for _, test := range tests {
		t.Run(test.name,func(t *testing.T){
			request,err := testhelpers.NewGetAdminRequest(test.token,serverUrl)
			if err!=nil{
				t.Fatalf("can not generate registration request:%v",err)
			}
			resp, err := http.DefaultClient.Do(request)
			if err!=nil{
				t.Fatalf("http request failed:%v",err)
			}
			testhelpers.AssertStatus(t, resp.StatusCode,test.wantedStatus)
			if resp.StatusCode == http.StatusOK{
				var gotAdmins []models.UserOutput
				body,_ := ioutil.ReadAll(resp.Body)
				err:=json.Unmarshal(body,&gotAdmins)
				if err!=nil{
					t.Fatalf("response is not in appropriate format:%v",err)
				}
				testhelpers.AssertAdmins(t,gotAdmins,admins)
			}
		})
	}
}



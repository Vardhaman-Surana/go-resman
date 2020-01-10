package main

import (
	"github.com/vds/go-resman/pkg/models"
	"github.com/vds/go-resman/pkg/testhelpers"
	"net/http"
	"testing"
)

func TestRegistration(t *testing.T){
	tests := []struct {
		name       string
		user 	   *models.UserReg
		wantedStatus 		int
	}{
		{name:"When an admin is successfully created",user:&testhelpers.AdminToRegister,wantedStatus: http.StatusOK},
		{name:"When a superAdmin is successfully created",user: &testhelpers.SuperAdminToRegister,wantedStatus: http.StatusOK},
		{name:"duplicate mail for admin",user: &testhelpers.AdminToRegister,wantedStatus: http.StatusBadRequest},
		{name:"duplicate mail for super admin",user: &testhelpers.SuperAdminToRegister,wantedStatus: http.StatusBadRequest},
		{name:"Empty Require Field",user: &models.UserReg{  Role:"admin",Email: "",Name: "admin1",Password: "pass1"},wantedStatus: http.StatusBadRequest},
		{name:"Registration with invalid role",user: &models.UserReg{  Role:"otherRole",Email: "a@gmail.com",Name: "admin1",Password: "pass1"},wantedStatus: http.StatusNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request,err := testhelpers.NewRegisterRequest(test.user,serverUrl)
			if err!=nil{
				t.Fatalf("can not generate registration request:%v",err)
			}
			resp, err := http.DefaultClient.Do(request)
			if err!=nil{
				t.Fatalf("http request failed:%v",err)
			}
			testhelpers.AssertStatus(t, resp.StatusCode,test.wantedStatus)
		})
	}
	testhelpers.ClearRegisteredUsers()
}

package main

import (
	"github.com/vds/go-resman/pkg/middleware"
	"github.com/vds/go-resman/pkg/models"
	"github.com/vds/go-resman/pkg/testhelpers"
	"net/http"
	"testing"
)

func TestLogInController(t *testing.T) {
	tests := []struct {
		name       string
		credentials *models.Credentials
		wantedStatus int
	}{
		{name: "When an admin is successfully logged in",credentials: testhelpers.AdminCredentials(),wantedStatus: http.StatusOK},
		{name: "When a superAdmin is successfully logged in",credentials: testhelpers.SuperAdminCredentials(),wantedStatus: http.StatusOK},
		{name: "When an owner is successfully logged in",credentials: testhelpers.OwnerCredentials(),wantedStatus: http.StatusOK},
		{name: "SuperAdmin with invalid credentials",credentials: &models.Credentials{ Role: middleware.Admin,Email:"a@email.com" ,Password: "dummySuperPa"},wantedStatus: http.StatusUnauthorized},
		{name: "with empty fields",credentials: &models.Credentials{ Role:"", Email:"",Password: ""}, wantedStatus: http.StatusBadRequest},
		{name: "For invalid role",credentials: &models.Credentials{ Role:"invalidRole",Email: "email",Password: "pass"},wantedStatus: http.StatusBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request,err := testhelpers.NewLogInRequest(test.credentials,serverUrl)
			if err!=nil{
				t.Fatalf("can not create http request:%v",err)
			}
			resp, err := http.DefaultClient.Do(request)
			if err!=nil{
				t.Fatalf("http request failed:%v",err)
			}
			testhelpers.AssertStatus(t, resp.StatusCode,test.wantedStatus)
		})
	}
	token,err := testhelpers.GetSuperAdminToken(serverUrl)
	if err!=nil{
		t.Fatalf("error in getting superAdmin token:%v",err)
	}
	///tests for logout
	testLogout := []struct {
		name       string
		token      string
		wantedStatus int
	}{
		{"when request is made with token", token, http.StatusOK},
		{"when token is not sent", "", http.StatusBadRequest},
	}
	for _, test := range testLogout {
		t.Run(test.name, func(t *testing.T) {
			request,err := testhelpers.NewLogOutRequest(test.token,serverUrl)
			if err!=nil{
				t.Fatalf("can not create http request:%v",err)
			}
			resp, err := http.DefaultClient.Do(request)
			if err!=nil{
				t.Fatalf("http request failed:%v",err)
			}
			testhelpers.AssertStatus(t, resp.StatusCode,test.wantedStatus)
		})
	}
	testhelpers.ClearInvalidTokens()
}

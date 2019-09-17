package controller_test

import (
	"encoding/json"
	"github.com/vds/go-resman/pkg/database/mysql"
	"github.com/vds/go-resman/pkg/models"
	"github.com/vds/go-resman/pkg/server"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	_ "testing"
)
var dummyAdmin=models.UserReg{"admin","dummyAdmin@gmail.com","dummyAdmin","dummyPass"}
var dummySuperAdmin=models.UserReg{"superAdmin","dummySuperAdmin@gmail.com","dummySuperAdmin","dummySuperPass"}
var dummyOwner=models.OwnerReg{"dummySuperOwner@gmail.com","dummySuperOwner","dummyOwnerPass"}
func TestLogInController(t *testing.T){
	var DB, _ = mysql.NewMySqlDB("restaurant_test")
	defer DB.Close()
	defer CleanDB(DB)
	svr,err:=server.NewServer(DB)
	router,_:=svr.Start()
	if err!=nil{
		panic(err)
	}

	tests:=[]struct{
		Name string
		Role string
		Email string
		Password string
		wantStatus int
	}{
		{"When an admin is successfully logged in",dummyAdmin.Role,dummyAdmin.Email,dummyAdmin.Password,http.StatusOK},
		{"SuperAdmin with invalid credentials",dummySuperAdmin.Role,dummySuperAdmin.Email,"dummySuperPa",http.StatusUnauthorized},
		{"with empty fields","","","",http.StatusBadRequest},
		{"For invalid role","invalidRole","email","pass",http.StatusNotFound},
		}
	for _,test :=range tests{
		t.Run(test.Name,func(t *testing.T){
			request:=NewLogInRequest(test.Role,test.Email,test.Password)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
		})
	}
	token:=GetSuperToken(router)
	///tests for logout
	testLogout:=[]struct{
		name string
		token string
		wantStatus int
	}{
		{"when request is made with token",token,http.StatusOK},
		{"when token is not sent","",http.StatusBadRequest},
	}
	for _,test:=range testLogout{
		t.Run(test.name,func(t *testing.T){
			request:=NewLogOutRequest(test.token)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
		})
	}


}



///
func NewLogInRequest(Role string,Email string,Password string) *http.Request{
	user:=models.Credentials{
		Role:     Role,
		Email:    Email,
		Password: Password,
	}
	data,err:=json.Marshal(user)
	if err!=nil{
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPost, "/login", strings.NewReader(string(data)))
	req.Header.Add("Content-Type", "application/json")
	return req
}
func NewLogOutRequest(token string) *http.Request{
	req, _ := http.NewRequest(http.MethodGet, "/logout",nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}
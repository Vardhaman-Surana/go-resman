package controller_test

import (
	"encoding/json"
	"fmt"
	"github.com/vds/go-resman/pkg/database/mysql"
	"github.com/vds/go-resman/pkg/server"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAdminController(t *testing.T) {
	var DB, _ = mysql.NewMySqlDB("restaurant_test")
	defer DB.Close()
	defer CleanDB(DB)
	svr, err := server.NewServer(DB)
	router, _ := svr.Start()
	if err != nil {
		panic(err)
	}
	token := GetSuperToken(router)
	t.Run("get admins with a valid token", func(t *testing.T) {
		request := NewGetAdminRequest(token)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusOK)
	})

	// For update admins
	admin := struct {
		Name  string
		Email string
		ID    string
	}{"dummyAdmin", "dummyAdmin@gmail.com", "ef8f8dac-de63-4bc3-911e-38d34864e39a"}
	testUpdateAdmin := []struct {
		name       string
		adminID    string
		userName   string
		email      string
		wantStatus int
	}{
		{"update admin with empty field", "invalidAdminID", "", "", http.StatusBadRequest},
		{"update non existing admin", "invalidAdminID", "name", "email", http.StatusBadRequest},
		{"update existing admin", admin.ID, admin.Name, admin.Email, http.StatusOK},
	}
	for _, test := range testUpdateAdmin {
		t.Run(test.name, func(t *testing.T) {
			request := NewUpdateAdminRequest(token, test.adminID, test.userName, test.email)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			assertStatus(t, response.Code, test.wantStatus)
		})
	}
	////Tests for admin deletion

	///create admins for deletion
	CreateAdmins(DB)
	testDeleteAdmins := []struct {
		name       string
		idArr      []string
		wantStatus int
	}{
		{"tests for valid deletion", []string{"id1", "id2", "id3"}, http.StatusOK},
		{"empty array of id", nil, http.StatusBadRequest},
		{"try to delete invalid id", []string{"idInvalid", "id2Invalid"}, http.StatusBadRequest},
	}

	for _, test := range testDeleteAdmins {
		t.Run(test.name, func(t *testing.T) {
			idToDelete := test.idArr
			request := NewDeleteAdminRequest(token, idToDelete)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			assertStatus(t, response.Code, test.wantStatus)
		})
	}

}

///
func NewGetAdminRequest(token string) *http.Request {
	req, err := http.NewRequest(http.MethodGet, "/manage/admins", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req
}
func NewUpdateAdminRequest(token string, adminID string, userName string, email string) *http.Request {
	data := fmt.Sprintf(`{"email":"%s","name":"%s"}`, email, userName)
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/manage/admins/%s", adminID), strings.NewReader(data))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req
}

func GetSuperToken(router http.Handler) string {
	request := NewLogInRequest(dummySuperAdmin.Role, dummySuperAdmin.Email, dummySuperAdmin.Password)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	decoder := json.NewDecoder(response.Body)
	var data map[string]string
	decoder.Decode(&data)
	token := data["token"]
	return token
}

func GetAdminToken(router http.Handler) string {
	request := NewLogInRequest(dummyAdmin.Role, dummyAdmin.Email, dummyAdmin.Password)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	decoder := json.NewDecoder(response.Body)
	var data map[string]string
	decoder.Decode(&data)
	token := data["token"]
	return token
}
func GetOwnerToken(router http.Handler) string {
	request := NewLogInRequest("owner", dummyOwner.Email, dummyOwner.Password)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	decoder := json.NewDecoder(response.Body)
	var data map[string]string
	decoder.Decode(&data)
	token := data["token"]
	return token
}

func CreateAdmins(db *mysql.MySqlDB) {
	stmt, _ := db.Prepare("insert into admins(id,email_id,name,password) values(?,?,?,?)")
	stmt.Exec("id1", "email1", "name1", "pass1")
	stmt.Exec("id2", "email2", "name2", "pass2")
	stmt.Exec("id3", "email3", "name3", "pass3")
}
func NewDeleteAdminRequest(token string, idArr []string) *http.Request {
	var adminID struct {
		IDArr []string `json:"idArr"`
	}
	adminID.IDArr = idArr
	data, err := json.Marshal(adminID)
	if err != nil {
		panic(err)
	}
	req, _ := http.NewRequest(http.MethodDelete, "/manage/admins", strings.NewReader(string(data)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req
}

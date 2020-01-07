package controller_test

import (
	"encoding/json"
	"fmt"
	"github.com/vds/go-resman/pkg/database/mysql"
	"github.com/vds/go-resman/pkg/models"
	"github.com/vds/go-resman/pkg/server"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOwnerController(t *testing.T) {
	var DB, _ = mysql.NewMySqlDB("restaurant_test")
	defer DB.Close()
	defer CleanDB(DB)
	svr, err := server.NewServer(DB)
	router, _ := svr.Start()
	if err != nil {
		panic(err)
	}
	//test get admins
	token := GetSuperToken(router)
	tokenAdmin := GetAdminToken(router)

	testGetOwners := []struct {
		name  string
		token string
	}{
		{"get owners with a valid token", token},
		{"get owners with a token of admin who has not created any owner", tokenAdmin},
	}
	for _, test := range testGetOwners {
		t.Run(test.name, func(t *testing.T) {
			request := NewGetOwnerRequest(test.token)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			assertStatus(t, response.Code, http.StatusOK)
		})
	}

	///owner update
	owner := struct {
		Name  string
		Email string
	}{"dummySuperOwner", "dummySuperOwner@gmail.com"}
	testUpdateOwner := []struct {
		name       string
		token      string
		ownerID    string
		ownerEmail string
		ownerName  string
		wantStatus int
	}{
		{"update owner with empty field", token, "5d1606c1-0d82-48c4-9bea-6db088e4ad", "", "", http.StatusBadRequest},
		{"update existing owner", token, "32758fde-4dd5-4635-9bf7-a4d46d6f0629", owner.Email, owner.Name, http.StatusOK},
		{"update non existing owner", token, "32758fde-4dd5-4635-9bf7-a4d46d6f", owner.Email, owner.Name, http.StatusBadRequest},
		{"update existing owner with invalid creator", tokenAdmin, "32758fde-4dd5-4635-9bf7-a4d46d6f0629", owner.Email, owner.Name, http.StatusUnauthorized},
	}
	for _, test := range testUpdateOwner {
		t.Run(test.name, func(t *testing.T) {
			request := NewUpdateOwnerRequest(test.token, test.ownerID, test.ownerEmail, test.ownerName)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			assertStatus(t, response.Code, test.wantStatus)
		})
	}

	// creating owner
	testCreateOwner := []struct {
		name       string
		ownerPass  string
		wantStatus int
	}{
		{"Create owner with empty field", "", http.StatusBadRequest},
		{"Create owner with valid entries", "password", http.StatusOK},
		{"Create owner with duplicate email", "password", http.StatusBadRequest},
	}
	for _, test := range testCreateOwner {
		t.Run(test.name, func(t *testing.T) {
			request := NewCreateOwnerRequest(token, "email", "name", test.ownerPass)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			assertStatus(t, response.Code, test.wantStatus)
		})
	}
	// deleting owners
	CreateOwners(DB)
	testDeleteOwners := []struct {
		name       string
		idArr      []string
		wantStatus int
	}{
		{"tests for valid deletion", []string{"id1", "id2", "id3"}, http.StatusOK},
		{"empty array of id", nil, http.StatusBadRequest},
		{"try to delete invalid id", []string{"idInvalid", "id2Invalid"}, http.StatusBadRequest},
	}

	for _, test := range testDeleteOwners {
		t.Run(test.name, func(t *testing.T) {
			idToDelete := test.idArr
			request := NewDeleteOwnerRequest(token, idToDelete)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			assertStatus(t, response.Code, test.wantStatus)
		})
	}

}

///
func NewGetOwnerRequest(token string) *http.Request {
	req, err := http.NewRequest(http.MethodGet, "/manage/owners", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req
}
func NewUpdateOwnerRequest(token string, ownerID string, email string, userName string) *http.Request {
	data := fmt.Sprintf(`{"email":"%s","name":"%s"}`, email, userName)
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/manage/owners/%s", ownerID), strings.NewReader(data))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req
}
func NewCreateOwnerRequest(token string, email string, userName string, pass string) *http.Request {
	user := models.OwnerReg{
		Email:    email,
		Name:     userName,
		Password: pass,
	}
	data, err := json.Marshal(user)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPost, "/manage/owners", strings.NewReader(string(data)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req
}

func CreateOwners(db *mysql.MySqlDB) {
	stmt, _ := db.Prepare("insert into owners(id,email_id,name,password,creator_id) values(?,?,?,?,?)")
	stmt.Exec("id1", "email1", "name1", "pass1", "creator1")
	stmt.Exec("id2", "email2", "name2", "pass2", "creator2")
	stmt.Exec("id3", "email3", "name3", "pass3", "creator3")
}
func NewDeleteOwnerRequest(token string, idArr []string) *http.Request {
	var ownerID struct {
		IDArr []string `json:"idArr"`
	}
	ownerID.IDArr = idArr
	data, err := json.Marshal(ownerID)
	if err != nil {
		panic(err)
	}
	req, _ := http.NewRequest(http.MethodDelete, "/manage/owners", strings.NewReader(string(data)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req
}

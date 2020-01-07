package controller_test

import (
	"encoding/json"
	"github.com/vds/go-resman/pkg/database/mysql"
	"github.com/vds/go-resman/pkg/server"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	_ "testing"
)

func TestRegisterController(t *testing.T) {
	var DB, _ = mysql.NewMySqlDB("restaurant_test")
	defer DB.Close()
	defer CleanDB(DB)
	svr, err := server.NewServer(DB)
	router, _ := svr.Start()
	if err != nil {
		panic(err)
	}
	tests := []struct {
		name       string
		role       string
		email      string
		userName   string
		password   string
		wantStatus int
	}{
		{"When an admin is successfully created", "admin", "admin1@gmail.com", "admin1", "pass1", http.StatusOK},
		{"When a superadmin is successfully created", "superAdmin", "superadmin@gmail.com", "superadmin1", "superpass1", http.StatusOK},
		{"duplicate mail for admin", "admin", "admin1@gmail.com", "admin1", "pass1", http.StatusBadRequest},
		{"duplicate mail for super admin", "superAdmin", "superadmin@gmail.com", "superadmin1", "superpass1", http.StatusBadRequest},
		{"Empty Require Field", "", "admin1@gmail.com", "admin1", "pass1", http.StatusBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := NewRegisterRequest(test.role, test.userName, test.email, test.password)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			assertStatus(t, response.Code, test.wantStatus)
		})
	}
	t.Run("For invalid role", func(t *testing.T) {
		request := NewRegisterRequest("AnyOtherRole", "name", "mail", "pass")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusNotFound)
	})
}

///Helpers
func NewRegisterRequest(Role string, UserName string, Email string, Password string) *http.Request {
	user := map[string]string{
		"role":     Role,
		"email":    Email,
		"name":     UserName,
		"password": Password,
	}
	data, err := json.Marshal(user)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPost, "/register", strings.NewReader(string(data)))
	req.Header.Add("Content-Type", "application/json")
	return req
}
func CleanDB(db *mysql.MySqlDB) {
	_, _ = db.Query("delete from admins where email_id<>?", dummyAdmin.Email)
	_, _ = db.Query("delete from super_admins where email_id<>?", dummySuperAdmin.Email)
	_, _ = db.Query("delete from restaurants where id<>? and id<>?", 3, 4)
	_, _ = db.Query("alter table restaurants AUTO_INCREMENT=5")
	_, _ = db.Query("delete from dishes where id<>1")
	_, _ = db.Query("alter table dishes AUTO_INCREMENT=2")
	_, _ = db.Query("delete from owners where email_id<>? and id<>? ", dummyOwner.Email, dummyOwnerID)
	_, _ = db.Query("delete from invalid_tokens")
}
func assertStatus(t *testing.T, got int, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("got status %v want status %v", got, want)
	}
}

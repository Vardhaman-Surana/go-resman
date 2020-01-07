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

func TestMenuController(t *testing.T) {
	var DB, _ = mysql.NewMySqlDB("restaurant_test")
	defer DB.Close()
	defer CleanDB(DB)
	svr, err := server.NewServer(DB)
	router, _ := svr.Start()
	if err != nil {
		panic(err)
	}
	token := GetSuperToken(router)

	//tests for deleting dishes
	CreateDishes(DB)
	testDeleteDishes := []struct {
		name       string
		resID      int
		idArr      []int
		wantStatus int
	}{
		{"tests for valid deletion", 3, []int{2, 3, 4}, http.StatusOK},
		{"empty array of id", 3, nil, http.StatusBadRequest},
		{"try to delete invalid id", 3, []int{10, 11}, http.StatusBadRequest},
	}
	for _, test := range testDeleteDishes {
		t.Run(test.name, func(t *testing.T) {
			idToDelete := test.idArr
			request := NewDeleteDishRequest(token, test.resID, idToDelete)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			assertStatus(t, response.Code, test.wantStatus)
		})
	}

	//tokenAdmin:=GetAdminToken(router)
	///test to add dishes to a restaurant
	testAddDishes := []struct {
		name       string
		resID      int
		dishes     []models.Dish
		wantStatus int
	}{
		{"Add dishes successfully", 3, []models.Dish{{"dish1", 100.0}, {"dish2", 200.0}}, http.StatusOK},
		{"Adding dishes for a non existing restaurant", 10, []models.Dish{{"dish1", 100.0}, {"dish2", 200.0}}, http.StatusBadRequest},
	}
	for _, test := range testAddDishes {
		t.Run(test.name, func(t *testing.T) {
			request := NewAddDishesRequest(token, test.resID, test.dishes)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			assertStatus(t, response.Code, test.wantStatus)
		})
	}

	// get menu test
	testGetMenu := []struct {
		name       string
		token      string
		resID      int
		wantStatus int
	}{
		{"for an existing restaurant", token, 3, http.StatusOK},
		{"for a non existing restaurant", token, 10, http.StatusBadRequest},
		{"when no items in menu", token, 4, http.StatusOK},
	}
	for _, test := range testGetMenu {
		t.Run(test.name, func(t *testing.T) {
			request := NewGetMenuRequest(token, test.resID)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			assertStatus(t, response.Code, test.wantStatus)
		})
	}

	//tests for updating a dish
	testUpdateDish := []struct {
		name       string
		token      string
		resID      int
		dishID     int
		dishName   string
		dishPrice  float32
		wantStatus int
	}{
		{"with empty fields", token, 3, 1, "", 10.0, http.StatusBadRequest},
		{"update an existing dish", token, 3, 1, "dish1", 100.0, http.StatusOK},
		{"when dish id does not exist", token, 3, 10, "dish1", 100.0, http.StatusBadRequest},
	}
	for _, test := range testUpdateDish {
		t.Run(test.name, func(t *testing.T) {
			request := NewUpdateDishRequest(token, test.resID, test.dishID, test.dishName, test.dishPrice)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			assertStatus(t, response.Code, test.wantStatus)
		})
	}

}

//////
func NewAddDishesRequest(token string, resID int, dishes []models.Dish) *http.Request {
	data, err := json.Marshal(dishes)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/manage/restaurants/%d/menu", resID), strings.NewReader(string(data)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req
}
func NewGetMenuRequest(token string, resID int) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/manage/restaurants/%d/menu", resID), nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req
}
func NewUpdateDishRequest(token string, resID, dishID int, dishName string, dishPrice float32) *http.Request {
	dish := models.Dish{
		Name:  dishName,
		Price: dishPrice,
	}
	data, err := json.Marshal(dish)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/manage/restaurants/%d/menu/%d", resID, dishID), strings.NewReader(string(data)))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req
}

func CreateDishes(db *mysql.MySqlDB) {
	stmt, _ := db.Prepare("insert into dishes(name,price,res_id) values(?,?,?)")
	stmt.Exec("dish10", 100, 3)
	stmt.Exec("dish20", 200, 3)
	stmt.Exec("dish30", 300, 3)
}
func NewDeleteDishRequest(token string, resID int, idArr []int) *http.Request {
	var dishID struct {
		IDArr []int `json:"idArr"`
	}
	dishID.IDArr = idArr
	data, err := json.Marshal(dishID)
	if err != nil {
		panic(err)
	}
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/manage/restaurants/%d/menu", resID), strings.NewReader(string(data)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req
}

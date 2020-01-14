package testhelpers

import (
	"encoding/json"
	"fmt"
	"github.com/vds/go-resman/pkg/models"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func NewAddDishRequest(token string, resID int, dishe *models.DishOutput,baseUrl string)(*http.Request,error){
	data, err := json.Marshal(dishe)
	if err != nil {
		return nil,err
	}
	req, err := http.NewRequest(http.MethodPost, baseUrl+fmt.Sprintf("/manage/restaurants/%d/menu", resID), strings.NewReader(string(data)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req,nil
}

func AssertDish(t *testing.T, got *models.DishOutput, want *models.DishOutput) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v got %v", want, got)
	}
}


func NewUpdateDishRequest(token string, resID int, dish *models.DishOutput,baseUrl string)(*http.Request,error){
	data, err := json.Marshal(dish)
	if err != nil {
		return nil,err
	}
	req, err := http.NewRequest(http.MethodPut, baseUrl+fmt.Sprintf("/manage/restaurants/%d/menu/%d", resID, dish.ID),strings.NewReader(string(data)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req,nil
}


func NewDeleteDishRequest(token string, resID int, dishId int,baseUrl string)(*http.Request,error) {
	req, err := http.NewRequest(http.MethodDelete, baseUrl+fmt.Sprintf("/manage/restaurants/%d/menu?id=%d", resID,dishId),nil)
	if err!=nil{
		return nil,err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req,nil
}

func AssertMenu(t *testing.T, got []models.DishOutput, want []models.DishOutput) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v got %v", want, got)
	}
}

func NewGetMenuRequest(token string, resID int,baseUrl string)(*http.Request,error) {
	req, err := http.NewRequest(http.MethodGet,baseUrl+ fmt.Sprintf("/manage/restaurants/%d/menu", resID), nil)
	if err!=nil{
		return nil,err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req,nil
}

func GetOwnerRestaurantMenu()[]models.DishOutput{
	return []models.DishOutput{
		{ID: 2,Name: ownerRestaurantDish.Name, Price: ownerRestaurantDish.Price},
	}
}

func GetAdminRestaurantMenu()[]models.DishOutput{
	return []models.DishOutput{
		{ID: 1,Name: adminRestaurantDish.Name, Price: adminRestaurantDish.Price},
	}
}

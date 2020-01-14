package testhelpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vds/go-resman/pkg/models"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func NewCreateRestaurantRequest(token string, restaurant *models.RestaurantOutput, baseUrl string) (*http.Request, error) {
	data, err := json.Marshal(restaurant)
	if err != nil {
		return nil, err
	}
	body := bytes.NewBuffer(data)
	req, err := http.NewRequest(http.MethodPost, baseUrl+"/manage/restaurants", body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req, nil
}

func AssertRestaurant(t *testing.T, got *models.RestaurantOutput, want *models.RestaurantOutput) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v got %v", want, got)
	}
}

func NewUpdateRestaurantRequest(token string, restaurant *models.RestaurantOutput, baseUrl string) (*http.Request, error) {
	data, err := json.Marshal(restaurant)
	if err != nil {
		return nil, err
	}
	body := bytes.NewBuffer(data)
	req, err := http.NewRequest(http.MethodPut, baseUrl+fmt.Sprintf("/manage/restaurants/%d", restaurant.ID), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req, nil
}

func NewDeleteRestaurantRequest(token string, id int, baseUrl string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodDelete, baseUrl+fmt.Sprintf("/manage/restaurants?id=%d", id), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req, nil
}

func GetAllRestaurants() []models.RestaurantOutput {
	return []models.RestaurantOutput{
		{ID: 1, Name: restaurantByAdmin.Name, Lat: restaurantByAdmin.Lat, Lng: restaurantByAdmin.Lng},
		{ID: 2, Name: restaurantOfOwner.Name, Lat: restaurantOfOwner.Lat, Lng: restaurantOfOwner.Lng},
	}
}

func GetAdminRestaurants() []models.RestaurantOutput {
	return []models.RestaurantOutput{
		{ID: 1, Name: restaurantByAdmin.Name, Lat: restaurantByAdmin.Lat, Lng: restaurantByAdmin.Lng},
	}
}
func GetOwnerByAdminRestaurants() []models.RestaurantOutput {
	return []models.RestaurantOutput{
		{ID: 2, Name: restaurantOfOwner.Name, Lat: restaurantOfOwner.Lat, Lng: restaurantOfOwner.Lng},
	}
}

func NewGetRestaurantRequest(token string, baseUrl string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, baseUrl+"/manage/restaurants", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req, nil
}

func AssertRestaurants(t *testing.T, got []models.RestaurantOutput, want []models.RestaurantOutput) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v got %v", want, got)
	}
}

func GetOwnerCreatedByAdminId() string{
	return ownerByAdminId
}
func GetOwnerCreatedBySuperAdminId() string{
	return ownerBySuperAdminId
}

func NewGetOwnerRestaurantRequest(token string, ownerID string,baseUrl string) (*http.Request,error) {
	req, err := http.NewRequest(http.MethodGet, baseUrl+fmt.Sprintf("/manage/owners/%s/restaurants", ownerID), nil)
	if err != nil {
		return nil,err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req,nil
}

func GetAvailableRestaurant() []models.RestaurantOutput{
	return []models.RestaurantOutput{
		models.RestaurantOutput{ID: 1,Name: restaurantByAdmin.Name,Lat:restaurantByAdmin.Lat,Lng: restaurantByAdmin.Lng},
	}
}

func NewGetRestaurantAvailableRequest(token string,baseUrl string)(*http.Request,error) {
	req, err := http.NewRequest(http.MethodGet, baseUrl+"/manage/available/restaurants", nil)
	if err != nil {
		return nil,err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req,nil
}


func NewGetNearByRestaurants(lat float32, lng float32,baseUrl string) (*http.Request,error) {
	location := map[string]float32{
		"lat": lat,
		"lng": lng,
	}
	data, err := json.Marshal(location)
	if err != nil {
		return nil,err
	}
	req, err := http.NewRequest(http.MethodGet, baseUrl+"/restaurantsNearBy", strings.NewReader(string(data)))
	if err != nil {
		return nil,err
	}
	req.Header.Add("Content-Type", "application/json")
	return req,nil
}


func NewAddOwnerRestaurantRequest(token string, ownerID string, assignIds []int,
	deAssignIds []int,baseUrl string) (*http.Request,error) {
	var resID struct {
		Assign   []int `json:"assign"`
		DeAssign []int `json:"deAssign"`
	}
	resID.Assign = assignIds
	resID.DeAssign = deAssignIds
	data, err := json.Marshal(resID)
	if err != nil {
		return nil,err
	}
	req, err := http.NewRequest(http.MethodPost, baseUrl+fmt.Sprintf("/manage/owners/%s/restaurants", ownerID), strings.NewReader(string(data)))
	if err != nil {
		return nil,err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	return req,nil
}

package main

import (
	"encoding/json"
	"github.com/vds/go-resman/pkg/models"
	"github.com/vds/go-resman/pkg/testhelpers"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestRestaurants(t *testing.T) {
	superAdminToken, err := testhelpers.GetSuperAdminToken(serverUrl)
	if err != nil {
		t.Fatalf("unable to get superAdminToken: %v", err)
	}
	adminToken, err := testhelpers.GetAdminToken(serverUrl)
	if err != nil {
		t.Fatalf("unable to get adminToken: %v", err)
	}
	ownerByAdminToken, err := testhelpers.GetOwnerByAdminToken(serverUrl)
	if err != nil {
		t.Fatalf("unable to get adminToken: %v", err)
	}
	ownerBySuperAdminToken, err := testhelpers.GetOwnerBySuperAdminToken(serverUrl)
	if err != nil {
		t.Fatalf("unable to get adminToken: %v", err)
	}

	restaurantByAdmin := models.RestaurantOutput{
		Name: "resByAdmin",
		Lat:  11.0,
		Lng:  15.0,
	}
	restaurantBySuperAdmin := models.RestaurantOutput{
		Name: "resByAdmin",
		Lat:  11.0,
		Lng:  15.0,
	}
	testCreateRestaurant := []struct {
		name         string
		token        string
		restaurant   *models.RestaurantOutput
		wantedStatus int
	}{
		{
			name:         "create restaurant by admin",
			token:        adminToken,
			restaurant:   &restaurantByAdmin,
			wantedStatus: http.StatusOK,
		},
		{
			name:         "create restaurant by superAdmin",
			token:        superAdminToken,
			restaurant:   &restaurantBySuperAdmin,
			wantedStatus: http.StatusOK,
		},
		{
			name:         "create restaurant with empty fields",
			token:        superAdminToken,
			restaurant:   &models.RestaurantOutput{Name: "", Lat: 1, Lng: 1},
			wantedStatus: http.StatusBadRequest,
		},
	}
	for _, test := range testCreateRestaurant {
		t.Run(test.name, func(t *testing.T) {
			request, err := testhelpers.NewCreateRestaurantRequest(test.token, test.restaurant, serverUrl)
			if err != nil {
				t.Fatalf("unable to create request:%v", err)
			}
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatalf("http request failed:%v", err)
			}
			testhelpers.AssertStatus(t, resp.StatusCode, test.wantedStatus)
			if test.wantedStatus == 200 {
				var createdRestaurant models.RestaurantOutput
				body, _ := ioutil.ReadAll(resp.Body)
				err := json.Unmarshal(body, &createdRestaurant)
				if err != nil {
					t.Fatalf("response not in correct format:%v", err)
				}
				test.restaurant.ID = createdRestaurant.ID
				testhelpers.AssertRestaurant(t, &createdRestaurant, test.restaurant)
			}
		})
	}

	restaurantByAdmin.Name = "restaurantUpdated"
	restaurantBySuperAdmin.Name = "superRestaurantUpdated"

	testUpdateRestaurant := []struct {
		name         string
		token        string
		restaurant   *models.RestaurantOutput
		wantedStatus int
	}{

		{
			name:         "update restaurant by admin created by him",
			token:        adminToken,
			restaurant:   &restaurantByAdmin,
			wantedStatus: http.StatusOK,
		},
		{
			name:         "update restaurant by admin not created by him",
			token:        adminToken,
			restaurant:   &restaurantBySuperAdmin,
			wantedStatus: http.StatusUnauthorized,
		},
		{
			name:         "update restaurant by superAdmin",
			token:        superAdminToken,
			restaurant:   &restaurantBySuperAdmin,
			wantedStatus: http.StatusOK,
		},

		{
			name:         "update restaurant with empty fields",
			token:        superAdminToken,
			restaurant:   &models.RestaurantOutput{Name: "", Lat: 1, Lng: 1},
			wantedStatus: http.StatusBadRequest,
		},
	}
	for _, test := range testUpdateRestaurant {
		t.Run(test.name, func(t *testing.T) {
			request, err := testhelpers.NewUpdateRestaurantRequest(test.token, test.restaurant, serverUrl)
			if err != nil {
				t.Fatalf("unable to create request:%v", err)
			}
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatalf("http request failed:%v", err)
			}
			testhelpers.AssertStatus(t, resp.StatusCode, test.wantedStatus)
			if test.wantedStatus == 200 {
				var createdRestaurant models.RestaurantOutput
				body, _ := ioutil.ReadAll(resp.Body)
				err := json.Unmarshal(body, &createdRestaurant)
				if err != nil {
					t.Fatalf("response not in correct format:%v", err)
				}
				test.restaurant.ID = createdRestaurant.ID
				testhelpers.AssertRestaurant(t, &createdRestaurant, test.restaurant)
			}
		})
	}

	testDeleteRestaurants := []struct {
		name         string
		token        string
		id           int
		wantedStatus int
	}{
		{
			name:         "admin deleting a restaurant not created by him",
			token:        adminToken,
			id:           restaurantBySuperAdmin.ID,
			wantedStatus: http.StatusBadRequest,
		},
		{
			name:         "admin deleting a restaurant created by him",
			token:        adminToken,
			id:           restaurantByAdmin.ID,
			wantedStatus: http.StatusOK,
		},
		{
			name:         "superAdmin deleting a restaurant created by him",
			token:        superAdminToken,
			id:           restaurantBySuperAdmin.ID,
			wantedStatus: http.StatusOK,
		},
		{
			name:         "superAdmin deleting a non existing restaurant",
			token:        superAdminToken,
			id:           1000,
			wantedStatus: http.StatusBadRequest,
		},
	}
	for _, test := range testDeleteRestaurants {
		t.Run(test.name, func(t *testing.T) {
			request, err := testhelpers.NewDeleteRestaurantRequest(test.token, test.id, serverUrl)
			if err != nil {
				t.Fatalf("unable to create request:%v", err)
			}
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatalf("http request failed:%v", err)
			}
			testhelpers.AssertStatus(t, resp.StatusCode, test.wantedStatus)
		})
	}

	testGetRestaurants := []struct {
		name              string
		token             string
		wantedRestaurants []models.RestaurantOutput
	}{
		{name: "get restaurants by admin", token: adminToken, wantedRestaurants: testhelpers.GetAdminRestaurants()},
		{name: "get restaurants by superAdmin", token: superAdminToken, wantedRestaurants: testhelpers.GetAllRestaurants()},
		{name: "get restaurants by owner", token: ownerByAdminToken, wantedRestaurants: testhelpers.GetOwnerByAdminRestaurants()},
		{name: "get restaurants by owner (empty restaurant list)", token: ownerBySuperAdminToken, wantedRestaurants: []models.RestaurantOutput{}},
	}
	for _, test := range testGetRestaurants {
		t.Run(test.name, func(t *testing.T) {
			request, err := testhelpers.NewGetRestaurantRequest(test.token, serverUrl)
			if err != nil {
				t.Fatalf("unable to create request:%v", err)
			}
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatalf("http request failed:%v", err)
			}
			testhelpers.AssertStatus(t, resp.StatusCode, http.StatusOK)
			var gotRestaurants []models.RestaurantOutput
			body, _ := ioutil.ReadAll(resp.Body)
			err = json.Unmarshal(body, &gotRestaurants)
			if err != nil {
				t.Fatalf("response not in correct format:%v", err)
			}
			testhelpers.AssertRestaurants(t, gotRestaurants, test.wantedRestaurants)
		})
	}
	// get restaurants for an owner

	ownerBySuperAdminId := testhelpers.GetOwnerCreatedBySuperAdminId()
	ownerByAdminId := testhelpers.GetOwnerCreatedByAdminId()

	testOwnerRestaurants := []struct {
		name              string
		token             string
		ownerID           string
		wantedStatus      int
		wantedRestaurants []models.RestaurantOutput
	}{
		{
			name:              "getting restaurants of an owner by superadmin",
			token:             superAdminToken,
			ownerID:           ownerBySuperAdminId,
			wantedStatus:      http.StatusOK,
			wantedRestaurants: []models.RestaurantOutput{},
		},
		{
			name:         "getting restaurants of an owner by admin not created by him",
			token:        adminToken,
			ownerID:      ownerBySuperAdminId,
			wantedStatus: http.StatusUnauthorized,
		},
		{
			name:              "getting restaurants of owner by admin created by him",
			token:             adminToken,
			ownerID:           ownerByAdminId,
			wantedStatus:      http.StatusOK,
			wantedRestaurants: testhelpers.GetOwnerByAdminRestaurants(),
		},
	}
	for _, test := range testOwnerRestaurants {
		t.Run(test.name, func(t *testing.T) {
			request, err := testhelpers.NewGetOwnerRestaurantRequest(test.token, test.ownerID, serverUrl)
			if err != nil {
				t.Fatalf("unable to create request:%v", err)
			}
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatalf("http request failed:%v", err)
			}
			testhelpers.AssertStatus(t, resp.StatusCode, test.wantedStatus)
			if test.wantedStatus == http.StatusOK {
				var gotRestaurants []models.RestaurantOutput
				body, _ := ioutil.ReadAll(resp.Body)
				err = json.Unmarshal(body, &gotRestaurants)
				if err != nil {
					t.Fatalf("response not in correct format:%v", err)
				}
				testhelpers.AssertRestaurants(t, gotRestaurants, test.wantedRestaurants)
			}
		})
	}

	// get available restaurants
	testGetAvailableRestaurants := []struct {
		name              string
		token             string
		wantedRestaurants []models.RestaurantOutput
	}{
		{
			name:              "get restaurants by superadmin(availabe here)",
			token:             superAdminToken,
			wantedRestaurants: testhelpers.GetAvailableRestaurant(),
		},
		{
			name:              "get restaurants by admin",
			token:             adminToken,
			wantedRestaurants: testhelpers.GetAvailableRestaurant(),
		},
	}
	for _, test := range testGetAvailableRestaurants {
		t.Run(test.name, func(t *testing.T) {
			request, err := testhelpers.NewGetRestaurantAvailableRequest(test.token, serverUrl)
			if err != nil {
				t.Fatalf("unable to create request:%v", err)
			}
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatalf("http request failed:%v", err)
			}
			testhelpers.AssertStatus(t, resp.StatusCode, http.StatusOK)
			var gotRestaurants []models.RestaurantOutput
			body, _ := ioutil.ReadAll(resp.Body)
			err = json.Unmarshal(body, &gotRestaurants)
			if err != nil {
				t.Fatalf("response not in correct format:%v", err)
			}
			testhelpers.AssertRestaurants(t, gotRestaurants, test.wantedRestaurants)

		})
	}

	///get restaurants near by
	testGetNearByRestaurants := []struct {
		name              string
		lat               float32
		lng               float32
		wantedRestaurants []models.RestaurantOutput
	}{
		{name: "get restaurants nearby(available)", lat: 10.08, lng: 15.01, wantedRestaurants: testhelpers.GetAvailableRestaurant()},
		{name: "get restaurants nearby(not available)", lat: 2.08, lng: 2.01, wantedRestaurants: []models.RestaurantOutput{}},
	}
	for _, test := range testGetNearByRestaurants {
		t.Run(test.name, func(t *testing.T) {
			request, err := testhelpers.NewGetNearByRestaurants(test.lat, test.lng, serverUrl)
			if err != nil {
				t.Fatalf("unable to create request:%v", err)
			}
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatalf("http request failed:%v", err)
			}
			testhelpers.AssertStatus(t, resp.StatusCode, http.StatusOK)
			var gotRestaurants []models.RestaurantOutput
			body, _ := ioutil.ReadAll(resp.Body)
			err = json.Unmarshal(body, &gotRestaurants)
			if err != nil {
				t.Fatalf("response not in correct format:%v", err)
			}
			testhelpers.AssertRestaurants(t, gotRestaurants, test.wantedRestaurants)

		})
	}

	//add owner to restaurants
	testAddOwnerRestaurants := []struct {
		name          string
		token         string
		ownerID       string
		toAssignIds   []int
		toDeAssignIds []int
		wantedStatus  int
	}{
		{
			name:          "admin trying to add owner(not created by him) to a restaurant",
			token:         adminToken,
			ownerID:       ownerBySuperAdminId,
			toAssignIds:   []int{},
			toDeAssignIds: []int{},
			wantedStatus:  http.StatusUnauthorized,
		},
		{
			name:          "admin trying to add owner to a restaurant(not created by him)",
			token:         adminToken,
			ownerID:       ownerByAdminId,
			toAssignIds:   []int{2},
			toDeAssignIds: []int{},
			wantedStatus:  http.StatusBadRequest,
		},
		{
			name:          "admin trying to remove owner of a restaurant(not created by him)",
			token:         adminToken,
			ownerID:       ownerByAdminId,
			toAssignIds:   []int{},
			toDeAssignIds: []int{2},
			wantedStatus:  http.StatusBadRequest,
		},
		{
			name:          "admin adding owner to a restaurant",
			token:         adminToken,
			ownerID:       ownerByAdminId,
			toAssignIds:   []int{1},
			toDeAssignIds: []int{},
			wantedStatus:  http.StatusOK,
		},
		{
			name:          "superAdmin removing owner of a restaurant(making it available)",
			token:         superAdminToken,
			ownerID:       ownerByAdminId,
			toAssignIds:   []int{},
			toDeAssignIds: []int{1},
			wantedStatus:  http.StatusOK,
		},
	}
	for _, test := range testAddOwnerRestaurants {
		t.Run(test.name, func(t *testing.T) {
			request,err := testhelpers.NewAddOwnerRestaurantRequest(test.token,
									test.ownerID, test.toAssignIds,test.toDeAssignIds,serverUrl)
			if err != nil {
				t.Fatalf("unable to create request:%v", err)
			}
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatalf("http request failed:%v", err)
			}
			testhelpers.AssertStatus(t, resp.StatusCode, test.wantedStatus)
		})
	}
}

package main

import (
	"encoding/json"
	"github.com/vds/go-resman/pkg/models"
	"github.com/vds/go-resman/pkg/testhelpers"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestMenu(t *testing.T) {
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

	superAdminDish := models.DishOutput{
		ID:    0,
		Name:  "dish10",
		Price: 50,
	}
	adminDish := models.DishOutput{
		ID:    0,
		Name:  "dish100",
		Price: 100,
	}
	ownerDish := models.DishOutput{
		ID:    0,
		Name:  "dish1000",
		Price: 100,
	}
	testAddDishes := []struct {
		name         string
		resID        int
		token        string
		dish         *models.DishOutput
		wantedStatus int
	}{
		{
			name:         "Add dish successfully by superAdmin",
			resID:        2,
			token:        superAdminToken,
			dish:         &superAdminDish,
			wantedStatus: http.StatusOK,
		},
		{
			name:         "Add dish successfully by owner",
			resID:        2,
			token:        ownerByAdminToken,
			dish:         &ownerDish,
			wantedStatus: http.StatusOK,
		},
		{
			name:         "Add dish successfully by admin",
			resID:        1,
			token:        adminToken,
			dish:         &adminDish,
			wantedStatus: http.StatusOK,
		},
		{
			name:         "Adding dishes for a non existing restaurant",
			resID:        10,
			token:        superAdminToken,
			dish:         &adminDish,
			wantedStatus: http.StatusBadRequest,
		},
		{
			name:         "Adding dishes for a restaurant(not created by him) by admin",
			resID:        2,
			token:        adminToken,
			dish:         &adminDish,
			wantedStatus: http.StatusUnauthorized,
		},
		{
			name:         "Adding dishes for a restaurant(not owned by him) by owner",
			resID:        1,
			token:        ownerByAdminToken,
			dish:         &adminDish,
			wantedStatus: http.StatusUnauthorized,
		},
		{
			name:         "Adding dishes with empty fields",
			resID:        1,
			token:        adminToken,
			dish:         &models.DishOutput{Name: "", Price: 0},
			wantedStatus: http.StatusBadRequest,
		},


	}
	for _, test := range testAddDishes {
		t.Run(test.name, func(t *testing.T) {
			request, err := testhelpers.NewAddDishRequest(test.token, test.resID, test.dish, serverUrl)
			if err != nil {
				t.Fatalf("unable to create request:%v", err)
			}
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatalf("http request failed:%v", err)
			}
			testhelpers.AssertStatus(t, resp.StatusCode, test.wantedStatus)
			if test.wantedStatus == 200 {
				var createdDish models.DishOutput
				body, _ := ioutil.ReadAll(resp.Body)
				err := json.Unmarshal(body, &createdDish)
				if err != nil {
					t.Fatalf("response not in correct format:%v", err)
				}
				test.dish.ID = createdDish.ID

				testhelpers.AssertDish(t, &createdDish, test.dish)
			}
		})
	}

	// updating the dishes
	adminDish.Name = "updatedNameAdmin"
	superAdminDish.Name = "updatedNameSuper"
	ownerDish.Price = 10.0
	testUpdateDish := []struct {
		name         string
		token        string
		resID        int
		dish         *models.DishOutput
		wantedStatus int
	}{
		{
			name:         "with empty fields",
			token:        superAdminToken,
			resID:        1,
			dish:         &models.DishOutput{ID: 4, Name: "", Price: 0},
			wantedStatus: http.StatusBadRequest,
		},
		{
			name:         "dish updated successfully by owner",
			token:        ownerByAdminToken,
			resID:        2,
			dish:         &ownerDish,
			wantedStatus: http.StatusOK,
		},
		{
			name:         "dish updated successfully by admin",
			token:        adminToken,
			resID:        1,
			dish:         &adminDish,
			wantedStatus: http.StatusOK,
		},
		{
			name:         "dish updated successfully by superAdmin",
			token:        superAdminToken,
			resID:        2,
			dish:         &superAdminDish,
			wantedStatus: http.StatusOK,
		},
		{
			name:         "updating dish of some other restaurant",
			token:        superAdminToken,
			resID:        1,
			dish:         &superAdminDish,
			wantedStatus: http.StatusBadRequest,
		},
		{
			name:         "updating dish of restaurant (not owned by him) by owner",
			token:        ownerByAdminToken,
			resID:        1,
			dish:         &ownerDish,
			wantedStatus: http.StatusUnauthorized,
		},
		{
			name:         "updating dish of restaurant (not created by him) by admin",
			token:        adminToken,
			resID:        2,
			dish:         &ownerDish,
			wantedStatus: http.StatusUnauthorized,
		},
	}
	for _, test := range testUpdateDish {
		t.Run(test.name, func(t *testing.T) {
			request,err := testhelpers.NewUpdateDishRequest(test.token, test.resID, test.dish, serverUrl)
			if err != nil {
				t.Fatalf("unable to create request:%v", err)
			}
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatalf("http request failed:%v", err)
			}
			testhelpers.AssertStatus(t, resp.StatusCode, test.wantedStatus)
			if test.wantedStatus == 200 {
				var updatedDish models.DishOutput
				body, _ := ioutil.ReadAll(resp.Body)
				err := json.Unmarshal(body, &updatedDish)
				if err != nil {
					t.Fatalf("response not in correct format:%v", err)
				}
				testhelpers.AssertDish(t, &updatedDish, test.dish)
			}
		})
	}


	//deleting the dish
	testDeleteDishes := []struct {
		name       string
		resID      int
		token string
		dishID      	int
		wantedStatus int
	}{
		{
			name:"admin trying to delete a restaurant(not created by him) dish",
			resID: 2,
			token: adminToken,
			dishID: ownerDish.ID,
			wantedStatus: http.StatusUnauthorized,
		},
		{
			name:"owner trying to delete a restaurant(not owned by him) dish",
			resID: 1,
			token: ownerByAdminToken,
			dishID: adminDish.ID,
			wantedStatus: http.StatusUnauthorized,
		},
		{
			name:"tests for valid deletion by superAdmin",
			resID: 2,
			token: superAdminToken,
			dishID: superAdminDish.ID,
			wantedStatus: http.StatusOK,
		},
		{
			name:"tests for valid deletion by admin",
			resID: 1,
			token: adminToken,
			dishID: adminDish.ID,
			wantedStatus: http.StatusOK,
		},
		{
			name:"tests for valid deletion by owner",
			resID: 2,
			token: ownerByAdminToken,
			dishID: ownerDish.ID,
			wantedStatus: http.StatusOK,
		},
		{
			name:"trying to delete a non existing dish",
			resID: 2,
			token: ownerByAdminToken,
			dishID: 12,
			wantedStatus: http.StatusBadRequest,
		},
	}
	for _, test := range testDeleteDishes {
		t.Run(test.name, func(t *testing.T) {
			request,err := testhelpers.NewDeleteDishRequest(test.token, test.resID,test.dishID,serverUrl)
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


	testGetMenu := []struct {
		name       string
		token      string
		resID      int
		wantedStatus int
		wantedMenu []models.DishOutput
	}{
		{
			name: "for an existing restaurant by admin",
			token: adminToken,
			resID: 1,
			wantedStatus: http.StatusOK,
			wantedMenu: testhelpers.GetAdminRestaurantMenu(),
		},
		{
			name: "for a non existing restaurant by admin",
			token: superAdminToken,
			resID: 10,
			wantedStatus: http.StatusBadRequest,
		},
		{
			name: "for an existing restaurant(not created by him) by admin",
			token: adminToken,
			resID: 2,
			wantedStatus: http.StatusUnauthorized,
		},
		{
			name: "for an existing restaurant(not owned by him) by owner",
			token: ownerByAdminToken,
			resID: 1,
			wantedStatus: http.StatusUnauthorized,
		},
		{
			name: "for an existing restaurant by owner",
			token: ownerByAdminToken,
			resID: 2,
			wantedStatus: http.StatusOK,
			wantedMenu: testhelpers.GetOwnerRestaurantMenu(),
		},
	}
	for _, test := range testGetMenu {
		t.Run(test.name, func(t *testing.T) {
			request,err := testhelpers.NewGetMenuRequest(test.token, test.resID,serverUrl)
			if err != nil {
				t.Fatalf("unable to create request:%v", err)
			}
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatalf("http request failed:%v", err)
			}
			testhelpers.AssertStatus(t, resp.StatusCode, test.wantedStatus)
			if test.wantedStatus == 200 {
				var gotMenu []models.DishOutput
				body, _ := ioutil.ReadAll(resp.Body)
				err := json.Unmarshal(body, &gotMenu)
				if err != nil {
					t.Fatalf("response not in correct format:%v", err)
				}
				testhelpers.AssertMenu(t, gotMenu, test.wantedMenu)
			}
		})
	}



}

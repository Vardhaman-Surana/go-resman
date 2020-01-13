package main

import (
	"encoding/json"
	"github.com/vds/go-resman/pkg/models"
	"github.com/vds/go-resman/pkg/testhelpers"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestOwners(t *testing.T) {
	superAdminToken, err := testhelpers.GetSuperAdminToken(serverUrl)
	if err != nil {
		t.Fatalf("unable to get superAdminToken: %v", err)
	}
	adminToken, err := testhelpers.GetAdminToken(serverUrl)
	if err != nil {
		t.Fatalf("unable to get adminToken: %v", err)
	}
	ownerByAdmin := models.OwnerReg{
		Email:    "owner100@gmail.com",
		Name:     "owner100",
		Password: "ownerPass100",
	}
	var ownerByAdminId string

	ownerBySuperAdmin := models.OwnerReg{
		Email:    "owner1000@gmail.com",
		Name:     "owner1000",
		Password: "ownerPass1000",
	}
	var ownerBySuperAdminId string

	testCreateOwner := []struct {
		name         string
		token        string
		owner        *models.OwnerReg
		wantedStatus int
	}{
		{name: "Create owner with empty field", token: superAdminToken, owner: &models.OwnerReg{Email: "", Name: "name", Password: "pa"}, wantedStatus: http.StatusBadRequest},
		{name: "Create owner with valid entries by admin", token: adminToken, owner: &ownerByAdmin, wantedStatus: http.StatusOK},
		{name: "Create owner with valid entries by superAdmin", token: superAdminToken, owner: &ownerBySuperAdmin, wantedStatus: http.StatusOK},
		{name: "Create owner with duplicate email", token: superAdminToken, owner: &ownerByAdmin, wantedStatus: http.StatusBadRequest},
	}
	for _, test := range testCreateOwner {
		t.Run(test.name, func(t *testing.T) {
			request, err := testhelpers.NewCreateOwnerRequest(test.token, test.owner, serverUrl)
			if err != nil {
				t.Fatalf("unable to create request:%v", err)
			}
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatalf("http request failed:%v", err)
			}
			testhelpers.AssertStatus(t, resp.StatusCode, test.wantedStatus)
			if test.wantedStatus == 200 {
				var createdOwner models.UserOutput
				body, _ := ioutil.ReadAll(resp.Body)
				err := json.Unmarshal(body, &createdOwner)
				if err != nil {
					t.Fatalf("response not in correct format:%v", err)
				}
				wantOwner := models.UserOutput{ID: createdOwner.ID, Email: test.owner.Email, Name: test.owner.Name}
				testhelpers.AssertOwner(t, &createdOwner, &wantOwner)
				if test.token == superAdminToken {
					ownerBySuperAdminId = createdOwner.ID
				} else {
					ownerByAdminId = createdOwner.ID
				}
			}
		})
	}
	// updating the data
	ownerByAdmin.Email = "updatedEmail"
	ownerBySuperAdmin.Email = "superUpdatedEmail"

	testUpdateOwner := []struct {
		name         string
		token        string
		owner        *models.UserOutput
		wantedStatus int
	}{
		{
			name:         "update owner with empty field",
			token:        superAdminToken,
			owner:        &models.UserOutput{ID: ownerByAdminId, Email: "", Name: ownerByAdmin.Name},
			wantedStatus: http.StatusBadRequest,
		},
		{
			name:         "update existing owner by admin",
			token:        adminToken,
			owner:        &models.UserOutput{ID: ownerByAdminId, Email: ownerByAdmin.Email, Name: ownerByAdmin.Name},
			wantedStatus: http.StatusOK,
		},
		{
			name:         "update existing owner by superAdmin",
			token:        superAdminToken,
			owner:        &models.UserOutput{ID: ownerBySuperAdminId, Email: ownerBySuperAdmin.Email, Name: ownerBySuperAdmin.Name},
			wantedStatus: http.StatusOK,
		},
		{
			name:         "update non existing owner",
			token:        superAdminToken,
			owner:        &models.UserOutput{ID: "invalidId", Email: ownerBySuperAdmin.Email, Name: ownerBySuperAdmin.Name},
			wantedStatus: http.StatusBadRequest,
		},
		{
			name:         "update existing owner with invalid creator",
			token:        adminToken,
			owner:        &models.UserOutput{ID: ownerBySuperAdminId, Email: ownerBySuperAdmin.Email, Name: ownerBySuperAdmin.Name},
			wantedStatus: http.StatusUnauthorized},
	}
	for _, test := range testUpdateOwner {
		t.Run(test.name, func(t *testing.T) {
			request, err := testhelpers.NewUpdateOwnerRequest(test.token, test.owner, serverUrl)
			if err != nil {
				t.Fatalf("unable to create request:%v", err)
			}
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatalf("http request failed:%v", err)
			}
			testhelpers.AssertStatus(t, resp.StatusCode, test.wantedStatus)
			if test.wantedStatus == 200 {
				var updatedOwner models.UserOutput
				body, _ := ioutil.ReadAll(resp.Body)
				err := json.Unmarshal(body, &updatedOwner)
				if err != nil {
					t.Fatalf("response not in correct format:%v", err)
				}
				testhelpers.AssertOwner(t, &updatedOwner, test.owner)
			}
		})
	}

	testDeleteOwner := []struct {
		name         string
		token        string
		id           string
		wantedStatus int
	}{
		{
			name:         "tests for invalid deletion by admin",
			token:        adminToken,
			id:           ownerBySuperAdminId,
			wantedStatus: http.StatusBadRequest,
		},
		{
			name:         "tests for valid deletion by admin",
			token:        adminToken,
			id:           ownerByAdminId,
			wantedStatus: http.StatusOK,
		},
		{
			name:         "tests for valid deletion by superAdmin",
			token:        superAdminToken,
			id:           ownerBySuperAdminId,
			wantedStatus: http.StatusOK,
		},
		{
			name:         "tests for non existing owner deletion",
			token:        superAdminToken,
			id:           "id123",
			wantedStatus: http.StatusBadRequest,
		},
	}
	for _, test := range testDeleteOwner {
		t.Run(test.name, func(t *testing.T) {
			request, err := testhelpers.NewDeleteOwnerRequest(test.token, test.id, serverUrl)
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

	allOwners := testhelpers.GetAllOwners()
	ownersByAdmin := testhelpers.GetOwnersByAdmin()

	testGetOwners := []struct {
		name         string
		token        string
		wantedOwners []models.UserOutput
		wantedStatus int
	}{
		{name: "get owners by superAdmin", token: superAdminToken, wantedOwners: allOwners, wantedStatus: http.StatusOK},
		{name: "get owners by admin", token: adminToken, wantedOwners: ownersByAdmin, wantedStatus: http.StatusOK},
	}
	for _, test := range testGetOwners {
		t.Run(test.name, func(t *testing.T) {
			request, err := testhelpers.NewGetOwnerRequest(test.token, serverUrl)
			if err != nil {
				t.Fatalf("unable to create request:%v", err)
			}
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatalf("http request failed:%v", err)
			}
			testhelpers.AssertStatus(t, resp.StatusCode, test.wantedStatus)
			var gotOwners []models.UserOutput
			body, _ := ioutil.ReadAll(resp.Body)
			err = json.Unmarshal(body, &gotOwners)
			if err != nil {
				t.Fatalf("response not in correct format:%v", err)
			}
			testhelpers.AssertOwners(t, gotOwners, test.wantedOwners)
		})
	}
}

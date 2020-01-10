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
	ownerToCreateByAdmin := models.OwnerReg{
		Email:    "owner100@gmail.com",
		Name:     "owner100",
		Password: "ownerPass100",
	}
	var ownerToCreateByAdminId string

	ownerToCreateBySuperAdmin := models.OwnerReg{
		Email:    "owner1000@gmail.com",
		Name:     "owner1000",
		Password: "ownerPass1000",
	}
	var ownerToCreateBySuperAdminId string

	testCreateOwner := []struct {
		name         string
		token        string
		owner        *models.OwnerReg
		wantedStatus int
	}{
		{name: "Create owner with empty field", token: superAdminToken, owner: &models.OwnerReg{Email: "", Name: "name", Password: "pa"}, wantedStatus: http.StatusBadRequest},
		{name: "Create owner with valid entries by admin", token: adminToken, owner: &ownerToCreateByAdmin, wantedStatus: http.StatusOK},
		{name: "Create owner with valid entries by superAdmin", token: superAdminToken, owner: &ownerToCreateBySuperAdmin, wantedStatus: http.StatusOK},
		{name: "Create owner with duplicate email", token: superAdminToken, owner: &ownerToCreateByAdmin, wantedStatus: http.StatusBadRequest},
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
			if test.wantedStatus==200{
				var createdOwner models.UserOutput
				body,_ := ioutil.ReadAll(resp.Body)
				err:=json.Unmarshal(body,&createdOwner)
				if err!=nil {
					t.Fatalf("response not in correct format:%v", err)
				}
				wantOwner:=models.UserOutput{ID: createdOwner.ID,Email:test.owner.Email,Name: test.owner.Name}
				testhelpers.AssertOwner(t,&createdOwner,&wantOwner)
				if test.token==superAdminToken{
					ownerToCreateBySuperAdminId = createdOwner.ID
				}else{
					ownerToCreateByAdminId = createdOwner.ID
				}
			}
		})
	}





}

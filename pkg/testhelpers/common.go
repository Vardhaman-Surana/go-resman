package testhelpers

import (
	"encoding/json"
	"net/http"
	"testing"
)

func AssertStatus(t *testing.T, got int, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("got status %v want status %v", got, want)
	}
}

func GetSuperAdminToken(baseUrl string) (string,error) {
	request,err := NewLogInRequest(SuperAdminCredentials(),baseUrl)
	if err!=nil{
		return "",err
	}
	resp, err := http.DefaultClient.Do(request)
	if err!=nil{
		return "",err
	}
	decoder := json.NewDecoder(resp.Body)
	var data map[string]string
	err=decoder.Decode(&data)
	if err!=nil{
		return "",err
	}
	token := data["token"]
	return token,err
}


func GetAdminToken(baseUrl string) (string,error) {
	request,err := NewLogInRequest(AdminCredentials(),baseUrl)
	if err!=nil{
		return "",err
	}
	resp, err := http.DefaultClient.Do(request)
	if err!=nil{
		return "",err
	}
	decoder := json.NewDecoder(resp.Body)
	var data map[string]string
	err=decoder.Decode(&data)
	if err!=nil{
		return "",err
	}
	token := data["token"]
	return token,err
}

func GetOwnerByAdminToken(baseUrl string) (string,error) {
	request,err := NewLogInRequest(OwnerByAdminCredentials(),baseUrl)
	if err!=nil{
		return "",err
	}
	resp, err := http.DefaultClient.Do(request)
	if err!=nil{
		return "",err
	}
	decoder := json.NewDecoder(resp.Body)
	var data map[string]string
	err=decoder.Decode(&data)
	if err!=nil{
		return "",err
	}
	token := data["token"]
	return token,err
}

func GetOwnerBySuperAdminToken(baseUrl string) (string,error) {
	request,err := NewLogInRequest(OwnerBySuperAdminCredentials(),baseUrl)
	if err!=nil{
		return "",err
	}
	resp, err := http.DefaultClient.Do(request)
	if err!=nil{
		return "",err
	}
	decoder := json.NewDecoder(resp.Body)
	var data map[string]string
	err=decoder.Decode(&data)
	if err!=nil{
		return "",err
	}
	token := data["token"]
	return token,err
}

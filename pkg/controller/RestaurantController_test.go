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

const dummySuperAdminID="0c647a43-5cef-443e-8688-aaea764d17d2"
const dummySuperOwnerID="32758fde-4dd5-4635-9bf7-a4d46d6f0629"//created by superadmin
const dummyOwnerID="451367e3-9b74-4bb6-9157-ac9a2c34da8d"//created by admin

func TestRestaurantController(t *testing.T){
	var DB, _ = mysql.NewMySqlDB("restaurant_test")
	defer DB.Close()
	defer  CleanDB(DB)
	svr,err:=server.NewServer(DB)
	router,_:=svr.Start()
	if err!=nil{
		panic(err)
	}
	token:=GetSuperToken(router)
	tokenAdmin:=GetAdminToken(router)
	testGetRestaurants:=[]struct{
		name string
		token string
	}{
		{"get restaurants with a valid token",token},
		{"get restaurants with a token of admin who has not created any restaurant",tokenAdmin},
	}
	for _,test:=range testGetRestaurants{
		t.Run(test.name,func(t *testing.T){
			request:=NewGetRestaurantRequest(test.token)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,http.StatusOK)
		})
	}
	//deleting restaurants
	CreateRestaurants(DB)
	testDeleteRestaurants:=[]struct{
		name string
		idArr []int
		wantStatus int
	}{
		{"tests for valid deletion",[]int{5,6,7},http.StatusOK},
		{"empty array of id",nil,http.StatusBadRequest},
		{"try to delete invalid id",[]int{10,11},http.StatusBadRequest},
	}
	for _,test:=range testDeleteRestaurants{
		t.Run(test.name,func(t *testing.T){
			idToDelete:=test.idArr
			request:=NewDeleteRestaurantRequest(token,idToDelete)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
		})
	}

	//add restaurant
	testCreateRestaurant:=[]struct{
		name string
		resName string
		wantStatus int
	}{
		{"Create restaurant with empty field","",http.StatusBadRequest},
		{"Create restaurant with valid entries","res100",http.StatusOK},
	}
	for _,test :=range testCreateRestaurant{
		t.Run(test.name,func(t *testing.T){
			request:=NewCreateRestaurantRequest(token,test.resName,1.2,5.0)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
		})
	}

	// updating restaurants
	tokenOwner:=GetOwnerToken(router)
	testUpdateRestaurant:=[]struct{
		name string
		resID int
		resName string
		token string
		wantStatus int
	}{
		{"update restaurants with empty field",3,"",token,http.StatusBadRequest},
		{"update restaurants with valid entries",3,"dummyRestaurant",token,http.StatusOK},
		{"update existing restaurant  by it's owner",3,"dummyRestaurant",tokenOwner,http.StatusUnauthorized},
		{"update non existing restaurant by superadmin",1,"dummyRestaurant",token,http.StatusBadRequest},
	}
	for _,test:=range testUpdateRestaurant{
		t.Run(test.name,func(t *testing.T){
			request:=NewUpdateRestaurantRequest(test.token,test.resID,test.resName,0.1,0.2)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
		})
	}

	//get restaurant for an owner
	testOwnerRestaurants:=[]struct{
		name string
		token string
		ownerID string
		wantStatus int
	}{
		{"getting restaurants of an owner by superadmin",token,dummySuperOwnerID,http.StatusOK},
		{"getting restaurants of an owner by admin",tokenAdmin,dummySuperOwnerID,http.StatusUnauthorized},
		{"getting restaurants of owner by admin created by him",tokenAdmin,dummyOwnerID,http.StatusOK},
	}
	for _,test:=range testOwnerRestaurants{
		t.Run(test.name,func(t *testing.T){
			request:=NewGetOwnerRestaurantRequest(test.token,test.ownerID)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
		})
	}

	//add owner to restaurants
	testAddOwnerRestaurants:=[]struct{
		name string
		token string
		ownerID string
		resID []int
		wantStatus int
	}{
		{"admin trying to add owner to a restaurant not created by him",tokenAdmin,dummySuperOwnerID,[]int{3},http.StatusUnauthorized},
		{"superAdmin trying to add owner to a restaurant",token,dummySuperOwnerID,[]int{3},http.StatusOK},
		{"empty restaurants field",token,dummySuperOwnerID,nil,http.StatusBadRequest},
		{"superAdmin trying to add a non existing owner to a restaurant",token,"1ajk",[]int{3},http.StatusBadRequest},
	}
	for _,test:=range testAddOwnerRestaurants{
		t.Run(test.name,func(t *testing.T){
			request:=NewAddOwnerRestaurantRequest(test.token,test.ownerID,test.resID)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
		})
	}

	////get available restaurants to add
	testGetAvailableRestaurants:=[]struct{
		name string
		token string
		wantStatus int
	}{
		{"get restaurants by superadmin(availabe here)",token,http.StatusOK},
		{"get restaurants by superadmin(not availabe here)",tokenAdmin,http.StatusOK},
	}
	for _,test:=range testGetAvailableRestaurants{
		t.Run(test.name,func(t *testing.T){
			request:=NewGetRestaurantAvailableRequest(test.token)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
		})
	}

	///get restaurants near by
	testGetNearByRestaurants:=[]struct{
		name string
		lat float32
		lng float32
		wantStatus int
	}{
		{"get restaurants nearby(available)",1.08,2.01,http.StatusOK},
		{"get restaurants nearby(not available)",11.08,12.01,http.StatusOK},
		{"empty lat and lng(value 0)",0,0,http.StatusBadRequest},
	}
	for _,test:=range testGetNearByRestaurants{
		t.Run(test.name,func(t *testing.T){
			request:=NewGetNearByRestaurants(test.lat,test.lng)
			response:=httptest.NewRecorder()
			router.ServeHTTP(response,request)
			assertStatus(t,response.Code,test.wantStatus)
		})
	}

}

///
func NewGetRestaurantRequest(token string) *http.Request{
	req, err := http.NewRequest(http.MethodGet, "/manage/restaurants",nil)
	if err!=nil{
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}
func NewCreateRestaurantRequest(token string,name string,lat float64,lng float64) *http.Request{
	restaurant:=models.Restaurant{
		Name: name,
		Lat: lat,
		Lng: lng,
	}
	data,err:=json.Marshal(restaurant)
	if err!=nil{
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPost, "/manage/restaurants", strings.NewReader(string(data)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}
func NewUpdateRestaurantRequest(token string,resID int,name string,lat float64,lng float64) *http.Request{
	restaurant:=models.Restaurant{
		Name: name,
		Lat: lat,
		Lng: lng,
	}
	data,err:=json.Marshal(restaurant)
	if err!=nil{
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/manage/restaurants/%d",resID),strings.NewReader(string(data)))
	if err!=nil{
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}

func CreateRestaurants(db *mysql.MySqlDB){
	stmt,_:=db.Prepare("insert into restaurants(name,lat,lng,creator_id,owner_id) values(?,?,?,?,null)")
	stmt.Exec("res10",10,10,dummySuperAdminID)
	stmt.Exec("res20",10,10,dummySuperAdminID)
	stmt.Exec("res30",10,10,dummySuperAdminID)
}
func NewDeleteRestaurantRequest(token string,idArr []int) *http.Request{
	var resID struct {
		IDArr []int	`json:"idArr"`
	}
	resID.IDArr=idArr
	data,err:=json.Marshal(resID)
	if err!=nil{
		panic(err)
	}
	req, _ := http.NewRequest(http.MethodDelete, "/manage/restaurants", strings.NewReader(string(data)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}
func NewGetOwnerRestaurantRequest(token string,ownerID string) *http.Request{
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/manage/owners/%s/restaurants",ownerID),nil)
	if err!=nil{
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}
func NewAddOwnerRestaurantRequest(token string,ownerID string,idArr []int) *http.Request{
	var resID struct {
		IDArr []int	`json:"idArr"`
	}
	resID.IDArr=idArr
	data,err:=json.Marshal(resID)
	if err!=nil{
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/manage/owners/%s/restaurants",ownerID),strings.NewReader(string(data)))
	if err!=nil{
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}
func NewGetRestaurantAvailableRequest(token string) *http.Request{
	req, err := http.NewRequest(http.MethodGet, "/manage/available/restaurants",nil)
	if err!=nil{
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token",token)
	return req
}

func NewGetNearByRestaurants(lat float32,lng float32) *http.Request{
	location:=map[string]float32{
		"lat":lat,
		"lng":lng,
	}
	data,err:=json.Marshal(location)
	if err!=nil{
		panic(err)
	}
	req, err := http.NewRequest(http.MethodGet, "/restaurantsNearBy",strings.NewReader(string(data)))
	if err!=nil{
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	return req
}
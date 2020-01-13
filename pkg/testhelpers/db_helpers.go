package testhelpers

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/vds/go-resman/pkg/database/mysql"
	"github.com/vds/go-resman/pkg/encryption"
	"github.com/vds/go-resman/pkg/models"
)

const (
	RestaurantTable   = "restaurants"
	MenuTable         = "dishes"
	InvalidTokenTable = "invalid_tokens"
)

var (
	admin = models.UserReg{
		Role:     "admin",
		Email:    "admin@gmail.com",
		Name:     "admin",
		Password: "pass",
	}
	adminForUD = models.UserReg{
		Role:     "admin",
		Email:    "admin100@gmail.com",
		Name:     "admin100",
		Password: "pass100",
	}
	superAdmin = models.UserReg{
		Role:     "superAdmin",
		Email:    "superAdmin@gmail.com",
		Name:     "superAdmin",
		Password: "superPass",
	}
	ownerByAdmin = models.UserReg{
		Role:     "owner",
		Email:    "ownerByAdmin@gmail.com",
		Name:     "ownerByAdmin",
		Password: "ownerByAdminPass",
	}
	ownerBySuperAdmin = models.UserReg{
		Role:     "owner",
		Email:    "ownerBySuperAdmin@gmail.com",
		Name:     "ownerBySuperAdmin",
		Password: "ownerBySuperAdmin",
	}
	restaurantByAdmin = models.Restaurant{
		Name: "restaurantByAdmin",
		Lat:  10,
		Lng:  15,
	}
	restaurantOfOwner = models.Restaurant{
		Name: "restaurantOfOwner",
		Lat:  5,
		Lng:  8,
	}
	ownerRestaurantDish = models.Dish{
		Name:  "ownerDish",
		Price: 100,
	}
	adminRestaurantDish = models.Dish{
		Name:  "adminDish",
		Price: 10,
	}

	adminId             string
	superAdminId        string
	ownerByAdminId      string
	ownerBySuperAdminId string
	adminForUDId        string
	Db *mysql.MySqlDB
)

func InitDB(db *mysql.MySqlDB) error {
	Db = db
	err := createAdmins(db)
	if err != nil {
		return err
	}
	err = createOwners(db)
	if err != nil {
		return err
	}
	err = createRestaurants(db)
	if err != nil {
		return err
	}
	err = createDishes(db)
	if err != nil {
		return err
	}
	return nil
}

func createAdmins(db *mysql.MySqlDB) error {
	adminId = uuid.New().String()
	adminForUDId = uuid.New().String()
	superAdminId = uuid.New().String()
	adminPass, err := encryption.GenerateHash(admin.Password)
	if err != nil {
		return err
	}
	superAdminPass, err := encryption.GenerateHash(superAdmin.Password)
	if err != nil {
		return err
	}
	adminForUDPass, err := encryption.GenerateHash(superAdmin.Password)
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf(
		`insert into %s(id,email_id,name,password) 
							value(?,?,?,?)
			`, mysql.AdminTable), adminId, admin.Email, admin.Name, adminPass)
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf(
		`insert into %s(id,email_id,name,password) 
							value(?,?,?,?)
			`, mysql.AdminTable), adminForUDId, adminForUD.Email, adminForUD.Name, adminForUDPass)
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf(
		`insert into %s(id,email_id,name,password) 
							value(?,?,?,?)
			`, mysql.SuperAdminTable), superAdminId, superAdmin.Email, superAdmin.Name, superAdminPass)
	if err != nil {
		return err
	}
	return nil
}

func createOwners(db *mysql.MySqlDB) error {
	ownerByAdminId = uuid.New().String()
	ownerBySuperAdminId = uuid.New().String()
	ownerByAdminPass, err := encryption.GenerateHash(ownerByAdmin.Password)
	if err != nil {
		return err
	}
	ownerBySuperAdminPass, err := encryption.GenerateHash(ownerBySuperAdmin.Password)
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf(
		`insert into %s(id,email_id,name,password,creator_id) 
							value(?,?,?,?,?)
			`, mysql.OwnerTable), ownerByAdminId, ownerByAdmin.Email, ownerByAdmin.Name, ownerByAdminPass, adminId)
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf(
		`insert into %s(id,email_id,name,password,creator_id) 
							value(?,?,?,?,?)
			`, mysql.OwnerTable), ownerBySuperAdminId, ownerBySuperAdmin.Email, ownerBySuperAdmin.Name, ownerBySuperAdminPass, superAdminId)
	if err != nil {
		return err
	}
	return nil
}

func createRestaurants(db *mysql.MySqlDB) error {
	_, err := db.Exec(fmt.Sprintf(
		`insert into %s(name,lat,lng,creator_id) 
							value(?,?,?,?)
			`, RestaurantTable), restaurantByAdmin.Name, restaurantByAdmin.Lat, restaurantByAdmin.Lng, adminId)
	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf(
		`insert into %s(name,lat,lng,creator_id,owner_id) 
							value(?,?,?,?,?)
			`, RestaurantTable), restaurantOfOwner.Name, restaurantOfOwner.Lat, restaurantOfOwner.Lng, superAdminId, ownerByAdminId)
	if err != nil {
		return err
	}
	return nil
}

func createDishes(db *mysql.MySqlDB) error {
	_, err := db.Exec(fmt.Sprintf(
		`insert into %s(name,price,res_id) 
							value(?,?,?)
			`, MenuTable), adminRestaurantDish.Name, adminRestaurantDish.Price, 1)
	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf(
		`insert into %s(name,price,res_id) 
							value(?,?,?)
			`, MenuTable), ownerRestaurantDish.Name, ownerRestaurantDish.Price, 2)
	if err != nil {
		return err
	}
	return nil
}

func ClearDB(db *mysql.MySqlDB) error {
	_, err := db.Exec(fmt.Sprintf("delete from %s", mysql.AdminTable))
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf("delete from %s", mysql.SuperAdminTable))
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf("delete from %s", mysql.OwnerTable))
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf("delete from %s", RestaurantTable))
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf("delete from %s", MenuTable))
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf("delete from %s", InvalidTokenTable))
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf("alter table %s  AUTO_INCREMENT=1", RestaurantTable))
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf("alter table %s  AUTO_INCREMENT=1", MenuTable))
	if err != nil {
		return err
	}
	return nil
}

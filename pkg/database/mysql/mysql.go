package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/mysql"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/google/uuid"
	"github.com/vds/go-resman/pkg/database"
	"github.com/vds/go-resman/pkg/encryption"
	"github.com/vds/go-resman/pkg/logger"
	"github.com/vds/go-resman/pkg/middleware"
	"github.com/vds/go-resman/pkg/models"
	"log"
	"os"
)

const (
	SuperAdminTable           = "super_admins"
	AdminTable                = "admins"
	OwnerTable                = "owners"
	InsertUser                = "insert into %s(id,email_id,name,password) values(?,?,?,?)"
	GetUserIDPassword         = "select id,password from %s where email_id=?"
	GetOwnersForSuperAdmin    = "select JSON_ARRAYAGG(JSON_OBJECT('id',id,'email',email_id,'name', name)) from owners order by id"
	InsertOwner               = "insert into owners(id,email_id,name,password,creator_id) values(?,?,?,?,?)"
	OwnerUpdate               = "update owners set email_id=?,name=? where id=?"
	SelectRestaurantsForSuper = "select JSON_ARRAYAGG(JSON_OBJECT('id',id,'name',name, 'lat',lat,'lng',lng)) from restaurants order by id"
	SelectRestaurantsForAdmin = "select JSON_ARRAYAGG(JSON_OBJECT('id',id,'name', name, 'lat',lat,'lng',lng)) from restaurants  where creator_id=? order by id"
	SelectRestaurantsForOwner = "select JSON_ARRAYAGG(JSON_OBJECT('id',id,'name', name, 'lat',lat,'lng',lng)) from restaurants  where owner_id=? order by id"
	InsertRestaurant          = "insert into restaurants(name,lat,lng,creator_id) values(?,?,?,?)"
	RestaurantUpdate          = "update restaurants set name=?,lat=?,lng=? where id=?"
	CheckRestaurantOwner      = "select owner_id from restaurants where id=?"
	CheckRestaurantCreator    = "select creator_id from restaurants where id=?"
	CheckRestaurantDish       = "select res_id from dishes where id=?"

	DeleteOwnerBySuperAdmin = "delete from owners where id=?"
	DeleteOwnerByAdmin      = "delete from owners where id=? and creator_id=?"

	DeleteRestaurantsBySuperAdmin = "delete from restaurants where id=?"
	DeleteRestaurantsByAdmin      = "delete from restaurants where id=? and creator_id=?"

	DeleteDishes = "delete from dishes where id=?"
)

type MySqlDB struct {
	*sql.DB
}

func NewMySqlDB(dbUrl string) (*MySqlDB, error) {
	////for docker servicename:3306 e.g. database:3306
	//serverName := "localhost:3306"
	//user := "root"
	//password := "password"

	//connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true&multiStatements=true", user, password, serverName, dbName)
	db, err := sql.Open("mysql", dbUrl)
		err = migrateDatabase(db)
		if err != nil {
			fmt.Print(err)
			return nil, err
		}
	mySqlDB := &MySqlDB{db}
	return mySqlDB, err
}

func (db *MySqlDB) ShowNearBy(ctx context.Context, location *models.Location) (string, error) {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogDebug(reqId, reqUrl, "executing GetNearByRestaurant query")
	var result sql.NullString
	rows, err := db.Query("select JSON_ARRAYAGG(JSON_OBJECT('id',id,'name',name, 'lat',lat,'lng',lng)) from restaurants where ST_Distance_Sphere(point(lat,lng),point(?,?))/1000 < 10", location.Lat, location.Lng)
	defer rows.Close()
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return "", database.ErrInternal
	}
	rows.Next()
	err = rows.Scan(&result)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in storing result in variable: %v", err), 0)
		return "", database.ErrInternal
	}
	logger.LogInfo(reqId, reqUrl, "getNearBy restaurant from db successful", 0)
	return result.String, nil
}

func (db *MySqlDB) CreateUser(ctx context.Context, user *models.UserReg) (string, error) {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogDebug(reqId, reqUrl, "executing createUser query")
	var tableName string
	switch user.Role {
	case middleware.Admin:
		tableName = AdminTable
	case middleware.SuperAdmin:
		tableName = SuperAdminTable
	}
	stmt, err := db.Prepare(fmt.Sprintf(InsertUser, tableName))
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return "", database.ErrInternal
	}
	logger.LogDebug(reqId, reqUrl, "generating password hash")
	pass, err := encryption.GenerateHash(user.Password)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in generating password hash: %v", err), 0)
		return "", database.ErrInternal
	}
	id := uuid.New().String()
	_, err = stmt.Exec(id, user.Email, user.Name, pass)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return "", database.ErrDupEmail
	}
	logger.LogInfo(reqId, reqUrl, "createUser in db successful", 0)
	return id, nil
}

func (db *MySqlDB) LogInUser(ctx context.Context, cred *models.Credentials) (string, error) {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogDebug(reqId, reqUrl, "logging in user")
	var tableName string
	switch cred.Role {
	case middleware.Admin:
		tableName = AdminTable
	case middleware.SuperAdmin:
		tableName = SuperAdminTable
	case middleware.Owner:
		tableName = OwnerTable
	}
	var id string
	var pass string
	rows, err := db.Query(fmt.Sprintf(GetUserIDPassword, tableName), cred.Email)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return "", database.ErrInternal
	}
	defer rows.Close()
	rows.Next()
	err = rows.Scan(&id, &pass)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in storing result in variable: %v", err), 0)
		return "", database.ErrInvalidCredentials
	}
	logger.LogDebug(reqId, reqUrl, "comparing passwords")
	isValid := encryption.ComparePasswords(ctx, pass, cred.Password)
	if !isValid {
		return "", database.ErrInvalidCredentials
	}

	logger.LogInfo(reqId, reqUrl, "login credentials verified from db", 0)
	return id, nil
}

func (db *MySqlDB) ShowOwners(ctx context.Context, userAuth *models.UserAuth) (string, error) {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogInfo(reqId, reqUrl, "selecting function to show owners according to role", 0)
	if userAuth.Role == middleware.SuperAdmin {
		return showOwnersForSuperAdmin(ctx, db)
	} else if userAuth.Role == middleware.Admin {
		return showOwnersForAdmin(ctx, db, userAuth.ID)
	}
	return "", database.ErrInternal
}

func (db *MySqlDB) CreateOwner(ctx context.Context, creatorID string, owner *models.OwnerReg) (*models.UserOutput, error) {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	stmt, err := db.Prepare(InsertOwner)
	logger.LogDebug(reqId, reqUrl, "executing query to create an owner")
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return nil, database.ErrInternal
	}
	logger.LogDebug(reqId, reqUrl, "generating password hash")
	pass, err := encryption.GenerateHash(owner.Password)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in generating hash of password: %v", err), 0)
		return nil, database.ErrInternal
	}
	id := uuid.New().String()
	_, err = stmt.Exec(id, owner.Email, owner.Name, pass, creatorID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return nil, database.ErrDupEmail
	}
	logger.LogDebug(reqId, reqUrl, "fetching created owner")
	var result models.UserOutput
	rows, err := db.Query("select id,name,email_id from owners where email_id=?", owner.Email)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return nil, database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&result.ID, &result.Name, &result.Email)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in storing result in variable: %v", err), 0)
		return nil, database.ErrInternal
	}
	logger.LogInfo(reqId, reqUrl, "owner created in db successful", 0)
	return &result, nil
}

func (db *MySqlDB) ShowAdmins(ctx context.Context) (string, error) {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	var result string
	logger.LogDebug(reqId, reqUrl, "executing query to get admins")
	rows, err := db.Query("select JSON_ARRAYAGG(JSON_OBJECT('id',id,'email',email_id,'name', name)) from admins")
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return "", database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&result)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in storing result in variable: %v", err), 0)
		return "", database.ErrInternal
	}

	logger.LogInfo(reqId, reqUrl, "get admins from db successful", 0)
	return result, nil
}

func (db *MySqlDB) CheckAdmin(ctx context.Context, adminID string) error {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	var count int
	logger.LogDebug(reqId, reqUrl, "executing query to check if admin exist")
	rows, err := db.Query("select count(*) from admins where id=?", adminID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&count)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in storing result in variable: %v", err), 0)
		return database.ErrInternal
	}
	if count != 1 {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("db error more than one account with same email: %v", err), 0)
		return database.ErrInternal
	}

	logger.LogInfo(reqId, reqUrl, "admin id verified from db", 0)
	return nil
}
func (db *MySqlDB) UpdateAdmin(ctx context.Context, admin *models.UserOutput) (string, error) {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogDebug(reqId, reqUrl, "executing query to update an admin")
	stmt, err := db.Prepare("update admins set email_id=?,name=? where id=?")
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in preparing query statement: %v", err), 0)
		return "", database.ErrInternal
	}
	_, err = stmt.Exec(admin.Email, admin.Name, admin.ID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return "", database.ErrInternal
	}
	logger.LogDebug(reqId, reqUrl, "executing query to fetch updated admin")
	var result sql.NullString
	rows, err := db.Query("select JSON_OBJECT('id',id,'email',email_id,'name', name) from admins where id=?", admin.ID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return "", database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&result)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in storing result in variable: %v", err), 0)
		return "", database.ErrInternal
	}
	logger.LogInfo(reqId, reqUrl, "admin updated in db successfully", 0)
	return result.String, nil
}
func (db *MySqlDB) RemoveAdmins(ctx context.Context, adminIDs ...string) error {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	var ErrEntries []int
	logger.LogDebug(reqId, reqUrl, "executing query to delete admin")
	stmt, err := db.Prepare("delete from admins where id=?")
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in preparing statement: %v", err), 0)
		return database.ErrInternal
	}
	for i, id := range adminIDs {
		result, err := stmt.Exec(id)
		if err != nil {
			logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
			return database.ErrInternal
		}
		numDeletedRows, _ := result.RowsAffected()
		if numDeletedRows == 0 {
			ErrEntries = append(ErrEntries, i)
		}
	}
	length := len(ErrEntries)
	if length != 0 {
		return sendErrorMessage(ctx, ErrEntries, length, "Admins")
	}
	logger.LogInfo(reqId, reqUrl, "admin deleted in db successfully", 0)
	return nil
}

func (db *MySqlDB) CheckOwnerCreator(ctx context.Context, creatorID string, ownerID string) error {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	var creatorIDOut string
	logger.LogDebug(reqId, reqUrl, "executing query to verify owner creator")
	rows, err := db.Query("select creator_id from owners where id=?", ownerID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&creatorIDOut)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in storing result in variable: %v", err), 0)
		return database.ErrInvalidOwner
	}
	if creatorIDOut != creatorID {
		return database.ErrInvalidOwnerCreator
	}
	logger.LogInfo(reqId, reqUrl, "owner creator verified from db", 0)
	return nil
}

func (db *MySqlDB) UpdateOwner(ctx context.Context, owner *models.UserOutput) (string, error) {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogDebug(reqId, reqUrl, "checking that owner exist")
	isValidOwnerID := CheckOwnerID(ctx, db, owner.ID)
	if !isValidOwnerID {
		return "", database.ErrInvalidOwner
	}
	logger.LogDebug(reqId, reqUrl, "executing query to update owner")
	stmt, err := db.Prepare(OwnerUpdate)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in preparing statement: %v", err), 0)
		return "", database.ErrInternal
	}
	_, err = stmt.Exec(owner.Email, owner.Name, owner.ID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return "", database.ErrDupEmail
	}
	logger.LogDebug(reqId, reqUrl, "fetching updated owner")
	var result sql.NullString
	rows, err := db.Query("select JSON_OBJECT('id',id,'email',email_id,'name', name) from owners where id=?", owner.ID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return "", database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&result)
	if err != nil {
		log.Printf("%v", err)
		return "", database.ErrInternal
	}
	logger.LogInfo(reqId, reqUrl, "owner updated in db successfully", 0)
	return result.String, nil
}

//func (db *MySqlDB) GetOwnerID(ctx context.Context, ownerEmail string) (string, error) {
//	reqIdVal := ctx.Value("reqId")
//	reqId := reqIdVal.(string)
//	reqUrlVal := ctx.Value("reqUrl")
//	reqUrl := reqUrlVal.(string)
//	id := ""
//	rows, err := db.Query(GetOwnerID, ownerEmail)
//	if err != nil {
//		logger.LogError(reqId,reqUrl,fmt.Sprintf("error in executing query: %v",err),0)
//		return "", database.ErrInternal
//	}
//	defer rows.Close()
//
//	rows.Next()
//	err = rows.Scan(&id)
//	if err != nil {
//		fmt.Print(err)
//		return "", database.ErrInvalidOwner
//	}
//	if id == "" {
//		return "", database.ErrInvalidOwner
//	}
//	return id, nil
//}

func (db *MySqlDB) RemoveOwners(ctx context.Context, userAuth *models.UserAuth, ownerIDs ...string) error {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogInfo(reqId, reqUrl, "selecting  owner delete function as per role", 0)
	switch userAuth.Role {
	case middleware.SuperAdmin:
		return removeOwnersBySuperAdmin(ctx, db, ownerIDs...)
	case middleware.Admin:
		return removeOwnersByAdmin(ctx, db, userAuth.ID, ownerIDs...)
	}
	return database.ErrInternal
}

//restaurants

func (db *MySqlDB) ShowRestaurants(ctx context.Context, userAuth *models.UserAuth) (string, error) {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogInfo(reqId, reqUrl, "selecting  show restaurant function as per role", 0)
	switch userAuth.Role {
	case middleware.SuperAdmin:
		return showRestaurantsForSuper(ctx, db)
	case middleware.Admin:
		return showRestaurantsForAdmin(ctx, db, userAuth.ID)
	case middleware.Owner:
		return showRestaurantsForOwner(ctx, db, userAuth.ID)
	}
	return "", database.ErrInternal
}

func (db *MySqlDB) InsertRestaurant(ctx context.Context, restaurant *models.Restaurant) (*models.RestaurantOutput, error) {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogDebug(reqId, reqUrl, "executing query to add a restaurant")
	stmt, err := db.Prepare(InsertRestaurant)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in preparing statement: %v", err), 0)
		return nil, database.ErrInternal
	}
	_, err = stmt.Exec(restaurant.Name, restaurant.Lat, restaurant.Lng, restaurant.CreatorID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return nil, database.ErrInternal
	}

	var result models.RestaurantOutput
	rows, err := db.Query("select id,name,lat,lng from restaurants order by id desc limit 1")
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return nil, database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&result.ID, &result.Name, &result.Lat, &result.Lng)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in storing result in variable: %v", err), 0)
		return nil, database.ErrInternal
	}
	logger.LogInfo(reqId, reqUrl, "restaurant added successfully in db", 0)
	return &result, nil
}

func (db *MySqlDB) CheckRestaurantCreator(ctx context.Context, creatorID string, resID int) error {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	var creatorIDOut string
	logger.LogDebug(reqId, reqUrl, "executing query to check restaurant creator")
	rows, err := db.Query(CheckRestaurantCreator, resID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&creatorIDOut)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in storing result in variable: %v", err), 0)
		return database.ErrNonExistingRestaurant
	}
	if creatorIDOut != creatorID {
		logger.LogError(reqId, reqUrl, "error invalid creator", 0)
		return database.ErrInvalidRestaurantCreator
	}
	logger.LogInfo(reqId, reqUrl, "restaurant creator verified", 0)
	return nil
}

func (db *MySqlDB) UpdateRestaurant(ctx context.Context, restaurant *models.RestaurantOutput) (*models.RestaurantOutput, error) {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	isValidRestaurant := CheckRestaurantID(ctx, db, restaurant.ID)
	if !isValidRestaurant {
		return nil, database.ErrNonExistingRestaurant
	}
	var err error
	logger.LogDebug(reqId, reqUrl, "executing query to update a restaurant")
	stmt, err := db.Prepare(RestaurantUpdate)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in preparing statement: %v", err), 0)
		return nil, database.ErrInternal
	}
	_, err = stmt.Exec(restaurant.Name, restaurant.Lat, restaurant.Lng, restaurant.ID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return nil, database.ErrInternal
	}
	logger.LogDebug(reqId, reqUrl, "executing query to fetch updated  restaurant")

	var result models.RestaurantOutput
	rows, err := db.Query("select id,name,lat,lng from restaurants where id=?", restaurant.ID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return nil, database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&result.ID, &result.Name, &result.Lat, &result.Lng)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in storing result in variable: %v", err), 0)
		return nil, database.ErrInternal
	}
	logger.LogInfo(reqId, reqUrl, "restaurant updated in db successfully", 0)
	return &result, nil
}

func (db *MySqlDB) RemoveRestaurants(ctx context.Context, userAuth *models.UserAuth, resIDs ...int) error {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogDebug(reqId, reqUrl, "selecting function to delete restaurant according to role")
	switch userAuth.Role {
	case middleware.SuperAdmin:
		return removeRestaurantsBySuperAdmin(ctx, db, resIDs...)
	case middleware.Admin:
		return removeRestaurantsByAdmin(ctx, db, userAuth.ID, resIDs...)
	}
	return database.ErrInternal
}

func (db *MySqlDB) ShowAvailableRestaurants(ctx context.Context, userAuth *models.UserAuth) (string, error) {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogDebug(reqId, reqUrl, "executing query to show available restaurant according to role")
	switch userAuth.Role {
	case middleware.SuperAdmin:
		return showAvailableRestaurantsForSuper(ctx, db)
	case middleware.Admin:
		return showAvailableRestaurantsForAdmin(ctx, db, userAuth.ID)
	}
	return "", database.ErrInternal
}

func (db *MySqlDB) InsertOwnerForRestaurants(ctx context.Context, userAuth *models.UserAuth, ownerID string, resIDs ...int) error {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	isValidOwnerID := CheckOwnerID(ctx, db, ownerID)
	if !isValidOwnerID {
		return database.ErrInvalidOwner
	}
	var ErrEntries []int
	logger.LogDebug(reqId, reqUrl, "executing query to add owner to restaurants")
	stmt, err := db.Prepare("update restaurants set owner_id=? where id=?")
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in preparing statement: %v", err), 0)
		return database.ErrInternal
	}
	for i, id := range resIDs {
		if userAuth.Role != middleware.SuperAdmin {
			err = db.CheckRestaurantCreator(ctx, userAuth.ID, id)
			if err != nil {
				ErrEntries = append(ErrEntries, i)
				continue
			}
		}
		_, err := stmt.Exec(ownerID, id)
		if err != nil {
			logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
			return database.ErrInternal
		}
	}
	length := len(ErrEntries)
	if length != 0 {
		return sendErrorMessage(ctx, ErrEntries, length, "Restaurants")
	}
	logger.LogInfo(reqId, reqUrl, "owner assigned for the requested restaurants in db successfully", 0)
	return nil
}

func (db *MySqlDB) RemoveOwnerForRestaurants(ctx context.Context, userAuth *models.UserAuth, ownerID string, resIDs ...int) error {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	isValidOwnerID := CheckOwnerID(ctx, db, ownerID)
	if !isValidOwnerID {
		return database.ErrInvalidOwner
	}
	var ErrEntries []int
	logger.LogDebug(reqId, reqUrl, "executing query to remove owner of restaurants")
	stmt, err := db.Prepare("update restaurants set owner_id=null where id=?")
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in preparing statement: %v", err), 0)
		return database.ErrInternal
	}
	for i, id := range resIDs {
		if userAuth.Role != middleware.SuperAdmin {
			err = db.CheckRestaurantCreator(ctx, userAuth.ID, id)
			if err != nil {
				ErrEntries = append(ErrEntries, i)
				continue
			}
		}
		_, err := stmt.Exec(id)
		if err != nil {
			logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
			return database.ErrInternal
		}
	}
	length := len(ErrEntries)
	if length != 0 {
		return sendErrorMessage(ctx, ErrEntries, length, "Restaurants")
	}
	logger.LogInfo(reqId, reqUrl, "owner removed of the requested restaurants in db successfully", 0)
	return nil
}

//menu
func (db *MySqlDB) CheckRestaurantOwner(ctx context.Context, ownerID string, resID int) error {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogDebug(reqId, reqUrl, "executing query to verify restaurant owner")
	var ownerIDOut string
	rows, err := db.Query(CheckRestaurantOwner, resID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&ownerIDOut)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in storing result in variable: %v", err), 0)
		return database.ErrNonExistingRestaurant
	}
	if ownerIDOut != ownerID {
		logger.LogError(reqId, reqUrl, "error invalid restaurant owner", 0)
		return database.ErrInvalidRestaurantOwner
	}
	logger.LogInfo(reqId, reqUrl, "restaurant owner validated from db successfully", 0)
	return nil
}
func (db *MySqlDB) ShowMenu(ctx context.Context, resID int) (string, error) {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogDebug(reqId, reqUrl, "executing query to get menu")
	isValidRestaurant := CheckRestaurantID(ctx, db, resID)
	if !isValidRestaurant {
		return "", database.ErrNonExistingRestaurant
	}
	var result string
	rows, err := db.Query("select JSON_ARRAYAGG(JSON_OBJECT('id',id,'name',name,'price',price)) from dishes where res_id=?", resID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return "", database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&result)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in storing result in variable: %v", err), 0)
		return "", database.ErrInternal
	}
	logger.LogInfo(reqId, reqUrl, "menu retrieved from db successfully", 0)
	return result, nil
}
func (db *MySqlDB) InsertDishes(ctx context.Context, dish models.Dish, resID int) (*models.DishOutput, error) {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogDebug(reqId, reqUrl, "executing query to insert dish")
	stmt, err := db.Prepare("insert into dishes(res_id,name,price) values(?,?,?)")
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in preparing statement: %v", err), 0)
		return nil, database.ErrInternal
	}
	_, err = stmt.Exec(resID, dish.Name, dish.Price)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return nil, database.ErrNonExistingRestaurant
	}
	logger.LogDebug(reqId, reqUrl, "executing query to fetch added dish")
	var addedDish models.DishOutput
	rows, err := db.Query("select id,name,price from dishes order by id desc limit 1")
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return nil, database.ErrInternal
	}
	defer rows.Close()
	rows.Next()
	err = rows.Scan(&addedDish.ID, &addedDish.Name, &addedDish.Price)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in storing result in variable: %v", err), 0)
		return nil, database.ErrInternal
	}
	logger.LogInfo(reqId, reqUrl, "dish added in db successfully", 0)
	return &addedDish, nil
}
func (db *MySqlDB) UpdateDish(ctx context.Context, dish *models.DishOutput) (*models.DishOutput, error) {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogDebug(reqId, reqUrl, "executing query to update dish")
	stmt, err := db.Prepare("update dishes set name=?,price=? where id=?")
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in preparing statement: %v", err), 0)
		return nil, database.ErrInternal
	}
	_, err = stmt.Exec(dish.Name, dish.Price, dish.ID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return nil, database.ErrInternal
	}
	logger.LogDebug(reqId, reqUrl, "executing query to fetch updated dish")
	var updatedDish models.DishOutput
	rows, err := db.Query("select id,name,price from dishes where id=?", dish.ID);
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return nil, database.ErrInternal
	}
	defer rows.Close()
	rows.Next()
	err = rows.Scan(&updatedDish.ID, &updatedDish.Name, &updatedDish.Price)
	if err != nil {
		log.Printf("%v", err)
		return nil, database.ErrInternal
	}
	logger.LogInfo(reqId, reqUrl, "dish updated in db successfully", 0)
	return &updatedDish, nil
}
func (db *MySqlDB) CheckRestaurantDish(ctx context.Context, resID int, dishID int) error {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	var resIDOut int
	logger.LogDebug(reqId, reqUrl, "executing query to check the requested dish in the restaurant")
	rows, err := db.Query(CheckRestaurantDish, dishID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&resIDOut)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in storing result in variable: %v", err), 0)
		return database.ErrInvalidDish
	}
	if resIDOut != resID {
		logger.LogError(reqId, reqUrl, "requested dish is of some other restaurant", 0)
		return database.ErrInvalidRestaurantDish
	}
	logger.LogInfo(reqId, reqUrl, "checking of the requested dish in the restaurant successful", 0)
	return nil
}

func (db *MySqlDB) RemoveDishes(ctx context.Context, dishIDs ...int) error {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	var ErrEntries []int
	logger.LogDebug(reqId, reqUrl, "executing query to delete the dish")
	stmt, err := db.Prepare(DeleteDishes)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in preparing statement: %v", err), 0)
		return database.ErrInternal
	}
	for i, id := range dishIDs {
		result, err := stmt.Exec(id)
		if err != nil {
			logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
			return database.ErrInternal
		}
		numDeletedRows, _ := result.RowsAffected()
		if numDeletedRows == 0 {
			ErrEntries = append(ErrEntries, i)
		}
	}
	length := len(ErrEntries)
	if length != 0 {
		return sendErrorMessage(ctx, ErrEntries, length, "Dishes")
	}
	logger.LogInfo(reqId, reqUrl, "dish deleted in db successfully", 0)
	return nil
}

//helpers
func showOwnersForSuperAdmin(ctx context.Context, db *MySqlDB) (string, error) {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	var result string
	logger.LogDebug(reqId, reqUrl, "executing query to get owners for superAdmin")
	rows, err := db.Query(GetOwnersForSuperAdmin)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return "", database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&result)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in storing result in variable: %v", err), 0)
		return "", database.ErrInternal
	}
	logger.LogInfo(reqId, reqUrl, "owners retrieved for superAdmin from db successfully", 0)
	return result, nil
}
func showOwnersForAdmin(ctx context.Context, db *MySqlDB, creatorID string) (string, error) {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	var result sql.NullString
	logger.LogDebug(reqId, reqUrl, "executing query to get owners for admin")
	rows, err := db.Query("select JSON_ARRAYAGG(JSON_OBJECT('id',id,'email',email_id,'name', name)) from owners where creator_id=?", creatorID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return "", database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&result)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in storing result in variable: %v", err), 0)
		return "", database.ErrInternal
	}
	logger.LogInfo(reqId, reqUrl, "owners retrieved for admin from db successfully", 0)
	return result.String, nil
}

func showRestaurantsForSuper(ctx context.Context, db *MySqlDB) (string, error) {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	var result string
	logger.LogDebug(reqId, reqUrl, "executing query to get restaurants for superAdmin")
	rows, err := db.Query(SelectRestaurantsForSuper)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return "", database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&result)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in storing result in variable: %v", err), 0)
		return "", database.ErrInternal
	}
	logger.LogInfo(reqId, reqUrl, "restaurants retrieved for superAdmin from db successfully", 0)
	return result, nil
}
func showRestaurantsForAdmin(ctx context.Context, db *MySqlDB, adminID string) (string, error) {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	var result string
	logger.LogDebug(reqId, reqUrl, "executing query to get restaurants for admin")

	rows, err := db.Query(SelectRestaurantsForAdmin, adminID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return "", database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&result)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in storing result in variable: %v", err), 0)
		return "", database.ErrInternal
	}
	logger.LogInfo(reqId, reqUrl, "restaurants retrieved for admin from db successfully", 0)
	return result, nil
}
func showRestaurantsForOwner(ctx context.Context, db *MySqlDB, ownerID string) (string, error) {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	var result sql.NullString
	logger.LogDebug(reqId, reqUrl, "executing query to get restaurants for owner")

	rows, err := db.Query(SelectRestaurantsForOwner, ownerID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return "", database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&result)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in storing result in variable: %v", err), 0)
		return "", database.ErrInternal
	}
	logger.LogInfo(reqId, reqUrl, "restaurants retrieved for owner from db successfully", 0)
	return result.String, nil
}

func showAvailableRestaurantsForSuper(ctx context.Context, db *MySqlDB) (string, error) {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	var result string
	logger.LogDebug(reqId, reqUrl, "executing query to get available restaurants for superAdmin")
	rows, err := db.Query("select JSON_ARRAYAGG(JSON_OBJECT('id',id,'name',name, 'lat',lat,'lng',lng)) from restaurants where owner_id IS NULL")
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return "", database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&result)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in storing result in variable: %v", err), 0)
		return "", database.ErrInternal
	}
	logger.LogInfo(reqId, reqUrl, "available restaurants retrieved for superAdmin from db successfully", 0)
	return result, nil
}
func showAvailableRestaurantsForAdmin(ctx context.Context, db *MySqlDB, creatorID string) (string, error) {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	var result string
	logger.LogDebug(reqId, reqUrl, "executing query to get available restaurants for superAdmin")
	rows, err := db.Query("select JSON_ARRAYAGG(JSON_OBJECT('id',id,'name',name, 'lat',lat,'lng',lng)) from restaurants where owner_id IS NULL and creator_id=?", creatorID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return "", database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&result)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in storing result in variable: %v", err), 0)
		return "", database.ErrInternal
	}
	logger.LogInfo(reqId, reqUrl, "available restaurants retrieved for admin from db successfully", 0)
	return result, nil
}

func removeOwnersBySuperAdmin(ctx context.Context, db *MySqlDB, ownerIDs ...string) error {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	var ErrEntries []int
	logger.LogDebug(reqId, reqUrl, "executing query to delete owner by superAdmin")

	stmt, err := db.Prepare(DeleteOwnerBySuperAdmin)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in preparing statement: %v", err), 0)
		return database.ErrInternal
	}
	for i, id := range ownerIDs {
		result, err := stmt.Exec(id)
		if err != nil {
			logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
			return database.ErrInternal
		}
		numDeletedRows, _ := result.RowsAffected()
		if numDeletedRows == 0 {
			ErrEntries = append(ErrEntries, i)
		} else {
			_, _ = db.Query("Update Restaurants set owner_id=null where owner_id=?", id)
		}
	}
	length := len(ErrEntries)
	if length != 0 {
		return sendErrorMessage(ctx, ErrEntries, length, "Owners")
	}
	logger.LogInfo(reqId, reqUrl, "owner deleted by superAdmin from db successfully", 0)
	return nil
}
func removeOwnersByAdmin(ctx context.Context, db *MySqlDB, creatorID string, ownerIDs ...string) error {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	var ErrEntries []int
	logger.LogDebug(reqId, reqUrl, "executing query to delete owner by admin")

	stmt, err := db.Prepare(DeleteOwnerByAdmin)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in preparing statement: %v", err), 0)
		return database.ErrInternal
	}
	for i, id := range ownerIDs {
		result, err := stmt.Exec(id, creatorID)
		if err != nil {
			logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
			return database.ErrInternal
		}
		numDeletedRows, _ := result.RowsAffected()
		if numDeletedRows == 0 {
			ErrEntries = append(ErrEntries, i)
		} else {
			_, err = db.Exec("Update Restaurants set owner_id=null where owner_id=?", id)
			if err != nil {
				logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
				return database.ErrInternal
			}
		}
	}
	length := len(ErrEntries)
	if length != 0 {
		return sendErrorMessage(ctx, ErrEntries, length, "Owners")
	}
	logger.LogInfo(reqId, reqUrl, "owner deleted by admin from db successfully", 0)
	return nil
}
func sendErrorMessage(ctx context.Context, ErrEntries []int, length int, data string) error {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogInfo(reqId, reqUrl, "generating error message", 0)
	errMsg := data + " Deleted Except entry no. "
	for i, j := range ErrEntries {
		if i == length-1 {
			errMsg = errMsg + fmt.Sprintf(" %v", j+1)
			break
		}
		errMsg = errMsg + fmt.Sprintf(" %v,", j+1)
	}
	logger.LogInfo(reqId, reqUrl, "error message generated successfully", 0)
	return errors.New(errMsg)
}

func removeRestaurantsBySuperAdmin(ctx context.Context, db *MySqlDB, resIDs ...int) error {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	var ErrEntries []int
	logger.LogDebug(reqId, reqUrl, "executing query to delete restaurant by superAdmin")

	stmt, err := db.Prepare(DeleteRestaurantsBySuperAdmin)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in preparing statement: %v", err), 0)
		return database.ErrInternal
	}
	for i, id := range resIDs {
		result, err := stmt.Exec(id)
		if err != nil {
			logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
			return database.ErrInternal
		}
		numDeletedRows, _ := result.RowsAffected()
		if numDeletedRows == 0 {
			ErrEntries = append(ErrEntries, i)
		}
	}
	length := len(ErrEntries)
	if length != 0 {
		return sendErrorMessage(ctx, ErrEntries, length, "Restaurants")
	}
	logger.LogInfo(reqId, reqUrl, "restaurant deleted by superAdmin from db successfully", 0)
	return nil
}
func removeRestaurantsByAdmin(ctx context.Context, db *MySqlDB, creatorID string, resIDs ...int) error {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	var ErrEntries []int
	logger.LogDebug(reqId, reqUrl, "executing query to delete restaurant by admin")
	stmt, err := db.Prepare(DeleteRestaurantsByAdmin)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in preparing statement: %v", err), 0)
		return database.ErrInternal
	}
	for i, id := range resIDs {
		result, err := stmt.Exec(id, creatorID)
		if err != nil {
			logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
			return database.ErrInternal
		}
		numDeletedRows, _ := result.RowsAffected()
		if numDeletedRows == 0 {
			ErrEntries = append(ErrEntries, i)
		}
	}
	length := len(ErrEntries)
	if length != 0 {
		return sendErrorMessage(ctx, ErrEntries, length, "Restaurants")
	}
	logger.LogInfo(reqId, reqUrl, "restaurant deleted by admin from db successfully", 0)
	return nil
}

func (db *MySqlDB) StoreToken(ctx context.Context, token string) error {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogDebug(reqId, reqUrl, "executing query to store logged out token")
	_, err := db.Query("insert into invalid_tokens(token) values(?)", token)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return database.ErrInternal
	}
	logger.LogInfo(reqId, reqUrl, "logged out token stored in db  successfully", 0)
	return nil
}
func (db *MySqlDB) VerifyToken(ctx context.Context, tokenIn string) bool {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	var count int
	logger.LogDebug(reqId, reqUrl, "executing query to verify logged out token")
	rows, err := db.Query("select count(*) from invalid_tokens where token=?", tokenIn)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return false
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&count)
	defer rows.Close()
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return false
	}
	if count==0{
		logger.LogInfo(reqId, reqUrl, "valid login token", 0)
		return true
	}else{
		logger.LogInfo(reqId, reqUrl, "invalid login token", 0)
		return false
	}
}

func CheckOwnerID(ctx context.Context, db *MySqlDB, ownerID string) bool {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	logger.LogDebug(reqId, reqUrl, "executing query to check that owner id exist")
	var count int
	rows, err := db.Query("select count(*) from owners where id=?", ownerID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return false
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&count)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in storing result in variable: %v", err), 0)
		return false
	}
	if count != 1 {
		return false
	}
	logger.LogInfo(reqId, reqUrl, "owner id exist in db ", 0)
	return true

}
func CheckRestaurantID(ctx context.Context, db *MySqlDB, resID int) bool {
	reqIdVal := ctx.Value("reqId")
	reqId := reqIdVal.(string)
	reqUrlVal := ctx.Value("reqUrl")
	reqUrl := reqUrlVal.(string)
	var count int
	logger.LogDebug(reqId, reqUrl, "executing query to check that restaurant id exist")

	rows, err := db.Query("select count(*) from restaurants where id=?", resID)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in executing query: %v", err), 0)
		return false
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&count)
	if err != nil {
		logger.LogError(reqId, reqUrl, fmt.Sprintf("error in storing result in variable: %v", err), 0)
		return false
	}
	if count != 1 {
		return false
	}
	logger.LogInfo(reqId, reqUrl, "restaurant id exist in db ", 0)
	return true
}

////////////////////
//Database migration

func migrateDatabase(db *sql.DB) error {
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return err
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	migration, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s/database", dir),
		"restaurant",
		driver,
	)
	if err != nil {
		return err
	}

	migration.Log = &models.MigrationLogger{}

	migration.Log.Printf("Applying database migrations")
	err = migration.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	version, _, err := migration.Version()
	if err != nil {
		return err
	}

	migration.Log.Printf("Active database version: %d", version)

	return nil

}

package database

import (
	"context"
	"errors"
	"github.com/vds/go-resman/pkg/models"
)

var (
	ErrInternal                 = errors.New("internal server error")
	ErrDupEmail                 = errors.New("email already used try a different one")
	ErrInvalidCredentials       = errors.New("incorrect login details")
	ErrInvalidOwner             = errors.New("owner does not exist")
	ErrInvalidOwnerCreator      = errors.New("can not update owner created by other admin")
	ErrInvalidRestaurantCreator = errors.New("can not update restaurant created by other admin")
	ErrNonExistingRestaurant    = errors.New("restaurant does not exist")
	ErrInvalidRestaurantOwner   = errors.New("can not update restaurant owned by others")
	ErrInvalidDish              = errors.New("dish does not exist")
	ErrInvalidRestaurantDish    = errors.New("can not update dish of other restaurant")
)

type Database interface {
	ShowNearBy(ctx context.Context, location *models.Location) (string, error)

	CreateUser(ctx context.Context, user *models.UserReg) (string, error)
	LogInUser(ctx context.Context, cred *models.Credentials) (string, error)
	ShowAdmins(ctx context.Context, ) (string, error)
	UpdateAdmin(ctx context.Context, admin *models.UserOutput) (string, error)
	RemoveAdmins(ctx context.Context, adminIDs ...string) error

	ShowOwners(ctx context.Context, userAuth *models.UserAuth) (string, error)
	CreateOwner(ctx context.Context, creatorID string, owner *models.OwnerReg) (*models.UserOutput, error)

	CheckOwnerCreator(ctx context.Context, creatorID string, ownerID string) error
	UpdateOwner(ctx context.Context, owner *models.UserOutput) (string, error)
	RemoveOwners(ctx context.Context, userAuth *models.UserAuth, ownerIDs ...string) error

	CheckAdmin(ctx context.Context, adminID string) error
	ShowRestaurants(ctx context.Context, userAuth *models.UserAuth) (string, error)
	InsertRestaurant(ctx context.Context, restaurant *models.Restaurant) (*models.RestaurantOutput, error)
	ShowAvailableRestaurants(ctx context.Context, userAuth *models.UserAuth) (string, error)
	InsertOwnerForRestaurants(ctx context.Context, userAuth *models.UserAuth, ownerID string, resIDs ...int) error
	RemoveOwnerForRestaurants(ctx context.Context, userAuth *models.UserAuth, ownerID string, resIDs ...int) error

	CheckRestaurantCreator(ctx context.Context, creatorID string, resID int) error
	UpdateRestaurant(ctx context.Context, restaurant *models.RestaurantOutput) (*models.RestaurantOutput, error)

	RemoveRestaurants(ctx context.Context, userAuth *models.UserAuth, resIDs ...int) error

	ShowMenu(ctx context.Context, resID int) (string, error)
	CheckRestaurantOwner(ctx context.Context, ownerID string, resID int) error
	InsertDishes(ctx context.Context, dishes models.Dish, resID int) (*models.DishOutput, error)
	UpdateDish(ctx context.Context, dish *models.DishOutput) (*models.DishOutput, error)
	CheckRestaurantDish(ctx context.Context, resID int, dishID int) error

	RemoveDishes(ctx context.Context, dishIDs ...int) error

	StoreToken(ctx context.Context, token string) error
	VerifyToken(ctx context.Context, token string) bool
}

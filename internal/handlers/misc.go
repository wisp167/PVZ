package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/wisp167/pvz/api"
	"github.com/wisp167/pvz/internal/db"
)

func TransformAddProductRowToProduct(row db.AddProductRow) api.Product {
	var dateTimePtr *time.Time
	if row.DateTime.Valid {
		dateTimePtr = &row.DateTime.Time
	}

	var idPtr *types.UUID
	if row.ID != uuid.Nil {
		id := types.UUID(row.ID)
		idPtr = &id
	}

	return api.Product{
		DateTime:    dateTimePtr,
		Id:          idPtr,
		ReceptionId: types.UUID(row.ReceptionID),
		Type:        api.ProductType(row.Type),
	}
}

func ConvertReceptionRowToAPI(row db.CreateOrGetReceptionRow) api.Reception {
	reception := api.Reception{
		PvzId:  openapi_types.UUID(row.PvzID),
		Status: api.ReceptionStatus(row.Status),
	}

	// Handle ID (convert to pointer)
	if row.ID != uuid.Nil {
		id := openapi_types.UUID(row.ID)
		reception.Id = &id
	}

	// Handle DateTime (convert sql.NullTime to time.Time)
	if row.DateTime.Valid {
		reception.DateTime = row.DateTime.Time
	} else {
		// Set default/zero time if NULL
		reception.DateTime = time.Time{}
	}

	return reception
}

func ConvertCloseReceptionRowToAPI(row db.CloseReceptionRow) api.Reception {
	fmt.Printf("DEBUG: Received row: %+v\n", row)
	reception := api.Reception{
		PvzId:  openapi_types.UUID(row.PvzID),
		Status: api.ReceptionStatus(row.Status),
	}

	// Handle ID (convert to pointer)
	if row.ID != uuid.Nil {
		id := openapi_types.UUID(row.ID)
		reception.Id = &id
	}

	// Handle DateTime (convert sql.NullTime to time.Time)
	if row.DateTime.Valid {
		reception.DateTime = row.DateTime.Time
	} else {
		// Set default/zero time if NULL
		reception.DateTime = time.Time{}
	}

	return reception
}

func ConvertCreatePVZRowToPVZ(row db.CreatePVZRow) api.PVZ {
	pvz := api.PVZ{
		City: api.PVZCity(row.City),
	}

	uuidVal := types.UUID(row.ID)
	pvz.Id = &uuidVal

	if row.RegistrationDate.Valid {
		pvz.RegistrationDate = &row.RegistrationDate.Time
	}

	return pvz
}
func ToUser(row *db.CreateUserRow) (*api.User, error) {
	// Convert UUID
	var userID *openapi_types.UUID
	if row.ID != uuid.Nil {
		parsedUUID := openapi_types.UUID(row.ID)
		userID = &parsedUUID
	}

	// Convert email
	email := openapi_types.Email(row.Email)

	// Convert role
	var userRole api.UserRole
	switch strings.ToLower(row.Role) {
	case "employee":
		userRole = api.UserRoleEmployee
	case "moderator":
		userRole = api.UserRoleModerator
	default:
		return nil, fmt.Errorf("invalid role: %s", row.Role)
	}

	return &api.User{
		Id:    userID,
		Email: email,
		Role:  userRole,
	}, nil
}

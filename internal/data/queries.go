package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/wisp167/pvz/api"
	"github.com/wisp167/pvz/internal/db"
	"github.com/wisp167/pvz/internal/helpers"
)

type PVZModelv1 struct {
	DB *sql.DB
}

type PVZModel struct {
	DB      *sql.DB
	Queries *db.Queries
}

func (m *Models) Login(reqCtx context.Context, req api.PostLoginJSONBody) (string, error) {

	var res string

	err := m.Transaction(reqCtx, func(q *db.Queries) error {
		user := db.GetUserByCredentialsParams{Email: string(req.Email), Md5: helpers.Md5(req.Password)}
		var err error
		res, err = q.GetUserByCredentials(reqCtx, user)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return "", err
	}
	return api.Token(res), nil
}

func (m *Models) Register(reqCtx context.Context, req api.PostRegisterJSONBody) (db.CreateUserRow, error) {

	var resp db.CreateUserRow

	err := m.Transaction(reqCtx, func(q *db.Queries) error {
		user := db.CreateUserParams{Email: string(req.Email), Md5: helpers.Md5(req.Password), Role: string(req.Role)}
		var err error
		_, err = q.GetUserByCredentials(reqCtx, db.GetUserByCredentialsParams{Email: string(req.Email), Md5: helpers.Md5(req.Password)})
		if !errors.Is(err, sql.ErrNoRows) {
			return sql.ErrNoRows
		}
		resp, err = q.CreateUser(reqCtx, user)
		return err
	})
	if err != nil {
		return db.CreateUserRow{}, err
	}
	return resp, nil
}

func (m *Models) AddPVZ(reqCtx context.Context, req api.PVZ) (db.CreatePVZRow, error) {

	var pvz db.CreatePVZRow

	err := m.Transaction(reqCtx, func(q *db.Queries) error {
		var err error
		pvz, err = q.CreatePVZ(reqCtx, string(req.City))
		return err
	})
	if err != nil {
		return db.CreatePVZRow{}, err
	}
	return pvz, nil
}

func (m *Models) AddReception(reqCtx context.Context, req api.PostReceptionsJSONBody) (db.CreateOrGetReceptionRow, error) {

	var reception db.CreateOrGetReceptionRow

	err := m.Transaction(reqCtx, func(q *db.Queries) error {
		var err error
		reception, err = q.CreateOrGetReception(reqCtx, uuid.UUID(req.PvzId))
		return err
	})
	if err != nil {
		return db.CreateOrGetReceptionRow{}, err
	}
	return reception, nil
}

func (m *Models) AddProduct(reqCtx context.Context, req api.PostProductsJSONBody) (db.AddProductRow, error) {

	var product db.AddProductRow

	err := m.Transaction(reqCtx, func(q *db.Queries) error {
		var err error
		product, err = q.AddProduct(reqCtx, db.AddProductParams{PvzID: uuid.UUID(req.PvzId), Type: string(req.Type)})
		return err
	})
	if err != nil {
		return db.AddProductRow{}, err
	}
	return product, nil
}

func (m *Models) CloseLastReception(reqCtx context.Context, req openapi_types.UUID) (db.CloseReceptionRow, error) {

	var reception db.CloseReceptionRow

	err := m.Transaction(reqCtx, func(q *db.Queries) error {
		var err error
		reception, err = q.CloseReception(reqCtx, uuid.UUID(req))
		return err
	})
	if err != nil {
		return db.CloseReceptionRow{}, err
	}
	fmt.Println(reception)
	return reception, nil
}

func (m *Models) DeleteLastProduct(reqCtx context.Context, req openapi_types.UUID) error {

	err := m.Transaction(reqCtx, func(q *db.Queries) error {
		var err error
		_, err = q.DeleteLastProduct(reqCtx, uuid.UUID(req))
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

// PVZWithReceptionsResponse matches the expected API response format
type PVZWithReceptionsResponse struct {
	PVZ        api.PVZ                 `json:"pvz"`
	Receptions []ReceptionWithProducts `json:"receptions"`
}

// ReceptionWithProducts matches the nested structure in the API spec
type ReceptionWithProducts struct {
	Reception api.Reception `json:"reception"`
	Products  []api.Product `json:"products"`
}

func (m *Models) GetPVZ(reqCtx context.Context, req api.GetPvzParams) ([]PVZWithReceptionsResponse, error) {
	var rows []db.GetPVZsWithReceptionsRow
	var result []PVZWithReceptionsResponse

	err := m.ReadOnlyTransaction(reqCtx, func(q *db.Queries) error {
		var err error
		params := ConvertPvzParamsToReceptionsParams(req)
		rows, err = q.GetPVZsWithReceptions(reqCtx, params)
		if err != nil {
			return err
		}

		// Convert each row to the API response format
		for _, row := range rows {
			var receptions []ReceptionWithProducts

			// Unmarshal the JSON receptions data
			if err := json.Unmarshal([]byte(row.ReceptionsJson), &receptions); err != nil {
				return fmt.Errorf("failed to unmarshal receptions JSON: %w", err)
			}

			// Convert UUID types
			pvzID := openapi_types.UUID(row.PvzID)

			response := PVZWithReceptionsResponse{
				PVZ: api.PVZ{
					Id:               &pvzID,
					RegistrationDate: &row.RegistrationDate.Time,
					City:             api.PVZCity(row.City),
				},
				Receptions: receptions,
			}
			result = append(result, response)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get PVZ data: %w", err)
	}

	return result, nil
}

func ConvertPvzParamsToReceptionsParams(src api.GetPvzParams) db.GetPVZsWithReceptionsParams {
	// Set defaults
	page := 1
	if src.Page != nil {
		page = *src.Page
	}

	limit := 10
	if src.Limit != nil {
		limit = *src.Limit
	}

	params := db.GetPVZsWithReceptionsParams{
		Limit:   int32(limit),
		Column4: int32(page),
	}

	// Handle date parameters
	if src.StartDate != nil {
		params.DateTime = sql.NullTime{
			Time:  *src.StartDate,
			Valid: true,
		}
	}

	if src.EndDate != nil {
		params.DateTime_2 = sql.NullTime{
			Time:  *src.EndDate,
			Valid: true,
		}
	}

	return params
}

/*
func (m *Models) GetPVZ(reqCtx context.Context, req api.GetPvzParams) ([]db.GetPVZsWithReceptionsRow, error) {

	var pvz []db.GetPVZsWithReceptionsRow

	err := m.ReadOnlyTransaction(reqCtx, func(q *db.Queries) error {
		var err error
		pvz, err = q.GetPVZsWithReceptions(reqCtx, ConvertPvzParamsToReceptionsParams(req))
		return err
	})
	if err != nil {
		return []db.GetPVZsWithReceptionsRow{}, err
	}
	return pvz, nil
}

func ConvertPvzParamsToReceptionsParams(src api.GetPvzParams) db.GetPVZsWithReceptionsParams {
	dest := db.GetPVZsWithReceptionsParams{
		Column4: src.Page, // Assuming Column4 corresponds to the ID
	}

	// Convert StartDate to DateTime (sql.NullTime)
	if src.StartDate != nil {
		dest.DateTime = sql.NullTime{
			Time:  *src.StartDate,
			Valid: true,
		}
	} else {
		dest.DateTime = sql.NullTime{Valid: false}
	}

	// Convert EndDate to DateTime_2 (sql.NullTime)
	if src.EndDate != nil {
		dest.DateTime_2 = sql.NullTime{
			Time:  *src.EndDate,
			Valid: true,
		}
	} else {
		dest.DateTime_2 = sql.NullTime{Valid: false}
	}

	// Convert Limit from *int to int32 (with default value if nil)
	if src.Limit != nil {
		dest.Limit = int32(*src.Limit)
	} else {
		dest.Limit = 10 // or whatever default value you prefer, based on your API default
	}

	return dest
}
*/

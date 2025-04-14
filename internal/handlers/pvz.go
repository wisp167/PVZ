package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/wisp167/pvz/api"
	"github.com/wisp167/pvz/internal/data"
	"github.com/wisp167/pvz/internal/helpers"
)

type ServerHandler struct {
	Model  *data.Models
	jwtkey []byte
	logger *log.Logger
}

func (h *ServerHandler) InitUnexportedVals(jwtkey []byte, logger *log.Logger) {
	h.jwtkey = jwtkey
	h.logger = logger
}

// Получение тестового токена
// (POST /dummyLogin)
func (h *ServerHandler) PostDummyLogin(ctx echo.Context) error {
	var req api.PostDummyLoginJSONBody
	if err := helpers.ReadJSON(ctx, &req); err != nil {
		h.logError(err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	switch req.Role {
	case api.PostDummyLoginJSONBodyRoleEmployee, api.PostDummyLoginJSONBodyRoleModerator:
		token, err := DummyLogin(string(req.Role), h.jwtkey)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate token")
		}
		return ctx.JSON(http.StatusOK, token)
	}
	/*
		token, err := DummyLogin(string(req.Role), h.jwtkey)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate token")
		}
	*/
	return echo.NewHTTPError(http.StatusBadRequest, "Invalid role")
}

// Авторизация пользователя
// (POST /login)
func (h *ServerHandler) PostLogin(ctx echo.Context) error {
	var req api.PostLoginJSONBody
	if err := helpers.ReadJSON(ctx, &req); err != nil {
		h.logError(err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}
	reqCtx := ctx.Request().Context()

	role, err := h.Model.Login(reqCtx, req)

	if err != nil {
		if err == data.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
		}
		h.logError(err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Failed to authenticate")
	}

	// Generate JWT token
	token, err := DummyLogin(role, h.jwtkey)
	if err != nil {
		h.logError(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate token")
	}

	return ctx.JSON(http.StatusOK, token)
}

// Добавление товара в текущую приемку (только для сотрудников ПВЗ)
// (POST /products)
func (h *ServerHandler) PostProducts(ctx echo.Context) error {
	//app.queue <- struct{}{}
	//defer func() {
	//	<-app.queue
	//}()

	var req api.PostProductsJSONBody
	if err := helpers.ReadJSON(ctx, &req); err != nil {
		h.logError(err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	reqCtx := ctx.Request().Context()

	product, err := h.Model.AddProduct(reqCtx, req)
	if errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusBadRequest, "product already exists")
	}
	if err != nil {
		h.logError(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register product")
	}
	return ctx.JSON(http.StatusCreated, TransformAddProductRowToProduct(product))
}

// Получение списка ПВЗ с фильтрацией по дате приемки и пагинацией
// (GET /pvz)
func (h *ServerHandler) GetPvz(ctx echo.Context, params api.GetPvzParams) error {

	reqCtx := ctx.Request().Context()

	pvz, err := h.Model.GetPVZ(reqCtx, params)
	if err != nil {
		h.logError(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user")
	}
	return ctx.JSON(http.StatusOK, pvz)

}

// Создание ПВЗ (только для модераторов)
// (POST /pvz)
func (h *ServerHandler) PostPvz(ctx echo.Context) error {
	var req api.PVZ
	if err := helpers.ReadJSON(ctx, &req); err != nil {
		h.logError(err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	reqCtx := ctx.Request().Context()

	pvz, err := h.Model.AddPVZ(reqCtx, req)
	if errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusBadRequest, "user already exists")
	}
	if err != nil {
		h.logError(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user")
	}
	return ctx.JSON(http.StatusCreated, ConvertCreatePVZRowToPVZ(pvz))
}

// Закрытие последней открытой приемки товаров в рамках ПВЗ
// (POST /pvz/{pvzId}/close_last_reception)
func (h *ServerHandler) PostPvzPvzIdCloseLastReception(ctx echo.Context, pvzId openapi_types.UUID) error {

	reqCtx := ctx.Request().Context()

	recep, err := h.Model.CloseLastReception(reqCtx, pvzId)
	if err != nil {
		h.logError(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user")
	}
	return ctx.JSON(http.StatusOK, ConvertCloseReceptionRowToAPI(recep))
}

// Удаление последнего добавленного товара из текущей приемки (LIFO, только для сотрудников ПВЗ)
// (POST /pvz/{pvzId}/delete_last_product)
func (h *ServerHandler) PostPvzPvzIdDeleteLastProduct(ctx echo.Context, pvzId openapi_types.UUID) error {

	reqCtx := ctx.Request().Context()

	err := h.Model.DeleteLastProduct(reqCtx, pvzId)
	if err != nil {
		h.logError(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user")
	}
	return ctx.JSON(http.StatusOK, nil)
}

// Создание новой приемки товаров (только для сотрудников ПВЗ)
// (POST /receptions)
func (h *ServerHandler) PostReceptions(ctx echo.Context) error {
	var req api.PostReceptionsJSONBody
	if err := helpers.ReadJSON(ctx, &req); err != nil {
		h.logError(err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	reqCtx := ctx.Request().Context()

	recep, err := h.Model.AddReception(reqCtx, req)
	if err != nil {
		h.logError(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user")
	}
	return ctx.JSON(http.StatusOK, ConvertReceptionRowToAPI(recep))
}

// Регистрация пользователя
// (POST /register)
func (h *ServerHandler) PostRegister(ctx echo.Context) error {
	//app.queue <- struct{}{}
	//defer func() {
	//	<-app.queue
	//}()

	var req api.PostRegisterJSONBody
	if err := helpers.ReadJSON(ctx, &req); err != nil {
		h.logError(err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	reqCtx := ctx.Request().Context()

	user, err := h.Model.Register(reqCtx, req)
	if errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusBadRequest, "user already exists")
	}
	if err != nil {
		h.logError(err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to register user")
	}

	resp, err := ToUser(&user)

	return ctx.JSON(http.StatusCreated, resp)
}

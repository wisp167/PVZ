package handlers

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/wisp167/pvz/api"
	"github.com/wisp167/pvz/internal/data"
	"github.com/wisp167/pvz/internal/helpers"
)

type ServerHandler struct {
	Model  data.Models
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

	token, err := DummyLogin(string(req.Role), h.jwtkey)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate token")
	}

	return ctx.JSON(http.StatusOK, token)
}

// Авторизация пользователя
// (POST /login)
func (*ServerHandler) PostLogin(ctx echo.Context) error { return nil }

// Добавление товара в текущую приемку (только для сотрудников ПВЗ)
// (POST /products)
func (*ServerHandler) PostProducts(ctx echo.Context) error { return nil }

// Получение списка ПВЗ с фильтрацией по дате приемки и пагинацией
// (GET /pvz)
func (*ServerHandler) GetPvz(ctx echo.Context, params api.GetPvzParams) error { return nil }

// Создание ПВЗ (только для модераторов)
// (POST /pvz)
func (*ServerHandler) PostPvz(ctx echo.Context) error { return nil }

// Закрытие последней открытой приемки товаров в рамках ПВЗ
// (POST /pvz/{pvzId}/close_last_reception)
func (*ServerHandler) PostPvzPvzIdCloseLastReception(ctx echo.Context, pvzId openapi_types.UUID) error {
	return nil
}

// Удаление последнего добавленного товара из текущей приемки (LIFO, только для сотрудников ПВЗ)
// (POST /pvz/{pvzId}/delete_last_product)
func (*ServerHandler) PostPvzPvzIdDeleteLastProduct(ctx echo.Context, pvzId openapi_types.UUID) error {
	return nil
}

// Создание новой приемки товаров (только для сотрудников ПВЗ)
// (POST /receptions)
func (*ServerHandler) PostReceptions(ctx echo.Context) error { return nil }

// Регистрация пользователя
// (POST /register)
//func (*ServerHandler) PostRegister(ctx echo.Context) error { return nil }

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

	token, err := DummyLogin(string(req.Role), h.jwtkey)
	if err != nil {
		h.logError(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate token")
	}

	// Return token response
	return ctx.JSON(http.StatusOK, token)
}

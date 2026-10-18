package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"ostos/backend/internal/apierrors"
	"ostos/backend/internal/platform"
	"ostos/backend/internal/shopping"
	httpmiddleware "ostos/backend/internal/transport/http/middleware"
	"ostos/backend/internal/transport/http/request"
	"ostos/backend/internal/transport/http/validation"
)

type ShoppingService interface {
	GetList(context.Context, platform.ID, platform.ID) ([]shopping.ListItem, error)
	CreateItem(context.Context, platform.ID, platform.ID, shopping.CreateItemInput) (shopping.ListItem, error)
	UpdateItem(context.Context, platform.ID, platform.ID, platform.ID, shopping.UpdateItemInput) (shopping.ListItem, error)
	DeleteItem(context.Context, platform.ID, platform.ID, platform.ID) error
	ReorderItems(context.Context, platform.ID, platform.ID, []platform.ID) ([]shopping.ListItem, error)
	CompleteList(context.Context, platform.ID, platform.ID, shopping.CompleteListInput) (*shopping.TripDetail, error)
	ListTrips(context.Context, platform.ID, platform.ID, string, int) ([]shopping.Trip, string, error)
	GetTrip(context.Context, platform.ID, platform.ID, platform.ID) (shopping.TripDetail, error)
}

type Shopping struct {
	service   ShoppingService
	validator *validation.Validator
}

func NewShopping(service ShoppingService, validator *validation.Validator) *Shopping {
	return &Shopping{service: service, validator: validator}
}

func (handler *Shopping) List(context *echo.Context) error {
	familyID, err := routeID(context, "family_id")
	if err != nil {
		return err
	}
	identity, _ := httpmiddleware.Identity(context)
	items, serviceErr := handler.service.GetList(context.Request().Context(), familyID, identity.UserID)
	if serviceErr != nil {
		return domainError(serviceErr)
	}
	return context.JSON(http.StatusOK, map[string]any{"items": items})
}

func (handler *Shopping) CreateItem(context *echo.Context) error {
	familyID, err := routeID(context, "family_id")
	if err != nil {
		return err
	}
	var body request.ListItemCreateRequest
	if err := request.DecodeJSON(context.Request(), &body, handler.validator); err != nil {
		return err
	}
	quantity, err := platform.ParseQuantity(body.Quantity)
	if err != nil {
		return apierrors.InvalidFormBody(map[string]string{"quantity": "Enter a valid positive decimal quantity"})
	}
	identity, _ := httpmiddleware.Identity(context)
	item, serviceErr := handler.service.CreateItem(context.Request().Context(), familyID, identity.UserID, shopping.CreateItemInput{
		Name:     body.Name,
		Quantity: quantity,
		Unit:     platform.Unit(body.Unit),
		Note:     body.Note,
	})
	if serviceErr != nil {
		return domainError(serviceErr)
	}
	context.Response().Header().Set("Location", "/v1/families/"+familyID.String()+"/shopping-list/items/"+item.ID.String())
	return context.JSON(http.StatusCreated, item)
}

func (handler *Shopping) UpdateItem(context *echo.Context) error {
	familyID, err := routeID(context, "family_id")
	if err != nil {
		return err
	}
	itemID, err := routeID(context, "item_id")
	if err != nil {
		return err
	}
	var body request.ListItemUpdateRequest
	if err := request.DecodeJSON(context.Request(), &body, handler.validator); err != nil {
		return err
	}
	input := shopping.UpdateItemInput{
		Name:              body.Name,
		Note:              body.Note,
		ClearNote:         body.ClearNote,
		IsChecked:         body.IsChecked,
		ClearPurchasedQty: body.ClearPurchasedQty,
		ClearTotalCents:   body.ClearTotalCents,
	}
	if body.Quantity != nil {
		q, parseErr := platform.ParseQuantity(*body.Quantity)
		if parseErr != nil {
			return apierrors.InvalidFormBody(map[string]string{"quantity": "Enter a valid positive decimal quantity"})
		}
		input.Quantity = &q
	}
	if body.Unit != nil {
		u := platform.Unit(*body.Unit)
		input.Unit = &u
	}
	if body.PurchasedQuantity != nil {
		pq, parseErr := platform.ParseQuantity(*body.PurchasedQuantity)
		if parseErr != nil {
			return apierrors.InvalidFormBody(map[string]string{"purchased_quantity": "Enter a valid positive decimal quantity"})
		}
		input.PurchasedQuantity = &pq
	}
	if body.TotalCents != nil {
		cents, parseErr := platform.NewCents(*body.TotalCents)
		if parseErr != nil {
			return apierrors.InvalidFormBody(map[string]string{"total_cents": "Cents must be nonnegative"})
		}
		input.TotalCents = &cents
	}
	identity, _ := httpmiddleware.Identity(context)
	item, serviceErr := handler.service.UpdateItem(context.Request().Context(), familyID, itemID, identity.UserID, input)
	if serviceErr != nil {
		return domainError(serviceErr)
	}
	return context.JSON(http.StatusOK, item)
}

func (handler *Shopping) DeleteItem(context *echo.Context) error {
	familyID, err := routeID(context, "family_id")
	if err != nil {
		return err
	}
	itemID, err := routeID(context, "item_id")
	if err != nil {
		return err
	}
	identity, _ := httpmiddleware.Identity(context)
	if err := handler.service.DeleteItem(context.Request().Context(), familyID, itemID, identity.UserID); err != nil {
		return domainError(err)
	}
	return context.NoContent(http.StatusNoContent)
}

func (handler *Shopping) ReorderItems(context *echo.Context) error {
	familyID, err := routeID(context, "family_id")
	if err != nil {
		return err
	}
	var body request.ListItemReorderRequest
	if err := request.DecodeJSON(context.Request(), &body, handler.validator); err != nil {
		return err
	}
	itemIDs := make([]platform.ID, 0, len(body.ItemIDs))
	for _, rawID := range body.ItemIDs {
		id, parseErr := parseID(rawID)
		if parseErr != nil {
			return apierrors.InvalidFormBody(map[string]string{"item_ids": "Invalid item ID"})
		}
		itemIDs = append(itemIDs, id)
	}
	identity, _ := httpmiddleware.Identity(context)
	items, serviceErr := handler.service.ReorderItems(context.Request().Context(), familyID, identity.UserID, itemIDs)
	if serviceErr != nil {
		return domainError(serviceErr)
	}
	return context.JSON(http.StatusOK, map[string]any{"items": items})
}

func (handler *Shopping) Complete(context *echo.Context) error {
	familyID, err := routeID(context, "family_id")
	if err != nil {
		return err
	}
	var body request.ShoppingListCompleteRequest
	if err := request.DecodeJSON(context.Request(), &body, handler.validator); err != nil {
		return err
	}
	overrides := make([]shopping.ItemActionOverride, 0, len(body.UncheckedItemOverrides))
	for _, o := range body.UncheckedItemOverrides {
		id, parseErr := parseID(o.ItemID)
		if parseErr != nil {
			return apierrors.InvalidFormBody(map[string]string{"unchecked_item_overrides": "Invalid item ID"})
		}
		overrides = append(overrides, shopping.ItemActionOverride{
			ItemID: id,
			Action: shopping.UncheckedAction(o.Action),
		})
	}
	identity, _ := httpmiddleware.Identity(context)
	trip, serviceErr := handler.service.CompleteList(context.Request().Context(), familyID, identity.UserID, shopping.CompleteListInput{
		SupermarketName:        body.SupermarketName,
		DefaultUncheckedAction: shopping.UncheckedAction(body.DefaultUncheckedAction),
		UncheckedItemOverrides: overrides,
	})
	if serviceErr != nil {
		return domainError(serviceErr)
	}
	return context.JSON(http.StatusOK, map[string]any{"trip": trip})
}

func (handler *Shopping) ListTrips(context *echo.Context) error {
	familyID, err := routeID(context, "family_id")
	if err != nil {
		return err
	}
	cursor := context.QueryParam("cursor")
	limit := 20
	if rawLimit := context.QueryParam("limit"); rawLimit != "" {
		if parsed, parseErr := strconv.Atoi(rawLimit); parseErr == nil && parsed > 0 {
			limit = parsed
		}
	}
	identity, _ := httpmiddleware.Identity(context)
	trips, nextCursor, serviceErr := handler.service.ListTrips(context.Request().Context(), familyID, identity.UserID, cursor, limit)
	if serviceErr != nil {
		return domainError(serviceErr)
	}
	res := map[string]any{"trips": trips}
	if nextCursor != "" {
		res["next_cursor"] = nextCursor
	}
	return context.JSON(http.StatusOK, res)
}

func (handler *Shopping) GetTrip(context *echo.Context) error {
	familyID, err := routeID(context, "family_id")
	if err != nil {
		return err
	}
	tripID, err := routeID(context, "trip_id")
	if err != nil {
		return err
	}
	identity, _ := httpmiddleware.Identity(context)
	trip, serviceErr := handler.service.GetTrip(context.Request().Context(), familyID, tripID, identity.UserID)
	if serviceErr != nil {
		return domainError(serviceErr)
	}
	return context.JSON(http.StatusOK, trip)
}

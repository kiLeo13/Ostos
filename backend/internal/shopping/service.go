package shopping

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"

	"ostos/backend/internal/platform"
)

type Service struct {
	repository Repository
	ids        platform.IDGenerator
	clock      platform.Clock
}

func NewService(repository Repository, ids platform.IDGenerator, clock platform.Clock) (*Service, error) {
	if repository == nil || ids == nil || clock == nil {
		return nil, errors.New("all shopping service dependencies are required")
	}
	return &Service{repository: repository, ids: ids, clock: clock}, nil
}

func (service *Service) GetList(ctx context.Context, familyID, userID platform.ID) ([]ListItem, error) {
	if err := service.repository.CheckMembership(ctx, familyID, userID); err != nil {
		return nil, err
	}
	items, err := service.repository.GetList(ctx, familyID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []ListItem{}
	}
	return items, nil
}

func (service *Service) CreateItem(ctx context.Context, familyID, userID platform.ID, input CreateItemInput) (ListItem, error) {
	if err := service.repository.CheckMembership(ctx, familyID, userID); err != nil {
		return ListItem{}, err
	}
	name := strings.TrimSpace(input.Name)
	if len(name) < 1 || len(name) > 160 {
		return ListItem{}, errors.New("item name must be between 1 and 160 characters")
	}
	if input.Quantity.Milli() <= 0 {
		return ListItem{}, ErrInvalidQuantity
	}
	if !input.Unit.Valid() {
		return ListItem{}, ErrInvalidUnit
	}
	var note *string
	if input.Note != nil {
		trimmed := strings.TrimSpace(*input.Note)
		if len(trimmed) > 500 {
			return ListItem{}, errors.New("note must not exceed 500 characters")
		}
		if trimmed != "" {
			note = &trimmed
		}
	}

	id, err := service.ids.Next(ctx)
	if err != nil {
		return ListItem{}, err
	}
	now := service.clock.Now().UTC()

	var created ListItem
	err = service.repository.WithinTransaction(ctx, func(repository Repository) error {
		maxPos, posErr := repository.GetMaxPosition(ctx, familyID)
		if posErr != nil {
			return posErr
		}
		item := ListItem{
			ID:          id,
			FamilyID:    familyID,
			Name:        name,
			Quantity:    input.Quantity,
			Unit:        input.Unit,
			Note:        note,
			IsChecked:   false,
			Position:    maxPos + 1000,
			AddedByID:   userID,
			UpdatedByID: userID,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if createErr := repository.CreateListItem(ctx, &item); createErr != nil {
			return createErr
		}
		created = item
		return nil
	})
	if err != nil {
		return ListItem{}, err
	}
	return created, nil
}

func (service *Service) UpdateItem(ctx context.Context, familyID, itemID, userID platform.ID, input UpdateItemInput) (ListItem, error) {
	if err := service.repository.CheckMembership(ctx, familyID, userID); err != nil {
		return ListItem{}, err
	}

	var updated ListItem
	err := service.repository.WithinTransaction(ctx, func(repository Repository) error {
		item, err := repository.GetListItem(ctx, familyID, itemID)
		if err != nil {
			return err
		}
		if item == nil {
			return ErrListItemNotFound
		}

		if input.Name != nil {
			name := strings.TrimSpace(*input.Name)
			if len(name) < 1 || len(name) > 160 {
				return errors.New("item name must be between 1 and 160 characters")
			}
			item.Name = name
		}
		if input.Quantity != nil {
			if input.Quantity.Milli() <= 0 {
				return ErrInvalidQuantity
			}
			item.Quantity = *input.Quantity
		}
		if input.Unit != nil {
			if !input.Unit.Valid() {
				return ErrInvalidUnit
			}
			item.Unit = *input.Unit
		}
		if input.ClearNote {
			item.Note = nil
		} else if input.Note != nil {
			trimmed := strings.TrimSpace(*input.Note)
			if len(trimmed) > 500 {
				return errors.New("note must not exceed 500 characters")
			}
			if trimmed != "" {
				item.Note = &trimmed
			} else {
				item.Note = nil
			}
		}

		if input.IsChecked != nil {
			if *input.IsChecked {
				item.IsChecked = true
				if input.ClearPurchasedQty {
					item.PurchasedQuantity = nil
				} else if input.PurchasedQuantity != nil {
					if input.PurchasedQuantity.Milli() <= 0 {
						return ErrInvalidQuantity
					}
					item.PurchasedQuantity = input.PurchasedQuantity
				}
				if input.ClearTotalCents {
					item.TotalCents = nil
				} else if input.TotalCents != nil {
					if int64(*input.TotalCents) < 0 {
						return errors.New("cents must be nonnegative")
					}
					item.TotalCents = input.TotalCents
				}
			} else {
				// Unchecking clears temporary purchase data per PLAN.md
				item.IsChecked = false
				item.PurchasedQuantity = nil
				item.TotalCents = nil
			}
		} else if item.IsChecked {
			if input.ClearPurchasedQty {
				item.PurchasedQuantity = nil
			} else if input.PurchasedQuantity != nil {
				if input.PurchasedQuantity.Milli() <= 0 {
					return ErrInvalidQuantity
				}
				item.PurchasedQuantity = input.PurchasedQuantity
			}
			if input.ClearTotalCents {
				item.TotalCents = nil
			} else if input.TotalCents != nil {
				if int64(*input.TotalCents) < 0 {
					return errors.New("cents must be nonnegative")
				}
				item.TotalCents = input.TotalCents
			}
		}

		item.UpdatedByID = userID
		item.UpdatedAt = service.clock.Now().UTC()
		if err := repository.UpdateListItem(ctx, item); err != nil {
			return err
		}
		updated = *item
		return nil
	})
	if err != nil {
		return ListItem{}, err
	}
	return updated, nil
}

func (service *Service) DeleteItem(ctx context.Context, familyID, itemID, userID platform.ID) error {
	if err := service.repository.CheckMembership(ctx, familyID, userID); err != nil {
		return err
	}
	return service.repository.DeleteListItem(ctx, familyID, itemID)
}

func (service *Service) ReorderItems(ctx context.Context, familyID, userID platform.ID, itemIDs []platform.ID) ([]ListItem, error) {
	if err := service.repository.CheckMembership(ctx, familyID, userID); err != nil {
		return nil, err
	}
	if len(itemIDs) == 0 {
		return service.GetList(ctx, familyID, userID)
	}

	var items []ListItem
	err := service.repository.WithinTransaction(ctx, func(repository Repository) error {
		if err := repository.ReorderListItems(ctx, familyID, itemIDs); err != nil {
			return err
		}
		var listErr error
		items, listErr = repository.GetList(ctx, familyID)
		return listErr
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (service *Service) CompleteList(ctx context.Context, familyID, userID platform.ID, input CompleteListInput) (*TripDetail, error) {
	if err := service.repository.CheckMembership(ctx, familyID, userID); err != nil {
		return nil, err
	}
	if !input.DefaultUncheckedAction.Valid() {
		return nil, ErrInvalidAction
	}
	overrideMap := make(map[platform.ID]UncheckedAction, len(input.UncheckedItemOverrides))
	for _, override := range input.UncheckedItemOverrides {
		if !override.Action.Valid() {
			return nil, ErrInvalidAction
		}
		overrideMap[override.ItemID] = override.Action
	}

	var result *TripDetail
	err := service.repository.WithinTransaction(ctx, func(repository Repository) error {
		if err := repository.LockFamily(ctx, familyID); err != nil {
			return err
		}
		currentItems, err := repository.LockListItemsForUpdate(ctx, familyID)
		if err != nil {
			return err
		}

		now := service.clock.Now().UTC()
		var purchased []ListItem
		var keptIDs []platform.ID
		var discardedIDs []platform.ID

		for _, item := range currentItems {
			if item.IsChecked {
				purchased = append(purchased, item)
			} else {
				action := input.DefaultUncheckedAction
				if override, ok := overrideMap[item.ID]; ok {
					action = override
				}
				switch action {
				case ActionKeep:
					keptIDs = append(keptIDs, item.ID)
				case ActionDiscard:
					discardedIDs = append(discardedIDs, item.ID)
				}
			}
		}

		if len(purchased) > 0 {
			tripID, idErr := service.ids.Next(ctx)
			if idErr != nil {
				return idErr
			}

			var supermarket *string
			if input.SupermarketName != nil {
				trimmed := strings.TrimSpace(*input.SupermarketName)
				if len(trimmed) > 120 {
					return errors.New("supermarket name must not exceed 120 characters")
				}
				if trimmed != "" {
					supermarket = &trimmed
				}
			}

			trip := Trip{
				ID:              tripID,
				FamilyID:        familyID,
				SupermarketName: supermarket,
				CompletedByID:   userID,
				CompletedAt:     now,
				CreatedAt:       now,
			}

			tripItems := make([]TripItem, 0, len(purchased))
			purchasedIDs := make([]platform.ID, 0, len(purchased))
			for _, item := range purchased {
				tripItemID, nextIDErr := service.ids.Next(ctx)
				if nextIDErr != nil {
					return nextIDErr
				}
				tripItems = append(tripItems, TripItem{
					ID:                tripItemID,
					TripID:            tripID,
					SourceItemID:      item.ID,
					Name:              item.Name,
					Quantity:          item.Quantity,
					Unit:              item.Unit,
					Note:              item.Note,
					AddedByID:         item.AddedByID,
					PurchasedQuantity: item.PurchasedQuantity,
					TotalCents:        item.TotalCents,
					CreatedAt:         now,
				})
				purchasedIDs = append(purchasedIDs, item.ID)
			}

			if err := repository.CreateTrip(ctx, &trip, tripItems); err != nil {
				return err
			}
			if err := repository.DeleteListItems(ctx, familyID, purchasedIDs); err != nil {
				return err
			}
			result = &TripDetail{Trip: trip, Items: tripItems}
		}

		if len(discardedIDs) > 0 {
			if err := repository.DeleteListItems(ctx, familyID, discardedIDs); err != nil {
				return err
			}
		}

		if len(keptIDs) > 0 {
			if err := repository.ResetUncheckedListItems(ctx, familyID, keptIDs, now); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (service *Service) ListTrips(ctx context.Context, familyID, userID platform.ID, cursorStr string, limit int) ([]Trip, string, error) {
	if err := service.repository.CheckMembership(ctx, familyID, userID); err != nil {
		return nil, "", err
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	var cursor *platform.ID
	if cursorStr != "" {
		id, err := DecodeCursor(cursorStr)
		if err != nil {
			return nil, "", ErrInvalidCursor
		}
		cursor = &id
	}

	trips, nextCursorID, err := service.repository.ListTrips(ctx, familyID, cursor, limit)
	if err != nil {
		return nil, "", err
	}
	if trips == nil {
		trips = []Trip{}
	}

	var nextCursor string
	if nextCursorID != nil {
		nextCursor = EncodeCursor(*nextCursorID)
	}
	return trips, nextCursor, nil
}

func (service *Service) GetTrip(ctx context.Context, familyID, tripID, userID platform.ID) (TripDetail, error) {
	if err := service.repository.CheckMembership(ctx, familyID, userID); err != nil {
		return TripDetail{}, err
	}
	detail, err := service.repository.GetTrip(ctx, familyID, tripID)
	if err != nil {
		return TripDetail{}, err
	}
	if detail == nil {
		return TripDetail{}, ErrTripNotFound
	}
	return *detail, nil
}

func EncodeCursor(id platform.ID) string {
	return base64.RawURLEncoding.EncodeToString([]byte(id.String()))
}

func DecodeCursor(cursor string) (platform.ID, error) {
	data, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return 0, ErrInvalidCursor
	}
	id, err := platform.ParseID(string(data))
	if err != nil {
		return 0, ErrInvalidCursor
	}
	return id, nil
}

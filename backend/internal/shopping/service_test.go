package shopping

import (
	"context"
	"testing"
	"time"

	"ostos/backend/internal/platform"
)

type memoryClock struct{ now time.Time }

func (clock memoryClock) Now() time.Time { return clock.now }

type memoryIDGenerator struct {
	nextID int64
}

func (gen *memoryIDGenerator) Next(ctx context.Context) (platform.ID, error) {
	gen.nextID++
	return platform.ID(gen.nextID), nil
}

type memoryRepository struct {
	items []ListItem
	trips []Trip
	tripItems []TripItem
}

func (repo *memoryRepository) WithinTransaction(ctx context.Context, fn func(Repository) error) error {
	return fn(repo)
}

func (repo *memoryRepository) CheckMembership(ctx context.Context, familyID, userID platform.ID) error {
	return nil
}

func (repo *memoryRepository) LockFamily(ctx context.Context, familyID platform.ID) error {
	return nil
}

func (repo *memoryRepository) GetList(ctx context.Context, familyID platform.ID) ([]ListItem, error) {
	return repo.items, nil
}

func (repo *memoryRepository) GetListItem(ctx context.Context, familyID, itemID platform.ID) (*ListItem, error) {
	for i := range repo.items {
		if repo.items[i].FamilyID == familyID && repo.items[i].ID == itemID {
			itemCopy := repo.items[i]
			return &itemCopy, nil
		}
	}
	return nil, nil
}

func (repo *memoryRepository) GetMaxPosition(ctx context.Context, familyID platform.ID) (int64, error) {
	var max int64
	for _, it := range repo.items {
		if it.FamilyID == familyID && it.Position > max {
			max = it.Position
		}
	}
	return max, nil
}

func (repo *memoryRepository) CreateListItem(ctx context.Context, item *ListItem) error {
	repo.items = append(repo.items, *item)
	return nil
}

func (repo *memoryRepository) UpdateListItem(ctx context.Context, item *ListItem) error {
	for i := range repo.items {
		if repo.items[i].FamilyID == item.FamilyID && repo.items[i].ID == item.ID {
			repo.items[i] = *item
			return nil
		}
	}
	return ErrListItemNotFound
}

func (repo *memoryRepository) DeleteListItem(ctx context.Context, familyID, itemID platform.ID) error {
	for i := range repo.items {
		if repo.items[i].FamilyID == familyID && repo.items[i].ID == itemID {
			repo.items = append(repo.items[:i], repo.items[i+1:]...)
			return nil
		}
	}
	return ErrListItemNotFound
}

func (repo *memoryRepository) ReorderListItems(ctx context.Context, familyID platform.ID, itemIDs []platform.ID) error {
	for idx, id := range itemIDs {
		for i := range repo.items {
			if repo.items[i].FamilyID == familyID && repo.items[i].ID == id {
				repo.items[i].Position = int64((idx + 1) * 1000)
			}
		}
	}
	return nil
}

func (repo *memoryRepository) LockListItemsForUpdate(ctx context.Context, familyID platform.ID) ([]ListItem, error) {
	return repo.items, nil
}

func (repo *memoryRepository) CreateTrip(ctx context.Context, trip *Trip, items []TripItem) error {
	repo.trips = append(repo.trips, *trip)
	repo.tripItems = append(repo.tripItems, items...)
	return nil
}

func (repo *memoryRepository) DeleteListItems(ctx context.Context, familyID platform.ID, itemIDs []platform.ID) error {
	idMap := make(map[platform.ID]bool)
	for _, id := range itemIDs {
		idMap[id] = true
	}
	filtered := make([]ListItem, 0, len(repo.items))
	for _, it := range repo.items {
		if it.FamilyID == familyID && idMap[it.ID] {
			continue
		}
		filtered = append(filtered, it)
	}
	repo.items = filtered
	return nil
}

func (repo *memoryRepository) ResetUncheckedListItems(ctx context.Context, familyID platform.ID, itemIDs []platform.ID, updatedAt time.Time) error {
	idMap := make(map[platform.ID]bool)
	for _, id := range itemIDs {
		idMap[id] = true
	}
	for i := range repo.items {
		if repo.items[i].FamilyID == familyID && idMap[repo.items[i].ID] {
			repo.items[i].IsChecked = false
			repo.items[i].PurchasedQuantity = nil
			repo.items[i].TotalCents = nil
			repo.items[i].UpdatedAt = updatedAt
		}
	}
	return nil
}

func (repo *memoryRepository) ListTrips(ctx context.Context, familyID platform.ID, cursor *platform.ID, limit int) ([]Trip, *platform.ID, error) {
	return repo.trips, nil, nil
}

func (repo *memoryRepository) GetTrip(ctx context.Context, familyID, tripID platform.ID) (*TripDetail, error) {
	for _, t := range repo.trips {
		if t.FamilyID == familyID && t.ID == tripID {
			return &TripDetail{Trip: t, Items: repo.tripItems}, nil
		}
	}
	return nil, nil
}

func TestUncheckingClearsPurchasedData(t *testing.T) {
	repo := &memoryRepository{}
	ids := &memoryIDGenerator{nextID: 100}
	clock := memoryClock{now: time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)}
	service, err := NewService(repo, ids, clock)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	qty, _ := platform.ParseQuantity("2.500")
	item, err := service.CreateItem(context.Background(), 1, 10, CreateItemInput{
		Name:     "Oranges",
		Quantity: qty,
		Unit:     platform.UnitKilogram,
	})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}

	checked := true
	pqty, _ := platform.ParseQuantity("3.000")
	cents, _ := platform.NewCents(1599)
	updated, err := service.UpdateItem(context.Background(), 1, item.ID, 11, UpdateItemInput{
		IsChecked:         &checked,
		PurchasedQuantity: &pqty,
		TotalCents:        &cents,
	})
	if err != nil {
		t.Fatalf("check item: %v", err)
	}
	if !updated.IsChecked || updated.PurchasedQuantity == nil || updated.TotalCents == nil {
		t.Fatalf("checked item missing purchase data: %+v", updated)
	}
	if updated.AddedByID != 10 || updated.UpdatedByID != 11 {
		t.Fatalf("attribution mismatch: added=%d updated=%d", updated.AddedByID, updated.UpdatedByID)
	}

	unchecked := false
	uncheckUpdated, err := service.UpdateItem(context.Background(), 1, item.ID, 12, UpdateItemInput{
		IsChecked: &unchecked,
	})
	if err != nil {
		t.Fatalf("uncheck item: %v", err)
	}
	if uncheckUpdated.IsChecked || uncheckUpdated.PurchasedQuantity != nil || uncheckUpdated.TotalCents != nil {
		t.Fatalf("unchecking did not clear purchase data: %+v", uncheckUpdated)
	}
	if uncheckUpdated.AddedByID != 10 || uncheckUpdated.UpdatedByID != 12 {
		t.Fatalf("attribution mismatch: added=%d updated=%d", uncheckUpdated.AddedByID, uncheckUpdated.UpdatedByID)
	}
}

func TestCompleteListWithKeepDiscardAndOverrides(t *testing.T) {
	repo := &memoryRepository{}
	ids := &memoryIDGenerator{nextID: 200}
	clock := memoryClock{now: time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)}
	service, _ := NewService(repo, ids, clock)

	qty, _ := platform.ParseQuantity("1.000")
	item1, _ := service.CreateItem(context.Background(), 1, 10, CreateItemInput{Name: "Item 1", Quantity: qty, Unit: platform.UnitEach})
	item2, _ := service.CreateItem(context.Background(), 1, 10, CreateItemInput{Name: "Item 2", Quantity: qty, Unit: platform.UnitEach})
	item3, _ := service.CreateItem(context.Background(), 1, 10, CreateItemInput{Name: "Item 3", Quantity: qty, Unit: platform.UnitEach})

	checked := true
	cents, _ := platform.NewCents(450)
	_, _ = service.UpdateItem(context.Background(), 1, item1.ID, 10, UpdateItemInput{IsChecked: &checked, TotalCents: &cents})

	market := "Supermarket Downtown"
	detail, err := service.CompleteList(context.Background(), 1, 10, CompleteListInput{
		SupermarketName:        &market,
		DefaultUncheckedAction: ActionKeep,
		UncheckedItemOverrides: []ItemActionOverride{
			{ItemID: item3.ID, Action: ActionDiscard},
		},
	})
	if err != nil {
		t.Fatalf("complete list: %v", err)
	}
	if detail == nil {
		t.Fatal("expected trip detail, got nil")
	}
	if len(detail.Items) != 1 || detail.Items[0].SourceItemID != item1.ID {
		t.Fatalf("unexpected trip items: %+v", detail.Items)
	}
	if len(repo.items) != 1 || repo.items[0].ID != item2.ID {
		t.Fatalf("expected only kept Item 2 in list, got %+v", repo.items)
	}
	if repo.items[0].IsChecked {
		t.Fatal("kept item must be unchecked")
	}
}

func TestCompleteZeroPurchasesCreatesNoTrip(t *testing.T) {
	repo := &memoryRepository{}
	ids := &memoryIDGenerator{nextID: 300}
	clock := memoryClock{now: time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)}
	service, _ := NewService(repo, ids, clock)

	qty, _ := platform.ParseQuantity("1.000")
	item1, _ := service.CreateItem(context.Background(), 1, 10, CreateItemInput{Name: "Item 1", Quantity: qty, Unit: platform.UnitEach})
	item2, _ := service.CreateItem(context.Background(), 1, 10, CreateItemInput{Name: "Item 2", Quantity: qty, Unit: platform.UnitEach})

	detail, err := service.CompleteList(context.Background(), 1, 10, CompleteListInput{
		DefaultUncheckedAction: ActionKeep,
		UncheckedItemOverrides: []ItemActionOverride{
			{ItemID: item1.ID, Action: ActionDiscard},
		},
	})
	if err != nil {
		t.Fatalf("complete list: %v", err)
	}
	if detail != nil {
		t.Fatalf("zero-purchase completion should return nil trip, got %+v", detail)
	}
	if len(repo.trips) != 0 {
		t.Fatalf("expected 0 trips created, got %d", len(repo.trips))
	}
	if len(repo.items) != 1 || repo.items[0].ID != item2.ID {
		t.Fatalf("expected only kept Item 2, got %+v", repo.items)
	}
}

func TestCursorEncodingAndDecoding(t *testing.T) {
	id := platform.ID(241895317264384000)
	encoded := EncodeCursor(id)
	if encoded == "" {
		t.Fatal("encoded cursor is empty")
	}
	decoded, err := DecodeCursor(encoded)
	if err != nil || decoded != id {
		t.Fatalf("decoded cursor %v != %v (err: %v)", decoded, id, err)
	}

	for _, invalid := range []string{"", "invalid!base64", "bm90LWFuLWlk"} {
		if _, err := DecodeCursor(invalid); err == nil {
			t.Fatalf("invalid cursor %q was accepted", invalid)
		}
	}
}

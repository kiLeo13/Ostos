package mysql_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"ostos/backend/internal/families"
	"ostos/backend/internal/platform"
	"ostos/backend/internal/shopping"
)

func TestShoppingListLifecycleAndAtomicCompletionOnRealMySQL(t *testing.T) {
	suite := newIntegrationSuite(t)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	owner := suite.registerAndVerify(t, "shopown")
	member := suite.registerAndVerify(t, "shopmem")
	outsider := suite.registerAndVerify(t, "shopout")

	family, err := suite.families.Create(ctx, owner.ID, "Groceries Family", "req-create-fam")
	if err != nil {
		t.Fatalf("create family: %v", err)
	}

	// Outsider cannot view or add to list
	if _, err := suite.shopping.GetList(ctx, family.ID, outsider.ID); !errors.Is(err, shopping.ErrForbidden) {
		t.Fatalf("outsider get list err = %v, want ErrForbidden", err)
	}

	// Owner adds item 1
	qty1, _ := platform.ParseQuantity("2.000")
	item1, err := suite.shopping.CreateItem(ctx, family.ID, owner.ID, shopping.CreateItemInput{
		Name:     "Bananas",
		Quantity: qty1,
		Unit:     platform.UnitKilogram,
	})
	if err != nil {
		t.Fatalf("create item 1: %v", err)
	}
	if item1.AddedByID != owner.ID || item1.UpdatedByID != owner.ID || item1.Position != 1000 {
		t.Fatalf("unexpected item 1 state: %+v", item1)
	}

	// Owner adds item 2
	qty2, _ := platform.ParseQuantity("1.000")
	note2 := "Low fat"
	item2, err := suite.shopping.CreateItem(ctx, family.ID, owner.ID, shopping.CreateItemInput{
		Name:     "Milk",
		Quantity: qty2,
		Unit:     platform.UnitLiter,
		Note:     &note2,
	})
	if err != nil {
		t.Fatalf("create item 2: %v", err)
	}
	if item2.Position <= item1.Position {
		t.Fatalf("item 2 position %d should be greater than %d", item2.Position, item1.Position)
	}

	// Invite member and accept
	if err := suite.families.Invite(ctx, owner.ID, family.ID, families.InviteTarget{UserID: &member.ID}, "req-inv"); err != nil {
		t.Fatalf("invite: %v", err)
	}
	invitations, err := suite.families.ReceivedInvitations(ctx, member.ID)
	if err != nil || len(invitations) != 1 {
		t.Fatalf("invitations: %v, %v", invitations, err)
	}
	if err := suite.families.AcceptInvitation(ctx, member.ID, invitations[0].ID, "req-acc"); err != nil {
		t.Fatalf("accept: %v", err)
	}

	// Member edits item 1 (marks checked, enters purchased quantity and cents)
	checked := true
	pqty, _ := platform.ParseQuantity("2.250")
	cents, _ := platform.NewCents(1290)
	updated1, err := suite.shopping.UpdateItem(ctx, family.ID, item1.ID, member.ID, shopping.UpdateItemInput{
		IsChecked:         &checked,
		PurchasedQuantity: &pqty,
		TotalCents:        &cents,
	})
	if err != nil {
		t.Fatalf("update item 1: %v", err)
	}
	if !updated1.IsChecked || updated1.AddedByID != owner.ID || updated1.UpdatedByID != member.ID {
		t.Fatalf("attribution or check mismatch: %+v", updated1)
	}

	// Unchecking clears temporary purchase data
	unchecked := false
	uncheck1, err := suite.shopping.UpdateItem(ctx, family.ID, item1.ID, member.ID, shopping.UpdateItemInput{
		IsChecked: &unchecked,
	})
	if err != nil {
		t.Fatalf("uncheck item 1: %v", err)
	}
	if uncheck1.IsChecked || uncheck1.PurchasedQuantity != nil || uncheck1.TotalCents != nil {
		t.Fatalf("uncheck failed to clear purchase data: %+v", uncheck1)
	}

	// Re-check item 1 with purchased data
	_, err = suite.shopping.UpdateItem(ctx, family.ID, item1.ID, member.ID, shopping.UpdateItemInput{
		IsChecked:         &checked,
		PurchasedQuantity: &pqty,
		TotalCents:        &cents,
	})
	if err != nil {
		t.Fatalf("re-check item 1: %v", err)
	}

	// Member adds item 3
	qty3, _ := platform.ParseQuantity("3.000")
	item3, err := suite.shopping.CreateItem(ctx, family.ID, member.ID, shopping.CreateItemInput{
		Name:     "Bread",
		Quantity: qty3,
		Unit:     platform.UnitEach,
	})
	if err != nil {
		t.Fatalf("create item 3: %v", err)
	}

	// Reorder items
	reordered, err := suite.shopping.ReorderItems(ctx, family.ID, owner.ID, []platform.ID{item3.ID, item1.ID, item2.ID})
	if err != nil || len(reordered) != 3 {
		t.Fatalf("reorder items: %v, len=%d", err, len(reordered))
	}
	if reordered[0].ID != item3.ID || reordered[1].ID != item1.ID || reordered[2].ID != item2.ID {
		t.Fatalf("unexpected order: %+v", reordered)
	}

	// Complete list: item1 is checked -> purchased, item2 is kept, item3 is discarded via override
	supermarket := "Super Central"
	detail, err := suite.shopping.CompleteList(ctx, family.ID, member.ID, shopping.CompleteListInput{
		SupermarketName:        &supermarket,
		DefaultUncheckedAction: shopping.ActionKeep,
		UncheckedItemOverrides: []shopping.ItemActionOverride{
			{ItemID: item3.ID, Action: shopping.ActionDiscard},
		},
	})
	if err != nil {
		t.Fatalf("complete list: %v", err)
	}
	if detail == nil {
		t.Fatal("expected trip detail, got nil")
	}
	if detail.Trip.CompletedByID != member.ID || *detail.Trip.SupermarketName != supermarket {
		t.Fatalf("unexpected trip: %+v", detail.Trip)
	}
	if len(detail.Items) != 1 || detail.Items[0].SourceItemID != item1.ID {
		t.Fatalf("unexpected trip items: %+v", detail.Items)
	}
	if detail.Items[0].AddedByID != owner.ID {
		t.Fatalf("trip item original adder mismatch: %d != %d", detail.Items[0].AddedByID, owner.ID)
	}

	// Current list should only contain item 2, unchecked and cleared
	remaining, err := suite.shopping.GetList(ctx, family.ID, owner.ID)
	if err != nil || len(remaining) != 1 {
		t.Fatalf("remaining list items: %v, len=%d", err, len(remaining))
	}
	if remaining[0].ID != item2.ID || remaining[0].IsChecked {
		t.Fatalf("remaining item state: %+v", remaining[0])
	}

	// History listing
	trips, nextCursor, err := suite.shopping.ListTrips(ctx, family.ID, member.ID, "", 10)
	if err != nil || len(trips) != 1 {
		t.Fatalf("list trips: %v, len=%d", err, len(trips))
	}
	if nextCursor != "" {
		t.Fatalf("expected empty nextCursor, got %s", nextCursor)
	}

	// Trip detail
	tripDetail, err := suite.shopping.GetTrip(ctx, family.ID, detail.Trip.ID, owner.ID)
	if err != nil || len(tripDetail.Items) != 1 {
		t.Fatalf("get trip: %v, items=%+v", err, tripDetail)
	}

	// Zero-purchase completion: complete remaining unchecked item 2 with keep
	zeroDetail, err := suite.shopping.CompleteList(ctx, family.ID, owner.ID, shopping.CompleteListInput{
		DefaultUncheckedAction: shopping.ActionKeep,
	})
	if err != nil {
		t.Fatalf("zero-purchase complete: %v", err)
	}
	if zeroDetail != nil {
		t.Fatalf("expected nil trip for zero purchases, got %+v", zeroDetail)
	}
	// Still only 1 trip in history
	tripsAfter, _, _ := suite.shopping.ListTrips(ctx, family.ID, member.ID, "", 10)
	if len(tripsAfter) != 1 {
		t.Fatalf("expected 1 trip in history, got %d", len(tripsAfter))
	}
}

package mysql_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm"
	"ostos/backend/internal/accounts"
	"ostos/backend/internal/families"
	"ostos/backend/internal/platform"
	mysqlrepository "ostos/backend/internal/repository/mysql"
	"ostos/backend/internal/shopping"
)

type integrationSuite struct {
	database *gorm.DB
	accounts *accounts.Service
	families *families.Service
	shopping *shopping.Service
	cipher   *accounts.TokenCipher
}

func newIntegrationSuite(t *testing.T) integrationSuite {
	t.Helper()
	dsn := os.Getenv("OSTOS_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("OSTOS_TEST_MYSQL_DSN is not set; run the real-MySQL Compose integration test")
	}
	database, sqlDatabase, err := mysqlrepository.Open(dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = sqlDatabase.Close() })
	ids, err := platform.NewSnowflakeGenerator(773)
	if err != nil {
		t.Fatalf("ID generator: %v", err)
	}
	jwt, err := accounts.NewJWTManager(bytes.Repeat([]byte{3}, 32), "ostos-integration", "integration-client", 30*time.Minute)
	if err != nil {
		t.Fatalf("JWT: %v", err)
	}
	cipher, err := accounts.NewTokenCipher(bytes.Repeat([]byte{4}, 32))
	if err != nil {
		t.Fatalf("cipher: %v", err)
	}
	accountRepository := mysqlrepository.NewAccountRepository(database)
	accountService, err := accounts.NewService(accountRepository, ids, platform.SystemClock{}, accounts.Argon2Hasher{}, jwt, cipher)
	if err != nil {
		t.Fatalf("account service: %v", err)
	}
	familyService, err := families.NewService(mysqlrepository.NewFamilyRepository(database), ids, platform.SystemClock{})
	if err != nil {
		t.Fatalf("family service: %v", err)
	}
	shoppingService, err := shopping.NewService(mysqlrepository.NewShoppingRepository(database), ids, platform.SystemClock{})
	if err != nil {
		t.Fatalf("shopping service: %v", err)
	}
	return integrationSuite{database: database, accounts: accountService, families: familyService, shopping: shoppingService, cipher: cipher}
}

func (suite integrationSuite) registerAndVerify(t *testing.T, label string) accounts.PublicUser {
	t.Helper()
	user, token := suite.register(t, label)
	if err := suite.accounts.VerifyEmail(context.Background(), token); err != nil {
		t.Fatalf("verify %s: %v", label, err)
	}
	return user
}

func (suite integrationSuite) register(t *testing.T, label string) (accounts.PublicUser, string) {
	t.Helper()
	suffix := fmt.Sprintf("%x", time.Now().UnixNano())
	username := label + suffix
	if len(username) > 32 {
		username = username[:32]
	}
	user, err := suite.accounts.Register(context.Background(), accounts.RegisterInput{
		Email: username + "@example.test", Username: username, DisplayName: label, Password: "correct-password",
	})
	if err != nil {
		t.Fatalf("register %s: %v", label, err)
	}
	var outbox accounts.EmailOutbox
	if err := suite.database.Where("user_id = ? AND template = ?", user.ID, accounts.EmailTemplateVerifyEmail).
		Order("created_at DESC").Take(&outbox).Error; err != nil {
		t.Fatalf("verification outbox: %v", err)
	}
	token, err := suite.cipher.Decrypt(outbox.EncryptedTokenPayload)
	if err != nil {
		t.Fatalf("decrypt verification token: %v", err)
	}
	return user, token
}

func TestVerificationExpiryResendAndSingleUseOnRealMySQL(t *testing.T) {
	suite := newIntegrationSuite(t)
	ctx := context.Background()
	user, token := suite.register(t, "verify")
	if _, err := suite.accounts.Login(ctx, user.Email, "correct-password", nil); !errors.Is(err, accounts.ErrEmailNotVerified) {
		t.Fatalf("unverified login error = %v", err)
	}
	var accountToken accounts.AccountToken
	if err := suite.database.Where("user_id = ? AND purpose = ?", user.ID, accounts.TokenPurposeVerifyEmail).Take(&accountToken).Error; err != nil {
		t.Fatalf("load token: %v", err)
	}
	if err := suite.database.Model(&accounts.AccountToken{}).Where("id = ?", accountToken.ID).
		Update("expires_at", accountToken.CreatedAt.Add(time.Microsecond)).Error; err != nil {
		t.Fatalf("expire token: %v", err)
	}
	if err := suite.accounts.VerifyEmail(ctx, token); !errors.Is(err, accounts.ErrInvalidAccountToken) {
		t.Fatalf("expired verification error = %v", err)
	}
	if err := suite.accounts.ResendVerification(ctx, strings.ToUpper(user.Email)); err != nil {
		t.Fatalf("resend: %v", err)
	}
	var outbox accounts.EmailOutbox
	if err := suite.database.Where("user_id = ? AND template = ?", user.ID, accounts.EmailTemplateVerifyEmail).
		Order("created_at DESC").Take(&outbox).Error; err != nil {
		t.Fatalf("new outbox: %v", err)
	}
	newToken, err := suite.cipher.Decrypt(outbox.EncryptedTokenPayload)
	if err != nil {
		t.Fatalf("decrypt resent token: %v", err)
	}
	if err := suite.accounts.VerifyEmail(ctx, newToken); err != nil {
		t.Fatalf("verify resent token: %v", err)
	}
	if err := suite.accounts.VerifyEmail(ctx, newToken); !errors.Is(err, accounts.ErrInvalidAccountToken) {
		t.Fatalf("reused verification error = %v", err)
	}
	if err := suite.accounts.ForgotPassword(ctx, "missing-account@example.test"); err != nil {
		t.Fatalf("forgot-password disclosed missing account through error: %v", err)
	}
}

func TestAccountsAndFamiliesCriticalTransactionsOnRealMySQL(t *testing.T) {
	suite := newIntegrationSuite(t)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	owner := suite.registerAndVerify(t, "owner")
	member := suite.registerAndVerify(t, "member")

	pair, err := suite.accounts.Login(ctx, owner.Email, "correct-password", nil)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	rotated, err := suite.accounts.Refresh(ctx, pair.RefreshToken)
	if err != nil || rotated.RefreshToken == pair.RefreshToken {
		t.Fatalf("refresh rotation: pair=%+v err=%v", rotated, err)
	}
	independent, err := suite.accounts.Login(ctx, owner.Email, "correct-password", nil)
	if err != nil {
		t.Fatalf("independent login: %v", err)
	}
	if _, err := suite.accounts.Refresh(ctx, pair.RefreshToken); !errors.Is(err, accounts.ErrRefreshTokenReuse) {
		t.Fatalf("reused refresh token error = %v", err)
	}
	if _, err := suite.accounts.Refresh(ctx, rotated.RefreshToken); !errors.Is(err, accounts.ErrInvalidRefreshToken) {
		t.Fatalf("descendant refresh after reuse error = %v", err)
	}
	if _, err := suite.accounts.Refresh(ctx, independent.RefreshToken); err != nil {
		t.Fatalf("independent session was revoked by another chain's reuse: %v", err)
	}
	memberPair, err := suite.accounts.Login(ctx, member.Username, "correct-password", nil)
	if err != nil {
		t.Fatalf("member login: %v", err)
	}
	if err := suite.accounts.ForgotPassword(ctx, member.Email); err != nil {
		t.Fatalf("forgot password: %v", err)
	}
	var resetOutbox accounts.EmailOutbox
	if err := suite.database.Where("user_id = ? AND template = ?", member.ID, accounts.EmailTemplateResetPassword).
		Order("created_at DESC").Take(&resetOutbox).Error; err != nil {
		t.Fatalf("reset outbox: %v", err)
	}
	resetToken, err := suite.cipher.Decrypt(resetOutbox.EncryptedTokenPayload)
	if err != nil {
		t.Fatalf("decrypt reset token: %v", err)
	}
	if err := suite.accounts.ResetPassword(ctx, resetToken, "new-correct-password"); err != nil {
		t.Fatalf("reset password: %v", err)
	}
	if _, err := suite.accounts.Refresh(ctx, memberPair.RefreshToken); !errors.Is(err, accounts.ErrInvalidRefreshToken) {
		t.Fatalf("refresh after reset error = %v", err)
	}
	if _, err := suite.accounts.Login(ctx, member.Email, "correct-password", nil); !errors.Is(err, accounts.ErrInvalidCredentials) {
		t.Fatalf("old password login error = %v", err)
	}
	if _, err := suite.accounts.Login(ctx, member.Email, "new-correct-password", nil); err != nil {
		t.Fatalf("new password login: %v", err)
	}

	family, err := suite.families.Create(ctx, owner.ID, "Home", "request-create")
	if err != nil {
		t.Fatalf("create Family: %v", err)
	}
	if err := suite.accounts.DeleteAccount(ctx, owner.ID, "correct-password"); !errors.Is(err, accounts.ErrOwnsFamilies) {
		t.Fatalf("owner deletion error = %v", err)
	}
	if err := suite.families.Invite(ctx, owner.ID, family.ID, families.InviteTarget{Username: &member.Username}, "request-invite"); err != nil {
		t.Fatalf("invite: %v", err)
	}
	received, err := suite.families.ReceivedInvitations(ctx, member.ID)
	if err != nil || len(received) != 1 {
		t.Fatalf("received invitations = %+v, %v", received, err)
	}
	if err := suite.families.AcceptInvitation(ctx, member.ID, received[0].ID, "request-accept"); err != nil {
		t.Fatalf("accept invitation: %v", err)
	}
	members, err := suite.families.Members(ctx, owner.ID, family.ID)
	if err != nil || len(members) != 2 || members[0].Role != families.RoleOwner || members[0].ID != nil {
		t.Fatalf("members = %+v, %v", members, err)
	}
	var memberID platform.ID
	for _, item := range members {
		if item.UserID == member.ID && item.ID != nil {
			memberID = *item.ID
		}
	}
	if memberID == 0 {
		t.Fatal("accepted member row is missing")
	}
	if err := suite.families.ChangeRole(ctx, owner.ID, family.ID, memberID, families.RoleManager, "request-promote"); err != nil {
		t.Fatalf("promote: %v", err)
	}
	if err := suite.families.TransferOwnership(ctx, owner.ID, family.ID, member.ID, "request-transfer"); err != nil {
		t.Fatalf("transfer: %v", err)
	}
	after, err := suite.families.Members(ctx, member.ID, family.ID)
	if err != nil || len(after) != 2 {
		t.Fatalf("members after transfer = %+v, %v", after, err)
	}
	if err := suite.accounts.DeleteAccount(ctx, owner.ID, "correct-password"); err != nil {
		t.Fatalf("former owner deletion: %v", err)
	}
	after, err = suite.families.Members(ctx, member.ID, family.ID)
	if err != nil || len(after) != 1 || after[0].UserID != member.ID || after[0].Role != families.RoleOwner {
		t.Fatalf("members after account deletion = %+v, %v", after, err)
	}
}

func TestConcurrentRegistrationAndFamilyLimitOnRealMySQL(t *testing.T) {
	suite := newIntegrationSuite(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	email := fmt.Sprintf("same-%x@example.test", time.Now().UnixNano())
	results := make(chan error, 2)
	var wait sync.WaitGroup
	for index := range 2 {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			candidateEmail := email
			if index == 0 {
				candidateEmail = "  " + strings.ToUpper(email) + "  "
			}
			_, err := suite.accounts.Register(ctx, accounts.RegisterInput{
				Email: candidateEmail, Username: fmt.Sprintf("same%x%d", time.Now().UnixNano(), index), DisplayName: "Same", Password: "correct-password",
			})
			results <- err
		}(index)
	}
	wait.Wait()
	close(results)
	successes, conflicts := 0, 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, accounts.ErrEmailTaken):
			conflicts++
		default:
			t.Fatalf("unexpected registration error: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("registration results: successes=%d conflicts=%d", successes, conflicts)
	}

	target := suite.registerAndVerify(t, "limit")
	for index := range 9 {
		if _, err := suite.families.Create(ctx, target.ID, fmt.Sprintf("Owned %d", index), "request-limit"); err != nil {
			t.Fatalf("create owned Family %d: %v", index, err)
		}
	}
	inviter := suite.registerAndVerify(t, "inviter")
	var invitationIDs []platform.ID
	for index := range 2 {
		family, err := suite.families.Create(ctx, inviter.ID, fmt.Sprintf("Invite %d", index), "request-inviter")
		if err != nil {
			t.Fatalf("create invitation Family: %v", err)
		}
		if err := suite.families.Invite(ctx, inviter.ID, family.ID, families.InviteTarget{UserID: &target.ID}, "request-invite-limit"); err != nil {
			t.Fatalf("invite target: %v", err)
		}
		var invitation families.Invitation
		if err := suite.database.Where("family_id = ? AND invited_user_id = ?", family.ID, target.ID).Take(&invitation).Error; err != nil {
			t.Fatalf("read invitation: %v", err)
		}
		invitationIDs = append(invitationIDs, invitation.ID)
	}
	acceptResults := make(chan error, 2)
	for _, invitationID := range invitationIDs {
		wait.Add(1)
		go func(id platform.ID) {
			defer wait.Done()
			acceptResults <- suite.families.AcceptInvitation(ctx, target.ID, id, "request-concurrent-accept")
		}(invitationID)
	}
	wait.Wait()
	close(acceptResults)
	successes, conflicts = 0, 0
	for err := range acceptResults {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, families.ErrFamilyLimit):
			conflicts++
		default:
			t.Fatalf("unexpected acceptance error: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("acceptance results: successes=%d conflicts=%d", successes, conflicts)
	}
	count, err := mysqlrepository.NewFamilyRepository(suite.database).CountUserFamilies(ctx, target.ID)
	if err != nil || count != 10 {
		t.Fatalf("Family count = %d, %v", count, err)
	}
}

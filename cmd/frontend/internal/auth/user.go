package auth

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
)

// MockCreateOrUpdateUser is used in tests to mock CreateOrUpdateUser.
var MockCreateOrUpdateUser func(db.NewUser, db.ExternalAccountSpec) (int32, error)

// CreateOrUpdateUser creates or updates a user using the provided information, looking up a user by
// the external account provided.
func CreateOrUpdateUser(ctx context.Context, newOrUpdatedUser db.NewUser, externalAccount db.ExternalAccountSpec) (userID int32, err error) {
	if MockCreateOrUpdateUser != nil {
		return MockCreateOrUpdateUser(newOrUpdatedUser, externalAccount)
	}

	// TEMPORARY: Copy external account info to db.NewUser (soon these fields will be removed from
	// db.NewUser).
	newOrUpdatedUser.ExternalProvider = externalAccount.ServiceID
	newOrUpdatedUser.ExternalID = externalAccount.AccountID

	usr, err := db.Users.GetByExternalID(ctx, newOrUpdatedUser.ExternalProvider, newOrUpdatedUser.ExternalID)
	if errcode.IsNotFound(err) {
		usr, err = db.Users.Create(ctx, newOrUpdatedUser)

		// Handle the race condition where the new user performs two requests and both try to create
		// the user.
		if err != nil {
			// If GetByExternalID fails, return original Create error (err); otherwise clear the error.
			var err2 error
			usr, err2 = db.Users.GetByExternalID(ctx, newOrUpdatedUser.ExternalProvider, newOrUpdatedUser.ExternalID)
			if err2 == nil {
				err = nil
			}
		}
	}
	if err != nil {
		return 0, err
	}

	// Update user in our DB if their profile info changed on the issuer. (Except username,
	// which the user is somewhat likely to want to control separately on Sourcegraph.)
	var userUpdate db.UserUpdate
	if usr.DisplayName != newOrUpdatedUser.DisplayName {
		userUpdate.DisplayName = &newOrUpdatedUser.DisplayName
	}
	if usr.AvatarURL != newOrUpdatedUser.AvatarURL {
		userUpdate.AvatarURL = &newOrUpdatedUser.AvatarURL
	}
	if userUpdate != (db.UserUpdate{}) {
		if err := db.Users.Update(ctx, usr.ID, userUpdate); err != nil {
			return 0, err
		}
	}
	return usr.ID, nil
}

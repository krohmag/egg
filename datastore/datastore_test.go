package datastore

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDatastoreInterfaces(t *testing.T) {
	eggIncID := uuid.New().String()
	testUser := User{
		EggIncID:    eggIncID,
		DiscordName: "krohmag",
	}

	testUserUpdate := User{
		EggIncID:        eggIncID,
		GameAccountName: "akroh",
		SoulFood:        1,
		ProphecyBonus:   1,
		SoulEggs:        1,
		ProphecyEggs:    1,
	}

	db, err := ConnectDatabase("sqlite-in-memory", true)
	require.NoError(t, err)

	datastore := Database{DB: db}

	require.NoError(t, datastore.DB.AutoMigrate(User{}))
	require.True(t, datastore.Ping())

	ctx := context.Background()

	t.Run("test rollback", func(t *testing.T) {
		require.NoError(t, runTransaction(t, ctx, datastore, testUser, testUserUpdate, false))
	})

	t.Run("test commit", func(t *testing.T) {
		require.NoError(t, runTransaction(t, ctx, datastore, testUser, testUserUpdate, true))

		var deletedUser User
		var txErr error
		tx, txErr := datastore.Transaction(ctx)
		require.NoError(t, txErr)

		user, txErr := tx.GetUserByEggIncUserID(eggIncID)
		require.NoError(t, txErr)
		require.Equal(t, testUserUpdate.ProphecyBonus, user.ProphecyBonus)

		usersByDiscordName, txErr := tx.GetUsersByDiscordName(testUser.DiscordName)
		require.NoError(t, txErr)
		require.Equal(t, testUserUpdate.SoulFood, usersByDiscordName[0].SoulFood)

		users, txErr := tx.GetUsers()
		require.NoError(t, txErr)
		require.Equal(t, []string{eggIncID}, users.GetEggIncIDs())

		require.NoError(t, tx.DeleteUser(testUser))
		require.NoError(t, tx.Commit())

		require.Errorf(t, datastore.DB.Where("egg_inc_id = ?", eggIncID).First(&deletedUser).Error, "record not found")
		require.Empty(t, deletedUser)

		require.NoError(t, datastore.DB.Unscoped().Where("egg_inc_id = ?", eggIncID).First(&deletedUser).Error)
		require.Equal(t, testUserUpdate.ProphecyEggs, deletedUser.ProphecyEggs)

		require.NoError(t, datastore.DB.Unscoped().Where("egg_inc_id = ?", eggIncID).Delete(&deletedUser).Error)
	})
}

func runTransaction(t *testing.T, ctx context.Context, datastore Database, testUser, testUserUpdate User, commit bool) error {
	tx, err := datastore.Transaction(ctx)
	if err != nil {
		return err
	}

	record, err := tx.CreateOrUpdateUser(testUser)
	if err != nil {
		return err
	}
	require.Equal(t, testUser.DiscordName, record.DiscordName)

	record2, err := tx.CreateOrUpdateUser(testUserUpdate)
	if err != nil {
		return err
	}
	require.Equal(t, testUserUpdate.SoulEggs, record2.SoulEggs)

	switch commit {
	case true:
		return tx.Commit()
	case false:
		return tx.Rollback()
	}

	return nil
}

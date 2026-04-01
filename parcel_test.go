package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func TestAddGetDelete(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	number, err := store.Add(parcel)
	parcel.Number = number
	require.NoError(t, err)
	assert.NotZero(t, number)

	p, err := store.Get(number)
	require.NoError(t, err)
	assert.Equal(t, parcel, p)

	err = store.Delete(number)
	require.NoError(t, err)
	_, err = store.Get(number)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestSetAddress(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	number, err := store.Add(parcel)
	require.NoError(t, err)
	assert.NotZero(t, number)

	newAddress := "new test address"
	err = store.SetAddress(number, newAddress)
	require.NoError(t, err)

	p, err := store.Get(number)
	require.NoError(t, err)
	assert.Equal(t, newAddress, p.Address)
}

func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	number, err := store.Add(parcel)
	require.NoError(t, err)
	assert.NotZero(t, number)

	err = store.SetStatus(number, ParcelStatusSent)
	require.NoError(t, err)

	p, err := store.Get(number)
	require.NoError(t, err)
	assert.NotEqual(t, parcel.Status, p.Status)
}

func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		assert.NotZero(t, id)

		parcels[i].Number = id

		parcelMap[id] = parcels[i]
	}

	storedParcels, err := store.GetByClient(client)

	require.NoError(t, err)
	assert.Equal(t, len(parcels), len(storedParcels))

	for _, parcel := range storedParcels {
		expectedParcel, exist := parcelMap[parcel.Number]
		assert.True(t, exist, "Parcel with number: %d not found", parcel.Number)

		assert.Equal(t, expectedParcel, parcel)
	}
}

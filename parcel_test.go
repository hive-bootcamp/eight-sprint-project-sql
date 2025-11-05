package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

func getTestParcel() Parcel {
	return Parcel{
		Client:    randRange.Intn(10_000_000),
		Status:    ParcelStatusRegistered,
		Address:   "test address",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// вспомогательная функция: создаёт in-memory базу и таблицу
func prepareTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE parcel (
		number INTEGER PRIMARY KEY AUTOINCREMENT,
		client INTEGER,
		status TEXT,
		address TEXT,
		created_at TEXT
	);`)
	require.NoError(t, err)

	return db
}

func TestAddGetDelete(t *testing.T) {
	db := prepareTestDB(t)
	store := NewParcelStore(db)
	p := getTestParcel()

	// add
	id, err := store.Add(p)
	require.NoError(t, err)
	require.NotZero(t, id)
	p.Number = id

	// get
	got, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, p.Client, got.Client)
	require.Equal(t, p.Status, got.Status)
	require.Equal(t, p.Address, got.Address)
	require.Equal(t, p.CreatedAt, got.CreatedAt)

	// delete
	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	require.Error(t, err)
}

func TestSetAddress(t *testing.T) {
	db := prepareTestDB(t)
	store := NewParcelStore(db)
	p := getTestParcel()

	id, err := store.Add(p)
	require.NoError(t, err)

	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	got, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, got.Address)
}

func TestSetStatus(t *testing.T) {
	db := prepareTestDB(t)
	store := NewParcelStore(db)
	p := getTestParcel()

	id, err := store.Add(p)
	require.NoError(t, err)

	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	got, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newStatus, got.Status)
}

func TestGetByClient(t *testing.T) {
	db := prepareTestDB(t)
	store := NewParcelStore(db)

	client := randRange.Intn(10_000_000)
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}

	for i := range parcels {
		parcels[i].Client = client
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		parcels[i].Number = id
	}

	got, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, got, len(parcels))

	for _, gp := range got {
		require.Equal(t, client, gp.Client)
		require.NotEmpty(t, gp.Address)
		require.NotEmpty(t, gp.Status)
	}
}

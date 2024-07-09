package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	// Создание объекта ParcelStore
	store := NewParcelStore(db)
	require.NotNil(t, store)

	// Подготовка тестовой посылки
	parcel := getTestParcel()

	// Добавление новой посылки в БД
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// Получение только что добавленной посылки из БД
	storedParcel, err := store.Get(id)
	require.NoError(t, err)

	// Проверка, что значения всех полей в полученной посылке соответствуют тестовой посылке
	require.Equal(t, parcel.Client, storedParcel.Client)
	require.Equal(t, parcel.Status, storedParcel.Status)
	require.Equal(t, parcel.Address, storedParcel.Address)
	require.Equal(t, parcel.CreatedAt, storedParcel.CreatedAt)

	// Удаление добавленной посылки из БД
	err = store.Delete(id)
	require.NoError(t, err)

	// Попытка получить удалённую посылку из БД
	_, err = store.Get(id)
	require.Error(t, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	// Создание объекта ParcelStore
	store := NewParcelStore(db)
	require.NotNil(t, store)

	// Добавление новой посылки в БД
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// Обновление адреса у добавленной посылки
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// Проверка, что адрес обновился
	updatedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, updatedParcel.Address)

	// Удаление добавленной посылки из БД
	err = store.Delete(id)
	require.NoError(t, err)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	// add
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)
	parcel.Number = id

	// set status
	err = store.SetStatus(parcel.Number, ParcelStatusSent)
	require.NoError(t, err)

	// check
	updatedParcel, err := store.Get(parcel.Number)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusSent, updatedParcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	// Создание объекта ParcelStore
	store := NewParcelStore(db)
	require.NotNil(t, store)

	// Подготовка тестовых посылок
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	client := randRange.Intn(10_000_000)

	// Добавление всех тестовых посылок с одинаковым идентификатором клиента
	for i := 0; i < len(parcels); i++ {
		parcels[i].Client = client
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		parcels[i].Number = id
	}

	// Получение списка посылок по идентификатору клиента
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)

	// Проверка, что количество полученных посылок совпадает с количеством добавленных
	require.Len(t, storedParcels, len(parcels))

	// Проверка, что все добавленные посылки присутствуют в полученном списке и их значения соответствуют
	for _, storedParcel := range storedParcels {
		found := false
		for _, p := range parcels {
			if storedParcel.Number == p.Number {
				found = true
				require.Equal(t, p.Client, storedParcel.Client)
				require.Equal(t, p.Status, storedParcel.Status)
				require.Equal(t, p.Address, storedParcel.Address)
				require.Equal(t, p.CreatedAt, storedParcel.CreatedAt)
				break
			}
		}
		require.True(t, found, "Не удалось найти посылку с идентификатором %d", storedParcel.Number)
	}

	// Удаление всех добавленных посылок из БД
	for _, p := range parcels {
		err = store.Delete(p.Number)
		require.NoError(t, err)
	}
}

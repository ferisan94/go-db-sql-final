package main

import (
	"database/sql"
	"errors"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	stmt := `INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)`
	result, err := s.db.Exec(stmt, p.Client, p.Status, p.Address, p.CreatedAt)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	var p Parcel
	row := s.db.QueryRow("SELECT * FROM parcel WHERE number = ?", number)
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return p, err
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	var parcels []Parcel
	rows, err := s.db.Query("SELECT * FROM parcel WHERE client = ?", client)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p Parcel
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		parcels = append(parcels, p)
	}

	return parcels, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// Проверяем, можно ли изменить статус
	if status != ParcelStatusSent && status != ParcelStatusDelivered {
		return errors.New("невозможно изменить статус на указанный")
	}

	stmt := `UPDATE parcel SET status = ? WHERE number = ? AND status = ?`
	result, err := s.db.Exec(stmt, status, number, ParcelStatusRegistered)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("невозможно изменить статус")
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	stmt := `UPDATE parcel SET address = ? WHERE number = ? AND status = ?`
	result, err := s.db.Exec(stmt, address, number, ParcelStatusRegistered)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("невозможно изменить адрес для данной посылки")
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	stmt := `DELETE FROM parcel WHERE number = ? AND status = ?`
	result, err := s.db.Exec(stmt, number, ParcelStatusRegistered)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("невозможно удалить посылку с данным статусом")
	}

	return nil
}

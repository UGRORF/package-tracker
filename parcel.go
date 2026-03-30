package main

import (
	"database/sql"
	"errors"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	query := "INSERT INTO parcel (client, status, address, created_at) " +
		"VALUES (:client, :status, :address, :created_at)"
	res, err := s.db.Exec(query,
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert parcel: %w", err)
	}
	number, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert number: %w", err)
	}

	return int(number), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	query := "SELECT number, client, status, address, created_at FROM parcel WHERE number = :number"
	p := Parcel{}
	err := s.db.QueryRow(query, sql.Named("number", number)).Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Parcel{}, fmt.Errorf("parcel with number %d not found: %w", number, err)
		}
		return Parcel{}, fmt.Errorf("failed to get parcel: %w", err)
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	query := "SELECT number, client, status, address, created_at " +
		"FROM parcel " +
		"WHERE client = :client"
	rows, err := s.db.Query(query, sql.Named("client", client))
	if err != nil {
		return nil, fmt.Errorf("failed to get parcels: %w", err)
	}
	defer rows.Close()

	var res []Parcel

	for rows.Next() {
		p := Parcel{}
		err = rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to get parcel: %w", err)
		}
		res = append(res, p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	query := "UPDATE parcel SET status = :status WHERE number = :number"
	res, err := s.db.Exec(query, sql.Named("status", status), sql.Named("number", number))
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("parcel with number %d not found", number)
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	parcel, err := s.Get(number)
	if err != nil {
		return err
	}
	if parcel.Status != ParcelStatusRegistered {
		return fmt.Errorf("parcel`s %d status is not registered", number)
	}
	query := "UPDATE parcel SET address = :address WHERE number = :number"
	res, err := s.db.Exec(query, sql.Named("address", address), sql.Named("number", number))
	if err != nil {
		return fmt.Errorf("failed to update address: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("parcel with number %d not found", number)
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	parcel, err := s.Get(number)
	if err != nil {
		return err
	}
	if parcel.Status != ParcelStatusRegistered {
		return fmt.Errorf("parcel`s %d status is not registered", number)
	}

	query := "DELETE FROM parcel WHERE number = :number"
	res, err := s.db.Exec(query, sql.Named("number", number))
	if err != nil {
		return fmt.Errorf("failed to delete parcel with number: %d. %w", number, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("parcel with number %d not found", number)
	}

	return nil
}

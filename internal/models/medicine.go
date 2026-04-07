package models

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Medicine struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	InStock     bool    `json:"in_stock"`
	Category    string  `json:"category"`
}

type MedicineStore struct{ db *pgxpool.Pool }

func NewMedicineStore(db *pgxpool.Pool) *MedicineStore { return &MedicineStore{db: db} }

func (s *MedicineStore) List(ctx context.Context) ([]Medicine, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, name, description, price, in_stock, category FROM medicines WHERE in_stock = true ORDER BY category, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var medicines []Medicine
	for rows.Next() {
		var m Medicine
		if err := rows.Scan(&m.ID, &m.Name, &m.Description, &m.Price, &m.InStock, &m.Category); err != nil {
			return nil, err
		}
		medicines = append(medicines, m)
	}
	return medicines, rows.Err()
}

func (s *MedicineStore) GetByID(ctx context.Context, id string) (*Medicine, error) {
	var m Medicine
	err := s.db.QueryRow(ctx,
		`SELECT id, name, description, price, in_stock, category FROM medicines WHERE id = $1`, id).
		Scan(&m.ID, &m.Name, &m.Description, &m.Price, &m.InStock, &m.Category)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

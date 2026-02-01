package models

import "time"

type RawMaterials struct {
	Id        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Price     float64   `json:"price" db:"price"`
	Quantity  float64   `json:"quantity" db:"quantity"`
	Unit      string    `json:"unit" db:"unit"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" `
}

type CreateRawMaterialType struct {
	Name     string  `json:"name" db:"name"`
	Price    float64 `json:"price" db:"price"`
	Quantity float64 `json:"quantity" db:"quantity"`
	Unit     string  `json:"unit" db:"unit"`
}

type UpdateRawMaterials struct {
	Id       string  `json:"id" db:"id"`
	Name     string  `json:"name" db:"name"`
	Price    float64 `json:"price" db:"price"`
	Quantity float64 `json:"quantity" db:"quantity"`
	Unit     string  `json:"unit" db:"unit"`
}

type DeleteRawMaterialsPayload struct {
	RawMaterialIds []string `json:"raw_materials_ids"`
}

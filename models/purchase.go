package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PurchaseItem struct {
	MedicineID   primitive.ObjectID `bson:"medicine_id" json:"medicine_id"`
	MedicineName string             `bson:"medicine_name" json:"medicine_name"`
	Quantity     int                `bson:"quantity" json:"quantity"`
	UnitPrice    float64            `bson:"unit_price" json:"unit_price"`
	Total        float64            `bson:"total" json:"total"`
	BatchNumber  string             `bson:"batch_number" json:"batch_number"`
	ExpiryDate   time.Time          `bson:"expiry_date" json:"expiry_date"`
}

type Purchase struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PONumber      string             `bson:"po_number" json:"po_number"`
	SupplierID    primitive.ObjectID `bson:"supplier_id" json:"supplier_id"`
	SupplierName  string             `bson:"supplier_name" json:"supplier_name"`
	Items         []PurchaseItem     `bson:"items" json:"items"`
	TotalAmount   float64            `bson:"total_amount" json:"total_amount"`
	Status        string             `bson:"status" json:"status"` // pending, received, cancelled
	Notes         string             `bson:"notes" json:"notes"`
	OrderedBy     string             `bson:"ordered_by" json:"ordered_by"`
	ReceivedAt    *time.Time         `bson:"received_at,omitempty" json:"received_at"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

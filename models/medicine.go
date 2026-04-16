package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Medicine struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name            string             `bson:"name" json:"name"`
	GenericName     string             `bson:"generic_name" json:"generic_name"`
	CategoryID      primitive.ObjectID `bson:"category_id" json:"category_id"`
	CategoryName    string             `bson:"category_name,omitempty" json:"category_name"`
	SupplierID      primitive.ObjectID `bson:"supplier_id" json:"supplier_id"`
	SupplierName    string             `bson:"supplier_name,omitempty" json:"supplier_name"`
	Manufacturer    string             `bson:"manufacturer" json:"manufacturer"`
	BatchNumber     string             `bson:"batch_number" json:"batch_number"`
	ExpiryDate      time.Time          `bson:"expiry_date" json:"expiry_date"`
	Quantity        int                `bson:"quantity" json:"quantity"`
	MinStockLevel   int                `bson:"min_stock_level" json:"min_stock_level"`
	PurchasePrice   float64            `bson:"purchase_price" json:"purchase_price"`
	SellingPrice    float64            `bson:"selling_price" json:"selling_price"`
	Unit            string             `bson:"unit" json:"unit"` // tablet, ml, mg, etc.
	Description     string             `bson:"description" json:"description"`
	Prescription    bool               `bson:"prescription" json:"prescription"`
	Active          bool               `bson:"active" json:"active"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}

func (m Medicine) IsLowStock() bool {
	return m.Quantity <= m.MinStockLevel
}

func (m Medicine) IsExpired() bool {
	if m.ExpiryDate.IsZero() {
		return false
	}
	return time.Now().After(m.ExpiryDate)
}

func (m Medicine) IsExpiringSoon() bool {
	if m.ExpiryDate.IsZero() {
		return false
	}
	return time.Now().AddDate(0, 1, 0).After(m.ExpiryDate) && !m.IsExpired()
}

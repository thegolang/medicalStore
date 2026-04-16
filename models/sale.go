package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SaleItem struct {
	MedicineID   primitive.ObjectID `bson:"medicine_id" json:"medicine_id"`
	MedicineName string             `bson:"medicine_name" json:"medicine_name"`
	Quantity     int                `bson:"quantity" json:"quantity"`
	UnitPrice    float64            `bson:"unit_price" json:"unit_price"`
	Total        float64            `bson:"total" json:"total"`
}

type Sale struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	InvoiceNumber string             `bson:"invoice_number" json:"invoice_number"`
	CustomerName  string             `bson:"customer_name" json:"customer_name"`
	CustomerPhone string             `bson:"customer_phone" json:"customer_phone"`
	Items         []SaleItem         `bson:"items" json:"items"`
	SubTotal      float64            `bson:"sub_total" json:"sub_total"`
	Discount      float64            `bson:"discount" json:"discount"`
	Tax           float64            `bson:"tax" json:"tax"`
	GrandTotal    float64            `bson:"grand_total" json:"grand_total"`
	PaymentMethod string             `bson:"payment_method" json:"payment_method"` // cash, card, upi
	AmountPaid    float64            `bson:"amount_paid" json:"amount_paid"`
	Change        float64            `bson:"change" json:"change"`
	Notes         string             `bson:"notes" json:"notes"`
	SoldBy        string             `bson:"sold_by" json:"sold_by"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
}

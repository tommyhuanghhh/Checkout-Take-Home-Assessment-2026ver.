package uuid

import (
	"PaymentGateway/internal/domain"
	"github.com/google/uuid"
)

//build time check
var _ domain.IDGenerator = (*UUIDGenerator)(nil)

type UUIDGenerator struct {}

func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

func (g *UUIDGenerator) Generate() string {
	id, _ := uuid.NewV7()
	return id.String()
}
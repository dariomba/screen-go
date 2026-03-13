package uuid

import "github.com/oklog/ulid/v2"

type UlidGenerator struct{}

func NewUlidGenerator() *UlidGenerator {
	return &UlidGenerator{}
}

func (u UlidGenerator) Generate() string {
	return ulid.Make().String()
}

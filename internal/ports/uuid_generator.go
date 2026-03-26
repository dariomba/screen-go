package ports

//go:generate go tool go.uber.org/mock/mockgen -source=uuid_generator.go -destination=../mocks/uuid_generator_mock.go -package=mocks
type UUIDGenerator interface {
	Generate() string
}

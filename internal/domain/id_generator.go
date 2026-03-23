package domain

type IDGenerator interface {
    Generate() string
}
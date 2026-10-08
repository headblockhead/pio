package simulation

import "uuid"

type ComponentIdentifier uuid.UUID

func NewComponentIdentifier() ComponentIdentifier {
	return ComponentIdentifier(uuid.New())
}

type Component interface {
	ID() ComponentIdentifier
	Label() string
}

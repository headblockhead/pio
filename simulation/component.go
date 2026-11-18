package simulation

import "uuid"

type ComponentIdentifier uuid.UUID

func NewComponentIdentifier() ComponentIdentifier {
	return ComponentIdentifier(uuid.New())
}

func (id ComponentIdentifier) String() string {
	return uuid.UUID(id).String()
}

type Component interface {
	ID() ComponentIdentifier
	Label() string
}

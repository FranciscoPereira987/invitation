package utils

import (
	"math/rand"
	"time"
)

type Choser struct {
	missing []uint
	source *rand.Rand
}

func NewChooser(from []uint) *Choser {
	missing := make([]uint, len(from))
	copy(missing, from)
	source := rand.New(rand.NewSource(time.Now().Unix()))

	return &Choser{
		missing,
		source,
	}
}

func (c *Choser) Choose() uint {
	chosenOneAt := c.source.Intn(len(c.missing))
	chosenOne := c.missing[chosenOneAt]
	c.missing[chosenOneAt] = c.missing[len(c.missing)-1]
	c.missing = c.missing[:len(c.missing)-1]
	return chosenOne
}
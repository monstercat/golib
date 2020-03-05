package operator

type Modifier rune

// Default Modifiers
var (
	ModifierOr  Modifier = '|'
	ModifierNot Modifier = '!'
	ModifierAnd Modifier = '&'

	Modifiers = []Modifier{
		ModifierOr, ModifierNot, ModifierAnd,
	}
)

func (m Modifier) Matches(r byte) bool {
	return byte(m) == r
}
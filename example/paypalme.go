package example

//go:generate cel-spec-go gen

type (
	// Paypalme returns a new URL that generates by given params.
	Paypalme func(username string, amount string, currency string) string
)

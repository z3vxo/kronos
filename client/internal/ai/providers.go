package ai

type Provider interface {
	Name() string
	Model() string
	SetModel(string)
	APIKeySet() bool
	Generate(prompt string, temp float32) (string, error)
}

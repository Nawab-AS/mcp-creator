package mcpserver

import (
	"github.com/sugarme/tokenizer"
	"github.com/sugarme/tokenizer/pretrained"
)

type Tokenizer struct {
	Tokenizer *tokenizer.Tokenizer
}

func LoadTokenizer(tokenizer_path string) (*Tokenizer, error) {
	t, err := pretrained.FromFile(tokenizer_path)
	if err != nil {
		return nil, err
	}
	return &Tokenizer{Tokenizer: t}, nil
}

package helper

import gonanoid "github.com/matoous/go-nanoid/v2"

const urlAlphabet = "useandom-26T198340PX75pxJACKVERYMINDBUSHWOLF_GQZbfghjklqvwyzrict"

func GenerateID() (string, error) {
	id, err := gonanoid.Generate(urlAlphabet, 16)
	if err != nil {
		return "", err
	}
	return id, nil
}

func Map[T any, R any](input []T, f func(T) R) []R {
	output := make([]R, len(input))
	for i, v := range input {
		output[i] = f(v)
	}
	return output
}

package summary

type Summarizer interface {
	Summarize(text string, maxLen int) (string, error)
}

type Stub struct {
}

func (s *Stub) Summarize(text string, maxLen int) (string, error) {
	runes := []rune(text)

	if len(runes) <= maxLen {
		return text, nil
	}

	runes = runes[:maxLen]

	for i := len(runes) - 1; i > 0; i-- {
		v := runes[i]
		if v == '.' || v == '!' || v == '?' {
			return string(runes[:i+1]), nil
		}
	}

	for i := len(runes) - 1; i > 0; i-- {
		v := runes[i]
		if v == ' ' {
			return string(runes[:i]) + "...", nil
		}
	}

	return string(runes[:maxLen]) + "...", nil
}

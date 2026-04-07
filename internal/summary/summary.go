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

	cut := string(runes[maxLen:])

	for i := len(cut) - 1; i > 0; i-- {
		v := cut[i]
		if v == '.' || v == '!' || v == '?' {
			return string(runes[:i]), nil
		}
	}

	for i := len(cut); i >= 0; i-- {
		v := cut[i]
		if v == ' ' {
			return string(runes[:i]), nil
		}
	}

	return string(runes[:maxLen]) + "...", nil
}

package summary

type Summarizer interface {
	Summarize(text string, maxLen int) (string, error)
}

type Sub struct {
}

func (s *Sub) Summarize(text string, maxLen int) (string, error) {
	// TODO:: Остановочка здесь
}

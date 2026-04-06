package news

type Source struct {
	ID   *int64 `json:"id"`
	Name string `json:"name"`
}

type NewsAPIArticle struct {
	Source      Source `json:"source"`
	Author      string `json:"author"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Url         string `json:"url"`
	UrlToImage  string `json:"urlToImage"`
	PublishedAt string `json:"publishedAt"`
	Content     string `json:"content"`
}

type NewsAPIArticles struct {
	Status       string           `json:"status"`
	TotalResults int              `json:"totalResults"`
	Articles     []NewsAPIArticle `json:"articles"`
}

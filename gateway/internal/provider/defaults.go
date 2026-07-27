package provider

func NewDefaultRouter() *Router {
	router := NewRouter()

	router.Register("mock", NewMockProvider())
	router.Register("ollama", NewOllamaProvider())

	return router
}

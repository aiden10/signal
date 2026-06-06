package testing

type TestData struct {
	TestGroupId string
	Username string
	GeminiPromptMessage string
	GeminiSearchQuestion string
	TestMessageGeneric string
}

var Data = TestData{
	TestGroupId: "",
	Username: "Test User",
	GeminiPromptMessage: "@gemini this is a test message. Hello",
	GeminiSearchQuestion: "@gemini what country had the most medals at the most recent olympic games?",
	TestMessageGeneric: "This is a test message",
}

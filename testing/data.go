package testing

type TestData struct {
	TestGroupId string
	Username string
	GeminiPromptMessage string
}

var Data = TestData{
	TestGroupId: "",
	Username: "Test User",
	GeminiPromptMessage: "@gemini this is a test message. Hello",
}

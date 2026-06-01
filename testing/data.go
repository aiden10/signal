package testing

type TestData struct {
	TestGroupId string
	Username string
	GeminiPromptMessage string
	TestMessageGeneric string
}

var Data = TestData{
	TestGroupId: "",
	Username: "Test User",
	GeminiPromptMessage: "@gemini this is a test message. Hello",
	TestMessageGeneric: "This is a test message",
}

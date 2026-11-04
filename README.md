# About

Signal bot that allows an AI to send messages in the chat by writing `@gemini`. DeepSeek generates the text and image responses, while Gemini is only used for embeddings. The bot also replies automatically when someone replies to one of its messages. Commands like `/history 50` and `/system-prompt ...` let anyone customize the bot per chat.

# Structure
## main.go
Entry point, initializes the services and retrieves data from `.env`.

## llm.go
Provides functions for generating chat responses with DeepSeek and embedding messages as vectors with Gemini. Request logic for the DeepSeek API lives in `llm/deepseek.go`.

## handlers/history.go
Provides an interface for storing and retrieving messages. When the bot is prompted, the prompt is turned into a vector and a cosine similarity search is performed against the memory database to find similar messages to be used as context. It also stores bot message timestamps and attachment metadata.

## handlers/events.go
Contains functions which are called depending on the event that is received

## handlers/commands.go
Parses `/command` messages and applies them before messages are embedded or sent to the LLM.

## handlers/settings.go
Loads and stores per-chat settings such as the context limits and system prompt.

## handlers/images.go
Fetches image attachments from the Signal API and prepares them for the vision model.

## handlers/socket.go
Establishes the socket connection to the API, listens to incoming events and calls the respective functions

## models/events.go
Contains models for the various message/event types, including quotes and attachments, and normalizes incoming messages.

## db/db.go
Contains database insertion and retrieval functions. Messages are placed into a `memory.db` sqlite database. Per-chat settings, bot message timestamps, and message attachment metadata are stored in separate tables.

## testing/
Because I have this hosted and setup on a remote server, I wanted a way to verify my changes without needing to push anything.

Allows the following commands to be run to test key functionality before pushing anything:

- `go test ./testing/... -v -run TestGenerateEmbedding`

- `go test ./testing/... -v -run TestMessageRequiringSearch`

- `go test ./testing/... -v -run TestHandleDataMessage`

Offline unit tests can be run with:

- `go test ./handlers ./llm ./models`

# Running
First you need to setup [this Signal API](https://github.com/bbernhard/signal-cli-rest-api) in json-rpc mode. This involves just running a docker command and then scanning a QR code in the Signal app.

After, copy `.env.example` to `.env` and fill in the values:

```
SERVER_URL=http://localhost:8080
SOCKET_URL=ws://localhost:8080/v1/receive/{phone_number}
GROUP_ID=the regular "id" of the target chat (can be found by going to the /v1/groups/{phone_number} endpoint)
PHONE_NUMBER=+11234567 (use your phone number and include the + and country code)
GEMINI_API_KEY= (used for embeddings)
DEEPSEEK_API_KEY= (used for chat)
EMBEDDING_MODEL=gemini-embedding-001 (can replace with desired embedding model)
DEEPSEEK_MODEL=deepseek-flash (optional, defaults to deepseek-flash)
DEEPSEEK_BASE_URL=https://api.deepseek.com (optional)
```

Only `DEEPSEEK_API_KEY` needs to be added to switch chat to DeepSeek; the model and base URL already have working defaults. Then run `go mod download` to download the dependencies and `go run .` to start the bot.

# Commands
Commands can be sent by anyone in the chat. They only affect the chat they are sent in, and settings persist across restarts.

- `/history <n>` - number of recent messages sent to the AI (default 20, max 200)
- `/relevant <n>` - number of semantically similar messages sent to the AI (default 10, max 200)
- `/system-prompt <prompt>` - set the system prompt, or view it without arguments
- `/settings` - show the current settings
- `/reset` - restore the default settings
- `/help` - list the available commands
- `/ping` - check whether the bot is online

Command messages are handled before embedding, so they are not added to the bot's conversation memory.

# Replies and Images
When the bot sends a message, it stores the returned Signal timestamp. A reply is treated as directed at the bot only when its quoted message timestamp matches one of those stored bot messages. This prevents replies to other messages sent from the same phone number from triggering the bot.

Image attachments are read from both the triggering message and the message it quotes. Quoted images are resolved from attachment metadata stored when the original message was received, with the quoted thumbnail as a fallback. Images are passed to DeepSeek as base64 data URLs.

Vision requires `DEEPSEEK_MODEL=deepseek-flash` (the default). `deepseek-v4-pro` does not support images. Supported formats are JPEG, PNG, GIF, and WebP. The Signal API must also be running in `json-rpc` mode with attachment downloads enabled, which is the default unless `JSON_RPC_IGNORE_ATTACHMENTS=true` is set.

# Improvements
## Autonomy 
Allowing the bot to respond to messages on its own. This could be handled in a couple different ways. The first would be to show the AI each message and ask it if it would like to respond. This could result in it sending too many messages though and could exceed the rate limit depending on the amount of messages. The other option is to do a random number check after each message to determine if the AI will respond or not. But that would feel less like a real person, since you could have times when you directly refer to Gemini and it still doesn't respond. 

## Customization
Per-chat commands now cover the context size and system prompt. The same approach could be used to switch the DeepSeek model per chat.

## Web Search
DeepSeek does not include the Google Search tool that Gemini had. Search could be added back with an external search API and DeepSeek tool calling.
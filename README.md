# About

Signal bot that allows Gemini to send messages in the chat by writing `@gemini`  

# Structure
## main.go
Entry point, initializes the services and retrieves data from `.env`.

## llm.go
Provides functions for interacting with the LLM and embedding messages as vectors in the memory database. 

## handlers/history.go
Provides an interface for storing and retrieving messages. When Gemini is propmted, the prompt is turned into a vector and a cosine similarity search is performed against the memory database to find similar messages to be used as context.

## handlers/events.go
Contains functions which are called depending on the event that is received

## handlers/socket.go
Establishes the socket connection to the API, listens to incoming events and calls the respective functions

## models/events.go
Contains models for the various message/event types, and a function for determining which event a message is based on the payload

## db/db.go
Contains database insertion and retrieval functions. Messages are placed into a `memory.db` sqlite database.

## testing/
Because I have this hosted and setup on a remote server, I wanted a way to verify my changes without needing to push anything.

Allows the following commands to be run to test key functionality before pushing anything:

- `go test ./testing/... -v -run TestGenerateEmbedding`

- `go test ./testing/... -v -run TestMessageRequiringSearch`

- `go test ./testing/... -v -run TestHandleDataMessage`

# Running
First you need to setup [this Signal API](https://github.com/bbernhard/signal-cli-rest-api) in json-rpc mode. This involves just running a docker command and then scanning a QR code in the Signal app.

After, create a `.env` file in this project's root directory with the following fields:

```
SERVER_URL=http://localhost:8080
SOCKET_URL=ws://localhost:8080/v1/receive/{phone_number}
GROUP_ID: the regular "id" of the target chat (can be found by going to the /v1/groups/{phone_number} endpoint)
PHONE_NUMBER=+11234567 (use your phone number and include the + and country code)
GEMINI_API_KEY=
GEMINI_ENDPOINT=https://generativelanguage.googleapis.com/v1beta/models/gemini-flash-latest:generateContent (can replace with desired model)
```

Then run `go mod download` to download the dependencies and `go run .` to start the bot.

# Improvements
## Images
Right now, I don't think it's able to see images. I don't know how this would work though. First, we'd need to get the image itself from the message, and then pass that to a vision model, and get the description of the image from that. However, I don't know if it's possible to get the image or if this approach of getting a description would work well.

## Replies
On Instagram, Meta AI will automatically reply if you reply to one of its messages. Also if you reply to a message and include `@gemini`, it should instead use the message you replied to as the "prompt message". This would be convenient, but with the bot using my phone number it means that I would likely need to record all of the bot's messages and check the replied message against those. I also don't know if the API differentiates between replies and regular messages or not.

## Autonomy 
Allowing the bot to respond to messages on its own. This could be handled in a couple different ways. The first would be to show the AI each message and ask it if it would like to respond. This could result in it sending too many messages though and could exceed the rate limit depending on the amount of messages. The other option is to do a random number check after each message to determine if the AI will respond or not. But that would feel less like a real person, since you could have times when you directly refer to Gemini and it still doesn't respond. 

## Customization
Allow for customizing certain properties with chat commands. For example updating the context window size, and the "system prompt". 
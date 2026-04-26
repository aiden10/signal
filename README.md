# About

Signal bot that allows Gemini to send messages in the chat by writing `@gemini`

# Structure

The "handlers" are also more like services so maybe I should rename them.

## main.go
Entry point, initializes the services and retrieves data from .env. Currently also contains the signal client struct and function for sending messages. These two things should be moved.

## llm.go
Provides functions for interacting with the LLM 

## handlers/history.go
Provides an interface for storing and retrieving messages

## handlers/events.go
Contains functions which are called depending on the event that is received

## handlers/socket.go
Establishes the socket connection to the API, listens to incoming events and calls the respective functions

## models/events.go
Contains models for the various message/event types, and a function for determining which event a message is based on the payload

# Issues
Currently listens to messages in all chats and only sends to one chat. Need to store both internal and external IDs. 

For some reason, the group ID that is sent along the event cannot be used when sending messages to that group because you need the external ID instead. To get around this, I only ever send to one chat and listen to all, but ideally I would send to the chat where the gemini message was sent from. So I need a way to get the external ID from the internal one. 

What I can do is fetch localhost:8080/v1/groups/+{phone_number} whenever a message containing @gemini is received and then iterate over each group to find the corresponding external ID and store it in a map. Then, I can pass that ID to the SendGroupMessage function.

# Improvements
## Memory
Memory is currently just in the form of providing the n most recent messages in each prompt. This isn't ideal because it means older details will always be lost, and it's pretty limited. Increasing n would help, but also result in more tokens being used. The context window size is also message based rather than character based.

There are three main solutions. 
- vector embeddings
- keyword message retrieval
- summary

### Vector Embeddings
When a message is received, use an embedding model to turn the message into a vector, then store this vector in a database. When Gemini is prompted, turn that message into a vector, and then search the database for the closest vectors and include their original messages in the prompt which is sent to Gemini.

### Keyword Message Retrieval
Requires a dict mapping words to message indices, and a list of messages. When a message is received, store it in a list and save its index. Then, iterate over each word in the message and do `dict[word] = message_index`.

Now, when a prompt is received, go over each word in the prompt and look them up in the dict, and get all the "relevant" messages and add those to the prompt.

The issue with this though is that there could be a lot of irrelevant messages for common words like "the" or "you", so you need a way to determine the keywords. Another issue is that this is only memory for words related to the prompt message and storing only a single message for a keyword might not be enough. The dict could be updated to store the previous n messages, but it would result in more tokens being used. It might also be better to do this process not only on the singular prompt message but also the past n messages. 

### Summary
This is the simplest option. Instead of sending all the previous messages, we could periodically feed the LLM previous messages and ask it to generate a summary. Subsequently, we would then use that summary along with some recent messages and get another updated summary. 

In the prompts, this summary would be included instead of the previous messages. 

## Images
Right now, I don't think it's able to see images. I don't know how this would work though. First, we'd need to get the image itself from the message, and then pass that to a vision model, and get the description of the image from that. However, I don't know if it's possible to get the image or if this approach of getting a description would work well.

## Replies
On Instagram, Meta AI will automatically reply if you reply to one of its messages. Also if you reply to a message and include @gemini, it should instead use the message you replied to as the "prompt message". This would be convenient, but with the bot using my phone number it means that I would likely need to record all of the bot's messages and check the replied message against those. I also don't know if the API differentiates between replies and regular messages or not.

## Autonomy 
Allowing the bot to respond to messages on its own. This could be handled in a couple different ways. The first would be to show the AI each message and ask it if it would like to respond. This could result in it sending too many messages though and could exceed the rate limit depending on the amount of messages. The other option is to do a random number check after each message to determine if the AI will respond or not. But that would feel less like a real person, since you could have times when you directly refer to Gemini and it still doesn't respond. 

## Customization
Allow for customizing certain properties with chat commands. For example updating the context window size, and the "system prompt". 
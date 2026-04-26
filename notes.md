
When a message in specific group chat is received, check the message contents. If it contains a certain string, i.e "@gemini", then send a request to gemini along with the past n messages, and then send the response to the group chat. Since I can't get the previous n messages easily, I need to save them as they come in myself. 

# Structure
- SocketHandler, establishes socket connection, depending on the message type, calls a different function
- EventHandler, defines functions for different messages
- HistoryHandler, functions for reading previous messages and recording messages
- Events, define the different events


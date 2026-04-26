
# Issues
Currently listens to messages in all chats and only sends to one chat. Need to store both internal and external IDs. 

For some reason, the group ID that is sent along the event cannot be used when sending messages to that group because you need the external ID instead. To get around this, I only ever send to one chat and listen to all, but ideally I would send to the chat where the gemini message was sent from. So I need a way to get the external ID from the internal one. 

What I can do is fetch localhost:8080/v1/groups/+16479234271 whenever a message containing @gemini is received and then iterate over each group to find the corresponding external ID and store it in a map. Then, I can pass that ID to the SendGroupMessage function.
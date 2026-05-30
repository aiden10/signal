
# Vector Embedding

When a message is received, turn it into a vector using the Gemini embedding API. Store the vector in an SQLite database along with the original message. 
Table schema:
```
CREATE TABLE memory (
    id INTEGER PRIMARY KEY,
    group_id TEXT,
    content TEXT,
    vector BLOB
);
```

And then the "GetContext" function in history.go will retrieve all vectors, load them into memory, and then I do a cosine similarity check over each one, comparing them with the prompt message. Alternatively, I can use the sqlite-vec extension which will perform this search must faster as a built-in SQL function.

Once I have the closest n vectors, I can then retrieve the corresponding texts, and add those to the prompt which is sent to Gemini.

# Commands
ssh -L 8080:localhost:8080 -i "ssh.key" ubuntu@aiden.photo

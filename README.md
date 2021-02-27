# simple-chat-app

This is a simple instant messaging application for learning purpose.

## Getting Started

### Deploy the chat server

```
$ git clone https://github.com/briansp8210/simple-chat-app.git
$ cd simple-chat-app
$ docker-compose up -d
```

### Run the chat client

```
$ cd cmd/client
$ go run main.go
```

Currently the client only invokes the implemented services for testing purpose. More features and interactive user interface will come soon.

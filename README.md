# NetCat / TCP-Chat

NetCat / TCP-Chat is a Go TCP server for a terminal-based group chat.

The server listens on a port, and clients connect from another terminal with
`nc`:

```bash
nc localhost 2525
```

This project is for an Ubuntu/Linux school environment. The main commands in
this README use Linux/bash syntax.

## Current State

The project is currently in **Milestone 5**:

* Go module setup.
* Running a TCP server.
* Welcome banner and non-empty name validation logic.
* Multi-client support with `ChatRoom` management.
* Message history and broadcasting.
* Integration tests for end-to-end validation.
* Capacity limit enforcement (10 clients).

Final audit hardening and code review are in progress.

## Requirements

For Ubuntu/Linux:

* Go
* Bash
* Netcat, available as `nc`

Install them with:

```bash
sudo apt update
sudo apt install -y golang-go netcat-openbsd
```

Check that the commands are available:

```bash
go version
nc -h
```


## Build

From the project folder, build the program:

```bash
go build -o TCPChat .
```

This creates an executable named `TCPChat`.

On Ubuntu/Linux or WSL, run it with:

```bash
./TCPChat
```

## Run The Server

Run with the default port `8989`:

```bash
./TCPChat
```

Expected output:

```text
Listening on the port :8989
```

Run with a custom port:

```bash
./TCPChat 2525
```

Expected output:

```text
Listening on the port :2525
```

Invalid usage:

```bash
./TCPChat 2525 localhost
```

Expected output:

```text
[USAGE]: ./TCPChat $port
```

## Connect A Client

Open a second terminal and connect with:

```bash
nc localhost 2525
```

For now, the connection will close because the chat session code is not finished
yet. The next implementation steps are the welcome message, name prompt,
multiple clients, and message broadcasting.

## Run Tests

Run all Go tests:

```bash
go test ./...
```

## Required Final Chat Behavior

When the implementation is complete, the app must support:

* Multiple clients connected to one TCP server.
* Maximum 10 simultaneous clients.
* Welcome banner and `[ENTER YOUR NAME]:` prompt.
* Non-empty client names.
* Timestamped chat messages.
* No broadcast for empty messages.
* Message history sent to newly joined clients.
* Join notifications.
* Leave notifications.
* Remaining clients stay connected when one client disconnects.
* Goroutines.
* Channels or mutexes.

Expected chat message format:

```text
[2020-01-20 15:48:41][client.name]:client.message
```

## Project Files

Important files:

* `main.go`: program startup, CLI argument parsing, and TCP listener.
* `main_test.go`: tests for the port parsing behavior.
* `go.mod`: Go module definition.
* `PRD.md`: project requirements.
* `docs/architecture.md`: planned architecture.
* `excersise.txt`: exercise summary.

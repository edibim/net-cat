# Project Context - NetCat / TCP-Chat

This repository is for the NetCat exercise: a Go TCP group chat server compatible with terminal clients such as `nc`.

The project targets a Linux/bash workflow. Use bash-style examples and prefer the audit binary command `./TCPChat` after building with `go build -o TCPChat .`.

## Current Phase

Milestone 5: Audit Hardening.

The mandatory assignment requirements have been normalized into:

* `PRD.md`
* `excersise.txt`
* `audit-questions.txt`
* `docs/architecture.md`

## Product Summary

The application must run as `./TCPChat [port]`.

If no port is provided, it listens on port `8989`. If one port is provided, it listens on that port. If extra arguments are provided, it prints:

```text
[USAGE]: ./TCPChat $port
```

Clients connect with:

```bash
nc <host> <port>
```

The server prompts each client for a non-empty name, then places them into a shared TCP chat room.

## Mandatory Features

* TCP server.
* Multiple clients.
* Maximum 10 simultaneous clients.
* Welcome banner and name prompt.
* Non-empty client names.
* Timestamped messages.
* Empty messages ignored.
* Message history for new clients.
* Join notifications.
* Leave notifications.
* Remaining clients stay connected after one client disconnects.
* Goroutines.
* Channels or mutexes.
* Allowed package constraints from the exercise.

## Design Direction

Use a simple mandatory-first design:

* `main` handles CLI parsing and startup.
* `Server` accepts connections.
* `Client` owns one connection session.
* `ChatRoom` owns shared clients/history state.
* `sync.Mutex` protects shared state.

Avoid bonus work until the mandatory audit path is stable.

Core design decisions:

* Use one Go package: `main`.
* Keep mandatory code inside the allowed package list.
* Store chat messages in history, not join/leave system events.
* Broadcast formatted chat messages to every connected client, including the sender.
* Allow duplicate non-empty names initially because uniqueness is not required by the exercise.
* Protect shared room state with `sync.Mutex`.
* Copy broadcast targets while locked, then write to TCP connections after unlocking.
* Reject the 11th client with a short capacity message and close only that connection.

## Current Implementation State

Milestone 5 is in progress:

* `go.mod` defines the local module.
* `main.go` validates CLI arguments, chooses the default/custom port, starts a TCP listener, and prints the required listening message.
* `banner.go` contains the mandatory ASCII art.
* `client.go` handles connection sessions, name validation, and the message loop.
* `chatroom.go` manages shared state, history, and broadcasting with a `sync.Mutex`.
* `integration_test.go` provides end-to-end verification of broadcasting and the 10-client limit.

Primary Linux/bash workflow:

```bash
go build -o TCPChat .
./TCPChat 2525
nc localhost 2525
go test ./...
```

## Design Milestones

1. Startup: CLI parsing and listener startup.
2. Client session: welcome banner, name prompt, non-empty name validation.
3. Chat room: client registry, max 10 clients, history, join/leave notifications.
4. Messaging: timestamp formatting, empty message filtering, broadcast.
5. Audit hardening: integration tests, manual `nc` checks, allowed package review, README update.

# NetCat / TCP-Chat Architecture

## Design Goal

Build the mandatory exercise first, but design it so larger later changes can be added without breaking the audit path.

The target operating environment is Linux with bash. Documentation, build commands, audit commands, and examples should use bash syntax and the `./TCPChat` executable.

The mandatory behavior must always remain stable:

* `./TCPChat` listens on port `8989`;
* `./TCPChat <port>` listens on the provided port;
* invalid argument count prints `[USAGE]: ./TCPChat $port`;
* clients connect with `nc`;
* each client receives the welcome banner and name prompt;
* non-empty names are required;
* up to 10 clients can chat together;
* valid messages are timestamped, stored in history, and broadcast;
* empty messages are ignored;
* join and leave notifications are sent;
* one disconnected client must not disconnect the others.

## Design Principles

1. Keep the mandatory server path simple and auditable.
2. Isolate decisions that may change later: message formatting, room membership, event broadcasting, and client command parsing.
3. Use only the exercise's allowed packages in mandatory code.
4. Prefer small functions that can be tested without real network connections.
5. Keep shared mutable state inside `ChatRoom`.
6. Do not introduce bonus features until the mandatory audit flow passes.

## Package And File Plan

Use one Go package: `main`.

This keeps the exercise simple and avoids package boundary overhead while still allowing clear files.

```text
.
+-- go.mod
+-- main.go
+-- main_test.go
+-- config.go
+-- server.go
+-- client.go
+-- chatroom.go
+-- message.go
+-- banner.go
+-- *_test.go
`-- docs/
```

Current files:

* `go.mod`
* `main.go`
* `main_test.go`

Planned files:

* `config.go`: constants and CLI parsing.
* `server.go`: listener ownership and accept loop.
* `client.go`: one connected user session.
* `chatroom.go`: shared room state, capacity, history, broadcast.
* `message.go`: timestamp/message formatting and input normalization.
* `banner.go`: TCP-Chat welcome banner.

If the project remains small, some planned files can stay combined. The important part is preserving the design boundaries.

## Linux/Bash Workflow

Build the binary:

```bash
go build -o TCPChat .
```

Run on the default port:

```bash
./TCPChat
```

Run on a custom port:

```bash
./TCPChat 2525
```

Connect clients from separate terminals:

```bash
nc localhost 2525
```

Run tests:

```bash
go test ./...
```

All manual audit instructions should use this Linux/bash workflow.

## Core Ownership Model

### `main`

Owns process startup only.

Responsibilities:

* parse arguments;
* print usage on invalid arguments;
* create the chat room;
* create and start the server;
* print fatal startup errors.

`main` should not know how clients are handled.

### `Server`

Owns the TCP listener and accept loop.

Responsibilities:

* listen on `:<port>`;
* print `Listening on the port :<port>`;
* accept incoming TCP connections;
* hand accepted connections to a client session goroutine;
* keep accepting after a single connection error.

Suggested shape:

```go
type Server struct {
    port string
    room *ChatRoom
}
```

Main methods:

```go
func NewServer(port string, room *ChatRoom) *Server
func (s *Server) ListenAndServe() error
func (s *Server) handleConn(conn net.Conn)
```

### `Client`

Owns one TCP connection.

Responsibilities:

* send welcome banner;
* prompt for a non-empty name;
* read messages line by line;
* submit valid messages to the room;
* trigger leave cleanup on disconnect.

Suggested shape:

```go
type Client struct {
    name string
    conn net.Conn
    room *ChatRoom
}
```

Main methods:

```go
func NewClient(conn net.Conn, room *ChatRoom) *Client
func (c *Client) Run()
func (c *Client) promptName(reader *bufio.Reader) (string, error)
func (c *Client) Write(message string)
```

### `ChatRoom`

Owns all shared chat state.

Responsibilities:

* track connected clients;
* enforce max 10 clients;
* store message history;
* broadcast chat messages;
* broadcast join/leave notifications;
* remove clients safely.

Suggested shape:

```go
type ChatRoom struct {
    mu      sync.Mutex
    clients map[*Client]bool
    history []string
    limit   int
}
```

Main methods:

```go
func NewChatRoom(limit int) *ChatRoom
func (r *ChatRoom) CanJoin() bool
func (r *ChatRoom) Join(client *Client) []string
func (r *ChatRoom) Leave(client *Client)
func (r *ChatRoom) BroadcastMessage(sender *Client, text string)
func (r *ChatRoom) BroadcastSystem(text string, except *Client)
```

## Runtime Flow

1. `main` calls `parsePort(os.Args[1:])`.
2. Invalid arguments print:

```text
[USAGE]: ./TCPChat $port
```

3. `main` creates `ChatRoom` with limit `10`.
4. `main` creates `Server`.
5. `Server.ListenAndServe` starts `net.Listen("tcp", ":"+port)`.
6. Server prints:

```text
Listening on the port :2525
```

7. Server accepts connections forever.
8. For each connection, server starts `go client.Run()`.
9. Client sends the welcome banner and `[ENTER YOUR NAME]:`.
10. Client reads name input until it receives a non-empty name or the connection closes.
11. If the room is full, the client receives a clear capacity message and the connection closes.
12. If accepted, the client joins the room.
13. The new client receives stored chat history.
14. Existing clients receive:

```text
<name> has joined our chat...
```

15. The client read loop starts.
16. Empty messages are ignored.
17. Valid messages are formatted, stored, and broadcast.
18. On disconnect, the client is removed from the room.
19. Remaining clients receive:

```text
<name> has left our chat...
```

## Message Contract

### Chat Message Format

```text
[YYYY-MM-DD HH:MM:SS][name]:message
```

Example:

```text
[2020-01-20 15:48:41][Yenlik]:hello
```

Use Go time format:

```go
"2006-01-02 15:04:05"
```

### Message History

Store only user chat messages in history.

Do not store:

* welcome banners;
* name prompts;
* join notifications;
* leave notifications;
* empty messages.

Reason: the exercise requires new clients to receive previous messages sent to the chat. System events are not required as history.

### Empty Message Rule

Normalize client input by removing line endings and trimming whitespace.

If the result is empty, ignore it.

### Sender Echo Decision

Broadcast formatted chat messages to every connected client, including the sender.

Reason: this makes the server the source of truth and ensures every client sees the same final formatted message. It is safer for audits than relying on terminal echo behavior.

## Concurrency Design

Use goroutines:

* one goroutine for the accept loop in the main server flow;
* one goroutine per connected client.

Use `sync.Mutex` inside `ChatRoom`:

* protect `clients`;
* protect `history`;
* protect capacity checks.

Preferred broadcast pattern:

1. Lock the room.
2. Update state.
3. Copy target clients and messages.
4. Unlock the room.
5. Write to network connections outside the lock.

This avoids one slow client blocking all room state changes.

## Error Handling

Expected errors should not crash the server:

* client disconnect;
* failed read from client;
* failed write to one client;
* temporary accept failure.

Startup errors can return from `ListenAndServe`:

* invalid/unavailable port;
* listener creation failure.

When a client connection fails, clean up only that client.

## Capacity Design

Maximum connected clients: `10`.

When the room is full:

1. accept the TCP connection;
2. write a short message such as:

```text
Chat room is full. Try again later.
```

3. close that connection;
4. keep existing clients connected.

This satisfies "control connections quantity" while keeping behavior clear.

## Name Design

Mandatory behavior:

* names must be non-empty after trimming whitespace.

Initial decision:

* duplicate names are allowed.

Reason: the exercise does not require unique names. Rejecting duplicates can create extra edge cases that do not improve audit success.

Potential later change:

* add duplicate-name rejection if needed, but keep it behind a small `ChatRoom.NameTaken(name string)` helper.

## Prompt And Terminal UX

The mandatory version can keep terminal UX minimal.

After each server message, clients may need to see their own prompt again. A future polish step can introduce a prompt helper:

```text
[2020-01-20 15:48:41][Yenlik]:
```

This is visible in the assignment examples, but the audit questions focus on message delivery and formatting. Implement the core chat first, then improve prompt redraw if needed.

## Test Strategy

### Unit Tests First

Test without network where possible:

* CLI parsing;
* message formatting;
* empty message normalization;
* capacity checks;
* join adds client;
* leave removes client;
* history stores chat messages;
* history excludes system events.

### Integration Tests Second

Use `net.Listen` and `net.Dial` for local TCP tests:

* client receives welcome banner;
* client receives name prompt;
* two clients can join;
* a message from one client is received by another;
* a late client receives history;
* full room rejects extra connection.

### Manual Audit Last

Use real terminals and `nc`:

* `go build -o TCPChat .`;
* `./TCPChat`;
* `./TCPChat 2525`;
* `./TCPChat 2525 localhost`;
* two-client chat;
* three-client chat;
* disconnect behavior;
* cross-machine behavior if available.

## Implementation Milestones

### Milestone 1 - Startup

Status: started.

Scope:

* Go module.
* CLI parsing.
* TCP listener startup.
* CLI parsing tests.

Next:

* move config constants into `config.go` only if the file split helps.
* add welcome banner and name prompt.

### Milestone 2 - Client Session

Scope:

* welcome banner;
* name prompt;
* non-empty name validation;
* one goroutine per client;
* basic client disconnect cleanup.

### Milestone 3 - Chat Room

Scope:

* `ChatRoom` struct;
* client registry;
* max 10 clients;
* join/leave notifications;
* history storage.

### Milestone 4 - Messaging

Scope:

* message normalization;
* timestamp formatting;
* empty message filtering;
* broadcast to all clients;
* write errors handled per client.

### Milestone 5 - Audit Hardening

Scope:

* integration tests;
* manual audit checklist;
* race-prone areas reviewed;
* allowed packages checked;
* README updated with run/test instructions.

## Safe Future Big Changes

These can be added later while preserving the exercise:

* better prompt redraw after broadcasts;
* duplicate-name rejection;
* server log file as bonus;
* name change command as bonus;
* multiple rooms as bonus;
* terminal UI as bonus, only using the allowed bonus package;
* command parser for slash commands, as long as normal chat still works.

The rule for all big changes: keep the mandatory `nc` chat path unchanged and passing before and after the change.

## Risk Register

* **Writes while holding lock:** can freeze room state if one client is slow.
* **Prompt redraw:** can make terminal output visually messy, even when logic works.
* **Sender echo ambiguity:** solved by broadcasting to sender too.
* **History ambiguity:** solved by storing chat messages only.
* **Allowed packages:** must be checked before final submission.
* **Test ports:** integration tests should use dynamic local ports where possible.

## Final Mandatory Definition

The project is ready for audit when:

* `go test ./...` passes;
* manual `nc` chat works with at least three clients;
* a new client receives previous messages;
* a disconnect leaves other clients connected;
* the 11th connection is controlled;
* invalid CLI usage prints the exact usage string;
* only allowed packages are used in mandatory code.

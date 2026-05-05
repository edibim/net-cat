# Product Requirements Document (PRD) - NetCat / TCP-Chat

## 1. Problem Statement

Build a Go implementation of a simplified NetCat-style group chat. The program runs as a TCP server, accepts multiple terminal clients, and lets them exchange chat messages in real time.

The project must demonstrate TCP networking, goroutines, synchronization, error handling, and clean Go structure while matching the required audit behavior.

The target execution environment is Linux/bash. The expected audit binary is built with `go build -o TCPChat .` and run as `./TCPChat`.

## 2. User / Use Case

* **Primary User:** A terminal user connecting with `nc <host> <port>`.
* **Secondary User:** A developer/auditor running `./TCPChat` and checking the project against the assignment requirements.
* **Use Case:** Start a TCP chat server, connect multiple clients, require each client to provide a name, broadcast messages, preserve chat history, and notify connected users when clients join or leave.

## 3. Architecture Approach

Use a single TCP server process with a central chat room state.

Recommended high-level components:

* **CLI/bootstrap:** validates arguments, chooses the port, starts the TCP listener.
* **Server:** accepts incoming TCP connections and starts one goroutine per client.
* **Client session:** handles welcome banner, name prompt, incoming messages, disconnects, and outgoing writes.
* **Chat room/state:** tracks connected clients, message history, and broadcasts.
* **Synchronization:** uses either `sync.Mutex` or channels to protect shared state.

Initial design preference: use a `ChatRoom` struct guarded by a mutex. This keeps the first version simple, auditable, and close to the assignment constraints.

## 4. CLI Contract

* **Build command:** `go build -o TCPChat .`
* **Command:** `./TCPChat [port]`
* **Default port:** `8989`
* **Valid usage examples:**
  * `./TCPChat`
  * `./TCPChat 2525`
* **Invalid usage example:**
  * `./TCPChat 2525 localhost`
* **Invalid usage output:**

```text
[USAGE]: ./TCPChat $port
```

* **Startup output:**

```text
Listening on the port :8989
Listening on the port :2525
```

## 5. Functional Requirements

1. The project must be written in Go.
2. The server must use TCP.
3. The server must accept multiple clients in a 1-to-many relationship.
4. The server must support a maximum of 10 simultaneous client connections.
5. Each accepted client must receive the TCP-Chat welcome banner and be prompted for a name.
6. A client name must be non-empty after trimming whitespace.
7. Clients must be able to send messages to the chat.
8. Empty messages must not be broadcast.
9. Each chat message must include the send timestamp and sender name:

```text
[2020-01-20 15:48:41][client.name]:client.message
```

10. A newly joined client must receive previous chat messages.
11. Existing clients must be notified when a client joins.
12. Existing clients must be notified when a client leaves.
13. All connected clients must receive messages sent by other clients.
14. If one client disconnects, all other clients must stay connected.
15. The server must handle expected client-side and server-side errors without crashing.
16. The implementation must use goroutines.
17. The implementation must use channels or mutexes.
18. The project must use only the allowed packages from the exercise unless a later design decision explicitly narrows or expands the scope.

## 6. Allowed Packages

Assignment allowed packages:

* `io`
* `log`
* `os`
* `fmt`
* `net`
* `sync`
* `time`
* `bufio`
* `errors`
* `strings`
* `reflect`

Design note: prefer the smallest useful subset. Avoid adding external dependencies for the mandatory version.

## 7. Non-Goals

The mandatory version does not include:

* UDP support.
* UNIX-domain sockets.
* File transfer.
* Terminal UI.
* Multiple chat rooms.
* Persistent logs.
* Name changes.
* Advanced NetCat flags.

These can be considered bonus work after the required audit behavior is stable.

## 8. Acceptance Criteria

### 8.1 Mandatory Audit Cases

The project is acceptable when:

* `go build -o TCPChat .` creates the expected Linux executable.
* `./TCPChat` listens on port `8989`.
* `./TCPChat 2525` listens on port `2525`.
* `./TCPChat 2525 localhost` prints `[USAGE]: ./TCPChat $port`.
* Two or more `nc` clients can connect successfully.
* Each client receives the welcome banner and name prompt.
* Non-empty names are accepted.
* Join notifications are sent to already connected clients.
* Messages are broadcast to all other connected clients.
* New clients receive message history.
* Messages include timestamp and sender name.
* Empty messages are ignored.
* Disconnecting one client does not disconnect the others.
* Leave notifications are sent to remaining clients.
* The implementation uses goroutines and synchronization.
* The project compiles and runs without panics during normal audit flows.

### 8.2 Additional Golden Tests

Useful extra cases:

* Reject or re-prompt for an empty name.
* Reject the 11th simultaneous client with a clear message and close only that connection.
* Keep history order stable under multiple clients.
* Handle `Ctrl+C` or closed `nc` sessions cleanly.
* Handle messages with leading/trailing spaces consistently.
* Avoid broadcasting partial state if a client disconnects mid-message.

## 9. Implementation Approach

Recommended first-pass implementation:

1. Add a Go module or simple Go package structure if needed.
2. Implement CLI argument validation and listener startup.
3. Define the welcome banner as a constant.
4. Define `Client` and `ChatRoom` structs.
5. Implement connection acceptance with one goroutine per client.
6. Prompt for a non-empty name.
7. Add client to room state.
8. Send history to the new client.
9. Broadcast join, message, and leave events.
10. Protect shared clients/history state with a mutex.
11. Add focused tests for formatting and room behavior where possible.
12. Manually verify real TCP behavior with `nc`.

## 10. Milestones

### Milestone 1 - Skeleton

* CLI validation.
* TCP listener.
* Welcome banner and name prompt.

### Milestone 2 - Core Chat

* Client tracking.
* Message formatting.
* Broadcast behavior.
* Join/leave notifications.

### Milestone 3 - Audit Hardening

* Max 10 clients.
* Message history for new clients.
* Empty message filtering.
* Graceful disconnects.

### Milestone 4 - Verification

* Unit tests for pure logic.
* Manual multi-terminal audit walkthrough.
* Clean up code structure and comments.

## 11. Risks / Open Questions

* Should duplicate client names be allowed? The exercise only requires non-empty names.
* Should the sender also receive their own formatted message from the server, or only see local terminal echo? Audit examples are mixed, so we should design deliberately.
* Should join/leave notifications be stored in history? The exercise requires previous messages, not necessarily events.
* How strict should port validation be beyond argument count?

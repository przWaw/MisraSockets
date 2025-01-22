# MISRA Token Passing Algorithm

This repository contains an implementation of the MISRA Token Passing Algorithm in Go. The algorithm is a distributed algorithm used for token-based mutual exclusion and communication in distributed systems. Algorithm allow to detect communication problem by passing two distinct tokens with only one granting permission to enter critical section.

## Overview

The MISRA Token Passing Algorithm is designed for systems where processes are arranged in a logical ring. Each process in the ring can:

- Pass a token to its neighbor.
- Ensure mutual exclusion when accessing shared resources.

This implementation is only example of implementation to present concept without any serious logic included. 

## Input Configuration

Arguments available to pass into application

| Parameter          | Flag    | Description                                       | Example Value       |
|--------------------|---------|---------------------------------------------------|---------------------|
| Listening Port     | `-l`    | The port to listen on                             | `8080`              |
| Send Address       | `-s`    | The address to send to (next process in the ring) | `127.0.0.1:8081`    |
| Start as Initiator | `-i`    | Start as the initiator (optional)                 | `true` (flag only)  |
| Ping Lost          | `-p`    | Ping lost probability (0 to 1, optional)          | `0.1`               |

### Example Usage

```bash
# Start as a listener on port 8080 and send to 127.0.0.1:8081
$ go run main.go -l 8080 -s 127.0.0.1:8081

# Start as an initiator on port 8080 and send to 127.0.0.1:8081
$ go run main.go -l 8080 -s 127.0.0.1:8081 -i

# Start as a listener with a 10% ping loss probability
$ go run main.go -l 8080 -s 127.0.0.1:8081 -p 0.1
```

It's preferred to start listener node first. It's possible to chain more than two nodes into circle, but every time initiator should be started at the end.
# Kronos C2
![](imgs/B7EF2F66-EC50-447B-8010-B79928404484.png)

> **Work in Progress** — not ready for production use.

Kronos is a command-and-control framework written in Go and c++. It consists of a teamserver, an operator CLI client, and a Windows agent.

![Kronos CLI](imgs/2026-04-22_04-08.png)

## Components

- **server** — teamserver exposing a REST API and SSE event stream for operators, with SQLite-backed agent and task tracking
- **client** — operator CLI for managing agents, listeners, and tasking
- **agent** — Windows implant (WIP)

## Features

- JWT-authenticated operator sessions
- HTTP/HTTPS listeners
- Fully encrypted communications
- Agent registration and check-in
- Real-time event streaming to connected operators
- SQLite persistence for agents, tasks, and listeners

## Status

This project is under active development. Many features are incomplete or subject to change.

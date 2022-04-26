# ppcserver

[![Go Reference](https://pkg.go.dev/badge/github.com/pom-pom-crafts/ppcserver.svg)](https://pkg.go.dev/github.com/pom-pom-crafts/ppcserver)
[![Go Report Card](https://goreportcard.com/badge/github.com/pom-pom-crafts/ppcserver)](https://goreportcard.com/report/github.com/pom-pom-crafts/ppcserver)

A Go multiplayer game server framework featured real-time WebSocket communication, Stateful Room, and distributed scaling via NATS.

## Documentation
- [Drawing example](./examples/drawing/README.md)

## Design Concept
- Bound with minimal package dependencies so that you can choose the ones according to your actual needs.
- Easy to connect through plain WebSocket API with no custom client library required.
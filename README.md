# ppcserver
Go Game Server communicates using WebSocket, groups players with Stateful Rooms, and is scalable through NATS.

## Design Concept
- Bound with minimal package dependencies so that you can choose the ones according to your actual needs.
- Easy to connect through plain WebSocket API with no custom client library required.
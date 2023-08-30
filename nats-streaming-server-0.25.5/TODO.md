
- [ ] Retry limits?
- [X] Server Store Limits (time, msgs, byte)
- [X] Change time to deltas
- [X] Server heartbeat, release dead clients.
- [X] Require clientID for published messages, error if not registered.
- [X] Check for need of ackMap (out of order re-delivery to queue subscribers).
- [X] Redelivered Flag for Msg.
- [X] Queue Subscribers
- [X] Durable Subscribers (survive reconnect, etc)
- [X] Start Positions on Subscribers
- [X] Ack for delivered just Reply? No need on ConnectedResponse?
- [X] PublishWithReply, or option.
- [X] Data Races in Server.
- [X] Manual Ack?
# voskos-rtc-sfu
`voskos Media [ Webrtc SFU ] server` is novel approach of implementing SFU media servers, which can act as a simple, single unit on the servers. 

The aim is - "*Simple Deployment*" for all the complex inherited functionalities of a generic SFU server.

Please note that the project is in development stage, we need all of your support in terms of Development, Documentation, Infrastructure for testing, and referrals for contribution :smiley:

[![Maintainer](https://img.shields.io/badge/Maintainer-bmonikraj-blue)](https://github.com/bmonikraj)
[![Maintainer](https://img.shields.io/badge/Maintainer-preetamnahak007-blue)](https://github.com/preetamnahak007)

---

## The current implementation consists of the server running ( using `go run main.go` in the root dir of the project) with functional aspects such as :- 

- [x] Multiple rooms(sessions) on a single server
- [x] Multiple users in a single session 
- [x] Audio/Video/Screen Share in a session

--- 

## To Do - Future Work 

- [ ] Migarte to CoTurn servers from Public STUN/TURN servers
- [ ] Implement Data Channels for Chats 
- [ ] Implementation of a secure authorization layer on top of signaling to generate "join token"
- [ ] Stress testing on multi-session::multi-user scenarios to observer the behaviour of mutex locks for re-negotiations
- [ ] Load testing to find the optimal hardware requirements
- [ ] Implementation of server stats endpoints 
- [ ] Integrate CI CD pipelines for build, test and release management

--- 

## Ideas ? Suggestions ? Contribution ? You are most welcome! 

**Just fork, raise a PR and that's it! :smiley:**

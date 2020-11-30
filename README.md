# voskos-rtc-sfu
voskos Media [ Webrtc SFU ] server

**Testinng Procedure is mentioned [here](https://github.com/voskos/voskos-testing-client/blob/main/README.md)**

### To Do List Development 

 - [ ] `log.Fatal()` Exits the program if Fatal error is occurred. Make sure the usage is correct 
 - [ ] In 'go.mod' file, there are *2* pion library versions available (v2 and v3). This would create version conflicts in later stage. Remove the **unwanted** version
 - [ ] Put the constants in `constant.go` like Actions, Message Types, etc
 - [ ] Since new logger is being used, put proper log formatting, no need to use directives like `%s`
 - [ ] Is the `router` struct's object maintained as *Global* scope ? Make sure it does, else `Lock` won't work as expected 
 - [ ] With the current code,  `lock` is acquired in Init, but released only when the "video" track is set. What if client is not giving "Video" track, gives only "Audio" track ? 
 - [ ] When the PC unlock happens ? Can find only when the client answer is coming for SDP - Does this means for 
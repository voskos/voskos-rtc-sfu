# voskos-testing-client-browser

Testing procedure :


1. Install node
2. `npm install --global http-server`
3. `git clone https://github.com/voskos/voskos-rtc-sfu/`
4. `go run main.go`
5. `cd /test/browser`
6. `http-server [path] [options]`. Visit [this](https://www.npmjs.com/package/http-server) page.
7. Open http://localhost:8080 in multiple browser tabs. 
8. Enter unique user ids and room ids and click on Call button to start simulation.


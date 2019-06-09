# remote-job-seeker

I wrote this for my friend [@jmfulgham](https://github.com/jmfulgham).

It's a simple Go program that will make HTTP requests to several JSON and RSS feeds, get the response data and parse that into a JSON file. The output JSON is an array of structs, each representing a single remote job opportunity.

It's not robust or tested. I spent a few hours on it just to try to work with Go's concurrency in a real world scenario.

The results are pretty impressive!

Presently it's requesting data from 3 URLs and encoding all of the results to a file in 0.9 - 1.2 seconds.

To run it:
- [install Go](https://golang.org/doc/install#install)
- clone this repo
- cd to the `remote-job-seeker` directory
- type `go run main.go`
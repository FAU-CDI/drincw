package source

// cspell:words helloworld

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
)

func ExampleReadAll_file() {
	// write "hello world" to the temporary file
	file := func() string {
		file, err := os.CreateTemp("", "")
		if err != nil {
			panic(err)
		}
		defer file.Close()
		_, err = file.WriteString("hello world")
		if err != nil {
			panic(err)
		}
		return file.Name()
	}()
	defer os.Remove(file)

	// read the local file!
	content, err := ReadAll(file)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(content))
	// Output: hello world
}

func ExampleReadAll_uri() {
	// listen on a new port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	// start a server that echoes back the url
	done := make(chan struct{})
	go func() {
		defer close(done)
		http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, r.URL.Path)
		}))
	}()

	defer func() {
		// stop accepting connections
		listener.Close()
		<-done
	}()

	// create a url to request
	url := fmt.Sprintf("http://%s/helloworld", listener.Addr().(*net.TCPAddr).AddrPort().String())

	// read the local file!
	content, err := ReadAll(url)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(content))
	// Output: /helloworld
}

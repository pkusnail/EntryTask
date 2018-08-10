package main

//usage:
//go run rpc_tcp_c.go  localhost:1234

import (
    "fmt"
    "log"
    "net/rpc"
    "os"
)

type Args struct {
    A, B string
}

type Args2 struct {
        A,B string
}

type Args3 struct {
        A,B,C string
}


type Args4 struct {
        A,B,C,D string
}

func main() {
    if len(os.Args) != 2 {
        fmt.Println("Usage: ", os.Args[0], "server:port")
        os.Exit(1)
    }
    service := os.Args[1]

    client, err := rpc.Dial("tcp", service)
    if err != nil {
        log.Fatal("dialing:", err)
    }
    // Synchronous call
    args := Args4{"what", "fuck","a","b"}
    var reply string
    err = client.Call("Query.SignUp", args, &reply)
    if err != nil {
        log.Fatal("arith error:", err)
    }
    fmt.Printf("response:%s\n",  reply)
}

package main

import (
	"fmt"
	"net"
	"net/rpc"
	remotelist "remotelist/pkg"
)

func main() {
	server := rpc.NewServer()
	remote := remotelist.NewRemoteList()

	server.Register(remote)

	l, err := net.Listen("tcp", ":5000")
	if err != nil {
		fmt.Println("Erro:", err)
		return
	}

	fmt.Println("Servidor RPC rodando na porta 5000...")

	for {
		conn, err := l.Accept()
		if err == nil {
			go server.ServeConn(conn)
		}
	}
}

package main

import (
	"fmt"
	"net/rpc"
)

func main() {
	client, _ := rpc.Dial("tcp", ":5000")

	var ok bool
	var num int
	
	client.Call("RemoteList.Append", [2]int{1, 10}, &ok)
	client.Call("RemoteList.Append", [2]int{1, 20}, &ok)
	client.Call("RemoteList.Append", [2]int{2, 15}, &ok)
	client.Call("RemoteList.Append", [2]int{1, 30}, &ok)
	client.Call("RemoteList.Append", [2]int{2, 40}, &ok)

	client.Call("RemoteList.Get", [2]int{1, 1}, &num)
	fmt.Println("Get pos 1 (lista 1) =", num)

	client.Call("RemoteList.Get", [2]int{2, 0}, &num)
	fmt.Println("Get pos 0 (lista 2) =", num)

	client.Call("RemoteList.Remove", 1, &num)
	fmt.Println("Remove =", num)

	client.Call("RemoteList.Size", 1, &num)
	fmt.Println("Tamanho =", num)
}

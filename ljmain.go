package main

import (
        "github.com/kouzdra/go-livejournal/client"
        "fmt"
        "log"
)


func main() {
	defer func () {
		switch err := recover ().(type) {
		case nil:
			break
		case client.Error:
			log.Printf("Error: %s", err.Show ())
			break
		default:
			panic (err)
		}
	} ()
	//client := &lj.Client { Url: "http://www.dreamwidth.org/interface/xmlrpc", UserName: "kouzdra", Password: "Glokaya239" }
	//client := &lj.Client { Url: "http://lj.rossia.org/interface/xmlrpc", UserName: "kouzdra", Password: "Ghjdthrf239" }
	c := &client.Client { Url: "http://www.livejournal.com/interface/xmlrpc", UserName: "kouzdra", Password: "Glokaya239" }
        //res := c.Call ("getreadpage");
        //res := c.Call ("getfriendspage");
        res := c.Call ("getdaycounts");
        //res := c.Call ("getchallenge");
	//for _, p := range res.(lj.Array) {
	for k, v := range res {
		fmt.Printf("%s=%v\n", k, v)
	}
	fmt.Println()
	//}
}

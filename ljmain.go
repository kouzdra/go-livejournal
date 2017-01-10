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
        //res, e := c.Call ("getreadpage");
        //res, e := c.Call ("getfriendspage");
        res, e := c.Call ("getdaycounts");
        //res, e := c.Call ("getchallenge");
        if e != nil {
                log.Fatal(e)
        } else {
		//for _, p := range res.(lj.Array) {
			for k, v := range res.(client.Struct) {
				fmt.Printf("%s=%v\n", k, v)
			}
			fmt.Println()
		//}
        }
}

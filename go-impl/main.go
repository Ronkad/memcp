/*
	cpdb smart clusterable distributed database working best on nvme

	https://pkelchte.wordpress.com/2013/12/31/scm-go/

*/
package main

import "fmt"

func main() {
	globalenv.vars["print"] = func (a ...scmer) scmer {
		fmt.Println(a[0].(string))
		return "ok"
	}
	globalenv.vars["nope"] = "nope"
	Repl()
}

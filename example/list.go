//
// lsof list opened file+pid
// support linux only
//

package main

import (
	"fmt"
	"os"

	"github.com/wheelcomplex/lsof"
)

func main() {
	prefix := ""
	list, err := lsof.Lsof(prefix)
	if err != nil {
		fmt.Printf("Lsof failed: %s\n", err)
		return
	}
	fmt.Printf(" ----- FILE LIST(MAP)\n")
	for key, val := range list.File2PIDsMap() {
		fmt.Printf("%s(%s): %v\n", key, val.File, val.PIDs)
	}
	list.Close()
	fmt.Printf(" -----\n")
	list, err = lsof.LsofPID(os.Getppid(), "")
	if err != nil {
		fmt.Printf("LsofPID failed: %s\n", err)
		return
	}
	fmt.Printf(" ----- FILE LIST(%d)\n", os.Getppid())
	for key, val := range list.File2PIDsMap() {
		fmt.Printf("%s(%s): %v\n", key, val.File, val.PIDs)
	}
	list.Close()
	fmt.Printf(" -----\n")
	list, err = lsof.LsofPID(os.Getpid(), "")
	if err != nil {
		fmt.Printf("LsofPID failed: %s\n", err)
		return
	}
	fmt.Printf(" ----- FILE LIST(%d)\n", os.Getpid())
	for key, val := range list.File2PIDsMap() {
		fmt.Printf("%s(%s): %v\n", key, val.File, val.PIDs)
	}
	list.Close()
	fmt.Printf(" -----\n")
}

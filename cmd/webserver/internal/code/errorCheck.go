package code

import "fmt"

func ErrCheck(err error, desc string, args ...interface{}) bool {
	if err != nil {
		if len(args) > 0 {
			fmt.Printf("%s:%v", fmt.Sprintf(desc, args...), err)
		} else {
			fmt.Printf("%s:%v", desc, err)
		}
		fmt.Printf("%s:%v", desc, err)
		return true
	}

	return false
}

package main

import "fmt"

func main() {
    var b interface{} = float64(0)
    fmt.Println(b != 0)
}

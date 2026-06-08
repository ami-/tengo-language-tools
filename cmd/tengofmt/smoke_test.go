//go:build ignore

package main

import (
    "fmt"
    "github.com/ami-/tengo-language-tools/internal/formatter"
)

func main() {
    src := []byte(`if oInfo.firstName != cInfo.firstName || oInfo.lastName != cInfo.lastName || oInfo.city != cInfo.city || oInfo.houseNumber != cInfo.houseNumber || oInfo.streetCode != cInfo.streetCode || oInfo.zipCode != cInfo.zipCode {
    return true
}

if a || b {
    return false
}

changed := oInfo.firstName != cInfo.firstName || oInfo.lastName != cInfo.lastName || oInfo.city != cInfo.city || oInfo.houseNumber != cInfo.houseNumber

for oInfo.firstName != cInfo.firstName || oInfo.lastName != cInfo.lastName || oInfo.city != cInfo.city || oInfo.houseNumber != cInfo.houseNumber {
    doSomething()
}
`)
    out, err := formatter.Format(src)
    if err != nil {
        fmt.Println("err:", err)
        return
    }
    fmt.Println(string(out))
}

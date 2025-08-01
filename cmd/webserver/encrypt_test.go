package main
import (
    "fmt"
    "github.com/mangohow/cloud-ide/pkg/utils/encrypt"
)
func main() {
    plainPassword := "12345678"
    encryptedPassword := encrypt.PasswdEncrypt(plainPassword)
    fmt.Printf("Plain: %s\n", plainPassword)
    fmt.Printf("Encrypted: %s\n", encryptedPassword)
    fmt.Printf("Length: %d\n", len(encryptedPassword))
    fmt.Printf("Verify: %t\n", encrypt.VerifyPasswd(plainPassword, encryptedPassword))
}

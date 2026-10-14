// hashlab: ทดลองแฮชรหัสผ่านด้วย bcrypt แล้วดูว่าแฮชเดียวกันเทียบกันได้ แต่รหัสเดียวกันให้แฮชไม่ซ้ำ
// ใช้: go run ./cmd/hashlab 'รหัสผ่านที่อยากลอง'
package main

import (
	"fmt"
	"os"
	"time"

	"moviehub/auth"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("ใช้: go run ./cmd/hashlab <password>")
		os.Exit(1)
	}
	pw := os.Args[1]

	start := time.Now()
	h1, err := auth.HashPassword(pw)
	if err != nil {
		fmt.Println("แฮชไม่ได้:", err)
		os.Exit(1)
	}
	fmt.Printf("แฮชครั้งที่ 1: %s (ใช้ %v)\n", h1, time.Since(start).Round(time.Millisecond))

	h2, _ := auth.HashPassword(pw)
	fmt.Printf("แฮชครั้งที่ 2: %s\n", h2)
	fmt.Println("สองแฮชเหมือนกันไหม:", h1 == h2, "(ไม่เหมือน เพราะมี salt สุ่มในแฮช)")
	fmt.Println("เทียบรหัสถูกกับแฮช 1:", auth.CheckPassword(h1, pw))
	fmt.Println("เทียบรหัสผิดกับแฮช 1:", auth.CheckPassword(h1, pw+"x"))
}

// peek: อ่านเนื้อความของตั๋วโดยไม่ต้องใช้กุญแจ เพื่อให้เห็นว่าตั๋วไม่ได้ซ่อนข้อมูล
// ใช้: go run ./cmd/peek <token>
package main

import (
	"fmt"
	"os"
	"time"

	"moviehub/auth"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("ใช้: go run ./cmd/peek <token>")
		os.Exit(1)
	}
	c, err := auth.Peek(os.Args[1])
	if err != nil {
		fmt.Println("อ่านตั๋วไม่ได้:", err)
		os.Exit(1)
	}
	fmt.Printf("เจ้าของ (member id): %d\n", c.Sub)
	fmt.Printf("บทบาท: %s\n", c.Role)
	fmt.Printf("ออกเมื่อ: %s\n", time.Unix(c.Iat, 0).Format(time.RFC3339))
	fmt.Printf("หมดอายุ: %s\n", time.Unix(c.Exp, 0).Format(time.RFC3339))
	fmt.Println("หมายเหตุ: อ่านได้โดยไม่ใช้กุญแจ แต่แก้แล้วลายเซ็นจะไม่ตรง จึงห้ามใส่ความลับลงตั๋ว")
}

package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "123456"

	// 生成密码hash
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Error generating hash: %v\n", err)
		return
	}

	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Hash: %s\n", string(hash))

	// 验证hash
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		fmt.Printf("Hash验证失败: %v\n", err)
	} else {
		fmt.Println("Hash验证成功!")
	}

	// 测试旧hash
	oldHash := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	err = bcrypt.CompareHashAndPassword([]byte(oldHash), []byte(password))
	if err != nil {
		fmt.Printf("\n旧hash验证失败: %v\n", err)
		fmt.Println("需要使用新生成的hash更新数据库")
	} else {
		fmt.Println("\n旧hash验证成功!")
	}
}

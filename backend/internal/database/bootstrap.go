package database

import "golang.org/x/crypto/bcrypt"

// HashPasswordBootstrap 生成默认管理员密码哈希。
func HashPasswordBootstrap(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

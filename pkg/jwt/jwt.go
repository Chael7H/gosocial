package jwt

import (
	"errors"
	"time"

	"github.com/spf13/viper"

	"github.com/golang-jwt/jwt/v4"
)

// CustomSecret 用于加盐的字符串
var CustomSecret = []byte("永恒世界")

// CustomClaims 自定义声明类型 并内嵌jwt.RegisteredClaims
// jwt包自带的jwt.RegisteredClaims只包含了官方字段
// 假设我们这里需要额外记录一个username字段，所以要自定义结构体
// 如果想要保存更多信息，都可以添加到这个结构体中
type CustomClaims struct {
	// 可根据需要自行添加字段
	UserID               uint64 `json:"user_id"`
	Username             string `json:"username"`
	jwt.RegisteredClaims        // 内嵌标准的声明
}

// GenToken 生成JWT
func GenToken(userID uint64, username string) (string, error) {
	// 创建一个我们自己的声明
	claims := CustomClaims{
		userID,
		username, // 自定义字段
		jwt.RegisteredClaims{ //标准字段
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(
				time.Duration(viper.GetInt("auth.jwt_expire")) * time.Hour)), //过期时间
			Issuer: "bluebell", // 签发人
		},
	}
	// 使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 使用指定的secret签名并获得完整的编码后的字符串token
	return token.SignedString(CustomSecret)
}

// ParseToken 解析JWT
func ParseToken(tokenString string) (*CustomClaims, error) {
	// 解析token
	// 如果是自定义Claim结构体则需要使用 ParseWithClaims 方法
	var cc = new(CustomClaims)
	token, err := jwt.ParseWithClaims(tokenString, cc, func(token *jwt.Token) (i interface{}, err error) {
		// 直接使用标准的Claim则可以直接使用Parse方法
		//token, err := jwt.Parse(tokenString, func(token *jwt.Token) (i interface{}, err error) {
		return CustomSecret, nil
	})
	if err != nil {
		return nil, err
	}
	// 对token对象中的Claim进行类型断言
	if token.Valid { // 校验token
		return cc, nil
	}
	return nil, errors.New("invalid token")
}

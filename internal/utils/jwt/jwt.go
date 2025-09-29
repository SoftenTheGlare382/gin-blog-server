package jwt

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

var (
	//  错误信息常量
	ErrTokenExpired     = errors.New("token 已经过期，请重新登录")  // token 过期
	ErrTokenNotValidYet = errors.New("token 无效，请重新登录")    // token 尚未生效
	ErrTokenMalFormed   = errors.New("token 不正确， 请重新登陆")  // token 格式错误
	ErrTokenInvalid     = errors.New("这不是一个 token，请重新登录") // token 无效
)

// MyClaims 结构体,自定义的Claims结构体，用于保存自定义的 payload 数据
// 通过这个结构体，可以将用户身份、角色以及标准的 JWT 元数据打包到一个 JWT 中，从而在服务端进行安全的身份验证和权限校验。
type MyClaims struct {
	UserId               int   `json:"user_id"`  // 用户id
	RoleIds              []int `json:"role_ids"` // 角色id列表
	jwt.RegisteredClaims       // jwt.RegisteredClaims,包含 JWT 的标准注册字段（如过期时间、签发者等）
}

// GenToken 生成新的JWT
// secret: 用于签名的密钥（通常是一个私钥或者密钥
// issuer: 签发者
// expireHour: Token 过期的小时数
// userId: 用户id
// roleIds: 角用户的角色ID数组
func GenToken(secret, issuer string, expireHour, userId int, roleIds []int) (string, error) {
	//创建一个 MyClaims 实例，填充JWT的Claims数据
	claims := MyClaims{
		UserId:  userId,  // 用户id
		RoleIds: roleIds, // 角色id列表
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expireHour))), // 过期时间
			Issuer:    issuer,                                                                    // 签发者
			IssuedAt:  jwt.NewNumericDate(time.Now()),                                            // 签发时间
		},
	}
	//使用 HS256 签名方法创建一个新的 JWT Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//使用指定的密钥对JWT进行签名并返回签名后的字符串
	return token.SignedString([]byte(secret))
}

// ParseToken 解析 JWT Token 并验证其合法性
// 参数解释：
// secret：用于验证签名的密钥（通常是一个私钥或公钥）
// token：要解析的 JWT Token 字符串
func ParseToken(secret, token string) (*MyClaims, error) {

	//解析token,并将解析出来的 Claims 存入 MyClaims 结构体中
	jwtToken, err := jwt.ParseWithClaims(token, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		//使用 secret 来验证 token 的签名
		return []byte(secret), nil
	})

	if err != nil {
		switch vError, ok := err.(jwt.ValidationError); ok {
		case vError.Errors&jwt.ValidationErrorMalformed != 0:
			// Token 格式错误
			return nil, ErrTokenMalFormed
		case vError.Errors&jwt.ValidationErrorExpired != 0:
			// Token 已经过期
			return nil, ErrTokenExpired
		case vError.Errors&jwt.ValidationErrorNotValidYet != 0:
			// Token 尚未生效
			return nil, ErrTokenNotValidYet
		default:
			// 其他验证错误
			return nil, ErrTokenInvalid
		}
	}
	// 判断 Token 是否有效，如果有效则返回 Claims，否则返回无效的错误
	if claims, ok := jwtToken.Claims.(*MyClaims); ok && jwtToken.Valid {
		// 返回有效的 Claims
		return claims, nil
	}
	// Token 无效，返回错误
	return nil, ErrTokenInvalid
}

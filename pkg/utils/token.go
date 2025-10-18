package utils

import (
	"fmt"
	"fsync/server/global"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func GenterateJWT(username string) (string, error) {
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(global.Configs.JWT.AccessTokenExpire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        "1",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(global.Configs.JWT.Access))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// GenerateTokenPair 生成访问令牌和刷新令牌
func GenerateTokenPair(username string) (*TokenPair, error) {
	// 生成访问令牌
	accessClaims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(global.Configs.JWT.AccessTokenExpire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        "1",
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(global.Configs.JWT.Access))
	if err != nil {
		return nil, err
	}

	// 生成刷新令牌
	refreshClaims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(global.Configs.JWT.RefreshTokenExpire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        "2",
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(global.Configs.JWT.Refresh))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}

// RefreshAccessToken 使用刷新令牌生成新的访问令牌
func RefreshAccessToken(refreshTokenString string) (*TokenPair, error) {
	// 解析刷新令牌
	token, err := jwt.ParseWithClaims(refreshTokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 确保签名方法是 HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(global.Configs.JWT.Refresh), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// 使用刷新令牌中的用户名生成新的访问令牌
		newAccessToken, err := GenterateJWT(claims.Username)
		if err != nil {
			return nil, fmt.Errorf("failed to generate new access token: %w", err)
		}

		// 检查刷新令牌是否即将过期，如果需要则生成新的刷新令牌
		timeUntilExpire := claims.ExpiresAt.Sub(time.Now())
		newRefreshToken := refreshTokenString // 默认使用现有的刷新令牌

		// 如果刷新令牌将在7天内过期，则生成新的刷新令牌
		if timeUntilExpire < time.Hour*24*7 {
			refreshClaims := &Claims{
				Username: claims.Username,
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(global.Configs.JWT.RefreshTokenExpire) * time.Second)),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
					ID:        "2",
				},
			}
			newRefreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
			newRefreshToken, err = newRefreshTokenObj.SignedString([]byte(global.Configs.JWT.Refresh))
			if err != nil {
				return nil, fmt.Errorf("failed to generate new refresh token: %w", err)
			}
		}

		return &TokenPair{
			AccessToken:  newAccessToken,
			RefreshToken: newRefreshToken,
		}, nil
	}

	return nil, fmt.Errorf("invalid refresh token")
}

func ParseToken(tokenString string) (*Claims, error) {
	// 检查token是否为空
	if tokenString == "" {
		return nil, fmt.Errorf("token is empty")
	}

	// 强制校验是否以 "Bearer " 开头（严格大小写匹配）
	if !strings.HasPrefix(tokenString, "Bearer ") {
		return nil, fmt.Errorf("token must start with 'Bearer ' prefix")
	}

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 确保签名方法是 HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(global.Configs.JWT.Access), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ParseRefreshToken 解析刷新令牌
func ParseRefreshToken(tokenString string) (*Claims, error) {
	// 解析刷新令牌
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 确保签名方法是 HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(global.Configs.JWT.Refresh), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid refresh token")
}

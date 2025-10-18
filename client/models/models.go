package models

// AuthReq 认证请求结构体
type AuthReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse 登录响应结构体
type LoginResponse struct {
	Message      string `json:"message"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserUUID     string `json:"user_uuid"`
	Salt         string `json:"salt"`
}

// RegisterResponse 注册响应结构体
type RegisterResponse struct {
	Message  string `json:"message"`
	UserUUID string `json:"user_uuid"`
}

// HealthCheckResponse 健康检查响应结构体
type HealthCheckResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}
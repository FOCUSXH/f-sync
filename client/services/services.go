package services

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"fsync/client/configs"
	"fsync/client/models"
	"fsync/pkg/utils"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

// AuthTokens 认证令牌结构
type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// getHTTPClient 获取HTTP客户端，根据证书存在情况决定是否跳过TLS验证
func getHTTPClient(config *configs.Config) (*http.Client, error) {
	// 获取证书文件路径 - 存储在与token相同的目录下
	certPath := filepath.Join(filepath.Dir(config.GetTokenFilePath()), "server.crt")

	// 检查证书文件是否存在
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		// 证书不存在，下载证书
		if err := downloadCertificate(config, certPath); err != nil {
			return nil, fmt.Errorf("下载证书失败: %w", err)
		}
	}

	// 读取证书文件
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("读取证书文件失败: %w", err)
	}

	// 创建CertPool并添加证书
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(certData) {
		return nil, fmt.Errorf("解析证书失败")
	}

	// 创建TLS配置
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	// 创建HTTP客户端
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: 30 * time.Second,
	}

	return client, nil
}

// downloadCertificate 从服务器下载证书
func downloadCertificate(config *configs.Config, certPath string) error {
	// 创建目录
	dir := filepath.Dir(certPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建证书目录失败: %w", err)
	}

	// 构建证书下载URL
	serverURL := config.GetServerURL()
	url := fmt.Sprintf("%s/certificate", serverURL)

	// 创建HTTP客户端（首次下载时跳过TLS验证）
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 30 * time.Second,
	}

	// 下载证书
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("请求证书失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载证书失败，状态码: %d", resp.StatusCode)
	}

	// 保存证书到文件
	certFile, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("创建证书文件失败: %w", err)
	}
	defer certFile.Close()

	_, err = io.Copy(certFile, resp.Body)
	if err != nil {
		return fmt.Errorf("保存证书文件失败: %w", err)
	}

	return nil
}

// CheckAndDownloadCertificate 检查并下载服务器证书（如果需要）
func CheckAndDownloadCertificate(config *configs.Config) error {
	// 获取HTTP客户端，这会自动处理证书下载
	_, err := getHTTPClient(config)
	return err
}

// HealthCheck performs a health check on the server
func HealthCheck(config *configs.Config) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get HTTP client with proper TLS config
	client, err := getHTTPClient(config)
	if err != nil {
		return fmt.Errorf("获取HTTP客户端失败: %w", err)
	}

	// Get server URL
	serverURL := config.GetServerURL()
	url := fmt.Sprintf("%s/health", serverURL)

	// Create a new request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("健康检查请求创建失败: %w", err)
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送健康检查请求失败: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("健康检查失败，状态码: %d", resp.StatusCode)
	}

	// Parse response
	var healthResp models.HealthCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		return fmt.Errorf("解析健康检查响应失败: %w", err)
	}

	// Check status
	if healthResp.Status != "ok" {
		return fmt.Errorf("服务器健康检查未通过: %s", healthResp.Message)
	}

	return nil
}

// Login 用户登录
func Login(config *configs.Config, username, password string) (*AuthTokens, error) {
	// Perform health check first
	if err := HealthCheck(config); err != nil {
		return nil, fmt.Errorf("健康检查失败: %w", err)
	}

	// 构造登录请求
	loginReq := models.AuthReq{
		Username: username,
		Password: password,
	}

	// 获取HTTP客户端
	client, err := getHTTPClient(config)
	if err != nil {
		return nil, fmt.Errorf("获取HTTP客户端失败: %w", err)
	}

	// 发送登录请求
	serverURL := config.GetServerURL()
	url := fmt.Sprintf("%s/api/v1/auth/login", serverURL)
	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return nil, fmt.Errorf("构造登录请求失败: %w", err)
	}

	resp, err := client.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("发送登录请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("登录失败，状态码: %d", resp.StatusCode)
	}

	var loginResp models.LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return nil, fmt.Errorf("解析登录响应失败: %w", err)
	}

	// 保存令牌到文件
	tokens := &AuthTokens{
		AccessToken:  loginResp.AccessToken,
		RefreshToken: loginResp.RefreshToken,
	}

	if err := saveTokens(config, tokens); err != nil {
		return nil, fmt.Errorf("保存令牌失败: %w", err)
	}

	// 保存salt到文件
	if err := saveSalt(config, loginResp.Salt); err != nil {
		return nil, fmt.Errorf("保存salt失败: %w", err)
	}

	return tokens, nil
}

// Register 用户注册
func Register(config *configs.Config, username, password string) error {
	// Perform health check first
	if err := HealthCheck(config); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	// 验证密码强度
	if err := utils.ValidatePassword(password); err != nil {
		return fmt.Errorf("密码强度不符合要求: %w", err)
	}

	// 获取管理员凭据
	fmt.Println("需要管理员权限进行用户注册")
	adminUsername, adminPassword, err := GetCredentials()
	if err != nil {
		return fmt.Errorf("获取管理员凭据失败: %w", err)
	}

	// 构造注册请求
	registerReq := models.AuthReq{
		Username: username,
		Password: password,
	}

	// 获取HTTP客户端（带TLS配置）
	client, err := getHTTPClient(config)
	if err != nil {
		return fmt.Errorf("获取HTTP客户端失败: %w", err)
	}

	// 发送注册请求
	serverURL := config.GetServerURL()
	url := fmt.Sprintf("%s/api/v1/auth/register", serverURL)
	jsonData, err := json.Marshal(registerReq)
	if err != nil {
		return fmt.Errorf("构造注册请求失败: %w", err)
	}

	// 创建请求并添加管理员凭据到请求头
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("创建注册请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Username", adminUsername)
	req.Header.Set("X-Admin-Password", adminPassword)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送注册请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	if resp.StatusCode != http.StatusOK {
		// 尝试解析错误响应
		var errorResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			if errMsg, ok := errorResp["error"].(string); ok {
				return fmt.Errorf("注册失败: %s", errMsg)
			}
		}
		return fmt.Errorf("注册失败，状态码: %d", resp.StatusCode)
	}

	var registerResp models.RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&registerResp); err != nil {
		return fmt.Errorf("解析注册响应失败: %w", err)
	}

	fmt.Printf("用户注册成功，用户UUID: %s\n", registerResp.UserUUID)

	// 注册成功后清除可能存在的旧token
	if err := ClearTokens(config); err != nil {
		return fmt.Errorf("清除旧令牌失败: %w", err)
	}

	return nil
}

// GetCredentialsWithConfirmation 从用户输入获取用户名和密码并确认密码
func GetCredentialsWithConfirmation() (string, string, error) {
	// 获取用户名
	fmt.Print("用户名: ")
	reader := bufio.NewReader(os.Stdin)
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", "", fmt.Errorf("读取用户名失败: %w", err)
	}
	username = strings.TrimSpace(username)

	// 获取密码（隐藏输入）
	fmt.Print("密码: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", fmt.Errorf("读取密码失败: %w", err)
	}
	password := string(bytePassword)
	fmt.Println() // 换行

	// 验证密码强度
	if err := utils.ValidatePassword(password); err != nil {
		return "", "", fmt.Errorf("密码强度不符合要求: %w", err)
	}

	// 确认密码
	fmt.Print("确认密码: ")
	byteConfirmPassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", fmt.Errorf("读取确认密码失败: %w", err)
	}
	confirmPassword := string(byteConfirmPassword)
	fmt.Println() // 换行

	// 检查密码是否匹配
	if password != confirmPassword {
		return "", "", fmt.Errorf("两次输入的密码不一致")
	}

	return username, password, nil
}

// GetCredentials 从用户输入获取用户名和密码
func GetCredentials() (string, string, error) {
	// 获取用户名
	fmt.Print("用户名: ")
	reader := bufio.NewReader(os.Stdin)
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", "", fmt.Errorf("读取用户名失败: %w", err)
	}
	username = strings.TrimSpace(username)

	// 获取密码（隐藏输入）
	fmt.Print("密码: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", fmt.Errorf("读取密码失败: %w", err)
	}
	password := string(bytePassword)
	fmt.Println() // 换行

	return username, password, nil
}

// GetCredentialsWithValidation 从用户输入获取用户名和密码并进行密码强度验证
func GetCredentialsWithValidation() (string, string, error) {
	// 获取用户名
	fmt.Print("用户名: ")
	reader := bufio.NewReader(os.Stdin)
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", "", fmt.Errorf("读取用户名失败: %w", err)
	}
	username = strings.TrimSpace(username)

	// 获取密码（隐藏输入）
	fmt.Print("密码: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", fmt.Errorf("读取密码失败: %w", err)
	}
	password := string(bytePassword)
	fmt.Println() // 换行

	// 验证密码强度
	if err := utils.ValidatePassword(password); err != nil {
		return "", "", fmt.Errorf("密码强度不符合要求: %w", err)
	}

	// 确认密码
	fmt.Print("确认密码: ")
	byteConfirmPassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", fmt.Errorf("读取确认密码失败: %w", err)
	}
	confirmPassword := string(byteConfirmPassword)
	fmt.Println() // 换行

	// 检查密码是否匹配
	if password != confirmPassword {
		return "", "", fmt.Errorf("两次输入的密码不一致")
	}

	return username, password, nil
}

// saveTokens 保存令牌到文件
func saveTokens(config *configs.Config, tokens *AuthTokens) error {
	// 确保Token目录存在
	tokenFilePath := config.GetTokenFilePath()
	tokenDir := filepath.Dir(tokenFilePath)

	if tokenDir != "" {
		if err := os.MkdirAll(tokenDir, 0700); err != nil {
			return fmt.Errorf("创建令牌目录失败: %w", err)
		}
	}

	// 写入令牌文件
	file, err := os.OpenFile(tokenFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("创建令牌文件失败: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(tokens); err != nil {
		return fmt.Errorf("编码令牌失败: %w", err)
	}

	return nil
}

// LoadTokens 从文件加载令牌
func LoadTokens(config *configs.Config) (*AuthTokens, error) {
	tokenFilePath := config.GetTokenFilePath()

	file, err := os.Open(tokenFilePath)
	if err != nil {
		return nil, fmt.Errorf("打开令牌文件失败: %w", err)
	}
	defer file.Close()

	var tokens AuthTokens
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&tokens); err != nil {
		return nil, fmt.Errorf("解码令牌失败: %w", err)
	}

	return &tokens, nil
}

// ClearTokens 清除存储的令牌（登出）
func ClearTokens(config *configs.Config) error {
	tokenFilePath := config.GetTokenFilePath()

	if err := os.Remove(tokenFilePath); err != nil {
		// 如果文件不存在，不视为错误
		if !os.IsNotExist(err) {
			return fmt.Errorf("删除令牌文件失败: %w", err)
		}
	}

	return nil
}

// saveSalt 保存salt到文件
func saveSalt(config *configs.Config, salt string) error {
	// 获取salt文件路径
	saltFile := filepath.Join(filepath.Dir(config.GetTokenFilePath()), "salt")

	// 写入salt到文件
	if err := os.WriteFile(saltFile, []byte(salt), 0600); err != nil {
		return fmt.Errorf("写入salt文件失败: %w", err)
	}

	return nil
}

// LoadSalt 从文件加载salt
func LoadSalt(config *configs.Config) (string, error) {
	// 获取salt文件路径
	saltFile := filepath.Join(filepath.Dir(config.GetTokenFilePath()), "salt")

	// 读取salt文件
	salt, err := os.ReadFile(saltFile)
	if err != nil {
		return "", fmt.Errorf("读取salt文件失败: %w", err)
	}

	return string(salt), nil
}

package aws

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// 全局常量定义
const (
	// metadataBaseURL EC2元数据服务基础URL
	metadataBaseURL = "http://169.254.169.254/latest"
)

// EC2Client AWS EC2客户端
type EC2Client struct {
	client *ec2.Client
}

// NewEC2Client 创建新的EC2客户端
func NewEC2Client() (*EC2Client, error) {
	// 获取当前实例所在region
	region, err := GetRegion()
	if err != nil {
		return nil, fmt.Errorf("failed to get region: %w", err)
	}

	// 加载AWS配置，自动使用EC2实例角色
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	// 创建EC2客户端
	client := ec2.NewFromConfig(cfg)

	return &EC2Client{
		client: client,
	}, nil
}

// getMetadataToken 获取EC2元数据token
func getMetadataToken() (string, error) {
	// 创建http客户端
	client := &http.Client{}
	// 创建PUT请求
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/token", metadataBaseURL), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}
	// 设置请求头
	req.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", "21600")
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get metadata token: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get metadata token, status code: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response: %w", err)
	}

	token := strings.TrimSpace(string(body))
	if token == "" {
		return "", fmt.Errorf("empty token returned")
	}

	return token, nil
}

// getMetadataWithToken 使用token获取EC2元数据
func getMetadataWithToken(path string) (string, error) {
	// 获取token
	token, err := getMetadataToken()
	if err != nil {
		return "", err
	}

	// 创建http客户端
	client := &http.Client{}
	// 创建GET请求
	url := fmt.Sprintf("%s/meta-data/%s", metadataBaseURL, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create metadata request: %w", err)
	}
	// 设置请求头
	req.Header.Set("X-aws-ec2-metadata-token", token)
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get metadata %s: %w", path, err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get metadata %s, status code: %d", path, resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read metadata response: %w", err)
	}

	metadata := strings.TrimSpace(string(body))
	if metadata == "" {
		return "", fmt.Errorf("empty %s returned", path)
	}

	return metadata, nil
}

// GetInstanceID 获取当前实例ID
func GetInstanceID() (string, error) {
	// 尝试从环境变量获取实例ID（用于测试）
	instanceID := os.Getenv("EC2_INSTANCE_ID")
	if instanceID != "" {
		return instanceID, nil
	}

	// 从EC2实例元数据获取实例ID，使用token认证
	return getMetadataWithToken("instance-id")
}

// GetRegion 获取当前实例所在region
func GetRegion() (string, error) {
	// 尝试从环境变量获取region（用于测试）
	region := os.Getenv("AWS_REGION")
	if region != "" {
		return region, nil
	}

	// 从EC2实例元数据获取region，使用token认证
	return getMetadataWithToken("placement/region")
}

// TerminateInstance 终止当前实例
func (ec *EC2Client) TerminateInstance(instanceID string) error {
	// 调用AWS EC2 API终止实例
	_, err := ec.client.TerminateInstances(context.TODO(), &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return fmt.Errorf("failed to terminate instance: %w", err)
	}

	return nil
}

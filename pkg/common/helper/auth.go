package helper

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"
)

// 身份校验工具

// ExtractTokenFromGrpc  从 gRPC 上下文（metadata）提取 Token
// 参数：
// 		- ctx context.Context: 上下文信息
// 返回：
// 		- string:
// 		- error:
func ExtractTokenFromGrpc(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}
    authHeaders := md.Get("authorization")
    if len(authHeaders) == 0 {
        return "", false
    }
    // 解析 "Bearer <token>" 格式
    parts := strings.Split(authHeaders[0], " ")
    if len(parts) != 2 || parts[0] != "Bearer" {
        return "", false
    }
    return parts[1], true
}

// ValidateToken 验证token
// 参数：
// 		- token string：
// 返回：
// 		- string
//		- error
func ValidateToken(token string) (string, error) {
	return "",nil
}
package middleware

import (
	"context"
	"my_workflow/pkg/common/helper"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)


// AuthInterceptor grpc拦截器: 核心验证
// 
//
// 参数：
// 		- ctx context.Context:
// 		- req interface{}:
// 		- info
//		- handler
// 返回：
// 		- 
func AuthInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 获取token并验证
	token, ok := helper.ExtractTokenFromGrpc(ctx)
    if !ok {
        return nil, status.Errorf(codes.Unauthenticated, "missing token")
    }
    
    // 验证 token
    userID, err := helper.ValidateToken(token)
    if err != nil {
        return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
    }
    
    // 将用户信息添加到 context
    ctx = context.WithValue(ctx, "user_id", userID)
    
    // 继续处理请求
    return handler(ctx, req)
}
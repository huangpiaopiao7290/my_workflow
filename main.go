package main

import (
	"context"
	"flag"
	"fmt"
	"my_workflow/config"
	"my_workflow/pkg/common/logger"
	"my_workflow/pkg/database/mongodb"
	"my_workflow/services/card"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "my_workflow/api/card/v1"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	rpcPort = flag.Int("rpc-port", 0, "grpc server port")
	httpPort = flag.Int("http-port", 0, "http gateway port")
)

// 启动grpc服务器
func startGRPCServer(port int) (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen: %v", err)
	}
	// 创建grpc服务器
	s := grpc.NewServer()
	// 注册卡片服务
	cardService := card.NewCardService()
	pb.RegisterCardServiceServer(s, cardService)

	logger.Info(context.TODO(), "starting grpc server", map[string]string{
		"port": fmt.Sprintf("%d", port),
	})

	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Error(context.TODO(), "failed to serve grpc", map[string]string{
				"error": err.Error(),
			})
		}
	}()

	return s, lis, nil
}

// 启动http网关
func startHTTPServer(grpcPort int, httpPort int) (*http.Server, error) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	// 创建gRPC-Gateway mux
	mux := runtime.NewServeMux()
	
	// 设置gRPC服务器地址
	grpcAddr := fmt.Sprintf(":%d", grpcPort)
	
	// 注册CardService处理程序
	err := pb.RegisterCardServiceHandlerFromEndpoint(ctx, mux, grpcAddr, []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register gateway: %v", err)
	}
	
	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", httpPort),
		Handler: mux,
	}
	
	logger.Info(context.TODO(), "starting HTTP gateway", map[string]string{
		"port": fmt.Sprintf("%d", httpPort),
	})
	
	// 启动HTTP服务器（非阻塞）
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(context.TODO(), "failed to serve HTTP", map[string]string{
				"error": err.Error(),
			})
		}
	}()
	
	return srv, nil
}

func main() {
	// 解析命令行参数
	flag.Parse()
	
	// 从配置文件获取端口，如果命令行有指定则使用命令行参数
	rpcPort := *rpcPort
	if rpcPort == 0 {
		rpcPort = config.GetInt("app.rpcPort")
	}
	
	httpPort := *httpPort
	if httpPort == 0 {
		httpPort = config.GetInt("app.httpPort")
	}
	
	// 初始化MongoDB连接（使用已封装的mongodb包）
	if err := mongodb.NewClient(); err != nil {
		logger.Fatal(context.TODO(), "failed to initialize mongodb", map[string]string{
			"error": err.Error(),
		})
	}
	defer func() {
		client := mongodb.GetClient()
		if client != nil {
			if err := client.Disconnect(context.TODO()); err != nil {
				logger.Error(context.TODO(), "failed to disconnect mongodb", map[string]string{
					"error": err.Error(),
				})
			}
		}
	}()
	
	// 启动gRPC服务器
	grpcServer, _, err := startGRPCServer(rpcPort)
	if err != nil {
		logger.Fatal(context.TODO(), "failed to start gRPC server", map[string]string{
			"error": err.Error(),
		})
	}
	
	// 启动HTTP网关
	httpServer, err := startHTTPServer(rpcPort, httpPort)
	if err != nil {
		logger.Fatal(context.TODO(), "failed to start HTTP server", map[string]string{
			"error": err.Error(),
		})
	}
	
	// 等待中断信号关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info(context.TODO(), "shutting down servers...")
	
	// 关闭HTTP服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error(context.TODO(), "HTTP server forced to shutdown", map[string]string{
			"error": err.Error(),
		})
	}
	
	// 关闭gRPC服务器
	grpcServer.GracefulStop()
	
	logger.Info(context.TODO(), "servers exiting")
}
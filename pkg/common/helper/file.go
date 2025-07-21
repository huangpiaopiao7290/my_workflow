package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"my_workflow/pkg/common/logger"
	"os"
	"os/exec"
	"strings"
)

// DownloadByCurl 使用curl命令从指定URL下载文件到指定输出路径
// 参数:
//
//	url: 需要下载的文件URL
//	outPath: 下载文件的输出路径
//
// 返回值:
//
//	error: 如果下载过程中发生错误，则返回错误信息，否则返回nil
func DownloadByCurl(url, outPath string) error {
	tempFile, err := os.Create(outPath)
	if err != nil {
		// 如果创建临时文件失败，返回错误信息
		return fmt.Errorf("create temp file error: %w", err)
	}
	// 确保在函数结束时关闭临时文件
	defer func(tempFile *os.File) {
		err := tempFile.Close()
		if err != nil {
			// 如果文件关闭失败，记录警告日志
			logger.Warn(context.Background(), "File close failed", map[string]string{
				"Error":    err.Error(),
				"FileName": tempFile.Name(),
			})
		}
	}(tempFile)
	// 因为后面已经Rename文件 这里无需删除临时文件
	defer func(name string) {
		// 确保临时文件在函数结束时被删除
	}(tempFile.Name())

	// 使用curl命令下载文件
	cmd := exec.Command("curl", "-o", tempFile.Name(), url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 如果下载过程中发生错误，记录错误日志并返回错误信息
		logger.Error(context.Background(), "curl download file error", map[string]string{
			"Error":  err.Error(),
			"URL":    url,
			"cmd":    cmd.String(),
			"output": string(output),
		})
		return err
	}
	// 检查curl命令的输出是否包含100%，以确认下载是否成功
	if !strings.Contains(string(output), "100") {
		// 如果输出不符合预期，记录错误日志并返回错误信息
		logger.Error(context.Background(), "curl download file unexpected output", map[string]string{
			"URL":    url,
			"cmd":    cmd.String(),
			"output": string(output),
		})
		return fmt.Errorf("unexpected curl output")
	}
	// 下载成功，返回nil
	return nil
}

// DownloadAndRenameFile 从给定的URL下载文件到本地，并将其重命名为指定的名字。
//
// 参数:
//
//	url string: 文件的下载地址
//	outputPath string: 文件保存的本地目录路径
//
// 返回值:
//
//	error: 如果下载或重命名过程中发生错误，则返回该错误
func DownloadAndRenameFile(url, outputPath string) error {
	if IsURL(url) {
		err := DownloadByCurl(url, outputPath)
		if err != nil {
			return err
		}
	} else {
		// 移动临时文件到目标位置并重命名
		if err := os.Rename(url, outputPath); err != nil {
			return fmt.Errorf("rename file error: %w", err)
		}
	}
	return nil
}

// UnmarshalByte 解析JSON字节切片数据到指定的结果对象。
// ctx: 上下文，用于传递请求范围的信息。
// data: 待解析的JSON字节切片。
// result: 解析结果将被存储的变量指针。
func UnmarshalByte(ctx context.Context, data []byte, result any) {
	err := json.Unmarshal(data, &result)
	if err != nil {
		logger.Error(ctx, "Unmarshal error", map[string]interface{}{
			"Error": err.Error(),
			"Data":  string(data),
		})
	}
}

package helper

import (
	"fmt"
	"net/url"
	"time"
)

const (
	DateFormat      = "20060102"                //'yyyyMMdd' 模式的日期格式
	TimestampFormat = "2006-01-02 15:04:05.000" //yyyy-mm-dd hh:ii:ss模式的时间格式
)

// RetryWithBackoff 使用退避重试策略执行一个函数，直到成功或达到最大重试次数。
// 这个函数接收最大重试次数（maxRetries）、初始延迟时间（delay）和一个错误返回函数（fn）作为参数。
// 如果函数执行成功，则返回nil；如果最终执行失败，则返回最后一次执行的错误。
// 当所有重试都失败后，返回一个自定义错误，指示操作在指定的重试次数后失败。
func RetryWithBackoff(maxRetries int, delay time.Duration, fn func() error) error {
	// 循环尝试执行传入的函数，直到成功或达到最大重试次数。
	for i := 0; i < maxRetries; i++ {
		// 尝试执行函数。
		err := fn()
		// 如果函数执行成功，没有错误，则直接返回nil。
		if err == nil {
			return nil
		}
		// 如果当前是最后一次尝试且执行失败，则返回执行错误。
		if i == maxRetries-1 {
			return err
		}
		// 在每次失败后，根据当前循环索引增加延迟，以实现退避重试策略。
		time.Sleep(delay * time.Duration(i+1))
	}
	// 如果所有重试都失败了，返回一个自定义错误，指示操作在指定的重试次数后失败。
	return fmt.Errorf("operation failed after %d retries", maxRetries)
}

// IsURL 检查给定的字符串是否构成一个有效的HTTP或HTTPS URL。
// 它通过解析输入字符串并验证其方案（scheme）和主机（host）来实现这一点。
//
// 参数:
//
//	input - 待验证的URL字符串。
//
// 返回值:
//
//	如果输入字符串是一个有效的HTTP或HTTPS URL，则返回true；否则返回false。
func IsURL(input string) bool {
	// 解析输入字符串以获取URL结构体。
	u, err := url.Parse(input)
	// 如果解析过程中出现错误，则返回false，表明输入不是一个有效的URL。
	if err != nil {
		return false
	}

	// 定义一个有效的URL方案映射，仅包含"http"和"https"。
	validSchemes := map[string]bool{
		"http":  true,
		"https": true,
	}

	// 检查URL的方案是否为"http"或"https"，并且有一个非空的主机。
	// 如果两者都满足，则返回true，表明输入是一个有效的URL。
	// 否则，返回false。
	return validSchemes[u.Scheme] && u.Host != ""
}

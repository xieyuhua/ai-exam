package util

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// ParsePagination 从查询参数中解析分页参数，返回 page, pageSize
func ParsePagination(pageStr, pageSizeStr string) (int, int) {
	page := 1
	pageSize := 15
	if pageStr != "" {
		if v, ok := parseInt(pageStr); ok && v > 0 {
			page = v
		}
	}
	if pageSizeStr != "" {
		if v, ok := parseInt(pageSizeStr); ok && v > 0 && v <= 100 {
			pageSize = v
		}
	}
	return page, pageSize
}

func parseInt(s string) (int, bool) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
	}
	return n, true
}

// NormalizeQuestionType 规范化题型名称（支持中英文）
func NormalizeQuestionType(s string) string {
	lower := strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.Contains(lower, "多选") || strings.Contains(lower, "multi") || lower == "multiple":
		return "multiple"
	case strings.Contains(lower, "判断") || strings.Contains(lower, "判") || lower == "judge" || lower == "boolean":
		return "judge"
	case strings.Contains(lower, "多") && strings.Contains(lower, "选"):
		// "多选题" 等包含"多选"的变体，已在上面处理；此处兜底"多选"类关键词
		return "multiple"
	case strings.Contains(lower, "填空") || lower == "fill" || lower == "blank":
		return "fill"
	case strings.Contains(lower, "简答") || strings.Contains(lower, "问答") || lower == "essay" || lower == "short_answer":
		return "essay"
	default:
		return "single"
	}
}

// StringSliceEqual 比较两个已排序的字符串切片是否相等
func StringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// SortSlice 返回排序后的副本
func SortSlice(s []string) []string {
	result := make([]string, len(s))
	copy(result, s)
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i] > result[j] {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result
}

// HashPassword 使用 SHA256 哈希密码
func HashPassword(password string) string {
	h := sha256.Sum256([]byte(password))
	return hex.EncodeToString(h[:])
}

// CheckPassword 校验密码
func CheckPassword(hashed, password string) bool {
	return hashed == HashPassword(password)
}

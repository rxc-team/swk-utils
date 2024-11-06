package helpers

import (
	"regexp"
	"strings"
)

// Escape 自动转义括号
func Escape(str string) string {
	// 正则匹配出现的所有特殊字符
	fbsArr := []string{"$", "(", ")", "\\", "*", "+", ".", "[", "]", "?", "^", "{", "}", "/"}
	for _, ch := range fbsArr {
		StrContainers := strings.Contains(str, ch)
		if StrContainers {
			str = strings.Replace(str, ch, "\\"+ch, -1)
		}
	}
	return str
}

// 特殊字符检查，合法返回true,否则返回false
func SpecialCheck(value string, special string) bool {
	// 特殊字符为空的情况,直接返回
	if len(special) == 0 {
		return true
	}

	specialReg := regexp.QuoteMeta(special)
	// 判断特殊字符是否包含减号
	hasMinus := strings.Contains(specialReg, "-")
	if hasMinus {
		specialReg = strings.Replace(specialReg, "-", "\\-", 1)
	}
	re := regexp.MustCompile("[" + specialReg + "]")
	return !re.MatchString(value)
}

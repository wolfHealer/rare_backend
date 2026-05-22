// Package search 提供关键词检索辅助（FULLTEXT ngram + 最短长度校验）。
package search

import (
	"strings"
	"unicode/utf8"
)

const MinRunes = 2

func Trim(keyword string) string {
	return strings.TrimSpace(keyword)
}

func Usable(keyword string) bool {
	return utf8.RuneCountInString(Trim(keyword)) >= MinRunes
}

// MatchClause 返回 FULLTEXT 条件片段： AND MATCH(cols) AGAINST(? IN NATURAL LANGUAGE MODE)
// 关键词为空或过短时 ok=false（避免 %keyword% 全表扫描）。
func MatchClause(cols string, keyword string) (clause string, arg string, ok bool) {
	kw := Trim(keyword)
	if kw == "" || !Usable(kw) {
		return "", "", false
	}
	return " AND MATCH(" + cols + ") AGAINST(? IN NATURAL LANGUAGE MODE)", kw, true
}

// MatchCondition 同 MatchClause，但不带前导 AND，便于拼接到 WHERE 条件切片。
func MatchCondition(cols string, keyword string) (condition string, arg string, ok bool) {
	clause, arg, ok := MatchClause(cols, keyword)
	if !ok {
		return "", "", false
	}
	return strings.TrimPrefix(clause, " AND "), arg, true
}

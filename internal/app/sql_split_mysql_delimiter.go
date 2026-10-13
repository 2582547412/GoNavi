package app

import "strings"

// MySQL 客户端的 DELIMITER 不是服务端语法：mysql 命令行、Navicat 等用它临时更换语句结束符，
// 好让 CREATE PROCEDURE 里的分号不被当成语句边界。服务端收到 DELIMITER 行会报 1064，
// 所以所有下发路径都必须在客户端识别并丢弃它，再按当前结束符拆分语句。
// 只在 mysql / mariadb 方言下识别，其它方言里的 delimiter 仍按普通标识符处理。

func isSQLStreamMySQLClientDialect(dbType string) bool {
	switch normalizeExplainLexicalDBType(dbType) {
	case "mysql", "mariadb":
		return true
	default:
		return false
	}
}

func (s *sqlStreamSplitter) canConsumeClientDirective() bool {
	return sqlCurrentLineWhitespaceOnly(&s.cur) && !hasExecutableSQLStatementContent(s.dbType, s.cur.String())
}

// scanSQLStreamDelimiterDirective 尝试在 text[start] 处解析一条 DELIMITER 指令。
// eof 为 true 时把文本结尾视为行尾；流式输入被截断而无法判定时返回 incomplete。
func scanSQLStreamDelimiterDirective(text string, start int, eof bool) (delimiter string, lineEnd int, matched bool, incomplete bool) {
	const keyword = "delimiter"
	if start < 0 || start >= len(text) {
		return "", start, false, false
	}
	remaining := text[start:]
	prefixLength := len(remaining)
	if prefixLength > len(keyword) {
		prefixLength = len(keyword)
	}
	if !strings.EqualFold(remaining[:prefixLength], keyword[:prefixLength]) {
		return "", start, false, false
	}
	if len(remaining) <= len(keyword) {
		return "", start, false, !eof
	}
	if !isSQLHorizontalWhitespace(remaining[len(keyword)]) {
		return "", start, false, false
	}

	newline := strings.IndexByte(remaining, '\n')
	if newline < 0 && !eof {
		return "", start, false, true
	}
	line := remaining
	lineEnd = len(text)
	if newline >= 0 {
		line = remaining[:newline]
		lineEnd = start + newline + 1
	}
	value := strings.TrimSpace(line[len(keyword):])
	if value == "" || strings.IndexFunc(value, func(r rune) bool { return r == ' ' || r == '\t' || r == '\r' || r == '\n' }) >= 0 {
		return "", start, false, false
	}
	return value, lineEnd, true, false
}

// mysqlTextHasDelimiterDirective 报告文本中是否有独占一行的 DELIMITER 指令。
// 只做行首粗筛，命中后交给与 SQL 导入相同的流式拆分器做准确判定。
func mysqlTextHasDelimiterDirective(dbType, text string) bool {
	if !isSQLStreamMySQLClientDialect(dbType) {
		return false
	}
	for lineStart := 0; lineStart < len(text); {
		next := len(text)
		if lineEnd := strings.IndexByte(text[lineStart:], '\n'); lineEnd >= 0 {
			next = lineStart + lineEnd + 1
		}
		trimmed := strings.TrimLeft(text[lineStart:next], " \t\r\f")
		if _, _, matched, _ := scanSQLStreamDelimiterDirective(trimmed, 0, true); matched {
			return true
		}
		lineStart = next
	}
	return false
}

// splitMySQLDelimiterScript 用 SQL 导入同款流式拆分器拆分含 DELIMITER 的脚本：
// DELIMITER 行被丢弃，自定义结束符之间的分号保留在语句内部。
func splitMySQLDelimiterScript(dbType, text string, backslashEscapes bool) []string {
	splitter := &sqlStreamSplitter{dbType: normalizeSQLClassifierDBType(dbType), noBackslashEscapes: !backslashEscapes}
	statements := splitter.Feed([]byte(text))
	return append(statements, splitter.flushStatements()...)
}

// mysqlBatchTextNeedsPerStatement 报告原文是否因含 DELIMITER 而不能直接下发。
// 原生多结果集与批量写这两条路径传的是用户原文，DELIMITER 行会让服务端报 1064，
// 必须退回逐条执行（逐条下发的是已经去掉 DELIMITER 的语句）。
func mysqlBatchTextNeedsPerStatement(dbType, query string) bool {
	return mysqlTextHasDelimiterDirective(dbType, query)
}

package app

import (
	"strings"
	"testing"

	"GoNavi-Wails/internal/connection"
	"GoNavi-Wails/internal/db"
	"GoNavi-Wails/internal/secretstore"
)

// MySQL 的 DELIMITER 是客户端指令：含它的脚本不能以原文下发（服务端报 1064），
// 必须逐条执行，且每条都是去掉 DELIMITER 后的完整语句。
func TestDBQueryMultiMySQLDelimiterScriptNeverReachesDriverAsRawText(t *testing.T) {
	installFakeOptionalDriverRuntime(t)
	originalNewDatabaseFunc := newDatabaseFunc
	t.Cleanup(func() {
		newDatabaseFunc = originalNewDatabaseFunc
	})

	query := strings.Join([]string{
		"DROP PROCEDURE IF EXISTS p;",
		"DELIMITER $$",
		"CREATE PROCEDURE p()",
		"BEGIN",
		"  WHILE x > 0 DO",
		"    SET x = x - 1;",
		"  END WHILE;",
		"END$$",
		"DELIMITER ;",
	}, "\n")
	baseDB := &fakeBatchWriteDB{
		queryErr:     map[string]error{},
		execAffected: map[string]int64{},
	}
	fakeDB := &fakeNativeMultiResultDB{fakeBatchWriteDB: baseDB}
	newDatabaseFunc = func(dbType string) (db.Database, error) {
		return fakeDB, nil
	}

	app := NewAppWithSecretStore(secretstore.NewUnavailableStore("test"))
	config := connection.ConnectionConfig{Type: "mysql", Host: "127.0.0.1", Port: 3306, User: "root"}

	result := app.DBQueryMulti(config, "demo", query, "mysql-delimiter-script-test")
	if !result.Success {
		t.Fatalf("expected DBQueryMulti success, got failure: %s", result.Message)
	}
	if fakeDB.multiCalls != 0 || baseDB.batchCalls != 0 {
		t.Fatalf("raw-text paths must not receive a DELIMITER script, multiCalls=%d batchCalls=%d", fakeDB.multiCalls, baseDB.batchCalls)
	}
	want := []string{
		"DROP PROCEDURE IF EXISTS p",
		"CREATE PROCEDURE p()\nBEGIN\n  WHILE x > 0 DO\n    SET x = x - 1;\n  END WHILE;\nEND",
	}
	if len(baseDB.execQueries) != len(want) {
		t.Fatalf("exec queries = %#v, want %#v", baseDB.execQueries, want)
	}
	for i, executed := range baseDB.execQueries {
		if executed != want[i] {
			t.Fatalf("exec query %d = %q, want %q", i, executed, want[i])
		}
	}
}

// 只有一条语句时会走批量写原文路径，含 DELIMITER 时同样必须退回逐条执行。
func TestDBQueryMultiMySQLSingleDelimiterStatementSkipsBatchWrite(t *testing.T) {
	installFakeOptionalDriverRuntime(t)
	originalNewDatabaseFunc := newDatabaseFunc
	t.Cleanup(func() {
		newDatabaseFunc = originalNewDatabaseFunc
	})

	query := "DELIMITER $$\nINSERT INTO demo(id) VALUES (1)$$\nDELIMITER ;"
	baseDB := &fakeBatchWriteDB{
		queryErr:     map[string]error{},
		execAffected: map[string]int64{},
	}
	newDatabaseFunc = func(dbType string) (db.Database, error) {
		return baseDB, nil
	}

	app := NewAppWithSecretStore(secretstore.NewUnavailableStore("test"))
	config := connection.ConnectionConfig{Type: "mysql", Host: "127.0.0.1", Port: 3306, User: "root"}

	result := app.DBQueryMulti(config, "demo", query, "mysql-delimiter-single-write-test")
	if !result.Success {
		t.Fatalf("expected DBQueryMulti success, got failure: %s", result.Message)
	}
	if baseDB.batchCalls != 0 {
		t.Fatalf("batch write must not receive a DELIMITER script, got %d calls", baseDB.batchCalls)
	}
	if len(baseDB.execQueries) != 1 || baseDB.execQueries[0] != "INSERT INTO demo(id) VALUES (1)" {
		t.Fatalf("exec queries = %#v", baseDB.execQueries)
	}
}

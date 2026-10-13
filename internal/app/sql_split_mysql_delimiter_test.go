package app

import (
	"reflect"
	"strings"
	"testing"
)

func TestSplitSQLStatementsForDialectHandlesMySQLDelimiter(t *testing.T) {
	tests := []struct {
		name   string
		dbType string
		sql    string
		want   []string
	}{
		{
			name:   "procedure body keeps inner semicolons",
			dbType: "mysql",
			sql: strings.Join([]string{
				"DROP PROCEDURE IF EXISTS p;",
				"DELIMITER $$",
				"CREATE PROCEDURE p()",
				"BEGIN",
				"  SELECT 1;",
				"  SELECT 2;",
				"END$$",
				"DELIMITER ;",
				"CALL p();",
			}, "\n"),
			want: []string{
				"DROP PROCEDURE IF EXISTS p",
				"CREATE PROCEDURE p()\nBEGIN\n  SELECT 1;\n  SELECT 2;\nEND",
				"CALL p()",
			},
		},
		{
			name:   "loops that the semicolon splitter cannot track",
			dbType: "mysql",
			sql: strings.Join([]string{
				"DELIMITER //",
				"CREATE PROCEDURE p()",
				"BEGIN",
				"  WHILE x > 0 DO",
				"    SET x = x - 1;",
				"  END WHILE;",
				"  REPEAT SET x = x + 1; UNTIL x > 3 END REPEAT;",
				"END //",
				"DELIMITER ;",
				"SELECT 'a;b';",
			}, "\n"),
			want: []string{
				"CREATE PROCEDURE p()\nBEGIN\n  WHILE x > 0 DO\n    SET x = x - 1;\n  END WHILE;\n  REPEAT SET x = x + 1; UNTIL x > 3 END REPEAT;\nEND",
				"SELECT 'a;b'",
			},
		},
		{
			name:   "comments before the directive and lowercase keyword",
			dbType: "mariadb",
			sql:    "-- install\ndelimiter ;;\nCREATE TRIGGER t BEFORE INSERT ON a FOR EACH ROW BEGIN SET NEW.x = 1; END;;\ndelimiter ;\nSELECT 1;",
			want: []string{
				"CREATE TRIGGER t BEFORE INSERT ON a FOR EACH ROW BEGIN SET NEW.x = 1; END",
				"SELECT 1",
			},
		},
		{
			name:   "delimiter text inside a string is data",
			dbType: "mysql",
			sql:    "SELECT 'x\nDELIMITER $$\ny';\nSELECT 2;",
			want:   []string{"SELECT 'x\nDELIMITER $$\ny'", "SELECT 2"},
		},
		{
			name:   "other dialects do not interpret delimiter",
			dbType: "sqlserver",
			sql:    "DELIMITER $$\nSELECT 1;",
			want:   []string{"DELIMITER $$\nSELECT 1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitSQLStatementsForDialect(tt.dbType, tt.sql)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("statements = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestSplitSQLStatementsForDialectModeHonorsNoBackslashEscapesWithDelimiter(t *testing.T) {
	sql := "DELIMITER $$\nSELECT 'a\\'$$\nSELECT 2$$\nDELIMITER ;"
	got := SplitSQLStatementsForDialectMode("mysql", sql, false)
	want := []string{"SELECT 'a\\'", "SELECT 2"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("statements = %#v, want %#v", got, want)
	}
}

func TestMySQLBatchTextNeedsPerStatement(t *testing.T) {
	tests := []struct {
		name   string
		dbType string
		sql    string
		want   bool
	}{
		{"directive line", "mysql", "DELIMITER $$\nSELECT 1$$", true},
		{"indented directive", "mariadb", "SELECT 1;\n  delimiter //\nSELECT 2//", true},
		{"plain batch", "mysql", "SELECT 1; SELECT 2;", false},
		{"identifier named delimiter", "mysql", "SELECT delimiter FROM t;", false},
		{"bare keyword", "mysql", "DELIMITER\nSELECT 1;", false},
		{"non-mysql dialect", "sqlserver", "DELIMITER $$\nSELECT 1$$", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mysqlBatchTextNeedsPerStatement(tt.dbType, tt.sql); got != tt.want {
				t.Fatalf("mysqlBatchTextNeedsPerStatement() = %v, want %v", got, tt.want)
			}
		})
	}
}

package sync

import (
	"reflect"
	"testing"

	"GoNavi-Wails/internal/connection"
)

func TestIsGeneratedColumnExtraMarkers(t *testing.T) {
	t.Parallel()

	generated := []string{
		"VIRTUAL GENERATED",
		"STORED GENERATED",
		"virtual generated",
		"  stored   generated  ",
		"GENERATED ALWAYS AS (...) STORED",
		"GENERATED",
		"MATERIALIZED",
		"ALIAS",
		"COMPUTED BY (x + y)",
	}
	for _, extra := range generated {
		if !connection.IsGeneratedColumnExtra(extra) {
			t.Fatalf("IsGeneratedColumnExtra(%q) = false, want true", extra)
		}
	}

	writable := []string{
		"",
		"auto_increment",
		"DEFAULT_GENERATED",
		"default_generated on update CURRENT_TIMESTAMP",
		"on update CURRENT_TIMESTAMP",
	}
	for _, extra := range writable {
		if connection.IsGeneratedColumnExtra(extra) {
			t.Fatalf("IsGeneratedColumnExtra(%q) = true, want false", extra)
		}
	}
}

func generatedColumnSyncConfig() SyncConfig {
	return SyncConfig{
		SourceConfig:        connection.ConnectionConfig{Type: "mysql", Database: "source_db", Host: "source.local"},
		TargetConfig:        connection.ConnectionConfig{Type: "mysql", Database: "target_db", Host: "target.local"},
		Tables:              []string{"crm_filing_offline"},
		Content:             "data",
		Mode:                "insert_update",
		TargetTableStrategy: "existing_only",
		TableOptions: map[string]TableOptions{
			"crm_filing_offline": {Insert: true, Update: true},
		},
	}
}

func generatedColumnSourceDB() *fakeMigrationDB {
	return &fakeMigrationDB{
		columns: map[string][]connection.ColumnDefinition{
			"source_db.crm_filing_offline": {
				{Name: "id", Type: "bigint", Nullable: "NO", Key: "PRI", Extra: "auto_increment"},
				{Name: "contract_no", Type: "varchar(64)", Nullable: "YES"},
				{Name: "company", Type: "varchar(128)", Nullable: "YES"},
			},
		},
		queryData: map[string][]map[string]interface{}{
			"SELECT `id`, `contract_no`, `company` FROM `source_db`.`crm_filing_offline` ORDER BY `id` ASC LIMIT 1000 OFFSET 0": {
				{"id": int64(1), "contract_no": "HT-1", "company": "甲公司"},
				{"id": int64(2), "contract_no": "HT-2", "company": "乙公司"},
			},
		},
	}
}

func generatedColumnTargetDB() *fakeQuerySyncTargetDB {
	return &fakeQuerySyncTargetDB{fakeMigrationDB: fakeMigrationDB{
		columns: map[string][]connection.ColumnDefinition{
			"target_db.crm_filing_offline": {
				{Name: "id", Type: "bigint", Nullable: "NO", Key: "PRI", Extra: "auto_increment"},
				{Name: "contract_no", Type: "varchar(64)", Nullable: "YES"},
				{Name: "company", Type: "varchar(128)", Nullable: "YES"},
				// 目标表比源表多出的生成列：数据库侧算值，写入必报 Error 3105。
				{Name: "contract_company", Type: "varchar(255)", Nullable: "YES", Extra: "VIRTUAL GENERATED"},
			},
		},
		queryData: map[string][]map[string]interface{}{
			// 生成列已被读路径剔除，目标主键反查不会带 contract_company。
			// 目标行 id=1 已存在且 company 不同，应走 UPDATE。
			"SELECT `id`, `contract_no`, `company` FROM `target_db`.`crm_filing_offline` WHERE `id` IN (1, 2)": {
				{"id": int64(1), "contract_no": "HT-1", "company": "旧公司"},
			},
		},
	}}
}

// 回归：crm_filing_offline 目标表带 VIRTUAL GENERATED 列时分页差异同步把
// 生成列名拼进 INSERT，MySQL 报 Error 3105。写路径必须用可写列集合过滤。
func TestRunSyncPagedDiffSkipsGeneratedColumnsInWrites(t *testing.T) {
	source := generatedColumnSourceDB()
	target := generatedColumnTargetDB()
	useSyncDatabaseFactorySequence(t,
		syncDatabaseFactoryStep{db: source},
		syncDatabaseFactoryStep{db: target},
	)

	result := NewSyncEngine(Reporter{}).RunSync(generatedColumnSyncConfig())
	if !result.Success || result.RowsInserted != 1 || result.RowsUpdated != 1 {
		t.Fatalf("RunSync() = %+v, want 1 insert and 1 update", result)
	}
	if len(target.appliedChanges.Inserts) != 1 {
		t.Fatalf("inserts = %#v, want one row", target.appliedChanges.Inserts)
	}
	insert := target.appliedChanges.Inserts[0]
	if _, ok := insert["contract_company"]; ok {
		t.Fatalf("generated column leaked into insert row: %#v", insert)
	}
	if insert["contract_no"] != "HT-2" || insert["company"] != "乙公司" {
		t.Fatalf("insert row lost writable columns: %#v", insert)
	}
	if len(target.appliedChanges.Updates) != 1 {
		t.Fatalf("updates = %#v, want one row", target.appliedChanges.Updates)
	}
	update := target.appliedChanges.Updates[0]
	if _, ok := update.Values["contract_company"]; ok {
		t.Fatalf("generated column leaked into update values: %#v", update.Values)
	}
	if update.Values["company"] != "甲公司" {
		t.Fatalf("update values = %#v, want company=甲公司", update.Values)
	}
}

// 快照差异路径（无主键分页回退）同样要剔除生成列，且不能只覆盖「缺列>0」分支。
func TestRunSyncSnapshotDiffSkipsGeneratedColumnsInWrites(t *testing.T) {
	source := generatedColumnSourceDB()
	// 复合主键让分页差异路径退回快照比对，命中快照分支的写过滤。
	source.columns["source_db.crm_filing_offline"] = append(
		append([]connection.ColumnDefinition(nil), source.columns["source_db.crm_filing_offline"]...),
		connection.ColumnDefinition{Name: "filing_id", Type: "bigint", Nullable: "NO", Key: "PRI"},
	)
	source.queryData = map[string][]map[string]interface{}{
		"SELECT * FROM `source_db`.`crm_filing_offline`": {
			{"id": int64(1), "contract_no": "HT-1", "company": "甲公司", "filing_id": int64(9)},
			{"id": int64(2), "contract_no": "HT-2", "company": "乙公司", "filing_id": int64(9)},
		},
	}

	target := generatedColumnTargetDB()
	target.columns["target_db.crm_filing_offline"] = append(
		append([]connection.ColumnDefinition(nil), target.columns["target_db.crm_filing_offline"]...),
		connection.ColumnDefinition{Name: "filing_id", Type: "bigint", Nullable: "NO", Key: "PRI"},
	)
	target.queryData = map[string][]map[string]interface{}{
		"SELECT * FROM `target_db`.`crm_filing_offline`": {
			{"id": int64(1), "contract_no": "HT-1", "company": "旧公司", "contract_company": "HT-1-旧公司", "filing_id": int64(9)},
		},
	}
	useSyncDatabaseFactorySequence(t,
		syncDatabaseFactoryStep{db: source},
		syncDatabaseFactoryStep{db: target},
	)

	result := NewSyncEngine(Reporter{}).RunSync(generatedColumnSyncConfig())
	if !result.Success || result.RowsInserted != 1 || result.RowsUpdated != 1 {
		t.Fatalf("RunSync() = %+v, want 1 insert and 1 update", result)
	}
	for _, insert := range target.appliedChanges.Inserts {
		if _, ok := insert["contract_company"]; ok {
			t.Fatalf("generated column leaked into insert row: %#v", insert)
		}
	}
	for _, update := range target.appliedChanges.Updates {
		if _, ok := update.Values["contract_company"]; ok {
			t.Fatalf("generated column leaked into update values: %#v", update.Values)
		}
	}
	if !reflect.DeepEqual(target.appliedChanges.Updates[0].Keys, map[string]interface{}{"id": int64(1), "filing_id": int64(9)}) {
		t.Fatalf("update keys = %#v, want composite key preserved", target.appliedChanges.Updates[0].Keys)
	}
}

func TestBuildWritableTargetColumnSetExcludesGenerated(t *testing.T) {
	t.Parallel()

	cols := []connection.ColumnDefinition{
		{Name: "id", Extra: "auto_increment"},
		{Name: "name"},
		{Name: " full_name ", Extra: "VIRTUAL GENERATED"},
		{Name: "snapshot_col", Extra: "MATERIALIZED"},
		{Name: "expr_col", Extra: "STORED GENERATED"},
	}
	set := buildWritableTargetColumnSet(cols)
	if !reflect.DeepEqual(set, map[string]struct{}{"id": {}, "name": {}}) {
		t.Fatalf("writable set = %#v, want only id+name", set)
	}
	full := buildTargetColumnSet(cols)
	for _, name := range []string{"id", "name", "full_name", "snapshot_col", "expr_col"} {
		if _, ok := full[name]; !ok {
			t.Fatalf("existence set missing %q: %#v", name, full)
		}
	}
}

package app

import (
	"archive/zip"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestXLSXNumericCellLiteral(t *testing.T) {
	cases := []struct {
		text string
		want string
		ok   bool
	}{
		{text: "0", want: "0", ok: true},
		{text: "1", want: "1", ok: true},
		{text: "-42", want: "-42", ok: true},
		{text: "1.25", want: "1.25", ok: true},
		{text: "12.50", want: "12.5", ok: true},
		{text: "1445663", want: "1445663", ok: true},
		{text: "9007199254740993", ok: false},
		{text: "0012", ok: false},
		{text: "alice", ok: false},
		{text: "", ok: false},
		{text: "1.2.3", ok: false},
	}
	for _, tc := range cases {
		got, ok := xlsxNumericCellLiteral(tc.text)
		if ok != tc.ok || got != tc.want {
			t.Errorf("xlsxNumericCellLiteral(%q) = %q, %v; want %q, %v", tc.text, got, ok, tc.want, tc.ok)
		}
	}
}

func TestWriteRowsToFile_XLSXNumericCellsAreValues(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "numeric-*.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	data := []map[string]interface{}{{
		"id":     1,
		"amount": 12.5,
		"code":   "0012",
		"huge":   "9007199254740993",
		"name":   "alice",
	}}
	columns := []string{"id", "amount", "code", "huge", "name"}
	if err := writeRowsToFile(f, data, columns, ExportFileOptions{Format: "xlsx"}); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	sheet := readXLSXSheetXML(t, f.Name())
	for _, want := range []string{
		`<c r="A2"><v>1</v></c>`,
		`<c r="B2"><v>12.5</v></c>`,
		`<c r="C2" t="inlineStr"><is><t>0012</t></is></c>`,
		`<c r="D2" t="inlineStr"><is><t>9007199254740993</t></is></c>`,
		`<c r="E2" t="inlineStr"><is><t>alice</t></is></c>`,
	} {
		if !strings.Contains(sheet, want) {
			t.Fatalf("sheet missing %s\n%s", want, sheet)
		}
	}

	book, err := excelize.OpenFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer book.Close()
	rows, err := book.GetRows("Sheet1")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[1][0] != "1" || rows[1][1] != "12.5" || rows[1][4] != "alice" {
		t.Fatalf("rows = %#v", rows)
	}
}

func readXLSXSheetXML(t *testing.T, path string) string {
	t.Helper()
	zr, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	for _, entry := range zr.File {
		if entry.Name != "xl/worksheets/sheet1.xml" {
			continue
		}
		rc, err := entry.Open()
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatal(err)
		}
		return string(body)
	}
	t.Fatal("sheet1.xml missing")
	return ""
}

package plugin

import (
	"context"
	"encoding/json"
	"strings"
	"unicode/utf8"
)

// EncodeExporterRequest preflights typed cells before encoding the exact identity-bearing envelope.
func EncodeExporterRequest(ctx context.Context, exporterID string, snapshot ExporterSnapshot) (json.RawMessage, error) {
	if ctx == nil || !exporterLocalID.MatchString(exporterID) {
		return nil, exporterError("invalid_request")
	}
	if err := preflightExporterSnapshot(ctx, snapshot); err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(ExporterRequest{ProtocolVersion: 1, ExporterID: exporterID, Columns: snapshot.Columns, Rows: snapshot.Rows})
	if err != nil {
		return nil, exporterError("invalid_request")
	}
	if len(encoded) > ExporterMaxEnvelopeBytes {
		return nil, exporterError("limit_exceeded")
	}
	return encoded, nil
}

// ValidateExporterSnapshot checks data budgets using a minimal identity; Handle checks the actual full wire envelope.
func ValidateExporterSnapshot(ctx context.Context, snapshot ExporterSnapshot) error {
	_, err := EncodeExporterRequest(ctx, "a", snapshot)
	return err
}

func preflightExporterSnapshot(ctx context.Context, snapshot ExporterSnapshot) error {
	columns, rows := snapshot.Columns, snapshot.Rows
	if len(columns) == 0 || len(columns) > 256 || rows == nil || len(rows) > 4096 || len(rows)*len(columns) > 16384 {
		return exporterError("invalid_request")
	}
	names := make(map[string]bool, len(columns))
	size := 0
	for _, name := range columns {
		if !utf8.ValidString(name) || len(name) > 1024 || strings.TrimSpace(name) == "" || strings.ContainsRune(name, 0) || names[name] {
			return exporterError("invalid_request")
		}
		names[name] = true
		size += len(name)
	}
	budget := exportJSONBudget{nodes: 5 + len(columns) + len(rows)}
	for _, row := range rows {
		if row == nil || len(row) != len(columns) {
			return exporterError("invalid_request")
		}
		for _, cell := range row {
			budget.nodes += 2
			size += len(cell.Kind)
			if cell.Value != nil {
				size += len(*cell.Value)
				if cell.Kind != ExporterJSON {
					budget.nodes++
				}
			}
			if size > ExporterMaxEnvelopeBytes || budget.nodes > 131072 {
				return exporterError("limit_exceeded")
			}
			if err := validateExporterCell(ctx, cell, &budget); err != nil {
				return err
			}
		}
	}
	return ctx.Err()
}

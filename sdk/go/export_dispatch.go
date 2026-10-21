package plugin

import (
	"context"
	"encoding/json"
	"time"
)

// ExporterDispatcher owns a copied registry of at most eight exact exporter identities.
// It runs no detached jobs: the single RPC context owns the entire render operation.
type ExporterDispatcher struct{ handlers map[string]ExporterRender }

// NewExporterDispatcher rejects wildcard/invalid registrations and never installs a global handler.
func NewExporterDispatcher(handlers map[string]ExporterRender) (*ExporterDispatcher, error) {
	if len(handlers) == 0 || len(handlers) > 8 {
		return nil, exporterError("invalid_request")
	}
	copied := make(map[string]ExporterRender, len(handlers))
	for id, render := range handlers {
		if !exporterLocalID.MatchString(id) || render == nil {
			return nil, exporterError("invalid_request")
		}
		copied[id] = render
	}
	return &ExporterDispatcher{handlers: copied}, nil
}

// Handle validates the full envelope before calling a renderer, then validates the complete encoded reply.
func (d *ExporterDispatcher) Handle(ctx context.Context, raw json.RawMessage) (any, error) {
	if ctx == nil {
		return nil, exporterError("invalid_request")
	}
	child, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	request, err := decodeExporterRequest(child, raw)
	if err != nil {
		return nil, err
	}
	render := d.handlers[request.ExporterID]
	if render == nil {
		return nil, exporterError("unsupported")
	}
	if err := child.Err(); err != nil {
		return nil, err
	}
	rowCount := len(request.Rows)
	text, err := callExporterRender(child, ExporterSnapshot{Columns: request.Columns, Rows: request.Rows}, render)
	if child.Err() != nil {
		return nil, child.Err()
	}
	if err != nil {
		return nil, exporterError("render_failed")
	}
	return validateExporterResult(child, text, rowCount)
}

func callExporterRender(ctx context.Context, snapshot ExporterSnapshot, render ExporterRender) (text string, err error) {
	defer func() {
		if recover() != nil {
			text = ""
			err = exporterError("render_failed")
		}
	}()
	return render(ctx, snapshot)
}

func exporterError(code string) error { return &Error{Code: -32020, Message: code} }

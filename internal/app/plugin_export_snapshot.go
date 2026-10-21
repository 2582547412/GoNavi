package app

import (
	"context"
	"errors"
	"io"
	"time"

	"GoNavi-Wails/internal/connection"
	"GoNavi-Wails/internal/dataexport"
	"GoNavi-Wails/internal/plugin"
)

func (a *App) exportPluginSnapshot(ctx context.Context, data []map[string]any, columns []string, defaultName string, options ExportFileOptions) connection.QueryResult {
	if !pluginSkillManagementAllowed(a) || ctx == nil {
		return a.pluginExportFailure("unavailable")
	}
	if err := ctx.Err(); err != nil {
		return a.exportTaskCanceledResult()
	}
	if err := validatePluginExportOptions(options); err != nil {
		return a.pluginExportFailure("invalid_selection")
	}
	if !pluginExportJobID.MatchString(options.JobID) {
		return a.pluginExportFailure("invalid_job")
	}
	snapshot, err := dataexport.PreparePluginExportSnapshot(ctx, data, columns, options.Columns)
	if err != nil {
		return a.pluginExportStageFailure(ctx, "invalid_snapshot")
	}
	manager, err := PluginSkillManagerForIntegration(ctx, a)
	if err != nil {
		return a.pluginExportStageFailure(ctx, "unavailable")
	}
	exporter, err := dataexport.NewPluginExporter(ctx, manager, options.Format, *options.PluginExporter)
	if err != nil {
		return a.pluginExportStageFailure(ctx, "unavailable")
	}
	root := a.pluginRuntimeOwner().configDir
	taskCtx, finish, err := a.beginPluginExportTask(ctx, options.JobID)
	if err != nil {
		return a.pluginExportStageFailure(ctx, "busy")
	}
	defer finish()
	targetPath, err := a.pickPluginExportTarget(taskCtx, defaultName, exporter.Descriptor().Extension)
	if err != nil {
		return a.pluginExportStageFailure(taskCtx, "destination_failed")
	}
	if targetPath == "" {
		return a.exportTaskCanceledResult()
	}
	// User interaction has no artificial deadline. Only selected work is bounded.
	workCtx, cancel := context.WithTimeout(taskCtx, 20*time.Second)
	defer cancel()
	check := func() error { return a.checkPluginExport(workCtx, manager, exporter, root, targetPath) }
	if err := check(); err != nil {
		return a.pluginExportStageFailure(workCtx, "unavailable")
	}
	options.TotalRowsKnown = true
	options.TotalRowsHint = int64(len(snapshot.Rows))
	reporter := newExportProgressReporter(a, options, defaultName, targetPath)
	reporter.Start(a.appText("data_export.progress.stage.preparing_export", nil))
	renderCtx, stopRender := context.WithTimeout(workCtx, 10*time.Second)
	content, err := exporter.Render(renderCtx, snapshot)
	stopRender()
	if err != nil {
		result := a.pluginExportStageFailure(workCtx, "render_failed")
		reporter.Error(0, result.Message)
		return result
	}
	if err := check(); err != nil {
		result := a.pluginExportStageFailure(workCtx, "unavailable")
		reporter.Error(0, result.Message)
		return result
	}
	return a.publishPluginExport(workCtx, targetPath, content, int64(len(snapshot.Rows)), reporter, check)
}

func (a *App) checkPluginExport(ctx context.Context, expected *plugin.Manager, exporter *dataexport.PluginExporter, root, target string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if a.pluginRuntimeOwner().configDir != root {
		return errors.New("plugin data root changed")
	}
	current, err := PluginSkillManagerForIntegration(ctx, a)
	if err != nil {
		return err
	}
	if current != expected {
		return errors.New("plugin execution owner changed")
	}
	if err := a.validatePluginExportTarget(target); err != nil {
		return err
	}
	return exporter.Check(ctx)
}

func (a *App) publishPluginExport(ctx context.Context, path string, content []byte, rows int64, reporter *exportProgressReporter, check func() error) connection.QueryResult {
	target, err := createAtomicExportTarget(path)
	if err != nil {
		result := a.pluginExportStageFailure(ctx, "write_failed")
		reporter.Error(0, result.Message)
		return result
	}
	defer target.abort()
	written, err := target.file.Write(content)
	if err == nil && written != len(content) {
		err = io.ErrShortWrite
	}
	if err != nil {
		result := a.pluginExportStageFailure(ctx, "write_failed")
		reporter.Error(0, result.Message)
		return result
	}
	return a.finishPluginExport(ctx, target, rows, reporter, check)
}

func (a *App) finishPluginExport(ctx context.Context, target *atomicExportTarget, rows int64, reporter *exportProgressReporter, check func() error) connection.QueryResult {
	reporter.Finalizing(rows)
	if err := target.commitWithCheck(ctx, check); err != nil {
		if target.replaceAttempted {
			result := a.pluginExportFailure("commit_unknown")
			result.Data = map[string]any{"code": "commit_unknown", "outcomeUnknown": true}
			reporter.Error(rows, result.Message)
			return result
		}
		result := a.pluginExportStageFailure(ctx, "write_failed")
		reporter.Error(0, result.Message)
		return result
	}
	reporter.Done(rows)
	return connection.QueryResult{Success: true, Message: a.appText("file.backend.message.export_completed", nil)}
}

func (a *App) pluginExportStageFailure(ctx context.Context, code string) connection.QueryResult {
	if ctx != nil && ctx.Err() != nil {
		return a.exportTaskCanceledResult()
	}
	return a.pluginExportFailure(code)
}

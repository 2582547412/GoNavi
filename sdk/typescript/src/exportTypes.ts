import type { JSONValue } from './types.js';

export const EXPORTER_MAX_ENVELOPE_BYTES = 512 * 1024;
export type ExporterCell = { kind: 'missing' | 'null' } | {
 kind: 'string' | 'boolean' | 'number' | 'json' | 'binary'; value: string;
};
export type ExporterSnapshot = { readonly columns: readonly string[]; readonly rows: readonly (readonly ExporterCell[])[] };
export type ExporterRequest = { protocolVersion: 1; exporterId: string; columns: string[]; rows: ExporterCell[][] };
export type ExporterResult = { protocolVersion: 1; rowCount: number; text: string };
/** Render only supplied data. Honor signal; never retain snapshots or perform destination work. */
export type ExporterRender = (snapshot: ExporterSnapshot, signal: AbortSignal) => string | Promise<string>;
export type ExporterHandlers = Readonly<Record<string, ExporterRender>>;
export type ExporterWireValue = JSONValue;

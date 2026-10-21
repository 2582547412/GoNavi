export type JSONValue = null | boolean | number | string | JSONValue[] | { [key: string]: JSONValue };
export type PluginState = 'disabled' | 'starting' | 'running' | 'stopping' | 'faulted' | 'blocked';
export interface PluginCommand {
  id: string;
  title: string;
  description?: string;
  inputSchema?: JSONValue;
  effect?: 'read' | 'write' | 'unknown';
}
export interface UINode {
  type: 'text' | 'alert' | 'column' | 'row' | 'button' | 'table';
  text?: string;
  title?: string;
  children?: UINode[];
  command?: string;
  arguments?: JSONValue;
  columns?: { key: string; title: string }[];
  rows?: Record<string, JSONValue>[];
}
export interface PluginManifest {
  schemaVersion: number;
  apiVersion: number;
  id: string;
  name: string;
  version: string;
  description?: string;
  author?: string;
  license?: string;
  minHostVersion?: string;
  permissions: string[];
  permissionNotes?: Record<string, string>;
  requires?: { id: string; minVersion?: string; optional?: boolean }[];
  contributes?: { commands?: PluginCommand[]; panels?: { id: string; title: string; body: unknown }[]; dataSources?: { id: string; name: string; dialect: string; command: string; defaultPort?: number; readOnly: true }[]; aiProviders?: { id: string; name: string; command: string; supportsTools: boolean; supportsImages: false; models: { id: string; name: string; contextWindow: number; maxOutputTokens: number }[] }[]; exporters?: { id: string; name: string; command: string; extension: string; contentType: string }[]; skills?: { id: string; name: string; description: string; entry: string; resources?: string[] }[] };
  backend?: { default?: { executable: string; args?: string[]; runtime?: string }; platforms?: Record<string, { executable: string; args?: string[]; runtime?: string }> };
  window?: { entry: string; title?: string; width?: number; height?: number };
  configSchema?: unknown;
}
export interface PluginStatus {
  manifest: PluginManifest;
  digest: string;
  state: PluginState;
  desiredEnabled: boolean;
  generation: number;
  lastError?: string;
  installedAt: string;
}
export interface PluginInspection {
  manifest: PluginManifest;
  digest: string;
  sourcePath: string;
  sizeBytes: number;
}
export interface PluginRuntimeInfo { safeMode: boolean; managementAllowed: boolean }

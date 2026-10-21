import type { DataSourceCapabilityContract } from '../../utils/dataSourceCapabilities';
import { parsePluginDataSourceType } from './identity';
import { findPluginDataSource } from './snapshot';

export function pluginDataSourceCapabilities(type: string, directType: boolean): DataSourceCapabilityContract {
  const identity = parsePluginDataSourceType(type);
  const available = directType && Boolean(identity && findPluginDataSource(type)?.available);
  const unsupported = () => ({ supported: false, messageKey: 'plugin.datasource.read_only' });
  const readable = () => ({ supported: available, ...(available ? {} : { messageKey: 'plugin.datasource.status.runtime_unavailable' }) });
  return {
    type, query: readable(), metadata: readable(), cancel: readable(), transaction: unsupported(),
    pagination: unsupported(), schema: unsupported(), sampling: unsupported(), streaming: unsupported(), dangerousOperations: unsupported(),
    ui: { explainDiagnosis: false, sqlQueryExport: false, copyInsert: false, mockData: false, copyTable: false,
      createDatabase: false, createDatabaseCharset: false, renameDatabase: false, dropDatabase: false, messagePublish: false,
      forceReadOnlyQueryResult: true, forceReadOnlyStructureDesigner: true, userManagement: false },
    navigation: { primaryVisibilitySupported: available, primaryKind: available ? 'database' : 'none',
      secondarySchemaVisibilitySupported: false, schemaIdentifierCaseSensitive: identity?.dialect === 'postgres' },
  };
}

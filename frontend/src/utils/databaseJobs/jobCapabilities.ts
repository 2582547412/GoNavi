import type { ConnectionConfig } from '../../types';
import { resolveSidebarMetadataDialect } from '../sidebarMetadata';

export const resolveDatabaseJobDialect = (
  config: Pick<ConnectionConfig, 'type' | 'driver' | 'oceanBaseProtocol'> | undefined,
): 'oracle' | 'dm' | null => {
  const type = String(config?.type || '').trim().toLowerCase();
  if (!['oracle', 'dameng', 'dm', 'custom'].includes(type)) return null;
  const dialect = resolveSidebarMetadataDialect(type, config?.driver || '', config?.oceanBaseProtocol);
  return dialect === 'oracle' || dialect === 'dm' ? dialect : null;
};

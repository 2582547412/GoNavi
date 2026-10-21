import { Tag, Typography } from 'antd';
import { useI18n } from '../../i18n/provider';
import type { PluginManifest } from '../../plugins/types';

export function PluginDataSourceSummary({ manifest }: { manifest: PluginManifest }) {
  const { t } = useI18n();
  const sources = manifest.contributes?.dataSources ?? [];
  if (!sources.length) return null;
  return <div><Typography.Text strong>{t('plugin.datasource.group')}</Typography.Text>
    <ul>{sources.map((source) => {
      const effect = manifest.contributes?.commands?.find((command) => command.id === source.command)?.effect;
      return <li key={source.id}><code>{source.id}</code> {source.dialect} <code>{effect ? `${source.command} (${effect})` : source.command}</code>{source.readOnly === true ? <Tag>{t('plugin.datasource.read_only')}</Tag> : null}</li>;
    })}</ul></div>;
}

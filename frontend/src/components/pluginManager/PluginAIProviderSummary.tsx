import { Tag, Typography } from 'antd';
import { useI18n } from '../../i18n/provider';
import type { PluginManifest } from '../../plugins/types';

export function PluginAIProviderSummary({ manifest }: { manifest: PluginManifest }) {
  const { t } = useI18n();
  const providers = manifest.contributes?.aiProviders ?? [];
  if (!providers.length) return null;
  return <div><Typography.Text strong>{t('plugin.ai_provider.label')}</Typography.Text>
    <ul>{providers.map((provider) => <li key={provider.id}><code>{provider.id}</code> {provider.models.map((model) => model.id).join(', ')}
      {provider.supportsImages === false ? <Tag>{t('plugin.ai_provider.no_images')}</Tag> : null}</li>)}</ul></div>;
}

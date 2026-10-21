import { useRef, useState } from 'react';
import { Alert, Button, Input, Typography } from 'antd';
import { useI18n } from '../../i18n/provider';
import { pluginAPI } from '../../plugins/api';
import { pluginErrorText } from '../../plugins/apiResult';
import type { PluginInspection } from '../../plugins/types';
import { PluginInstallPreview } from './PluginInstallPreview';

interface Props { allowed: boolean; onInstalled: () => void; installedIds?: ReadonlySet<string> }
export function LocalInstallPanel({ allowed, onInstalled, installedIds }: Props) {
  const { t } = useI18n();
  const [path, setPath] = useState('');
  const [inspection, setInspection] = useState<PluginInspection | null>(null);
  const [busy, setBusy] = useState(false);
  const [installing, setInstalling] = useState(false);
  const [error, setError] = useState('');
  const request = useRef(0);
  const inspect = async () => {
    if (!path.trim() || !allowed || busy || installing) return;
    const token = ++request.current;
    setBusy(true); setError('');
    try { const result = await pluginAPI.inspect(path.trim()); if (token === request.current) setInspection(result); }
    catch (cause) { if (token === request.current) setError(pluginErrorText(cause)); }
    finally { if (token === request.current) setBusy(false); }
  };
  return <section className="plugin-local-install plugin-stack" aria-label={t('plugin.install.local')}>
    <Typography.Text strong>{t('plugin.install.local')}</Typography.Text>
    <Typography.Text type="secondary">{t('plugin.install.path_hint')}</Typography.Text>
    <div className="plugin-toolbar">
      <Input aria-label={t('plugin.install.path')} placeholder={t('plugin.install.path')} value={path} disabled={!allowed || installing}
        onChange={(event) => { request.current++; setBusy(false); setPath(event.target.value); setInspection(null); }} onPressEnter={() => void inspect()} />
      <Button onClick={() => void inspect()} loading={busy} disabled={!allowed || installing || !path.trim()}>{t('plugin.install.inspect')}</Button>
    </div>
    {installing && <Alert type="info" showIcon message={t('plugin.install.in_progress')} />}
    {error && <Alert showIcon type="error" message={error} />}
    {inspection && <PluginInstallPreview key={inspection.digest} inspection={inspection} allowed={allowed} onCancel={() => setInspection(null)}
      replace={installedIds?.has(inspection.manifest.id) === true}
      onInstall={async () => {
        setInstalling(true); setError('');
        const replacing = installedIds?.has(inspection.manifest.id) === true;
        try {
          if (replacing) await pluginAPI.replace(inspection.sourcePath, inspection.digest, true);
          else await pluginAPI.install(inspection.sourcePath, inspection.digest, true);
          setInspection(null);
        }
        catch (cause) { setError(pluginErrorText(cause)); throw cause; }
        finally { setInstalling(false); onInstalled(); }
      }} />}
  </section>;
}

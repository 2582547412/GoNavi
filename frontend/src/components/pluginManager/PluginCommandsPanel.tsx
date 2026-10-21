import { useState } from 'react';
import { Alert, Button, Collapse, Input, Select, Tabs, Typography } from 'antd';
import { useI18n } from '../../i18n/provider';
import { boundedJSON } from '../../plugins/jsonBoundary';
import { DeclarativePanel } from '../../plugins/DeclarativePanel';
import { usePluginInvocation } from '../../plugins/usePluginInvocation';
import type { PluginStatus } from '../../plugins/types';
import { publicPluginCommands } from '../../plugins/dataSource/reservedCommands';

export function PluginCommandsPanel({ plugin, allowed }: { plugin: PluginStatus; allowed: boolean }) {
  const { t } = useI18n();
  const commands = publicPluginCommands(plugin.manifest);
  const [command, setCommand] = useState<string | undefined>(commands[0]?.id);
  const [args, setArgs] = useState('{}');
  const [validationError, setValidationError] = useState(false);
  const invocation = usePluginInvocation(plugin.manifest.id, plugin.generation, allowed);
  const run = () => {
    if (!command || !commands.some((item) => item.id === command)) return;
    try { const value = boundedJSON(JSON.parse(args)); setValidationError(false); void invocation.invoke(command, value); }
    catch { setValidationError(true); }
  };
  return <div className="plugin-stack">
    {plugin.manifest.window && <Button disabled={!allowed} onClick={() => void invocation.openWindow()}>{t('plugin.action.open_window')}</Button>}
    {commands.length > 0 && <Collapse items={[{ key: 'commands', label: t('plugin.commands.title'), children: <div className="plugin-stack">
      <Select aria-label={t('plugin.commands.select')} value={command} onChange={setCommand} options={commands.map((item) => ({ label: item.title, value: item.id }))} />
      <Input.TextArea aria-label={t('plugin.commands.arguments')} value={args} onChange={(event) => setArgs(event.target.value)} maxLength={65536} autoSize={{ minRows: 2, maxRows: 8 }} />
      {validationError && <Alert type="error" message={t('plugin.commands.invalid_json')} />}
      <Typography.Text type="secondary">{t('plugin.commands.effect_warning')}</Typography.Text>
      <Button onClick={run} loading={invocation.busy} disabled={!allowed || !command}>{t('plugin.commands.run')}</Button>
    </div> }]} />}
    {Boolean(plugin.manifest.contributes?.panels?.length) && <Tabs destroyInactiveTabPane items={(plugin.manifest.contributes?.panels ?? []).map((panel) => ({
      key: panel.id, label: panel.title, children: <section aria-label={panel.title} className="plugin-embedded-panel">
        <DeclarativePanel body={panel.body} commands={commands} disabled={!allowed || invocation.busy} invoke={(id, value) => void invocation.invoke(id, value)} />
      </section>,
    }))} />}
    {invocation.error && <Alert type="error" showIcon message={invocation.error} />}
    {invocation.result && <pre className="plugin-result" aria-label={t('plugin.commands.result')}>{invocation.result}</pre>}
  </div>;
}

import { Button, Switch } from 'antd';
import { useId } from 'react';
import { useBackgroundImageStore } from '../../backgroundImageStore';
import { useI18n } from '../../i18n/provider';
import { BUILTIN_BACKGROUNDS } from '../../utils/builtinBackgrounds';
import BackgroundArtworkNotice from './BackgroundArtworkNotice';
import './BuiltinBackgroundSettings.css';

interface Props {
  disabled: boolean;
  runAction: (action: () => void) => void;
}

export default function BuiltinBackgroundSettings({ disabled, runAction }: Props) {
  const { t } = useI18n();
  const id = useId();
  const enabled = useBackgroundImageStore((state) => state.enabled);
  const builtinId = useBackgroundImageStore((state) => state.builtinId);
  const image = useBackgroundImageStore((state) => state.image);
  const setEnabled = useBackgroundImageStore((state) => state.setEnabled);
  const selectBuiltin = useBackgroundImageStore((state) => state.selectBuiltin);
  return (
    <section className="gonavi-builtin-backgrounds" aria-label={t('app.theme.background.builtin.title')}>
      <div className="gonavi-background-enable">
        <label htmlFor={id}>{t('app.theme.background.enabled')}</label>
        <Switch id={id} checked={enabled} disabled={disabled || (!image && !builtinId)}
          aria-describedby={`${id}-hint`} onChange={(value) => runAction(() => setEnabled(value))} />
      </div>
      <p id={`${id}-hint`} className="gonavi-background-image-hint">
        {t('app.theme.background.enabled_hint')}
      </p>
      <div className="gonavi-builtin-background-options">
        {BUILTIN_BACKGROUNDS.map((background) => (
          <Button key={background.id} className="gonavi-builtin-background-option"
            disabled={disabled} aria-pressed={builtinId === background.id}
            onClick={() => runAction(() => selectBuiltin(background.id))}>
            <span className="gonavi-builtin-background-thumbnail" aria-hidden="true"
              style={{ backgroundImage: `url("${background.url}")` }} />
            <span>{t('app.theme.background.builtin.use', { name: t(background.nameKey) })}</span>
          </Button>
        ))}
      </div>
      <BackgroundArtworkNotice />
    </section>
  );
}

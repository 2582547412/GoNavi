import type { MouseEvent } from 'react';
import { BrowserOpenURL } from '../../../wailsjs/runtime';
import { useI18n } from '../../i18n/provider';
import { BACKGROUND_ARTWORK_LICENSE, BACKGROUND_ARTWORK_SOURCE } from '../../utils/builtinBackgrounds';
import notice from '../../assets/backgrounds/whale-maid/NOTICE?raw';
import license from '../../assets/backgrounds/whale-maid/LICENSE-ARTWORK?raw';

function openLink(event: MouseEvent<HTMLAnchorElement>, url: string) {
  // Leave ordinary anchors to the browser in Web mode; use the OS browser in Wails.
  if (!window.runtime) return;
  event.preventDefault();
  BrowserOpenURL(url);
}

export default function BackgroundArtworkNotice() {
  const { t } = useI18n();
  return (
    <div className="gonavi-background-artwork-notice">
      <p className="gonavi-background-image-hint">{t('app.theme.background.builtin.attribution')}</p>
      <p className="gonavi-background-image-hint">{t('app.theme.background.builtin.terms')}</p>
      <div className="gonavi-background-image-actions">
        <a href={BACKGROUND_ARTWORK_SOURCE} target="_blank" rel="noopener noreferrer"
          onClick={(event) => openLink(event, BACKGROUND_ARTWORK_SOURCE)}>
          {t('app.theme.background.builtin.source')}
        </a>
        <a href={BACKGROUND_ARTWORK_LICENSE} target="_blank" rel="noopener noreferrer"
          onClick={(event) => openLink(event, BACKGROUND_ARTWORK_LICENSE)}>
          {t('app.theme.background.builtin.license')}
        </a>
      </div>
      <details>
        <summary>{t('app.theme.background.builtin.full_notice')}</summary>
        <pre>{notice}</pre>
        <pre>{license}</pre>
      </details>
    </div>
  );
}

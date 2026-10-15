// @vitest-environment jsdom
import { Alert, Button, Slider, Switch } from 'antd';
import { act, create, type ReactTestRenderer } from 'react-test-renderer';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useBackgroundImageStore as store } from '../../backgroundImageStore';
import { I18nProvider } from '../../i18n/provider';
import { BUILTIN_BACKGROUNDS, BACKGROUND_ARTWORK_LICENSE, BACKGROUND_ARTWORK_SOURCE } from '../../utils/builtinBackgrounds';
import { prepareBackgroundImage } from '../../utils/backgroundImage';
import BackgroundImageSettings from './BackgroundImageSettings';
import BackgroundImageStyleHost from '../theme/BackgroundImageStyleHost';

vi.mock('../../utils/backgroundImage', async (original) => ({
  ...await original<typeof import('../../utils/backgroundImage')>(), prepareBackgroundImage: vi.fn(),
}));
let view: ReactTestRenderer;
let root: HTMLElement;
const builtinButton = (index: number) => view.root.findAllByType(Button)
  .filter((button) => button.props.className === 'gonavi-builtin-background-option')[index];
const toggle = () => view.root.findByType(Switch);
const mount = () => act(() => {
  view = create(<I18nProvider preference="en-US" onPreferenceChange={() => undefined}>
    <BackgroundImageStyleHost /><BackgroundImageSettings />
  </I18nProvider>);
});
beforeEach(() => {
  localStorage.clear();
  store.getState().reload();
  root = document.createElement('main');
  root.className = 'gn-v2-app-root';
  document.body.append(root);
});
afterEach(() => {
  act(() => view?.unmount());
  document.body.replaceChildren();
  vi.restoreAllMocks();
});

describe('builtin background settings and style integration', () => {
  it('starts unchecked without a selected image and explains opt-in behavior', () => {
    mount();
    expect(toggle().props).toMatchObject({ checked: false, disabled: true });
    expect(builtinButton(0).props['aria-pressed']).toBe(false);
    expect(builtinButton(1).props['aria-pressed']).toBe(false);
    expect(JSON.stringify(view.toJSON())).toContain('Off by default');
    expect(root.hasAttribute('data-background-image')).toBe(false);
  });

  it.each([0, 1])('selects builtin %s explicitly, previews contain, persists ID and restores after disable', (index) => {
    mount();
    expect(builtinButton(index).findAllByType('span').some((span) =>
      typeof span.props.children === 'string' && span.props.children.startsWith('Select and enable'))).toBe(true);
    act(() => builtinButton(index).props.onClick());
    const selected = BUILTIN_BACKGROUNDS[index];
    expect(store.getState()).toMatchObject({ builtinId: selected.id, enabled: true, image: null });
    expect(builtinButton(index).props['aria-pressed']).toBe(true);
    expect(view.root.findByType('img').props).toMatchObject({ src: selected.url, style: { objectFit: 'contain' } });
    expect(root.style.getPropertyValue('--gn-background-size')).toBe('contain');
    expect(root.style.getPropertyValue('--gn-background-image')).toBe(`url("${selected.url}")`);
    const axes = view.root.findAllByType(Slider).filter((slider) => slider.props.min === -100);
    act(() => axes[0].props.onChange(12));
    expect(root.style.getPropertyValue('--gn-background-offset-x')).toBe('12vw');
    act(() => axes[0].props.onChangeComplete());
    act(() => view.root.findAllByType(Slider).find((slider) => slider.props.min === 5)!.props.onChange(25));
    expect(root.style.getPropertyValue('--gn-background-tint')).toBe('75%');
    act(() => toggle().props.onChange(false));
    expect(root.hasAttribute('data-background-image')).toBe(false);
    for (const property of ['image', 'size', 'tint', 'offset-x', 'offset-y']) {
      expect(root.style.getPropertyValue(`--gn-background-${property}`)).toBe('');
    }
    expect(view.root.findByType('img').props.src).toBe(selected.url);
    act(() => store.getState().reload());
    expect(toggle().props.checked).toBe(false);
    act(() => toggle().props.onChange(true));
    expect(root.style.getPropertyValue('--gn-background-offset-x')).toBe('12vw');
    expect(root.style.getPropertyValue('--gn-background-tint')).toBe('75%');
  });

  it('leaves migrated uploads disabled while retaining the editable preview', () => {
    localStorage.setItem('gonavi-background-image-v1', JSON.stringify({
      version: 1, image: { name: 'legacy.jpg', dataUrl: 'data:image/jpeg;base64,/9j//9k=' }, opacity: 24,
    }));
    store.getState().reload();
    mount();
    expect(toggle().props).toMatchObject({ checked: false, disabled: false });
    expect(view.root.findByType('img').props.style.objectFit).toBe('cover');
    expect(root.hasAttribute('data-background-image')).toBe(false);
    act(() => toggle().props.onChange(true));
    expect(root.style.getPropertyValue('--gn-background-size')).toBe('cover');
  });

  it('replaces a builtin with a safely prepared upload and enables the cover preview', async () => {
    store.getState().selectBuiltin('whale-maid-a');
    store.getState().setEnabled(false);
    const image = { name: 'new.jpg', dataUrl: 'data:image/jpeg;base64,/9j//9k=' };
    vi.mocked(prepareBackgroundImage).mockResolvedValue(image);
    mount();
    const file = new File(['image'], image.name, { type: 'image/jpeg' });
    await act(async () => view.root.findByType('input').props.onChange({ currentTarget: { files: [file], value: '' } }));
    expect(prepareBackgroundImage).toHaveBeenCalledWith(file);
    expect(store.getState()).toMatchObject({ enabled: true, builtinId: null, image });
    expect(root.style.getPropertyValue('--gn-background-size')).toBe('cover');
  });

  it.each(['select', 'toggle'])('reports %s storage failure without losing the selected background', (action) => {
    store.getState().selectBuiltin('whale-maid-a');
    mount();
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => { throw new Error('quota'); });
    act(() => action === 'select' ? builtinButton(1).props.onClick() : toggle().props.onChange(false));
    expect(store.getState()).toMatchObject({ enabled: true, builtinId: 'whale-maid-a' });
    expect(view.root.findByType(Alert).props.message).toContain('Unable to save background settings locally');
  });

  it('provides full offline attribution and license, plus public source and license links', () => {
    mount();
    const links = view.root.findAllByType('a');
    expect(links.map((link) => link.props.href)).toEqual([BACKGROUND_ARTWORK_SOURCE, BACKGROUND_ARTWORK_LICENSE]);
    const text = view.root.findAllByType('pre').map((node) => node.props.children).join('\n');
    for (const value of ['上善', 'ZipZipPipe', 'Small-tailqwq', 'GPT Image 2', '62155430', '18604994',
      'Section 8 -- Interpretation.', 'NOT covered by GoNavi', 'OpenAI imagegen']) expect(text).toContain(value);
    expect(JSON.stringify(view.toJSON())).toContain('Non-commercial use only');
  });
});

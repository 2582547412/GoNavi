import maidLeft from '../assets/backgrounds/whale-maid/maid-left.webp';
import maidRight from '../assets/backgrounds/whale-maid/maid-right.webp';

export const BUILTIN_BACKGROUNDS = [
  { id: 'whale-maid-a', url: maidLeft, nameKey: 'app.theme.background.builtin.maid_a' },
  { id: 'whale-maid-b', url: maidRight, nameKey: 'app.theme.background.builtin.maid_b' },
] as const;

export type BuiltinBackgroundId = typeof BUILTIN_BACKGROUNDS[number]['id'];

// Only IDs from this bundled catalog are accepted; persisted URLs are never trusted.
export function getBuiltinBackground(id: unknown) {
  return BUILTIN_BACKGROUNDS.find((background) => background.id === id) ?? null;
}

export const BACKGROUND_ARTWORK_SOURCE = 'https://github.com/Small-tailqwq/dsh-deep-whale';
export const BACKGROUND_ARTWORK_LICENSE = 'https://creativecommons.org/licenses/by-nc-sa/4.0/';

import { sanitizeBackgroundImage } from './backgroundImage';
import { getBuiltinBackground } from './builtinBackgrounds';

export interface ResolvedBackground {
  dataUrl: string;
  name: string;
  nameKey?: string;
  fit: 'contain' | 'cover';
}

// Resolve the selection even while disabled so settings can still preview/edit it.
// This is shared by the settings and the style host, not a general URL escape hatch.
export function resolvedBackground(selection: { builtinId?: unknown; image: unknown }): ResolvedBackground | null {
  const builtin = getBuiltinBackground(selection.builtinId);
  if (builtin) return { dataUrl: builtin.url, name: '', nameKey: builtin.nameKey, fit: 'contain' };
  const image = sanitizeBackgroundImage(selection.image);
  return image ? { ...image, fit: 'cover' } : null;
}

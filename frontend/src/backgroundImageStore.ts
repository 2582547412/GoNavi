import { create } from 'zustand';
import {
  BackgroundImageError,
  prepareBackgroundImage,
  sanitizeBackgroundImage,
  type BackgroundImage,
} from './utils/backgroundImage';

import { normalizeBackgroundPosition, type BackgroundImagePosition } from './utils/backgroundImagePosition';
import { getBuiltinBackground, type BuiltinBackgroundId } from './utils/builtinBackgrounds';

export const BACKGROUND_IMAGE_STORAGE_KEY = 'gonavi-background-image-v1';
const DEFAULT_OPACITY = 15;

interface BackgroundImageSnapshot extends BackgroundImagePosition {
  enabled: boolean;
  builtinId: BuiltinBackgroundId | null;
  image: BackgroundImage | null;
  opacity: number;
}

export interface BackgroundImageState extends BackgroundImageSnapshot {
  // Transient offsets are never serialized with the image; commit once per gesture.
  previewPosition: BackgroundImagePosition | null;
  previewPositionAt: (x: number, y: number) => void;
  cancelPositionPreview: () => void;
  setPosition: (x: number, y: number) => void;
  importImage: (file: File) => Promise<void>;
  selectBuiltin: (id: BuiltinBackgroundId) => void;
  setEnabled: (enabled: boolean) => void;
  clearImage: () => void;
  reset: () => void;
  setOpacity: (value: number) => void;
  reload: () => void;
}

const defaults = (): BackgroundImageSnapshot => ({
  image: null, opacity: DEFAULT_OPACITY, positionX: 0, positionY: 0, enabled: false, builtinId: null,
});
const snapshotOf = ({ image, opacity, positionX, positionY, enabled, builtinId }: BackgroundImageSnapshot): BackgroundImageSnapshot => ({
  image, opacity, positionX, positionY, enabled, builtinId,
});
const normalizeOpacity = (value: number): number => Number.isFinite(value)
  ? Math.min(30, Math.max(5, value)) : DEFAULT_OPACITY;

const loadSnapshot = (): BackgroundImageSnapshot => {
  try {
    const serialized = globalThis.localStorage.getItem(BACKGROUND_IMAGE_STORAGE_KEY);
    if (!serialized) return defaults();
    const raw: unknown = JSON.parse(serialized);
    if (!raw || typeof raw !== 'object') return defaults();
    const snapshot = raw as Record<string, unknown>;
    if (snapshot.version !== 1) return defaults();
    const builtinId = getBuiltinBackground(snapshot.builtinId)?.id ?? null;
    const image = builtinId ? null : sanitizeBackgroundImage(snapshot.image);
    if (!builtinId && snapshot.image != null && !image) return defaults();
    const selected = Boolean(image || builtinId);
    return {
      image, builtinId,
      // Legacy snapshots keep the selection and adjustments, but never opt in.
      enabled: selected && snapshot.enabled === true,
      opacity: typeof snapshot.opacity === 'number' ? normalizeOpacity(snapshot.opacity) : DEFAULT_OPACITY,
      ...normalizeBackgroundPosition(selected ? snapshot.positionX : 0, selected ? snapshot.positionY : 0),
    };
  } catch {
    return defaults();
  }
};

const persistSnapshot = (snapshot: BackgroundImageSnapshot): void => {
  try {
    globalThis.localStorage.setItem(BACKGROUND_IMAGE_STORAGE_KEY, JSON.stringify({ version: 1, ...snapshot }));
  } catch {
    throw new BackgroundImageError('storage-failed');
  }
};

export const useBackgroundImageStore = create<BackgroundImageState>((set, get) => {
  let importRevision = 0;
  return {
    ...loadSnapshot(),
    previewPosition: null,
    previewPositionAt: (x, y) => set({ previewPosition: normalizeBackgroundPosition(x, y) }),
    cancelPositionPreview: () => set({ previewPosition: null }),
    setPosition: (x, y) => {
      const next = { ...snapshotOf(get()), ...normalizeBackgroundPosition(x, y) };
      try {
        persistSnapshot(next);
      } catch (error) {
        // Revert the entire window to the last committed position on save failure.
        set({ previewPosition: null });
        throw error;
      }
      set({ ...next, previewPosition: null });
    },
    importImage: async (file) => {
      const revision = ++importRevision;
      const image = await prepareBackgroundImage(file);
      // A later import, clear or reload takes precedence over pending decoding.
      if (revision !== importRevision) return;
      const next = { image, opacity: get().opacity, positionX: 0, positionY: 0, enabled: true, builtinId: null };
      persistSnapshot(next);
      set({ ...next, previewPosition: null });
    },
    selectBuiltin: (id) => {
      const builtin = getBuiltinBackground(id);
      if (!builtin) throw new BackgroundImageError('invalid-image');
      const next = { ...snapshotOf(get()), image: null, builtinId: builtin.id, enabled: true };
      persistSnapshot(next);
      importRevision += 1;
      set({ ...next, previewPosition: null });
    },
    setEnabled: (enabled) => {
      const current = snapshotOf(get());
      const next = { ...current, enabled: enabled && Boolean(current.image || current.builtinId) };
      persistSnapshot(next);
      importRevision += 1;
      set({ ...next, previewPosition: null });
    },
    clearImage: () => {
      const next = { ...defaults(), opacity: get().opacity };
      persistSnapshot(next);
      importRevision += 1;
      set({ ...next, previewPosition: null });
    },
    reset: () => {
      const next = defaults();
      persistSnapshot(next);
      importRevision += 1;
      set({ ...next, previewPosition: null });
    },
    setOpacity: (value) => {
      const next = { ...snapshotOf(get()), opacity: normalizeOpacity(value) };
      persistSnapshot(next);
      set(next);
    },
    reload: () => {
      importRevision += 1;
      set({ ...loadSnapshot(), previewPosition: null });
    },
  };
});

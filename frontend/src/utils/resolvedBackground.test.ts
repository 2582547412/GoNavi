import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';
import { createHash } from 'node:crypto';
import { resolve } from 'node:path';
import { resolvedBackground } from './resolvedBackground';
import { BUILTIN_BACKGROUNDS } from './builtinBackgrounds';

const image = { name: 'photo.jpg', dataUrl: 'data:image/jpeg;base64,/9j//9k=' };
describe('trusted background resolution', () => {
  it.each(BUILTIN_BACKGROUNDS)('resolves $id with a static URL and contain fit', (builtin) => {
    expect(resolvedBackground({ builtinId: builtin.id, image })).toEqual({
      dataUrl: builtin.url, name: '', nameKey: builtin.nameKey, fit: 'contain',
    });
    expect(builtin.url).not.toMatch(/^(https?:|data:)/);
  });
  it('keeps uploaded JPEG validation and cover sizing', () => {
    expect(resolvedBackground({ image })).toEqual({ ...image, fit: 'cover' });
    expect(resolvedBackground({ image: null })).toBeNull();
  });
  it.each(['https://evil.example/a.webp', 'data:image/svg+xml,<svg/>', 'blob:local', '__proto__'])('cannot inject URL %s', (url) => {
    expect(resolvedBackground({ builtinId: url, image: null })).toBeNull();
    expect(resolvedBackground({ image: { ...image, dataUrl: url } })).toBeNull();
  });
  it.each([
    ['maid-left.webp', '2818e5359596b13c5d64f5b1c5d4a9e0c05be477824779af8af37a0e10016b07'],
    ['maid-right.webp', '7c5a5493e84418bc0435013c8605222718923872c15eb46fdd36158d201c0b1e'],
  ])('retains original transparent WebP bytes for %s', (file, hash) => {
    const data = readFileSync(resolve(process.cwd(), 'src/assets/backgrounds/whale-maid', file));
    expect(createHash('sha256').update(data).digest('hex')).toBe(hash);
    expect(data.subarray(0, 4).toString()).toBe('RIFF');
    expect(data.subarray(8, 12).toString()).toBe('WEBP');
    expect(data.length).toBeLessThan(1.3 * 1024 * 1024);
  });
});

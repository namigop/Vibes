/**
 * There is no user table in the contract — `assignee` is free text. To keep a
 * person visually recognisable across the board we derive a stable colour from
 * their name with FNV-1a, so "Ada Lovelace" is always the same hue.
 */
export function hashName(name: string): number {
  let hash = 0x811c9dc5; // FNV-1a 32-bit offset basis
  const value = name.trim().toLowerCase();
  for (let i = 0; i < value.length; i += 1) {
    hash ^= value.charCodeAt(i);
    hash = Math.imul(hash, 0x01000193);
  }
  return hash >>> 0;
}

/** Deterministic hue in degrees for an assignee name. */
export function hueForName(name: string): number {
  return hashName(name) % 360;
}

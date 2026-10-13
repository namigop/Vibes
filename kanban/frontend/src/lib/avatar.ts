/**
 * `assignee` is free text and deliberately NOT a user reference, even though
 * the app now has a users table.
 *
 * Boards are owner-only, so the only person who can open a given card is its
 * owner. An FK here would name someone who by construction cannot see the
 * card. Assignee instead names a person involved in the work being tracked —
 * a deal contact, a colleague with no account.
 *
 * Do not "fix" this into a foreign key without also introducing board sharing.
 *
 * To keep a name visually recognisable across the board we derive a stable
 * colour from it with FNV-1a, so "Ada Lovelace" is always the same hue.
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

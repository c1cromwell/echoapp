export type BarQuery =
  | { kind: "empty" }
  | { kind: "emailNotLive" }
  | { kind: "username"; handle: string };

/** Same rules as `BarSession.interpret` in the iOS/Mac window. */
export function interpretBarQuery(raw: string): BarQuery {
  const trimmed = raw.trim();
  if (!trimmed) return { kind: "empty" };
  if (trimmed.includes("@") && !trimmed.startsWith("@")) {
    return { kind: "emailNotLive" };
  }
  const handle = (trimmed.startsWith("@") ? trimmed.slice(1) : trimmed).trim();
  if (!handle || handle.includes("@")) return { kind: "empty" };
  return { kind: "username", handle };
}

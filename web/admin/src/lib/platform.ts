/**
 * Returns the primary keyboard modifier label for the current platform.
 * - macOS / iOS family: ⌘
 * - Windows / Linux / others: Ctrl
 */
export function getPrimaryModifierKeyLabel(): "⌘" | "Ctrl" {
  if (typeof navigator === "undefined") return "⌘";

  const platform = `${navigator.platform || ""} ${navigator.userAgent || ""}`.toLowerCase();
  return /mac|iphone|ipad|ipod/.test(platform) ? "⌘" : "Ctrl";
}

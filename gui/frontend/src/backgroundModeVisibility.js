export function sectionsForBackgroundMode(mode) {
  switch (mode) {
    case "transparent":
      return { showForeground: true, showBackground: false };
    case "cutout":
      return { showForeground: false, showBackground: true };
    case "solid":
    default:
      return { showForeground: true, showBackground: true };
  }
}

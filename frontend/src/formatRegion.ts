export function formatRegion({ admin1, country }: { admin1: string; country: string }): string {
  return [admin1, country].filter(Boolean).join(', ')
}

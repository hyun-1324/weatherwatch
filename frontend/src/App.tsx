import { LocationCard } from '@/components/LocationCard'
import { PlaceSearch } from '@/components/PlaceSearch'
import { Button } from '@/components/ui/button'
import { errorMessage } from '@/errorMessage'
import { useLocations } from '@/hooks'

export function App() {
  const locations = useLocations()
  const savedExternalIds = new Set(locations.data?.map((l) => l.externalId))

  return (
    <main className="mx-auto max-w-6xl px-4 py-8 sm:py-12">
      <header className="mb-6">
        <h1 className="text-3xl font-bold tracking-tight">Weather Watchlist</h1>
        <p className="mt-1 text-muted-foreground">Save cities and see their 7-day forecast.</p>
      </header>

      <PlaceSearch savedExternalIds={savedExternalIds} />

      {locations.isPending && <p className="text-muted-foreground">Loading…</p>}
      {locations.isError && (
        <div className="flex items-center gap-3 text-sm text-destructive">
          {errorMessage(locations.error, 'Could not load your watchlist. Please try again.')}
          <Button variant="outline" size="sm" onClick={() => locations.refetch()}>
            Retry
          </Button>
        </div>
      )}
      {locations.data?.length === 0 && (
        <p className="mt-16 text-center text-muted-foreground">
          No saved locations yet. Search for a city above to add one.
        </p>
      )}

      <div className="grid grid-cols-[repeat(auto-fill,minmax(320px,1fr))] gap-4">
        {locations.data?.map((loc) => (
          <LocationCard key={loc.id} location={loc} />
        ))}
      </div>
    </main>
  )
}

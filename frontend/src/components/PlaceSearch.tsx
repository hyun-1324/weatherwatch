import { Check, Plus, Search } from 'lucide-react'
import { useState } from 'react'

import type { Place } from '@/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { errorMessage } from '@/errorMessage'
import { formatRegion } from '@/formatRegion'
import { MIN_QUERY_LENGTH, useAddLocation, usePlaceSearch } from '@/hooks'
import { useDebouncedValue } from '@/lib/useDebouncedValue'

interface PlaceSearchProps {
  savedExternalIds: Set<number>
}

export function PlaceSearch({ savedExternalIds }: PlaceSearchProps) {
  const [query, setQuery] = useState('')
  const debouncedQuery = useDebouncedValue(query)
  const search = usePlaceSearch(debouncedQuery)
  const add = useAddLocation()

  function handleAdd(place: Place) {
    add.mutate(place, { onSuccess: () => setQuery('') })
  }

  const showResults = debouncedQuery.trim().length >= MIN_QUERY_LENGTH && query.trim().length >= MIN_QUERY_LENGTH

  return (
    <section className="mb-8 max-w-lg">
      <div className="relative">
        <Search className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          type="search"
          className="pl-9"
          placeholder="Search for a city, e.g. Helsinki"
          aria-label="Search for a city"
          value={query}
          onChange={(e) => {
            setQuery(e.target.value)
            add.reset() // a save error belongs to the previous search
          }}
        />
      </div>

      {add.isError && (
        <p className="mt-2 text-sm text-destructive">
          {errorMessage(add.error, 'Could not save this city. Please try again.')}
        </p>
      )}

      {showResults && (
        <ul className="mt-2 divide-y rounded-lg border bg-card text-sm shadow-sm">
          {search.isPending && <li className="px-3 py-2.5 text-muted-foreground">Searching…</li>}
          {search.isError && (
            <li className="px-3 py-2.5 text-destructive">
              {errorMessage(search.error, 'Could not search for cities. Please try again.')}
            </li>
          )}
          {search.data?.length === 0 && <li className="px-3 py-2.5 text-muted-foreground">No matches</li>}
          {search.data?.map((place) => {
            const saved = savedExternalIds.has(place.id)
            return (
              <li key={place.id} className="flex items-center justify-between gap-2 px-3 py-2">
                <span className="truncate">
                  <span className="font-medium">{place.name}</span>
                  <span className="text-muted-foreground"> {formatRegion(place)}</span>
                </span>
                <Button
                  size="sm"
                  variant={saved ? 'secondary' : 'default'}
                  onClick={() => handleAdd(place)}
                  disabled={saved || add.isPending}
                >
                  {saved ? <Check aria-hidden /> : <Plus aria-hidden />}
                  {saved ? 'Saved' : 'Add'}
                </Button>
              </li>
            )
          })}
        </ul>
      )}
    </section>
  )
}

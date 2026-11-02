import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { api } from '@/api'

// Matches the backend's minimum; shorter queries are not sent at all.
export const MIN_QUERY_LENGTH = 2

const keys = {
  locations: ['locations'] as const,
  forecast: (id: number) => ['forecast', id] as const,
  places: (q: string) => ['places', q] as const,
}

export function useLocations() {
  return useQuery({ queryKey: keys.locations, queryFn: api.listLocations })
}

export function useForecast(locationId: number) {
  return useQuery({
    queryKey: keys.forecast(locationId),
    queryFn: () => api.forecast(locationId),
    staleTime: 10 * 60 * 1000,
    refetchInterval: 15 * 60 * 1000, // keeps an open dashboard current
  })
}

export function usePlaceSearch(query: string) {
  const q = query.trim()
  return useQuery({
    queryKey: keys.places(q),
    queryFn: () => api.searchPlaces(q),
    enabled: q.length >= MIN_QUERY_LENGTH,
    placeholderData: keepPreviousData, // old results stay visible while the new query loads
  })
}

// Refetch even on failure: a 409 or 404 means the list changed in another tab.
export function useAddLocation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: api.addLocation,
    onSettled: () => qc.invalidateQueries({ queryKey: keys.locations }),
  })
}

export function useRemoveLocation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: api.removeLocation,
    onSettled: () => qc.invalidateQueries({ queryKey: keys.locations }),
  })
}

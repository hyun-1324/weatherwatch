import { ApiError } from '@/api'

const messages = new Map<string, string>([
  ['QUERY_TOO_SHORT', 'Enter at least two characters to search for a city.'],
  ['LOCATION_ALREADY_SAVED', 'This city is already in your watchlist.'],
  ['LOCATION_NOT_FOUND', 'This city is no longer in your watchlist.'],
  ['PLACE_SEARCH_UNAVAILABLE', 'City search is temporarily unavailable. Please try again later.'],
  ['FORECAST_UNAVAILABLE', 'Weather information is temporarily unavailable. Please try again later.'],
])

export function errorMessage(error: unknown, fallback: string): string {
  return error instanceof ApiError ? (messages.get(error.code) ?? fallback) : fallback
}

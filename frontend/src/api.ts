export interface Place {
  id: number
  name: string
  country: string
  admin1: string
  latitude: number
  longitude: number
}

export interface SavedLocation {
  id: number
  externalId: number
  name: string
  country: string
  admin1: string
  latitude: number
  longitude: number
  createdAt: string
}

export interface Forecast {
  timezone: string
  current: {
    time: string
    temperature: number
    apparentTemperature: number
    humidity: number
    windSpeed: number
    weatherCode: number
    condition: string
    isDay: boolean
  }
  daily: {
    date: string
    weatherCode: number
    condition: string
    tempMax: number
    tempMin: number
    precipitationProbability: number | null
  }[]
}

interface ErrorBody {
  statusCode: number
  code: string // stable identifier
  message: string // diagnostic explanation for API consumers; never render directly
}

export class ApiError extends Error {
  readonly status: number
  readonly code: string

  constructor(status: number, code: string, message: string) {
    super(message)
    this.status = status
    this.code = code
  }
}

// Error bodies may come from a proxy, so keep only fields with the expected types.
function parseErrorBody(body: unknown): Partial<ErrorBody> {
  if (typeof body !== 'object' || body === null) return {}
  return {
    code: 'code' in body && typeof body.code === 'string' ? body.code : undefined,
    message: 'message' in body && typeof body.message === 'string' ? body.message : undefined,
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    ...init,
    headers: { 'Content-Type': 'application/json', ...init?.headers },
  })
  if (!res.ok) {
    const body: unknown = await res.json().catch(() => null)
    const error = parseErrorBody(body)
    throw new ApiError(res.status, error.code ?? 'UNKNOWN', error.message ?? `Request failed (${res.status})`)
  }
  return res.status === 204 ? (undefined as T) : res.json()
}

export const api = {
  searchPlaces: (q: string) => request<Place[]>(`/api/places?q=${encodeURIComponent(q)}`),

  listLocations: () => request<SavedLocation[]>('/api/locations'),

  addLocation: (place: Place) =>
    request<SavedLocation>('/api/locations', {
      method: 'POST',
      body: JSON.stringify({
        externalId: place.id,
        name: place.name,
        country: place.country,
        admin1: place.admin1,
        latitude: place.latitude,
        longitude: place.longitude,
      }),
    }),

  removeLocation: (id: number) => request<void>(`/api/locations/${id}`, { method: 'DELETE' }),

  forecast: (id: number) => request<Forecast>(`/api/locations/${id}/forecast`),
}

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { PlaceSearch } from '@/components/PlaceSearch'

const helsinki = {
  id: 658225,
  name: 'Helsinki',
  country: 'Finland',
  admin1: 'Uusimaa',
  latitude: 60.17,
  longitude: 24.94,
}

function renderSearch(savedExternalIds = new Set<number>()) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={client}>
      <PlaceSearch savedExternalIds={savedExternalIds} />
    </QueryClientProvider>,
  )
}

function mockFetch(handler: (url: string, init?: RequestInit) => unknown) {
  const fn = vi.fn<(url: string, init?: RequestInit) => Promise<Response>>(async (url, init) => {
    const body = handler(url, init)
    return new Response(JSON.stringify(body), { status: init?.method === 'POST' ? 201 : 200 })
  })
  vi.stubGlobal('fetch', fn)
  return fn
}

afterEach(() => vi.unstubAllGlobals())

describe('PlaceSearch', () => {
  it('searches and saves a place', async () => {
    const fetch = mockFetch((url) => (url.startsWith('/api/places') ? [helsinki] : { id: 1 }))
    const user = userEvent.setup()
    renderSearch()

    await user.type(screen.getByRole('searchbox'), 'Hels')
    await user.click(await screen.findByRole('button', { name: 'Add' }))

    const post = fetch.mock.calls.find(([, init]) => init?.method === 'POST')
    expect(post?.[0]).toBe('/api/locations')
    expect(JSON.parse(post?.[1]?.body as string)).toMatchObject({ externalId: 658225, name: 'Helsinki' })
  })

  it.each([
    ['LOCATION_ALREADY_SAVED', 'This city is already in your watchlist.'],
    ['NEW_ERROR', 'Could not save this city. Please try again.'],
  ])('shows frontend copy for a failed save: %s', async (code, message) => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async (_url: string, init?: RequestInit) =>
        init?.method === 'POST'
          ? new Response(JSON.stringify({ code, message: 'Developer diagnostic: duplicate externalId' }), {
              status: 409,
            })
          : new Response(JSON.stringify([helsinki])),
      ),
    )
    const user = userEvent.setup()
    renderSearch()
    await user.type(screen.getByRole('searchbox'), 'Hels')
    await user.click(await screen.findByRole('button', { name: 'Add' }))
    expect(await screen.findByText(message)).toBeInTheDocument()
    expect(screen.queryByText(/Developer diagnostic/)).not.toBeInTheDocument()
  })

  it('marks places that are already saved', async () => {
    mockFetch(() => [helsinki])
    const user = userEvent.setup()
    renderSearch(new Set([658225]))

    await user.type(screen.getByRole('searchbox'), 'Hels')

    expect(await screen.findByRole('button', { name: 'Saved' })).toBeDisabled()
  })
})

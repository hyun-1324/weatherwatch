import { afterEach, describe, expect, it, vi } from 'vitest'

import { api, ApiError } from '@/api'

afterEach(() => vi.unstubAllGlobals())

describe('API errors', () => {
  it('preserves diagnostics and uses the HTTP status rather than the body status', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            statusCode: 500,
            code: 'LOCATION_ALREADY_SAVED',
            message: 'Duplicate externalId',
          }),
          { status: 409 },
        ),
      ),
    )
    await expect(api.listLocations()).rejects.toMatchObject({
      status: 409,
      code: 'LOCATION_ALREADY_SAVED',
      message: 'Duplicate externalId',
    })
  })

  it.each(['<html>Bad gateway</html>', 'null', '42', '{"code":{},"message":[]}'])(
    'handles an untrusted error body: %s',
    async (body) => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(body, { status: 502 })))
      const error = await api.listLocations().catch((error: unknown) => error)
      expect(error).toBeInstanceOf(ApiError)
      expect(error).toMatchObject({ status: 502, code: 'UNKNOWN', message: 'Request failed (502)' })
    },
  )
})

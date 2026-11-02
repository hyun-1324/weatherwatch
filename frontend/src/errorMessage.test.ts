import { describe, expect, it } from 'vitest'

import { ApiError } from '@/api'
import { errorMessage } from '@/errorMessage'

describe('errorMessage', () => {
  it('uses the stable code regardless of diagnostic wording', () => {
    expect(errorMessage(new ApiError(409, 'LOCATION_ALREADY_SAVED', 'Duplicate externalId'), 'Save failed')).toBe(
      'This city is already in your watchlist.',
    )
  })

  it.each([
    new ApiError(500, 'NEW_ERROR', 'Internal diagnostic'),
    new ApiError(500, 'constructor', 'Internal diagnostic'),
    new TypeError('Failed to fetch'),
    new SyntaxError('Unexpected token'),
    null,
  ])('uses operation-specific copy for unmapped errors: %s', (error) => {
    expect(errorMessage(error, 'Could not save this city.')).toBe('Could not save this city.')
    expect(errorMessage(error, 'Could not remove this city.')).toBe('Could not remove this city.')
  })
})

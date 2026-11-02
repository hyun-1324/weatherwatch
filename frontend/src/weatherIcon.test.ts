import { Cloud, CloudDrizzle, CloudLightning, CloudRain, HelpCircle, Moon, Snowflake, Sun } from 'lucide-react'
import { describe, expect, it } from 'vitest'

import { weatherIcon } from '@/weatherIcon'

describe('weatherIcon', () => {
  it.each([
    [0, true, Sun],
    [0, false, Moon],
    [3, true, Cloud],
    [53, true, CloudDrizzle],
    [63, true, CloudRain],
    [75, true, Snowflake],
    [95, true, CloudLightning],
    [97, true, HelpCircle], // not a WMO code; 95, 96 and 99 are the only thunderstorms
    [1234, true, HelpCircle],
  ])('code %i (day=%s)', (code, isDay, icon) => {
    expect(weatherIcon(code, isDay)).toBe(icon)
  })
})

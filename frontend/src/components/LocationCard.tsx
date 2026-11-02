import { Droplets, X, type LucideProps } from 'lucide-react'
import { createElement } from 'react'

import type { Forecast, SavedLocation } from '@/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { errorMessage } from '@/errorMessage'
import { formatRegion } from '@/formatRegion'
import { useForecast, useRemoveLocation } from '@/hooks'
import { weatherIcon } from '@/weatherIcon'

const weekday = new Intl.DateTimeFormat('en', { weekday: 'short', timeZone: 'UTC' })

interface LocationCardProps {
  location: SavedLocation
}

export function LocationCard({ location }: LocationCardProps) {
  const forecast = useForecast(location.id)
  const remove = useRemoveLocation()

  return (
    <Card>
      <CardHeader>
        <div className="min-w-0">
          <CardTitle className="truncate">{location.name}</CardTitle>
          <CardDescription className="truncate">{formatRegion(location)}</CardDescription>
        </div>
        <Button
          variant="ghost"
          size="icon"
          className="-mt-1 -mr-2 text-muted-foreground"
          aria-label={`Remove ${location.name}`}
          onClick={() => remove.mutate(location.id)}
          disabled={remove.isPending}
        >
          <X />
        </Button>
      </CardHeader>

      <CardContent>
        {remove.isError && (
          <p role="alert" className="mb-3 text-sm text-destructive">
            {errorMessage(remove.error, 'Could not remove this city. Please try again.')}
          </p>
        )}
        {forecast.isPending && <ForecastSkeleton />}
        {forecast.isError && (
          <div className="flex items-center gap-3 text-sm text-destructive">
            {errorMessage(forecast.error, 'Could not load the forecast. Please try again.')}
            <Button variant="outline" size="sm" onClick={() => forecast.refetch()}>
              Retry
            </Button>
          </div>
        )}
        {forecast.data && <ForecastView forecast={forecast.data} />}
      </CardContent>
    </Card>
  )
}

interface ForecastViewProps {
  forecast: Forecast
}

function ForecastView({ forecast }: ForecastViewProps) {
  const { current, daily } = forecast

  return (
    <>
      <div className="flex items-center gap-4">
        <WeatherIcon
          code={current.weatherCode}
          isDay={current.isDay}
          className="size-12 shrink-0 text-primary"
          strokeWidth={1.5}
          aria-hidden
        />
        <span className="text-5xl font-semibold tabular-nums">{Math.round(current.temperature)}°</span>
        <div className="min-w-0 text-sm">
          <div className="font-medium">{current.condition}</div>
          <div className="flex flex-wrap gap-x-3 text-muted-foreground">
            <span>Feels {Math.round(current.apparentTemperature)}°</span>
            <span>Wind {Math.round(current.windSpeed)} km/h</span>
            <span>Humidity {current.humidity}%</span>
          </div>
        </div>
      </div>

      <ol className="mt-4 grid grid-cols-7 gap-1 border-t pt-3 text-center text-xs">
        {daily.map((day) => (
          <li key={day.date} title={day.condition} className="flex flex-col items-center gap-1">
            {/* Dates are local to the location; parse as UTC so the weekday doesn't shift. */}
            <span className="text-muted-foreground">{weekday.format(new Date(`${day.date}T00:00:00Z`))}</span>
            <WeatherIcon code={day.weatherCode} className="size-5" strokeWidth={1.75} aria-label={day.condition} />
            <span className="font-medium tabular-nums">{Math.round(day.tempMax)}°</span>
            <span className="text-muted-foreground tabular-nums">{Math.round(day.tempMin)}°</span>
            {day.precipitationProbability != null && (
              <span className="flex items-center text-[10px] text-muted-foreground tabular-nums">
                <Droplets className="size-2.5" aria-hidden />
                {day.precipitationProbability}%
              </span>
            )}
          </li>
        ))}
      </ol>
    </>
  )
}

interface WeatherIconProps extends LucideProps {
  code: number
  isDay?: boolean
}

function WeatherIcon({ code, isDay, ...props }: WeatherIconProps) {
  return createElement(weatherIcon(code, isDay), props)
}

function ForecastSkeleton() {
  return (
    <div aria-label="Loading forecast">
      <div className="flex items-center gap-4">
        <Skeleton className="size-12 rounded-full" />
        <Skeleton className="h-12 w-20" />
        <div className="flex-1 space-y-2">
          <Skeleton className="h-4 w-24" />
          <Skeleton className="h-3 w-40" />
        </div>
      </div>
      <Skeleton className="mt-4 h-20 w-full" />
    </div>
  )
}

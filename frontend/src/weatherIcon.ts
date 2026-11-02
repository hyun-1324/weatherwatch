import {
  Cloud,
  CloudDrizzle,
  CloudFog,
  CloudLightning,
  CloudMoon,
  CloudRain,
  CloudSnow,
  CloudSun,
  HelpCircle,
  Moon,
  Snowflake,
  Sun,
  type LucideIcon,
} from 'lucide-react'

export function weatherIcon(code: number, isDay = true): LucideIcon {
  if (code === 0) return isDay ? Sun : Moon
  if (code <= 2) return isDay ? CloudSun : CloudMoon
  if (code === 3) return Cloud
  if (code === 45 || code === 48) return CloudFog
  if (code >= 51 && code <= 57) return CloudDrizzle
  if (code >= 61 && code <= 67) return CloudRain
  if (code >= 71 && code <= 77) return Snowflake
  if (code >= 80 && code <= 82) return CloudRain
  if (code === 85 || code === 86) return CloudSnow
  if (code === 95 || code === 96 || code === 99) return CloudLightning
  return HelpCircle
}

const activityTypeLabels: Record<string, string> = {
  alpineski: 'Alpine ski',
  backcountryski: 'Backcountry ski',
  canoeing: 'Canoeing',
  golf: 'Golf',
  handcycle: 'Handcycle',
  highintensityintervaltraining: 'High-intensity interval training',
  iceskate: 'Ice skate',
  inlineskate: 'Inline skate',
  kayaking: 'Kayaking',
  kitesurf: 'Kitesurfing',
  nordicski: 'Nordic ski',
  pickleball: 'Pickleball',
  pilates: 'Pilates',
  racquetball: 'Racquetball',
  run: 'Run',
  ride: 'Ride',
  rockclimbing: 'Rock climbing',
  rollerski: 'Roller ski',
  rowing: 'Rowing',
  sail: 'Sailing',
  skateboard: 'Skateboarding',
  snowboard: 'Snowboarding',
  snowshoe: 'Snowshoeing',
  soccer: 'Soccer',
  squash: 'Squash',
  standuppaddling: 'Stand-up paddling',
  surfing: 'Surfing',
  swim: 'Swim',
  tabletennis: 'Table tennis',
  tennis: 'Tennis',
  velomobile: 'Velomobile',
  virtualrow: 'Virtual row',
  walk: 'Walk',
  wheelchair: 'Wheelchair',
  windsurf: 'Windsurfing',
  workout: 'Workout',
  yoga: 'Yoga',
}

const defaultNumberLocale = 'en-US'

export function formatDistance(
  meters: number,
  locale: Intl.LocalesArgument = defaultNumberLocale,
): string {
  if (meters < 1000) {
    return `${new Intl.NumberFormat(locale, {
      maximumFractionDigits: 0,
    }).format(meters)} m`
  }

  return `${new Intl.NumberFormat(locale, {
    maximumFractionDigits: 1,
  }).format(meters / 1000)} km`
}

export function formatDuration(seconds: number): string {
  const totalMinutes = Math.round(seconds / 60)
  const hours = Math.floor(totalMinutes / 60)
  const minutes = totalMinutes % 60

  if (hours === 0) {
    return `${minutes} min`
  }

  if (minutes === 0) {
    return `${hours} h`
  }

  return `${hours} h ${minutes} min`
}

export function formatElevation(
  meters: number | undefined,
  locale: Intl.LocalesArgument = defaultNumberLocale,
): string {
  if (meters === undefined) {
    return '—'
  }

  return `${new Intl.NumberFormat(locale, {
    maximumFractionDigits: 0,
  }).format(meters)} m`
}

export function formatActivityDate(
  timestamp: string,
  locale?: Intl.LocalesArgument,
  timeZone?: string,
): string {
  return new Intl.DateTimeFormat(locale, {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
    timeZone,
  }).format(new Date(timestamp))
}

export function formatActivityType(activityType: string): string {
  const normalizedType = activityType.trim().toLowerCase()
  const knownLabel = activityTypeLabels[normalizedType]

  if (knownLabel) {
    return knownLabel
  }

  const readableType = normalizedType
    .replace(/[_-]+/g, ' ')
    .replace(/\s+/g, ' ')

  if (!readableType) {
    return 'Unknown'
  }

  return readableType[0].toUpperCase() + readableType.slice(1)
}

function parseLocalDateKey(dateKey: string): Date {
  const [year, month, day] = dateKey.split('-').map(Number)
  return new Date(year, month - 1, day)
}

export function formatPeriod(
  startDateKey: string,
  endDateKey: string,
  locale?: Intl.LocalesArgument,
): string {
  const startDate = parseLocalDateKey(startDateKey)
  const endDate = parseLocalDateKey(endDateKey)
  const sameYear = startDate.getFullYear() === endDate.getFullYear()
  const startFormatter = new Intl.DateTimeFormat(locale, {
    day: 'numeric',
    month: 'long',
    year: sameYear ? undefined : 'numeric',
  })
  const endFormatter = new Intl.DateTimeFormat(locale, {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  })

  return `${startFormatter.format(startDate)} – ${endFormatter.format(endDate)}`
}

import { describe, expect, it } from 'vitest'
import {
  formatActivityDate,
  formatActivityType,
  formatDistance,
  formatDuration,
  formatElevation,
  formatPeriod,
} from './formatters'

describe('formatDistance', () => {
  it('uses a decimal point by default', () => {
    expect(formatDistance(7400)).toBe('7.4 km')
  })

  it('formats distances of at least one kilometer as kilometers', () => {
    expect(formatDistance(7400, 'de-DE')).toBe('7,4 km')
  })

  it('formats shorter distances as rounded meters', () => {
    expect(formatDistance(842.4, 'de-DE')).toBe('842 m')
  })
})

describe('formatDuration', () => {
  it('formats hours and minutes', () => {
    expect(formatDuration(5220)).toBe('1 h 27 min')
  })

  it('omits minutes for an exact number of hours', () => {
    expect(formatDuration(7200)).toBe('2 h')
  })

  it('formats durations shorter than one hour as minutes', () => {
    expect(formatDuration(1500)).toBe('25 min')
  })
})

describe('formatElevation', () => {
  it('formats elevation as rounded meters', () => {
    expect(formatElevation(839.6, 'de-DE')).toBe('840 m')
  })

  it('shows a placeholder when elevation is unavailable', () => {
    expect(formatElevation(undefined, 'de-DE')).toBe('—')
  })
})

describe('formatActivityDate', () => {
  it('formats a timestamp for the requested locale and time zone', () => {
    expect(
      formatActivityDate('2026-06-18T07:30:00Z', 'de-DE', 'UTC'),
    ).toBe('18. Juni 2026')
  })
})

describe('formatActivityType', () => {
  it('uses readable names for known canonical activity types', () => {
    expect(formatActivityType('run')).toBe('Run')
    expect(formatActivityType('ride')).toBe('Ride')
    expect(formatActivityType('alpineski')).toBe('Alpine ski')
    expect(formatActivityType('backcountryski')).toBe('Backcountry ski')
    expect(formatActivityType('standuppaddling')).toBe('Stand-up paddling')
    expect(formatActivityType('rockclimbing')).toBe('Rock climbing')
  })

  it('turns an unknown machine-readable type into readable text', () => {
    expect(formatActivityType('cross_country_ski')).toBe('Cross country ski')
  })
})

describe('formatPeriod', () => {
  it('formats a date-key range in the requested locale', () => {
    expect(formatPeriod('2026-05-22', '2026-06-18', 'de-DE')).toBe(
      '22. Mai – 18. Juni 2026',
    )
  })
})

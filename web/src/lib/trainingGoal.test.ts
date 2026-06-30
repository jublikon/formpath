import { describe, expect, it } from 'vitest'
import {
  calculateCalendarDaysRemaining,
  goalToFormValues,
  localDateKey,
  parseDurationInput,
  validateTrainingGoalForm,
} from './trainingGoal'

describe('localDateKey', () => {
  it('formats the local calendar date', () => {
    expect(localDateKey(new Date(2026, 5, 18, 16))).toBe('2026-06-18')
  })
})

describe('calculateCalendarDaysRemaining', () => {
  it('counts calendar days instead of elapsed hours', () => {
    expect(
      calculateCalendarDaysRemaining(
        '2026-06-20',
        new Date(2026, 5, 18, 23),
      ),
    ).toBe(2)
  })

  it('returns null for invalid target dates', () => {
    expect(
      calculateCalendarDaysRemaining('2026-02-30', new Date(2026, 5, 18)),
    ).toBeNull()
  })
})

describe('parseDurationInput', () => {
  it('accepts hours and minutes', () => {
    expect(parseDurationInput('3:30')).toBe(12600)
  })

  it('accepts hours, minutes, and seconds', () => {
    expect(parseDurationInput('3:30:15')).toBe(12615)
  })

  it('treats a blank value as omitted', () => {
    expect(parseDurationInput('')).toBeUndefined()
  })

  it('rejects invalid durations', () => {
    expect(parseDurationInput('3:75')).toBeNull()
    expect(parseDurationInput('soon')).toBeNull()
    expect(parseDurationInput('0:00')).toBeNull()
    expect(parseDurationInput('596524:00')).toBeNull()
  })
})

describe('goalToFormValues', () => {
  it('converts stored base units for editing', () => {
    expect(
      goalToFormValues({
        user_id: '00000000-0000-0000-0000-000000000042',
        goal_type: 'distance_event',
        sport: 'run',
        name: 'Berlin Marathon',
        target_distance_meters: 42195,
        target_date: '2026-09-27',
        target_duration_seconds: 12600,
      }),
    ).toEqual({
      name: 'Berlin Marathon',
      sport: 'run',
      targetDistanceKilometers: '42.195',
      targetDate: '2026-09-27',
      targetDuration: '3:30',
    })
  })
})

describe('validateTrainingGoalForm', () => {
  const today = new Date(2026, 5, 18)

  it('builds a goal payload in base units', () => {
    const result = validateTrainingGoalForm(
      {
        name: ' Berlin Marathon ',
        sport: 'run',
        targetDistanceKilometers: '42.195',
        targetDate: '2026-09-27',
        targetDuration: '3:30',
      },
      today,
    )

    expect(result).toEqual({
      valid: true,
      payload: {
        name: 'Berlin Marathon',
        sport: 'run',
        target_distance_meters: 42195,
        target_date: '2026-09-27',
        target_duration_seconds: 12600,
      },
    })
  })

  it('allows an omitted target duration', () => {
    const result = validateTrainingGoalForm(
      {
        name: 'Summer Century',
        sport: 'ride',
        targetDistanceKilometers: '160.934',
        targetDate: '2026-07-12',
        targetDuration: '',
      },
      today,
    )

    expect(result).toEqual({
      valid: true,
      payload: {
        name: 'Summer Century',
        sport: 'ride',
        target_distance_meters: 160934,
        target_date: '2026-07-12',
      },
    })
  })

  it('rejects target dates before the browser local date', () => {
    const result = validateTrainingGoalForm(
      {
        name: 'Past Race',
        sport: 'run',
        targetDistanceKilometers: '10',
        targetDate: '2026-06-17',
        targetDuration: '',
      },
      today,
    )

    expect(result).toEqual({
      valid: false,
      message: 'Choose today or a future target date.',
    })
  })

  it('returns understandable validation messages', () => {
    expect(
      validateTrainingGoalForm(
        {
          name: '',
          sport: 'run',
          targetDistanceKilometers: '10',
          targetDate: '2026-06-18',
          targetDuration: '',
        },
        today,
      ),
    ).toEqual({ valid: false, message: 'Enter an event name.' })

    expect(
      validateTrainingGoalForm(
        {
          name: 'Race',
          sport: 'run',
          targetDistanceKilometers: '0',
          targetDate: '2026-06-18',
          targetDuration: '',
        },
        today,
      ),
    ).toEqual({
      valid: false,
      message: 'Enter a target distance greater than 0.',
    })
  })
})

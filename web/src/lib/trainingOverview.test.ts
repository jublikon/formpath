import { describe, expect, it } from 'vitest'
import type { Activity } from '../types/activity'
import { buildTrainingOverview } from './trainingOverview'

function localDate(
  year: number,
  monthIndex: number,
  day: number,
  hour = 12,
): Date {
  return new Date(year, monthIndex, day, hour)
}

function activity(
  overrides: Partial<Activity> & Pick<Activity, 'id' | 'started_at'>,
): Activity {
  return {
    name: 'Training activity',
    activity_type: 'run',
    distance_meters: 0,
    duration_seconds: 0,
    moving_seconds: 0,
    ...overrides,
  }
}

describe('buildTrainingOverview', () => {
  const today = localDate(2026, 5, 18)

  it('creates 28 consecutive local calendar days ending today', () => {
    const overview = buildTrainingOverview([], today)

    expect(overview.period).toEqual({
      startDate: '2026-05-22',
      endDate: '2026-06-18',
    })
    expect(overview.days).toHaveLength(28)
    expect(overview.days[0]).toEqual({
      date: '2026-05-22',
      movingSeconds: 0,
      runDistanceMeters: 0,
      rideDistanceMeters: 0,
      workoutMovingSeconds: 0,
    })
    expect(overview.days[27]).toEqual({
      date: '2026-06-18',
      movingSeconds: 0,
      runDistanceMeters: 0,
      rideDistanceMeters: 0,
      workoutMovingSeconds: 0,
    })
  })

  it('adds multiple activities from the same local day', () => {
    const overview = buildTrainingOverview(
      [
        activity({
          id: 'morning-run',
          started_at: localDate(2026, 5, 16, 8).toISOString(),
          moving_seconds: 2700,
        }),
        activity({
          id: 'evening-ride',
          started_at: localDate(2026, 5, 16, 18).toISOString(),
          moving_seconds: 1800,
        }),
      ],
      today,
    )

    expect(
      overview.days.find((day) => day.date === '2026-06-16'),
    ).toEqual({
      date: '2026-06-16',
      movingSeconds: 4500,
      runDistanceMeters: 0,
      rideDistanceMeters: 0,
      workoutMovingSeconds: 0,
    })
  })

  it('builds activity-specific daily distance and time series', () => {
    const overview = buildTrainingOverview(
      [
        activity({
          id: 'morning-run',
          activity_type: 'run',
          started_at: localDate(2026, 5, 16, 8).toISOString(),
          distance_meters: 7200,
          moving_seconds: 2400,
        }),
        activity({
          id: 'evening-run',
          activity_type: ' Run ',
          started_at: localDate(2026, 5, 16, 18).toISOString(),
          distance_meters: 3800,
          moving_seconds: 1500,
        }),
        activity({
          id: 'ride',
          activity_type: 'ride',
          started_at: localDate(2026, 5, 17).toISOString(),
          distance_meters: 42500,
          moving_seconds: 5400,
        }),
        activity({
          id: 'workout',
          activity_type: 'workout',
          started_at: localDate(2026, 5, 18).toISOString(),
          distance_meters: 900,
          moving_seconds: 2700,
        }),
        activity({
          id: 'walk',
          activity_type: 'walk',
          started_at: localDate(2026, 5, 18).toISOString(),
          distance_meters: 5000,
          moving_seconds: 3600,
        }),
      ],
      today,
    )

    expect(
      overview.days.find((day) => day.date === '2026-06-16'),
    ).toMatchObject({
      runDistanceMeters: 11000,
      rideDistanceMeters: 0,
      workoutMovingSeconds: 0,
    })
    expect(
      overview.days.find((day) => day.date === '2026-06-17'),
    ).toMatchObject({
      runDistanceMeters: 0,
      rideDistanceMeters: 42500,
      workoutMovingSeconds: 0,
    })
    expect(
      overview.days.find((day) => day.date === '2026-06-18'),
    ).toMatchObject({
      runDistanceMeters: 0,
      rideDistanceMeters: 0,
      workoutMovingSeconds: 2700,
    })
  })

  it('calculates totals from activities inside the period', () => {
    const overview = buildTrainingOverview(
      [
        activity({
          id: 'run',
          started_at: localDate(2026, 5, 10).toISOString(),
          distance_meters: 8400,
          duration_seconds: 4000,
          moving_seconds: 3600,
          elevation_gain_meters: 120,
        }),
        activity({
          id: 'ride',
          started_at: localDate(2026, 5, 18).toISOString(),
          distance_meters: 10000,
          duration_seconds: 4300,
          moving_seconds: 3600,
          elevation_gain_meters: 190,
        }),
      ],
      today,
    )

    expect(overview.totals).toEqual({
      activityCount: 2,
      distanceMeters: 18400,
      movingSeconds: 7200,
      elevationGainMeters: 310,
    })
  })

  it('ignores activities outside the 28-day period', () => {
    const overview = buildTrainingOverview(
      [
        activity({
          id: 'before-period',
          started_at: localDate(2026, 4, 21, 23).toISOString(),
          distance_meters: 5000,
          moving_seconds: 1800,
        }),
        activity({
          id: 'period-start',
          started_at: localDate(2026, 4, 22, 0).toISOString(),
          distance_meters: 7000,
          moving_seconds: 2400,
        }),
        activity({
          id: 'after-period',
          started_at: localDate(2026, 5, 19, 0).toISOString(),
          distance_meters: 9000,
          moving_seconds: 3000,
        }),
      ],
      today,
    )

    expect(overview.totals).toEqual({
      activityCount: 1,
      distanceMeters: 7000,
      movingSeconds: 2400,
      elevationGainMeters: undefined,
    })
  })

  it('uses moving time rather than elapsed duration', () => {
    const overview = buildTrainingOverview(
      [
        activity({
          id: 'paused-run',
          started_at: localDate(2026, 5, 18).toISOString(),
          duration_seconds: 5400,
          moving_seconds: 3600,
        }),
      ],
      today,
    )

    expect(overview.totals.movingSeconds).toBe(3600)
    expect(overview.days[27].movingSeconds).toBe(3600)
  })

  it('keeps elevation unavailable when no activity provides it', () => {
    const overview = buildTrainingOverview(
      [
        activity({
          id: 'run-without-elevation',
          started_at: localDate(2026, 5, 18).toISOString(),
        }),
      ],
      today,
    )

    expect(overview.totals.elevationGainMeters).toBeUndefined()
  })
})

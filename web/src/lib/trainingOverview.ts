import type { Activity } from '../types/activity'

const overviewDayCount = 28

export type TrainingDay = {
  date: string
  movingSeconds: number
  runDistanceMeters: number
  rideDistanceMeters: number
  workoutMovingSeconds: number
}

export type TrainingOverview = {
  period: {
    startDate: string
    endDate: string
  }
  totals: {
    activityCount: number
    distanceMeters: number
    movingSeconds: number
    elevationGainMeters: number | undefined
  }
  days: TrainingDay[]
}

function startOfLocalDay(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate())
}

function addLocalDays(date: Date, days: number): Date {
  const result = new Date(date)
  result.setDate(result.getDate() + days)
  return result
}

function localDateKey(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')

  return `${year}-${month}-${day}`
}

export function buildTrainingOverview(
  activities: Activity[],
  today: Date,
): TrainingOverview {
  const periodEnd = startOfLocalDay(today)
  const periodStart = addLocalDays(periodEnd, -(overviewDayCount - 1))
  const days = Array.from({ length: overviewDayCount }, (_, index) => ({
    date: localDateKey(addLocalDays(periodStart, index)),
    movingSeconds: 0,
    runDistanceMeters: 0,
    rideDistanceMeters: 0,
    workoutMovingSeconds: 0,
  }))
  const daysByDate = new Map(days.map((day) => [day.date, day]))

  let activityCount = 0
  let distanceMeters = 0
  let movingSeconds = 0
  let elevationGainMeters = 0
  let hasElevationData = false

  for (const activity of activities) {
    const activityDate = new Date(activity.started_at)
    if (Number.isNaN(activityDate.getTime())) {
      continue
    }

    const day = daysByDate.get(localDateKey(activityDate))
    if (!day) {
      continue
    }

    activityCount += 1
    distanceMeters += activity.distance_meters
    movingSeconds += activity.moving_seconds
    day.movingSeconds += activity.moving_seconds

    switch (activity.activity_type.trim().toLowerCase()) {
      case 'run':
        day.runDistanceMeters += activity.distance_meters
        break
      case 'ride':
        day.rideDistanceMeters += activity.distance_meters
        break
      case 'workout':
        day.workoutMovingSeconds += activity.moving_seconds
        break
    }

    if (activity.elevation_gain_meters !== undefined) {
      elevationGainMeters += activity.elevation_gain_meters
      hasElevationData = true
    }
  }

  return {
    period: {
      startDate: days[0].date,
      endDate: days[days.length - 1].date,
    },
    totals: {
      activityCount,
      distanceMeters,
      movingSeconds,
      elevationGainMeters: hasElevationData ? elevationGainMeters : undefined,
    },
    days,
  }
}

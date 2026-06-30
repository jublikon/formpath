import type {
  TrainingGoal,
  TrainingGoalPayload,
  TrainingGoalSport,
} from '../types/trainingGoal'

export type TrainingGoalFormValues = {
  name: string
  sport: TrainingGoalSport
  targetDistanceKilometers: string
  targetDate: string
  targetDuration: string
}

export type TrainingGoalValidationResult =
  | {
      valid: true
      payload: TrainingGoalPayload
    }
  | {
      valid: false
      message: string
    }

export function localDateKey(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')

  return `${year}-${month}-${day}`
}

function parseLocalDateKey(dateKey: string): Date | null {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(dateKey)) {
    return null
  }

  const [year, month, day] = dateKey.split('-').map(Number)
  const parsed = new Date(year, month - 1, day)

  if (localDateKey(parsed) !== dateKey) {
    return null
  }

  return parsed
}

export function calculateCalendarDaysRemaining(
  targetDate: string,
  today: Date,
): number | null {
  const parsedTargetDate = parseLocalDateKey(targetDate)
  if (!parsedTargetDate) {
    return null
  }

  const todayStart = parseLocalDateKey(localDateKey(today))
  if (!todayStart) {
    return null
  }

  const millisecondsPerDay = 24 * 60 * 60 * 1000
  return Math.round(
    (parsedTargetDate.getTime() - todayStart.getTime()) / millisecondsPerDay,
  )
}

export function goalToFormValues(
  goal: TrainingGoal | null,
): TrainingGoalFormValues {
  if (!goal) {
    return {
      name: '',
      sport: 'run',
      targetDistanceKilometers: '',
      targetDate: '',
      targetDuration: '',
    }
  }

  return {
    name: goal.name,
    sport: goal.sport,
    targetDistanceKilometers: String(goal.target_distance_meters / 1000),
    targetDate: goal.target_date,
    targetDuration:
      goal.target_duration_seconds === undefined
        ? ''
        : formatDurationInput(goal.target_duration_seconds),
  }
}

export function parseDurationInput(value: string): number | undefined | null {
  const trimmedValue = value.trim()
  if (trimmedValue === '') {
    return undefined
  }

  const parts = trimmedValue.split(':')
  if (parts.length < 2 || parts.length > 3) {
    return null
  }

  const parsedParts = parts.map((part) => {
    if (!/^\d+$/.test(part)) {
      return Number.NaN
    }
    return Number(part)
  })

  if (parsedParts.some((part) => Number.isNaN(part))) {
    return null
  }

  const [hours, minutes, seconds = 0] = parsedParts
  if (minutes > 59 || seconds > 59) {
    return null
  }

  const totalSeconds = hours * 60 * 60 + minutes * 60 + seconds
  return totalSeconds > 0 ? totalSeconds : null
}

export function formatDurationInput(seconds: number): string {
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const remainingSeconds = seconds % 60

  if (remainingSeconds === 0) {
    return `${hours}:${String(minutes).padStart(2, '0')}`
  }

  return `${hours}:${String(minutes).padStart(2, '0')}:${String(
    remainingSeconds,
  ).padStart(2, '0')}`
}

export function validateTrainingGoalForm(
  values: TrainingGoalFormValues,
  today: Date,
): TrainingGoalValidationResult {
  const name = values.name.trim()
  if (name === '') {
    return { valid: false, message: 'Enter an event name.' }
  }

  if (values.sport !== 'run' && values.sport !== 'ride') {
    return { valid: false, message: 'Choose running or cycling.' }
  }

  const targetDistanceKilometers = Number(values.targetDistanceKilometers)
  if (
    !Number.isFinite(targetDistanceKilometers) ||
    targetDistanceKilometers <= 0
  ) {
    return { valid: false, message: 'Enter a target distance greater than 0.' }
  }

  if (!parseLocalDateKey(values.targetDate)) {
    return { valid: false, message: 'Enter a valid target date.' }
  }

  const daysRemaining = calculateCalendarDaysRemaining(values.targetDate, today)
  if (daysRemaining === null || daysRemaining < 0) {
    return {
      valid: false,
      message: 'Choose today or a future target date.',
    }
  }

  const targetDurationSeconds = parseDurationInput(values.targetDuration)
  if (targetDurationSeconds === null) {
    return {
      valid: false,
      message: 'Enter target time as hours and minutes.',
    }
  }

  return {
    valid: true,
    payload: {
      name,
      sport: values.sport,
      target_distance_meters: targetDistanceKilometers * 1000,
      target_date: values.targetDate,
      ...(targetDurationSeconds === undefined
        ? {}
        : { target_duration_seconds: targetDurationSeconds }),
    },
  }
}

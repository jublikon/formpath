import { useState, type FormEvent } from 'react'
import {
  calculateCalendarDaysRemaining,
  goalToFormValues,
  localDateKey,
  validateTrainingGoalForm,
  type TrainingGoalFormValues,
} from '../lib/trainingGoal'
import type { TrainingGoal, TrainingGoalPayload } from '../types/trainingGoal'
import { formatDistance, formatDuration } from '../lib/formatters'

type TrainingGoalPanelProps = {
  goal: TrainingGoal | null
  today: Date
  loading: boolean
  saving: boolean
  error: string | null
  loadFailed: boolean
  onRetryLoad: () => Promise<void>
  onSave: (goal: TrainingGoalPayload) => Promise<void>
  onDelete: () => Promise<void>
}

function formatGoalSport(sport: TrainingGoal['sport']): string {
  return sport === 'run' ? 'Running' : 'Cycling'
}

function formatCalendarDate(dateKey: string): string {
  const [year, month, day] = dateKey.split('-').map(Number)
  return new Intl.DateTimeFormat(undefined, {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  }).format(new Date(year, month - 1, day))
}

function describeDaysRemaining(daysRemaining: number | null): string {
  if (daysRemaining === null) {
    return 'Target date unavailable'
  }

  if (daysRemaining === 0) {
    return 'Today'
  }

  if (daysRemaining === 1) {
    return '1 day remaining'
  }

  if (daysRemaining > 1) {
    return `${daysRemaining} days remaining`
  }

  if (daysRemaining === -1) {
    return '1 day ago'
  }

  return `${Math.abs(daysRemaining)} days ago`
}

export function TrainingGoalPanel({
  goal,
  today,
  loading,
  saving,
  error,
  loadFailed,
  onRetryLoad,
  onSave,
  onDelete,
}: TrainingGoalPanelProps) {
  const [isEditing, setIsEditing] = useState(goal === null)
  const [formValues, setFormValues] = useState<TrainingGoalFormValues>(() =>
    goalToFormValues(goal),
  )
  const [formError, setFormError] = useState<string | null>(null)

  function updateFormValue<Name extends keyof TrainingGoalFormValues>(
    name: Name,
    value: TrainingGoalFormValues[Name],
  ) {
    setFormValues((currentValues) => ({
      ...currentValues,
      [name]: value,
    }))
  }

  async function submitGoal(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const validation = validateTrainingGoalForm(formValues, today)
    if (!validation.valid) {
      setFormError(validation.message)
      return
    }

    setFormError(null)
    try {
      await onSave(validation.payload)
      setIsEditing(false)
    } catch {
      return
    }
  }

  async function deleteGoal() {
    if (!window.confirm('Remove this training goal?')) {
      return
    }

    setFormError(null)
    try {
      await onDelete()
    } catch {
      return
    }
  }

  const todayKey = localDateKey(today)
  const daysRemaining = goal
    ? calculateCalendarDaysRemaining(goal.target_date, today)
    : null
  const showLoadFailure = loadFailed && !goal

  return (
    <section className="goal-panel" aria-labelledby="goal-heading">
      <div className="section-heading">
        <div>
          <p className="section-eyebrow">Training goal</p>
          <h2 id="goal-heading">
            {goal ? goal.name : 'Add a race or event goal'}
          </h2>
        </div>

        {goal && !isEditing && (
          <div className="goal-actions" role="group" aria-label="Goal actions">
            <button
              className="secondary-action"
              type="button"
              onClick={() => setIsEditing(true)}
            >
              Edit
            </button>
            <button
              className="secondary-action danger-action"
              type="button"
              disabled={saving}
              onClick={deleteGoal}
            >
              Remove
            </button>
          </div>
        )}
      </div>

      {error && (
        <p className="goal-message goal-message-error" role="alert">
          {error}
        </p>
      )}

      {showLoadFailure && (
        <div className="goal-retry-state">
          <p>
            The current goal is temporarily unavailable. Try loading it again
            before creating or replacing a goal.
          </p>
          <button
            className="secondary-action"
            type="button"
            disabled={loading}
            onClick={() => {
              setFormError(null)
              void onRetryLoad()
            }}
          >
            {loading ? 'Loading...' : 'Retry'}
          </button>
        </div>
      )}

      {goal && !isEditing && (
        <dl className="goal-summary">
          <div>
            <dt>Sport</dt>
            <dd>{formatGoalSport(goal.sport)}</dd>
          </div>
          <div>
            <dt>Distance</dt>
            <dd>{formatDistance(goal.target_distance_meters)}</dd>
          </div>
          <div>
            <dt>Target date</dt>
            <dd>{formatCalendarDate(goal.target_date)}</dd>
          </div>
          <div>
            <dt>Time remaining</dt>
            <dd>{describeDaysRemaining(daysRemaining)}</dd>
          </div>
          {goal.target_duration_seconds !== undefined && (
            <div>
              <dt>Target time</dt>
              <dd>{formatDuration(goal.target_duration_seconds)}</dd>
            </div>
          )}
        </dl>
      )}

      {isEditing && !showLoadFailure && (
        <form className="goal-form" onSubmit={submitGoal} noValidate>
          <div className="goal-form-grid">
            <label className="field">
              <span>Event name</span>
              <input
                type="text"
                value={formValues.name}
                onChange={(event) =>
                  updateFormValue('name', event.currentTarget.value)
                }
                autoComplete="off"
              />
            </label>

            <label className="field">
              <span>Sport</span>
              <select
                value={formValues.sport}
                onChange={(event) =>
                  updateFormValue(
                    'sport',
                    event.currentTarget.value === 'ride' ? 'ride' : 'run',
                  )
                }
              >
                <option value="run">Running</option>
                <option value="ride">Cycling</option>
              </select>
            </label>

            <label className="field">
              <span>Distance (km)</span>
              <input
                type="number"
                min="0.001"
                step="0.001"
                placeholder="42.195"
                value={formValues.targetDistanceKilometers}
                onChange={(event) =>
                  updateFormValue(
                    'targetDistanceKilometers',
                    event.currentTarget.value,
                  )
                }
              />
            </label>

            <label className="field">
              <span>Target date</span>
              <input
                type="date"
                min={todayKey}
                value={formValues.targetDate}
                onChange={(event) =>
                  updateFormValue('targetDate', event.currentTarget.value)
                }
              />
            </label>

            <label className="field">
              <span>Target time (optional)</span>
              <input
                type="text"
                placeholder="3:30"
                value={formValues.targetDuration}
                onChange={(event) =>
                  updateFormValue('targetDuration', event.currentTarget.value)
                }
              />
            </label>
          </div>

          {formError && (
            <p className="goal-message goal-message-error" role="alert">
              {formError}
            </p>
          )}

          <div className="goal-actions">
            <button className="primary-action" type="submit" disabled={saving}>
              {saving ? 'Saving...' : 'Save goal'}
            </button>
            {goal && (
              <button
                className="secondary-action"
                type="button"
                disabled={saving}
                onClick={() => {
                  setFormValues(goalToFormValues(goal))
                  setFormError(null)
                  setIsEditing(false)
                }}
              >
                Cancel
              </button>
            )}
          </div>
        </form>
      )}
    </section>
  )
}

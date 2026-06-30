import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import { TrainingGoalPanel } from './TrainingGoalPanel'
import type { TrainingGoal } from '../types/trainingGoal'

function renderPanel(goal: TrainingGoal | null) {
  return renderToStaticMarkup(
    <TrainingGoalPanel
      goal={goal}
      today={new Date(2026, 5, 18)}
      saving={false}
      error={null}
      onSave={async () => undefined}
      onDelete={async () => undefined}
    />,
  )
}

describe('TrainingGoalPanel', () => {
  it('renders the create form when no goal exists', () => {
    const html = renderPanel(null)

    expect(html).toContain('Add a race or event goal')
    expect(html).toContain('Event name')
    expect(html).toContain('Distance (km)')
    expect(html).toContain('placeholder="42.195"')
    expect(html).toContain('Save goal')
  })

  it('renders active goal facts without readiness claims', () => {
    const html = renderPanel({
      user_id: '00000000-0000-0000-0000-000000000042',
      goal_type: 'distance_event',
      sport: 'run',
      name: 'Berlin Marathon',
      target_distance_meters: 42195,
      target_date: '2026-09-27',
      target_duration_seconds: 12600,
    })

    expect(html).toContain('Berlin Marathon')
    expect(html).toContain('Running')
    expect(html).toContain('42.2 km')
    expect(html).toContain('101 days remaining')
    expect(html).toContain('3 h 30 min')
    expect(html).toContain('Edit')
    expect(html).toContain('Remove')
    expect(html).not.toContain('ready')
    expect(html).not.toContain('on track')
  })
})

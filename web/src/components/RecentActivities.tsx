import { useState } from 'react'
import {
  formatActivityDate,
  formatActivityType,
  formatDistance,
  formatDuration,
  formatElevation,
} from '../lib/formatters'
import type { Activity } from '../types/activity'

type RecentActivitiesProps = {
  activities: Activity[]
}

const initialActivityCount = 10
const numberFormatter = new Intl.NumberFormat('en-US')

export function RecentActivities({ activities }: RecentActivitiesProps) {
  const [showAll, setShowAll] = useState(false)
  const visibleActivities = showAll
    ? activities
    : activities.slice(0, initialActivityCount)
  const hasHiddenActivities = activities.length > initialActivityCount
  const visibleActivityCount = numberFormatter.format(visibleActivities.length)
  const totalActivityCount = numberFormatter.format(activities.length)

  return (
    <section className="recent-activities" aria-labelledby="activities-heading">
      <div className="section-heading">
        <div>
          <p className="section-eyebrow">Activity history</p>
          <h2 id="activities-heading">Recent activities</h2>
        </div>
        <p className="period">
          Showing {visibleActivityCount} of {totalActivityCount}
        </p>
      </div>

      <ul className="activity-list" id="activity-list">
        {visibleActivities.map((activity) => (
          <li className="activity" key={activity.id}>
            <div>
              <h3>{activity.name}</h3>
              <p>{formatActivityType(activity.activity_type)}</p>
            </div>

            <dl>
              <div>
                <dt>Date</dt>
                <dd>{formatActivityDate(activity.started_at)}</dd>
              </div>
              <div>
                <dt>Distance</dt>
                <dd>{formatDistance(activity.distance_meters)}</dd>
              </div>
              <div>
                <dt>Moving time</dt>
                <dd>{formatDuration(activity.moving_seconds)}</dd>
              </div>
              <div>
                <dt>Elevation</dt>
                <dd>{formatElevation(activity.elevation_gain_meters)}</dd>
              </div>
            </dl>
          </li>
        ))}
      </ul>

      {hasHiddenActivities && (
        <button
          className="secondary-action activity-list-action"
          type="button"
          aria-controls="activity-list"
          aria-expanded={showAll}
          onClick={() => setShowAll((currentValue) => !currentValue)}
        >
          {showAll
            ? 'Show fewer'
            : `Show all ${totalActivityCount} activities`}
        </button>
      )}
    </section>
  )
}

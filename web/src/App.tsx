import { useEffect, useState } from 'react'
import './App.css'

type Activity = {
  id: string
  name: string
  activity_type: string
  started_at: string
  distance_meters: number
  duration_seconds: number
}

function App() {
  const [activities, setActivities] = useState<Activity[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function loadActivities() {
      try {
        const response = await fetch('/api/activities')

        if (!response.ok) {
          throw new Error(`Request failed with status ${response.status}`)
        }

        const data: Activity[] | null = await response.json()
        setActivities(data ?? [])
      } catch (error) {
        const message =
          error instanceof Error ? error.message : 'An unknown error occurred'
        setError(message)
      } finally {
        setIsLoading(false)
      }
    }

    loadActivities()
  }, [])

  return (
    <main>
      <header>
        <p>Formpath</p>
        <h1>Activities</h1>
      </header>

      {isLoading && <p className="status">Loading activities...</p>}

      {error && (
        <p className="status status-error">Could not load activities: {error}</p>
      )}

      {!isLoading && !error && activities.length === 0 && (
        <p className="status">No activities found.</p>
      )}

      {!isLoading && !error && activities.length > 0 && (
        <ul className="activity-list">
          {activities.map((activity) => (
            <li className="activity" key={activity.id}>
              <div>
                <h2>{activity.name}</h2>
                <p>{activity.activity_type}</p>
              </div>

              <dl>
                <div>
                  <dt>Date</dt>
                  <dd>{activity.started_at}</dd>
                </div>
                <div>
                  <dt>Distance</dt>
                  <dd>{activity.distance_meters} m</dd>
                </div>
                <div>
                  <dt>Duration</dt>
                  <dd>{activity.duration_seconds} s</dd>
                </div>
              </dl>
            </li>
          ))}
        </ul>
      )}
    </main>
  )
}

export default App

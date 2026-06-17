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

type StravaIntegration = {
  provider: 'strava'
  connected: boolean
}

function App() {
  const [activities, setActivities] = useState<Activity[]>([])
  const [stravaIntegration, setStravaIntegration] =
    useState<StravaIntegration | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isSyncing, setIsSyncing] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [syncError, setSyncError] = useState<string | null>(null)
  const [stravaStatus] = useState<string | null>(() => {
    const params = new URLSearchParams(window.location.search)
    return params.get('strava')
  })

  useEffect(() => {
    async function loadInitialData() {
      try {
        const [activitiesResponse, integrationResponse] = await Promise.all([
          fetch('/api/activities'),
          fetch('/api/integrations/strava'),
        ])

        if (!activitiesResponse.ok) {
          throw new Error(
            `Activities request failed with status ${activitiesResponse.status}`,
          )
        }

        if (!integrationResponse.ok) {
          throw new Error(
            `Strava integration request failed with status ${integrationResponse.status}`,
          )
        }

        const activitiesData: Activity[] | null =
          await activitiesResponse.json()
        const integrationData: StravaIntegration =
          await integrationResponse.json()

        setActivities(activitiesData ?? [])
        setStravaIntegration(integrationData)
      } catch (error) {
        const message =
          error instanceof Error ? error.message : 'An unknown error occurred'
        setError(message)
      } finally {
        setIsLoading(false)
      }
    }

    loadInitialData()
  }, [])

  useEffect(() => {
    if (!stravaStatus) {
      return
    }

    const url = new URL(window.location.href)
    url.searchParams.delete('strava')
    window.history.replaceState({}, '', url)
  }, [stravaStatus])

  async function syncActivities() {
    setIsSyncing(true)
    setSyncError(null)

    try {
      const response = await fetch('/api/activities/sync', {
        method: 'POST',
      })

      if (!response.ok) {
        throw new Error(`Sync request failed with status ${response.status}`)
      }

      const data: Activity[] | null = await response.json()
      setActivities(data ?? [])
    } catch (error) {
      const message =
        error instanceof Error ? error.message : 'An unknown error occurred'
      setSyncError(message)
    } finally {
      setIsSyncing(false)
    }
  }

  return (
    <main>
      <header>
        <p>Formpath</p>
        <h1>Activities</h1>
      </header>

      {stravaStatus === 'connected' && (
        <p className="status status-success">Strava connected successfully.</p>
      )}

      {!isLoading && !error && stravaIntegration && (
        <section className="integration-bar" aria-label="Strava integration">
          <div>
            <p className="integration-label">Strava</p>
            {stravaIntegration.connected ? (
              <p className="integration-copy">Connected and ready to sync.</p>
            ) : (
              <p className="integration-copy">
                Connect Strava to sync your activities.
              </p>
            )}
          </div>

          {stravaIntegration.connected ? (
            <button
              className="primary-action"
              type="button"
              disabled={isSyncing}
              onClick={syncActivities}
            >
              {isSyncing ? 'Syncing...' : 'Sync activities'}
            </button>
          ) : (
            <a className="primary-action" href="/auth/strava">
              Connect with Strava
            </a>
          )}
        </section>
      )}

      {syncError && (
        <p className="status status-error">
          Could not sync activities: {syncError}
        </p>
      )}

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

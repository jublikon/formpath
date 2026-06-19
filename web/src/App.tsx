import { useEffect, useState } from 'react'
import './App.css'
import { IntegrationPanel } from './components/IntegrationPanel'
import { Metric } from './components/Metric'
import { RecentActivities } from './components/RecentActivities'
import { TrainingVolumeChart } from './components/TrainingVolumeChart'
import {
  formatDistance,
  formatDuration,
  formatElevation,
  formatPeriod,
} from './lib/formatters'
import { buildTrainingOverview } from './lib/trainingOverview'
import type { Activity } from './types/activity'

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
      } catch {
        setError(
          "We couldn't load your training data. Please refresh the page.",
        )
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
    } catch {
      setSyncError(
        "We couldn't sync Strava right now. Your existing activities are still available.",
      )
    } finally {
      setIsSyncing(false)
    }
  }

  const overview = buildTrainingOverview(activities, new Date())
  const hasActivities = activities.length > 0
  const isConnected = stravaIntegration?.connected === true
  const hasLoaded = !isLoading && !error

  return (
    <main>
      <header>
        <p>Formpath</p>
        <h1>Training overview</h1>
      </header>

      {stravaStatus === 'connected' && (
        <p className="status status-success" role="status">
          Strava connected successfully.
        </p>
      )}

      {isLoading && (
        <p className="status" role="status">
          Loading your training data...
        </p>
      )}

      {error && (
        <p className="status status-error" role="alert">
          {error}
        </p>
      )}

      {hasLoaded && stravaIntegration && (
        <IntegrationPanel
          connected={isConnected}
          syncing={isSyncing}
          hasActivities={hasActivities}
          onSync={syncActivities}
        />
      )}

      {syncError && (
        <p className="status status-error" role="alert">
          {syncError}
        </p>
      )}

      {hasLoaded && hasActivities && (
        <>
          <section
            className="overview-summary"
            aria-labelledby="overview-heading"
          >
            <div className="section-heading">
              <div>
                <p className="section-eyebrow">Last 28 days</p>
                <h2 id="overview-heading">Your training at a glance</h2>
              </div>
              <p className="period">
                {formatPeriod(
                  overview.period.startDate,
                  overview.period.endDate,
                )}
              </p>
            </div>

            <dl className="metric-grid">
              <Metric
                label="Activities"
                value={new Intl.NumberFormat('en-US').format(
                  overview.totals.activityCount,
                )}
              />
              <Metric
                label="Distance"
                value={formatDistance(overview.totals.distanceMeters)}
              />
              <Metric
                label="Moving time"
                value={formatDuration(overview.totals.movingSeconds)}
              />
              <Metric
                label="Elevation"
                value={formatElevation(overview.totals.elevationGainMeters)}
              />
            </dl>
          </section>

          <TrainingVolumeChart
            days={overview.days}
            startDate={overview.period.startDate}
            endDate={overview.period.endDate}
          />

          <RecentActivities activities={activities} />
        </>
      )}

      {hasLoaded && !hasActivities && (
        <section className="empty-state" aria-labelledby="empty-state-heading">
          <p className="section-eyebrow">
            {isConnected ? 'Ready to sync' : 'Get started'}
          </p>
          <h2 id="empty-state-heading">
            {isConnected
              ? 'Build your first training overview'
              : 'Connect your training data'}
          </h2>
          <p>
            {isConnected
              ? 'Use the sync action above to import your Strava activities.'
              : 'Connect Strava above to import activities and see your recent training volume.'}
          </p>
        </section>
      )}
    </main>
  )
}

export default App

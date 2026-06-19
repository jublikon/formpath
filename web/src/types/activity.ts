export type Activity = {
  id: string
  name: string
  activity_type: string
  started_at: string
  distance_meters: number
  duration_seconds: number
  moving_seconds: number
  elevation_gain_meters?: number
}

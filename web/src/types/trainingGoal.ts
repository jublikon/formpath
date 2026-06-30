export type TrainingGoalSport = 'run' | 'ride'

export type TrainingGoal = {
  id?: string
  user_id: string
  goal_type: 'distance_event'
  sport: TrainingGoalSport
  name: string
  target_distance_meters: number
  target_date: string
  target_duration_seconds?: number
  created_at?: string
  updated_at?: string
}

export type TrainingGoalPayload = {
  sport: TrainingGoalSport
  name: string
  target_distance_meters: number
  target_date: string
  target_duration_seconds?: number
}

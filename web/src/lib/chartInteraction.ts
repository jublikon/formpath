import type { TrainingChartPoint } from './trainingChart'

export function pointerClientXToChartX(
  clientX: number,
  boundsLeft: number,
  boundsWidth: number,
  chartWidth: number,
): number {
  if (boundsWidth <= 0) {
    return 0
  }

  const relativeX = (clientX - boundsLeft) / boundsWidth
  return Math.max(0, Math.min(chartWidth, relativeX * chartWidth))
}

export function findNearestPointIndex(
  points: TrainingChartPoint[],
  x: number,
): number {
  if (points.length === 0) {
    return -1
  }

  return points.reduce((nearestIndex, point, index) => {
    const nearestDistance = Math.abs(points[nearestIndex].x - x)
    const pointDistance = Math.abs(point.x - x)
    return pointDistance < nearestDistance ? index : nearestIndex
  }, 0)
}

export function movePointIndex(
  currentIndex: number | null,
  direction: -1 | 1,
  pointCount: number,
): number | null {
  if (pointCount === 0) {
    return null
  }

  if (currentIndex === null) {
    return direction === 1 ? 0 : pointCount - 1
  }

  return Math.max(0, Math.min(pointCount - 1, currentIndex + direction))
}

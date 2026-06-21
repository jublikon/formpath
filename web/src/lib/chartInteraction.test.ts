import { describe, expect, it } from 'vitest'
import type { TrainingChartPoint } from './trainingChart'
import {
  findNearestPointIndex,
  movePointIndex,
  pointerClientXToChartX,
} from './chartInteraction'

function chartPoints(xPositions: number[]): TrainingChartPoint[] {
  return xPositions.map((x, index) => ({
    date: `2026-06-${String(index + 1).padStart(2, '0')}`,
    value: index * 100,
    x,
    y: 100,
  }))
}

describe('findNearestPointIndex', () => {
  it('selects the point nearest to the horizontal pointer position', () => {
    const points = chartPoints([20, 50, 80])

    expect(findNearestPointIndex(points, 24)).toBe(0)
    expect(findNearestPointIndex(points, 62)).toBe(1)
    expect(findNearestPointIndex(points, 78)).toBe(2)
  })

  it('keeps the earlier point when the pointer is exactly between two points', () => {
    const points = chartPoints([20, 50])

    expect(findNearestPointIndex(points, 35)).toBe(0)
  })

  it('returns no selection for an empty chart', () => {
    expect(findNearestPointIndex([], 20)).toBe(-1)
  })
})

describe('pointerClientXToChartX', () => {
  it('maps a rendered pointer position into chart coordinates', () => {
    expect(pointerClientXToChartX(250, 50, 400, 800)).toBe(400)
  })

  it('clamps pointer positions to the chart bounds', () => {
    expect(pointerClientXToChartX(0, 50, 400, 800)).toBe(0)
    expect(pointerClientXToChartX(500, 50, 400, 800)).toBe(800)
  })

  it('handles a zero-width rendered chart', () => {
    expect(pointerClientXToChartX(250, 50, 0, 800)).toBe(0)
  })
})

describe('movePointIndex', () => {
  it('starts at the appropriate edge when no point is selected', () => {
    expect(movePointIndex(null, 1, 28)).toBe(0)
    expect(movePointIndex(null, -1, 28)).toBe(27)
  })

  it('moves one point and stops at either edge', () => {
    expect(movePointIndex(12, 1, 28)).toBe(13)
    expect(movePointIndex(0, -1, 28)).toBe(0)
    expect(movePointIndex(27, 1, 28)).toBe(27)
  })

  it('returns no selection for an empty chart', () => {
    expect(movePointIndex(null, 1, 0)).toBeNull()
  })
})

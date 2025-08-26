package agvCollider

import "math"

// --------------------- 工具函数 ---------------------

// segmentIntersect 判断两个 Segment 是否相交 / 重合，考虑车宽
// 参数:
//   s1, s2: 两条线段
//   width : AGV宽度
// 返回:
//   bool: 是否相交或重合
//   Point: 相交点或重合区间中心点
func segmentIntersect(s1, s2 Segment, width float64) (bool, Point) {
	dx1 := s1.End.X - s1.Start.X
	dy1 := s1.End.Y - s1.Start.Y
	dx2 := s2.End.X - s2.Start.X
	dy2 := s2.End.Y - s2.Start.Y

	denom := cross(dx1, dy1, dx2, dy2)

	// ============== Case1: 平行/共线 =================
	if denom == 0 {
		// 检查是否共线： (s2.Start - s1.Start) 与 (s1 向量) 是否共线
		if cross(s2.Start.X-s1.Start.X, s2.Start.Y-s1.Start.Y, dx1, dy1) == 0 {
			// 共线 -> 检查投影是否重叠
			if dx1 != 0 { // 横向线段
				minA, maxA := math.Min(s1.Start.X, s1.End.X), math.Max(s1.Start.X, s1.End.X)
				minB, maxB := math.Min(s2.Start.X, s2.End.X), math.Max(s2.Start.X, s2.End.X)
				if math.Max(minA, minB) <= math.Min(maxA, maxB)+width/2 {
					overlapStart := math.Max(minA, minB)
					overlapEnd := math.Min(maxA, maxB)
					ix := (overlapStart + overlapEnd) / 2
					iy := s1.Start.Y
					return true, Point{ix, iy}
				}
			} else { // 纵向线段
				minA, maxA := math.Min(s1.Start.Y, s1.End.Y), math.Max(s1.Start.Y, s1.End.Y)
				minB, maxB := math.Min(s2.Start.Y, s2.End.Y), math.Max(s2.Start.Y, s2.End.Y)
				if math.Max(minA, minB) <= math.Min(maxA, maxB)+width/2 {
					overlapStart := math.Max(minA, minB)
					overlapEnd := math.Min(maxA, maxB)
					iy := (overlapStart + overlapEnd) / 2
					ix := s1.Start.X
					return true, Point{ix, iy}
				}
			}
		}
		// 平行但不共线 → 安全
		return false, Point{}
	}

	// ============== Case2: 非平行 -> 正常相交判定 ==============
	t := cross(s2.Start.X-s1.Start.X, s2.Start.Y-s1.Start.Y, dx2, dy2) / denom
	u := cross(s2.Start.X-s1.Start.X, s2.Start.Y-s1.Start.Y, dx1, dy1) / denom

	if t >= 0 && t <= 1 && u >= 0 && u <= 1 {
		ix := s1.Start.X + t*dx1
		iy := s1.Start.Y + t*dy1
		return true, Point{ix, iy}
	}

	return false, Point{}
}

// 计算点到线段最近点
func closestPointOnSegment(p Point, seg Segment) Point {
	apx := p.X - seg.Start.X
	apy := p.Y - seg.Start.Y
	abx := seg.End.X - seg.Start.X
	aby := seg.End.Y - seg.Start.Y

	ab2 := abx*abx + aby*aby
	if ab2 == 0 {
		return seg.Start
	}
	t := (apx*abx + apy*aby) / ab2
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return Point{seg.Start.X + abx*t, seg.Start.Y + aby*t}
}

// segmentDistance 两Segment最近距离及接触点
func segmentDistance(s1, s2 Segment) (float64, Point) {
	candidates := []Point{
		closestPointOnSegment(s1.Start, s2),
		closestPointOnSegment(s1.End, s2),
		closestPointOnSegment(s2.Start, s1),
		closestPointOnSegment(s2.End, s1),
	}

	minD := math.MaxFloat64
	var best Point
	for _, c := range candidates {
		d := math.Min(
			math.Min(getDistance(s1.Start, c), getDistance(s1.End, c)),
			math.Min(getDistance(s2.Start, c), getDistance(s2.End, c)),
		)
		if d < minD {
			minD = d
			best = c
		}
	}
	return minD, best
}

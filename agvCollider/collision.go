package agvCollider

import "math"

// Collision 表示路径交点信息（用于计算到达时间）
type Collision struct {
	Point     Point   // 碰撞点坐标
	PathADist float64 // A车：从路径起点到该点的累计距离
	PathBDist float64 // B车：从路径起点到该点的累计距离
	TimeA     float64 // A车到达时间 (s)
	TimeB     float64 // B车到达时间 (s)
	TimeDiff  float64 // 两车到达时间差
}

// CollisionEvent 表示两辆AGV之间的最终潜在碰撞事件
type CollisionEvent struct {
	AGV1   *AGV    // 指向第1辆AGV
	AGV2   *AGV    // 指向第2辆AGV
	Point  Point   // 碰撞点
	Time1  float64 // AGV1到达时间
	Time2  float64 // AGV2到达时间
	DeltaT float64 // 时间差
}

// ====================== 碰撞检测逻辑 ======================

// findAllCollisions 找出两条路径的所有交点/重叠点，并计算对应路径长度和时间
// 参数:
//   pathA, pathB: 两条路径
//   vA, vB:       两车速度 (m/s)
//   width:        AGV宽度
// 返回: []Collision，包含所有可能的碰撞事件
func findAllCollisions(pathA, pathB []Point, vA, vB, width float64) []Collision {
	var collisions []Collision
	for i := 0; i < len(pathA)-1; i++ {
		s1 := Segment{Start: pathA[i], End: pathA[i+1]}
		for j := 0; j < len(pathB)-1; j++ {
			s2 := Segment{Start: pathB[j], End: pathB[j+1]}
			// 判断该两段是否有交点/重合
			if ok, inter := segmentIntersect(s1, s2, width); ok {
				// 路径累计长度 = 起点→交点的距离
				sA := pathDistanceToPoint(pathA, inter)
				sB := pathDistanceToPoint(pathB, inter)
				// 到达时间 = 距离 / 速度
				tA := sA / vA
				tB := sB / vB
				// 把这个潜在碰撞事件存入结果集
				collisions = append(collisions, Collision{
					Point:     inter,
					PathADist: sA,
					PathBDist: sB,
					TimeA:     tA,
					TimeB:     tB,
					TimeDiff:  math.Abs(tA - tB),
				})
			}
		}
	}
	return collisions
}

// earliestCollision 从所有交点中选出最早可能的碰撞事件
// 参数:
//   tol: 时间差容忍度 (秒)，小于等于此值认为会相撞
// 返回:
//   bool: 是否存在潜在碰撞
//   Collision: 最早的碰撞事件
func earliestCollision(pathA, pathB []Point, vA, vB, width, tol float64) (bool, Collision) {
	all := findAllCollisions(pathA, pathB, vA, vB, width)
	if len(all) == 0 {
		return false, Collision{}
	}

	minTime := math.MaxFloat64
	var best Collision
	found := false

	for _, c := range all {
		if c.TimeDiff <= tol { // 两车几乎同时到达，才算关键危险点
			earliest := math.Min(c.TimeA, c.TimeB)
			if earliest < minTime {
				minTime = earliest
				best = c
				found = true
			}
		}
	}
	return found, best
}

// DetectCollisionsWithKDTree 使用KD树对所有AGV进行邻居碰撞检测
// 参数：
//   agvs: 所有AGV
//   tol: 时间差容忍度 (s)
//   radius: KD树范围查询半径
func DetectCollisionsWithKDTree(agvs []*AGV, tol, radius float64) []CollisionEvent {
	root := buildKDTree(agvs, 0)
	results := []CollisionEvent{}
	seen := make(map[int]map[int]bool)
	for i := range agvs {
		neighbors := []*AGV{}
		rangeSearch(root, agvs[i], radius, &neighbors)
		for _, other := range neighbors {
			if seen[agvs[i].Id] == nil {
				seen[agvs[i].Id] = make(map[int]bool)
			}
			if seen[other.Id] == nil {
				seen[other.Id] = make(map[int]bool)
			}
			if seen[agvs[i].Id][other.Id] || seen[other.Id][agvs[i].Id] {
				continue
			}
			if ok, event := agvs[i].DetectCollisionWith(other, tol); ok {
				results = append(results, event)
			}
			seen[agvs[i].Id][other.Id] = true
		}
	}
	return results
}

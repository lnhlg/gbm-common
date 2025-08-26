package agvCollider

import "math"

// CollisionPrediction 表示基于位置预测的碰撞信息
type CollisionPrediction struct {
	AGV1               *AGV    // 第一辆AGV
	AGV2               *AGV    // 第二辆AGV
	CollisionTime      float64 // 预测碰撞时间（秒）
	CollisionPoint     Point   // 碰撞点坐标（两车中心的中点）
	AGV1Pose           Pose    // AGV1在碰撞时刻的位姿
	AGV2Pose           Pose    // AGV2在碰撞时刻的位姿
	Distance           float64 // 碰撞时刻两车中心距离
	CollisionThreshold float64 // 碰撞距离阈值
}

// calculateOptimalSearchRadius 计算最优搜索半径
func calculateOptimalSearchRadius(agvs []*AGV, timeRange, collisionThreshold float64) float64 {
	maxWidth := 0.0
	maxSpeed := 0.0

	for _, agv := range agvs {
		if agv.Width > maxWidth {
			maxWidth = agv.Width
		}
		if agv.Speed > maxSpeed {
			maxSpeed = agv.Speed
		}
	}

	// 搜索半径 = 最大车宽 + 最大速度 * 时间范围 + 碰撞阈值 + 安全边距
	safetyMargin := maxWidth * 0.5 // 50%的安全边距
	return maxWidth + maxSpeed*timeRange + collisionThreshold + safetyMargin
}

// GetCollisionRiskLevel 根据碰撞预测信息评估碰撞风险等级
// 参数:
//   collision: 碰撞预测信息
// 返回:
//   string: 风险等级描述
func (cp *CollisionPrediction) GetCollisionRiskLevel() string {
	// 根据碰撞时间和距离评估风险
	if cp.CollisionTime <= 1.0 {
		return "极高风险" // 1秒内碰撞
	} else if cp.CollisionTime <= 3.0 {
		return "高风险" // 3秒内碰撞
	} else if cp.CollisionTime <= 5.0 {
		return "中等风险" // 5秒内碰撞
	} else {
		return "低风险" // 5秒后碰撞
	}
}

// predictCollisionsForFleetWithKDTree 使用KD树优化的车队碰撞检测
func predictCollisionsForFleetWithKDTree(agvs []*AGV, timeRange, timeStep, collisionThreshold float64) []CollisionPrediction {
	var collisions []CollisionPrediction
	seen := make(map[int]map[int]bool)

	// 构建KD树
	root := buildKDTree(agvs, 0)

	// 计算动态搜索半径
	searchRadius := calculateOptimalSearchRadius(agvs, timeRange, collisionThreshold)

	for _, agv1 := range agvs {
		// 使用KD树查找潜在邻居
		neighbors := []*AGV{}
		rangeSearch(root, agv1, searchRadius, &neighbors)

		for _, agv2 := range neighbors {
			// 避免重复检查同一对AGV
			if agv1.Id >= agv2.Id {
				continue
			}

			// 检查是否已经处理过这对AGV
			if seen[agv1.Id] == nil {
				seen[agv1.Id] = make(map[int]bool)
			}
			if seen[agv2.Id] == nil {
				seen[agv2.Id] = make(map[int]bool)
			}
			if seen[agv1.Id][agv2.Id] || seen[agv2.Id][agv1.Id] {
				continue
			}

			// 快速预筛选：检查初始距离
			initialDistance := math.Hypot(agv1.Pose.X-agv2.Pose.X, agv1.Pose.Y-agv2.Pose.Y)
			if initialDistance > searchRadius {
				continue
			}

			// 检测碰撞
			if hasCollision, collision := agv1.PredictCollisionWith(agv2, timeRange, timeStep, collisionThreshold); hasCollision {
				collisions = append(collisions, collision)
			}

			// 标记已处理
			seen[agv1.Id][agv2.Id] = true
		}
	}

	return collisions
}

// PredictCollisionsForFleet 检测AGV车队中所有可能的碰撞
// 参数:
//   agvs: AGV车队
//   timeRange: 预测时间范围（秒）
//   timeStep: 时间步长（秒）
//   collisionThreshold: 碰撞距离阈值（米）
// 返回:
//   []CollisionPrediction: 所有预测的碰撞事件
func PredictCollisionsForFleet(agvs []*AGV, timeRange, timeStep, collisionThreshold float64) []CollisionPrediction {
	var collisions []CollisionPrediction
	seen := make(map[int]map[int]bool)

	// 使用KD树优化邻居搜索
	// 构建KD树用于空间索引
	root := buildKDTree(agvs, 0)

	// 计算搜索半径：基于AGV最大宽度和预测时间范围
	maxWidth := 0.0
	maxSpeed := 0.0
	for _, agv := range agvs {
		if agv.Width > maxWidth {
			maxWidth = agv.Width
		}
		if agv.Speed > maxSpeed {
			maxSpeed = agv.Speed
		}
	}

	// 搜索半径 = 最大车宽 + 最大速度 * 时间范围 + 碰撞阈值
	searchRadius := maxWidth + maxSpeed*timeRange + collisionThreshold

	for i := range agvs {
		// 使用KD树查找潜在邻居
		neighbors := []*AGV{}
		rangeSearch(root, agvs[i], searchRadius, &neighbors)

		for j := range neighbors {
			// 避免重复检查同一对AGV
			if i >= j {
				continue
			}

			// 检查是否已经处理过这对AGV
			if seen[agvs[i].Id] == nil {
				seen[agvs[i].Id] = make(map[int]bool)
			}
			if seen[agvs[j].Id] == nil {
				seen[agvs[j].Id] = make(map[int]bool)
			}
			if seen[agvs[i].Id][agvs[j].Id] || seen[agvs[j].Id][agvs[i].Id] {
				continue
			}

			// 检测碰撞
			if hasCollision, collision := agvs[i].PredictCollisionWith(agvs[j], timeRange, timeStep, collisionThreshold); hasCollision {
				collisions = append(collisions, collision)
			}

			// 标记已处理
			seen[agvs[i].Id][agvs[j].Id] = true
		}
	}

	return collisions
}

// PredictCollisionsForFleetOptimized 使用更高级优化的车队碰撞检测
// 参数:
//   agvs: AGV车队
//   timeRange: 预测时间范围（秒）
//   timeStep: 时间步长（秒）
//   collisionThreshold: 碰撞距离阈值（米）
//   useSpatialIndex: 是否使用空间索引优化
// 返回:
//   []CollisionPrediction: 所有预测的碰撞事件
func PredictCollisionsForFleetOptimized(agvs []*AGV, timeRange, timeStep, collisionThreshold float64, useSpatialIndex bool) []CollisionPrediction {
	var collisions []CollisionPrediction
	seen := make(map[int]map[int]bool)

	if useSpatialIndex && len(agvs) > 10 {
		// 对于大型车队，使用KD树优化
		return predictCollisionsForFleetWithKDTree(agvs, timeRange, timeStep, collisionThreshold)
	}

	// 对于小型车队，使用原始方法
	for i := range agvs {
		for j := range agvs {
			if i >= j {
				continue // 避免重复检查同一对AGV
			}

			// 检查是否已经处理过这对AGV
			if seen[agvs[i].Id] == nil {
				seen[agvs[i].Id] = make(map[int]bool)
			}
			if seen[agvs[j].Id] == nil {
				seen[agvs[j].Id] = make(map[int]bool)
			}
			if seen[agvs[i].Id][agvs[j].Id] || seen[agvs[j].Id][agvs[i].Id] {
				continue
			}

			// 检测碰撞
			if hasCollision, collision := agvs[i].PredictCollisionWith(agvs[j], timeRange, timeStep, collisionThreshold); hasCollision {
				collisions = append(collisions, collision)
			}

			// 标记已处理
			seen[agvs[i].Id][agvs[j].Id] = true
		}
	}

	return collisions
}

package agvCollider

import "math"

// ===================== 基本数据结构 =====================

// Pose 表示AGV的位姿（位置+朝向）
// - X: 横坐标
// - Y: 纵坐标
// - T: 航向角，单位弧度（0表示朝向x正方向）
type Pose struct {
	X float64
	Y float64
	T float64
}

// Point 表示二维平面上的一个点（无方向）
// - X: 横坐标
// - Y: 纵坐标
type Point struct {
	X float64
	Y float64
}

// Segment 表示路径中的一条线段
// - Start: 起点
// - End:   终点
type Segment struct {
	Start Point
	End   Point
}

// AGV 表示自动引导车结构体
// - Pose:     当前位姿（位置+方向）
// - Width:  AGV的宽度（m）
// - Speed:    行驶速度（m/s）
// - Path:     全局路径（所有规划好的路径点）
// - SubPath:  当前子路径（缓存，用于避免每次从全局路径起点计算）
// - InitDone: 是否已经初始化过子路径（首次计算需要用全局路径）
type AGV struct {
	Id       int
	Width    float64
	Pose     Pose
	Speed    float64
	Path     []Point
	SubPath  []Point
	InitDone bool
}

// ===================== 基础工具函数 =====================

// getDistance 计算两点之间的欧式距离
// 参数:
//   p1, p2: 两个点
// 返回:
//   float64: 距离
func getDistance(p1, p2 Point) float64 {
	return math.Hypot(p2.X-p1.X, p2.Y-p1.Y)
}

// dot 向量点积
// 参数:
//   ax, ay: 向量A
//   bx, by: 向量B
// 返回:
//   float64: 点积结果
func dot(ax, ay, bx, by float64) float64 {
	return ax*bx + ay*by
}

// cross 向量叉积
// 参数:
//   ax, ay: 向量A
//   bx, by: 向量B
// 返回:
//   float64: 叉积结果
// 说明:
//   - 结果为正时, 表示向量A在向量B的逆时针方向
//   - 结果为负时, 表示向量A在向量B的顺时针方向
//   - 结果为0时, 表示向量A和向量B共线
func cross(ax, ay, bx, by float64) float64 {
	return ax*by - ay*bx
}

// interpolate 在两点之间进行插值
// 参数:
//   p1: 起点
//   p2: 终点
//   ratio: 插值比例 (0=起点, 1=终点)
// 返回:
//   Point: 插值得到的新点
func interpolate(p1, p2 Point, ratio float64) Point {
	return Point{
		X: p1.X + (p2.X-p1.X)*ratio,
		Y: p1.Y + (p2.Y-p1.Y)*ratio,
	}
}

// projectPointOnSegment 将一个点投影到一条线段上
// 参数:
//   pose: 待投影点（AGV的当前位置）
//   seg:  路径段
// 返回:
//   Point: 投影后的点坐标
//   t:     投影比例 (0=落在seg.Start, 1=落在seg.End)
func projectPointOnSegment(pose Pose, seg Segment) (Point, float64) {
	// 路径段 向量
	vx := seg.End.X - seg.Start.X
	vy := seg.End.Y - seg.Start.Y
	// 点到起点的向量
	wx := pose.X - seg.Start.X
	wy := pose.Y - seg.Start.Y

	segLen2 := vx*vx + vy*vy
	if segLen2 == 0 {
		// 起点和终点重合 -> 投影点就是起点
		return seg.Start, 0
	}
	// 投影比例
	t := dot(wx, wy, vx, vy) / segLen2

	// 约束 t ∈ [0,1]
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	// 计算投影点
	p := Point{
		X: seg.Start.X + t*vx,
		Y: seg.Start.Y + t*vy,
	}
	return p, t
}

// ===================== AGV方法 =====================

// GenerateSubPath 从当前Pose生成子路径
// 逻辑:
//   - 如果是首次调用, 从全局Path投影, 生成子路径
//   - 如果已有SubPath缓存, 则从SubPath起始段开始计算, 节省开销
// 返回:
//   []Point: 新的子路径（起点为投影点, 包含后续路径点）
func (agv *AGV) GenerateSubPath() []Point {
	var basePath []Point

	// 首次计算 → 使用全局路径
	if !agv.InitDone || len(agv.SubPath) < 2 {
		basePath = agv.Path
		agv.InitDone = true
	} else {
		// 后续计算 → 使用上次的子路径
		basePath = agv.SubPath
	}

	n := len(basePath)
	if n < 2 {
		return basePath
	}

	minDist := math.MaxFloat64
	var segIdx int
	var proj Point

	// 遍历路径段，找到距离AGV最近的投影点
	for i := 0; i < n-1; i++ {
		seg := Segment{Start: basePath[i], End: basePath[i+1]}
		p, _ := projectPointOnSegment(agv.Pose, seg)
		d := getDistance(Point{agv.Pose.X, agv.Pose.Y}, p)
		if d < minDist {
			minDist = d
			segIdx = i
			proj = p
		}
	}

	// 构造新的子路径 ＝ [投影点 + 后续路径点]
	newPath := []Point{}
	newPath = append(newPath, proj)
	for j := segIdx + 1; j < n; j++ {
		newPath = append(newPath, basePath[j])
	}

	// 更新缓存
	agv.SubPath = newPath
	return newPath
}

// PredictPosition 预测AGV在dt秒后的位姿
// 步骤:
//   1. 调用 GenerateSubPath 获取子路径（自动复用缓存）
//   2. 计算子路径的累计里程表
//   3. 根据 v*dt 找到目标距离 targetS
//   4. 在目标处进行插值，得到预测位置和方向
// 参数:
//   dt: 预测的时间间隔，单位秒
// 返回:
//   Pose: 预测出的位姿
func (agv *AGV) PredictPosition(dt float64) Pose {
	newPath := agv.GenerateSubPath()
	n := len(newPath)
	if n < 2 {
		return agv.Pose
	}

	// Step1: 路径每段的长度 & 累计长度
	segmentLengths := make([]float64, n-1)
	cumulativeLengths := make([]float64, n)
	for i := 1; i < n; i++ {
		segmentLengths[i-1] = getDistance(newPath[i-1], newPath[i])
		cumulativeLengths[i] = cumulativeLengths[i-1] + segmentLengths[i-1]
	}

	// Step2: 目标行驶距离 S = v * dt
	targetS := agv.Speed * dt
	totalLen := cumulativeLengths[n-1]

	// Step3: 如果超出路径总长，直接返回最后一个点
	if targetS >= totalLen {
		last := newPath[n-1]
		// 航向角取最后一段的方向
		dx := newPath[n-1].X - newPath[n-2].X
		dy := newPath[n-1].Y - newPath[n-2].Y
		theta := math.Atan2(dy, dx)
		agv.Pose = Pose{X: last.X, Y: last.Y, T: theta}
		return agv.Pose
	}

	// Step4: 找出targetS在哪个路径段上
	var segIdx int
	for i := 1; i < n; i++ {
		if targetS <= cumulativeLengths[i] {
			segIdx = i - 1
			break
		}
	}

	// Step5: 在该段上进行插值
	segStart := newPath[segIdx]
	segEnd := newPath[segIdx+1]
	segLen := segmentLengths[segIdx]

	distOnSeg := targetS - cumulativeLengths[segIdx]
	ratio := distOnSeg / segLen

	pt := interpolate(segStart, segEnd, ratio)
	theta := math.Atan2(segEnd.Y-segStart.Y, segEnd.X-segStart.X)

	// 更新AGV姿态并返回
	agv.Pose = Pose{X: pt.X, Y: pt.Y, T: theta}
	return agv.Pose
}

// ====================== AGV方法扩展 ======================

// DetectCollisionWith 检测当前AGV和另一辆AGV的潜在碰撞
func (agv *AGV) DetectCollisionWith(other *AGV, tol float64) (bool, CollisionEvent) {
	ok, col := earliestCollision(
		agv.Path, other.Path,
		agv.Speed, other.Speed,
		(agv.Width+other.Width)/2, tol,
	)
	if ok {
		return true, CollisionEvent{
			AGV1:   agv,
			AGV2:   other,
			Point:  col.Point,
			Time1:  col.TimeA,
			Time2:  col.TimeB,
			DeltaT: col.TimeDiff,
		}
	}
	return false, CollisionEvent{}
}

// ====================== 基于PredictPosition的碰撞检测 ======================

// PredictCollisionWith 使用PredictPosition方法检测两辆AGV是否会相撞
// 参数:
//   other: 另一辆AGV
//   timeRange: 预测时间范围（秒），默认检查0到timeRange秒内的所有时间点
//   timeStep: 时间步长（秒），用于离散化时间检查
//   collisionThreshold: 碰撞距离阈值（米），两车中心距离小于此值认为碰撞
// 返回:
//   bool: 是否会发生碰撞
//   CollisionPrediction: 碰撞预测信息
func (agv *AGV) PredictCollisionWith(other *AGV, timeRange, timeStep, collisionThreshold float64) (bool, CollisionPrediction) {
	if timeStep <= 0 {
		timeStep = 0.1 // 默认0.1秒步长
	}
	if collisionThreshold <= 0 {
		collisionThreshold = (agv.Width + other.Width) / 2 // 默认使用两车半宽之和
	}

	var earliestCollision CollisionPrediction
	earliestTime := timeRange + 1 // 初始化为超出范围的值
	found := false

	// 离散化时间检查
	for t := 0.0; t <= timeRange; t += timeStep {
		// 预测两车在时间t的位置
		pose1 := agv.PredictPosition(t)
		pose2 := other.PredictPosition(t)

		// 计算两车中心距离
		distance := math.Hypot(pose1.X-pose2.X, pose1.Y-pose2.Y)

		// 检查是否碰撞
		if distance <= collisionThreshold {
			// 找到碰撞，记录最早的时间
			if t < earliestTime {
				earliestTime = t
				earliestCollision = CollisionPrediction{
					AGV1:               agv,
					AGV2:               other,
					CollisionTime:      t,
					CollisionPoint:     Point{(pose1.X + pose2.X) / 2, (pose1.Y + pose2.Y) / 2},
					AGV1Pose:           pose1,
					AGV2Pose:           pose2,
					Distance:           distance,
					CollisionThreshold: collisionThreshold,
				}
				found = true
			}
		}
	}

	return found, earliestCollision
}

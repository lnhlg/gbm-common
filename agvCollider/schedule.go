package agvCollider

// ScheduleAction 调度动作
type ScheduleAction struct {
	AGV       *AGV
	Action    string         // "GO" 或 "WAIT"
	WaitTime  float64        // 等待时间 (s)
	Collision CollisionEvent // 对应的冲突事件
}

// ResolveCollision 自动调度决策
// 参数：
//   e: 碰撞事件
//   safeGap: 安全时间间隔 (秒)，要求后一辆车至少等待这么久
// 返回：两个调度动作（一个GO，一个WAIT）
func ResolveCollision(e CollisionEvent, safeGap float64) (ScheduleAction, ScheduleAction) {

	// 决定谁优先通过
	if e.Time1 <= e.Time2 {
		// AGV1先到 → GO，AGV2等待
		return ScheduleAction{
				AGV: e.AGV1, Action: "GO", WaitTime: 0, Collision: e,
			}, ScheduleAction{
				AGV: e.AGV2, Action: "WAIT",
				WaitTime:  (e.Time1 + safeGap) - e.Time2,
				Collision: e,
			}
	} else {
		// AGV2先到 → GO，AGV1等待
		return ScheduleAction{
				AGV: e.AGV2, Action: "GO", WaitTime: 0, Collision: e,
			}, ScheduleAction{
				AGV: e.AGV1, Action: "WAIT",
				WaitTime:  (e.Time2 + safeGap) - e.Time1,
				Collision: e,
			}
	}
}

// DetectAndSchedule 使用KD树检测潜在碰撞并下发调度建议
func DetectAndSchedule(agvs []*AGV, tol, radius, safeGap float64) []ScheduleAction {
	events := DetectCollisionsWithKDTree(agvs, tol, radius)
	actions := []ScheduleAction{}
	for _, e := range events {
		a1, a2 := ResolveCollision(e, safeGap)
		actions = append(actions, a1, a2)
	}
	return actions
}

package agvCollider

import (
	"math"
	"sort"
)

// KDNode 表示KD树节点
type KDNode struct {
	AGV   *AGV    // 当前节点存储的AGV
	Left  *KDNode // 左子树
	Right *KDNode // 右子树
	Depth int     // 当前节点深度，用来选择分割轴
}

// buildKDTree 递归构建KD树
func buildKDTree(agvs []*AGV, depth int) *KDNode {
	if len(agvs) == 0 {
		return nil
	}
	// 根据当前层决定比较的坐标轴 (0=X轴, 1=Y轴)
	axis := depth % 2

	// 排序
	sort.Slice(agvs, func(i, j int) bool {
		if axis == 0 {
			return agvs[i].Pose.X < agvs[j].Pose.X
		}
		return agvs[i].Pose.Y < agvs[j].Pose.Y
	})

	// 取中位数作为分割点
	median := len(agvs) / 2

	// 递归构造
	return &KDNode{
		AGV:   agvs[median],
		Left:  buildKDTree(agvs[:median], depth+1),
		Right: buildKDTree(agvs[median+1:], depth+1),
		Depth: depth,
	}
}

// distance 计算两AGV欧式距离
func distance(a, b *AGV) float64 {
	dx := a.Pose.X - b.Pose.X
	dy := a.Pose.Y - b.Pose.Y
	return math.Hypot(dx, dy)
}

// rangeSearch 在KD树中查询目标AGV半径r的邻居
func rangeSearch(node *KDNode, target *AGV, r float64, results *[]*AGV) {
	if node == nil {
		return
	}

	// Step 1: 检查当前节点
	if node.AGV.Id != target.Id && distance(node.AGV, target) <= r {
		*results = append(*results, node.AGV)
	}

	// Step 2: 判定搜索方向
	axis := node.Depth % 2
	var diff float64
	if axis == 0 {
		diff = target.Pose.X - node.AGV.Pose.X
	} else {
		diff = target.Pose.Y - node.AGV.Pose.Y
	}

	// Step 3: 搜索对应子树
	if diff <= 0 {
		rangeSearch(node.Left, target, r, results)
		if math.Abs(diff) <= r {
			rangeSearch(node.Right, target, r, results)
		}
	} else {
		rangeSearch(node.Right, target, r, results)
		if math.Abs(diff) <= r {
			rangeSearch(node.Left, target, r, results)
		}
	}
}

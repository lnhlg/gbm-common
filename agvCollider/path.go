package agvCollider

import "math"

// --------------------- 核心: 路径交点检测 ---------------------

// checkPathIntersection 检查两条路径是否存在相交/重叠
// 参数:
//   pathA, pathB: 两条路径(点集)
//   width: AGV宽度
// 返回:
//   bool: 是否发现相交/重叠
//   Point: 最近的相交点/重叠点
func checkPathIntersection(pathA, pathB []Point, width float64) (bool, Point) {
	minDist := math.MaxFloat64
	var nearest Point
	found := false

	// 遍历所有路径段组合
	for i := 0; i < len(pathA)-1; i++ {
		s1 := Segment{Start: pathA[i], End: pathA[i+1]}
		for j := 0; j < len(pathB)-1; j++ {
			s2 := Segment{Start: pathB[j], End: pathB[j+1]}

			// 先检测是否直接相交或共线重叠
			if ok, inter := segmentIntersect(s1, s2, width); ok {
				found = true
				d := getDistance(pathA[0], inter)
				if d < minDist {
					minDist = d
					nearest = inter
				}
			} else {
				// 再考虑「最近距离 < 半车宽」时，算作碰撞
				d, pt := segmentDistance(s1, s2)
				if d <= width/2.0 {
					found = true
					if d < minDist {
						minDist = d
						nearest = pt
					}
				}
			}
		}
	}
	return found, nearest
}

// pathDistanceToPoint 计算路径起点到指定点的累计路径长度
// 参数:
//   path: 路径点集
//   p   : 目标点（相交点）
// 返回:
//   float64: 从路径起点到目标点的距离
func pathDistanceToPoint(path []Point, p Point) float64 {
	total := 0.0
	for i := 0; i < len(path)-1; i++ {
		seg := Segment{Start: path[i], End: path[i+1]}
		// 判断p是否在该段上(在直线上且投影在段内)
		// 方法: 向量叉积==0 且 点在范围内
		if cross(p.X-seg.Start.X, p.Y-seg.Start.Y,
			seg.End.X-seg.Start.X, seg.End.Y-seg.Start.Y) == 0 &&
			p.X >= math.Min(seg.Start.X, seg.End.X)-1e-6 &&
			p.X <= math.Max(seg.Start.X, seg.End.X)+1e-6 &&
			p.Y >= math.Min(seg.Start.Y, seg.End.Y)-1e-6 &&
			p.Y <= math.Max(seg.Start.Y, seg.End.Y)+1e-6 {
			// 点在该段
			total += getDistance(seg.Start, p)
			return total
		}
		// 否则整段加入累计
		total += getDistance(seg.Start, seg.End)
	}
	return total
}

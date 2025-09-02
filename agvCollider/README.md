# AGV碰撞检测系统

这是一个用于检测两台或多台AGV（自动导引车）之间碰撞的Go语言实现。

## 功能特性

- **实时碰撞检测**: 基于当前位置和速度进行实时碰撞检测
- **预测性碰撞检测**: 预测未来时间范围内的潜在碰撞
- **多AGV支持**: 支持同时检测多台AGV之间的碰撞
- **可配置参数**: 可自定义安全距离和预测时间范围
- **高性能**: 使用数学优化算法，支持高频检测

## 核心算法

### 碰撞检测原理

1. **静态碰撞检测**: 当两台AGV之间的距离小于安全距离时，立即判定为碰撞
2. **动态碰撞预测**: 基于AGV的运动轨迹和速度，预测未来是否会发生碰撞
3. **二次方程求解**: 使用二次方程求解碰撞时间点

### 数学公式

对于两台AGV，碰撞检测基于以下方程：

```
(dx + vrx*t)² + (dy + vry*t)² = (width1/2 + width2/2)²
```

其中：
- `dx, dy`: 两AGV之间的相对位置
- `vrx, vry`: 相对速度分量
- `t`: 碰撞时间
- `width1, width2`: AGV宽度

## 使用方法

### 基本用法

```go
package main

import (
    "fmt"
    "math"
    "your-project/agvCollider"
)

func main() {
    // 创建碰撞检测器
    detector := agvCollider.NewCollisionDetector(1.0, 5.0) // 安全距离1米，预测5秒
    
    // 创建AGV
    agv1 := agvCollider.CreateAGV(1, 0, 0, 0, 2.0, nil)        // 位置(0,0)，向东2m/s
    agv2 := agvCollider.CreateAGV(2, 10, 0, math.Pi, 2.0, nil) // 位置(10,0)，向西2m/s
    
    // 检测碰撞
    result := detector.CheckCollision(agv1, agv2)
    
    if result.IsCollision {
        fmt.Printf("检测到碰撞！距离: %.2fm, 预计时间: %.2fs\n", 
            result.Distance, result.TimeToCollision)
    }
}
```

### 多AGV检测

```go
// 创建多台AGV
agvs := []*agvCollider.AGV{
    agvCollider.CreateAGV(1, 0, 0, 0, 2.0, nil),
    agvCollider.CreateAGV(2, 5, 0, math.Pi, 1.5, nil),
    agvCollider.CreateAGV(3, 0, 5, math.Pi/2, 1.0, nil),
}

// 检测所有AGV之间的碰撞
results := detector.CheckMultipleAGVs(agvs)
for _, result := range results {
    if result.IsCollision {
        fmt.Printf("碰撞检测: 距离=%.2fm, 时间=%.2fs\n", 
            result.Distance, result.TimeToCollision)
    }
}
```

### 实时位置更新

```go
// 更新AGV位置信息
agv1.UpdateAGVPosition(5, 5, math.Pi/2, 1.5) // 新位置(5,5)，向北1.5m/s

// 重新检测碰撞
result := detector.CheckCollision(agv1, agv2)
```

## API文档

### 结构体

#### AGV
```go
type AGV struct {
    ID        int           // AGV唯一标识
    Position  Position      // 当前位置
    Velocity  float64       // 当前速度 (m/s)
    Width     float64       // AGV宽度 (默认0.5米)
    Path      [][]float64   // 全局路径
    Timestamp time.Time     // 时间戳
}
```

#### Position
```go
type Position struct {
    X     float64 // X坐标
    Y     float64 // Y坐标
    Theta float64 // 弧度方向
}
```

#### CollisionResult
```go
type CollisionResult struct {
    IsCollision      bool    // 是否发生碰撞
    Distance         float64 // 两AGV之间的距离
    TimeToCollision  float64 // 预计碰撞时间 (秒)
    CollisionPoint   Point   // 碰撞点坐标
}
```

### 主要函数

#### NewCollisionDetector
```go
func NewCollisionDetector(safetyDistance, timeHorizon float64) *CollisionDetector
```
创建新的碰撞检测器
- `safetyDistance`: 安全距离（米）
- `timeHorizon`: 预测时间范围（秒）

#### CheckCollision
```go
func (cd *CollisionDetector) CheckCollision(agv1, agv2 *AGV) CollisionResult
```
检查两台AGV是否相撞

#### CheckMultipleAGVs
```go
func (cd *CollisionDetector) CheckMultipleAGVs(agvs []*AGV) []CollisionResult
```
检查多个AGV之间的碰撞

#### CreateAGV
```go
func CreateAGV(id int, x, y, theta, velocity float64, path [][]float64) *AGV
```
创建新的AGV实例

#### UpdateAGVPosition
```go
func (agv *AGV) UpdateAGVPosition(x, y, theta, velocity float64)
```
更新AGV位置信息

## 运行测试

```bash
# 运行所有测试
go test ./agvCollider

# 运行基准测试
go test -bench=. ./agvCollider

# 运行示例
go run agvCollider/example.go
```

## 性能特点

- **时间复杂度**: O(n²) 用于n台AGV的碰撞检测
- **空间复杂度**: O(1) 用于单次碰撞检测
- **精度**: 支持毫米级精度的碰撞检测
- **实时性**: 单次检测耗时 < 1ms

## 注意事项

1. **坐标系统**: 使用笛卡尔坐标系，X轴向右为正，Y轴向上为正
2. **角度单位**: 方向角度使用弧度制，0表示向东，π/2表示向北
3. **AGV宽度**: 默认AGV宽度为0.5米，可通过修改AGV结构体自定义
4. **安全距离**: 建议设置安全距离大于AGV宽度，以确保安全缓冲
5. **预测时间**: 根据实际应用场景调整预测时间范围

## 扩展功能

- 支持AGV路径规划集成
- 支持动态障碍物检测
- 支持碰撞避免策略
- 支持3D空间碰撞检测



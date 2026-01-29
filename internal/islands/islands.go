package islands

type Cell struct {
	X int
	Y int
}

type Island struct {
	Cells []Cell
	MinX  int
	MinY  int
	MaxX  int
	MaxY  int
}

type Connectivity int

const (
	Connectivity4 Connectivity = iota
	Connectivity8
)

type CornerMask uint8

const (
	CornerTopLeft CornerMask = 1 << iota
	CornerTopRight
	CornerBottomRight
	CornerBottomLeft
)

func (mask CornerMask) Has(flag CornerMask) bool {
	return mask&flag != 0
}

func FindIslands(matrix [][]bool) []Island {
	return FindIslandsWithConnectivity(matrix, Connectivity4)
}

func FindIslandsWithConnectivity(matrix [][]bool, connectivity Connectivity) []Island {
	if len(matrix) == 0 {
		return nil
	}
	visited := make([][]bool, len(matrix))
	for i := range matrix {
		visited[i] = make([]bool, len(matrix[i]))
	}

	var islands []Island
	for y, row := range matrix {
		for x, cell := range row {
			if !cell || visited[y][x] {
				continue
			}
			island := Island{MinX: x, MaxX: x, MinY: y, MaxY: y}
			queue := []Cell{{X: x, Y: y}}
			visited[y][x] = true
			for len(queue) > 0 {
				current := queue[0]
				queue = queue[1:]
				island.Cells = append(island.Cells, current)
				if current.X < island.MinX {
					island.MinX = current.X
				}
				if current.X > island.MaxX {
					island.MaxX = current.X
				}
				if current.Y < island.MinY {
					island.MinY = current.Y
				}
				if current.Y > island.MaxY {
					island.MaxY = current.Y
				}
				neighbors := []Cell{
					{X: current.X + 1, Y: current.Y},
					{X: current.X - 1, Y: current.Y},
					{X: current.X, Y: current.Y + 1},
					{X: current.X, Y: current.Y - 1},
				}
				if connectivity == Connectivity8 {
					neighbors = append(neighbors,
						Cell{X: current.X + 1, Y: current.Y + 1},
						Cell{X: current.X - 1, Y: current.Y - 1},
						Cell{X: current.X + 1, Y: current.Y - 1},
						Cell{X: current.X - 1, Y: current.Y + 1},
					)
				}
				for _, neighbor := range neighbors {
					if neighbor.Y < 0 || neighbor.Y >= len(matrix) {
						continue
					}
					if neighbor.X < 0 || neighbor.X >= len(matrix[neighbor.Y]) {
						continue
					}
					if visited[neighbor.Y][neighbor.X] {
						continue
					}
					if !matrix[neighbor.Y][neighbor.X] {
						continue
					}
					visited[neighbor.Y][neighbor.X] = true
					queue = append(queue, neighbor)
				}
			}
			islands = append(islands, island)
		}
	}
	return islands
}

func CornerMaskAt(matrix [][]bool, x int, y int, connectivity Connectivity) CornerMask {
	if y < 0 || y >= len(matrix) {
		return 0
	}
	if x < 0 || x >= len(matrix[y]) {
		return 0
	}
	if !matrix[y][x] {
		return 0
	}
	up := y > 0 && x < len(matrix[y-1]) && matrix[y-1][x]
	down := y+1 < len(matrix) && x < len(matrix[y+1]) && matrix[y+1][x]
	left := x > 0 && matrix[y][x-1]
	right := x+1 < len(matrix[y]) && matrix[y][x+1]
	diagTL := connectivity == Connectivity8 && y > 0 && x > 0 && matrix[y-1][x-1]
	diagTR := connectivity == Connectivity8 && y > 0 && x+1 < len(matrix[y-1]) && matrix[y-1][x+1]
	diagBR := connectivity == Connectivity8 && y+1 < len(matrix) && x+1 < len(matrix[y+1]) && matrix[y+1][x+1]
	diagBL := connectivity == Connectivity8 && y+1 < len(matrix) && x > 0 && matrix[y+1][x-1]

	var mask CornerMask
	if !up && !left && !diagTL {
		mask |= CornerTopLeft
	}
	if !up && !right && !diagTR {
		mask |= CornerTopRight
	}
	if !down && !right && !diagBR {
		mask |= CornerBottomRight
	}
	if !down && !left && !diagBL {
		mask |= CornerBottomLeft
	}
	return mask
}

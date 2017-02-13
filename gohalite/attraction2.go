package gohalite

import (
    "sort"
)

// -------------------------------------------------------------------------------------------------------------
// Sort related stuff

type CellAndScore struct {
    Index       int
    Score       int
    Distance    int
}

type CellScoreQueue []CellAndScore

func (s CellScoreQueue) Len() int {
    return len(s)
}
func (s CellScoreQueue) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}
func (s CellScoreQueue) Less(i, j int) bool {
    if s[i].Score < s[j].Score {
        return true
    } else if s[i].Score > s[j].Score {
        return false
    } else {
        return s[i].Index < s[j].Index      // Break ties deterministically
    }
}

// -------------------------------------------------------------------------------------------------------------

func (g *Game) SimpleBorderScore(i int) int {               // Used in v41 onwards...
    return g.Production[i] * 4 - (g.Strength[i] * 7) / 10
}

func (g *Game) AttractionMap_v2(nice_minimum int) ([]int, []int) {

    result := make([]int, g.Size)
    distances := make([]int, g.Size)

    for i := 0 ; i < g.Size ; i++ {
        result[i] = -2147483647
    }

    frontier_neutrals := g.ListTouchingNiceNeutrals(nice_minimum)

    var queue CellScoreQueue

    for _, i := range frontier_neutrals {
        result[i] = g.SimpleBorderScore(i)
        distances[i] = 0                            // Redundant, it's 0 already, but for clarity
        c_s := CellAndScore{
            Index: i,
            Score: result[i],
            Distance: 0,
        }
        queue = append(queue, c_s)
    }

    sort.Sort(sort.Reverse(queue))

    for {
        if len(queue) == 0 {
            break
        }

        propagator := queue[0]
        queue = queue[1:]

        i := propagator.Index

        for _, neighbour := range g.Neighbours[i] {

            if g.Owner[neighbour.Index] == g.Id && result[neighbour.Index] == -2147483647 {

                result[neighbour.Index] = propagator.Score - (g.Production[neighbour.Index] + 2)
                distances[neighbour.Index] = propagator.Distance + 1
                c_s := CellAndScore{
                    Index: neighbour.Index,
                    Score: result[neighbour.Index],
                    Distance: distances[neighbour.Index],
                }

                if len(queue) == 0 || queue[len(queue) - 1].Score > c_s.Score {     // Safe (correct) to append to end

                    queue = append(queue, c_s)

                } else {

                    // Fast(ish) insertion...

                    index := sort.Search(len(queue), func(i int) bool { return queue[i].Score <= c_s.Score })

                    queue = append(queue, CellAndScore{0, 0, 0})     // Dummy item, gets overwritten
                    copy(queue[index + 1:], queue[index:])
                    queue[index] = c_s
                }
            }
        }
    }

    return result, distances
}

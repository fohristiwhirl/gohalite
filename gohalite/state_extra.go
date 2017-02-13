package gohalite

// What is i?

func (g *Game) NiceNeutralCheck(i, nice_minimum int) bool {
    if g.Owner[i] == 0 && (g.Strength[i] == 0 || g.Production[i] >= nice_minimum) {
        return true
    }
    return false
}

// What does i touch?

func (g *Game) TouchesEnemy(i int) bool {
    for _, neighbour := range g.Neighbours[i] {
        if g.Owner[neighbour.Index] != g.Id && g.Owner[neighbour.Index] != 0 {
            return true
        }
    }
    return false
}

func (g *Game) TouchesFriendly(i int) bool {
    for _, neighbour := range g.Neighbours[i] {
        if g.Owner[neighbour.Index] == g.Id {
            return true
        }
    }
    return false
}

func (g *Game) TouchesNeutral(i int) bool {
    for _, neighbour := range g.Neighbours[i] {
        if g.Owner[neighbour.Index] == 0 {
            return true
        }
    }
    return false
}

func (g *Game) TouchesNiceNeutral(i, nice_minimum int) bool {
    for _, neighbour := range g.Neighbours[i] {
        if g.NiceNeutralCheck(neighbour.Index, nice_minimum) {
            return true
        }
    }
    return false
}

// Lists...

func (g *Game) ListWarZones() []int {
    var result []int
    for i := 0 ; i < g.Size ; i++ {
        if g.Owner[i] == 0 && g.TouchesFriendly(i) && g.TouchesEnemy(i) {
            result = append(result, i)
        }
    }
    return result
}

func (g *Game) ListTouchingNeutrals() []int {
    var result []int
    for i := 0 ; i < g.Size ; i++ {
        if g.Owner[i] == 0 && g.TouchesFriendly(i) {
            result = append(result, i)
        }
    }
    return result
}

func (g *Game) ListTouchingNiceNeutrals(nice_minimum int) []int {
    var result []int
    for i := 0 ; i < g.Size ; i++ {
        if g.TouchesFriendly(i) && g.NiceNeutralCheck(i, nice_minimum) {
            result = append(result, i)
        }
    }
    return result
}

func (g *Game) ListFriendlies() []int {
    var result []int
    for i := 0 ; i < g.Size ; i++ {
        if g.Owner[i] == g.Id {
            result = append(result, i)
        }
    }
    return result
}

func (g *Game) ListFrontierFriendlies() []int {
    var result []int
    for i := 0 ; i < g.Size ; i++ {
        if g.Owner[i] == g.Id && g.TouchesNeutral(i) {
            result = append(result, i)
        }
    }
    return result
}

func (g *Game) ListNiceFrontierFriendlies(nice_minimum int) []int {
    var result []int
    for i := 0 ; i < g.Size ; i++ {
        if g.Owner[i] == g.Id && g.TouchesNiceNeutral(i, nice_minimum) {
            result = append(result, i)
        }
    }
    return result
}

func (g *Game) ListStillFriendlies(min_strength int) []int {
    var result []int
    for i := 0 ; i < g.Size ; i++ {
        if g.Owner[i] == g.Id && g.Moves[i] == STILL && g.Strength[i] >= min_strength {
            result = append(result, i)
        }
    }
    return result
}

// "Maps"... (not in a code sense of the word)

func (g *Game) NiceFrontierDistances(nice_minimum int) []int {

    // Map (sensu lato) of distances to the "nice" frontier.

    result := make([]int, g.Size)

    frontier_friendlies := g.ListNiceFrontierFriendlies(nice_minimum)

    // We treat certain neutrals as voids to be ignored.
    // These are set in the map with values of 99.

    if nice_minimum > 0 {
        for i := 0 ; i < g.Size ; i++ {
            if g.Owner[i] == 0 && g.Production[i] < nice_minimum && g.Strength[i] > 0 {
                result[i] = 99
            }
        }
    }

    for _, i := range frontier_friendlies {
        result[i] = 1
    }

    depth := 2
    last_iteration_hits := frontier_friendlies

    for {
        this_iteration_hits := make([]int, 0, 8)

        for _, i := range last_iteration_hits {
            for _, neighbour := range g.Neighbours[i] {
                if g.Owner[neighbour.Index] == g.Id && result[neighbour.Index] == 0 {
                    result[neighbour.Index] = depth
                    this_iteration_hits = append(this_iteration_hits, neighbour.Index)
                }
            }
        }

        if len(this_iteration_hits) == 0 {
            break
        }

        last_iteration_hits = this_iteration_hits
        depth++
    }

    return result
}

func (g *Game) FrontierDistances() []int {
    return g.NiceFrontierDistances(0)
}

func (g *Game) EnemyDistances() []int {

    // Map (sensu lato) of distances to enemies.

    result := make([]int, g.Size)

    first_hits := make([]int, 0, 8)

    for i := 0 ; i < g.Size ; i++ {
        if g.Owner[i] != g.Id && g.Owner[i] != 0 {
            for _, neighbour := range g.Neighbours[i] {
                if g.Owner[neighbour.Index] == g.Id || g.Owner[neighbour.Index] == 0 {
                    result[neighbour.Index] = 1
                    first_hits = append(first_hits, neighbour.Index)
                }
            }
        }
    }

    last_iteration_hits := first_hits
    depth := 1

    for {
        this_iteration_hits := make([]int, 0, 8)
        depth++

        for _, i := range last_iteration_hits {
            for _, neighbour := range g.Neighbours[i] {
                if g.Owner[neighbour.Index] == g.Id || g.Owner[neighbour.Index] == 0 {
                    if result[neighbour.Index] == 0 {
                        result[neighbour.Index] = depth
                        this_iteration_hits = append(this_iteration_hits, neighbour.Index)
                    }
                }
            }
        }

        if len(this_iteration_hits) == 0 {
            break
        }

        last_iteration_hits = this_iteration_hits
    }

    return result
}

func (g *Game) BestAttractions(attraction_map []int) []int {

    // For any given cell, what's the score of the cell its most attracted to?

    result := make([]int, g.Size)

    for i := 0 ; i < g.Size ; i++ {
        result[i] = -2147483647
    }

    for i := 0 ; i < g.Size ; i++ {

        if g.Owner[i] != g.Id {
            continue
        }

        best_score := attraction_map[i]

        for _, neighbour := range g.Neighbours[i] {
            if attraction_map[neighbour.Index] > best_score {
                best_score = attraction_map[neighbour.Index]
            }
        }

        result[i] = best_score
    }

    return result
}

// Special list of lists...

func (g *Game) ListFriendliesByValue(value_map []int) [][]int {

    // value_map is any slice representing the whole board, giving a value to each index

    max_value := 0
    for i := 0 ; i < g.Size ; i++ {
        if value_map[i] > max_value {
            max_value = value_map[i]
        }
    }

    result := make([][]int, max_value + 1)

    for i := 0 ; i < g.Size ; i++ {
        if g.Owner[i] == g.Id {
            value := value_map[i]
            result[value] = append(result[value], i)
        }
    }

    return result
}

func (g *Game) SetStartLoc() {
    for i := 0 ; i < g.Size ; i++ {
        if g.Owner[i] == g.Id {
            g.StartLoc = i
            break
        }
    }
}

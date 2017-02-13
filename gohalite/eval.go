package gohalite

func (g *Game) Goodness(i int) float32 {
    // Assuming i is neutral, this provides a way to rank desirability.
    // Higher is better. Stolen from DanielVF.
    return float32(g.Production[i]) / (float32(g.Strength[i]) + 0.01)
}

func (g *Game) SecondOrderGoodness(i int) float32 {
    result := float32(g.Production[i]) / (float32(g.Strength[i]) + 0.01)
    for _, neighbour := range g.Neighbours[i] {
        if g.Owner[neighbour.Index] == 0 {
            result += float32(g.Production[neighbour.Index]) / (float32(g.Strength[neighbour.Index]) + 0.01)
        }
    }
    return result
}

func (g *Game) NiceExcludeFraction(nice_min int) float32 {
    excluded := 0
    for i := 0 ; i < g.Size ; i++ {
        if g.Production[i] < nice_min {
            excluded++
        }
    }
    return float32(excluded) / float32(g.Size)
}

func (g *Game) CountPlayers() int {
    set := make(map[int]bool)
    for i := 0 ; i < g.Size ; i++ {
        if g.Owner[i] > 0 {
            set[g.Owner[i]] = true
        }
    }
    return len(set)
}

func (g *Game) CountFriendlyCells() int {
    result := 0
    for i := 0 ; i < g.Size ; i++ {
        if g.Owner[i] == g.Id {
            result++
        }
    }
    return result
}

func (g *Game) CountEnemyCells() int {
    result := 0
    for i := 0 ; i < g.Size ; i++ {
        if g.Owner[i] != g.Id && g.Owner[i] != 0 {
            result++
        }
    }
    return result
}

func (g *Game) TotalStrengths() []int {
    result := make([]int, g.InitialPlayerCount + 1, g.InitialPlayerCount + 1)
    for i := 0 ; i < g.Size ; i++ {
        result[g.Owner[i]] += g.Strength[i]
    }
    return result
}

func (g *Game) StrengthOfPlayer(id int) int {
    result := 0
    for i := 0 ; i < g.Size ; i++ {
        if g.Owner[i] == id {
            result += g.Strength[i]
        }
    }
    return result
}

func (g *Game) MyProduction() int {
    result := 0
    for i := 0 ; i < g.Size ; i++ {
        if g.Owner[i] == g.Id {
            result += g.Production[i]
        }
    }
    return result
}

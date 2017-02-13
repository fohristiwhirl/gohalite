package gohalite

type OpeningSimulator struct {
    G                       *Game
    placements              []int
}

func NewOpeningSimulator(g *Game, true_player_id int) *OpeningSimulator {
    s := new(OpeningSimulator)
    s.G = g.Copy()

    s.placements = make([]int, g.Size)

    for i := 0 ; i < s.G.Size ; i++ {
        if s.G.Owner[i] != true_player_id {
            s.G.Owner[i] = 0
        } else {
            s.G.Owner[i] = 1
        }
    }

    s.G.Id = 1
    s.G.InitialPlayerCount = 1

    return s
}

func (s *OpeningSimulator) Simulate() {

    g := s.G

    for i := 0 ; i < g.Size ; i++ {
        s.placements[i] = 0
    }

    // Set placements from movement...

    for i := 0 ; i < g.Size ; i++ {
        if g.Owner[i] == 0 {
            continue
        }
        target := g.Movement_to_I(i, g.Moves[i])
        s.placements[target] += g.Strength[i]
    }

    // Add production from cells that didn't move...

    for i := 0 ; i < g.Size ; i++ {
        if g.Owner[i] != 0 {
            if g.Moves[i] == STILL {
                s.placements[i] += g.Production[i]
            }
        }
    }

    // Cap at 255...

    for i := 0 ; i < g.Size ; i++ {
        if s.placements[i] > 255 {
            s.placements[i] = 255
        }
    }

    // Internal merging...

    for i := 0 ; i < g.Size ; i++ {
        if g.Owner[i] == 1 {
            g.Strength[i] = s.placements[i]
        }
    }

    // Combat vs neutrals...

    for i := 0 ; i < g.Size ; i++ {
        if g.Owner[i] == 0 && s.placements[i] > 0 {
            if g.Strength[i] < s.placements[i] {
                g.Owner[i] = 1
                g.Strength[i] = s.placements[i] - g.Strength[i]
            } else {
                g.Strength[i] -= s.placements[i]
            }
        }
    }

    g.SetExtraState()   // Includes g.Turn += 1 and resets various slices
    return
}

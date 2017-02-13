package gohalite

import (
    "runtime"
    "strconv"
    "strings"
    "time"
)

func (g *Game) Update() {
    g.ParseMap()
    g.SetExtraState()
}

func (g *Game) SetExtraState() {

    // Every turn, set state as appropriate.

    g.Turn += 1

    for i := 0 ; i < g.Size ; i++ {

        g.Moves[i] = STILL
        g.HasOrders[i] = false

        if g.Owner[i] == g.Id {
            g.Allocation[i] = g.Strength[i]     // How much strength we expect to send to this square. If we have a piece on it already
        } else {                                // then by default we will be "sending" that much strength, just by standing still.
            g.Allocation[i] = 0
        }

        g.Incoming[i] = 0                       // Like Allocation, but incoming strength only
    }
}

func (g *Game) Startup() {

    g.ParseInitialMessages()    // Gets size info, needed for next calls
    g.MakeLookupTable()
    g.MakeSlices()

    g.Turn = -1

    g.ParseProduction()
    g.ParseMap()

    g.InitialPlayerCount = g.CountPlayers()
}

func (g *Game) MakeLookupTable() {
    g.Neighbours = make([][]Neighbour, g.Size, g.Size)
    for i := 0 ; i < g.Size ; i++ {

        g.Neighbours[i] = make([]Neighbour, 4, 4)

        x, y := g.I_to_XY(i)

        // Same ordering as the Cardinals, but offset by 1. Movement_to_I() depends on this.

        g.Neighbours[i][0] = Neighbour{g.XY_to_I(x, y - 1), UP}
        g.Neighbours[i][1] = Neighbour{g.XY_to_I(x + 1, y), RIGHT}
        g.Neighbours[i][2] = Neighbour{g.XY_to_I(x, y + 1), DOWN}
        g.Neighbours[i][3] = Neighbour{g.XY_to_I(x - 1, y), LEFT}
    }
}

func (g *Game) MakeSlices() {
    g.Production = make([]int, g.Size)
    g.Owner = make([]int, g.Size)
    g.Strength = make([]int, g.Size)

    g.Moves = make([]int, g.Size)
    g.HasOrders = make([]bool, g.Size)
    g.Allocation = make([]int, g.Size)
    g.Incoming = make([]int, g.Size)

    g.MovementNotes = make([]string, g.Size)
}

func (g *Game) ParseInitialMessages() {
    scanner.Scan()
    g.Id, _ = strconv.Atoi(scanner.Text())
    g.GameStart = time.Now()                    // After 1st message received

    scanner.Scan()
    width_and_height := strings.Fields(scanner.Text())
    g.Width, _ = strconv.Atoi(width_and_height[0])
    g.Height, _ = strconv.Atoi(width_and_height[1])
    g.Size = g.Width * g.Height
}

func (g *Game) SetMove(index, direction int, note string) {
/*
    if SOME_CONDITION {
        x, y := g.I_to_XY(index)
        g.Log("Turn %d [%d,%d], called by %s", g.Turn, x, y, MyCaller())
    }
*/
    g.HasOrders[index] = true
    g.MovementNotes[index] = note

    old_direction := g.Moves[index]
    if old_direction == direction {
        return
    }

    // Undo the allocation if it was already moving...

    if old_direction != STILL {
        g.Allocation[index] += g.Strength[index]
        g.Allocation[g.Movement_to_I(index, old_direction)] -= g.Strength[index]
        g.Incoming[g.Movement_to_I(index, old_direction)] -= g.Strength[index]
    }

    g.Moves[index] = direction

    if direction != STILL {
        g.Allocation[index] -= g.Strength[index]
        g.Allocation[g.Movement_to_I(index, direction)] += g.Strength[index]
        g.Incoming[g.Movement_to_I(index, direction)] += g.Strength[index]
    }
}

// Function for tracking down weird moves.
// http://stackoverflow.com/questions/35212985/is-it-possible-get-information-about-caller-function-in-golang

func MyCaller() string {

    // We get the callers as uintptrs - but we just need 1...

    fpcs := make([]uintptr, 1)

    // Skip 3 levels to get to the caller of whoever called Caller()...

    n := runtime.Callers(3, fpcs)
    if n == 0 {
        return "n/a"
    }

    // Get the info of the actual function that's in the pointer...

    fun := runtime.FuncForPC(fpcs[0]-1)
    if fun == nil {
        return "n/a"
    }

    return fun.Name()
}

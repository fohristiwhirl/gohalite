package gohalite

func (g *Game) I_to_XY(i int) (int, int) {
    x := i % g.Width
    y := i / g.Width
    return x, y
}

func (g *Game) XY_to_I(x, y int) (int) {

    if x < 0 {
        x += -(x / g.Width) * g.Width + g.Width      // Can make x == g.Width, so must still use % later
    }

    x %= g.Width

    if y < 0 {
        y += -(y / g.Height) * g.Height + g.Height   // Can make y == g.Height, so must still use % later
    }

    y %= g.Height

    return y * g.Width + x
}

func (g *Game) Movement_to_I(src, direction int) int {

    // Depends on the ordering of the Neighbours lookup table matching that of the Cardinals

    if direction > 4 || direction <= 0 {
        return src
    }

    return g.Neighbours[src][direction - 1].Index
}

func (g *Game) Cardinal(src, dst int) int {

    // Return the direction that a piece on src must move to get to dst in 1 move.
    // Assumes they are adjacent. The arguments are indices of the map.

    x, y := g.I_to_XY(src)

    if g.XY_to_I(x - 1, y) == dst {
        return LEFT
    } else if g.XY_to_I(x + 1, y) == dst {
        return RIGHT
    } else if g.XY_to_I(x, y - 1) == dst {
        return UP
    } else if g.XY_to_I(x, y + 1) == dst {
        return DOWN
    }

    return STILL
}

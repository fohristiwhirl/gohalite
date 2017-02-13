package gohalite

import (
    "time"
)

const (
    INITIAL_TIMEOUT = 12 * time.Second
    TIMEOUT = 950 * time.Millisecond
    STILL = 0
    NORTH = 1
    EAST = 2
    SOUTH = 3
    WEST = 4
    UP = NORTH
    RIGHT = EAST
    DOWN = SOUTH
    LEFT = WEST
)

type Game struct {

    // Lookup table of neighbouring indices and directions:
    Neighbours          [][]Neighbour

    // Constant after game started:
    GameStart           time.Time
    Width               int
    Height              int
    Size                int
    Id                  int
    InitialPlayerCount  int
    Logfile             *Logfile

    // Single values that get set each turn:
    TurnStart           time.Time
    Turn                int

    // Slices handled by the main parsers:
    Production          []int
    Owner               []int
    Strength            []int

    // Other slices, updated each turn:
    Moves               []int           // Direction to move this turn?
    HasOrders           []bool          // Has this piece explicitly been given orders (even if STILL)?
    Allocation          []int           // How much strength are we sending to this spot?
    Incoming            []int           // Like allocation, but not including strength that started on the spot

    // Flags and vars used by the AI:
    OpeningFlag         bool
    StartLoc            int
    IsSim               bool
    MovementNotes       []string        // For logging
}

func (g *Game) Copy() *Game {

    // There must be a better way of deep copying a struct?

    result := new(Game)

    result.GameStart = g.GameStart
    result.Width = g.Width
    result.Height = g.Height
    result.Size = g.Size
    result.Id = g.Id
    result.InitialPlayerCount = g.InitialPlayerCount
    result.Logfile = g.Logfile

    result.TurnStart = g.TurnStart
    result.Turn = g.Turn

    result.Neighbours = g.Neighbours    // OK since the original is read-only in practice
    result.MakeSlices()

    copy(result.Production, g.Production)
    copy(result.Owner, g.Owner)
    copy(result.Strength, g.Strength)
    copy(result.Moves, g.Moves)
    copy(result.HasOrders, g.HasOrders)
    copy(result.Allocation, g.Allocation)
    copy(result.Incoming, g.Incoming)

    result.OpeningFlag = g.OpeningFlag
    result.StartLoc = g.StartLoc
    result.IsSim = g.IsSim

    return result
}

type Neighbour struct {
    Index               int
    Dir                 int
}

func Dir_to_str(dir int) string {       // For debugging / logging
    switch dir {
    case STILL:
        return "still"
    case RIGHT:
        return "right"
    case LEFT:
        return "left"
    case UP:
        return "up"
    case DOWN:
        return "down"
    default:
        return "screwball"
    }
}

func Opposite(dir int) int {
    switch dir {
    case STILL:
        return STILL
    case RIGHT:
        return LEFT
    case LEFT:
        return RIGHT
    case UP:
        return DOWN
    case DOWN:
        return UP
    default:
        return STILL
    }
}

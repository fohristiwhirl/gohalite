package gohalite

import (
    "crypto/sha1"
    "fmt"
    "os"
    "strconv"
    "strings"
)

type Logfile struct {
    outfile         *os.File
    outfilename     string
    enabled         bool
    logged_once     map[string]bool
}

func NewLog(outfilename string, enabled bool) *Logfile {
    logged_once := make(map[string]bool)
    return &Logfile{nil, outfilename, enabled, logged_once}
}

func (log *Logfile) Dump(format_string string, args ...interface{}) {

    if log == nil {
        return
    }

    if log.enabled == false {
        return
    }

    if log.outfile == nil {

        var err error

        if _, tmp_err := os.Stat(log.outfilename); tmp_err == nil {
            // File exists
            log.outfile, err = os.OpenFile(log.outfilename, os.O_APPEND|os.O_WRONLY, 0666)
        } else {
            // File needs creating
            log.outfile, err = os.Create(log.outfilename)
        }

        if err != nil {
            log.enabled = false
            return
        }
    }

    fmt.Fprintf(log.outfile, format_string, args...)
    fmt.Fprintf(log.outfile, "\r\n")                    // Because I use Windows...
}

func (g *Game) Log(format_string string, args ...interface{}) {

    if g.IsSim {
        return
    }

    g.Logfile.Dump(format_string, args...)
}

func (g *Game) LogOnce(format_string string, args ...interface{}) bool {

    if g.IsSim {
        return false
    }

    if g.Logfile.logged_once[format_string] == false {
        g.Logfile.logged_once[format_string] = true         // Note that it's format_string that is checked / saved
        g.Logfile.Dump(format_string, args...)
        return true
    }
    return false
}

func (g *Game) LogInSim(format_string string, args ...interface{}) {
    g.Logfile.Dump(format_string, args...)
}

func (g *Game) LogValueMap(value_map []int, translate map[int]string) {

    var add string
    var ok bool

    var all_strings []string

    max_width := 0
    for i := 0 ; i < g.Size ; i++ {
        if translate != nil {
            add, ok = translate[value_map[i]]
            if ok == false {
                add = strconv.Itoa(value_map[i])
            }
        } else {
            add = fmt.Sprintf("%d", value_map[i])
        }
        if len(add) > max_width {
            max_width = len(add)
        }
        all_strings = append(all_strings, add)
    }

    format_string := fmt.Sprintf(" %%%ds", max_width)

    s := ""
    for i := 0 ; i < g.Size ; i++ {
        s += fmt.Sprintf(format_string, all_strings[i])
        if i % g.Width == g.Width - 1 {
            g.Log(s)
            s = ""
        }
    }
    g.Log("")
}

func (g *Game) LogAttractMap(attraction_map []int) {
    g.LogValueMap(attraction_map, map[int]string{-2147483647:"."})
}

func (g *Game) LogAttractDirections(attraction_map []int) {

    direction_map := make([]int, g.Size)

    for y := 0 ; y < g.Height ; y++ {
        for x := 0 ; x < g.Width ; x++ {
            i := g.XY_to_I(x,y)

            if g.Owner[i] != g.Id {
                if g.Owner[i] == 0 {
                    direction_map[i] = -1
                } else {
                    direction_map[i] = -2
                }
                continue
            }

            best := attraction_map[i]
            direction := STILL

            for _, neighbour := range g.Neighbours[i] {
                if attraction_map[neighbour.Index] > best {
                    best = attraction_map[neighbour.Index]
                    direction = neighbour.Dir
                }
            }

            direction_map[i] = direction
        }
    }

    g.LogValueMap(direction_map, map[int]string{-2: "O", -1:".", 0:"?", 1:"^", 2:">", 3:"v", 4:"<"})
}

func (g *Game) LogPredictions(predicted_strengths []int, my_forces_set map[int]bool) {

    s := ""
    for i := 0 ; i < g.Size ; i++ {
        if my_forces_set[i] {
            s += "O "
        } else if predicted_strengths[i] > 0 {
            s += "+ "
        } else {
            s += ". "
        }
        if i % g.Width == g.Width - 1 {
            g.Log(s)
            s = ""
        }
    }
    g.Log("")
}

func (g *Game) LogOwner() {
    g.LogValueMap(g.Owner, nil)
}

func (g *Game) LogStrength() {
    g.LogValueMap(g.Strength, nil)
}

func (g *Game) LogProduction() {
    g.LogValueMap(g.Production, nil)
}

func (g *Game) LogList(list []int) {

    bools := make([]bool, g.Size)

    for _, i := range list {
        bools[i] = true
    }

    s := ""
    for i := 0 ; i < g.Size ; i++ {
        if bools[i] {
            s += "O"
        } else {
            s += "."
        }
        if i % g.Width == g.Width - 1 {
            g.Log(s)
            s = ""
        }
    }
    g.Log("")
}

func (g *Game) LogLists(lists ...[]int) {

    display := make([]int, g.Size)

    for n, list := range lists {
        for _, i := range list {
            if display[i] == 0 {
                display[i] = n + 1      // Because of zeroth index
            } else {
                display[i] = -1
            }
        }
    }

    g.LogValueMap(display, map[int]string{-1:"?", 0:"."})
}

func (g *Game) LogMoves() {
    g.LogValueMap(g.Moves, map[int]string{0:".", 1:"^", 2:">", 3:"v", 4:"<"})
}

func (g *Game) LogOverallocation(cumulative int) int {
    s := ""
    for i := 0 ; i < g.Size ; i++ {
        if g.Allocation[i] > 255 {
            x, y := g.I_to_XY(i)

            report := fmt.Sprintf("[%d,%d]", x, y)
            for _, neighbour := range g.Neighbours[i] {
                if g.Moves[neighbour.Index] == Opposite(neighbour.Dir) {
                    mover_x, mover_y := g.I_to_XY(neighbour.Index)
                    report += fmt.Sprintf("(%s: %d,%d)", g.MovementNotes[neighbour.Index], mover_x, mover_y)
                }
            }

            s += report + "  "
            cumulative += g.Allocation[i] - 255
        }
    }

    if len(s) > 0 {
        g.Log("Turn %d: Over-allocation:  %s-- cumulative: %d", g.Turn, s, cumulative)
    }
    return cumulative
}

func hash_from_string(datastring string) string {
    data := []byte(datastring)
    sum := sha1.Sum(data)
    return fmt.Sprintf("%x", sum)
}

func (g *Game) MovesHash() string {
    var components []string
    for i := 0 ; i < g.Size ; i++ {
        if g.Owner[i] == g.Id && g.Moves[i] != STILL {
            components = append(components, fmt.Sprintf("%d %d", i, g.Moves[i]))
        }
    }
    fullstring := strings.Join(components, " ")
    return hash_from_string(fullstring)
}

func (g *Game) BoardHash() string {
    var components []string

    for i := 0 ; i < g.Size ; i++ {
        components = append(components, fmt.Sprintf("%d %d %d", g.Owner[i], g.Production[i], g.Strength[i]))
    }

    fullstring := strings.Join(components, " ")
    return hash_from_string(fullstring)
}

func (g *Game) LogMovesHash() {
    g.Log("Moves SHA1: %s", g.MovesHash())
}

func (g *Game) LogBoardHash() {
    g.Log("Board SHA1: %s", g.BoardHash())
}

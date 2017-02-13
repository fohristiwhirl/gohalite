package gohalite

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
    "time"
)

var scanner *bufio.Scanner = bufio.NewScanner(os.Stdin)

func (g *Game) ParseProduction() {
    scanner.Scan()
    production_fields := strings.Fields(scanner.Text())
    for i := 0; i < len(production_fields); i++ {
        g.Production[i], _ = strconv.Atoi(production_fields[i])
    }
}

func (g *Game) ParseMap() {

    // See https://halite.io/advanced_writing_sp.php

    scanner.Scan()
    g.TurnStart = time.Now()        // Do this immediately after .Scan() succeeds

    map_fields := strings.Fields(scanner.Text())

    field_index := 0

    game_index := 0
    for game_index < g.Size {

        num, _ := strconv.Atoi(map_fields[field_index])
        owner, _ := strconv.Atoi(map_fields[field_index + 1])
        field_index += 2

        for n := 0 ; n < num ; n++ {
            g.Owner[game_index] = owner
            game_index++
        }
    }

    game_index = 0
    for game_index < g.Size {
        g.Strength[game_index], _ = strconv.Atoi(map_fields[field_index])
        field_index++
        game_index++
    }
}

func (g *Game) SendMoves() {

    for i := 0 ; i < g.Size ; i++ {
        if g.Owner[i] == g.Id && g.Moves[i] != STILL {
            x, y := g.I_to_XY(i)
            fmt.Printf("%d %d %d ", x, y, g.Moves[i])
        }
    }
    fmt.Printf("\n")
}

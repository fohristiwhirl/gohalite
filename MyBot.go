package main

/*  GoHalite - by Github user Fohristiwhirl
    My entry in the Halite programming competition

    OPENING SIMULATIONS:

        - During the opening 15 seconds, we generate potential openings and then simulate future expansion for each.
        - The best opening plan (i.e. a sequence of 3 or 4 squares to capture) is used, unless it is worse than no plan at all.
        - See functions AI_Startup(), SimulateOpenings() and its alternative CheapOpeningTester() used on big boards.

    BREADTH-FIRST-SEARCH: (thanks to DanielVF for helpful discussions)

        - Originally, a breadth-first search was run to pull central pieces to the borders.
        - Later I learned that this was unnecessary, so the code is now just used in the very early game.
        - It is used to pull the strength needed for the first few (planned) captures, as early as possible.
        - See function OneFinger() here, but also the file bfs.go.

    ATTRACTION OF PIECES TO BORDERS: (again thanks to DanielVF for helpful discussions)

        - Borders (neutrals touching us) are scored based on production and strength.
        - These scores propagate through our friendly territory, creating "gradients" or "slides" for pieces to travel down.
        - At each stage a penalty is applied based on production, so we try to avoid travelling over high production places.
        - Pieces that are moving simply try to go to the highest scoring neighbour (unless that's suicidal).
        - See the file attraction2.go and its method AttractionMap_v2().
        - See also function Attract() in this file.

    WAR:

        - The war code itself is rather stupid - see War().
        - However, we like to ensure a checkerboard pattern of pieces arriving at the frontline (empirically quite effective).
        - This is achieved by penalising the scores in the attraction map of half of the squares that are near the frontline.
        - See function checker_penalty().
*/

import (
    "fmt"
    "math"
    "math/rand"
    "sort"
    "time"

    hal "./gohalite"
)

const NAME = "v52"

const (
    LOGGING_ENABLED = true
    DETERMINIST = true
    FINGER_LENGTH = 4
    INTERNAL_MULTIPLIER = 6
    NORMAL_NICE_MIN = 1
    MAX_OPENING_COMBO = 4
)

var TotalSimPos int
var BestNiceMin int
var NiceMin int
var OpeningEvalDepth int

func main() {

    g := new(hal.Game)
    g.Logfile = hal.NewLog("Log" + "_" + NAME + ".log", LOGGING_ENABLED)
    g.Startup()

    var longest_ponder time.Duration
    var longest_ponder_turn int
    var total_overallocation int

    g.Log("----------------------------------------------------------------")

    best_opening := AI_Startup(g)

    fmt.Printf("%s\n", NAME)                        // Tell the engine we're ready

    rand.Seed(1)                                    // Match the seed used by the simulations (do this last, before real loop)

    for {

        // Main loop...

        g.Update()
        MakeMoves(g, best_opening)
        g.SendMoves()

        // The rest is just logging...

        total_overallocation = g.LogOverallocation(total_overallocation)

        if time.Since(g.TurnStart) > longest_ponder {
            longest_ponder = time.Since(g.TurnStart)
            longest_ponder_turn = g.Turn
        }

        if g.OpeningFlag == false {
            g.LogOnce("OpeningFlag became false on turn %d", g.Turn)
        }

        if time.Since(g.TurnStart) >= hal.TIMEOUT {
            g.Log("Turn %d: Apparently timed out", g.Turn)
        }

        MaybeLogEnd(g, longest_ponder, longest_ponder_turn)
    }
}

// -------------------------------------------------------------------------------------------------------------

func LogStartInfo(g *hal.Game) {
    x,y := g.I_to_XY(g.StartLoc)
    g.Log("%d players ... [%d x %d] ... [%d,%d] ... %v", g.InitialPlayerCount, g.Width, g.Height, x, y, g.GameStart.Format("2006-01-02  15:04:05"))
    g.Log("Determinism: %v ... BestNiceMin == %d, excludes %.2f%%", DETERMINIST, BestNiceMin, g.NiceExcludeFraction(BestNiceMin) * 100)
}

func MaybeLogEnd(g *hal.Game, longest_ponder time.Duration, longest_ponder_turn int) bool {

    if g.IsSim {
        return false      // Waste of time doing anything in this case
    }

    didlog := false

    friendlies := g.CountFriendlyCells()
    enemies := g.CountEnemyCells()

    if enemies > 0 {
        if float32(friendlies) / float32(enemies) < 0.05 {
            didlog = g.LogOnce("Turn %d: imminent DEFEAT", g.Turn)
        } else if float32(friendlies) / float32(enemies) > 20 {
            didlog = g.LogOnce("Turn %d: imminent VICTORY", g.Turn)
        }
    }

    if didlog {
        g.Log("Longest ponder by turn %d: %v (turn %d)", g.Turn, longest_ponder, longest_ponder_turn)
        g.LogBoardHash()
    }

    return didlog
}

// -------------------------------------------------------------------------------------------------------------

func AI_Startup(g *hal.Game) []int {

    if g.CountFriendlyCells() == 1 {
        g.OpeningFlag = true
    } else {
        g.OpeningFlag = false
    }

    g.SetStartLoc()     // But will be wrong value if loaded into midgame.

    BestNiceMin = NORMAL_NICE_MIN
    LogStartInfo(g)

    if g.OpeningFlag {
        if g.Width * g.Height >= 30 * 30 {
            OpeningEvalDepth = 145 - 5 * int(math.Sqrt(float64(g.Width * g.Height))) / 3            // Can get away with a longer depth since we're not as wide
            return CheapOpeningTester(g)
        } else {
            OpeningEvalDepth = 80 - 5 * int(math.Sqrt(float64(g.Width * g.Height))) / 3             // Also the time we have before war is shorter in small maps
            return SimulateOpenings(g)
        }
    } else {
        g.Log("Apparently loaded into midgame.")
        return nil
    }
}

type BestCombo struct {
    score int
    combo []int     // The winning combo contains our first n cells, in capture order (including our start cell)
}

func SimulateOpenings(g *hal.Game) []int {

    g.Log("Using SimulateOpenings()")

    nil_result := EvaluateCombo(g, nil)

    var combo []int
    combo = append(combo, g.StartLoc)

    var bc BestCombo
    abortflag := Recursor(g, combo, &bc)

    if abortflag {
        g.Log("Search was aborted due to time.")
    }

    g.Log("%d total simulated positions; depth: %d, time taken: %v", TotalSimPos, OpeningEvalDepth, time.Since(g.GameStart))

    if bc.score > nil_result {
        g.Log("Using combo: %v, score: %d (nil score: %d)", bc.combo, bc.score, nil_result)
        return bc.combo
    } else {
        g.Log("NOT using combo: %v, score: %d (nil score: %d)", bc.combo, bc.score, nil_result)
        return nil
    }

    // Note that, because the OpeningFlag affects what functions are called, the result of using
    // nil can be better than the result of using the combo that is equivalent to it, just because
    // of differences in the state of the RNG. This is fine.
}

func Recursor(g *hal.Game, combo []int, bc *BestCombo) bool {

    if time.Since(g.GameStart) > hal.INITIAL_TIMEOUT {
        return true
    }

    // Given a partly formed combo, recurse through all possible continuations
    // except when max depth is reached, in which case update the best result
    // if we are the best result...

    if len(combo) >= MAX_OPENING_COMBO {
        score := EvaluateCombo(g, combo)
        if score > bc.score {
            bc.score = score
            bc.combo = make([]int, len(combo))
            copy(bc.combo, combo)
        }
        if score == -1 {        // Emergency timeout initiated in EvaluateCombo()
            return true
        }
        return false
    }

    // For all items in the combo already...

    for _, i := range(combo) {

        // For all 4 neighbours...

        NEIGHBOUR_LOOP:
        for _, neighbour := range g.Neighbours[i] {

            neigh_i := neighbour.Index

            for _, z := range combo {
                if z == neigh_i {
                    continue NEIGHBOUR_LOOP
                }
            }

            // So neigh_i is a cell we don't already have...

            extended_combo := append(combo, neigh_i)

            abortflag := Recursor(g, extended_combo, bc)
            if abortflag {
                return true
            }
        }
    }

    return false    // no time problem
}

func CheapOpeningTester(g *hal.Game) []int {

    g.Log("Using CheapOpeningTester()")

    U, R, D, L := hal.UP, hal.RIGHT, hal.DOWN, hal.LEFT

    combos_to_try := [][]int{
        []int{U,U,U}, []int{R,R,R}, []int{D,D,D}, []int{L,L,L},         // Straight

        []int{U,U,R}, []int{U,U,L}, []int{R,R,U}, []int{R,R,D},         // Turn at end
        []int{L,L,U}, []int{L,L,D}, []int{D,D,L}, []int{D,D,R},

        []int{U,R,R}, []int{U,L,L}, []int{R,U,U}, []int{R,D,D},         // Turn at start
        []int{L,U,U}, []int{L,D,D}, []int{D,L,L}, []int{D,R,R},

        []int{U,R,U}, []int{U,L,U}, []int{R,U,R}, []int{R,D,R},         // Turn then turn
        []int{L,U,L}, []int{L,D,L}, []int{D,L,D}, []int{D,R,D},

        []int{U,U,U,U}, []int{R,R,R,R}, []int{D,D,D,D}, []int{L,L,L,L},

        []int{U,U,U,R}, []int{U,U,U,L}, []int{R,R,R,U}, []int{R,R,R,D},
        []int{L,L,L,U}, []int{L,L,L,D}, []int{D,D,D,L}, []int{D,D,D,R},

        []int{U,R,R,R}, []int{U,L,L,L}, []int{R,U,U,U}, []int{R,D,D,D},
        []int{L,U,U,U}, []int{L,D,D,D}, []int{D,L,L,L}, []int{D,R,R,R},
    }

    nil_result := EvaluateCombo(g, nil)

    var bc BestCombo

    for _, arr := range combos_to_try {

        var combo []int
        combo = append(combo, g.StartLoc)

        current_pos := g.StartLoc
        for _, direction := range arr {
            next_pos := g.Movement_to_I(current_pos, direction)
            combo = append(combo, next_pos)
            current_pos = next_pos
        }

        score := EvaluateCombo(g, combo)
        if score > bc.score {
            bc.score = score
            bc.combo = make([]int, len(combo))
            copy(bc.combo, combo)
        }
        if score == -1 {        // Emergency timeout initiated in EvaluateCombo()
            g.Log("Search was aborted due to time.")
            break
        }
    }

    g.Log("%d total simulated positions; depth: %d, time taken: %v", TotalSimPos, OpeningEvalDepth, time.Since(g.GameStart))

    if bc.score > nil_result {
        g.Log("Using combo: %v, score: %d (nil score: %d)", bc.combo, bc.score, nil_result)
        return bc.combo
    } else {
        g.Log("NOT using combo: %v, score: %d (nil score: %d)", bc.combo, bc.score, nil_result)
        return nil
    }
}

func EvaluateCombo(realgame *hal.Game, combo []int) int {

    rand.Seed(1)

    s := hal.NewOpeningSimulator(realgame, realgame.Id)
    s.G.OpeningFlag = true
    s.G.IsSim = true
    s.G.Turn = 0

    for n := 0 ; n < OpeningEvalDepth ; n++ {

        if time.Since(s.G.GameStart) > hal.INITIAL_TIMEOUT {          // Emergency timeout.
            return -1
        }

        MakeMoves(s.G, combo)

        if time.Since(s.G.GameStart) > hal.INITIAL_TIMEOUT {          // Emergency timeout.
            return -1
        }

        s.Simulate()
        TotalSimPos++
    }

    return s.G.MyProduction()
}

func OpeningFingers(g *hal.Game, combo []int) error {

    // FIXME: gotta deal with the case where we've made the final valid pull

    tried_a_finger := false
    war := false

    // Iterate through the combo's valid targets and break if war, or when a pull fails to happen

    for _, i := range(combo) {
        if g.Owner[i] != g.Id && g.TouchesFriendly(i) {

            if g.TouchesEnemy(i) {
                war = true
                break
            }

            success := OneFinger(g, i)
            tried_a_finger = true

            if success == false {
                break
            }
        }
    }

    if war {
        return fmt.Errorf("war detected by OpeningFingers")
    }
    if tried_a_finger == false {
        return fmt.Errorf("no valid target found by OpeningFingers")        // "valid target" means a neutral that could be attacked (maybe later)
    }
    return nil
}

// -------------------------------------------------------------------------------------------------------------

func FixNiceMin(g *hal.Game) {

    // If we're trapped because of NiceMin being low, that fact will show up as
    // a zero length list of nice touching neutrals. Note that, once war begins,
    // there are always such neutrals. (Since strength 0 neutrals are always
    // present during war, and strength 0 counts as nice.)

    NiceMin = BestNiceMin

    for {
        if NiceMin == 0 {
            return
        }

        touch_list := g.ListTouchingNiceNeutrals(NiceMin)
        if len(touch_list) == 0 {
            NiceMin -= 1
            g.LogOnce("Turn %d: NiceMin reduced", g.Turn)
        } else {
            return
        }
    }
}

// -------------------------------------------------------------------------------------------------------------

func MakeMoves(g *hal.Game, opening_combo []int) {

    FixNiceMin(g)

    if len(opening_combo) == 0 {
        g.OpeningFlag = false
    }

    var war_zones []int

    if g.IsSim == false {
        war_zones = g.ListWarZones()

        if len(war_zones) > 0 {
            g.OpeningFlag = false
            War(g, war_zones)
        }
    }

    if g.OpeningFlag {
        err := OpeningFingers(g, opening_combo)
        if err != nil {
            g.OpeningFlag = false
        }
    }

    if g.OpeningFlag == false {

        attraction_map, target_distances := g.AttractionMap_v2(NiceMin)

        if len(war_zones) > 0 {
            checker_penalty(g, attraction_map)
        }

        Attract(g, attraction_map, target_distances)
        ForcedAttacks(g)
    }
}

// -------------------------------------------------------------------------------------------------------------

func War(g *hal.Game, war_zones []int) {

    // This is the worst part of the code I'm sure, and is poorly thought out.
    // In particular, its choice of moves only makes sense if the enemy is not
    // moving, which of course is almost never true.

    my_forces_set := make(map[int]bool)         // Using a set to avoid duplicate entries

    for _, i := range war_zones {
        for _, neighbour := range g.Neighbours[i] {
            neigh_i := neighbour.Index
            if g.Owner[neigh_i] == g.Id && g.Strength[neigh_i] > 0 {
                my_forces_set[neigh_i] = true
            }
        }
    }

    my_forces := key_list(my_forces_set, DETERMINIST)

    tmp := hal.SortStruct{g, my_forces}
    sort.Sort(sort.Reverse(hal.ByStrength(tmp)))

    for _, i := range my_forces {

        best_score := 0
        best_move := hal.STILL

        for _, neighbour := range g.Neighbours[i] {     // Ranging over the various moves we can make
            neigh_i := neighbour.Index

            if g.Owner[neigh_i] != 0 {
                continue
            }

            score := g.Strength[neigh_i] * -1           // Reduce our score by the damage we take from the neutral

            for _, second_order_neighbour := range g.Neighbours[neigh_i] {
                second_order_i := second_order_neighbour.Index
                if g.Owner[second_order_i] != g.Id && g.Owner[second_order_i] != 0 {
                    score += g.Strength[second_order_i] + 1         // + 1 so we attack even strength 0 enemies
                }
            }

            if score > best_score {
                if g.Allocation[neigh_i] + g.Strength[i] <= 255 {   // Suicide prevention
                    best_score = score
                    best_move = g.Cardinal(i, neigh_i)
                }
            }
        }

        g.SetMove(i, best_move, "War")
    }
}

// -------------------------------------------------------------------------------------------------------------

func OneFinger(g *hal.Game, puller_index int) bool {

    pull_happened := false

    possibles := hal.BFS(g, puller_index, FINGER_LENGTH, false, DETERMINIST)

    sort.Sort(hal.BFSO_ByDepth(possibles))

    need := g.Strength[puller_index] + 1

    available := 0
    cell_count := 0
    depth := 1

    for n, poss := range possibles {

        if poss.Depth > depth {             // Relies on the sort, above

            if available >= need {          // Previous depth was enough
                break
            }

            // Available strength increases due to one extra turn of production

            diff := poss.Depth - depth

            for p := 0 ; p < n ; p++ {      // "p < n" so we hit every cell PRIOR to this one
                available += possibles[p].Production * diff
            }
/*
            if available >= need {          // This has interesting ramifications. Seems OK without.
                return false
            }
*/
            depth = poss.Depth
        }

        available += poss.Strength
        cell_count++
    }

    if available >= need {

        all_pulls := make([]hal.BFS_Object, cell_count, cell_count)
        copy(all_pulls, possibles[0:cell_count])

        // Lame knapsacking...

        sort.Sort(sort.Reverse(hal.BFSO_ByStrength(all_pulls))) // Reverse puts the weak at the end, which we need...

        for p := cell_count - 1 ; p >= 0 ; p-- {                // ...because we must go backwards here.

            pull := all_pulls[p]

            production_turns := depth - pull.Depth
            production := production_turns * pull.Production
            c := production + pull.Strength

            if available - c >= need {
                 all_pulls = append(all_pulls[:p], all_pulls[p + 1:]...)
                 available -= c
            }
        }

        // Only issue orders if something is actually moving...

        found_outer_mover := false

        for _, pull := range all_pulls {
            if pull.Depth == depth {
                found_outer_mover = true
                break
            }
        }

        if found_outer_mover {
            pull_happened = true
            for _, pull := range all_pulls {
                if pull.Depth == depth {
                    g.SetMove(pull.Index, pull.Direction, "OneFinger")
                } else {
                    g.SetMove(pull.Index, hal.STILL, "OneFinger")
                }
            }
        }
    }

    return pull_happened
}

// -------------------------------------------------------------------------------------------------------------

func Attract(g *hal.Game, attraction_map []int, target_distances []int) {

    cells_by_dist := g.ListFriendliesByValue(target_distances)

    for dist := 1 ; dist < len(cells_by_dist) ; dist++ {        // Make a pass over the internal pieces, from rim to hub...

        tmp := hal.SortStruct{g, cells_by_dist[dist]}
        sort.Sort(sort.Reverse(hal.ByStrength(tmp)))

        for _, i := range cells_by_dist[dist] {

            if g.HasOrders[i] {             // This cell already got used by some other function...
                continue
            }

            if g.Strength[i] < g.Production[i] * INTERNAL_MULTIPLIER {
                continue
            }

            best := attraction_map[i]
            preferred_dir := hal.STILL
            best_index := i

            for _, neighbour := range g.Neighbours[i] {
                neigh_i := neighbour.Index

                projected_allocation := g.Allocation[neigh_i] + g.Strength[i]
                projected_incoming := g.Incoming[neigh_i] + g.Strength[i]

                if projected_allocation <= 255 || g.TouchesNeutral(neigh_i) && g.Moves[neigh_i] == hal.STILL && projected_incoming <= 255 {

                    if attraction_map[neigh_i] > best {

                        best = attraction_map[neigh_i]
                        preferred_dir = neighbour.Dir
                        best_index = neigh_i

                    } else if attraction_map[neigh_i] == best && preferred_dir != hal.STILL {

                        // Break ties by combining stuff where possible... (new in v39)
                        if g.Allocation[neigh_i] > g.Allocation[best_index] {
                            best = attraction_map[neigh_i]
                            preferred_dir = neighbour.Dir
                            best_index = neigh_i
                        }
                    }
                }
            }

            if preferred_dir != hal.STILL {
                if g.Owner[best_index] != 0 || g.Strength[best_index] < g.Strength[i] || (g.Strength[best_index] == 255 && g.Strength[i] == 255) {
                    g.SetMove(i, preferred_dir, "Attract")
                }
            }
        }
    }

    // Now make obvious captures that we missed because the capturer was under the INTERNAL_BUILDUP threshold...

    if len(cells_by_dist) == 1 {    // FIXME: is this still adequate to deal with the problem of whole-map-capture-in-sim ?
        return
    }

    for _, i := range cells_by_dist[1] {

        if g.HasOrders[i] == false {

            best := attraction_map[i]
            preferred_dir := hal.STILL
            best_index := i

            for _, neighbour := range g.Neighbours[i] {
                if attraction_map[neighbour.Index] > best {
                    best = attraction_map[neighbour.Index]
                    preferred_dir = neighbour.Dir
                    best_index = neighbour.Index
                }
            }

            if g.Owner[best_index] == 0 {
                if g.Strength[best_index] < g.Strength[i] {
                    projected_allocation := g.Allocation[best_index] + g.Strength[i]
                    if projected_allocation <= 255 {
                        g.SetMove(i, preferred_dir, "Attract")
                    }
                }
            }
        }
    }
}

// -------------------------------------------------------------------------------------------------------------

func ForcedAttacks(g *hal.Game) {

    // Make a pass over the frontier pieces.
    // If any are both STILL, and about to be suicided, force an attack.

    frontier_cells := g.ListFrontierFriendlies()

    for _, i := range frontier_cells {

        if g.Moves[i] != hal.STILL {
            continue
        }

        if g.Allocation[i] > 255 {

            targets := make([]hal.Neighbour, 0, 4)
            retreats := make([]hal.Neighbour, 0, 4)

            for _, neighbour := range g.Neighbours[i] {
                if g.Owner[neighbour.Index] == 0 {
                    targets = append(targets, neighbour)
                } else {
                    retreats = append(retreats, neighbour)
                }
            }

            forced_dir := -1

            if len(targets) > 0 {

                var tar_indices []int
                for _, target := range targets {
                    tar_indices = append(tar_indices, target.Index)
                }

                tmp := hal.SortStruct{g, tar_indices}
                sort.Sort(sort.Reverse(hal.ByGoodness(tmp)))

                for _, tar_i := range tar_indices {
                    if g.Allocation[tar_i] + g.Strength[i] <= 255 {
                        forced_dir = g.Cardinal(i, tar_i)
                        break
                    }
                }
            }

            if forced_dir == -1 {

                if len(retreats) > 0 {

                    var ret_indices []int
                    for _, retreat := range retreats {
                        ret_indices = append(ret_indices, retreat.Index)
                    }

                    tmp := hal.SortStruct{g, ret_indices}
                    sort.Sort(hal.ByAllocation(tmp))
                    forced_dir = g.Cardinal(i, ret_indices[0])
                }
            }

            if forced_dir != -1 {
                g.SetMove(i, forced_dir, "ForcedAttacks")
                // x, y := g.I_to_XY(i)
                // g.Log("Turn %d: unit %d [%d,%d] went kamikaze %s", g.Turn, i, x, y, hal.Dir_to_str(forced_dir))
            }
        }
    }
}

// -------------------------------------------------------------------------------------------------------------
// Utilities.

func shuffle(src []int) {
    dest := make([]int, len(src))
    perm := rand.Perm(len(src))
    for i, v := range perm {
        dest[v] = src[i]
    }
    copy(src, dest)
}

func key_list(m map[int]bool, determinist bool) []int {

    // Assuming the map is a set of game indices (locations) this will
    // return the set in a pseudorandom but deterministic order.

    var list_of_map_keys []int
    for k, _ := range m {
        if m[k] {
            list_of_map_keys = append(list_of_map_keys, k)
        }
    }
    if determinist {
        sort.Ints(list_of_map_keys)
        shuffle(list_of_map_keys)
    }
    return list_of_map_keys
}

func checker_penalty(g *hal.Game, attraction_map []int) {

    war_distances := g.NiceFrontierDistances(999)   // Using a NiceMin above 16 means only war zones are considered.

    left := -(g.Width / 2)
    top := -(g.Height / 2)
    right := g.Width - (g.Width / 2)
    bottom := g.Height - (g.Height / 2)

    startx, starty := g.I_to_XY(g.StartLoc)

    for xo := left ; xo < right ; xo++ {
        x := startx + xo
        for yo := top ; yo < bottom ; yo++ {
            y := starty + yo
            i := g.XY_to_I(x, y)
            if war_distances[i] < 4 {
                if g.Turn % 2 == 0 {
                    if (xo % 2 == 0 && yo % 2 == 0 || xo % 2 != 0 && yo % 2 != 0) {
                        attraction_map[i] -= 100
                    }
                } else {
                    if (xo % 2 != 0 && yo % 2 == 0 || xo % 2 == 0 && yo % 2 != 0) {
                        attraction_map[i] -= 100
                    }
                }
            }
        }
    }
}

func minimum(a, b int) int {
    if a <= b {return a} else {return b}
}

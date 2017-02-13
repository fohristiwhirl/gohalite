package gohalite

import (
    "math/rand"
    "sort"
)

type BFS_Object struct {
    Index       int
    Strength    int
    Production  int
    Depth       int
    Direction   int
}

type BFSO_ByDepth       []BFS_Object
type BFSO_ByStrength    []BFS_Object


func (s BFSO_ByDepth) Len() int {
    return len(s)
}
func (s BFSO_ByDepth) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}
func (s BFSO_ByDepth) Less(i, j int) bool {
    return s[i].Depth < s[j].Depth
}


func (s BFSO_ByStrength) Len() int {
    return len(s)
}
func (s BFSO_ByStrength) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}
func (s BFSO_ByStrength) Less(i, j int) bool {
    return s[i].Strength < s[j].Strength
}

func BFS(g *Game, puller int, max_depth int, pass_through_zero_strength bool, determinist bool) []BFS_Object {

    // Simply return all reachable friendlies near the puller.
    // Reachable means "going only through friendlies with no orders and some strength".
    // The caller can prune the list as needed.

    all_pulls := make(map[int]BFS_Object)
    all_pulls[puller] = BFS_Object{Index: puller, Strength: g.Strength[puller], Production: g.Production[puller], Depth: 0, Direction: STILL}

    var pullers_next_depth []int
    pullers_next_depth = append(pullers_next_depth, puller)

    depth := 0

    for {
        depth += 1
        if depth > max_depth {
            break
        }

        pullers_this_depth := make([]int, len(pullers_next_depth), len(pullers_next_depth))
        copy(pullers_this_depth, pullers_next_depth)
        pullers_next_depth = nil

        for _, i := range pullers_this_depth {
            for _, neighbour := range g.Neighbours[i] {
                neigh_i := neighbour.Index
                if g.Owner[neigh_i] == g.Id && g.HasOrders[neigh_i] == false && (g.Strength[neigh_i] > 0 || pass_through_zero_strength) {
                    _, ok := all_pulls[neigh_i]
                    if ok == false {
                        dir := g.Cardinal(neigh_i, i)
                        all_pulls[neigh_i] = BFS_Object{Index: neigh_i, Strength: g.Strength[neigh_i], Production: g.Production[neigh_i], Depth: depth, Direction: dir}
                        pullers_next_depth = append(pullers_next_depth, neigh_i)
                    }
                }
            }
        }
    }

    delete(all_pulls, puller)       // Delete the root

    result := make([]BFS_Object, 0, len(all_pulls))

    keys := key_list_bfso(all_pulls, determinist)

    for _, k := range keys {
        result = append(result, all_pulls[k])
    }

    return result
}


func key_list_bfso(m map[int]BFS_Object, determinist bool) []int {

    var list_of_map_keys []int
    for k, _ := range m {
        list_of_map_keys = append(list_of_map_keys, k)
    }
    if determinist {
        sort.Ints(list_of_map_keys)
        shuffle(list_of_map_keys)
    }
    return list_of_map_keys
}


func shuffle(src []int) {
    dest := make([]int, len(src))
    perm := rand.Perm(len(src))
    for i, v := range perm {
        dest[v] = src[i]
    }
    copy(src, dest)
}

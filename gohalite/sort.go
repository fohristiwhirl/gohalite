package gohalite

// Sort-related stuff. Note that in these things, i and j are indexes in the list, not indexes in the game.
// Also note some BFS-related sort stuff is in bfs.go

type SortStruct struct {
    G           *Game
    Contents    []int               // Because slices are really pointers, the original gets altered by sort
}

type ByGoodness SortStruct
type ByStrength SortStruct
type ByAllocation SortStruct


func (s ByGoodness) Len() int {
    return len(s.Contents)
}
func (s ByGoodness) Swap(i, j int) {
    s.Contents[i], s.Contents[j] = s.Contents[j], s.Contents[i]
}
func (s ByGoodness) Less(i, j int) bool {
    return s.G.Goodness(s.Contents[i]) < s.G.Goodness(s.Contents[j])
}


func (s ByStrength) Len() int {
    return len(s.Contents)
}
func (s ByStrength) Swap(i, j int) {
    s.Contents[i], s.Contents[j] = s.Contents[j], s.Contents[i]
}
func (s ByStrength) Less(i, j int) bool {
    return s.G.Strength[s.Contents[i]] < s.G.Strength[s.Contents[j]]
}


func (s ByAllocation) Len() int {
    return len(s.Contents)
}
func (s ByAllocation) Swap(i, j int) {
    s.Contents[i], s.Contents[j] = s.Contents[j], s.Contents[i]
}
func (s ByAllocation) Less(i, j int) bool {
    return s.G.Allocation[s.Contents[i]] < s.G.Allocation[s.Contents[j]]
}

// ----------------------------------------

type ValueMapSortStruct struct {
    G           *Game
    Contents    []int               // Because slices are really pointers, the original gets altered by sort
    ValueMap    []int
}

func (s ValueMapSortStruct) Len() int {
    return len(s.Contents)
}
func (s ValueMapSortStruct) Swap(i, j int) {
    s.Contents[i], s.Contents[j] = s.Contents[j], s.Contents[i]
}
func (s ValueMapSortStruct) Less(i, j int) bool {

    index_one := s.Contents[i]
    index_two := s.Contents[j]

    if s.ValueMap[index_one] < s.ValueMap[index_two] {
        return true
    } else if s.ValueMap[index_one] > s.ValueMap[index_two] {
        return false
    } else {
        return index_one < index_two
    }
}

// ----------------------------------------

type StrengthAndValueSortStruct struct {
    G           *Game
    Contents    []int               // Because slices are really pointers, the original gets altered by sort
    ValueMap    []int
}

type ByStrengthAndValue StrengthAndValueSortStruct

func (s ByStrengthAndValue) Len() int {
    return len(s.Contents)
}
func (s ByStrengthAndValue) Swap(i, j int) {
    s.Contents[i], s.Contents[j] = s.Contents[j], s.Contents[i]
}
func (s ByStrengthAndValue) Less(i, j int) bool {

    index_one := s.Contents[i]
    index_two := s.Contents[j]

    if s.G.Strength[index_one] < s.G.Strength[index_two] {
        return true
    } else if s.G.Strength[index_one] > s.G.Strength[index_two] {
        return false
    }

    // Strengths equal, break by value map...

    if s.ValueMap[index_one] < s.ValueMap[index_two] {
        return true
    } else if s.ValueMap[index_one] > s.ValueMap[index_two] {
        return false
    }

    // Everything equal, so break ties in a deterministic way...

    if index_one < index_two {
        return true
    } else {
        return false
    }
}

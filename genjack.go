package main

import (
	"flag"
	"fmt"
	"math/rand"
	"runtime"
	"sort"
	"time"
)

type Actor struct {
	actions       [10][42]Action
	ratio         int
	lifetimeRatio int
}

const ACTIONS = 2
const ROUTINES = 8

type Action int

const (
	HOLD Action = 0
	HIT  Action = 1
)

func (actor *Actor) initialize() {
	for ix := range actor.actions {
		for ij := range actor.actions[ix] {
			actor.actions[ix][ij] = Action(rand.Intn(ACTIONS))
		}
	}
	actor.lifetimeRatio += actor.ratio
	actor.ratio = 0
}
func (actor *Actor) mutate(mut float32) {
	for ix := range actor.actions {
		for ij := range actor.actions[ix] {
			if rand.Float32() < mut {
				actor.actions[ix][ij] = Action(rand.Intn(ACTIONS))
			}
		}
	}
	actor.lifetimeRatio += actor.ratio
	actor.ratio = 0
}

func main() {
	runtime.GOMAXPROCS(8)
	rand.Seed(time.Now().UnixNano())
	fmt.Printf("GenJack v0.2 - %v CPU cores detected.\n", runtime.NumCPU())
	var actorCount = flag.Int("a", 2000, "Number of actors per generation")
	var genCount = flag.Int("n", 10000, "Number of generation")
	var gamesPerGen = flag.Int("g", 200, "Number of games per generation")
	flag.Parse()
	fmt.Printf("actors: %v, generations: %v, games: %v\n", *actorCount, *genCount, *gamesPerGen)
	t0 := time.Now()
	runDNA(*actorCount, *genCount, *gamesPerGen)
	t1 := time.Now()
	fmt.Printf("The call took %v to run.\n", t1.Sub(t0))
}

func runDNA(actorCount int, genCount int, gamesPerGen int) {
	actors := make([]Actor, actorCount)
	for ix := range actors {
		actors[ix].initialize()
	}

	for gen := 0; gen < genCount; gen++ {
		actorCount := len(actors)
		tops := int(actorCount / 5)
		for ix := 0; ix < tops; ix++ {
			actors[ix].mutate(0.01)
		}
		for ix := tops; ix < actorCount; ix++ {
			actors[ix].mutate(0.05)
		}
		dispatchHands(&actors, gamesPerGen)
		sort.Sort(ByRatio(actors))
	}
	fmt.Println("Final results:")
	bestActor := &actors[0]
	lifetimeActor := &actors[0]
	for ix := range actors {
		if actors[ix].ratio > bestActor.ratio {
			bestActor = &actors[ix]
		}
		if actors[ix].lifetimeRatio > lifetimeActor.lifetimeRatio {
			lifetimeActor = &actors[ix]
		}
		fmt.Printf("actor - %v - %v\n", actors[ix].ratio, actors[ix].lifetimeRatio)
	}
	fmt.Printf("best actor - %v - %v\n", bestActor.ratio, bestActor.lifetimeRatio)
	printRules(bestActor)
	fmt.Printf("lifetime best actor - %v - %v\n", lifetimeActor.ratio, lifetimeActor.lifetimeRatio)
	printRules(lifetimeActor)
}

func dispatchHands(actors *[]Actor, numHands int) {
	total := len(*actors)
	chunk := total / ROUTINES
	results := 0
	rchan := make(chan int)
	for ix := 0; ix < ROUTINES-1; ix++ {
		go handRoutine(actors, numHands, ix*chunk, (ix+1)*chunk, rchan)
	}
	go handRoutine(actors, numHands, (ROUTINES-1)*chunk, total, rchan)
	for finished := range rchan {
		results += finished
		if results >= total {
			close(rchan)
		}
	}
}

func handRoutine(actors *[]Actor, numHands int, start int, end int, res chan int) {
	for ix := start; ix < end; ix++ {
		playHands(&(*actors)[ix], numHands)
	}
	res <- (end - start)
}

func playHands(actor *Actor, numHands int) {
	for i := 0; i < numHands; i++ {
		hand := []int{rand.Intn(10) + 1, rand.Intn(10) + 1}
		dealerHand := []int{rand.Intn(10) + 1, rand.Intn(10) + 1}
		finalSum := playHand(&hand, actor, dealerHand[0])
		if finalSum == 21 {
			actor.ratio += 4
		} else if finalSum > 21 {
			actor.ratio -= 2
		} else {
			dealerSum := playDealerHand(&dealerHand)
			if dealerSum > 21 {
				actor.ratio++
			} else if finalSum > dealerSum {
				actor.ratio += 2
			} else if finalSum == dealerSum {
				actor.ratio++
			} else {
				actor.ratio--
			}
		}
	}
}

func playHand(hand *[]int, actor *Actor, dealerCard int) int {
	for {
		//sum, soft := dualSumHand(hand)
		sum := sumHand(hand)
		switch {
		case sum > 21:
			return sum
		case sum == 21:
			return sum
		}
		switch actor.actions[dealerCard-1][sum-1] {
		case HOLD:
			return sum
		case HIT:
			*hand = append(*hand, rand.Intn(10)+1)
		}
	}
}

func sumToAction(sum int, soft bool) int {
	return sum - 1
	//	if soft {
	//		return sum + 20
	//	} else {
	//		return sum - 1
	//	}
}

func playDealerHand(hand *[]int) int {
	sum := sumHand(hand)
	for sum < 17 {
		sum = drawCard(hand)
	}
	return sum
}

func sumHand(hand *[]int) int {
	sum := 0
	sort.Sort(sort.Reverse(sort.IntSlice(*hand)))
	for ix, card := range *hand {
		if card == 1 && (sum+(len(*hand)-ix)) < 12 {
			sum += 11
		} else {
			sum += card
		}
	}
	return sum
}

func dualSumHand(hand *[]int) (int, bool) {
	sum := 0
	soft := false
	sort.Sort(sort.Reverse(sort.IntSlice(*hand)))
	for ix, card := range *hand {
		if card == 1 && (sum+(len(*hand)-ix)) < 12 {
			sum += 11
			soft = true
		} else {
			sum += card
		}
	}
	return sum, soft
}

func drawCard(hand *[]int) int {
	*hand = append(*hand, rand.Intn(10)+1)
	return sumHand(hand)
}

func printRules(actor *Actor) {
	fmt.Println("Dealer:  A  2  3  4  5  6  7  8  9  10")
	//	ruleCount := len(actor.actions[0])
	for rx := 3; rx < 21; rx++ {
		fmt.Printf("    %2d:  ", rx+1)
		for ry := 0; ry < 10; ry++ {
			fmt.Printf("%v  ", actor.actions[ry][rx])
		}
		fmt.Printf("\n")
	}
}

type ByRatio []Actor

func (a ByRatio) Len() int           { return len(a) }
func (a ByRatio) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByRatio) Less(i, j int) bool { return a[i].lifetimeRatio > a[j].lifetimeRatio }

// cmd/sevencard/main.go

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/paulhankin/poker/v2/poker"
)

var (
	handsFlag = flag.String("hands", "", "seven card hands to compare (format: AcKhQdJsTs9h8d)")
)

func parseHand(s string) ([7]poker.Card, error) {
	if len(s) != 14 {
		return [7]poker.Card{}, fmt.Errorf("hand must be exactly 7 cards (14 characters), got %q", s)
	}

	var hand [7]poker.Card
	for i := 0; i < 7; i++ {
		cardStr := s[i*2 : i*2+2]
		// Try first with suit then rank
		c, ok := poker.NameToCard[strings.ToUpper(cardStr)]
		if !ok {
			// Try with rank and suit reversed
			c, ok = poker.NameToCard[strings.ToUpper(string(cardStr[1])+string(cardStr[0]))]
		}
		if !ok {
			return hand, fmt.Errorf("invalid card at position %d: %q", i, cardStr)
		}
		hand[i] = c
	}
	return hand, nil
}

func main() {
	flag.Parse()

	if *handsFlag == "" {
		fmt.Fprintf(os.Stderr, "must specify one or more 7-card hands via the -hands flag\n")
		os.Exit(1)
	}

	handStrings := strings.Fields(*handsFlag)
	var hands [][7]poker.Card
	var scores []int16

	// Parse and evaluate each hand
	for _, handStr := range handStrings {
		hand, err := parseHand(handStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing hand %q: %v\n", handStr, err)
			os.Exit(1)
		}
		hands = append(hands, hand)
		score := poker.Eval7(&hand)
		scores = append(scores, score)
	}

	// Print evaluations and compare hands
	fmt.Println("Hand Evaluations:")
	for i, hand := range hands {
		desc, err := poker.Describe(hand[:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error describing hand: %v\n", err)
			continue
		}
		fmt.Printf("%s: %s (score: %d)\n", handStrings[i], desc, scores[i])
	}

	// Find and announce winner
	bestScore := scores[0]
	bestHand := 0
	ties := []int{0}

	for i := 1; i < len(scores); i++ {
		if scores[i] > bestScore {
			bestScore = scores[i]
			bestHand = i
			ties = []int{i}
		} else if scores[i] == bestScore {
			ties = append(ties, i)
		}
	}

	if len(ties) > 1 {
		fmt.Printf("\nTie between hands: ")
		for i, idx := range ties {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(handStrings[idx])
		}
		fmt.Println()
	} else {
		fmt.Printf("\nWinning hand: %s\n", handStrings[bestHand])
	}
}

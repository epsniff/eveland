package main

import (
	"fmt"
	"math/rand"
	"time"
)

// TradingHub represents a trading hub in the graph.
type TradingHub struct {
	ID    int
	Items []TradeItem
}

// TradeItem represents a trade item in the graph.
type TradeItem struct {
	ID   int
	Buy  float64
	Sell float64
}

// TradeRoute represents a trade route between two trading hubs.
type TradeRoute struct {
	Start    *TradingHub
	End      *TradingHub
	Distance float64
	Profit   float64
}

// Ant represents an ant that traverses the graph to find the best trade route.
type Ant struct {
	ID           int
	CurrentHub   *TradingHub
	VisitedHubs  map[int]bool
	CurrentRoute *TradeRoute
	Profit       float64
	TravelTime   float64
}

// Graph represents the trading hubs and the trade routes between them.
type Graph struct {
	Hubs   []*TradingHub
	Routes []*TradeRoute
}

// InitGraph initializes the graph with random data.
func InitGraph(numHubs int, numItems int) *Graph {
	rand.Seed(time.Now().UnixNano())

	// Initialize the hubs.
	hubs := make([]*TradingHub, numHubs)
	for i := 0; i < numHubs; i++ {
		items := make([]TradeItem, numItems)
		for j := 0; j < numItems; j++ {
			items[j] = TradeItem{
				ID:   j,
				Buy:  rand.Float64() * 10,
				Sell: rand.Float64() * 10,
			}
		}
		hubs[i] = &TradingHub{
			ID:    i,
			Items: items,
		}
	}

	// Initialize the routes.
	routes := make([]*TradeRoute, 0)
	for i := 0; i < numHubs; i++ {
		for j := i + 1; j < numHubs; j++ {
			distance := rand.Float64() * 1000
			profit := rand.Float64() * 10
			routes = append(routes, &TradeRoute{
				Start:    hubs[i],
				End:      hubs[j],
				Distance: distance,
				Profit:   profit,
			})
			routes = append(routes, &TradeRoute{
				Start:    hubs[j],
				End:      hubs[i],
				Distance: distance,
				Profit:   profit,
			})
		}
	}

	return &Graph{
		Hubs:   hubs,
		Routes: routes,
	}
}

// AntColony implements the ant colony algorithm to find the best trade route.
func AntColony(graph *Graph, numAnts int, numIterations int, maxHops int, minMargin float64) *TradeRoute {
	// Initialize the ants.
	ants := make([]*Ant, numAnts)
	for i := 0; i < numAnts; i++ {
		ants[i] = &Ant{
			ID:           i,
			CurrentHub:   graph.Hubs[rand.Intn(len(graph.Hubs))],
			VisitedHubs:  make(map[int]bool),
			CurrentRoute: nil,
			Profit:       0,
			TravelTime:   0,
		}
	}

	// Initialize the pheromone trails.
	pheromones := make(map[*TradeRoute]float64)
	for _, route := range graph.Routes {
		pheromones[route] = 1.0
	}

	// Run the ant colony algorithm for the specified number of iterations.
	bestRoute := &TradeRoute{
		Profit: -1,
	}
	for i := 0; i < numIterations; i++ {
		// Move each ant to a new hub.
		for _, ant := range ants {
			// If the ant has not visited any hubs, choose a random hub to visit.
			if len(ant.VisitedHubs) == 0 {
				ant.CurrentHub = graph.Hubs[rand.Intn(len(graph.Hubs))]
			} else {
				// Calculate the probability of each route from the current hub.
				probabilities := make(map[*TradeRoute]float64)
				totalProb := 0.0
				for _, route := range graph.Routes {
					if route.Start != ant.CurrentHub {
						continue
					}
					if ant.VisitedHubs[route.End.ID] {
						continue
					}
					if len(ant.VisitedHubs) >= maxHops {
						continue
					}
					if route.Profit/route.Distance < minMargin {
						continue
					}
					prob := pheromones[route] * (route.Profit / route.Distance)
					probabilities[route] = prob
					totalProb += prob
				}

				// Choose the next hub based on the probabilities.
				if totalProb == 0 {
					ant.CurrentHub = graph.Hubs[rand.Intn(len(graph.Hubs))]
				} else {
					r := rand.Float64() * totalProb
					p := 0.0
					for route, prob := range probabilities {
						p += prob
						if p >= r {
							ant.CurrentHub = route.End
							ant.CurrentRoute = route
							break
						}
					}
				}
			}

			// Update the ant's visited hubs and profit.
			if ant.CurrentRoute != nil {
				ant.VisitedHubs[ant.CurrentHub.ID] = true
				ant.Profit += ant.CurrentRoute.Profit
				ant.TravelTime += ant.CurrentRoute.Distance
			}
		}

		// Update the pheromone trails based on the ant's paths.
		for _, route := range graph.Routes {
			deltaPheromone := 0.0
			for _, ant := range ants {
				if ant.CurrentRoute == route {
					deltaPheromone += ant.Profit / ant.TravelTime
				}
			}
			pheromones[route] = (1-0.5)*pheromones[route] + deltaPheromone
		}

		// Find the best trade route so far.
		for _, ant := range ants {
			if ant.CurrentRoute != nil && ant.Profit > bestRoute.Profit {
				bestRoute = ant.CurrentRoute
			}
		}
	}

	return bestRoute
}

func main() {
	graph := InitGraph(10, 5)
	route := AntColony(graph, 100, 1000, 10, 0.05)
	fmt.Println(route.Start.ID, "->", route.End.ID)
}

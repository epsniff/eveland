package dbjumpcal

import (
	"container/list"
	"database/sql"
	"fmt"

	"github.com/epsniff/eveland/src/dbsdeutils"
)

func GetSystemID(systemName, dbpath string) (int, error) {
	// Open the SDE SQLite database
	db, err := dbsdeutils.LoadDB(dbpath)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	// Prepare an SQL query to retrieve the systemID for the given systemName
	// SAMPLE
	/*
		regionID: 10000001 constellationID: 20000001 solarSystemID: 30000002 solarSystemName: Lashesih x: -1.0330096826312646e+17 y: 4.1707503568269944e+16 z: -2.985630412979509e+16 xMin: -1.0330156508982766e+17 xMax: -1.0329952828984021e+17 yMin: 4.1707467566727784e+16 yMax: 4.170759041342664e+16 zMin: 2.985617328250696e+16 zMax: 2.9856912011988012e+16 luminosity: 0.01282 border: true fringe: false corridor: true hub: false international: true regional: true constellation: <nil> security: 0.7516891466979871 factionID: 500007 radius: 1.018399993728e+12 sunTypeID: 45037 securityClass: B
	*/
	stmt, err := db.Prepare("SELECT solarSystemID FROM mapSolarSystems WHERE solarSystemName = ?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	// Execute the query with the given systemName as a parameter
	var systemID int
	err = stmt.QueryRow(systemName).Scan(&systemID)
	if err != nil {
		return 0, err
	}

	// Return the systemID
	return systemID, nil
}

func SystemsWithinNJumps(startSystemId, nJumps int, dbpath string) (SystemGraph, error) {
	db, err := dbsdeutils.LoadDB(dbpath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	res, err := search(db, int(startSystemId), nJumps)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("Systems within %d jumps of %d:  results:\n%v\n", nJumps, startSystemId, res)

	return res, nil
}

type Node struct {
	ID        int
	Depth     int
	Neighbors []int
}

type SystemGraph map[int]*Node

func (s SystemGraph) String() string {
	res := ""
	for sid, node := range s {
		res = res + fmt.Sprintf("   system:%v (depth:%v) -> %v \n", sid, node.Depth, node.Neighbors)
	}
	return res
}

// search - uses BFS to find all systems within maxDepth jumps of systemID
func search(db *sql.DB, systemID int, maxDepth int) (SystemGraph, error) {
	queue := list.New()

	graph := SystemGraph{}

	visited := make(map[int]int)

	// Add starting node to queue
	queue.PushBack(Node{ID: systemID, Depth: 0, Neighbors: []int{}})
	visited[systemID] = 0

	for queue.Len() != 0 {
		// Get next node from queue
		node := queue.Front()
		queue.Remove(node)
		currNode := node.Value.(Node)

		// Check if max depth has been reached
		if currNode.Depth == maxDepth {
			continue
		}

		graph[currNode.ID] = &currNode

		// Query for all systems that can be reached from current system
		/* Example data
		fromRegionID: 10000009 fromConstellationID: 20000114 fromSolarSystemID: 30000777 toSolarSystemID: 30000778 toConstellationID: 20000114 toRegionID: 10000009
		fromRegionID: 10000009 fromConstellationID: 20000114 fromSolarSystemID: 30000777 toSolarSystemID: 30000761 toConstellationID: 20000111 toRegionID: 10000009
		fromRegionID: 10000009 fromConstellationID: 20000114 fromSolarSystemID: 30000777 toSolarSystemID: 30000782 toConstellationID: 20000114 toRegionID: 10000009
		*/
		rows, err := db.Query("SELECT toSolarSystemID FROM mapSolarSystemJumps WHERE fromSolarSystemID=?", currNode.ID)
		if err != nil {
			return nil, fmt.Errorf("error querying database: %s", err.Error())
		}
		defer rows.Close()

		// Add unvisited neighbors to queue
		for rows.Next() {
			var neighbor int
			err := rows.Scan(&neighbor)
			if err != nil {
				return nil, fmt.Errorf("error scanning row: %s", err.Error())
			}
			currNode.Neighbors = append(currNode.Neighbors, neighbor)
			if _, ok := visited[neighbor]; !ok {
				visited[neighbor] = currNode.Depth + 1
				queue.PushBack(Node{ID: neighbor, Depth: currNode.Depth + 1, Neighbors: []int{}})
			}
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating over rows: %s", err.Error())
		}
	}

	return graph, nil
}

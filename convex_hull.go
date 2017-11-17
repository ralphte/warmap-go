package main

import "sort"

//Swap swaps the location of two Point structs in a Points struct
func (points Points) Swap(i, j int) {
	points[i], points[j] = points[j], points[i]
}

//Len is custom length definition for Points
func (points Points) Len() int {
	return len(points)
}

//Less sorts Points by x and, if equal, by y
func (points Points) Less(i, j int) bool {
	if points[i].X == points[j].X {
		return points[i].Y < points[j].Y
	}
	return points[i].X < points[j].X
}

//crossProduct returns the modulo (and sign) of the cross product between vetors OA and OB
func crossProduct(A, B, O Point) float64 {
	return (A.X-O.X)*(B.Y-O.Y) - (A.Y-O.Y)*(B.X-O.X)
}

// findConvexHull returns a slice of Point with a convex hull
// it is counterclockwise and starts and ends at the same point
func findConvexHull(points Points) Points {
	n := len(points)  // number of points to find convex hull
	var result Points // final result
	count := 0        // size of our convex hull (number of points added)
	// lets sort our points by x and if equal by y
	sort.Sort(points)
	if n == 0 {
		return result
	}
	// add the first element:
	result = append(result, points[0])
	count++
	//find the lower hull
	for i := 1; i < n; i++ {
		// remove points which are not part of the lower hull
		for count > 1 && crossProduct(result[count-2], result[count-1], points[i]) < 0.00000000000000001 {
			count--
			result = result[:count]
		}
		// add a new better point than the removed ones
		result = append(result, points[i])
		count++
	}
	count0 := count // our base counter for the upper hull
	// find the upper hull
	for i := n - 2; i >= 0; i-- {
		// remove points which are not part of the upper hull
		for count-count0 > 0 && crossProduct(result[count-2], result[count-1], points[i]) < 0.00000000000000001 {
			count--
			result = result[:count]
		}
		// add a new better point than the removed ones
		result = append(result, points[i])
		count++
	}
	return result
}

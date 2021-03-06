package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"sort"
)

// Object is an object in space.
type Object byte

const (
	objectSpace    Object = '.'
	objectAsteroid Object = '#'
)

// Point is a point in space.
type Point struct {
	X, Y int
}

// Field is a field of asteroids.
type Field struct {
	Width, Height int
	Asteroids     map[Point]Object
}

// ParseAsteroids parses input into an asteroid field.
func ParseAsteroids(r io.Reader) (*Field, error) {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanBytes)

	f := &Field{
		Asteroids: make(map[Point]Object),
	}

	var (
		p          Point
		maxX, maxY int
	)

	for scanner.Scan() {
		obj := Object(scanner.Bytes()[0])
		switch obj {
		case '\n':
			p.X = 0
			p.Y++

		case objectSpace:
			p.X++

		case objectAsteroid:
			f.Asteroids[p] = Object(obj)

			if p.X > maxX {
				maxX = p.X
			}
			if p.Y > maxY {
				maxY = p.Y
			}

			p.X++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan input: %w", err)
	}

	f.Width = maxX + 1
	f.Height = maxY + 1

	return f, nil
}

// IsVisible returns true if point p2 is visible from point p1.
func IsVisible(field *Field, p1, p2 Point) bool {
	// Check each other point to see if it is between p1 and p2.
	for p3 := range field.Asteroids {
		if p3 == p1 || p3 == p2 {
			continue
		}

		// Ideas from: https://stackoverflow.com/a/328122

		// If cross product is zero, points are colinear. Otherwise can't occlude.
		cross := crossproduct(p2, p3, p1)
		if cross != 0 {
			continue
		}

		d12 := lineDistance(p1, p2)
		d13 := lineDistance(p1, p3)

		dot := dotproduct(p2, p3, p1)

		if dot >= 0 && (float64(dot) <= (lineDistance(p2, p3) * lineDistance(p2, p3))) {
			// p1 is between p2 and p3.  p3 cannot occlude p2.
			continue
		}

		if d13 < d12 {
			// p3 is closer to p1 than p2.
			return false
		}
	}

	return true
}

// VisibleAsteroids finds the number of asteroids visible from the given
// point.
func VisibleAsteroids(field *Field, p1 Point) []Point {
	asteroids := []Point{}

	for p2 := range field.Asteroids {
		if p2 == p1 {
			continue
		}

		if IsVisible(field, p1, p2) {
			asteroids = append(asteroids, p2)
		}
	}

	return asteroids
}

// FindMonitoringStation finds the asteroids point that can see the most other
// asteroids.  Returns that point, and the number of asteroids.
func FindMonitoringStation(field *Field) (Point, int) {
	var (
		maxCount int
		maxPoint Point
	)

	for p := range field.Asteroids {
		count := len(VisibleAsteroids(field, p))
		if count > maxCount {
			maxCount = count
			maxPoint = p
		}
	}

	return maxPoint, maxCount
}

// sortAngle sorts asteroids clockwise around point.
func sortAngle(asteroids []Point, c Point) {
	sort.Slice(asteroids, func(i, j int) bool {
		p1 := asteroids[i]
		p2 := asteroids[j]

		return getAngle(p2, c) > getAngle(p1, c)
	})
}

func getAngle(p1, p2 Point) float64 {
	angle := math.Atan2(float64(p2.Y-p1.Y), float64(p2.X-p1.X))

	// Rotate so that 0 is up
	angle -= (math.Pi / 2)

	// Extend range from 0 -> Pi -> -Pi -> 0 to 0 -> 2Pi
	if angle < 0 {
		angle = (2 * math.Pi) + angle
	}

	return angle - (math.Pi / 2)
}

// Vaporize returns the order of asteroid points vaporized by the rotating
// laser at c.
func Vaporize(field *Field, c Point) []Point {
	vaporized := make([]Point, 0, len(field.Asteroids))

	for len(field.Asteroids) > 1 {
		next := VisibleAsteroids(field, c)

		sortAngle(next, c)

		for _, p := range next {
			vaporized = append(vaporized, p)
			delete(field.Asteroids, p)
		}
	}

	return vaporized
}

func crossproduct(p1, p2, p3 Point) int {
	return ((p3.Y - p1.Y) * (p2.X - p1.X)) - ((p3.X - p1.X) * (p2.Y - p1.Y))
}

func dotproduct(p1, p2, p3 Point) int {
	return ((p3.X - p1.X) * (p2.X - p1.X)) + ((p3.Y - p1.Y) * (p2.Y - p1.Y))
}

// positive if to the right and below, negative if to the left and above
func lineDistance(p1, p2 Point) float64 {
	x1 := float64(p1.X)
	y1 := float64(p1.Y)

	x2 := float64(p2.X)
	y2 := float64(p2.Y)

	x := x2 - x1
	y := y2 - y1

	d := math.Sqrt((x * x) + (y * y))

	return d
}

func onLine(m, b float64, p Point) bool {
	return m*float64(p.X)+b == float64(p.Y)
}

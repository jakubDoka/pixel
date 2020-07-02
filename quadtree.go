package pixel

type Collidable interface {
	GetRect() Rect
}

// i separated this data so i can simply copy it to subnodes
type Common struct {
	Depth int // maximal level tree can reach
	level int // this takes track of level
	Cap   int // max amount of objects per quadrant, if there is more quadrant splits
}

type Quadtree struct {
	Rect
	Common
	nodes  []*Quadtree
	shapes []Collidable
}

// Creates new quad tree reference.
// bounds - defines position of quad tree and its size. If shapes goes out of bounds they
// will not be assigned to quadrants and the tree will be ineffective.
// depth - resolution of quad tree. It lavais splits in half so if bounds size is 100 x 100
// and depth is 2 smallest quadrants will be 25 x 25. Making resolution too high is redundant
// if shapes cannot fit into smallest quadrants.
// cap - sets maximal capacity of quadrant before it splits to 4 smaller. Making can too big is
// inefficient. optimal value can be 10 but its allways better to test what works the best.
func NewQuadtree(bounds Rect, depth, cap int) *Quadtree {
	return &Quadtree{
		Rect: bounds,
		Common: Common{
			Depth: depth,
			Cap:   cap,
		},
	}
}

//generates subquadrants
func (q *Quadtree) split() {
	q.nodes = make([]*Quadtree, 4)
	newCommon := q.Common
	newCommon.level++
	halfH := q.H() / 2
	halfW := q.W() / 2
	center := q.Center()
	//top-left
	q.nodes[0] = &Quadtree{
		Rect: Rect{
			Min: V(q.Min.X, q.Min.Y+halfH),
			Max: V(q.Max.X-halfW, q.Max.Y),
		},
		Common: newCommon,
	}
	//top-right
	q.nodes[1] = &Quadtree{
		Rect: Rect{
			Min: center,
			Max: q.Max,
		},
		Common: newCommon,
	}
	//bottom-left
	q.nodes[2] = &Quadtree{
		Rect: Rect{
			Min: q.Min,
			Max: center,
		},
		Common: newCommon,
	}
	//bottom-right
	q.nodes[3] = &Quadtree{
		Rect: Rect{
			Min: V(q.Min.X+halfW, q.Min.Y),
			Max: V(q.Max.X, q.Min.Y+halfH),
		},
		Common: newCommon,
	}
}

// finds out to witch subquadrant the shape belongs to. Shape has to overlap only with one quadrant,
// otherwise it returns -1
func (q *Quadtree) getSub(rect Rect) int8 {
	vertical := q.Min.X + q.W()/2
	horizontal := q.Min.Y + q.H()/2

	if rect.Max.X < q.Min.X || rect.Max.X > q.Max.X || rect.Min.Y < q.Min.Y || rect.Max.Y > q.Max.Y {
		return -1
	}

	left := rect.Max.X < vertical
	right := rect.Min.X > vertical
	if rect.Min.Y > horizontal {
		// top
		if left {
			return 0 // left
		} else if right {
			return 1 // right
		}
	} else if rect.Max.Y < horizontal {
		// bottom
		if left {
			return 2 // left
		} else if right {
			return 3 // right
		}
	}
	return -1
}

// Adds the shape to quad tree and asians it to correct quadrant.
// Proper way is adding all shapes first and then detecting collisions.
// For struct to implement Collidable interface it has to have
// GetRect() *pixel.Rect defined. GetRect function also slightly affects performance.
func (q *Quadtree) Insert(collidable Collidable) {
	rect := collidable.GetRect()
	// this is little memory expensive but it makes acesing shapes faster
	q.shapes = append(q.shapes, collidable)
	if len(q.nodes) != 0 {
		i := q.getSub(rect)
		if i != -1 {
			q.nodes[i].Insert(collidable)
		}
		return
	} else if q.Cap == len(q.shapes) && q.level != q.Depth {
		q.split()
		for _, s := range q.shapes {
			i := q.getSub(s.GetRect())
			if i != -1 {
				q.nodes[i].Insert(s)
			}
		}
	}
}

// gets smallest generated quadrant that rect fits into
func (q *Quadtree) getQuad(rect Rect) *Quadtree {
	if len(q.nodes) == 0 {
		return q
	}
	subIdx := q.getSub(rect)
	if subIdx == -1 {
		return q
	}
	return q.nodes[subIdx].getQuad(rect)
}

// returns all collidables that this rect can possibly collide with
// thought it also returns the shape it self if it wos inserted
func (q *Quadtree) Retrieve(rect Rect) []Collidable {
	return q.getQuad(rect).shapes
}

// returns all coliding shapes
func (q *Quadtree) GetColliding(collidable Collidable) []Collidable {
	var res []Collidable
	rect := collidable.GetRect()
	for _, c := range q.Retrieve(rect) {
		if c.GetRect().Intersects(rect) && c != collidable {
			res = append(res, c)
		}
	}
	return res
}

// Resets the tree, use this every frame before inserting all shapes
// other wise you will run out of memory eventually and tree will not even work properly
func (q *Quadtree) Clear() {
	q.shapes = []Collidable{}
	q.nodes = []*Quadtree{}
}

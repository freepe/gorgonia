package tensori

import (
	"testing"

	"github.com/chewxy/gorgonia/tensor/types"
	"github.com/stretchr/testify/assert"
)

func TestAt(t *testing.T) {
	backing := RangeInt(0, 6)
	T := NewTensor(WithShape(2, 3), WithBacking(backing))
	zeroone := T.At(0, 1)
	assert.Equal(t, int(1), zeroone)

	oneone := T.At(1, 1)
	assert.Equal(t, int(4), oneone)

	fail := func() {
		T.At(1, 2, 3)
	}
	assert.Panics(t, fail, "Expected too many coordinates to panic")

	backing = RangeInt(0, 24)
	T = NewTensor(WithShape(2, 3, 4), WithBacking(backing))
	/*
		T = [0, 1, 2, 3]
			[4, 5, 6, 7]
			[8, 9, 10, 11]

			[12, 13, 14, 15]
			[16, 17, 18, 19]
			[20, 21, 22, 23]
	*/
	oneoneone := T.At(1, 1, 1)
	assert.Equal(t, int(17), oneoneone)
	zthreetwo := T.At(0, 2, 2)
	assert.Equal(t, int(10), zthreetwo)
	onetwothree := T.At(1, 2, 3)
	assert.Equal(t, int(23), onetwothree)

	fail = func() {
		T.At(0, 3, 2)
	}
	assert.Panics(t, fail)
}

func TestT_transposeIndex(t *testing.T) {
	assert := assert.New(t)
	var T *Tensor

	T = NewTensor(WithShape(2, 2), WithBacking(RangeInt(0, 4)))

	correct := []int{0, 2, 1, 3}
	for i, v := range correct {
		assert.Equal(v, T.transposeIndex(i, []int{1, 0}, []int{2, 1}))
	}
}

func TestTranspose(t *testing.T) {
	assert := assert.New(t)
	var backing []int
	var correct []int
	var T *Tensor
	var err error
	backing = []int{1, 2, 3, 4}

	t.Log("Testing 4x1 column vector transpose")
	T = NewTensor(WithShape(4, 1), WithBacking(backing))
	T.T()
	// don't actually have to do transpose. We can hence test to see if the thunking works
	assert.Equal(types.Shape{1, 4}, T.Shape())
	assert.Equal([]int{1}, T.Strides())

	t.Log("Testing untransposing of a thunk'd 1x4 vector transpose - tests thunk")
	T.T()
	assert.Equal(types.Shape{4, 1}, T.Shape())
	assert.Equal([]int{1}, T.Strides())
	assert.Nil(T.old)

	t.Log("Testing actually transposing a column vector into a row vector")
	T.T()
	T.Transpose()
	assert.Nil(T.old)
	assert.Equal(types.Shape{1, 4}, T.Shape()) // note, not the getter... but the actual data
	assert.Equal([]int{1}, T.ostrides())

	t.Log("Testing 2x2 matrix: standard transpose")
	T = NewTensor(WithShape(2, 2), WithBacking(backing))
	T.T()
	T.Transpose()

	correct = []int{1, 3, 2, 4}
	assert.Equal(correct, T.data, "Transpose of 2x2 matrix isn't correct")
	assert.Nil(T.old, "Expected transposeInfo to be nil after a DoTranspose()")
	assert.Nil(T.transposeWith)

	t.Log("Testing 2x2 Matrix: untransposing previously transposed")
	T.T()
	T.Transpose()
	assert.Equal(backing, T.data)
	assert.Nil(T.transposeWith)

	t.Log("Testing Transposing a transpose that is purely thunked")
	T.T()
	T.T()
	assert.Nil(T.old)
	assert.Nil(T.transposeWith)
	assert.Equal(backing, T.data, "Thunk'd transpose should have the same data as the original")

	t.Log("Texting 2x2 Matrix: do-nothing transpose")
	T.T(0, 1) // the axis is exactly the same as the axis
	t.Logf("%v", T.old)
	T.Transpose()
	assert.Equal(backing, T.data, "Do-Nothing transpose of 2x2 matrix isn't correct")
	assert.Nil(T.transposeWith)

	t.Log("Testing 2x2 Matrix: impossible axes")
	err = T.T(1, 2, 3, 4) // waay more axes than what the matrix has
	assert.NotNil(err, "Transpose should have failed")

	t.Log("Testing 2x2 Matrix: invalid axes")
	err = T.T(0, 5) // one of the axes is invalid
	assert.NotNil(err, "Transpose should have failed - one of the axes were invalid")

	t.Log("Testing 2x2 Matrix: repeated axes")
	T.T(0, 0) // meaningless permutation
	assert.NotNil(err, "Transpose should have failed - the axes were repeated")

	// This part onwards actually fully stress tests the algorithm
	// Basically trying on different variations of tensors.
	t.Log("Testing 4x2 Matrix: standard transpose")
	backing = RangeInt(0, 8)
	T = NewTensor(WithShape(4, 2), WithBacking(backing))
	t.Log("\tTesting thunked info while we're at it...")
	T.T()
	assert.Equal([]int{1, 0}, T.transposeWith, "Expected the transpose axes to be {1,0}")
	assert.NotNil(T.old)

	correct = []int{
		0, 2, 4, 6,
		1, 3, 5, 7,
	}
	T.Transpose()
	assert.Equal(correct, T.data, "Transpose of 4x2 matrix isn't correct")

	t.Log("Testing 3-Tensor (2x3x4): standard transpose")
	backing = RangeInt(0, 24)
	T = NewTensor(WithShape(2, 3, 4), WithBacking(backing))

	correct = []int{
		0, 12,
		4, 16,
		8, 20,

		1, 13,
		5, 17,
		9, 21,

		2, 14,
		6, 18,
		10, 22,

		3, 15,
		7, 19,
		11, 23,
	}
	T.T()
	T.Transpose()
	assert.Equal(correct, T.data, "Transpose of a (2,3,4) 3-tensor was incorrect")

	// backing has changed, so we need to actually create a new one
	t.Log("Testing 3-Tensor (2x3x4): (2,0,1) transpose")
	backing = RangeInt(0, 24)
	T = NewTensor(WithShape(2, 3, 4), WithBacking(backing))

	correct = []int{
		0, 4, 8,
		12, 16, 20,

		1, 5, 9,
		13, 17, 21,

		2, 6, 10,
		14, 18, 22,

		3, 7, 11,
		15, 19, 23,
	}
	T.T(2, 0, 1)
	T.Transpose()
	assert.Equal(correct, T.data, "Transpose(2,0,1) of a (2,3,4) 3-tensor was incorrect")

	t.Log("Testing Thunk'd transpose where it's a direct reverse")
	backing = RangeInt(0, 24)
	T = NewTensor(WithShape(2, 3, 4), WithBacking(backing))
	T.T(2, 0, 1)
	T.T(1, 2, 0) // reverse of 201
	assert.Nil(T.old)
	assert.Nil(T.transposeWith)

	t.Log("Testing Thunk'd transpose where it's NOT a direct reverse")
	T.T(2, 0, 1)
	T.T(1, 0, 2) // needs the result of the previous transpose before this can be done
	assert.Equal(correct, T.data, "The data should be as if a (2,0,1) transpose was done")
	assert.Equal(types.Shape{2, 4, 3}, T.Shape(), "The Shape() should be 2x4x3")
	assert.NotNil(T.old)
	assert.NotNil(T.transposeWith)

	t.Log("Testing 4-Tensor (2x3x4x5): Basic Transpose")
	backing = RangeInt(0, 2*3*4*5)
	T = NewTensor(WithShape(2, 3, 4, 5), WithBacking(backing))

	correct = []int{
		0, 60,
		20, 80,
		40, 100,

		5, 65,
		25, 85,
		45, 105,

		10, 70,
		30, 90,
		50, 110,

		15, 75,
		35, 95,
		55, 115,

		// new layer
		1, 61,
		21, 81,
		41, 101,

		6, 66,
		26, 86,
		46, 106,

		11, 71,
		31, 91,
		51, 111,

		16, 76,
		36, 96,
		56, 116,

		// new layer
		2, 62,
		22, 82,
		42, 102,

		7, 67,
		27, 87,
		47, 107,

		12, 72,
		32, 92,
		52, 112,

		17, 77,
		37, 97,
		57, 117,

		// new layer
		3, 63,
		23, 83,
		43, 103,

		8, 68,
		28, 88,
		48, 108,

		13, 73,
		33, 93,
		53, 113,

		18, 78,
		38, 98,
		58, 118,

		// new layer
		4, 64,
		24, 84,
		44, 104,

		9, 69,
		29, 89,
		49, 109,

		14, 74,
		34, 94,
		54, 114,

		19, 79,
		39, 99,
		59, 119,
	}
	T.T()
	T.Transpose()
	assert.Equal(correct, T.data, "Transpose of (2,3,4,5) 4-tensor isn't correct")
}

func TestTUT(t *testing.T) {
	assert := assert.New(t)
	var T *Tensor

	T = NewTensor(WithShape(2, 3, 4))
	T.T()
	T.UT()
	assert.Nil(T.old)
	assert.Nil(T.transposeWith)

	T.T(2, 0, 1)
	T.UT()
	assert.Nil(T.old)
	assert.Nil(T.transposeWith)
}

func TestTRepeat(t *testing.T) {
	assert := assert.New(t)
	var T, T2 *Tensor
	var expectedShape types.Shape
	var expectedData []int
	var err error

	// SCALARS

	T = NewTensor(AsScalar(int(3)))
	T2, err = T.Repeat(0, 3)
	if err != nil {
		t.Error(err)
	}

	if T == T2 {
		t.Error("Not supposed to be the same pointer")
	}
	expectedShape = types.Shape{3}
	expectedData = []int{3, 3, 3}
	assert.Equal(expectedData, T2.data)
	assert.Equal(expectedShape, T2.Shape())

	T2, err = T.Repeat(1, 3)
	if err != nil {
		t.Error(err)
	}

	expectedShape = types.Shape{1, 3}
	assert.Equal(expectedData, T2.data)
	assert.Equal(expectedShape, T2.Shape())

	// VECTORS

	// These are the rules for vector repeats:
	// 	- Vectors can repeat on axis 0 and 1
	// 	- For vanilla vectors, repeating on axis 0 and 1 is as if it were a colvec
	// 	- For non vanilla vectors, it's as if it were a matrix being repeated

	var backing = []int{1, 2}

	// repeats on axis 1: colvec
	T = NewTensor(WithShape(2, 1), WithBacking(backing))
	T2, err = T.Repeat(1, 3)
	if err != nil {
		t.Error(err)
	}

	expectedShape = types.Shape{2, 3}
	expectedData = []int{1, 1, 1, 2, 2, 2}
	assert.Equal(expectedData, T2.data)
	assert.Equal(expectedShape, T2.Shape())

	// repeats on axis 1: vanilla vector
	T = NewTensor(WithShape(2), WithBacking(backing))
	T2, err = T.Repeat(1, 3)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(expectedData, T2.data)
	assert.Equal(expectedShape, T2.Shape())

	// repeats on axis 1: rowvec
	T = NewTensor(WithShape(1, 2), WithBacking(backing))
	T2, err = T.Repeat(1, 3)
	if err != nil {
		t.Error(err)
	}
	expectedShape = types.Shape{1, 6}
	assert.Equal(expectedData, T2.data)
	assert.Equal(expectedShape, T2.Shape())

	// repeats on axis 0: vanilla vectors
	T = NewTensor(WithShape(2), WithBacking(backing))
	T2, err = T.Repeat(0, 3)
	if err != nil {
		t.Error(err)
	}
	expectedShape = types.Shape{6}
	assert.Equal(expectedData, T2.data)
	assert.Equal(expectedShape, T2.Shape())

	// repeats on axis 0: colvec
	T = NewTensor(WithShape(2, 1), WithBacking(backing))
	T2, err = T.Repeat(0, 3)
	if err != nil {
		t.Error(err)
	}
	expectedShape = types.Shape{6, 1}
	assert.Equal(expectedData, T2.data)
	assert.Equal(expectedShape, T2.Shape())

	// repeats on axis 0: rowvec
	T = NewTensor(WithShape(1, 2), WithBacking(backing))
	T2, err = T.Repeat(0, 3)
	if err != nil {
		t.Error(err)
	}
	expectedData = []int{1, 2, 1, 2, 1, 2}
	expectedShape = types.Shape{3, 2}
	assert.Equal(expectedData, T2.data)
	assert.Equal(expectedShape, T2.Shape())

	// repeats on -1 : all should have shape of (6)
	T = NewTensor(WithShape(2, 1), WithBacking(backing))
	T2, err = T.Repeat(-1, 3)
	if err != nil {
		t.Error(err)
	}
	expectedData = []int{1, 1, 1, 2, 2, 2}
	expectedShape = types.Shape{6}
	assert.Equal(expectedData, T2.data)
	assert.Equal(expectedShape, T2.Shape())

	T = NewTensor(WithShape(1, 2), WithBacking(backing))
	T2, err = T.Repeat(-1, 3)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(expectedData, T2.data)
	assert.Equal(expectedShape, T2.Shape())

	T = NewTensor(WithShape(2), WithBacking(backing))
	T2, err = T.Repeat(-1, 3)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(expectedData, T2.data)
	assert.Equal(expectedShape, T2.Shape())

	// MATRICES

	backing = []int{1, 2, 3, 4}

	/*
		1, 2,
		3, 4
	*/

	T = NewTensor(WithShape(2, 2), WithBacking(backing))
	T2, err = T.Repeat(-1, 1, 2, 1, 1)
	if err != nil {
		t.Error(err)
	}

	expectedShape = types.Shape{5}
	expectedData = []int{1, 2, 2, 3, 4}

	assert.Equal(expectedData, T2.data)
	assert.Equal(expectedShape, T2.Shape())

	/*
		1, 1, 2
		3, 3, 4
	*/
	T2, err = T.Repeat(1, 2, 1)
	if err != nil {
		t.Error(err)
	}
	expectedShape = types.Shape{2, 3}
	expectedData = []int{1, 1, 2, 3, 3, 4}
	assert.Equal(expectedData, T2.data)
	assert.Equal(expectedShape, T2.Shape())

	/*
		1, 2, 2,
		3, 4, 4
	*/
	T2, err = T.Repeat(1, 1, 2)
	if err != nil {
		t.Error(err)
	}
	expectedShape = types.Shape{2, 3}
	expectedData = []int{1, 2, 2, 3, 4, 4}
	assert.Equal(expectedData, T2.data)
	assert.Equal(expectedShape, T2.Shape())

	/*
		1, 2,
		3, 4,
		3, 4
	*/
	T2, err = T.Repeat(0, 1, 2)
	if err != nil {
		t.Error(err)
	}
	expectedShape = types.Shape{3, 2}
	expectedData = []int{1, 2, 3, 4, 3, 4}
	assert.Equal(expectedData, T2.data)
	assert.Equal(expectedShape, T2.Shape())

	/*
		1, 2,
		1, 2,
		3, 4
	*/
	T2, err = T.Repeat(0, 2, 1)
	if err != nil {
		t.Error(err)
	}
	expectedShape = types.Shape{3, 2}
	expectedData = []int{1, 2, 1, 2, 3, 4}
	assert.Equal(expectedData, T2.data)
	assert.Equal(expectedShape, T2.Shape())

	// MORE THAN 2D!!
	/*
		In:
			1, 2,
			3, 4,
			5, 6,

			7, 8,
			9, 10,
			11, 12
		Out:
			1, 2,
			3, 4
			3, 4
			5, 6

			7, 8,
			9, 10,
			9, 10,
			11, 12
	*/
	T = NewTensor(WithShape(2, 3, 2), WithBacking(RangeInt(1, 2*3*2+1)))
	T2, err = T.Repeat(1, 1, 2, 1)
	if err != nil {
		t.Error(err)
	}
	expectedShape = types.Shape{2, 4, 2}
	expectedData = []int{1, 2, 3, 4, 3, 4, 5, 6, 7, 8, 9, 10, 9, 10, 11, 12}
	assert.Equal(expectedData, T2.data)
	assert.Equal(expectedShape, T2.Shape())

	// broadcast errors
	T2, err = T.Repeat(0, 1, 2, 1)
	if err == nil {
		t.Error("Expected a broadacast/shapeMismatch error")
	}

	// generic repeat - repeat EVERYTHING by 2
	T2, err = T.Repeat(types.AllAxes, 2)
	if err != nil {
		t.Error(err)
	}
	expectedShape = types.Shape{24}
	expectedData = []int{1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9, 10, 10, 11, 11, 12, 12}
	assert.Equal(expectedData, T2.data)
	assert.Equal(expectedShape, T2.Shape())

	// generic repeat, axis specified
	T2, err = T.Repeat(2, 2)
	if err != nil {
		t.Error(err)
	}
	expectedShape = types.Shape{2, 3, 4}
	expectedData = []int{1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9, 10, 10, 11, 11, 12, 12}
	assert.Equal(expectedData, T2.data)
	assert.Equal(expectedShape, T2.Shape())

	// repeat scalars!
	T = NewTensor(AsScalar(int(3)))
	T2, err = T.Repeat(0, 5)
	if err != nil {
		t.Error(err)
	}
	expectedData = []int{3, 3, 3, 3, 3}
	expectedShape = types.Shape{5}
	assert.Equal(expectedData, T2.data)
	assert.Equal(expectedShape, T2.Shape())

	/* IDIOTS SECTION */

	// trying to repeat on a nonexistant axis - Vector
	T = NewTensor(WithShape(2, 1), WithBacking([]int{1, 2}))
	fails := func() {
		T.Repeat(2, 3)
	}
	assert.Panics(fails)

	T = NewTensor(WithShape(2, 3), WithBacking([]int{1, 2, 3, 4, 5, 6}))
	fails = func() {
		T.Repeat(3, 3)
	}
	assert.Panics(fails)
}

func TestTSlice(t *testing.T) {
	assert := assert.New(t)
	var T, V *Tensor
	var err error
	var correct []int
	var correctShape types.Shape
	var correctStride []int

	// slicing of vectors

	// vanillavec
	T = NewTensor(WithBacking(RangeInt(0, 4)), WithShape(4))
	t.Log("T[0]")
	if V, err = T.Slice(ss(0)); err != nil {
		t.Error(err)
	}
	assert.Equal([]int{int(0)}, V.data)

	t.Log("T[0:2]")
	V, err = T.Slice(makeRS(0, 2))
	if err != nil {
		t.Error(err)
	}
	assert.Equal([]int{0, 1}, V.data)

	// colvec
	T = NewTensor(WithBacking(RangeInt(0, 4)), WithShape(4, 1))
	t.Log("T[0]")
	V, err = T.Slice(ss(0))
	if err != nil {
		t.Error(err)
	}
	assert.Equal([]int{int(0)}, V.data)

	t.Log("T[1:3]")
	V, err = T.Slice(makeRS(1, 3))
	if err != nil {
		t.Error(err)
	}
	assert.Equal([]int{1, 2}, V.data)

	// rowvec
	T = NewTensor(WithBacking(RangeInt(0, 4)), WithShape(1, 4))
	t.Log("T[0]")
	if V, err = T.Slice(ss(0)); err != nil {
		t.Error(err)
	}
	assert.Equal([]int{0, 1, 2, 3}, V.data)

	t.Log("T[1:3] - will Error")
	if _, err = T.Slice(makeRS(1, 3)); err == nil {
		t.Error("Expected an error - dimension 0 only has a size of 1")
	}

	t.Log("T[:, 1:3]")
	if V, err = T.Slice(nil, makeRS(1, 3)); err != nil {
		t.Error(err)
	}
	assert.Equal([]int{1, 2}, V.data)

	t.Log("T[0, 0]")
	if V, err = T.Slice(ss(0), ss(0)); err != nil {
		t.Error(err)
	}
	assert.Equal([]int{int(0)}, V.data)

	// slicing of matrix
	t.Log("SLICING MATRICES")

	T = NewTensor(WithBacking(RangeInt(0, 12)), WithShape(3, 4))

	/*
		0,  1,  2,  3
		4,  5,  6,  7
		8,  9, 10, 11

		should yield

		0, 1, 2, 3
		4, 5, 6, 7
	*/
	t.Log("T[0:2]")
	V, err = T.Slice(makeRS(0, 2))
	if err != nil {
		t.Error(err)
	}

	correct = RangeInt(0, 8)
	correctShape = types.Shape{2, 4}
	correctStride = []int{4, 1}
	assert.Equal(correct, V.data)
	assert.Equal(correctShape, V.Shape())
	assert.Equal(correctStride, V.ostrides())

	/*
		0,  1,  2,  3
		4,  5,  6,  7
		8,  9, 10, 11

		should yield

		4, 5, 6, 7
	*/
	t.Log("T[1]")
	V, err = T.Slice(ss(1))
	if err != nil {
		t.Error(err)
	}

	correct = RangeInt(4, 8)
	correctShape = types.Shape{1, 4}
	correctStride = []int{1}

	assert.Equal(correct, V.data)
	assert.Equal(correctShape, V.Shape())
	assert.Equal(correctStride, V.ostrides())

	// should be the same as above - this is more testing rangeSlice and singleSlice similarity than anything
	t.Log("T[1:2]")
	V, err = T.Slice(makeRS(1, 2))
	if err != nil {
		t.Error(err)
	}

	assert.Equal(correct, V.data)
	assert.Equal(correctShape, V.Shape())
	assert.Equal(correctStride, V.ostrides())

	/*
		0,  1,  2,  3
		4,  5,  6,  7
		8,  9, 10, 11

		should yield

		C[2, 6, 10]
	*/
	t.Log("T[:, 2]")
	V, err = T.Slice(nil, ss(2))
	if err != nil {
		t.Error(err)
	}

	correct = RangeInt(2, 11)
	correctShape = types.Shape{3, 1}
	correctStride = []int{4}

	assert.Equal(correct, V.data)
	assert.Equal(correctShape, V.Shape())
	assert.Equal(correctStride, V.ostrides())

	/*
		0,  1,  2,  3
		4,  5,  6,  7
		8,  9, 10, 11

		should yield

		0, 1
		4, 5
		8, 9
	*/
	t.Log("T[:, 0:2]")
	V, err = T.Slice(nil, makeRS(0, 2))
	if err != nil {
		t.Error(err)
	}

	correct = RangeInt(0, 10)
	correctShape = types.Shape{3, 2}
	correctStride = []int{4, 1}

	assert.Equal(correct, V.data)
	assert.Equal(correctShape, V.Shape())
	assert.Equal(correctStride, V.ostrides())

	// please put on your realD)) 3D glasses

	T = NewTensor(WithBacking(RangeInt(0, 24)), WithShape(2, 3, 4))

	/*
		0   1   2   3
		4   5   6   7
		8   9  10  11

		12  13  14  15
		16  17  18  19
		20  21  22  23

		yields

		13  14
		17  18

	*/
	t.Log("T[1, 0:2, 1:3]")
	V, err = T.Slice(ss(1), makeRS(0, 2), makeRS(1, 3))
	if err != nil {
		t.Error(err)
	}
	correct = RangeInt(13, 19)
	correctShape = types.Shape{2, 2}
	correctStride = []int{4, 1}

	assert.Equal(correct, V.data)
	assert.Equal(correctShape, V.Shape())
	assert.Equal(correctStride, V.ostrides())

	/*
		0   1   2   3
		4   5   6   7
		8   9  10  11

		12  13  14  15
		16  17  18  19
		20  21  22  23

		yields

		17  18

	*/
	t.Log("T[1, 1, 1:3]")
	V, err = T.Slice(makeRS(1, 2), ss(1), makeRS(1, 3))
	if err != nil {
		t.Error(err)
	}

	correct = RangeInt(17, 19)
	correctShape = types.Shape{1, 2}
	correctStride = []int{1}

	assert.Equal(correct, V.data)
	assert.Equal(correctShape, V.Shape())
	assert.Equal(correctStride, V.ostrides())

	/*
		0   1   2   3
		4   5   6   7
		8   9  10  11

		12  13  14  15
		16  17  18  19
		20  21  22  23

		yields

		5    6

		17  18

	*/
	t.Log("T[:, 1, 1:3]")
	V, err = T.Slice(nil, ss(1), makeRS(1, 3))
	if err != nil {
		t.Error(err)
	}

	correct = RangeInt(5, 19)
	correctShape = types.Shape{2, 2}
	correctStride = []int{12, 1}

	assert.Equal(correct, V.data)
	assert.Equal(correctShape, V.Shape())
	assert.Equal(correctStride, V.ostrides())

	// T[0, :, 2]
	t.Log("T[0, :, 2]")
	V, err = T.Slice(ss(0), nil, ss(2))
	if err != nil {
		t.Error(err)
	}
	correct = RangeInt(2, 11)
	correctShape = types.Shape{3, 1}
	correctStride = []int{4}

	assert.Equal(correct, V.data)
	assert.Equal(correctShape, V.Shape())
	assert.Equal(correctStride, V.ostrides())

	// T[0, 1, 2]
	// willl yield a scalar
	t.Log("T[0,1,2]")
	V, err = T.Slice(ss(0), ss(1), ss(2))
	if err != nil {
		t.Error(err)
	}
	assert.True(V.IsScalar())

	// And now, ladies and gentlemen, the idiots!

	// too many slices
	_, err = T.Slice(ss(1), ss(2), ss(3), ss(4))
	if err == nil {
		t.Error("Expected a DimMismatchError error")
	}

	// out of range sliced
	_, err = T.Slice(makeRS(1, 5))
	if err == nil {
		t.Error("Expected a IndexError")
	}

	// surely nobody can be this dumb? Having a start of negatives
	_, err = T.Slice(makeRS(-1, 1))
	if err == nil {
		t.Error("Expected a IndexError")
	}

}

func TestT_at_itol(t *testing.T) {
	assert := assert.New(t)
	var err error
	var T *Tensor
	var shape types.Shape

	T = NewTensor(WithBacking(RangeInt(0, 12)), WithShape(3, 4))
	t.Logf("%+v", T)

	shape = T.Shape()
	for i := 0; i < shape[0]; i++ {
		for j := 0; j < shape[1]; j++ {
			coord := []int{i, j}
			idx, err := T.at(coord...)
			if err != nil {
				t.Error(err)
			}

			got, err := T.itol(idx)
			if err != nil {
				t.Error(err)
			}

			assert.Equal(coord, got)
		}
	}

	T = NewTensor(WithBacking(RangeInt(0, 24)), WithShape(2, 3, 4))

	shape = T.Shape()
	for i := 0; i < shape[0]; i++ {
		for j := 0; j < shape[1]; j++ {
			for k := 0; k < shape[2]; k++ {
				coord := []int{i, j, k}
				idx, err := T.at(coord...)
				if err != nil {
					t.Error(err)
				}

				got, err := T.itol(idx)
				if err != nil {
					t.Error(err)
				}

				assert.Equal(coord, got)
			}
		}
	}

	/* Transposes */

	T = NewTensor(WithBacking(RangeInt(0, 6)), WithShape(2, 3))
	t.Logf("%+v", T)
	err = T.T()
	if err != nil {
		t.Error(err)
	}
	t.Logf("%v, %v", T.Shape(), T.Shape())
	t.Logf("%v, %v", T.Strides(), T.ostrides())

	shape = T.Shape()
	for i := 0; i < shape[0]; i++ {
		for j := 0; j < shape[1]; j++ {
			coord := []int{i, j}
			idx, err := T.at(coord...)
			if err != nil {
				t.Error(err)
				continue
			}

			got, err := T.itol(idx)
			if err != nil {
				t.Error(err)
				continue
			}

			assert.Equal(coord, got)
		}
	}

	/* IDIOT OF THE WEEK */

	T = NewTensor(WithBacking(RangeInt(0, 24)), WithShape(2, 3, 4))

	_, err = T.at(1, 3, 2) // the 3 is out of range
	if err == nil {
		t.Error("Expected an error")
	}
	t.Log(err)

	_, err = T.itol(24) // 24 is out of range
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestCopyTo(t *testing.T) {
	assert := assert.New(t)
	var T, T2, T3 *Tensor
	var err error

	T = NewTensor(WithShape(2), WithBacking([]int{1, 2}))
	T2 = NewTensor(WithShape(1, 2))

	err = T.CopyTo(T2)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(T2.data, T.data)

	// now, modify T1's data
	T.data[0] = 5000
	assert.NotEqual(T2.data, T.data)

	// test views
	T = NewTensor(WithShape(3, 3))
	T2 = NewTensor(WithShape(2, 2))
	T3, _ = T.Slice(makeRS(0, 2), makeRS(0, 2)) // T[0:2, 0:2], shape == (2,2)
	if err = T2.CopyTo(T3); err != nil {
		t.Log(err) // for now it's a not yet implemented error. TODO: FIX THIS
	}

	// dumbass time

	T = NewTensor(WithShape(3, 3))
	T2 = NewTensor(WithShape(2, 2))
	if err = T.CopyTo(T2); err == nil {
		t.Error("Expected an error")
	}

	if err = T.CopyTo(T); err != nil {
		t.Error("Copying a *Tensor to itself should yield no error. ")
	}

}

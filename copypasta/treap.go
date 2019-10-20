package copypasta

// 耗时大约是红黑树的 1.3 倍

import (
	. "fmt"
	"time"
)

var seed = uint(time.Now().UnixNano())

// https://www.jstatsoft.org/article/view/v008i14/xorshift.pdf
// https://en.wikipedia.org/wiki/Xorshift
func fastRand() uint {
	x := seed
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	seed = x
	return x
}

type tKeyType int   // *custom* 图方便可以全局替换
type tValueType int // *custom* 图方便可以全局替换

type tNode struct {
	lr       [2]*tNode
	priority uint // max heap
	sz       int
	msz      int
	key      tKeyType
	value    tValueType
}

func (o *tNode) size() int {
	if o != nil {
		return o.sz
	}
	return 0
}

func (o *tNode) mSize() int {
	if o != nil {
		return o.msz
	}
	return 0
}

func (o *tNode) pushUp() {
	sz := 1
	msz := int(o.value)
	if ol := o.lr[0]; ol != nil {
		sz += ol.sz
		msz += ol.msz
	}
	if or := o.lr[1]; or != nil {
		sz += or.sz
		msz += or.msz
	}
	o.sz = sz
	o.msz = msz
}

type treap struct {
	root       *tNode
	comparator func(a, b tKeyType) int
}

func newTreap() *treap {
	// 设置如下返回值是为了方便使用 tNode 中的 lr 数组
	return &treap{comparator: func(a, b tKeyType) int {
		if a < b {
			return 0
		}
		if a > b {
			return 1
		}
		return -1
	}}
}

func (t *treap) size() int   { return t.root.size() }
func (t *treap) empty() bool { return t.size() == 0 }

// d=0: left
// d=1: right
func (t *treap) rotate(o *tNode, d int) *tNode {
	x := o.lr[d^1]
	o.lr[d^1] = x.lr[d]
	x.lr[d] = o
	// x.msz = o.msz; o.pushUp()
	o.pushUp()
	x.pushUp()
	return x
}

func (t *treap) _put(o *tNode, key tKeyType, value tValueType) *tNode {
	if o == nil {
		return &tNode{priority: fastRand(), sz: 1, msz: 1, key: key, value: value}
	}
	if cmp := t.comparator(key, o.key); cmp >= 0 {
		o.lr[cmp] = t._put(o.lr[cmp], key, value)
		if o.lr[cmp].priority > o.priority {
			o = t.rotate(o, cmp^1)
		}
	} else {
		o.value = value
		//o.value += value
	}
	o.pushUp()
	return o
}

func (t *treap) put(key tKeyType, value tValueType) { t.root = t._put(t.root, key, value) }

func (t *treap) get(key tKeyType) *tNode {
	for o := t.root; o != nil; {
		if cmp := t.comparator(key, o.key); cmp >= 0 {
			o = o.lr[cmp]
		} else {
			return o
		}
	}
	return nil
}

func (t *treap) _delete(o *tNode, key tKeyType) *tNode {
	if o == nil {
		return nil
	}
	if cmp := t.comparator(key, o.key); cmp >= 0 {
		o.lr[cmp] = t._delete(o.lr[cmp], key)
	} else {
		//if o.value > 1 {
		//	o.value--
		//} else {
		if o.lr[1] == nil {
			return o.lr[0]
		}
		if o.lr[0] == nil {
			return o.lr[1]
		}
		// o 有两颗子树，先把优先级高的子树旋转到根，然后递归在另一颗子树中删除 o
		cmp2 := 0
		if o.lr[0].priority > o.lr[1].priority {
			cmp2 = 1
		}
		o = t.rotate(o, cmp2)
		o.lr[cmp2] = t._delete(o.lr[cmp2], key)
		//}
	}
	o.pushUp()
	return o
}

func (t *treap) delete(key tKeyType) { t.root = t._delete(t.root, key) }

//

func (t *treap) floor(key tKeyType) (floor *tNode) {
	for o := t.root; o != nil; {
		switch cmp := t.comparator(key, o.key); {
		case cmp == 0:
			o = o.lr[0]
		case cmp > 0:
			floor = o
			o = o.lr[1]
		default:
			return o
		}
	}
	return
}

// 前驱（小于 key，且最大的数）
func (t *treap) prev(key tKeyType) (prev *tNode) {
	// 另一种写法
	// rank, _ := t.mRank(key)
	// return t.mSelect(rank - 1)
	for o := t.root; o != nil; {
		if cmp := t.comparator(key, o.key); cmp <= 0 {
			o = o.lr[0]
		} else {
			prev = o
			o = o.lr[1]
		}
	}
	return
}

func (t *treap) ceiling(key tKeyType) (ceiling *tNode) {
	for o := t.root; o != nil; {
		switch cmp := t.comparator(key, o.key); {
		case cmp == 0:
			ceiling = o
			o = o.lr[0]
		case cmp > 0:
			o = o.lr[1]
		default:
			return o
		}
	}
	return
}

// 后继（大于 key，且最小的数)
func (t *treap) next(key tKeyType) (next *tNode) {
	// 另一种写法
	// rank, o := t.mRank(key)
	// if o != nil {
	// 	 rank += int(o.value)
	// }
	// return t.mSelect(rank)
	for o := t.root; o != nil; {
		if cmp := t.comparator(key, o.key); cmp != 0 {
			o = o.lr[1]
		} else {
			next = o
			o = o.lr[0]
		}
	}
	return
}

// 小于 key 的键的数量
func (t *treap) mRank(key tKeyType) (cnt int, o *tNode) {
	for o = t.root; o != nil; {
		switch cmp := t.comparator(key, o.key); {
		case cmp == 0:
			o = o.lr[0]
		case cmp > 0:
			cnt += int(o.value) + o.lr[0].mSize()
			o = o.lr[1]
		default:
			cnt += o.lr[0].mSize()
			return
		}
	}
	return
}

// 排名为 k 的节点 o（即有 k 个键小于 o.key）
func (t *treap) mSelect(k int) (o *tNode) {
	//if k < 0 {
	//	return
	//}
	for o = t.root; o != nil; {
		switch ls := o.lr[0].mSize(); {
		case k < ls:
			o = o.lr[0]
		case k > ls:
			k -= int(o.value) + ls
			if k < 0 {
				return
			}
			o = o.lr[1]
		default:
			return
		}
	}
	return
}

func (t *treap) min() (min *tNode) {
	for o := t.root; o != nil; o = o.lr[0] {
		min = o
	}
	return
}

func (t *treap) max() (max *tNode) {
	for o := t.root; o != nil; o = o.lr[1] {
		max = o
	}
	return
}

//

func (o *tNode) String() string {
	var s string
	if o.value == 1 {
		s = Sprintf("%v", o.key)
	} else {
		s = Sprintf("%v(%v)", o.key, o.value)
	}
	s += Sprintf("[sz:%d,msz:%d,p:%d]", o.sz, o.msz, o.priority)
	return s
}

func (o *tNode) draw(prefix string, isTail bool, str *string) {
	if o.lr[1] != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "│   "
		} else {
			newPrefix += "    "
		}
		o.lr[1].draw(newPrefix, false, str)
	}
	*str += prefix
	if isTail {
		*str += "└── "
	} else {
		*str += "┌── "
	}
	*str += o.String() + "\n"
	if o.lr[0] != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}
		o.lr[0].draw(newPrefix, true, str)
	}
}

func (t *treap) String() string {
	if t.empty() {
		return "Treap (empty)\n"
	}
	str := "Treap\n"
	t.root.draw("", true, &str)
	return str
}

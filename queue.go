package log_rotation

import (
	"sync"
)

var (
	// 触发缩容的outIndex位置
	n uint = 100
	// 触发扩容的nextInsert位置
	m uint = 800
	// 默认长度
	minLength uint = 1024
	// slice当前长度
	length = minLength
	// 扩容比例
	scaleNum uint = 2
)

type uChan struct {
	data       [][]byte
	lock       sync.Mutex
	nextInsert uint
	outIndex   uint
}

func newChan() *uChan {
	return &uChan{lock: sync.Mutex{}, data: make([][]byte, minLength), nextInsert: 0}
}

func (c *uChan) put(data []byte) {
	//fmt.Println("put ", string(data))
	c.lock.Lock()
	defer c.lock.Unlock()
	c.elastic()
	// 以为有些日志库底层集成了buf复用机制，故这里直接赋值会导致问题，需使用make申请足够的空间后将数据进行copy
	c.data[c.nextInsert] = make([]byte, len(data))
	copy(c.data[c.nextInsert], data)
	//c.data[c.nextInsert] = data
	c.nextInsert++
	//for i := range c.data {
	//	if i == int(c.nextInsert) {
	//		return
	//	}
	//	fmt.Println(string(c.data[i]))
	//}
	//fmt.Println("----------------")
}

// 用于slice的扩、缩容量
// 扩容之前检查outIndex的位置，如果可前移则迁移，且index = 移动前的index - 移动前的outIndex
// c.data[c.outIndex:c.nextInsert] 就是当前队列的有效数据
// 这里的机制就是将有效数据进行整体向前移动，且outIndex和nextInsert进行对应的位置调整
// *在调整之后应当尽快释放nextInsert之后的数据，以免数据没有产生覆盖的情况下导致内存泄漏
func (c *uChan) elastic() {
	//oldOutIndex := c.outIndex
	//oldNextIndex := c.nextInsert
	// 这里触发slice元素移动操作，用于缩减空间
	if c.outIndex >= n && c.nextInsert >= m {
		c._reset()
		//copy(c.data, c.data[c.outIndex:c.nextInsert])
		//// TODO: free invalid (data[:nextInsert]全部置零，避免持续的指针引用)
		//c.nextInsert -= c.outIndex
		//c.outIndex = 0
	}
	// 如果元素使用量 > 触发扩容机制基数m - 触发Index移动位置基数n，则对slice进行扩容操作，扩容公式 length = length * scaleNum,且对应的n和m也要对应的扩大
	if c.nextInsert >= m {
		length = length * scaleNum
		m = m * scaleNum
		n = n * scaleNum
		c._reset()
		//newData := make([]*[]byte, length)
		//copy(newData, c.data[c.outIndex:c.nextInsert])
		//// 扩容之后，索引位置发生改变
		//c.nextInsert -= c.outIndex
		//c.outIndex = 0
		//c.data = newData
	}
	// 缩容逻辑
	// 触发条件 有效数据个数（nextInsert - outIndex） < length / scalaNum * 2 && length > minLength
	if (c.nextInsert-c.outIndex < length/(scaleNum*2)) && length > minLength {
		length = length / scaleNum
		m = m / scaleNum
		n = n / scaleNum
		c._reset()
		//newData := make([]*[]byte, length)
		//copy(newData, c.data[c.outIndex:c.nextInsert])
		//// 扩容之后，索引位置发生改变
		//c.nextInsert -= c.outIndex
		//c.outIndex = 0
		//c.data = newData
	}
}

func (c *uChan) _reset() {
	//fmt.Println("reset size", length)
	//os.Exit(1)
	newData := make([][]byte, length)
	copy(newData, c.data[c.outIndex:c.nextInsert])
	// 扩容之后，索引位置发生改变
	c.nextInsert -= c.outIndex
	c.outIndex = 0
	c.data = newData
}

func (c *uChan) get() []byte {
	c.lock.Lock()
	defer c.lock.Unlock()
	// outIndex永远不会超过nextInsert的位置
	if c.outIndex == c.nextInsert {
		c.outIndex, c.nextInsert = 0, 0
		return nil
	}
	oldOut := c.outIndex
	c.outIndex++
	return c.data[oldOut]
}

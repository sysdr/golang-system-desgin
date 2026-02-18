package struct_analyzer

// WidgetBad demonstrates poor field ordering leading to more padding.
type WidgetBad struct {
	ID        int32  // 4 bytes
	IsActive  bool   // 1 byte
	Count     int64  // 8 bytes
	Name      string // 16 bytes (pointer + len)
	Timestamp int32  // 4 bytes
}

// WidgetGood demonstrates optimized field ordering to minimize padding.
type WidgetGood struct {
	Name      string // 16 bytes
	Count     int64  // 8 bytes
	ID        int32  // 4 bytes
	Timestamp int32  // 4 bytes
	IsActive  bool   // 1 byte
}

// Counter for demonstrating false sharing concept.
type Counter struct {
	Value int64    // 8 bytes
	_     [7]byte  // Pad for cache line
}

// PaddedCounter explicitly adds padding for cache line alignment.
type PaddedCounter struct {
	Value int64    // 8 bytes
	_     [56]byte // Padding to 64 bytes
}

// ProductInfo for assignment - initially poorly ordered.
type ProductInfo struct {
	Price    float64 // 8 bytes
	InStock  bool    // 1 byte
	SKU      string  // 16 bytes
	Quantity int16   // 2 bytes
	Category int32   // 4 bytes
}

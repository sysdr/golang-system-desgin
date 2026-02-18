package struct_analyzer

import (
	"fmt"
	"reflect"
	"unsafe"
)

// FieldInfo holds one field's analysis.
type FieldInfo struct {
	Name     string
	Type     string
	Offset   uintptr
	Size     uintptr
	Align    uintptr
	Padding  uintptr
}

// StructResult holds analysis result for one struct (for dashboard).
type StructResult struct {
	Name      string
	Size      uintptr
	Alignment uintptr
	Fields    []FieldInfo
	EndPadding uintptr
}

// AnalyzeStruct prints the size, alignment, and field offsets.
func AnalyzeStruct(name string, s interface{}, structSize uintptr) {
	val := reflect.ValueOf(s)
	typ := val.Type()
	if structSize == 0 {
		structSize = typ.Size()
	}
	alignment := unsafe.Alignof(s)
	fmt.Printf("\n--- Analyzing Struct: %s ---\n", name)
	fmt.Printf("Total Size: %d bytes\n", structSize)
	fmt.Printf("Alignment: %d bytes\n", alignment)
	var currentOffset uintptr = 0
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldOffset := field.Offset
		fieldSize := field.Type.Size()
		fieldAlign := field.Type.Align()
		if fieldOffset > currentOffset {
			padding := fieldOffset - currentOffset
			fmt.Printf("  [Padding: %d bytes]\n", padding)
		}
		fmt.Printf("  Field: %-15s Type: %-10s Offset: %-5d Size: %-5d Alignment: %d\n",
			field.Name, field.Type.String(), fieldOffset, fieldSize, fieldAlign)
		currentOffset = fieldOffset + fieldSize
	}
	if structSize > currentOffset {
		padding := structSize - currentOffset
		fmt.Printf("  [End Padding: %d bytes]\n", padding)
	}
	fmt.Println("------------------------------------")
}

// RunAnalysisToResults returns analysis results for all structs (for dashboard).
func RunAnalysisToResults() []StructResult {
	results := make([]StructResult, 0, 5)
	for _, nvs := range []struct {
		name string
		v    interface{}
		size uintptr
		align uintptr
	}{
		{"WidgetBad", WidgetBad{}, unsafe.Sizeof(WidgetBad{}), unsafe.Alignof(WidgetBad{})},
		{"WidgetGood", WidgetGood{}, unsafe.Sizeof(WidgetGood{}), unsafe.Alignof(WidgetGood{})},
		{"Counter", Counter{}, unsafe.Sizeof(Counter{}), unsafe.Alignof(Counter{})},
		{"PaddedCounter", PaddedCounter{}, unsafe.Sizeof(PaddedCounter{}), unsafe.Alignof(PaddedCounter{})},
		{"ProductInfo", ProductInfo{}, unsafe.Sizeof(ProductInfo{}), unsafe.Alignof(ProductInfo{})},
	} {
		results = append(results, analyzeToResult(nvs.name, nvs.v, nvs.size, nvs.align))
	}
	return results
}

func analyzeToResult(name string, s interface{}, structSize, alignment uintptr) StructResult {
	val := reflect.ValueOf(s)
	typ := val.Type()
	fields := make([]FieldInfo, 0, typ.NumField())
	var currentOffset uintptr = 0
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldOffset := field.Offset
		fieldSize := field.Type.Size()
		fieldAlign := field.Type.Align()
		var padding uintptr
		if fieldOffset > currentOffset {
			padding = fieldOffset - currentOffset
		}
		fields = append(fields, FieldInfo{
			Name:    field.Name,
			Type:    field.Type.String(),
			Offset:  fieldOffset,
			Size:    fieldSize,
			Align:   uintptr(fieldAlign),
			Padding: padding,
		})
		currentOffset = fieldOffset + fieldSize
	}
	var endPadding uintptr
	if structSize > currentOffset {
		endPadding = structSize - currentOffset
	}
	return StructResult{
		Name:       name,
		Size:       structSize,
		Alignment:  alignment,
		Fields:     fields,
		EndPadding: endPadding,
	}
}

// RunAnalysis orchestrates the analysis of all predefined structs (CLI).
func RunAnalysis() {
	AnalyzeStruct("WidgetBad", WidgetBad{}, unsafe.Sizeof(WidgetBad{}))
	AnalyzeStruct("WidgetGood", WidgetGood{}, unsafe.Sizeof(WidgetGood{}))
	AnalyzeStruct("Counter", Counter{}, unsafe.Sizeof(Counter{}))
	AnalyzeStruct("PaddedCounter", PaddedCounter{}, unsafe.Sizeof(PaddedCounter{}))
	AnalyzeStruct("ProductInfo (Assignment - Unoptimized)", ProductInfo{}, unsafe.Sizeof(ProductInfo{}))
	fmt.Println("\nInsight: WidgetGood is smaller than WidgetBad due to reordering.")
	fmt.Println("Insight: PaddedCounter demonstrates cache line alignment for avoiding false sharing.")
}

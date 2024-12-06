package core

import (
	"fmt"
	"reflect"
	"strconv"

	"time"
)

func ConvertToString(data interface{}) string {
	if num, ok := data.(float64); ok {
		str := strconv.FormatFloat(num, 'f', -1, 64)
		return str
	}
	if num, ok := data.(float32); ok {
		str := strconv.FormatFloat(float64(num), 'f', -1, 32)
		return str
	}
	if num, ok := data.(int); ok {
		str := strconv.FormatInt(int64(num), 10)
		return str
	}
	if num, ok := data.(int32); ok {
		str := strconv.FormatInt(int64(num), 10)
		return str
	}
	if num, ok := data.(int16); ok {
		str := strconv.FormatInt(int64(num), 10)
		return str
	}
	if num, ok := data.(int64); ok {
		str := strconv.FormatInt(num, 10)
		return str
	}
	if num, ok := data.(uint64); ok {
		str := strconv.FormatUint(num, 10)
		return str
	}
	if num, ok := data.(uint32); ok {
		str := strconv.FormatUint(uint64(num), 10)
		return str
	}
	if num, ok := data.(uint8); ok {
		str := strconv.FormatUint(uint64(num), 10)
		return str
	}
	if num, ok := data.(uint); ok {
		str := strconv.FormatUint(uint64(num), 10)
		return str
	}
	if byteSlice, ok := data.([]uint8); ok {
		str := string(byteSlice)
		return str
	}
	if str, ok := data.(string); ok {
		return str
	}
	if b, ok := data.(bool); ok {
		if b {
			return "1"
		} else {
			return "0"
		}
	}
	if t, ok := data.(time.Time); ok {
		return FormatTimeToString(t)
	}

	return ""
}

func ConvertToInt64(data interface{}) int64 {
	if num, ok := data.(int64); ok {
		return num
	}
	str := ConvertToString(data)
	num, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return num
}

func ConvertToInt32(data interface{}) int32 {
	if num, ok := data.(int32); ok {
		return num
	}
	num := ConvertToInt64(data)
	return int32(num)
}
func ConvertToInt16(data interface{}) int16 {
	if num, ok := data.(int16); ok {
		return num
	}
	num := ConvertToInt64(data)
	return int16(num)
}
func ConvertToInt(data interface{}) int {
	if num, ok := data.(int); ok {
		return num
	}
	num := ConvertToInt64(data)
	return int(num)
}

func ConvertToUint64(data interface{}) uint64 {
	if num, ok := data.(uint64); ok {
		return num
	}
	str := ConvertToString(data)
	num, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0
	}
	return num
}

func ConvertToUint32(data interface{}) uint32 {
	if num, ok := data.(uint32); ok {
		return num
	}
	num := ConvertToUint64(data)
	return uint32(num)
}
func ConvertToUint16(data interface{}) uint16 {
	if num, ok := data.(uint16); ok {
		return num
	}
	num := ConvertToUint64(data)
	return uint16(num)
}
func ConvertToUint8(data interface{}) uint8 {
	if num, ok := data.(uint8); ok {
		return num
	}
	num := ConvertToUint64(data)
	return uint8(num)
}
func ConvertToUint(data interface{}) uint {
	if num, ok := data.(uint); ok {
		return num
	}
	num := ConvertToUint64(data)
	return uint(num)
}

func ConvertToFloat64(data interface{}) float64 {
	if num, ok := data.(float64); ok {
		return num
	}
	str := ConvertToString(data)
	num, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0.0
	}
	return num
}

func ConvertToFloat32(data interface{}) float32 {
	if num, ok := data.(float32); ok {
		return num
	}
	num := ConvertToFloat64(data)

	return float32(num)
}

func ConvertToTimeString(data interface{}) string {
	if t, ok := data.(time.Time); ok {
		return FormatTimeToString(t)
	}
	t := ConvertToString(data)

	return t
}

func ConvertToTimeStringFromLocaleToUTC(data interface{}, locationString string) string {
	timeString := ConvertToTimeString(data)
	if timeString == "" {
		return timeString
	}
	return FormatTimeFromLocaleToUTC(timeString, locationString)
}

func ConvertToTimeStringFromUTCToLocale(data interface{}, locationString string) string {
	timeString := ConvertToTimeString(data)
	if timeString == "" {
		return timeString
	}
	return FormatTimeFromUTCToLocale(timeString, locationString)
}

func ConvertToTime(data interface{}) time.Time {
	if t, ok := data.(time.Time); ok {
		return t
	}
	t, err := time.Parse("2006-01-02 15:04:05", ConvertToString(data))
	if err != nil {
		fmt.Println("解析时间失败:", err)
		return time.Time{}
	}

	// 转换为 UTC 时间
	utcTime := t.UTC()

	return utcTime
}

func FormatTimeToString(t time.Time) string {
	// 指定时间格式
	layout := "2006-01-02 15:04:05"

	// 格式化时间为指定格式
	formattedTime := t.Format(layout)
	return formattedTime
}

func FormatTimeFromLocaleToUTC(input string, locationString string) string {
	return FormatTimeToString(ConvertStringFromLocaleToUTCTime(input, locationString))
}
func FormatTimeFromUTCToLocale(input string, locationString string) string {
	return FormatTimeToString(ConvertStringFromUTCToLocaleTime(input, locationString))
}

func ConvertStringFromLocaleToUTCTime(input string, locationString string) time.Time {
	layout := "2006-01-02 15:04:05"             // 输入字符串的格式
	loc, _ := time.LoadLocation(locationString) // 加载UTC时区

	// 使用ParseInLocation函数将字符串解析为UTC时间
	localTime, err := time.ParseInLocation(layout, input, loc)
	if err != nil {
		return time.Time{}
	}
	utcTime := localTime.UTC()
	return utcTime
}

func ConvertStringFromUTCToLocaleTime(input string, locationString string) time.Time {
	location, err := time.LoadLocation(locationString) // 指定要输出的时区
	if err != nil {
		return time.Time{}
	}
	layout := "2006-01-02 15:04:05" // 输入字符串的格式
	t, err := time.Parse(layout, input)
	if err != nil {
		return time.Time{}
	}

	utcTime := t.UTC()                // 将时间转换为 UTC 时间
	localTime := utcTime.In(location) // 转换为指定时区的时间

	return localTime
}

func updateField(obj interface{}, fieldName string, value interface{}) {
	if value == nil {
		return
	}
	// 獲取 obj 的反射值
	v := reflect.ValueOf(obj)

	// 檢查 obj 是否為指針
	if v.Kind() != reflect.Ptr {
		return
	}

	// 獲取 obj 指向的元素值
	elem := v.Elem()

	// 獲取指定 field 的反射值
	field := elem.FieldByName(fieldName)

	// 檢查 field 是否存在
	if !field.IsValid() {
		return
	}

	// 檢查 field 的類型是否與 value 的類型兼容
	if !field.Type().AssignableTo(reflect.TypeOf(value)) {
		return
	}

	// 設置 field 的值
	field.Set(reflect.ValueOf(value))

}
